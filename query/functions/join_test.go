package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestJoinOperation_Marshaling(t *testing.T) {
	//TODO: Implement expression marshalling
	t.Skip()
	data := []byte(`{"id":"join","kind":"join","spec":{"keys":["t1","t2"]}}`)
	op := &query.Operation{
		ID: "join",
		Spec: &functions.JoinOpSpec{
			Keys: []string{"t1", "t2"},
			Expression: &expression.BinaryNode{
				Operator: expression.AdditionOperator,
				Left: &expression.ReferenceNode{
					Name: "_measurement",
					Kind: "identifier",
				},
				Right: &expression.IntegerLiteralNode{
					Value: 42,
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
