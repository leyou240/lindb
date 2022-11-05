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

package concurrent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	"github.com/lindb/lindb/internal/linmetric"
	"github.com/lindb/lindb/metrics"
)

var statistics = metrics.NewConcurrentStatistics("test", linmetric.BrokerRegistry)

func Test_Pool_Submit(t *testing.T) {
	// num. of pool + 1 dispatcher, workers has not been spawned
	pool := NewPool("test", 2, 0, statistics)

	var c atomic.Int32

	finished := make(chan struct{})
	do := func(iterations int) {
		for i := 0; i < iterations; i++ {
			pool.Submit(context.TODO(), NewTask(func() {
				c.Inc()
			}, nil))
		}
		finished <- struct{}{}
	}
	go do(100)
	<-finished
	pool.Stop()
	pool.Stop()
	// reject all task
	go do(100)
	<-finished
	assert.Equal(t, int32(100), c.Load())
}

func TestPool_Submit_PanicTask(t *testing.T) {
	pool := NewPool("test", 0, time.Millisecond*200, statistics)
	var wait sync.WaitGroup
	wait.Add(1)
	pool.Submit(context.TODO(), NewTask(func() {
		panic("err")
	}, func(_ error) {
		wait.Done()
	}))
	wait.Wait()

	wp := pool.(*workerPool)
	assert.Equal(t, float64(1), wp.statistics.WorkersAlive.Get())
	time.Sleep(time.Second)
	assert.Equal(t, float64(0), wp.statistics.WorkersAlive.Get())
	pool.Stop()
}

func TestPool_Submit_Task_Timeout(t *testing.T) {
	pool := NewPool("test", 0, time.Millisecond*100, statistics)
	submit := func() {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*2)
		defer cancel()
		pool.Submit(ctx, NewTask(func() {
			time.Sleep(20 * time.Millisecond)
		}, nil))
	}
	for i := 0; i < 100; i++ {
		submit()
	}
	time.Sleep(time.Second)
}

func TestPool_idle(t *testing.T) {
	p := NewPool("test", 0, time.Millisecond*100, statistics)
	// no worker
	time.Sleep(time.Second)

	p1 := p.(*workerPool)
	p1.statistics.WorkersAlive.Incr()
	p1.readyWorkers <- newWorker(p1)
	ch := make(chan struct{})
	go func() {
		p1.idle()
		time.Sleep(10 * time.Millisecond)
		p1.cancel()
		time.Sleep(10 * time.Millisecond)
		p1.idle()
		ch <- struct{}{}
	}()
	p1.idle()
	<-ch
}
