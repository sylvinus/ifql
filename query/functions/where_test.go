package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestWhereOperation_Marshaling(t *testing.T) {
	//TODO: Implement expression marshalling
	t.Skip()
	data := []byte(`{"id":"where","kind":"where","spec":{}}`)
	op := &query.Operation{
		ID: "where",
		Spec: &functions.WhereOpSpec{
			Expression: &expression.BinaryNode{
				Operator: expression.EqualOperator,
				Left: &expression.ReferenceNode{
					Name: "_measurement",
					Kind: "tag",
				},
				Right: &expression.StringLiteralNode{
					Value: "mem",
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
