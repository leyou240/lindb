package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/check.v1"

	"github.com/lindb/lindb/config"
	"github.com/lindb/lindb/constants"
	"github.com/lindb/lindb/mock"
	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/pathutil"
	"github.com/lindb/lindb/pkg/server"
	"github.com/lindb/lindb/pkg/state"
)

type testStorageRuntimeSuite struct {
	mock.RepoTestSuite
}

func TestStorageRuntime(t *testing.T) {
	check.Suite(&testStorageRuntimeSuite{})
	check.TestingT(t)
}

func (ts *testStorageRuntimeSuite) TestStorageRun(c *check.C) {
	// test normal storage run
	cfg := config.Storage{
		Server: config.Server{
			Port: 9999,
			TTL:  1,
		},
		Engine: config.Engine{Path: "/tmp/storage/data"},
		Coordinator: state.Config{
			Namespace: "/test/storage",
			Endpoints: ts.Cluster.Endpoints,
		},
		Replication: config.Replication{
			Path: "/tmp/storage/replication",
		},
	}
	storage := NewStorageRuntime(cfg)
	err := storage.Run()
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(server.Running, check.Equals, storage.State())
	// wait register success
	time.Sleep(200 * time.Millisecond)

	runtime, _ := storage.(*runtime)
	nodePath := pathutil.GetNodePath(constants.ActiveNodesPath, runtime.node.Indicator())
	nodeBytes, err := runtime.repo.Get(context.TODO(), nodePath)
	if err != nil {
		c.Fatal(err)
	}
	nodeInfo := models.ActiveNode{}
	_ = json.Unmarshal(nodeBytes, &nodeInfo)

	c.Assert(runtime.node, check.Equals, nodeInfo.Node)
	c.Assert("storage", check.Equals, storage.Name())

	_ = storage.Stop()
	c.Assert(server.Terminated, check.Equals, storage.State())
}