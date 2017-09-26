package query_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/functions"
)

func TestOperation_JSON(t *testing.T) {
	testCases := map[string]struct {
		json string
		op   *query.Operation
	}{
		"select": {
			json: `{
				"id": "select",
				"kind": "select",
				"spec": {
					"database":"mydb"
				}
			}`,
			op: &query.Operation{
				ID: "select",
				Spec: &functions.SelectOpSpec{
					Database: "mydb",
				},
			},
		},
		"where": {
			json: `{
				"id": "where",
				"kind": "where",
				"spec": {
					"exp": {}
				}
			}`,
			op: &query.Operation{
				ID: "where",
				Spec: &functions.WhereOpSpec{
					Exp: &query.ExpressionSpec{},
				},
			},
		},
		"range": {
			json: `{
				"id": "range",
				"kind": "range",
				"spec": {
					"start": "-4h",
					"stop": "now"
				}
			}`,
			op: &query.Operation{
				ID: "range",
				Spec: &functions.RangeOpSpec{
					Start: query.Time{
						Relative: -4 * time.Hour,
					},
					Stop: query.Time{},
				},
			},
		},
		//"clear": {
		//	json: `{
		//		"id": "clear",
		//		"kind": "clear"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "clear",
		//		Spec: &functions.ClearOpSpec{},
		//	},
		//},
		"window": {
			json: `{
				"id": "window",
				"kind": "window",
				"spec":{
					"every":"1m",
					"period":"1h",
					"start": "2017-08-01T00:00:00Z"
				}
			}`,
			op: &query.Operation{
				ID: "window",
				Spec: &functions.WindowOpSpec{
					Every:  query.Duration(time.Minute),
					Period: query.Duration(time.Hour),
					Start: query.Time{
						Absolute: time.Date(2017, 8, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		//TODO implement full spec unmarshalling for all OpSpecs below
		"merge": {
			json: `{
				"id": "merge",
				"kind": "merge"
			}`,
			op: &query.Operation{
				ID:   "merge",
				Spec: &functions.MergeOpSpec{},
			},
		},
		//"keys": {
		//	json: `{
		//		"id": "keys",
		//		"kind": "keys"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "keys",
		//		Spec: &functions.KeysOpSpec{},
		//	},
		//},
		//"values": {
		//	json: `{
		//		"id": "values",
		//		"kind": "values"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "values",
		//		Spec: &functions.ValuesOpSpec{},
		//	},
		//},
		//"cardinality": {
		//	json: `{
		//		"id": "cardinality",
		//		"kind": "cardinality"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "cardinality",
		//		Spec: &functions.CardinalityOpSpec{},
		//	},
		//},
		"limit": {
			json: `{
				"id": "limit",
				"kind": "limit"
			}`,
			op: &query.Operation{
				ID:   "limit",
				Spec: &functions.LimitOpSpec{},
			},
		},
		//"shift": {
		//	json: `{
		//		"id": "shift",
		//		"kind": "shift"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "shift",
		//		Spec: &functions.ShiftOpSpec{},
		//	},
		//},
		//"interpolate": {
		//	json: `{
		//		"id": "interpolate",
		//		"kind": "interpolate"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "interpolate",
		//		Spec: &functions.InterpolateOpSpec{},
		//	},
		//},
		"join": {
			json: `{
				"id": "join",
				"kind": "join"
			}`,
			op: &query.Operation{
				ID:   "join",
				Spec: &functions.JoinOpSpec{},
			},
		},
		//"union": {
		//	json: `{
		//		"id": "union",
		//		"kind": "union"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "union",
		//		Spec: &functions.UnionOpSpec{},
		//	},
		//},
		//"filter": {
		//	json: `{
		//		"id": "filter",
		//		"kind": "filter"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "filter",
		//		Spec: &functions.FilterOpSpec{},
		//	},
		//},
		//"sort": {
		//	json: `{
		//		"id": "sort",
		//		"kind": "sort"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "sort",
		//		Spec: &functions.SortOpSpec{},
		//	},
		//},
		//"rate": {
		//	json: `{
		//		"id": "rate",
		//		"kind": "rate"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "rate",
		//		Spec: &functions.RateOpSpec{},
		//	},
		//},
		"count": {
			json: `{
				"id": "count",
				"kind": "count"
			}`,
			op: &query.Operation{
				ID:   "count",
				Spec: &functions.CountOpSpec{},
			},
		},
		"sum": {
			json: `{
				"id": "sum",
				"kind": "sum"
			}`,
			op: &query.Operation{
				ID:   "sum",
				Spec: &functions.SumOpSpec{},
			},
		},
		//"mean": {
		//	json: `{
		//		"id": "mean",
		//		"kind": "mean"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "mean",
		//		Spec: &functions.MeanOpSpec{},
		//	},
		//},
		//"percentile": {
		//	json: `{
		//		"id": "percentile",
		//		"kind": "percentile"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "percentile",
		//		Spec: &functions.PercentileOpSpec{},
		//	},
		//},
		//"stddev": {
		//	json: `{
		//		"id": "stddev",
		//		"kind": "stddev"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "stddev",
		//		Spec: &functions.StddevOpSpec{},
		//	},
		//},
		//"min": {
		//	json: `{
		//		"id": "min",
		//		"kind": "min"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "min",
		//		Spec: &functions.MinOpSpec{},
		//	},
		//},
		//"max": {
		//	json: `{
		//		"id": "max",
		//		"kind": "max"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "max",
		//		Spec: &functions.MaxOpSpec{},
		//	},
		//},
		//"top": {
		//	json: `{
		//		"id": "top",
		//		"kind": "top"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "top",
		//		Spec: &functions.TopOpSpec{},
		//	},
		//},
		//"difference": {
		//	json: `{
		//		"id": "difference",
		//		"kind": "difference"
		//	}`,
		//	op: &query.Operation{
		//		ID:   "difference",
		//		Spec: &functions.DifferenceOpSpec{},
		//	},
		//},
	}
	if got, exp := len(testCases), query.NumberOfOperations(); got != exp {
		t.Fatalf("unexpected number of test cases, have %d test cases, there are %d kinds", got, exp)
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// Ensure we can properly unmarshal a spec
			gotO := &query.Operation{}
			if err := json.Unmarshal([]byte(tc.json), &gotO); err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(gotO, tc.op) {
				t.Errorf("unexpected operation -want/+got %s", cmp.Diff(tc.op, gotO))
			}

			// Marshal the spec and ensure we can unmarshal it again.
			data, err := json.Marshal(tc.op)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &gotO); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(gotO, tc.op) {
				t.Errorf("unexpected operation after marshalling -want/+got %s", cmp.Diff(tc.op, gotO))
			}
		})
	}
}
