package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ifql/pl"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		script  string
		want    *ast.Program
		wantErr bool
	}{
		{
			name:   "simple",
			script: "var a = x",
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: "a"},
								Init: &ast.StringLiteral{Value: "x"},
							},
						},
					},
				},
			},
		},
		{
			name:    "error",
			script:  "var aa",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := pl.Parse(tc.script)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse() error = %v, wantErr: %t", err, tc.wantErr)
			}
			if !cmp.Equal(got, tc.want) {
				t.Errorf("Parse() unexpected -want/+got:\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
