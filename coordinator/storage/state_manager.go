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

package storage

import (
	"path/filepath"
	"sync"

	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/encoding"
	"github.com/lindb/lindb/pkg/logger"
	"github.com/lindb/lindb/tsdb"
)

type StateManager interface {
	// OnNodeStartup triggers when node online.
	OnNodeStartup(key string, data []byte)
	// OnNodeFailure trigger when node offline.
	OnNodeFailure(key string)
	OnShardAssignmentChange(key string, data []byte)
	OnDatabaseDelete(key string)
}

type stateManager struct {
	engine  tsdb.Engine
	current *models.StatefulNode
	nodes   map[string]models.StatefulNode // storage live nodes

	mutex sync.Mutex

	logger *logger.Logger
}

func NewStateManager(current *models.StatefulNode,
	engine tsdb.Engine) StateManager {
	return &stateManager{
		current: current,
		engine:  engine,
		logger:  logger.GetLogger("storage", "StateManager"),
	}
}

func (m *stateManager) OnShardAssignmentChange(key string, data []byte) {
	m.logger.Info("shard assignment is changed",
		logger.String("key", key),
		logger.String("data", string(data)))
	param := models.DatabaseAssignment{}
	if err := encoding.JSONUnmarshal(data, &param); err != nil {
		return
	}
	if param.ShardAssignment == nil {
		return
	}
	var shardIDs []models.ShardID
	for shardID, replica := range param.ShardAssignment.Shards {
		if replica.Contain(m.current.ID) {
			shardIDs = append(shardIDs, shardID)
		}
	}
	if len(shardIDs) == 0 {
		return
	}
	if err := m.engine.CreateShards(
		param.ShardAssignment.Name,
		param.Option,
		shardIDs...,
	); err != nil {
		return
	}
}

func (m *stateManager) OnDatabaseDelete(key string) {
	panic("implement me")
}

func (m *stateManager) OnNodeStartup(key string, data []byte) {
	m.logger.Info("new node online",
		logger.String("key", key),
		logger.String("data", string(data)))

	node := &models.StatefulNode{}
	if err := encoding.JSONUnmarshal(data, node); err != nil {
		m.logger.Error("new node online but unmarshal error", logger.Error(err))
		return
	}

	_, fileName := filepath.Split(key)
	nodeID := fileName

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.nodes[nodeID] = *node
}

func (m *stateManager) OnNodeFailure(key string) {
	_, fileName := filepath.Split(key)
	nodeID := fileName

	m.logger.Info("node online => offline",
		logger.String("nodeID", nodeID),
		logger.String("key", key))

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.nodes, nodeID)
}