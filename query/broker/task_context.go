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
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/lindb/common/pkg/fasttime"

	"github.com/lindb/lindb/aggregation"
	"github.com/lindb/lindb/aggregation/function"
	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/encoding"
	protoCommonV1 "github.com/lindb/lindb/proto/gen/v1/common"
	"github.com/lindb/lindb/series"
	"github.com/lindb/lindb/series/field"
	"github.com/lindb/lindb/sql/stmt"
)

var (
	newGroupingAgg = aggregation.NewGroupingAggregator
)

//go:generate mockgen -source=./task_context.go -destination=./task_context_mock.go -package=brokerquery

// TaskContext represents the task context for distribution query and computing
type TaskContext interface {
	// Context returns the context.
	Context() context.Context
	// Expired returns if this task is expired
	Expired(ttl time.Duration) bool
	// TaskID returns the id of the task
	TaskID() string
	// TaskType returns the task type
	TaskType() TaskType
	// ParentNode returns the parent node's indicator for sending task result
	ParentNode() string
	// ParentTaskID returns the parent node's task id for tracking task
	ParentTaskID() string

	WriteResponse(resp *protoCommonV1.TaskResponse, fromNode string)
	// Done returns if the task has been done
	Done() bool
}

type baseTaskContext struct {
	ctx          context.Context
	createTime   int64
	taskID       string
	taskType     TaskType
	parentTaskID string
	parentNode   string
	// race condition, we cannot make sure that
	// if another response wouldn't write to a closed channel without lock
	mu            sync.Mutex
	expectResults int32
	closed        bool
}

func (c *baseTaskContext) Context() context.Context {
	return c.ctx
}

func (c *baseTaskContext) Expired(ttl time.Duration) bool {
	return fasttime.UnixMilliseconds()-c.createTime > ttl.Milliseconds()
}

func (c *baseTaskContext) TaskType() TaskType {
	return c.taskType
}

// ParentNode returns the parent node's indicator for sending task result
func (c *baseTaskContext) ParentNode() string {
	return c.parentNode
}

// ParentTaskID returns the parent node's task id for tracking task
func (c *baseTaskContext) ParentTaskID() string {
	return c.parentTaskID
}

// TaskID returns the task id under current node
func (c *baseTaskContext) TaskID() string {
	return c.taskID
}

// Done returns if the task is completes
func (c *baseTaskContext) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.expectResults <= 0
}

// intermediateAckTaskContext represents the task context for root node
// tacking how many intermediate ack the intermediate task
type intermediateAckTaskContext struct {
	baseTaskContext

	eventCh chan<- error
}

func newIntermediateAckTaskContext(
	ctx context.Context,
	taskID string,
	taskType TaskType,
	expectResults int32,
	eventCh chan<- error,
) TaskContext {
	return &intermediateAckTaskContext{
		baseTaskContext: baseTaskContext{
			ctx:           ctx,
			taskID:        taskID,
			taskType:      taskType,
			parentTaskID:  "",
			parentNode:    "",
			expectResults: expectResults,
			closed:        false,
			createTime:    fasttime.UnixMilliseconds(),
		},
		eventCh: eventCh,
	}
}

func (c *intermediateAckTaskContext) WriteResponse(resp *protoCommonV1.TaskResponse, _ string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.expectResults--

	// preventing close channel twice
	if c.closed {
		return
	}
	defer func() {
		if c.expectResults <= 0 {
			close(c.eventCh)
			c.closed = true
		}
	}()
	if resp.ErrMsg != "" {
		select {
		case c.eventCh <- errors.New(resp.ErrMsg):
		default:
			// reader gone
		}
		return
	}
	// not done yet
	if c.expectResults > 0 {
		return
	}

	select {
	case c.eventCh <- nil:
	default:
		// reader gone
	}
}

// metricTaskContext represents the task context for tacking task execution state
type metricTaskContext struct {
	baseTaskContext

	eventCh   chan<- *series.TimeSeriesEvent
	stmtQuery *stmt.Query
	groupAgg  aggregation.GroupingAggregator
	stats     *models.QueryStats
	// field name -> aggregator spec
	// we will use it during intermediate tasks
	aggregatorSpecs map[string]*protoCommonV1.AggregatorSpec
	// tolerantNotFounds keeps the number of how many not found errors can be returned
	// if all nodes return not-found errors, it will be treated as a error
	// other error will be returned immediately
	tolerantNotFounds int32
	startTime         time.Time // task start time
}

// metricTaskContext creates the task context based on params
func newMetricTaskContext(
	ctx context.Context,
	taskID string,
	taskType TaskType,
	parentTaskID string,
	parentNode string,
	stmtQuery *stmt.Query,
	expectResults int32,
	eventCh chan<- *series.TimeSeriesEvent,
) TaskContext {
	return &metricTaskContext{
		baseTaskContext: baseTaskContext{
			ctx:           ctx,
			taskID:        taskID,
			taskType:      taskType,
			parentTaskID:  parentTaskID,
			parentNode:    parentNode,
			expectResults: expectResults,
			closed:        false,
			createTime:    fasttime.UnixMilliseconds(),
		},
		aggregatorSpecs:   make(map[string]*protoCommonV1.AggregatorSpec),
		stmtQuery:         stmtQuery,
		eventCh:           eventCh,
		tolerantNotFounds: expectResults,
		startTime:         time.Now(),
	}
}

// checkError checks if it has an error should be returned.
// node of the cluster may return not found error,
// ignoreResponse=true symbols that the response should be ignored
func (c *metricTaskContext) checkError(errMsg string) (ignoreResponse bool, err error) {
	if errMsg == "" {
		return false, nil
	}
	// real error
	if !strings.Contains(errMsg, "not found") {
		goto ReturnError
	}
	c.tolerantNotFounds--
	// not found, but there may be still more responses not reached
	if c.tolerantNotFounds > 0 {
		return true, nil
	}
	// fallthrough, all node returns not found errors
ReturnError:
	return true, errors.New(errMsg)
}

func (c *metricTaskContext) WriteResponse(resp *protoCommonV1.TaskResponse, fromNode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.expectResults--

	// preventing close channel twice
	if c.closed {
		return
	}
	defer func() {
		if c.expectResults <= 0 {
			close(c.eventCh)
			c.closed = true
		}
	}()

	if err := c.handleTaskResponse(resp, fromNode); err != nil {
		select {
		case c.eventCh <- &series.TimeSeriesEvent{Err: err, Stats: c.stats}:
		default:
			// reader gone
		}
		return
	}
	// not done yet
	if c.expectResults > 0 {
		return
	}

	var seriesList series.GroupedIterators
	if c.groupAgg != nil {
		seriesList = c.groupAgg.ResultSet()
	}
	if c.stats != nil {
		c.stats.End = time.Now().UnixNano()
	}
	select {
	case c.eventCh <- &series.TimeSeriesEvent{
		AggregatorSpecs: c.aggregatorSpecs,
		SeriesList:      seriesList,
		Stats:           c.stats,
	}:
	default:
		// reader gone
	}
}

func (c *metricTaskContext) handleStats(resp *protoCommonV1.TaskResponse, fromNode string) {
	if len(resp.Stats) == 0 {
		return
	}
	// if has query stats, need merge task query stats
	if c.stats == nil {
		c.stats = models.NewQueryStats()
		c.stats.Start = c.startTime.UnixNano()
	}
	switch resp.Type {
	// from intermediate node
	case protoCommonV1.TaskType_Intermediate:
		queryStats := models.NewQueryStats()
		_ = encoding.JSONUnmarshal(resp.Stats, queryStats)
		c.stats.MergeBrokerTaskStats(fromNode, queryStats)
	default:
		// from leaf node
		storageStats := &models.LeafNodeStats{}
		_ = encoding.JSONUnmarshal(resp.Stats, storageStats)
		storageStats.NetPayload = int64(len(resp.Stats) + len(resp.Payload))
		c.stats.MergeLeafTaskStats(fromNode, storageStats)
		c.stats.NetPayload += storageStats.NetPayload
	}
}

func (c *metricTaskContext) handleTaskResponse(resp *protoCommonV1.TaskResponse, fromNode string) error {
	c.handleStats(resp, fromNode)

	ignoreResponse, err := c.checkError(resp.ErrMsg)
	if err != nil {
		return err
	}
	// partial not-found errors
	if ignoreResponse {
		return nil
	}

	if c.stats != nil {
		start := time.Now()
		defer func() {
			c.stats.TotalCost = time.Since(start).Nanoseconds()
		}()
	}

	tsList := &protoCommonV1.TimeSeriesList{}
	if err := tsList.Unmarshal(resp.Payload); err != nil {
		return err
	}

	if len(tsList.FieldAggSpecs) == 0 {
		// if it gets empty aggregator spec(empty response), need ignore response.
		// if not ignore, will build empty group aggregator, and cannot aggregate real response data.
		return nil
	}

	for _, spec := range tsList.FieldAggSpecs {
		c.aggregatorSpecs[spec.FieldName] = spec
	}

	if c.groupAgg == nil {
		AggregatorSpecs := make(aggregation.AggregatorSpecs, len(tsList.FieldAggSpecs))
		for idx, aggSpec := range tsList.FieldAggSpecs {
			AggregatorSpecs[idx] = aggregation.NewAggregatorSpec(
				field.Name(aggSpec.FieldName),
				field.Type(aggSpec.FieldType),
			)
			for _, funcType := range aggSpec.FuncTypeList {
				AggregatorSpecs[idx].AddFunctionType(function.FuncType(funcType))
			}
		}
		// interval ratio is 1 when do merge result.
		c.groupAgg = newGroupingAgg(
			c.stmtQuery.Interval,
			1,
			c.stmtQuery.TimeRange,
			AggregatorSpecs,
		)
	}

	for _, ts := range tsList.TimeSeriesList {
		// if no field data, ignore this response
		if len(ts.Fields) == 0 {
			return nil
		}
		fields := make(map[field.Name][]byte)
		for k, v := range ts.Fields {
			fields[field.Name(k)] = v
		}
		c.groupAgg.Aggregate(series.NewGroupedIterator(ts.Tags, fields))
	}

	return nil
}

// metaDataTaskContext represents the task context for tacking task execution state
type metaDataTaskContext struct {
	baseTaskContext

	taskResponseCh chan<- *protoCommonV1.TaskResponse
}

// metricTaskContext creates the task context based on params
func newMetaDataTaskContext(
	ctx context.Context,
	taskID string,
	taskType TaskType,
	parentTaskID string,
	parentNode string,
	expectResults int32,
	taskResponseCh chan<- *protoCommonV1.TaskResponse,
) TaskContext {
	return &metaDataTaskContext{
		baseTaskContext: baseTaskContext{
			ctx:           ctx,
			taskID:        taskID,
			taskType:      taskType,
			parentTaskID:  parentTaskID,
			parentNode:    parentNode,
			expectResults: expectResults,
			closed:        false,
			createTime:    fasttime.UnixMilliseconds(),
		},
		taskResponseCh: taskResponseCh,
	}
}

func (c *metaDataTaskContext) WriteResponse(resp *protoCommonV1.TaskResponse, _ string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.expectResults--

	// preventing close channel twice
	if c.closed {
		return
	}
	select {
	case c.taskResponseCh <- resp:
		if c.expectResults <= 0 {
			close(c.taskResponseCh)
			c.closed = true
		}
	default:
		// has been closed, just drop the data
	}
}
