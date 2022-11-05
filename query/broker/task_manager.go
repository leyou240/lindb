// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package brokerquery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/lindb/lindb/constants"
	"github.com/lindb/lindb/internal/concurrent"
	"github.com/lindb/lindb/metrics"
	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/encoding"
	"github.com/lindb/lindb/pkg/logger"
	"github.com/lindb/lindb/pkg/utils"
	protoCommonV1 "github.com/lindb/lindb/proto/gen/v1/common"
	"github.com/lindb/lindb/query"
	"github.com/lindb/lindb/rpc"
	"github.com/lindb/lindb/series"
	"github.com/lindb/lindb/sql/stmt"
)

//go:generate mockgen -source=./task_manager.go -destination=./task_manager_mock.go -package=brokerquery

// TaskManager represents the task manager for current node
type TaskManager interface {
	// SubmitMetricTask concurrently send query task to multi intermediates and leafs.
	// If intermediates are empty, the root waits response from leafs.
	// Otherwise, the roots waits response from intermediates.
	// 1. api -> metric-query -> SubmitMetricTask (query without intermediate nodes) -> leaf nodes
	//                                                                            -> leaf nodes -> response
	// 2. api -> metric-query -> SubmitMetricTask (query with intermediate nodes) <-> peer broker <->
	SubmitMetricTask(
		ctx context.Context,
		physicalPlan *models.PhysicalPlan,
		stmtQuery *stmt.Query,
	) (eventCh <-chan *series.TimeSeriesEvent, err error)

	// SubmitIntermediateMetricTask creates a intermediate task from leaf nodes
	// leaf response will also be merged in task-context.
	// when all intermediate response arrives to root, event will be returned to the caller
	SubmitIntermediateMetricTask(
		ctx context.Context,
		physicalPlan *models.PhysicalPlan,
		stmtQuery *stmt.Query,
		parentTaskID string,
	) (eventCh <-chan *series.TimeSeriesEvent)

	// SubmitMetaDataTask concurrently send query metadata task to multi leafs.
	SubmitMetaDataTask(
		ctx context.Context,
		physicalPlan *models.PhysicalPlan,
		suggest *stmt.MetricMetadata,
	) (taskResponse <-chan *protoCommonV1.TaskResponse, err error)

	// SendRequest sends the task request to target node based on node's indicator
	SendRequest(targetNodeID string, req *protoCommonV1.TaskRequest) error
	// SendResponse sends the task response to parent node
	SendResponse(targetNodeID string, resp *protoCommonV1.TaskResponse) error

	// Receive receives task response from rpc handler asynchronous
	Receive(req *protoCommonV1.TaskResponse, targetNode string) error
}

// taskManager implements the task manager interface, tracks all task of the current node
type taskManager struct {
	ctx               context.Context
	currentNodeID     string
	seq               *atomic.Int64
	taskClientFactory rpc.TaskClientFactory
	taskServerFactory rpc.TaskServerFactory

	workerPool concurrent.Pool // workers for
	tasks      sync.Map        // taskID -> taskCtx
	ttl        time.Duration

	statistics *metrics.BrokerQueryStatistics
	logger     *logger.Logger
}

// NewTaskManager creates the task manager
func NewTaskManager(
	ctx context.Context,
	currentNode models.Node,
	taskClientFactory rpc.TaskClientFactory,
	taskServerFactory rpc.TaskServerFactory,
	taskPool concurrent.Pool,
	ttl time.Duration,
) TaskManager {
	tm := &taskManager{
		ctx:               ctx,
		currentNodeID:     currentNode.Indicator(),
		taskClientFactory: taskClientFactory,
		taskServerFactory: taskServerFactory,
		seq:               atomic.NewInt64(0),
		workerPool:        taskPool,
		ttl:               ttl,
		statistics:        metrics.NewBrokerQueryStatistics(),
		logger:            logger.GetLogger("Query", "TaskManager"),
	}
	duration := ttl
	if ttl < time.Minute {
		duration = time.Minute
	}
	go tm.cleaner(duration)
	return tm
}

// cleaner cleans expired tasks in
func (t *taskManager) cleaner(duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.tasks.Range(func(key, value interface{}) bool {
				taskCtx := value.(TaskContext)
				if taskCtx.Expired(t.ttl) {
					t.statistics.AliveTask.Decr()
					t.statistics.ExpireTasks.Incr()
					t.tasks.Delete(key)
				}
				return true
			})
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *taskManager) evictTask(taskID string) {
	_, loaded := t.tasks.LoadAndDelete(taskID)
	if loaded {
		t.statistics.AliveTask.Decr()
	}
}

func (t *taskManager) storeTask(taskID string, taskCtx TaskContext) {
	t.tasks.Store(taskID, taskCtx)
	t.statistics.CreatedTasks.Incr()
	t.statistics.AliveTask.Incr()
}

func (t *taskManager) ensureIntermediateAckTasks(
	ctx context.Context,
	physicalPlan *models.PhysicalPlan,
	taskRequest *protoCommonV1.TaskRequest,
) error {
	var (
		wg        sync.WaitGroup
		sendError atomic.Error
	)
	responseCh := make(chan error)
	taskCtx := newIntermediateAckTaskContext(
		ctx,
		taskRequest.ParentTaskID,
		RootTask,
		int32(len(physicalPlan.Intermediates)),
		responseCh,
	)

	t.storeTask(taskRequest.ParentTaskID, taskCtx)
	defer t.evictTask(taskRequest.ParentTaskID)

	wg.Add(len(physicalPlan.Intermediates))
	for _, intermediate := range physicalPlan.Intermediates {
		intermediate := intermediate
		t.workerPool.Submit(ctx, concurrent.NewTask(func() {
			defer wg.Done()
			if err := t.SendRequest(intermediate.Indicator, taskRequest); err != nil {
				sendError.Store(err)
			}
		}, nil))
	}
	wg.Wait()
	if sendError.Load() != nil {
		return sendError.Load()
	}

	select {
	case event, ok := <-responseCh:
		if !ok {
			return fmt.Errorf("missing acks from intermdiate nodes")
		}
		return event
	case <-ctx.Done():
		return ErrTimeout
	}
}

func (t *taskManager) SubmitMetricTask(
	ctx context.Context,
	physicalPlan *models.PhysicalPlan,
	stmtQuery *stmt.Query,
) (eventCh <-chan *series.TimeSeriesEvent, err error) {
	rootTaskID := t.AllocTaskID()
	marshalledPhysicalPlan := encoding.JSONMarshal(physicalPlan)
	marshalledPayload, _ := stmtQuery.MarshalJSON()

	// checkpoint to ensure that all intermediate established the tasks.
	// in distributed environment, storage responses may be faster than intermediate nodes
	//
	// send task to intermediates firstly, then to the leafs
	// in case of too early response arriving without reader
	if len(physicalPlan.Intermediates) > 0 {
		req := &protoCommonV1.TaskRequest{
			ParentTaskID: rootTaskID,
			Type:         protoCommonV1.TaskType_Intermediate,
			RequestType:  protoCommonV1.RequestType_Data,
			PhysicalPlan: marshalledPhysicalPlan,
			Payload:      marshalledPayload,
		}
		if err := t.ensureIntermediateAckTasks(ctx, physicalPlan, req); err != nil {
			return nil, err
		}
	}

	// get request from context
	reqFromCtx := utils.GetFromContext(ctx, constants.ContextKeySQL)
	request, ok := reqFromCtx.(*models.Request)
	requestID := ""
	if ok {
		requestID = request.RequestID
	}

	responseCh := make(chan *series.TimeSeriesEvent)
	taskCtx := newMetricTaskContext(
		ctx,
		rootTaskID,
		RootTask,
		"",
		"",
		stmtQuery,
		physicalPlan.Root.NumOfTask,
		responseCh,
	)
	t.storeTask(rootTaskID, taskCtx)

	// return the channel for reader, then send the rpc request
	// in case of too early response arriving without reader
	var (
		wg        sync.WaitGroup
		sendError atomic.Error
	)
	// notify error to other peer nodes
	req := &protoCommonV1.TaskRequest{
		RequestID:    requestID,
		ParentTaskID: rootTaskID,
		Type:         protoCommonV1.TaskType_Leaf,
		RequestType:  protoCommonV1.RequestType_Data,
		PhysicalPlan: marshalledPhysicalPlan,
		Payload:      marshalledPayload,
	}
	wg.Add(len(physicalPlan.Leaves))
	for _, leaf := range physicalPlan.Leaves {
		leaf := leaf
		t.workerPool.Submit(ctx, concurrent.NewTask(func() {
			defer wg.Done()
			if err := t.SendRequest(leaf.Indicator, req); err != nil {
				sendError.Store(err)
			}
		}, nil))
	}
	wg.Wait()

	if sendError.Load() != nil {
		t.evictTask(rootTaskID)
	}
	return responseCh, sendError.Load()
}

func (t *taskManager) SubmitIntermediateMetricTask(
	ctx context.Context,
	physicalPlan *models.PhysicalPlan,
	stmtQuery *stmt.Query,
	parentTaskID string,
) (eventCh <-chan *series.TimeSeriesEvent) {
	responseCh := make(chan *series.TimeSeriesEvent)
	taskCtx := newMetricTaskContext(
		ctx,
		parentTaskID,
		IntermediateTask,
		parentTaskID,
		physicalPlan.Root.Indicator,
		stmtQuery,
		int32(len(physicalPlan.Leaves)),
		responseCh,
	)

	t.storeTask(parentTaskID, taskCtx)
	return responseCh
}

func (t *taskManager) SubmitMetaDataTask(
	ctx context.Context,
	physicalPlan *models.PhysicalPlan,
	suggest *stmt.MetricMetadata,
) (taskResponse <-chan *protoCommonV1.TaskResponse, err error) {
	taskID := t.AllocTaskID()

	suggestMarshalData, _ := suggest.MarshalJSON()
	req := &protoCommonV1.TaskRequest{
		RequestType:  protoCommonV1.RequestType_Metadata,
		ParentTaskID: taskID,
		PhysicalPlan: encoding.JSONMarshal(physicalPlan),
		Payload:      suggestMarshalData,
	}

	responseCh := make(chan *protoCommonV1.TaskResponse)
	taskCtx := newMetaDataTaskContext(
		ctx,
		taskID,
		RootTask,
		"",
		"",
		physicalPlan.Root.NumOfTask,
		responseCh)

	t.storeTask(taskID, taskCtx)

	var (
		wg        sync.WaitGroup
		sendError atomic.Error
	)
	wg.Add(len(physicalPlan.Leaves))
	for _, leafNode := range physicalPlan.Leaves {
		leafNode := leafNode
		t.workerPool.Submit(ctx, concurrent.NewTask(func() {
			defer wg.Done()
			if err := t.SendRequest(leafNode.Indicator, req); err != nil {
				sendError.Store(err)
			}
		}, nil))
	}
	wg.Wait()
	if sendError.Load() != nil {
		t.evictTask(taskID)
	}
	return responseCh, sendError.Load()
}

// AllocTaskID allocates the task id for new task, before task submits
func (t *taskManager) AllocTaskID() string {
	seq := t.seq.Inc()
	return fmt.Sprintf("%s-%d", t.currentNodeID, seq)
}

// Get returns the task context by task id
func (t *taskManager) Get(taskID string) TaskContext {
	if task, ok := t.tasks.Load(taskID); ok {
		return task.(TaskContext)
	}
	return nil
}

// SendRequest sends the task request to target node based on node's indicator,
// if fail, returns err
func (t *taskManager) SendRequest(targetNodeID string, req *protoCommonV1.TaskRequest) error {
	t.logger.Debug("send query task", logger.String("target", targetNodeID))
	client := t.taskClientFactory.GetTaskClient(targetNodeID)
	if client == nil {
		t.statistics.SentRequestFailures.Incr()
		return fmt.Errorf("SendRequest: %w, targetNodeID: %s", query.ErrNoSendStream, targetNodeID)
	}
	if err := client.Send(req); err != nil {
		t.statistics.SentRequestFailures.Incr()
		return fmt.Errorf("%w, targetNodeID: %s", query.ErrTaskSend, targetNodeID)
	}
	t.statistics.SentRequest.Incr()
	return nil
}

// SendResponse sends the task response to parent node,
// if fail, returns err
func (t *taskManager) SendResponse(parentNodeID string, resp *protoCommonV1.TaskResponse) error {
	stream := t.taskServerFactory.GetStream(parentNodeID)
	if stream == nil {
		t.statistics.SentResponseFailures.Incr()
		return fmt.Errorf("SendResponse: %w, parentNodeID: %s", query.ErrNoSendStream, parentNodeID)
	}
	if err := stream.Send(resp); err != nil {
		t.statistics.SentResponseFailures.Incr()
		return fmt.Errorf("SendResponse: %w, parentNodeID: %s", query.ErrResponseSend, parentNodeID)
	}
	t.statistics.SentResponses.Incr()
	return nil
}

func (t *taskManager) Receive(resp *protoCommonV1.TaskResponse, targetNode string) error {
	taskCtx := t.Get(resp.TaskID)
	if taskCtx == nil {
		t.statistics.OmitResponse.Incr()
		return fmt.Errorf("TaskID: %s may be evicted", resp.TaskID)
	}
	t.statistics.EmitResponse.Incr()
	t.workerPool.Submit(taskCtx.Context(), concurrent.NewTask(func() {
		// for root task and intermediate task
		taskCtx.WriteResponse(resp, targetNode)

		if taskCtx.Done() {
			t.evictTask(resp.TaskID)
		}
	}, nil))
	return nil
}
