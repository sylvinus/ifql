package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestJoinOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"join",
		"kind":"join",
		"spec":{
			"keys":["t1","t2"],
			"expression":{
				"root":{
					"type":"binary",
					"operator": "==",
					"left":{
						"type":"reference",
						"name":"_measurement",
						"kind":"identifier"
					},
					"right":{
						"type":"stringLiteral",
						"value":"abc"
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "join",
		Spec: &functions.JoinOpSpec{
			Keys: []string{"t1", "t2"},
			Expression: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.EqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "identifier",
					},
					Right: &expression.StringLiteralNode{
						Value: "abc",
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
