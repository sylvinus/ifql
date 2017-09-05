package ifql

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:generate pigeon -o ifql.go ifql.peg

var g = &grammar{
	rules: []*rule{
		{
			name: "Start",
			pos:  position{line: 7, col: 1, offset: 60},
			expr: &actionExpr{
				pos: position{line: 8, col: 5, offset: 70},
				run: (*parser).callonStart1,
				expr: &seqExpr{
					pos: position{line: 8, col: 5, offset: 70},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 8, col: 5, offset: 70},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 8, col: 8, offset: 73},
							label: "program",
							expr: &ruleRefExpr{
								pos:  position{line: 8, col: 16, offset: 81},
								name: "Program",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 8, col: 24, offset: 89},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Program",
			pos:  position{line: 12, col: 1, offset: 127},
			expr: &actionExpr{
				pos: position{line: 13, col: 5, offset: 139},
				run: (*parser).callonProgram1,
				expr: &labeledExpr{
					pos:   position{line: 13, col: 5, offset: 139},
					label: "body",
					expr: &ruleRefExpr{
						pos:  position{line: 13, col: 10, offset: 144},
						name: "SourceElements",
					},
				},
			},
		},
		{
			name: "SourceElements",
			pos:  position{line: 17, col: 1, offset: 210},
			expr: &actionExpr{
				pos: position{line: 18, col: 5, offset: 229},
				run: (*parser).callonSourceElements1,
				expr: &seqExpr{
					pos: position{line: 18, col: 5, offset: 229},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 18, col: 5, offset: 229},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 18, col: 10, offset: 234},
								name: "SourceElement",
							},
						},
						&labeledExpr{
							pos:   position{line: 18, col: 24, offset: 248},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 18, col: 29, offset: 253},
								expr: &seqExpr{
									pos: position{line: 18, col: 30, offset: 254},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 18, col: 30, offset: 254},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 18, col: 33, offset: 257},
											name: "SourceElement",
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
			name: "SourceElement",
			pos:  position{line: 22, col: 1, offset: 316},
			expr: &ruleRefExpr{
				pos:  position{line: 23, col: 5, offset: 334},
				name: "Statement",
			},
		},
		{
			name: "Statement",
			pos:  position{line: 25, col: 1, offset: 345},
			expr: &choiceExpr{
				pos: position{line: 26, col: 5, offset: 359},
				alternatives: []interface{}{
					&labeledExpr{
						pos:   position{line: 26, col: 5, offset: 359},
						label: "varstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 26, col: 13, offset: 367},
							name: "VariableStatement",
						},
					},
					&labeledExpr{
						pos:   position{line: 27, col: 5, offset: 389},
						label: "exprstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 27, col: 14, offset: 398},
							name: "ExpressionStatement",
						},
					},
				},
			},
		},
		{
			name: "VariableStatement",
			pos:  position{line: 29, col: 1, offset: 419},
			expr: &actionExpr{
				pos: position{line: 30, col: 5, offset: 441},
				run: (*parser).callonVariableStatement1,
				expr: &seqExpr{
					pos: position{line: 30, col: 5, offset: 441},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 30, col: 5, offset: 441},
							name: "VarToken",
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 14, offset: 450},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 30, col: 17, offset: 453},
							label: "declarations",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 30, offset: 466},
								name: "VariableDeclarationList",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableDeclarationList",
			pos:  position{line: 34, col: 1, offset: 549},
			expr: &actionExpr{
				pos: position{line: 35, col: 5, offset: 577},
				run: (*parser).callonVariableDeclarationList1,
				expr: &seqExpr{
					pos: position{line: 35, col: 5, offset: 577},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 35, col: 5, offset: 577},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 35, col: 10, offset: 582},
								name: "VariableDeclaration",
							},
						},
						&labeledExpr{
							pos:   position{line: 35, col: 30, offset: 602},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 35, col: 35, offset: 607},
								expr: &seqExpr{
									pos: position{line: 35, col: 36, offset: 608},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 35, col: 36, offset: 608},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 35, col: 39, offset: 611},
											val:        ",",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 35, col: 43, offset: 615},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 35, col: 46, offset: 618},
											name: "VariableDeclaration",
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
			name: "VarToken",
			pos:  position{line: 39, col: 1, offset: 685},
			expr: &litMatcher{
				pos:        position{line: 39, col: 12, offset: 696},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 41, col: 1, offset: 703},
			expr: &actionExpr{
				pos: position{line: 42, col: 5, offset: 727},
				run: (*parser).callonVariableDeclaration1,
				expr: &seqExpr{
					pos: position{line: 42, col: 5, offset: 727},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 42, col: 5, offset: 727},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 8, offset: 730},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 42, col: 19, offset: 741},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 42, col: 22, offset: 744},
							label: "initExpr",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 31, offset: 753},
								name: "Initializer",
							},
						},
					},
				},
			},
		},
		{
			name: "Initializer",
			pos:  position{line: 46, col: 1, offset: 824},
			expr: &actionExpr{
				pos: position{line: 47, col: 5, offset: 840},
				run: (*parser).callonInitializer1,
				expr: &seqExpr{
					pos: position{line: 47, col: 5, offset: 840},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 47, col: 5, offset: 840},
							val:        "=",
							ignoreCase: false,
						},
						&notExpr{
							pos: position{line: 47, col: 9, offset: 844},
							expr: &litMatcher{
								pos:        position{line: 47, col: 10, offset: 845},
								val:        "=",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 47, col: 14, offset: 849},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 47, col: 17, offset: 852},
							label: "expression",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 28, offset: 863},
								name: "VariableExpression",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 52, col: 1, offset: 990},
			expr: &choiceExpr{
				pos: position{line: 53, col: 5, offset: 1013},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 53, col: 5, offset: 1013},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 54, col: 5, offset: 1032},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 5, offset: 1050},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 56, col: 5, offset: 1079},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 5, offset: 1092},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 58, col: 5, offset: 1105},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 59, col: 5, offset: 1116},
						name: "Field",
					},
				},
			},
		},
		{
			name: "ExpressionStatement",
			pos:  position{line: 61, col: 1, offset: 1123},
			expr: &actionExpr{
				pos: position{line: 62, col: 5, offset: 1147},
				run: (*parser).callonExpressionStatement1,
				expr: &labeledExpr{
					pos:   position{line: 62, col: 5, offset: 1147},
					label: "call",
					expr: &ruleRefExpr{
						pos:  position{line: 62, col: 10, offset: 1152},
						name: "CallExpression",
					},
				},
			},
		},
		{
			name: "MemberExpression",
			pos:  position{line: 66, col: 1, offset: 1221},
			expr: &actionExpr{
				pos: position{line: 67, col: 5, offset: 1242},
				run: (*parser).callonMemberExpression1,
				expr: &seqExpr{
					pos: position{line: 67, col: 5, offset: 1242},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 67, col: 5, offset: 1242},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 67, col: 10, offset: 1247},
								name: "Identifier",
							},
						},
						&labeledExpr{
							pos:   position{line: 68, col: 5, offset: 1289},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 68, col: 10, offset: 1294},
								expr: &actionExpr{
									pos: position{line: 69, col: 9, offset: 1304},
									run: (*parser).callonMemberExpression7,
									expr: &seqExpr{
										pos: position{line: 69, col: 9, offset: 1304},
										exprs: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 69, col: 9, offset: 1304},
												name: "__",
											},
											&litMatcher{
												pos:        position{line: 69, col: 12, offset: 1307},
												val:        ".",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 69, col: 16, offset: 1311},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 69, col: 19, offset: 1314},
												label: "property",
												expr: &ruleRefExpr{
													pos:  position{line: 69, col: 28, offset: 1323},
													name: "Identifier",
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
			name: "CallExpression",
			pos:  position{line: 77, col: 1, offset: 1449},
			expr: &actionExpr{
				pos: position{line: 78, col: 5, offset: 1468},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 78, col: 5, offset: 1468},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 78, col: 5, offset: 1468},
							label: "head",
							expr: &actionExpr{
								pos: position{line: 79, col: 7, offset: 1481},
								run: (*parser).callonCallExpression4,
								expr: &seqExpr{
									pos: position{line: 79, col: 7, offset: 1481},
									exprs: []interface{}{
										&labeledExpr{
											pos:   position{line: 79, col: 7, offset: 1481},
											label: "callee",
											expr: &ruleRefExpr{
												pos:  position{line: 79, col: 14, offset: 1488},
												name: "MemberExpression",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 79, col: 31, offset: 1505},
											name: "__",
										},
										&labeledExpr{
											pos:   position{line: 79, col: 34, offset: 1508},
											label: "args",
											expr: &ruleRefExpr{
												pos:  position{line: 79, col: 39, offset: 1513},
												name: "Arguments",
											},
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 83, col: 5, offset: 1596},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 83, col: 10, offset: 1601},
								expr: &choiceExpr{
									pos: position{line: 84, col: 9, offset: 1611},
									alternatives: []interface{}{
										&actionExpr{
											pos: position{line: 84, col: 9, offset: 1611},
											run: (*parser).callonCallExpression14,
											expr: &seqExpr{
												pos: position{line: 84, col: 9, offset: 1611},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 84, col: 9, offset: 1611},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 84, col: 12, offset: 1614},
														label: "args",
														expr: &ruleRefExpr{
															pos:  position{line: 84, col: 17, offset: 1619},
															name: "Arguments",
														},
													},
												},
											},
										},
										&actionExpr{
											pos: position{line: 87, col: 9, offset: 1701},
											run: (*parser).callonCallExpression19,
											expr: &seqExpr{
												pos: position{line: 87, col: 9, offset: 1701},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 87, col: 9, offset: 1701},
														name: "__",
													},
													&litMatcher{
														pos:        position{line: 87, col: 12, offset: 1704},
														val:        ".",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 87, col: 16, offset: 1708},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 87, col: 19, offset: 1711},
														label: "property",
														expr: &ruleRefExpr{
															pos:  position{line: 87, col: 28, offset: 1720},
															name: "Identifier",
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
		},
		{
			name: "Arguments",
			pos:  position{line: 95, col: 1, offset: 1871},
			expr: &actionExpr{
				pos: position{line: 96, col: 5, offset: 1885},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 96, col: 5, offset: 1885},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 96, col: 5, offset: 1885},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 96, col: 9, offset: 1889},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 96, col: 12, offset: 1892},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 96, col: 17, offset: 1897},
								expr: &ruleRefExpr{
									pos:  position{line: 96, col: 18, offset: 1898},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 96, col: 33, offset: 1913},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 96, col: 36, offset: 1916},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 100, col: 1, offset: 1952},
			expr: &actionExpr{
				pos: position{line: 101, col: 5, offset: 1969},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 101, col: 5, offset: 1969},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 101, col: 5, offset: 1969},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 101, col: 11, offset: 1975},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 101, col: 23, offset: 1987},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 101, col: 26, offset: 1990},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 101, col: 31, offset: 1995},
								expr: &ruleRefExpr{
									pos:  position{line: 101, col: 31, offset: 1995},
									name: "FunctionArgsRest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgsRest",
			pos:  position{line: 105, col: 1, offset: 2070},
			expr: &actionExpr{
				pos: position{line: 106, col: 5, offset: 2091},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 106, col: 5, offset: 2091},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 106, col: 5, offset: 2091},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 9, offset: 2095},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 106, col: 13, offset: 2099},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 106, col: 17, offset: 2103},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 110, col: 1, offset: 2146},
			expr: &actionExpr{
				pos: position{line: 111, col: 5, offset: 2162},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 111, col: 5, offset: 2162},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 111, col: 5, offset: 2162},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 111, col: 9, offset: 2166},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 111, col: 20, offset: 2177},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 111, col: 24, offset: 2181},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 111, col: 28, offset: 2185},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 111, col: 31, offset: 2188},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 111, col: 37, offset: 2194},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 115, col: 1, offset: 2270},
			expr: &choiceExpr{
				pos: position{line: 116, col: 5, offset: 2292},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 116, col: 5, offset: 2292},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 117, col: 5, offset: 2306},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 118, col: 5, offset: 2324},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 119, col: 5, offset: 2353},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 120, col: 5, offset: 2366},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 121, col: 5, offset: 2379},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 122, col: 5, offset: 2390},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 124, col: 1, offset: 2402},
			expr: &actionExpr{
				pos: position{line: 125, col: 5, offset: 2416},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 125, col: 5, offset: 2416},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 125, col: 5, offset: 2416},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 125, col: 9, offset: 2420},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 125, col: 12, offset: 2423},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 125, col: 17, offset: 2428},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 125, col: 22, offset: 2433},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 125, col: 26, offset: 2437},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 136, col: 1, offset: 2660},
			expr: &ruleRefExpr{
				pos:  position{line: 137, col: 5, offset: 2669},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 139, col: 1, offset: 2678},
			expr: &actionExpr{
				pos: position{line: 140, col: 5, offset: 2699},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 140, col: 6, offset: 2700},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 140, col: 6, offset: 2700},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 140, col: 14, offset: 2708},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 144, col: 1, offset: 2760},
			expr: &actionExpr{
				pos: position{line: 145, col: 5, offset: 2772},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 145, col: 5, offset: 2772},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 145, col: 5, offset: 2772},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 145, col: 10, offset: 2777},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 145, col: 19, offset: 2786},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 145, col: 24, offset: 2791},
								expr: &seqExpr{
									pos: position{line: 145, col: 26, offset: 2793},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 145, col: 26, offset: 2793},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 145, col: 30, offset: 2797},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 145, col: 47, offset: 2814},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 145, col: 51, offset: 2818},
											name: "Equality",
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
			name: "EqualityOperators",
			pos:  position{line: 149, col: 1, offset: 2897},
			expr: &actionExpr{
				pos: position{line: 150, col: 5, offset: 2919},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 150, col: 6, offset: 2920},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 150, col: 6, offset: 2920},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 150, col: 13, offset: 2927},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 154, col: 1, offset: 2973},
			expr: &actionExpr{
				pos: position{line: 155, col: 5, offset: 2986},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 155, col: 5, offset: 2986},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 155, col: 5, offset: 2986},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 155, col: 10, offset: 2991},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 155, col: 21, offset: 3002},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 155, col: 26, offset: 3007},
								expr: &seqExpr{
									pos: position{line: 155, col: 28, offset: 3009},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 155, col: 28, offset: 3009},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 155, col: 31, offset: 3012},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 155, col: 49, offset: 3030},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 155, col: 52, offset: 3033},
											name: "Relational",
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
			name: "RelationalOperators",
			pos:  position{line: 159, col: 1, offset: 3113},
			expr: &actionExpr{
				pos: position{line: 160, col: 5, offset: 3137},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 160, col: 9, offset: 3141},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 160, col: 9, offset: 3141},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 161, col: 9, offset: 3154},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 162, col: 9, offset: 3166},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 163, col: 9, offset: 3179},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 164, col: 9, offset: 3191},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 165, col: 9, offset: 3213},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 166, col: 9, offset: 3227},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 167, col: 9, offset: 3248},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 172, col: 1, offset: 3306},
			expr: &actionExpr{
				pos: position{line: 173, col: 5, offset: 3321},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 173, col: 5, offset: 3321},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 173, col: 5, offset: 3321},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 173, col: 10, offset: 3326},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 173, col: 19, offset: 3335},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 173, col: 24, offset: 3340},
								expr: &seqExpr{
									pos: position{line: 173, col: 26, offset: 3342},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 173, col: 26, offset: 3342},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 173, col: 29, offset: 3345},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 173, col: 49, offset: 3365},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 173, col: 52, offset: 3368},
											name: "Additive",
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
			name: "AdditiveOperator",
			pos:  position{line: 177, col: 1, offset: 3446},
			expr: &actionExpr{
				pos: position{line: 178, col: 5, offset: 3467},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 178, col: 6, offset: 3468},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 178, col: 6, offset: 3468},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 178, col: 12, offset: 3474},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 182, col: 1, offset: 3522},
			expr: &actionExpr{
				pos: position{line: 183, col: 5, offset: 3535},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 183, col: 5, offset: 3535},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 183, col: 5, offset: 3535},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 183, col: 10, offset: 3540},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 183, col: 25, offset: 3555},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 183, col: 30, offset: 3560},
								expr: &seqExpr{
									pos: position{line: 183, col: 32, offset: 3562},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 183, col: 32, offset: 3562},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 183, col: 35, offset: 3565},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 183, col: 52, offset: 3582},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 183, col: 55, offset: 3585},
											name: "Multiplicative",
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
			name: "MultiplicativeOperator",
			pos:  position{line: 187, col: 1, offset: 3670},
			expr: &actionExpr{
				pos: position{line: 188, col: 5, offset: 3697},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 188, col: 6, offset: 3698},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 188, col: 6, offset: 3698},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 188, col: 12, offset: 3704},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 192, col: 1, offset: 3748},
			expr: &actionExpr{
				pos: position{line: 193, col: 5, offset: 3767},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 193, col: 5, offset: 3767},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 193, col: 5, offset: 3767},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 193, col: 10, offset: 3772},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 193, col: 18, offset: 3780},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 193, col: 23, offset: 3785},
								expr: &seqExpr{
									pos: position{line: 193, col: 25, offset: 3787},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 193, col: 25, offset: 3787},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 193, col: 28, offset: 3790},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 193, col: 51, offset: 3813},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 193, col: 54, offset: 3816},
											name: "Primary",
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
			name: "Primary",
			pos:  position{line: 197, col: 1, offset: 3893},
			expr: &choiceExpr{
				pos: position{line: 198, col: 5, offset: 3905},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 198, col: 5, offset: 3905},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 198, col: 5, offset: 3905},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 198, col: 5, offset: 3905},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 198, col: 9, offset: 3909},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 198, col: 12, offset: 3912},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 198, col: 17, offset: 3917},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 198, col: 25, offset: 3925},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 198, col: 28, offset: 3928},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 201, col: 5, offset: 3967},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 5, offset: 3985},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 203, col: 5, offset: 4014},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 204, col: 5, offset: 4027},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 5, offset: 4040},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 206, col: 5, offset: 4051},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 207, col: 5, offset: 4061},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 209, col: 1, offset: 4073},
			expr: &seqExpr{
				pos: position{line: 210, col: 5, offset: 4090},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 210, col: 5, offset: 4090},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 210, col: 11, offset: 4096},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 210, col: 17, offset: 4102},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 210, col: 23, offset: 4108},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 212, col: 1, offset: 4115},
			expr: &seqExpr{
				pos: position{line: 214, col: 5, offset: 4140},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 214, col: 5, offset: 4140},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 214, col: 11, offset: 4146},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 216, col: 1, offset: 4153},
			expr: &seqExpr{
				pos: position{line: 219, col: 5, offset: 4223},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 219, col: 5, offset: 4223},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 219, col: 11, offset: 4229},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 221, col: 1, offset: 4236},
			expr: &seqExpr{
				pos: position{line: 223, col: 5, offset: 4260},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 223, col: 5, offset: 4260},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 223, col: 11, offset: 4266},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 225, col: 1, offset: 4273},
			expr: &seqExpr{
				pos: position{line: 227, col: 5, offset: 4299},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 227, col: 5, offset: 4299},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 227, col: 11, offset: 4305},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 229, col: 1, offset: 4312},
			expr: &seqExpr{
				pos: position{line: 232, col: 5, offset: 4384},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 232, col: 5, offset: 4384},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 232, col: 11, offset: 4390},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 234, col: 1, offset: 4397},
			expr: &seqExpr{
				pos: position{line: 235, col: 5, offset: 4413},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 235, col: 5, offset: 4413},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 235, col: 9, offset: 4417},
						expr: &ruleRefExpr{
							pos:  position{line: 235, col: 9, offset: 4417},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 237, col: 1, offset: 4425},
			expr: &seqExpr{
				pos: position{line: 238, col: 5, offset: 4443},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 238, col: 6, offset: 4444},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 238, col: 6, offset: 4444},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 238, col: 12, offset: 4450},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 238, col: 17, offset: 4455},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 238, col: 26, offset: 4464},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 238, col: 30, offset: 4468},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 240, col: 1, offset: 4480},
			expr: &choiceExpr{
				pos: position{line: 241, col: 6, offset: 4496},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 241, col: 6, offset: 4496},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 241, col: 12, offset: 4502},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 243, col: 1, offset: 4518},
			expr: &seqExpr{
				pos: position{line: 244, col: 5, offset: 4534},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 244, col: 5, offset: 4534},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 244, col: 14, offset: 4543},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 244, col: 18, offset: 4547},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 244, col: 29, offset: 4558},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 244, col: 33, offset: 4562},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 244, col: 44, offset: 4573},
						expr: &ruleRefExpr{
							pos:  position{line: 244, col: 44, offset: 4573},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 246, col: 1, offset: 4587},
			expr: &seqExpr{
				pos: position{line: 247, col: 5, offset: 4600},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 247, col: 5, offset: 4600},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 247, col: 18, offset: 4613},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 247, col: 22, offset: 4617},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 247, col: 32, offset: 4627},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 247, col: 36, offset: 4631},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 249, col: 1, offset: 4641},
			expr: &seqExpr{
				pos: position{line: 250, col: 5, offset: 4654},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 250, col: 5, offset: 4654},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 250, col: 17, offset: 4666},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 252, col: 1, offset: 4678},
			expr: &actionExpr{
				pos: position{line: 253, col: 5, offset: 4691},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 253, col: 5, offset: 4691},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 253, col: 5, offset: 4691},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 253, col: 14, offset: 4700},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 18, offset: 4704},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 257, col: 1, offset: 4759},
			expr: &litMatcher{
				pos:        position{line: 258, col: 5, offset: 4779},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 260, col: 1, offset: 4785},
			expr: &choiceExpr{
				pos: position{line: 261, col: 6, offset: 4807},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 261, col: 6, offset: 4807},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 13, offset: 4814},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 20, offset: 4822},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 263, col: 1, offset: 4830},
			expr: &litMatcher{
				pos:        position{line: 264, col: 5, offset: 4851},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 266, col: 1, offset: 4857},
			expr: &litMatcher{
				pos:        position{line: 267, col: 5, offset: 4873},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 269, col: 1, offset: 4878},
			expr: &litMatcher{
				pos:        position{line: 270, col: 5, offset: 4894},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 272, col: 1, offset: 4899},
			expr: &litMatcher{
				pos:        position{line: 273, col: 5, offset: 4913},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 275, col: 1, offset: 4918},
			expr: &choiceExpr{
				pos: position{line: 277, col: 9, offset: 4946},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 277, col: 9, offset: 4946},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 278, col: 9, offset: 4970},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 279, col: 9, offset: 4995},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 280, col: 9, offset: 5020},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 281, col: 9, offset: 5040},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 282, col: 9, offset: 5060},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 285, col: 1, offset: 5077},
			expr: &seqExpr{
				pos: position{line: 286, col: 5, offset: 5096},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 286, col: 5, offset: 5096},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 286, col: 12, offset: 5103},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 288, col: 1, offset: 5118},
			expr: &actionExpr{
				pos: position{line: 289, col: 5, offset: 5131},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 289, col: 5, offset: 5131},
					expr: &ruleRefExpr{
						pos:  position{line: 289, col: 5, offset: 5131},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 293, col: 1, offset: 5200},
			expr: &choiceExpr{
				pos: position{line: 294, col: 5, offset: 5218},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 294, col: 5, offset: 5218},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 294, col: 7, offset: 5220},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 294, col: 7, offset: 5220},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 294, col: 11, offset: 5224},
									expr: &ruleRefExpr{
										pos:  position{line: 294, col: 11, offset: 5224},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 294, col: 29, offset: 5242},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 297, col: 5, offset: 5306},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 297, col: 7, offset: 5308},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 297, col: 7, offset: 5308},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 297, col: 11, offset: 5312},
									expr: &ruleRefExpr{
										pos:  position{line: 297, col: 11, offset: 5312},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 297, col: 31, offset: 5332},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 297, col: 31, offset: 5332},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 297, col: 37, offset: 5338},
											name: "EOF",
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
			name: "DoubleStringChar",
			pos:  position{line: 301, col: 1, offset: 5420},
			expr: &choiceExpr{
				pos: position{line: 302, col: 5, offset: 5441},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 302, col: 5, offset: 5441},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 302, col: 5, offset: 5441},
								expr: &choiceExpr{
									pos: position{line: 302, col: 8, offset: 5444},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 302, col: 8, offset: 5444},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 302, col: 14, offset: 5450},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 21, offset: 5457},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 302, col: 27, offset: 5463},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 303, col: 5, offset: 5478},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 303, col: 5, offset: 5478},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 303, col: 10, offset: 5483},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 305, col: 1, offset: 5503},
			expr: &choiceExpr{
				pos: position{line: 306, col: 5, offset: 5526},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 306, col: 5, offset: 5526},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 307, col: 5, offset: 5534},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 307, col: 7, offset: 5536},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 307, col: 7, offset: 5536},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 307, col: 20, offset: 5549},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 307, col: 26, offset: 5555},
									name: "EOF",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Identifier",
			pos:  position{line: 311, col: 1, offset: 5627},
			expr: &actionExpr{
				pos: position{line: 312, col: 5, offset: 5642},
				run: (*parser).callonIdentifier1,
				expr: &oneOrMoreExpr{
					pos: position{line: 312, col: 5, offset: 5642},
					expr: &ruleRefExpr{
						pos:  position{line: 312, col: 5, offset: 5642},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 316, col: 1, offset: 5702},
			expr: &seqExpr{
				pos: position{line: 317, col: 5, offset: 5717},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 317, col: 5, offset: 5717},
						expr: &choiceExpr{
							pos: position{line: 317, col: 7, offset: 5719},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 317, col: 7, offset: 5719},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 318, col: 9, offset: 5731},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 319, col: 9, offset: 5743},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 320, col: 9, offset: 5755},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 321, col: 9, offset: 5767},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 322, col: 9, offset: 5779},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 323, col: 9, offset: 5791},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 324, col: 9, offset: 5803},
									val:        "$",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 325, col: 9, offset: 5815},
									val:        ".",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 326, col: 9, offset: 5827},
									name: "ws",
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 327, col: 7, offset: 5836},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 329, col: 1, offset: 5848},
			expr: &actionExpr{
				pos: position{line: 330, col: 5, offset: 5859},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 330, col: 5, offset: 5859},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 330, col: 5, offset: 5859},
							expr: &litMatcher{
								pos:        position{line: 330, col: 5, offset: 5859},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 330, col: 10, offset: 5864},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 330, col: 18, offset: 5872},
							expr: &seqExpr{
								pos: position{line: 330, col: 20, offset: 5874},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 330, col: 20, offset: 5874},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 330, col: 24, offset: 5878},
										expr: &ruleRefExpr{
											pos:  position{line: 330, col: 24, offset: 5878},
											name: "Digit",
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
			name: "Integer",
			pos:  position{line: 334, col: 1, offset: 5939},
			expr: &choiceExpr{
				pos: position{line: 335, col: 5, offset: 5951},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 335, col: 5, offset: 5951},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 336, col: 5, offset: 5959},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 336, col: 5, offset: 5959},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 336, col: 18, offset: 5972},
								expr: &ruleRefExpr{
									pos:  position{line: 336, col: 18, offset: 5972},
									name: "Digit",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "NonZeroDigit",
			pos:  position{line: 338, col: 1, offset: 5980},
			expr: &charClassMatcher{
				pos:        position{line: 339, col: 5, offset: 5997},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 341, col: 1, offset: 6004},
			expr: &charClassMatcher{
				pos:        position{line: 342, col: 5, offset: 6014},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 344, col: 1, offset: 6021},
			expr: &actionExpr{
				pos: position{line: 345, col: 5, offset: 6031},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 5, offset: 6031},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 345, col: 11, offset: 6037},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 349, col: 1, offset: 6091},
			expr: &actionExpr{
				pos: position{line: 350, col: 5, offset: 6141},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 350, col: 5, offset: 6141},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 350, col: 5, offset: 6141},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 350, col: 9, offset: 6145},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 350, col: 17, offset: 6153},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 350, col: 39, offset: 6175},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 354, col: 1, offset: 6217},
			expr: &actionExpr{
				pos: position{line: 355, col: 5, offset: 6243},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 355, col: 5, offset: 6243},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 355, col: 11, offset: 6249},
						expr: &ruleRefExpr{
							pos:  position{line: 355, col: 11, offset: 6249},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 359, col: 1, offset: 6327},
			expr: &choiceExpr{
				pos: position{line: 360, col: 5, offset: 6353},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 360, col: 5, offset: 6353},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 360, col: 5, offset: 6353},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 360, col: 5, offset: 6353},
									expr: &charClassMatcher{
										pos:        position{line: 360, col: 6, offset: 6354},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 360, col: 12, offset: 6360},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 360, col: 15, offset: 6363},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 363, col: 5, offset: 6425},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 365, col: 1, offset: 6461},
			expr: &choiceExpr{
				pos: position{line: 366, col: 5, offset: 6500},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 366, col: 5, offset: 6500},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 366, col: 5, offset: 6500},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 369, col: 5, offset: 6538},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 369, col: 5, offset: 6538},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 369, col: 10, offset: 6543},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 371, col: 1, offset: 6575},
			expr: &actionExpr{
				pos: position{line: 372, col: 5, offset: 6610},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 372, col: 5, offset: 6610},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 372, col: 5, offset: 6610},
							expr: &ruleRefExpr{
								pos:  position{line: 372, col: 6, offset: 6611},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 372, col: 21, offset: 6626},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 376, col: 1, offset: 6677},
			expr: &anyMatcher{
				line: 377, col: 5, offset: 6692,
			},
		},
		{
			name: "__",
			pos:  position{line: 379, col: 1, offset: 6695},
			expr: &zeroOrMoreExpr{
				pos: position{line: 380, col: 5, offset: 6702},
				expr: &choiceExpr{
					pos: position{line: 380, col: 7, offset: 6704},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 380, col: 7, offset: 6704},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 380, col: 12, offset: 6709},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 382, col: 1, offset: 6717},
			expr: &charClassMatcher{
				pos:        position{line: 383, col: 5, offset: 6724},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 385, col: 1, offset: 6735},
			expr: &charClassMatcher{
				pos:        position{line: 386, col: 5, offset: 6754},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 388, col: 1, offset: 6762},
			expr: &litMatcher{
				pos:        position{line: 389, col: 5, offset: 6770},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 391, col: 1, offset: 6776},
			expr: &notExpr{
				pos: position{line: 392, col: 5, offset: 6784},
				expr: &anyMatcher{
					line: 392, col: 6, offset: 6785,
				},
			},
		},
	},
}

func (c *current) onStart1(program interface{}) (interface{}, error) {
	return program, nil

}

func (p *parser) callonStart1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStart1(stack["program"])
}

func (c *current) onProgram1(body interface{}) (interface{}, error) {
	return program(body, c.text, c.pos)

}

func (p *parser) callonProgram1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onProgram1(stack["body"])
}

func (c *current) onSourceElements1(head, tail interface{}) (interface{}, error) {
	return srcElems(head, tail)

}

func (p *parser) callonSourceElements1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSourceElements1(stack["head"], stack["tail"])
}

func (c *current) onVariableStatement1(declarations interface{}) (interface{}, error) {
	return varstmt(declarations, c.text, c.pos)

}

func (p *parser) callonVariableStatement1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onVariableStatement1(stack["declarations"])
}

func (c *current) onVariableDeclarationList1(head, tail interface{}) (interface{}, error) {
	return vardecls(head, tail)

}

func (p *parser) callonVariableDeclarationList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onVariableDeclarationList1(stack["head"], stack["tail"])
}

func (c *current) onVariableDeclaration1(id, initExpr interface{}) (interface{}, error) {
	return vardecl(id, initExpr, c.text, c.pos)

}

func (p *parser) callonVariableDeclaration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onVariableDeclaration1(stack["id"], stack["initExpr"])
}

func (c *current) onInitializer1(expression interface{}) (interface{}, error) {
	return expression, nil

}

func (p *parser) callonInitializer1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInitializer1(stack["expression"])
}

func (c *current) onExpressionStatement1(call interface{}) (interface{}, error) {
	return exprstmt(call, c.text, c.pos)

}

func (p *parser) callonExpressionStatement1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onExpressionStatement1(stack["call"])
}

func (c *current) onMemberExpression7(property interface{}) (interface{}, error) {
	return property, nil

}

func (p *parser) callonMemberExpression7() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMemberExpression7(stack["property"])
}

func (c *current) onMemberExpression1(head, tail interface{}) (interface{}, error) {
	return memberexprs(head, tail, c.text, c.pos)

}

func (p *parser) callonMemberExpression1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMemberExpression1(stack["head"], stack["tail"])
}

func (c *current) onCallExpression4(callee, args interface{}) (interface{}, error) {
	return callexpr(callee, args, c.text, c.pos)

}

func (p *parser) callonCallExpression4() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCallExpression4(stack["callee"], stack["args"])
}

func (c *current) onCallExpression14(args interface{}) (interface{}, error) {
	return callexpr(nil, args, c.text, c.pos)

}

func (p *parser) callonCallExpression14() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCallExpression14(stack["args"])
}

func (c *current) onCallExpression19(property interface{}) (interface{}, error) {
	return memberexpr(nil, property, c.text, c.pos)

}

func (p *parser) callonCallExpression19() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCallExpression19(stack["property"])
}

func (c *current) onCallExpression1(head, tail interface{}) (interface{}, error) {
	return callexprs(head, tail, c.text, c.pos)

}

func (p *parser) callonCallExpression1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCallExpression1(stack["head"], stack["tail"])
}

func (c *current) onArguments1(args interface{}) (interface{}, error) {
	return args, nil

}

func (p *parser) callonArguments1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onArguments1(stack["args"])
}

func (c *current) onFunctionArgs1(first, rest interface{}) (interface{}, error) {
	return object(first, rest, c.text, c.pos)

}

func (p *parser) callonFunctionArgs1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionArgs1(stack["first"], stack["rest"])
}

func (c *current) onFunctionArgsRest1(arg interface{}) (interface{}, error) {
	return arg, nil

}

func (p *parser) callonFunctionArgsRest1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionArgsRest1(stack["arg"])
}

func (c *current) onFunctionArg1(key, value interface{}) (interface{}, error) {
	return property(key, value, c.text, c.pos)

}

func (p *parser) callonFunctionArg1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionArg1(stack["key"], stack["value"])
}

func (c *current) onWhereExpr1(expr interface{}) (interface{}, error) {
	return expr, nil

}

func (p *parser) callonWhereExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWhereExpr1(stack["expr"])
}

func (c *current) onLogicalOperators1() (interface{}, error) {
	return logicalOp(c.text)

}

func (p *parser) callonLogicalOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLogicalOperators1()
}

func (c *current) onLogical1(head, tail interface{}) (interface{}, error) {
	return logicalExpression(head, tail, c.text, c.pos)

}

func (p *parser) callonLogical1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLogical1(stack["head"], stack["tail"])
}

func (c *current) onEqualityOperators1() (interface{}, error) {
	return binaryOp(c.text)

}

func (p *parser) callonEqualityOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEqualityOperators1()
}

func (c *current) onEquality1(head, tail interface{}) (interface{}, error) {
	return binaryExpression(head, tail, c.text, c.pos)

}

func (p *parser) callonEquality1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEquality1(stack["head"], stack["tail"])
}

func (c *current) onRelationalOperators1() (interface{}, error) {
	return binaryOp(c.text)

}

func (p *parser) callonRelationalOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRelationalOperators1()
}

func (c *current) onRelational1(head, tail interface{}) (interface{}, error) {
	return binaryExpression(head, tail, c.text, c.pos)

}

func (p *parser) callonRelational1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRelational1(stack["head"], stack["tail"])
}

func (c *current) onAdditiveOperator1() (interface{}, error) {
	return binaryOp(c.text)

}

func (p *parser) callonAdditiveOperator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAdditiveOperator1()
}

func (c *current) onAdditive1(head, tail interface{}) (interface{}, error) {

	return binaryExpression(head, tail, c.text, c.pos)

}

func (p *parser) callonAdditive1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAdditive1(stack["head"], stack["tail"])
}

func (c *current) onMultiplicativeOperator1() (interface{}, error) {
	return binaryOp(c.text)

}

func (p *parser) callonMultiplicativeOperator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiplicativeOperator1()
}

func (c *current) onMultiplicative1(head, tail interface{}) (interface{}, error) {
	return binaryExpression(head, tail, c.text, c.pos)

}

func (p *parser) callonMultiplicative1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiplicative1(stack["head"], stack["tail"])
}

func (c *current) onPrimary2(expr interface{}) (interface{}, error) {
	return expr, nil

}

func (p *parser) callonPrimary2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrimary2(stack["expr"])
}

func (c *current) onDateTime1() (interface{}, error) {
	return datetime(c.text, c.pos)

}

func (p *parser) callonDateTime1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDateTime1()
}

func (c *current) onDuration1() (interface{}, error) {
	return durationLiteral(c.text, c.pos)

}

func (p *parser) callonDuration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDuration1()
}

func (c *current) onStringLiteral2() (interface{}, error) {
	return stringLiteral(c.text, c.pos)

}

func (p *parser) callonStringLiteral2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStringLiteral2()
}

func (c *current) onStringLiteral8() (interface{}, error) {
	return "", errors.New("string literal not terminated")

}

func (p *parser) callonStringLiteral8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStringLiteral8()
}

func (c *current) onDoubleStringEscape3() (interface{}, error) {
	return nil, errors.New("invalid escape character")

}

func (p *parser) callonDoubleStringEscape3() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDoubleStringEscape3()
}

func (c *current) onIdentifier1() (interface{}, error) {
	return identifier(c.text, c.pos)

}

func (p *parser) callonIdentifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifier1()
}

func (c *current) onNumber1() (interface{}, error) {
	return numberLiteral(c.text, c.pos)

}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onField1(field interface{}) (interface{}, error) {
	return fieldLiteral(c.text, c.pos)

}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["field"])
}

func (c *current) onRegularExpressionLiteral1(pattern interface{}) (interface{}, error) {
	return pattern, nil

}

func (p *parser) callonRegularExpressionLiteral1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegularExpressionLiteral1(stack["pattern"])
}

func (c *current) onRegularExpressionBody1(chars interface{}) (interface{}, error) {
	return regexLiteral(chars, c.text, c.pos)

}

func (p *parser) callonRegularExpressionBody1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegularExpressionBody1(stack["chars"])
}

func (c *current) onRegularExpressionChar2(re interface{}) (interface{}, error) {
	return re, nil

}

func (p *parser) callonRegularExpressionChar2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegularExpressionChar2(stack["re"])
}

func (c *current) onRegularExpressionBackslashSequence2() (interface{}, error) {
	return "/", nil

}

func (p *parser) callonRegularExpressionBackslashSequence2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegularExpressionBackslashSequence2()
}

func (c *current) onRegularExpressionNonTerminator1() (interface{}, error) {
	return string(c.text), nil

}

func (p *parser) callonRegularExpressionNonTerminator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegularExpressionNonTerminator1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// GlobalStore creates an Option to set a key to a certain value in
// the globalStore.
func GlobalStore(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.globalStore[key]
		p.cur.globalStore[key] = value
		return GlobalStore(key, old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (i interface{}, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
	}()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match

	// the globalStore allows the parser to store arbitrary values
	globalStore map[string]interface{}
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos             position
	val             string
	basicLatinChars [128]bool
	chars           []rune
	ranges          []rune
	classes         []*unicode.RangeTable
	ignoreCase      bool
	inverted        bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner    error
	pos      position
	prefix   string
	expected []string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
		cur: current{
			globalStore: make(map[string]interface{}),
		},
		maxFailPos:      position{col: 1, line: 1},
		maxFailExpected: make([]string, 0, 20),
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	depth   int
	recover bool
	debug   bool

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position, []string{})
}

func (p *parser) addErrAt(err error, pos position, expected []string) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String(), expected: expected}
	p.errs.add(pe)
}

func (p *parser) failAt(fail bool, pos position, want string) {
	// process fail if parsing fails and not inverted or parsing succeeds and invert is set
	if fail == p.maxFailInvertExpected {
		if pos.offset < p.maxFailPos.offset {
			return
		}

		if pos.offset > p.maxFailPos.offset {
			p.maxFailPos = pos
			p.maxFailExpected = p.maxFailExpected[:0]
		}

		if p.maxFailInvertExpected {
			want = "!" + want
		}
		p.maxFailExpected = append(p.maxFailExpected, want)
	}
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n == 1 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// If parsing fails, but no errors have been recorded, the expected values
			// for the farthest parser position are returned as error.
			maxFailExpectedMap := make(map[string]struct{}, len(p.maxFailExpected))
			for _, v := range p.maxFailExpected {
				maxFailExpectedMap[v] = struct{}{}
			}
			expected := make([]string, 0, len(maxFailExpectedMap))
			eof := false
			if _, ok := maxFailExpectedMap["!."]; ok {
				delete(maxFailExpectedMap, "!.")
				eof = true
			}
			for k := range maxFailExpectedMap {
				expected = append(expected, k)
			}
			sort.Strings(expected)
			if eof {
				expected = append(expected, "EOF")
			}
			p.addErrAt(errors.New("no match found, expected: "+listJoin(expected, ", ", "or")), p.maxFailPos, expected)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func listJoin(list []string, sep string, lastSep string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return fmt.Sprintf("%s %s %s", strings.Join(list[:len(list)-1], sep), lastSep, list[len(list)-1])
	}
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	var ok bool
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position, []string{})
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		p.failAt(true, start.position, ".")
		return p.sliceFrom(start), true
	}
	p.failAt(false, p.pt.position, ".")
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	start := p.pt

	// can't match EOF
	if cur == utf8.RuneError {
		p.failAt(false, start.position, chr.val)
		return nil, false
	}

	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		p.failAt(true, start.position, chr.val)
		return p.sliceFrom(start), true
	}
	p.failAt(false, start.position, chr.val)
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	ignoreCase := ""
	if lit.ignoreCase {
		ignoreCase = "i"
	}
	val := fmt.Sprintf("%q%s", lit.val, ignoreCase)
	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.failAt(false, start.position, val)
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	p.failAt(true, start.position, val)
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	_, ok := p.parseExpr(not.expr)
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	vals := make([]interface{}, 0, len(seq.exprs))

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}
