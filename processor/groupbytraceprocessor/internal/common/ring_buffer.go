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

package common // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbytraceprocessor"

import "go.opentelemetry.io/collector/pdata/pcommon"

// RingBuffer keeps an in-memory bounded buffer with the in-flight trace IDs
type RingBuffer struct {
	index     int
	size      int
	ids       []pcommon.TraceID
	idToIndex map[pcommon.TraceID]int // key is traceID, value is the index on the 'ids' slice
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		index:     -1, // the first span to be received will be placed at position '0'
		size:      size,
		ids:       make([]pcommon.TraceID, size),
		idToIndex: make(map[pcommon.TraceID]int),
	}
}

func (r *RingBuffer) Put(traceID pcommon.TraceID) pcommon.TraceID {
	// calculates the item in the ring that we'll store the trace
	r.index = (r.index + 1) % r.size

	// see if the ring has an item already
	evicted := r.ids[r.index]

	if !evicted.IsEmpty() {
		// clear space for the new item
		r.Delete(evicted)
	}

	// place the traceID in memory
	r.ids[r.index] = traceID
	r.idToIndex[traceID] = r.index

	return evicted
}

func (r *RingBuffer) Contains(traceID pcommon.TraceID) bool {
	_, found := r.idToIndex[traceID]
	return found
}

func (r *RingBuffer) Delete(traceID pcommon.TraceID) bool {
	index, found := r.idToIndex[traceID]
	if !found {
		return false
	}

	delete(r.idToIndex, traceID)
	r.ids[index] = pcommon.InvalidTraceID()
	return true
}
