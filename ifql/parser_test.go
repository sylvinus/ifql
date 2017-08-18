package ifql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewAST(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "select",
			raw:  `select()`,
			want: &Function{
				Name:     "select",
				Args:     []*FunctionArg{},
				Children: []*Function{},
			},
		},
		{
			name: "select with database",
			raw:  `select(database:"telegraf")`,
			want: &Function{
				Name: "select",
				Args: []*FunctionArg{
					{
						Name: "database",
						Arg:  &StringLiteral{"telegraf"},
					},
				},
				Children: []*Function{},
			},
		},
		{
			name: "select with where with no parens",
			raw:  `select(database:"telegraf").where(exp:{"other"="mem" and "this"="that" or "these"!="those"})`,
			want: &Function{
				Name: "select",
				Args: []*FunctionArg{
					&FunctionArg{
						Name: "database",
						Arg: &StringLiteral{
							String: "telegraf",
						},
					},
				},
				Children: []*Function{
					&Function{
						Name: "where",
						Args: []*FunctionArg{
							&FunctionArg{
								Name: "exp",
								Arg: &WhereExpr{
									Expr: &BinaryExpression{
										Left: &BinaryExpression{
											Left: &BinaryExpression{
												Left: &StringLiteral{
													String: "other",
												},
												Operator: "=",
												Right: &StringLiteral{
													String: "mem",
												},
											},
											Operator: "and",
											Right: &BinaryExpression{
												Left: &StringLiteral{
													String: "this",
												}, Operator: "=",
												Right: &StringLiteral{
													String: "that",
												},
											},
										},
										Operator: "or",
										Right: &BinaryExpression{
											Left: &StringLiteral{
												String: "these",
											},
											Operator: "!=",
											Right: &StringLiteral{
												String: "those"},
										},
									},
								},
							},
						},
						Children: []*Function{},
					},
				},
			},
		},
		/*{
			name: "select with range",
			raw:  `select(database:"telegraf").range(start:-1h, end:10m)`,
		},
		{
			name: "select with range",
			raw:  `select(database:"telegraf").where(exp:{(("other"="mem") and ("this"="that")) or ("this"!="that")}).range(start:-1h, end:10m).window(period:1m).count()`,
		},*/
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, err := Parse("", []byte(tt.raw))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, g) {
				t.Errorf("Parse() = %s", cmp.Diff(tt.want, g))
			}
		})
	}
}
