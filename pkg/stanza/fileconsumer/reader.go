// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileconsumer // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

// Reader manages a single file
type Reader struct {
	Fingerprint *Fingerprint
	Offset      int64

	generation     int
	fileInput      *Input
	file           *os.File
	fileAttributes *FileAttributes

	splitter *helper.Splitter

	*zap.SugaredLogger `json:"-"`
}

// NewReader creates a new file reader
func (f *Input) NewReader(file *os.File, fp *Fingerprint, splitter *helper.Splitter) (*Reader, error) {
	path := file.Name()
	attrs, err := resolveFileAttributes(path)
	if err != nil {
		f.Errorf("resolve attributes: %w", err)
	}
	r := &Reader{
		SugaredLogger:  f.SugaredLogger.With("path", path),
		Fingerprint:    fp,
		file:           file,
		fileInput:      f,
		fileAttributes: attrs,
		splitter:       splitter,
	}
	return r, nil
}

// Copy creates a deep copy of a Reader
func (r *Reader) Copy(file *os.File) (*Reader, error) {
	reader, err := r.fileInput.NewReader(file, r.Fingerprint.Copy(), r.splitter)
	if err != nil {
		return nil, err
	}
	reader.Offset = r.Offset
	return reader, nil
}

// InitializeOffset sets the starting offset
func (r *Reader) InitializeOffset(startAtBeginning bool) error {
	if !startAtBeginning {
		info, err := r.file.Stat()
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}
		r.Offset = info.Size()
	}
	return nil
}

// ReadToEnd will read until the end of the file
func (r *Reader) ReadToEnd(ctx context.Context) {
	if _, err := r.file.Seek(r.Offset, 0); err != nil {
		r.Errorw("Failed to seek", zap.Error(err))
		return
	}

	scanner := NewPositionalScanner(r, r.fileInput.MaxLogSize, r.Offset, r.splitter.SplitFunc)

	// Iterate over the tokenized file, emitting entries as we go
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ok := scanner.Scan()
		if !ok {
			if err := scanner.getError(); err != nil {
				r.Errorw("Failed during scan", zap.Error(err))
			}
			break
		}

		token, err := r.splitter.Encoding.Decode(scanner.Bytes())
		if err != nil {
			r.Errorw("decode: %w", zap.Error(err))
		} else {
			r.fileInput.emit(ctx, r.fileAttributes, token)
		}

		r.Offset = scanner.Pos()
	}
}

// Close will close the file
func (r *Reader) Close() {
	if r.file != nil {
		if err := r.file.Close(); err != nil {
			r.Debugw("Problem closing reader", zap.Error(err))
		}
	}
}

// Read from the file and update the fingerprint if necessary
func (r *Reader) Read(dst []byte) (int, error) {
	// Skip if fingerprint is already built
	// or if fingerprint is behind Offset
	if len(r.Fingerprint.FirstBytes) == r.fileInput.fingerprintSize || int(r.Offset) > len(r.Fingerprint.FirstBytes) {
		return r.file.Read(dst)
	}
	n, err := r.file.Read(dst)
	appendCount := min0(n, r.fileInput.fingerprintSize-int(r.Offset))
	// return for n == 0 or r.Offset >= r.fileInput.fingerprintSize
	if appendCount == 0 {
		return n, err
	}

	// for appendCount==0, the following code would add `0` to fingerprint
	r.Fingerprint.FirstBytes = append(r.Fingerprint.FirstBytes[:r.Offset], dst[:appendCount]...)
	return n, err
}

func min0(a, b int) int {
	if a < 0 || b < 0 {
		return 0
	}
	if a < b {
		return a
	}
	return b
}
