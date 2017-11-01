package influxql

import (
	"testing"

	"github.com/influxdata/influxql"
)

func mustParseExpr(t *testing.T, s string) influxql.Expr {
	t.Helper()

	if e, err := influxql.ParseExpr(s); err != nil {
		panic("ParseExpr failed:" + err.Error())

	} else {
		return e
	}
}

func TestValidateTimeRangeExpr(t *testing.T) {
	cases := []struct {
		in  influxql.Expr
		err bool
	}{
		{
			in:  mustParseExpr(t, `a = 1 OR (time > 0 AND time < 10)`),
			err: true,
		},
		{
			in: mustParseExpr(t, `a = 1 AND (time > 0 AND time < 10)`),
		},
		{
			in:  mustParseExpr(t, `a = 1 AND (time > 0 OR time < 10)`),
			err: true,
		},
		{
			in: mustParseExpr(t, `a = 1`),
		},
		{
			in:  mustParseExpr(t, `a = 1 OR time > 0`),
			err: true,
		},
		{
			in: mustParseExpr(t, `a = 1 AND (time >= now() - 10 AND time <= now() - 5)`),
		},
		{
			in:  mustParseExpr(t, `(a = 1 AND (time > 0 AND time < 20)) OR (a = 2 AND (time > 40 AND time < 60))`),
			err: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.in.String(), func(t *testing.T) {
			if err := ValidateTimeRangeExpr(tc.in); (err != nil) != tc.err {
				t.Error("expected error but got nil")
			}
		})
	}
}
