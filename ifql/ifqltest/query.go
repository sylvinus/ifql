package ifqltest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
)

type NewQueryTestCase struct {
	Name    string
	Raw     string
	Want    *query.QuerySpec
	WantErr bool
}

func NewQueryTestHelper(t *testing.T, tc NewQueryTestCase) {
	t.Helper()

	got, err := ifql.NewQuery(tc.Raw, ifql.Debug(false))
	if (err != nil) != tc.WantErr {
		t.Errorf("ifql.NewQuery() error = %v, wantErr %v", err, tc.WantErr)
		return
	}
	if tc.WantErr {
		return
	}
	opts := []cmp.Option{cmp.AllowUnexported(query.QuerySpec{}), cmpopts.IgnoreUnexported(query.QuerySpec{})}
	if !cmp.Equal(tc.Want, got, opts...) {
		t.Errorf("ifql.NewQuery() = -want/+got %s", cmp.Diff(tc.Want, got, opts...))
	}
}
