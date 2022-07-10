// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestRingBufferCapacity(t *testing.T) {
	// prepare
	buffer :=  NewRingBuffer(5)

	// test
	traceIDs := []pcommon.TraceID{
		pcommon.NewTraceID([16]byte{1, 2, 3, 4}),
		pcommon.NewTraceID([16]byte{2, 3, 4, 5}),
		pcommon.NewTraceID([16]byte{3, 4, 5, 6}),
		pcommon.NewTraceID([16]byte{4, 5, 6, 7}),
		pcommon.NewTraceID([16]byte{5, 6, 7, 8}),
		pcommon.NewTraceID([16]byte{6, 7, 8, 9}),
	}
	for _, traceID := range traceIDs {
		buffer.Put(traceID)
	}

	// verify
	for i := 5; i > 0; i-- { // last 5 traces
		traceID := traceIDs[i]
		assert.True(t, buffer.Contains(traceID))
	}

	// the first trace should have been evicted
	assert.False(t, buffer.Contains(traceIDs[0]))
}

func TestDeleteFromBuffer(t *testing.T) {
	// prepare
	buffer :=  NewRingBuffer(2)
	traceID := pcommon.NewTraceID([16]byte{1, 2, 3, 4})
	buffer.Put(traceID)

	// test
	deleted := buffer.Delete(traceID)

	// verify
	assert.True(t, deleted)
	assert.False(t, buffer.Contains(traceID))
}

func TestDeleteNonExistingFromBuffer(t *testing.T) {
	// prepare
	buffer :=  NewRingBuffer(2)
	traceID := pcommon.NewTraceID([16]byte{1, 2, 3, 4})

	// test
	deleted := buffer.Delete(traceID)

	// verify
	assert.False(t, deleted)
	assert.False(t, buffer.Contains(traceID))
}
