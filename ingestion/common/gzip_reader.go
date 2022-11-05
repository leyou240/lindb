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

package common

import (
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
)

// for testing
var (
	resetReaderFn = resetReader
)

var gzipReaderPool sync.Pool

// GetGzipReader picks a cached reader from the pool
func GetGzipReader(r io.Reader) (*gzip.Reader, error) {
	reader := gzipReaderPool.Get()
	if reader == nil {
		return gzip.NewReader(r)
	}
	gzipReader := reader.(*gzip.Reader)
	if err := resetReaderFn(gzipReader, r); err != nil {
		// illegal reader, put it back
		PutGzipReader(gzipReader)
		return nil, err
	}
	return gzipReader, nil
}

// PutGzipReader puts the gzipReader back to the pool
func PutGzipReader(gzipReader *gzip.Reader) {
	if gzipReader == nil {
		return
	}
	_ = gzipReader.Close()
	gzipReaderPool.Put(gzipReader)
}

// resetReader resets gzip reader.
func resetReader(reader *gzip.Reader, r io.Reader) error {
	return reader.Reset(r)
}
