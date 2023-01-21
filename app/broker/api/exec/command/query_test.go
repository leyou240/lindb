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

package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	depspkg "github.com/lindb/lindb/app/broker/deps"
	"github.com/lindb/lindb/config"
	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/query"
	"github.com/lindb/lindb/sql/stmt"
)

func TestQueryCommand(t *testing.T) {
	defer func() {
		metricDataSearchFn = query.MetricDataSearch
	}()

	metricDataSearchFn = func(_ context.Context, _ *models.ExecuteParam, _ *stmt.Query, _ *query.SearchMgr) (any, error) {
		return nil, nil
	}

	rs, err := QueryCommand(context.Background(), &depspkg.HTTPDeps{
		Node: &models.StatelessNode{},
		BrokerCfg: &config.Broker{
			Query: *config.NewDefaultQuery(),
		},
	}, nil, &stmt.Query{})
	assert.NoError(t, err)
	assert.Nil(t, rs)
}