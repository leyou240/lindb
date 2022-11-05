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

package function

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncTypeString(t *testing.T) {
	assert.Equal(t, "sum", Sum.String())
	assert.Equal(t, "min", Min.String())
	assert.Equal(t, "max", Max.String())
	assert.Equal(t, "count", Count.String())
	assert.Equal(t, "avg", Avg.String())
	assert.Equal(t, "last", Last.String())
	assert.Equal(t, "first", First.String())
	assert.Equal(t, "quantile", Quantile.String())
	assert.Equal(t, "stddev", Stddev.String())
	assert.Equal(t, "rate", Rate.String())
	assert.Equal(t, "unknown", Unknown.String())
}

func TestIsSupportOrderBy(t *testing.T) {
	assert.True(t, IsSupportOrderBy(Max))
	assert.False(t, IsSupportOrderBy(Quantile))
	assert.False(t, IsSupportOrderBy(Unknown))
}
