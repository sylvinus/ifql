package expression_test

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/expression"
)

func TestNodeMarshal(t *testing.T) {
	testCases := []struct {
		json []byte
		expr expression.Expression
	}{
		{
			json: []byte(`{
				"root":{
					"type":"binary",
					"operator": "==",
					"left":{
						"type": "reference",
						"name": "_measurement",
						"kind": "tag"
					},
					"right":{
						"type": "stringLiteral",
						"value": "abc"
					}
				}
			}`),
			expr: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.EqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "tag",
					},
					Right: &expression.StringLiteralNode{
						Value: "abc",
					},
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type":"unary",
					"operator": "-",
					"node":{
						"type": "floatLiteral",
						"value": 42.0
					}
				}
			}`),
			expr: expression.Expression{
				Root: &expression.UnaryNode{
					Operator: expression.SubtractionOperator,
					Node: &expression.FloatLiteralNode{
						Value: 42,
					},
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type":"stringLiteral",
					"value": "abcxyz"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.StringLiteralNode{
					Value: "abcxyz",
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "integerLiteral",
					"value": "9223372036854775807"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.IntegerLiteralNode{
					Value: 9223372036854775807,
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "booleanLiteral",
					"value": true
				}
			}`),
			expr: expression.Expression{
				Root: &expression.BooleanLiteralNode{
					Value: true,
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "floatLiteral",
					"value": 24.1
				}
			}`),
			expr: expression.Expression{
				Root: &expression.FloatLiteralNode{
					Value: 24.1,
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "durationLiteral",
					"value": "1h3s"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.DurationLiteralNode{
					Value: time.Hour + 3*time.Second,
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "timeLiteral",
					"value": "2017-10-10T10:10:10Z"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.TimeLiteralNode{
					Value: time.Date(2017, 10, 10, 10, 10, 10, 0, time.UTC),
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "regexpLiteral",
					"value": ".*"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.RegexpLiteralNode{
					Value: ".*",
				},
			},
		},
		{
			json: []byte(`{
				"root":{
					"type": "reference",
					"name": "t1",
					"kind": "tag"
				}
			}`),
			expr: expression.Expression{
				Root: &expression.ReferenceNode{
					Name: "t1",
					Kind: "tag",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Test Unmarshal
			gotExp := expression.Expression{}
			if err := json.Unmarshal(tc.json, &gotExp); err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(gotExp, tc.expr) {
				t.Errorf("unexpected expression after unmarshaling -want/+got\n%s", cmp.Diff(tc.expr, gotExp))
			}

			// Test marshal
			data, err := json.Marshal(tc.expr)
			if err != nil {
				t.Fatal(err)
			}
			gotExp = expression.Expression{}
			if err := json.Unmarshal(data, &gotExp); err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(gotExp, tc.expr) {
				t.Errorf("unexpected expression after marshalling -want/+got\n%s", cmp.Diff(tc.expr, gotExp))
			}
		})
	}
}
