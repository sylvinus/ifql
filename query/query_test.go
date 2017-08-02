package query_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/query"
)

var ignoreUnexportedQuerySpec = cmpopts.IgnoreUnexported(query.QuerySpec{})

func TestQuery_JSON(t *testing.T) {
	srcData := []byte(`
{
	"operations":[
		{
			"id": "select",
			"kind": "select",
			"spec": {
				"database":"mydb"
			}
		},
		{
			"id": "range",
			"kind": "range",
			"spec": {
				"start": "-4h",
				"stop": "now"
			}
		},
		{
			"id": "clear",
			"kind": "clear"
		}
	],
	"edges":[
		{"parent":"select","child":"range"},
		{"parent":"range","child":"clear"}
	]
}
	`)

	// Ensure we can properly unmarshal a query
	gotQ := query.QuerySpec{}
	if err := json.Unmarshal(srcData, &gotQ); err != nil {
		t.Fatal(err)
	}
	expQ := query.QuerySpec{
		Operations: []*query.Operation{
			{
				OperationID: "select",
				Spec: &query.SelectOpSpec{
					Database: "mydb",
				},
			},
			{
				OperationID: "range",
				Spec: &query.RangeOpSpec{
					Start: query.Time{
						Relative: -4 * time.Hour,
					},
					Stop: query.Time{},
				},
			},
			{
				OperationID: "clear",
				Spec:        &query.ClearOpSpec{},
			},
		},
		Edges: []query.Edge{
			{Parent: "select", Child: "range"},
			{Parent: "range", Child: "clear"},
		},
	}
	if !cmp.Equal(gotQ, expQ, ignoreUnexportedQuerySpec) {
		t.Errorf("unexpected query:\n%s", cmp.Diff(gotQ, expQ))
	}

	// Ensure we can properly marshal a query
	data, err := json.Marshal(expQ)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &gotQ); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(gotQ, expQ, ignoreUnexportedQuerySpec) {
		t.Errorf("unexpected query after marshalling:\n%s", cmp.Diff(gotQ, expQ))
	}
}

func TestQuery_Walk(t *testing.T) {
	testCases := []struct {
		query     *query.QuerySpec
		walkOrder []query.OperationID
		err       error
	}{
		{
			query: &query.QuerySpec{},
			err:   errors.New("query has no root nodes"),
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
				},
			},
			err: errors.New("edge references unknown child operation \"c\""),
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "b"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "b"},
				},
			},
			err: errors.New("found duplicate operation ID \"b\""),
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
					{OperationID: "d"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
					{Parent: "d", Child: "b"},
				},
			},
			err: errors.New("found cycle in query"),
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
					{OperationID: "d"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []query.OperationID{
				"a", "b", "c", "d",
			},
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
					{OperationID: "d"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "b"},
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []query.OperationID{
				"a", "c", "b", "d",
			},
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
					{OperationID: "d"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "c"},
					{Parent: "c", Child: "d"},
				},
			},
			walkOrder: []query.OperationID{
				"b", "a", "c", "d",
			},
		},
		{
			query: &query.QuerySpec{
				Operations: []*query.Operation{
					{OperationID: "a"},
					{OperationID: "b"},
					{OperationID: "c"},
					{OperationID: "d"},
				},
				Edges: []query.Edge{
					{Parent: "a", Child: "c"},
					{Parent: "b", Child: "d"},
				},
			},
			walkOrder: []query.OperationID{
				"b", "d", "a", "c",
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var gotOrder []query.OperationID
			err := tc.query.Walk(func(o *query.Operation) error {
				gotOrder = append(gotOrder, o.OperationID)
				return nil
			})
			if tc.err == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error: %q", tc.err)
				} else if got, exp := err.Error(), tc.err.Error(); got != exp {
					t.Fatalf("unexpected errors: got %q exp %q", got, exp)
				}
			}

			if !cmp.Equal(gotOrder, tc.walkOrder) {
				t.Fatalf("unexpected walk order:\n%s", cmp.Diff(gotOrder, tc.walkOrder))
			}
		})
	}
}
