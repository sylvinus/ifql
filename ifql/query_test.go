package ifql

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/ifql/query"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "select",
			raw:     `select()`,
			wantErr: true,
		},
		{
			name: "select with database",
			raw:  `select(database:"mydb").range(start:-4h stop:-2h).clear()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "clear",
						Spec: &query.ClearOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "clear"},
				},
			},
		},
		{
			name: "select with database",
			raw:  `select(database:"mydb").range(start:-4h stop:-2h).sum()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "sum",
						Spec: &query.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "sum"},
				},
			},
		},
		{
			name: "select with database",
			raw:  `select(database:"mydb").range(start:-4h stop:-2h).count()`,
			want: &query.QuerySpec{
				Operations: []*query.Operation{
					{
						ID: "select",
						Spec: &query.SelectOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range",
						Spec: &query.RangeOpSpec{
							Start: query.Time{
								Relative: -4 * time.Hour,
							},
							Stop: query.Time{
								Relative: -2 * time.Hour,
							},
						},
					},
					{
						ID:   "count",
						Spec: &query.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "select", Child: "range"},
					{Parent: "range", Child: "count"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewQuery(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. NewQuery() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
