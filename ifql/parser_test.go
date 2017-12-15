package ifql

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ast/asttest"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "from",
			raw:  `from()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
						},
					},
				},
			},
		},
		{
			name: "identifier with number",
			raw:  `tan2()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "tan2",
							},
						},
					},
				},
			},
		},
		{
			name: "declare variable as an int",
			raw:  `var howdy = 1`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
				},
			},
		},
		{
			name: "declare variable as a float",
			raw:  `var howdy = 1.1`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.FloatLiteral{Value: 1.1},
						}},
					},
				},
			},
		},
		{
			name: "declare variable as an array",
			raw:  `var howdy = [1, 2, 3, 4]`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{Name: "howdy"},
							Init: &ast.ArrayExpression{
								Elements: []ast.Expression{
									&ast.IntegerLiteral{Value: 1},
									&ast.IntegerLiteral{Value: 2},
									&ast.IntegerLiteral{Value: 3},
									&ast.IntegerLiteral{Value: 4},
								},
							},
						}},
					},
				},
			},
		},
		{
			name: "use variable to declare something",
			raw: `var howdy = 1
			from()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "howdy"},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
						},
					},
				},
			},
		},
		{
			name: "variable is from statement",
			raw: `var howdy = from()
			howdy.count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "howdy",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "howdy",
								},
								Property: &ast.Identifier{
									Name: "count",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "two variables for two froms",
			raw: `var howdy = from()
			var doody = from()
			howdy.count()
			doody.sum()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "howdy",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "doody",
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{
									Name: "from",
								},
							},
						}},
					},

					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "howdy",
								},
								Property: &ast.Identifier{
									Name: "count",
								},
							},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "doody",
								},
								Property: &ast.Identifier{
									Name: "sum",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "from with database",
			raw:  `from(db:"telegraf")`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{
								Name: "from",
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{
												Name: "db",
											},
											Value: &ast.StringLiteral{
												Value: "telegraf",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "map member expressions",
			raw: `var m = {key1: 1, key2:"value2"}
			m.key1
			m["key2"]
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "m",
							},
							Init: &ast.ObjectExpression{
								Properties: []*ast.Property{
									{
										Key:   &ast.Identifier{Name: "key1"},
										Value: &ast.IntegerLiteral{Value: 1},
									},
									{
										Key:   &ast.Identifier{Name: "key2"},
										Value: &ast.StringLiteral{Value: "value2"},
									},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "m"},
							Property: &ast.Identifier{Name: "key1"},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "m"},
							Property: &ast.StringLiteral{Value: "key2"},
						},
					},
				},
			},
		},
		{
			name: "var as binary expression of other vars",
			raw: `var a = 1
            var b = 2
            var c = a + b
            var d = a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 1},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.IntegerLiteral{Value: 2},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.Identifier{Name: "b"},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "d",
							},
							Init: &ast.Identifier{Name: "a"},
						}},
					},
				},
			},
		},
		{
			name: "var as unary expression of other vars",
			raw: `var a = 5
            var c = -a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 5},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.UnaryExpression{
								Operator: ast.SubtractionOperator,
								Argument: &ast.Identifier{Name: "a"},
							},
						}},
					},
				},
			},
		},
		{
			name: "var as both binary and unary expressions",
			raw: `var a = 5
            var c = 10 * -a`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.IntegerLiteral{Value: 5},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "c",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.MultiplicationOperator,
								Left:     &ast.IntegerLiteral{Value: 10},
								Right: &ast.UnaryExpression{
									Operator: ast.SubtractionOperator,
									Argument: &ast.Identifier{Name: "a"},
								},
							},
						}},
					},
				},
			},
		},
		{
			name: "unary expressions within logical expression",
			raw: `var a = 5.0
            10.0 * -a == -0.5 or a == 6.0`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.FloatLiteral{Value: 5},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.BinaryExpression{
									Operator: ast.MultiplicationOperator,
									Left:     &ast.FloatLiteral{Value: 10},
									Right: &ast.UnaryExpression{
										Operator: ast.SubtractionOperator,
										Argument: &ast.Identifier{Name: "a"},
									},
								},
								Right: &ast.UnaryExpression{
									Operator: ast.SubtractionOperator,
									Argument: &ast.FloatLiteral{Value: 0.5},
								},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.FloatLiteral{Value: 6},
							},
						},
					},
				},
			},
		},
		{
			name: "expressions with function calls",
			raw:  `var a = foo() == 10`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "foo"},
								},
								Right: &ast.IntegerLiteral{Value: 10},
							},
						}},
					},
				},
			},
		},
		{
			name: "mix unary logical and binary expressions",
			raw: `
            not (f() == 6.0 * x) or fail()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.UnaryExpression{
								Operator: ast.NotOperator,
								Argument: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "f"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.MultiplicationOperator,
										Left:     &ast.FloatLiteral{Value: 6},
										Right:    &ast.Identifier{Name: "x"},
									},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "fail"},
							},
						},
					},
				},
			},
		},
		{
			name: "mix unary logical and binary expressions with extra parens",
			raw: `
            (not (f() == 6.0 * x) or fail())`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.UnaryExpression{
								Operator: ast.NotOperator,
								Argument: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "f"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.MultiplicationOperator,
										Left:     &ast.FloatLiteral{Value: 6},
										Right:    &ast.Identifier{Name: "x"},
									},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "fail"},
							},
						},
					},
				},
			},
		},
		{
			name: "arrow function called",
			raw: `var plusOne = (r) => r + 1
			plusOne(r:5)
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "plusOne",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "r"},
									Right:    &ast.IntegerLiteral{Value: 1},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plusOne"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{
												Name: "r",
											},
											Value: &ast.IntegerLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "arrow function return map",
			raw:  `var toMap = (r) =>({r:r})`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "toMap",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.ObjectExpression{
									Properties: []*ast.Property{{
										Key:   &ast.Identifier{Name: "r"},
										Value: &ast.Identifier{Name: "r"},
									}},
								},
							},
						}},
					},
				},
			},
		},
		{
			name: "arrow function called in binary expression",
			raw: `
            var plusOne = (r) => r + 1
            plusOne(r:5) == 6 or die()
			`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "plusOne",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "r"},
									Right:    &ast.IntegerLiteral{Value: 1},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: ast.OrOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "plusOne"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{
														Name: "r",
													},
													Value: &ast.IntegerLiteral{
														Value: 5,
													},
												},
											},
										},
									},
								},
								Right: &ast.IntegerLiteral{Value: 6},
							},
							Right: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "die"},
							},
						},
					},
				},
			},
		},
		{
			name: "arrow function as single expression",
			raw:  `var f = (r) => r["_measurement"] == "cpu"`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "f",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "r"},
										Property: &ast.StringLiteral{Value: "_measurement"},
									},
									Right: &ast.StringLiteral{Value: "cpu"},
								},
							},
						}},
					},
				},
			},
		},
		{
			name: "arrow function as block",
			raw: `var f = (r) => { 
                var m = r["_measurement"]
                return m == "cpu"
            }`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "f",
							},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{{Name: "r"}},
								Body: &ast.BlockStatement{
									Body: []ast.Statement{
										&ast.VariableDeclaration{
											Declarations: []*ast.VariableDeclarator{{
												ID: &ast.Identifier{
													Name: "m",
												},
												Init: &ast.MemberExpression{
													Object:   &ast.Identifier{Name: "r"},
													Property: &ast.StringLiteral{Value: "_measurement"},
												},
											}},
										},
										&ast.ReturnStatement{
											Argument: &ast.BinaryExpression{
												Operator: ast.EqualOperator,
												Left:     &ast.Identifier{Name: "m"},
												Right:    &ast.StringLiteral{Value: "cpu"},
											},
										},
									},
								},
							},
						}},
					},
				},
			},
		},
		{
			name: "from with filter with no parens",
			raw:  `from(db:"telegraf").filter(fn: (r) => r["other"]=="mem" and r["this"]=="that" or r["these"]!="those")`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Property: &ast.Identifier{Name: "filter"},
								Object: &ast.CallExpression{
									Callee: &ast.Identifier{
										Name: "from",
									},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "db"},
													Value: &ast.StringLiteral{Value: "telegraf"},
												},
											},
										},
									},
								},
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Identifier{{Name: "r"}},
												Body: &ast.LogicalExpression{
													Operator: ast.OrOperator,
													Left: &ast.LogicalExpression{
														Operator: ast.AndOperator,
														Left: &ast.BinaryExpression{
															Operator: ast.EqualOperator,
															Left: &ast.MemberExpression{
																Object:   &ast.Identifier{Name: "r"},
																Property: &ast.StringLiteral{Value: "other"},
															},
															Right: &ast.StringLiteral{Value: "mem"},
														},
														Right: &ast.BinaryExpression{
															Operator: ast.EqualOperator,
															Left: &ast.MemberExpression{
																Object:   &ast.Identifier{Name: "r"},
																Property: &ast.StringLiteral{Value: "this"},
															},
															Right: &ast.StringLiteral{Value: "that"},
														},
													},
													Right: &ast.BinaryExpression{
														Operator: ast.NotEqualOperator,
														Left: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "r"},
															Property: &ast.StringLiteral{Value: "these"},
														},
														Right: &ast.StringLiteral{Value: "those"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "from with range",
			raw:  `from(db:"telegraf").range(start:-1h, end:10m)`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "from"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "db"},
													Value: &ast.StringLiteral{Value: "telegraf"},
												},
											},
										},
									},
								},
								Property: &ast.Identifier{Name: "range"},
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "start"},
											Value: &ast.UnaryExpression{
												Operator: ast.SubtractionOperator,
												Argument: &ast.DurationLiteral{Value: time.Hour},
											},
										},
										{
											Key:   &ast.Identifier{Name: "end"},
											Value: &ast.DurationLiteral{Value: 10 * time.Minute},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "from with limit",
			raw:  `from(db:"telegraf").limit(limit:100, offset:10)`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "from"},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "db"},
													Value: &ast.StringLiteral{Value: "telegraf"},
												},
											},
										},
									},
								},
								Property: &ast.Identifier{Name: "limit"},
							},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key:   &ast.Identifier{Name: "limit"},
											Value: &ast.IntegerLiteral{Value: 100},
										},
										{
											Key:   &ast.Identifier{Name: "offset"},
											Value: &ast.IntegerLiteral{Value: 10},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "from with range and count",
			raw: `from(db:"mydb")
						.range(start:-4h, stop:-2h)
						.count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object: &ast.CallExpression{
											Callee: &ast.Identifier{Name: "from"},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key:   &ast.Identifier{Name: "db"},
															Value: &ast.StringLiteral{Value: "mydb"},
														},
													},
												},
											},
										},
										Property: &ast.Identifier{Name: "range"},
									},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{Value: 4 * time.Hour},
													},
												},
												{
													Key: &ast.Identifier{Name: "stop"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{Value: 2 * time.Hour},
													},
												},
											},
										},
									},
								},
								Property: &ast.Identifier{Name: "count"},
							},
							Arguments: nil,
						},
					},
				},
			},
		},
		{
			name: "from with range, limit and count",
			raw: `from(db:"mydb")
						.range(start:-4h, stop:-2h)
						.limit(limit:10)
						.count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object: &ast.CallExpression{
											Callee: &ast.MemberExpression{
												Object: &ast.CallExpression{
													Callee: &ast.Identifier{Name: "from"},
													Arguments: []ast.Expression{
														&ast.ObjectExpression{
															Properties: []*ast.Property{
																{
																	Key:   &ast.Identifier{Name: "db"},
																	Value: &ast.StringLiteral{Value: "mydb"},
																},
															},
														},
													},
												},
												Property: &ast.Identifier{Name: "range"},
											},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key: &ast.Identifier{Name: "start"},
															Value: &ast.UnaryExpression{
																Operator: ast.SubtractionOperator,
																Argument: &ast.DurationLiteral{Value: 4 * time.Hour},
															},
														},
														{
															Key: &ast.Identifier{Name: "stop"},
															Value: &ast.UnaryExpression{
																Operator: ast.SubtractionOperator,
																Argument: &ast.DurationLiteral{Value: 2 * time.Hour},
															},
														},
													},
												},
											},
										},
										Property: &ast.Identifier{Name: "limit"},
									},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key:   &ast.Identifier{Name: "limit"},
													Value: &ast.IntegerLiteral{Value: 10},
												},
											},
										},
									},
								},
								Property: &ast.Identifier{Name: "count"},
							},
							Arguments: nil,
						},
					},
				},
			},
		},
		{
			name: "from with filter, range and count",
			raw: `from(db:"mydb")
                        .filter(fn: (r) => r["_field"] == 10.1)
                        .range(start:-4h, stop:-2h)
                        .count()`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object: &ast.CallExpression{
											Callee: &ast.MemberExpression{
												Object: &ast.CallExpression{
													Callee: &ast.Identifier{Name: "from"},
													Arguments: []ast.Expression{
														&ast.ObjectExpression{
															Properties: []*ast.Property{
																{
																	Key:   &ast.Identifier{Name: "db"},
																	Value: &ast.StringLiteral{Value: "mydb"},
																},
															},
														},
													},
												},
												Property: &ast.Identifier{Name: "filter"},
											},
											Arguments: []ast.Expression{
												&ast.ObjectExpression{
													Properties: []*ast.Property{
														{
															Key: &ast.Identifier{Name: "fn"},
															Value: &ast.ArrowFunctionExpression{
																Params: []*ast.Identifier{{Name: "r"}},
																Body: &ast.BinaryExpression{
																	Operator: ast.EqualOperator,
																	Left: &ast.MemberExpression{
																		Object:   &ast.Identifier{Name: "r"},
																		Property: &ast.StringLiteral{Value: "_field"},
																	},
																	Right: &ast.FloatLiteral{Value: 10.1},
																},
															},
														},
													},
												},
											},
										},
										Property: &ast.Identifier{Name: "range"},
									},
									Arguments: []ast.Expression{
										&ast.ObjectExpression{
											Properties: []*ast.Property{
												{
													Key: &ast.Identifier{Name: "start"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{Value: 4 * time.Hour},
													},
												},
												{
													Key: &ast.Identifier{Name: "stop"},
													Value: &ast.UnaryExpression{
														Operator: ast.SubtractionOperator,
														Argument: &ast.DurationLiteral{Value: 2 * time.Hour},
													},
												},
											},
										},
									},
								},
								Property: &ast.Identifier{Name: "count"},
							},
							Arguments: nil,
						},
					},
				},
			},
		},
		{
			name: "from with join",
			raw: `
var a = from(db:"dbA").range(start:-1h)
var b = from(db:"dbB").range(start:-1h)
join(tables:[a,b], on:["host"], fn: (a,b) => a["_field"] + b["_field"])`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "db"},
														Value: &ast.StringLiteral{Value: "dbA"},
													},
												},
											},
										},
									},
									Property: &ast.Identifier{Name: "range"},
								},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key: &ast.Identifier{Name: "start"},
												Value: &ast.UnaryExpression{
													Operator: ast.SubtractionOperator,
													Argument: &ast.DurationLiteral{Value: 1 * time.Hour},
												},
											},
										},
									},
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object: &ast.CallExpression{
										Callee: &ast.Identifier{Name: "from"},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key:   &ast.Identifier{Name: "db"},
														Value: &ast.StringLiteral{Value: "dbB"},
													},
												},
											},
										},
									},
									Property: &ast.Identifier{Name: "range"},
								},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key: &ast.Identifier{Name: "start"},
												Value: &ast.UnaryExpression{
													Operator: ast.SubtractionOperator,
													Argument: &ast.DurationLiteral{Value: 1 * time.Hour},
												},
											},
										},
									},
								},
							}},
						},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "join"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "tables"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.Identifier{Name: "a"},
													&ast.Identifier{Name: "b"},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "on"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{&ast.StringLiteral{Value: "host"}},
											},
										},
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Identifier{
													{Name: "a"},
													{Name: "b"},
												},
												Body: &ast.BinaryExpression{
													Operator: ast.AdditionOperator,
													Left: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "a"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
													Right: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "b"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "from with join with complex expression",
			raw: `var a = from(db:"ifql").filter(fn: (r) => r["_measurement"] == "a").range(start:-1h)
			var b = from(db:"ifql").filter(fn: (r) => r["_measurement"] == "b").range(start:-1h)
			join(tables:[a,b], on:["t1"], fn: (a,b) => (a["_field"] - b["_field"]) / b["_field"])`,
			want: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "a",
							},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object: &ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object: &ast.CallExpression{
												Callee: &ast.Identifier{Name: "from"},
												Arguments: []ast.Expression{
													&ast.ObjectExpression{
														Properties: []*ast.Property{
															{
																Key:   &ast.Identifier{Name: "db"},
																Value: &ast.StringLiteral{Value: "ifql"},
															},
														},
													},
												},
											},
											Property: &ast.Identifier{Name: "filter"},
										},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.ArrowFunctionExpression{
															Params: []*ast.Identifier{{Name: "r"}},
															Body: &ast.BinaryExpression{
																Operator: ast.EqualOperator,
																Left: &ast.MemberExpression{
																	Object:   &ast.Identifier{Name: "r"},
																	Property: &ast.StringLiteral{Value: "_measurement"},
																},
																Right: &ast.StringLiteral{Value: "a"},
															},
														},
													},
												},
											},
										},
									},
									Property: &ast.Identifier{Name: "range"},
								},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key: &ast.Identifier{Name: "start"},
												Value: &ast.UnaryExpression{
													Operator: ast.SubtractionOperator,
													Argument: &ast.DurationLiteral{Value: 1 * time.Hour},
												},
											},
										},
									},
								},
							},
						}},
					},
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{
								Name: "b",
							},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object: &ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object: &ast.CallExpression{
												Callee: &ast.Identifier{Name: "from"},
												Arguments: []ast.Expression{
													&ast.ObjectExpression{
														Properties: []*ast.Property{
															{
																Key:   &ast.Identifier{Name: "db"},
																Value: &ast.StringLiteral{Value: "ifql"},
															},
														},
													},
												},
											},
											Property: &ast.Identifier{Name: "filter"},
										},
										Arguments: []ast.Expression{
											&ast.ObjectExpression{
												Properties: []*ast.Property{
													{
														Key: &ast.Identifier{Name: "fn"},
														Value: &ast.ArrowFunctionExpression{
															Params: []*ast.Identifier{{Name: "r"}},
															Body: &ast.BinaryExpression{
																Operator: ast.EqualOperator,
																Left: &ast.MemberExpression{
																	Object:   &ast.Identifier{Name: "r"},
																	Property: &ast.StringLiteral{Value: "_measurement"},
																},
																Right: &ast.StringLiteral{Value: "b"},
															},
														},
													},
												},
											},
										},
									},
									Property: &ast.Identifier{Name: "range"},
								},
								Arguments: []ast.Expression{
									&ast.ObjectExpression{
										Properties: []*ast.Property{
											{
												Key: &ast.Identifier{Name: "start"},
												Value: &ast.UnaryExpression{
													Operator: ast.SubtractionOperator,
													Argument: &ast.DurationLiteral{Value: 1 * time.Hour},
												},
											},
										},
									},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "join"},
							Arguments: []ast.Expression{
								&ast.ObjectExpression{
									Properties: []*ast.Property{
										{
											Key: &ast.Identifier{Name: "tables"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.Identifier{Name: "a"},
													&ast.Identifier{Name: "b"},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "on"},
											Value: &ast.ArrayExpression{
												Elements: []ast.Expression{
													&ast.StringLiteral{
														Value: "t1",
													},
												},
											},
										},
										{
											Key: &ast.Identifier{Name: "fn"},
											Value: &ast.ArrowFunctionExpression{
												Params: []*ast.Identifier{
													{Name: "a"},
													{Name: "b"},
												},
												Body: &ast.BinaryExpression{
													Operator: ast.DivisionOperator,
													Left: &ast.BinaryExpression{
														Operator: ast.SubtractionOperator,
														Left: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "a"},
															Property: &ast.StringLiteral{Value: "_field"},
														},
														Right: &ast.MemberExpression{
															Object:   &ast.Identifier{Name: "b"},
															Property: &ast.StringLiteral{Value: "_field"},
														},
													},
													Right: &ast.MemberExpression{
														Object:   &ast.Identifier{Name: "b"},
														Property: &ast.StringLiteral{Value: "_field"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "parse error extra gibberish",
			raw:     `from(db:"ifql") &^*&H#IUJBN`,
			wantErr: true,
		},
		{
			name:    "parse error extra gibberish and valid content",
			raw:     `from(db:"ifql") &^*&H#IUJBN from(db:"other")`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse("", []byte(tt.raw), Debug(false))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got, asttest.CompareOptions...) {
				t.Errorf("Parse() = -want/+got %s", cmp.Diff(tt.want, got, asttest.CompareOptions...))
			}
		})
	}
}
