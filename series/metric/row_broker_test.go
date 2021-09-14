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

package metric

import (
	"bytes"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/fasttime"
	"github.com/lindb/lindb/proto/gen/v1/flatMetricsV1"
)

func Test_BrokerBatchRows(t *testing.T) {
	for i := 0; i < 10; i++ {
		brokerRows := NewBrokerBatchRows()
		assertBrokerBatchRows(t, brokerRows)
		brokerRows.Release()
	}
}

func assertBrokerBatchRows(t *testing.T, brokerRows *BrokerBatchRows) {
	now := fasttime.UnixMilliseconds()

	assert.Zero(t, brokerRows.Len())
	for i := 0; i < 1000; i++ {
		i := i
		assert.NoError(t, brokerRows.TryAppend(func(row *BrokerRow) error {
			buildRow(row, now-int64(i)*1000*60)
			return nil
		}))
	}
	assert.Equal(t, 1000, brokerRows.Len())

	// only one shard
	itr := brokerRows.NewShardGroupIterator(1)
	assert.True(t, itr.HasRowsForNextShard())
	shardID, rows := itr.RowsForNextShard()
	assert.Equal(t, models.ShardID(0), shardID)
	assert.Len(t, rows, 1000)
	assert.False(t, itr.HasRowsForNextShard())

	itr = brokerRows.NewShardGroupIterator(10)
	for i := 0; i < 10; i++ {
		assert.True(t, itr.HasRowsForNextShard())
		shardID, rows = itr.RowsForNextShard()
		assert.Equal(t, models.ShardID(i), shardID)
		assert.True(t, len(rows) > 0)
		assert.Len(t, brokerRows.Rows(), 1000)
	}
	assert.False(t, itr.HasRowsForNextShard())

	// eviction
	assert.Equal(t, 999,
		brokerRows.EvictOutOfTimeRange(100, 100))
}

func buildRow(row *BrokerRow, timestamp int64) {
	builder, releaseFunc := NewRowBuilder()
	defer releaseFunc(builder)

	builder.AddMetricName([]byte("test"))
	_ = builder.AddTag([]byte("ts"), []byte(strconv.FormatInt(timestamp, 10)))
	_ = builder.AddSimpleField([]byte("f1"), flatMetricsV1.SimpleFieldTypeDeltaSum, 100)
	builder.AddTimestamp(timestamp)
	_ = builder.BuildTo(row)

}

func Test_BrokerBatchRows_AppendError(t *testing.T) {
	batch := NewBrokerBatchRows()
	defer batch.Release()

	assert.Error(t, batch.TryAppend(func(row *BrokerRow) error {
		return io.ErrShortBuffer
	}))
	assert.Equal(t, 0, batch.Len())
}

func Test_BrokerRow_Writer(t *testing.T) {
	var row BrokerRow
	row.IsOutOfTimeRange = true
	row.buffer = append(row.buffer, []byte{1, 2, 3, 4}...)

	var buf bytes.Buffer
	assert.Equal(t, 0, row.Size())
	n, err := row.WriteTo(&buf)
	assert.Equal(t, 0, n)
	assert.NoError(t, err)

	_ = row.Metric()
	row.IsOutOfTimeRange = false
	assert.Equal(t, 4, row.Size())
	n, err = row.WriteTo(&buf)
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
}