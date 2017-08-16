package promql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		promql  string
		opts    []Option
		want    string
		wantErr bool
	}{
		{
			name:   "testing comments",
			promql: "# http_requests_total",
			want:   "# http_requests_total",
		},
		{
			name:   "vector",
			promql: "http_requests_total",
			want:   "http_requests_total",
		},
		{
			name:   "vector with label matching",
			promql: `http_requests_total{a="b"}`,
			want:   `http_requests_total{a="b"}`,
		},
		{
			name:   "vector with two labels matching",
			promql: `http_requests_total{a="b", c="d"}`,
			want:   `http_requests_total{a="b", c="d"}`,
		},
		{
			name:   "vector with numeric label matcher",
			promql: `http_requests_total{a=500}`,
			want:   `http_requests_total{a=500}`,
		},
		{
			name:    "invalid operator in label matcher",
			promql:  `http_requests_total{a > 500}`,
			wantErr: true,
		},
		{
			name:    "no metric name",
			promql:  `{}`,
			wantErr: true,
		},
		{
			name:   "vector with multiple regular expressions",
			promql: `foo{a="b", foo!="bar", test=~"test", bar!~"baz"}`,
			want:   `foo{a="b", foo!="bar", test=~"test", bar!~"baz"}`,
		},
		{
			name:   "vector with offset",
			promql: "http_requests_total OFFSET 5m",
			want:   "http_requests_total OFFSET 5m",
		},
		{
			name:   "vector with range",
			promql: "http_requests_total[5y]",
			want:   "http_requests_total[5y]",
		},
		{
			name:   "vector with label matches, range, and offset",
			promql: `test{a="b"}[5y] OFFSET 3d`,
			want:   `test{a="b"}[5y] OFFSET 3d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewQuery(tt.promql, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("NewQuery() = -got/+want %s", cmp.Diff(tt.want, got))
			}
		})
	}
}
