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

package monitoring

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/lindb/lindb/internal/mock"
	"github.com/lindb/lindb/pkg/fileutil"
)

func TestLoggerAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	path := "."
	logFile := filepath.Join(path, "1.log")
	f, err := os.Create(logFile)
	assert.NoError(t, err)
	defer func() {
		ctrl.Finish()
		readDirFn = os.ReadDir
		_ = f.Close()
		_ = fileutil.RemoveFile(logFile)
	}()

	api := NewLoggerAPI(path)
	r := gin.New()
	api.Register(r)
	resp := mock.DoRequest(t, r, http.MethodGet, LogListPath, "")
	assert.Equal(t, http.StatusOK, resp.Code)

	readDirFn = func(dirname string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("err")
	}
	resp = mock.DoRequest(t, r, http.MethodGet, LogListPath, "")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	readDirFn = os.ReadDir
	resp = mock.DoRequest(t, r, http.MethodGet, LogViewPath, "")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// file not exist
	resp = mock.DoRequest(t, r, http.MethodGet, LogViewPath+"?file=log_handler.go", "")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	// ok
	resp = mock.DoRequest(t, r, http.MethodGet, LogViewPath+"?file=log_handle.go", "")
	assert.Equal(t, http.StatusOK, resp.Code)
}
