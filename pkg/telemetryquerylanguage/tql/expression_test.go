// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/telemetryquerylanguage/tql/tqltest"
)

func hello() (ExprFunc, error) {
	return func(ctx TransformContext) interface{} {
		return "world"
	}, nil
}

func Test_newGetter(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		want interface{}
	}{
		{
			name: "string literal",
			val: Value{
				String: tqltest.Strp("str"),
			},
			want: "str",
		},
		{
			name: "float literal",
			val: Value{
				Float: tqltest.Floatp(1.2),
			},
			want: 1.2,
		},
		{
			name: "int literal",
			val: Value{
				Int: tqltest.Intp(12),
			},
			want: int64(12),
		},
		{
			name: "bytes literal",
			val: Value{
				Bytes: (*Bytes)(&[]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			},
			want: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "nil literal",
			val: Value{
				IsNil: (*IsNil)(tqltest.Boolp(true)),
			},
			want: nil,
		},
		{
			name: "bool literal",
			val: Value{
				Bool: (*Boolean)(tqltest.Boolp(true)),
			},
			want: true,
		},
		{
			name: "path expression",
			val: Value{
				Path: &Path{
					Fields: []Field{
						{
							Name: "name",
						},
					},
				},
			},
			want: "bear",
		},
		{
			name: "function call",
			val: Value{
				Invocation: &Invocation{
					Function: "hello",
				},
			},
			want: "world",
		},
	}

	functions := map[string]interface{}{"hello": hello}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewGetter(tt.val, functions, testParsePath)
			assert.NoError(t, err)
			val := reader.Get(tqltest.TestTransformContext{
				Item: tt.want,
			})
			assert.Equal(t, tt.want, val)
		})
	}

	t.Run("empty value", func(t *testing.T) {
		_, err := NewGetter(Value{}, functions, testParsePath)
		assert.Error(t, err)
	})
}

// pathGetSetter is a getSetter which has been resolved using a path expression provided by a user.
type testGetSetter struct {
	getter ExprFunc
	setter func(ctx TransformContext, val interface{})
}

func (path testGetSetter) Get(ctx TransformContext) interface{} {
	return path.getter(ctx)
}

func (path testGetSetter) Set(ctx TransformContext, val interface{}) {
	path.setter(ctx, val)
}
