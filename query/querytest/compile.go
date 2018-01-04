package querytest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/ast/asttest"
	"github.com/influxdata/ifql/query"
)

type NewQueryTestCase struct {
	Name    string
	Raw     string
	Want    *query.QuerySpec
	WantErr bool
}

var opts = append(asttest.CompareOptions, cmp.AllowUnexported(query.QuerySpec{}), cmpopts.IgnoreUnexported(query.QuerySpec{}))

func NewQueryTestHelper(t *testing.T, tc NewQueryTestCase) {
	t.Helper()

	got, err := query.Compile(context.Background(), tc.Raw)
	if (err != nil) != tc.WantErr {
		t.Errorf("ifql.NewQuery() error = %v, wantErr %v", err, tc.WantErr)
		return
	}
	if tc.WantErr {
		return
	}
	if !cmp.Equal(tc.Want, got, opts...) {
		t.Errorf("ifql.NewQuery() = -want/+got %s", cmp.Diff(tc.Want, got, opts...))
	}
}
