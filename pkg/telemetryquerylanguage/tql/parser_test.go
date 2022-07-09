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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/telemetryquerylanguage/tql/tqltest"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		query    string
		expected *ParsedQuery
	}{
		{
			query: `set("foo")`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: tqltest.Strp("foo"),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `met(1.2)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "met",
					Arguments: []Value{
						{
							Float: tqltest.Floatp(1.2),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `fff(12)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "fff",
					Arguments: []Value{
						{
							Int: tqltest.Intp(12),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `set("foo", get(bear.honey))`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: tqltest.Strp("foo"),
						},
						{
							Invocation: &Invocation{
								Function: "get",
								Arguments: []Value{
									{
										Path: &Path{
											Fields: []Field{
												{
													Name: "bear",
												},
												{
													Name: "honey",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `set(foo.attributes["bar"].cat, "dog")`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: tqltest.Strp("dog"),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `set(foo.attributes["bar"].cat, "dog") where name == "fido"`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: tqltest.Strp("dog"),
						},
					},
				},
				Condition: &Condition{
					Left: Value{
						Path: &Path{
							Fields: []Field{
								{
									Name: "name",
								},
							},
						},
					},
					Op: "==",
					Right: Value{
						String: tqltest.Strp("fido"),
					},
				},
			},
		},
		{
			query: `set(foo.attributes["bar"].cat, "dog") where name != "fido"`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: tqltest.Strp("dog"),
						},
					},
				},
				Condition: &Condition{
					Left: Value{
						Path: &Path{
							Fields: []Field{
								{
									Name: "name",
								},
							},
						},
					},
					Op: "!=",
					Right: Value{
						String: tqltest.Strp("fido"),
					},
				},
			},
		},
		{
			query: `set  ( foo.attributes[ "bar"].cat,   "dog")   where name=="fido"`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name: "foo",
									},
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("bar"),
									},
									{
										Name: "cat",
									},
								},
							},
						},
						{
							String: tqltest.Strp("dog"),
						},
					},
				},
				Condition: &Condition{
					Left: Value{
						Path: &Path{
							Fields: []Field{
								{
									Name: "name",
								},
							},
						},
					},
					Op: "==",
					Right: Value{
						String: tqltest.Strp("fido"),
					},
				},
			},
		},
		{
			query: `set("fo\"o")`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							String: tqltest.Strp("fo\"o"),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `convert_gauge_to_sum("cumulative", false)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "convert_gauge_to_sum",
					Arguments: []Value{
						{
							String: tqltest.Strp("cumulative"),
						},
						{
							Bool: (*Boolean)(tqltest.Boolp(false)),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `convert_gauge_to_sum("cumulative", true)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "convert_gauge_to_sum",
					Arguments: []Value{
						{
							String: tqltest.Strp("cumulative"),
						},
						{
							Bool: (*Boolean)(tqltest.Boolp(true)),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `set(attributes["bytes"], 0x0102030405060708)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("bytes"),
									},
								},
							},
						},
						{
							Bytes: (*Bytes)(&[]byte{1, 2, 3, 4, 5, 6, 7, 8}),
						},
					},
				},
				Condition: nil,
			},
		},
		{
			query: `set(attributes["test"], nil)`,
			expected: &ParsedQuery{
				Invocation: Invocation{
					Function: "set",
					Arguments: []Value{
						{
							Path: &Path{
								Fields: []Field{
									{
										Name:   "attributes",
										MapKey: tqltest.Strp("test"),
									},
								},
							},
						},
						{
							IsNil: (*IsNil)(tqltest.Boolp(true)),
						},
					},
				},
				Condition: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			parsed, err := parseQuery(tt.query)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func Test_parse_failure(t *testing.T) {
	tests := []string{
		`set(`,
		`set("foo)`,
		`set(name.)`,
		`("foo")`,
		`set("foo") where name =||= "fido"`,
		`set(span_id, SpanIDWrapper{not a hex string})`,
		`set(span_id, SpanIDWrapper{01})`,
		`set(span_id, SpanIDWrapper{010203040506070809})`,
		`set(trace_id, TraceIDWrapper{not a hex string})`,
		`set(trace_id, TraceIDWrapper{0102030405060708090a0b0c0d0e0f})`,
		`set(trace_id, TraceIDWrapper{0102030405060708090a0b0c0d0e0f1011})`,
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := parseQuery(tt)
			assert.Error(t, err)
		})
	}
}

func testParsePath(val *Path) (GetSetter, error) {
	if val != nil && len(val.Fields) > 0 && val.Fields[0].Name == "name" {
		return &testGetSetter{
			getter: func(ctx TransformContext) interface{} {
				return ctx.GetItem()
			},
			setter: func(ctx TransformContext, val interface{}) {
				ctx.GetItem()
			},
		}, nil
	}
	return nil, fmt.Errorf("bad path %v", val)
}
