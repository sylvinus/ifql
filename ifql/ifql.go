package ifql

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
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
							label: "declaration",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 29, offset: 465},
								name: "VariableDeclaration",
							},
						},
					},
				},
			},
		},
		{
			name: "VarToken",
			pos:  position{line: 34, col: 1, offset: 543},
			expr: &litMatcher{
				pos:        position{line: 34, col: 12, offset: 554},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 36, col: 1, offset: 561},
			expr: &actionExpr{
				pos: position{line: 37, col: 5, offset: 585},
				run: (*parser).callonVariableDeclaration1,
				expr: &seqExpr{
					pos: position{line: 37, col: 5, offset: 585},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 37, col: 5, offset: 585},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 8, offset: 588},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 19, offset: 599},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 37, col: 22, offset: 602},
							label: "initExpr",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 31, offset: 611},
								name: "Initializer",
							},
						},
					},
				},
			},
		},
		{
			name: "Initializer",
			pos:  position{line: 41, col: 1, offset: 682},
			expr: &actionExpr{
				pos: position{line: 42, col: 5, offset: 698},
				run: (*parser).callonInitializer1,
				expr: &seqExpr{
					pos: position{line: 42, col: 5, offset: 698},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 42, col: 5, offset: 698},
							val:        "=",
							ignoreCase: false,
						},
						&notExpr{
							pos: position{line: 42, col: 9, offset: 702},
							expr: &litMatcher{
								pos:        position{line: 42, col: 10, offset: 703},
								val:        "=",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 42, col: 14, offset: 707},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 42, col: 17, offset: 710},
							label: "expression",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 28, offset: 721},
								name: "VariableExpression",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 47, col: 1, offset: 848},
			expr: &choiceExpr{
				pos: position{line: 48, col: 5, offset: 871},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 48, col: 5, offset: 871},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 49, col: 5, offset: 890},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 50, col: 5, offset: 908},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 51, col: 5, offset: 937},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 5, offset: 950},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 53, col: 5, offset: 963},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 54, col: 5, offset: 974},
						name: "Field",
					},
				},
			},
		},
		{
			name: "ExpressionStatement",
			pos:  position{line: 56, col: 1, offset: 981},
			expr: &actionExpr{
				pos: position{line: 57, col: 5, offset: 1005},
				run: (*parser).callonExpressionStatement1,
				expr: &labeledExpr{
					pos:   position{line: 57, col: 5, offset: 1005},
					label: "call",
					expr: &ruleRefExpr{
						pos:  position{line: 57, col: 10, offset: 1010},
						name: "CallExpression",
					},
				},
			},
		},
		{
			name: "MemberExpression",
			pos:  position{line: 61, col: 1, offset: 1079},
			expr: &actionExpr{
				pos: position{line: 62, col: 5, offset: 1100},
				run: (*parser).callonMemberExpression1,
				expr: &seqExpr{
					pos: position{line: 62, col: 5, offset: 1100},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 62, col: 5, offset: 1100},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 62, col: 10, offset: 1105},
								name: "Identifier",
							},
						},
						&labeledExpr{
							pos:   position{line: 63, col: 5, offset: 1147},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 63, col: 10, offset: 1152},
								expr: &actionExpr{
									pos: position{line: 64, col: 9, offset: 1162},
									run: (*parser).callonMemberExpression7,
									expr: &seqExpr{
										pos: position{line: 64, col: 9, offset: 1162},
										exprs: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 64, col: 9, offset: 1162},
												name: "__",
											},
											&litMatcher{
												pos:        position{line: 64, col: 12, offset: 1165},
												val:        ".",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 64, col: 16, offset: 1169},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 64, col: 19, offset: 1172},
												label: "property",
												expr: &ruleRefExpr{
													pos:  position{line: 64, col: 28, offset: 1181},
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
			pos:  position{line: 72, col: 1, offset: 1307},
			expr: &actionExpr{
				pos: position{line: 73, col: 5, offset: 1326},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 73, col: 5, offset: 1326},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 73, col: 5, offset: 1326},
							label: "head",
							expr: &actionExpr{
								pos: position{line: 74, col: 7, offset: 1339},
								run: (*parser).callonCallExpression4,
								expr: &seqExpr{
									pos: position{line: 74, col: 7, offset: 1339},
									exprs: []interface{}{
										&labeledExpr{
											pos:   position{line: 74, col: 7, offset: 1339},
											label: "callee",
											expr: &ruleRefExpr{
												pos:  position{line: 74, col: 14, offset: 1346},
												name: "MemberExpression",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 74, col: 31, offset: 1363},
											name: "__",
										},
										&labeledExpr{
											pos:   position{line: 74, col: 34, offset: 1366},
											label: "args",
											expr: &ruleRefExpr{
												pos:  position{line: 74, col: 39, offset: 1371},
												name: "Arguments",
											},
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 78, col: 5, offset: 1454},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 78, col: 10, offset: 1459},
								expr: &choiceExpr{
									pos: position{line: 79, col: 9, offset: 1469},
									alternatives: []interface{}{
										&actionExpr{
											pos: position{line: 79, col: 9, offset: 1469},
											run: (*parser).callonCallExpression14,
											expr: &seqExpr{
												pos: position{line: 79, col: 9, offset: 1469},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 79, col: 9, offset: 1469},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 79, col: 12, offset: 1472},
														label: "args",
														expr: &ruleRefExpr{
															pos:  position{line: 79, col: 17, offset: 1477},
															name: "Arguments",
														},
													},
												},
											},
										},
										&actionExpr{
											pos: position{line: 82, col: 9, offset: 1559},
											run: (*parser).callonCallExpression19,
											expr: &seqExpr{
												pos: position{line: 82, col: 9, offset: 1559},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 82, col: 9, offset: 1559},
														name: "__",
													},
													&litMatcher{
														pos:        position{line: 82, col: 12, offset: 1562},
														val:        ".",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 82, col: 16, offset: 1566},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 82, col: 19, offset: 1569},
														label: "property",
														expr: &ruleRefExpr{
															pos:  position{line: 82, col: 28, offset: 1578},
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
			pos:  position{line: 90, col: 1, offset: 1729},
			expr: &actionExpr{
				pos: position{line: 91, col: 5, offset: 1743},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 91, col: 5, offset: 1743},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 91, col: 5, offset: 1743},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 91, col: 9, offset: 1747},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 91, col: 12, offset: 1750},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 91, col: 17, offset: 1755},
								expr: &ruleRefExpr{
									pos:  position{line: 91, col: 18, offset: 1756},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 91, col: 33, offset: 1771},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 91, col: 36, offset: 1774},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 95, col: 1, offset: 1810},
			expr: &actionExpr{
				pos: position{line: 96, col: 5, offset: 1827},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 96, col: 5, offset: 1827},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 96, col: 5, offset: 1827},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 96, col: 11, offset: 1833},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 96, col: 23, offset: 1845},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 96, col: 26, offset: 1848},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 96, col: 31, offset: 1853},
								expr: &ruleRefExpr{
									pos:  position{line: 96, col: 31, offset: 1853},
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
			pos:  position{line: 100, col: 1, offset: 1928},
			expr: &actionExpr{
				pos: position{line: 101, col: 5, offset: 1949},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 101, col: 5, offset: 1949},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 101, col: 5, offset: 1949},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 101, col: 9, offset: 1953},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 101, col: 13, offset: 1957},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 101, col: 17, offset: 1961},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 105, col: 1, offset: 2004},
			expr: &actionExpr{
				pos: position{line: 106, col: 5, offset: 2020},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 106, col: 5, offset: 2020},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 106, col: 5, offset: 2020},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 106, col: 9, offset: 2024},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 20, offset: 2035},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 106, col: 24, offset: 2039},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 28, offset: 2043},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 106, col: 31, offset: 2046},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 106, col: 37, offset: 2052},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 110, col: 1, offset: 2128},
			expr: &choiceExpr{
				pos: position{line: 111, col: 5, offset: 2150},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 111, col: 5, offset: 2150},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 5, offset: 2164},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 5, offset: 2182},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 114, col: 5, offset: 2211},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 115, col: 5, offset: 2224},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 116, col: 5, offset: 2237},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 117, col: 5, offset: 2248},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 119, col: 1, offset: 2260},
			expr: &actionExpr{
				pos: position{line: 120, col: 5, offset: 2274},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 120, col: 5, offset: 2274},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 120, col: 5, offset: 2274},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 120, col: 9, offset: 2278},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 120, col: 12, offset: 2281},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 120, col: 17, offset: 2286},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 120, col: 22, offset: 2291},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 120, col: 26, offset: 2295},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 131, col: 1, offset: 2518},
			expr: &ruleRefExpr{
				pos:  position{line: 132, col: 5, offset: 2527},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 134, col: 1, offset: 2536},
			expr: &actionExpr{
				pos: position{line: 135, col: 5, offset: 2557},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 135, col: 6, offset: 2558},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 135, col: 6, offset: 2558},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 135, col: 14, offset: 2566},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 139, col: 1, offset: 2618},
			expr: &actionExpr{
				pos: position{line: 140, col: 5, offset: 2630},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 140, col: 5, offset: 2630},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 140, col: 5, offset: 2630},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 140, col: 10, offset: 2635},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 140, col: 19, offset: 2644},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 140, col: 24, offset: 2649},
								expr: &seqExpr{
									pos: position{line: 140, col: 26, offset: 2651},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 140, col: 26, offset: 2651},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 140, col: 30, offset: 2655},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 140, col: 47, offset: 2672},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 140, col: 51, offset: 2676},
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
			pos:  position{line: 144, col: 1, offset: 2755},
			expr: &actionExpr{
				pos: position{line: 145, col: 5, offset: 2777},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 145, col: 6, offset: 2778},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 145, col: 6, offset: 2778},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 145, col: 13, offset: 2785},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 149, col: 1, offset: 2831},
			expr: &actionExpr{
				pos: position{line: 150, col: 5, offset: 2844},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 150, col: 5, offset: 2844},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 150, col: 5, offset: 2844},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 150, col: 10, offset: 2849},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 150, col: 21, offset: 2860},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 150, col: 26, offset: 2865},
								expr: &seqExpr{
									pos: position{line: 150, col: 28, offset: 2867},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 150, col: 28, offset: 2867},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 150, col: 31, offset: 2870},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 150, col: 49, offset: 2888},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 150, col: 52, offset: 2891},
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
			pos:  position{line: 154, col: 1, offset: 2971},
			expr: &actionExpr{
				pos: position{line: 155, col: 5, offset: 2995},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 155, col: 9, offset: 2999},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 155, col: 9, offset: 2999},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 156, col: 9, offset: 3012},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 157, col: 9, offset: 3024},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 158, col: 9, offset: 3037},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 159, col: 9, offset: 3049},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 160, col: 9, offset: 3071},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 161, col: 9, offset: 3085},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 162, col: 9, offset: 3106},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 167, col: 1, offset: 3164},
			expr: &actionExpr{
				pos: position{line: 168, col: 5, offset: 3179},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 168, col: 5, offset: 3179},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 168, col: 5, offset: 3179},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 168, col: 10, offset: 3184},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 168, col: 19, offset: 3193},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 168, col: 24, offset: 3198},
								expr: &seqExpr{
									pos: position{line: 168, col: 26, offset: 3200},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 168, col: 26, offset: 3200},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 168, col: 29, offset: 3203},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 168, col: 49, offset: 3223},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 168, col: 52, offset: 3226},
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
			pos:  position{line: 172, col: 1, offset: 3304},
			expr: &actionExpr{
				pos: position{line: 173, col: 5, offset: 3325},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 173, col: 6, offset: 3326},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 173, col: 6, offset: 3326},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 173, col: 12, offset: 3332},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 177, col: 1, offset: 3380},
			expr: &actionExpr{
				pos: position{line: 178, col: 5, offset: 3393},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 178, col: 5, offset: 3393},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 178, col: 5, offset: 3393},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 10, offset: 3398},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 178, col: 25, offset: 3413},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 178, col: 30, offset: 3418},
								expr: &seqExpr{
									pos: position{line: 178, col: 32, offset: 3420},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 178, col: 32, offset: 3420},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 178, col: 35, offset: 3423},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 178, col: 52, offset: 3440},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 178, col: 55, offset: 3443},
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
			pos:  position{line: 182, col: 1, offset: 3527},
			expr: &actionExpr{
				pos: position{line: 183, col: 5, offset: 3554},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 183, col: 6, offset: 3555},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 183, col: 6, offset: 3555},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 183, col: 12, offset: 3561},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 187, col: 1, offset: 3605},
			expr: &actionExpr{
				pos: position{line: 188, col: 5, offset: 3624},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 188, col: 5, offset: 3624},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 188, col: 5, offset: 3624},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 188, col: 10, offset: 3629},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 188, col: 18, offset: 3637},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 188, col: 23, offset: 3642},
								expr: &seqExpr{
									pos: position{line: 188, col: 25, offset: 3644},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 188, col: 25, offset: 3644},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 188, col: 28, offset: 3647},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 188, col: 51, offset: 3670},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 188, col: 54, offset: 3673},
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
			pos:  position{line: 192, col: 1, offset: 3750},
			expr: &choiceExpr{
				pos: position{line: 193, col: 5, offset: 3762},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 193, col: 5, offset: 3762},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 193, col: 5, offset: 3762},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 193, col: 5, offset: 3762},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 193, col: 9, offset: 3766},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 193, col: 12, offset: 3769},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 193, col: 17, offset: 3774},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 193, col: 25, offset: 3782},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 193, col: 28, offset: 3785},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 196, col: 5, offset: 3824},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 197, col: 5, offset: 3842},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 198, col: 5, offset: 3871},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 199, col: 5, offset: 3884},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 200, col: 5, offset: 3897},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 201, col: 5, offset: 3908},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 5, offset: 3918},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 204, col: 1, offset: 3930},
			expr: &seqExpr{
				pos: position{line: 205, col: 5, offset: 3947},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 205, col: 5, offset: 3947},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 11, offset: 3953},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 17, offset: 3959},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 23, offset: 3965},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 207, col: 1, offset: 3972},
			expr: &seqExpr{
				pos: position{line: 209, col: 5, offset: 3997},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 209, col: 5, offset: 3997},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 209, col: 11, offset: 4003},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 211, col: 1, offset: 4010},
			expr: &seqExpr{
				pos: position{line: 214, col: 5, offset: 4080},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 214, col: 5, offset: 4080},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 214, col: 11, offset: 4086},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 216, col: 1, offset: 4093},
			expr: &seqExpr{
				pos: position{line: 218, col: 5, offset: 4117},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 218, col: 5, offset: 4117},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 218, col: 11, offset: 4123},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 220, col: 1, offset: 4130},
			expr: &seqExpr{
				pos: position{line: 222, col: 5, offset: 4156},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 222, col: 5, offset: 4156},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 222, col: 11, offset: 4162},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 224, col: 1, offset: 4169},
			expr: &seqExpr{
				pos: position{line: 227, col: 5, offset: 4241},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 227, col: 5, offset: 4241},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 227, col: 11, offset: 4247},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 229, col: 1, offset: 4254},
			expr: &seqExpr{
				pos: position{line: 230, col: 5, offset: 4270},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 230, col: 5, offset: 4270},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 230, col: 9, offset: 4274},
						expr: &ruleRefExpr{
							pos:  position{line: 230, col: 9, offset: 4274},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 232, col: 1, offset: 4282},
			expr: &seqExpr{
				pos: position{line: 233, col: 5, offset: 4300},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 233, col: 6, offset: 4301},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 233, col: 6, offset: 4301},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 233, col: 12, offset: 4307},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 233, col: 17, offset: 4312},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 233, col: 26, offset: 4321},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 233, col: 30, offset: 4325},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 235, col: 1, offset: 4337},
			expr: &choiceExpr{
				pos: position{line: 236, col: 6, offset: 4353},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 236, col: 6, offset: 4353},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 236, col: 12, offset: 4359},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 238, col: 1, offset: 4375},
			expr: &seqExpr{
				pos: position{line: 239, col: 5, offset: 4391},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 239, col: 5, offset: 4391},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 239, col: 14, offset: 4400},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 239, col: 18, offset: 4404},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 239, col: 29, offset: 4415},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 239, col: 33, offset: 4419},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 239, col: 44, offset: 4430},
						expr: &ruleRefExpr{
							pos:  position{line: 239, col: 44, offset: 4430},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 241, col: 1, offset: 4444},
			expr: &seqExpr{
				pos: position{line: 242, col: 5, offset: 4457},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 242, col: 5, offset: 4457},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 242, col: 18, offset: 4470},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 242, col: 22, offset: 4474},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 242, col: 32, offset: 4484},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 242, col: 36, offset: 4488},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 244, col: 1, offset: 4498},
			expr: &seqExpr{
				pos: position{line: 245, col: 5, offset: 4511},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 245, col: 5, offset: 4511},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 245, col: 17, offset: 4523},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 247, col: 1, offset: 4535},
			expr: &actionExpr{
				pos: position{line: 248, col: 5, offset: 4548},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 248, col: 5, offset: 4548},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 248, col: 5, offset: 4548},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 248, col: 14, offset: 4557},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 18, offset: 4561},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 252, col: 1, offset: 4616},
			expr: &litMatcher{
				pos:        position{line: 253, col: 5, offset: 4636},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 255, col: 1, offset: 4642},
			expr: &choiceExpr{
				pos: position{line: 256, col: 6, offset: 4664},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 256, col: 6, offset: 4664},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 256, col: 13, offset: 4671},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 256, col: 20, offset: 4679},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 258, col: 1, offset: 4687},
			expr: &litMatcher{
				pos:        position{line: 259, col: 5, offset: 4708},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 261, col: 1, offset: 4714},
			expr: &litMatcher{
				pos:        position{line: 262, col: 5, offset: 4730},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 264, col: 1, offset: 4735},
			expr: &litMatcher{
				pos:        position{line: 265, col: 5, offset: 4751},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 267, col: 1, offset: 4756},
			expr: &litMatcher{
				pos:        position{line: 268, col: 5, offset: 4770},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 270, col: 1, offset: 4775},
			expr: &choiceExpr{
				pos: position{line: 272, col: 9, offset: 4803},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 272, col: 9, offset: 4803},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 273, col: 9, offset: 4827},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 274, col: 9, offset: 4852},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 275, col: 9, offset: 4877},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 276, col: 9, offset: 4897},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 277, col: 9, offset: 4917},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 280, col: 1, offset: 4934},
			expr: &seqExpr{
				pos: position{line: 281, col: 5, offset: 4953},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 281, col: 5, offset: 4953},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 281, col: 12, offset: 4960},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 283, col: 1, offset: 4975},
			expr: &actionExpr{
				pos: position{line: 284, col: 5, offset: 4988},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 284, col: 5, offset: 4988},
					expr: &ruleRefExpr{
						pos:  position{line: 284, col: 5, offset: 4988},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 288, col: 1, offset: 5057},
			expr: &choiceExpr{
				pos: position{line: 289, col: 5, offset: 5075},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 289, col: 5, offset: 5075},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 289, col: 7, offset: 5077},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 289, col: 7, offset: 5077},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 289, col: 11, offset: 5081},
									expr: &ruleRefExpr{
										pos:  position{line: 289, col: 11, offset: 5081},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 289, col: 29, offset: 5099},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 292, col: 5, offset: 5163},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 292, col: 7, offset: 5165},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 292, col: 7, offset: 5165},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 292, col: 11, offset: 5169},
									expr: &ruleRefExpr{
										pos:  position{line: 292, col: 11, offset: 5169},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 292, col: 31, offset: 5189},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 292, col: 31, offset: 5189},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 292, col: 37, offset: 5195},
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
			pos:  position{line: 296, col: 1, offset: 5277},
			expr: &choiceExpr{
				pos: position{line: 297, col: 5, offset: 5298},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 297, col: 5, offset: 5298},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 297, col: 5, offset: 5298},
								expr: &choiceExpr{
									pos: position{line: 297, col: 8, offset: 5301},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 297, col: 8, offset: 5301},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 297, col: 14, offset: 5307},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 297, col: 21, offset: 5314},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 297, col: 27, offset: 5320},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 298, col: 5, offset: 5335},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 298, col: 5, offset: 5335},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 298, col: 10, offset: 5340},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 300, col: 1, offset: 5360},
			expr: &choiceExpr{
				pos: position{line: 301, col: 5, offset: 5383},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 301, col: 5, offset: 5383},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 302, col: 5, offset: 5391},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 302, col: 7, offset: 5393},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 302, col: 7, offset: 5393},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 302, col: 20, offset: 5406},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 302, col: 26, offset: 5412},
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
			pos:  position{line: 306, col: 1, offset: 5484},
			expr: &actionExpr{
				pos: position{line: 307, col: 5, offset: 5499},
				run: (*parser).callonIdentifier1,
				expr: &oneOrMoreExpr{
					pos: position{line: 307, col: 5, offset: 5499},
					expr: &ruleRefExpr{
						pos:  position{line: 307, col: 5, offset: 5499},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 311, col: 1, offset: 5559},
			expr: &seqExpr{
				pos: position{line: 312, col: 5, offset: 5574},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 312, col: 5, offset: 5574},
						expr: &choiceExpr{
							pos: position{line: 312, col: 7, offset: 5576},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 312, col: 7, offset: 5576},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 313, col: 9, offset: 5588},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 314, col: 9, offset: 5600},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 315, col: 9, offset: 5612},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 316, col: 9, offset: 5624},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 317, col: 9, offset: 5636},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 318, col: 9, offset: 5648},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 319, col: 9, offset: 5660},
									val:        "$",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 320, col: 9, offset: 5672},
									val:        ".",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 321, col: 9, offset: 5684},
									name: "ws",
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 322, col: 7, offset: 5693},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 324, col: 1, offset: 5705},
			expr: &choiceExpr{
				pos: position{line: 325, col: 5, offset: 5716},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 325, col: 5, offset: 5716},
						run: (*parser).callonNumber2,
						expr: &seqExpr{
							pos: position{line: 325, col: 5, offset: 5716},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 325, col: 5, offset: 5716},
									name: "Integer",
								},
								&litMatcher{
									pos:        position{line: 325, col: 13, offset: 5724},
									val:        ".",
									ignoreCase: false,
								},
								&oneOrMoreExpr{
									pos: position{line: 325, col: 17, offset: 5728},
									expr: &ruleRefExpr{
										pos:  position{line: 325, col: 17, offset: 5728},
										name: "Digit",
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 328, col: 6, offset: 5789},
						run: (*parser).callonNumber8,
						expr: &ruleRefExpr{
							pos:  position{line: 328, col: 6, offset: 5789},
							name: "Integer",
						},
					},
				},
			},
		},
		{
			name: "Integer",
			pos:  position{line: 332, col: 1, offset: 5849},
			expr: &choiceExpr{
				pos: position{line: 333, col: 5, offset: 5861},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 333, col: 5, offset: 5861},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 334, col: 5, offset: 5869},
						exprs: []interface{}{
							&zeroOrOneExpr{
								pos: position{line: 334, col: 5, offset: 5869},
								expr: &litMatcher{
									pos:        position{line: 334, col: 5, offset: 5869},
									val:        "-",
									ignoreCase: false,
								},
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 10, offset: 5874},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 334, col: 23, offset: 5887},
								expr: &ruleRefExpr{
									pos:  position{line: 334, col: 23, offset: 5887},
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
			pos:  position{line: 336, col: 1, offset: 5895},
			expr: &charClassMatcher{
				pos:        position{line: 337, col: 5, offset: 5912},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 339, col: 1, offset: 5919},
			expr: &charClassMatcher{
				pos:        position{line: 340, col: 5, offset: 5929},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 342, col: 1, offset: 5936},
			expr: &actionExpr{
				pos: position{line: 343, col: 5, offset: 5946},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 343, col: 5, offset: 5946},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 343, col: 11, offset: 5952},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 347, col: 1, offset: 6006},
			expr: &actionExpr{
				pos: position{line: 348, col: 5, offset: 6056},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 348, col: 5, offset: 6056},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 348, col: 5, offset: 6056},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 348, col: 9, offset: 6060},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 348, col: 17, offset: 6068},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 348, col: 39, offset: 6090},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 352, col: 1, offset: 6132},
			expr: &actionExpr{
				pos: position{line: 353, col: 5, offset: 6158},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 353, col: 5, offset: 6158},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 353, col: 11, offset: 6164},
						expr: &ruleRefExpr{
							pos:  position{line: 353, col: 11, offset: 6164},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 357, col: 1, offset: 6242},
			expr: &choiceExpr{
				pos: position{line: 358, col: 5, offset: 6268},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 358, col: 5, offset: 6268},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 358, col: 5, offset: 6268},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 358, col: 5, offset: 6268},
									expr: &charClassMatcher{
										pos:        position{line: 358, col: 6, offset: 6269},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 358, col: 12, offset: 6275},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 358, col: 15, offset: 6278},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 361, col: 5, offset: 6340},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 363, col: 1, offset: 6376},
			expr: &choiceExpr{
				pos: position{line: 364, col: 5, offset: 6415},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 364, col: 5, offset: 6415},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 364, col: 5, offset: 6415},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 367, col: 5, offset: 6453},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 367, col: 5, offset: 6453},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 367, col: 10, offset: 6458},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 369, col: 1, offset: 6490},
			expr: &actionExpr{
				pos: position{line: 370, col: 5, offset: 6525},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 370, col: 5, offset: 6525},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 370, col: 5, offset: 6525},
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 6, offset: 6526},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 21, offset: 6541},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 374, col: 1, offset: 6592},
			expr: &anyMatcher{
				line: 375, col: 5, offset: 6607,
			},
		},
		{
			name: "__",
			pos:  position{line: 377, col: 1, offset: 6610},
			expr: &zeroOrMoreExpr{
				pos: position{line: 378, col: 5, offset: 6617},
				expr: &choiceExpr{
					pos: position{line: 378, col: 7, offset: 6619},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 378, col: 7, offset: 6619},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 12, offset: 6624},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 380, col: 1, offset: 6632},
			expr: &charClassMatcher{
				pos:        position{line: 381, col: 5, offset: 6639},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 383, col: 1, offset: 6650},
			expr: &charClassMatcher{
				pos:        position{line: 384, col: 5, offset: 6669},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 386, col: 1, offset: 6677},
			expr: &litMatcher{
				pos:        position{line: 387, col: 5, offset: 6685},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 389, col: 1, offset: 6691},
			expr: &notExpr{
				pos: position{line: 390, col: 5, offset: 6699},
				expr: &anyMatcher{
					line: 390, col: 6, offset: 6700,
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

func (c *current) onVariableStatement1(declaration interface{}) (interface{}, error) {
	return varstmt(declaration, c.text, c.pos)

}

func (p *parser) callonVariableStatement1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onVariableStatement1(stack["declaration"])
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

func (c *current) onNumber2() (interface{}, error) {
	return numberLiteral(c.text, c.pos)

}

func (p *parser) callonNumber2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber2()
}

func (c *current) onNumber8() (interface{}, error) {
	return integerLiteral(c.text, c.pos)

}

func (p *parser) callonNumber8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber8()
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

	// errMaxExprCnt is used to signal that the maximum number of
	// expressions have been parsed.
	errMaxExprCnt = errors.New("max number of expresssions parsed")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// MaxExpressions creates an Option to stop parsing after the provided
// number of expressions have been parsed, if the value is 0 then the parser will
// parse for as many steps as needed (possibly an infinite number).
//
// The default for maxExprCnt is 0.
func MaxExpressions(maxExprCnt uint64) Option {
	return func(p *parser) Option {
		oldMaxExprCnt := p.maxExprCnt
		p.maxExprCnt = maxExprCnt
		return MaxExpressions(oldMaxExprCnt)
	}
}

// Statistics adds a user provided Stats struct to the parser to allow
// the user to process the results after the parsing has finished.
// Also the key for the "no match" counter is set.
//
// Example usage:
//
//     input := "input"
//     stats := Stats{}
//     _, err := Parse("input-file", []byte(input), Statistics(&stats, "no match"))
//     if err != nil {
//         log.Panicln(err)
//     }
//     b, err := json.MarshalIndent(stats.ChoiceAltCnt, "", "  ")
//     if err != nil {
//         log.Panicln(err)
//     }
//     fmt.Println(string(b))
//
func Statistics(stats *Stats, choiceNoMatch string) Option {
	return func(p *parser) Option {
		oldStats := p.Stats
		p.Stats = stats
		oldChoiceNoMatch := p.choiceNoMatch
		p.choiceNoMatch = choiceNoMatch
		if p.Stats.ChoiceAltCnt == nil {
			p.Stats.ChoiceAltCnt = make(map[string]map[string]int)
		}
		return Statistics(oldStats, oldChoiceNoMatch)
	}
}

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
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
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

type recoveryExpr struct {
	pos          position
	expr         interface{}
	recoverExpr  interface{}
	failureLabel []string
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type throwExpr struct {
	pos   position
	label string
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
	stats := Stats{
		ChoiceAltCnt: make(map[string]map[string]int),
	}

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
		Stats:           &stats,
	}
	p.setOptions(opts)

	if p.maxExprCnt == 0 {
		p.maxExprCnt = math.MaxUint64
	}

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

const choiceNoMatch = -1

// Stats stores some statistics, gathered during parsing
type Stats struct {
	// ExprCnt counts the number of expressions processed during parsing
	// This value is compared to the maximum number of expressions allowed
	// (set by the MaxExpressions option).
	ExprCnt uint64

	// ChoiceAltCnt is used to count for each ordered choice expression,
	// which alternative is used how may times.
	// These numbers allow to optimize the order of the ordered choice expression
	// to increase the performance of the parser
	//
	// The outer key of ChoiceAltCnt is composed of the name of the rule as well
	// as the line and the column of the ordered choice.
	// The inner key of ChoiceAltCnt is the number (one-based) of the matching alternative.
	// For each alternative the number of matches are counted. If an ordered choice does not
	// match, a special counter is incremented. The name of this counter is set with
	// the parser option Statistics.
	// For an alternative to be included in ChoiceAltCnt, it has to match at least once.
	ChoiceAltCnt map[string]map[string]int
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

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool

	// max number of expressions to be parsed
	maxExprCnt uint64

	*Stats

	choiceNoMatch string
	// recovery expression stack, keeps track of the currently available recovery expression, these are traversed in reverse
	recoveryStack []map[string]interface{}
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

// push a recovery expression with its labels to the recoveryStack
func (p *parser) pushRecovery(labels []string, expr interface{}) {
	if cap(p.recoveryStack) == len(p.recoveryStack) {
		// create new empty slot in the stack
		p.recoveryStack = append(p.recoveryStack, nil)
	} else {
		// slice to 1 more
		p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)+1]
	}

	m := make(map[string]interface{}, len(labels))
	for _, fl := range labels {
		m[fl] = expr
	}
	p.recoveryStack[len(p.recoveryStack)-1] = m
}

// pop a recovery expression from the recoveryStack
func (p *parser) popRecovery() {
	// GC that map
	p.recoveryStack[len(p.recoveryStack)-1] = nil

	p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)-1]
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

	p.ExprCnt++
	if p.ExprCnt > p.maxExprCnt {
		panic(errMaxExprCnt)
	}

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
	case *recoveryExpr:
		val, ok = p.parseRecoveryExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *throwExpr:
		val, ok = p.parseThrowExpr(expr)
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

func (p *parser) incChoiceAltCnt(ch *choiceExpr, altI int) {
	choiceIdent := fmt.Sprintf("%s %d:%d", p.rstack[len(p.rstack)-1].name, ch.pos.line, ch.pos.col)
	m := p.ChoiceAltCnt[choiceIdent]
	if m == nil {
		m = make(map[string]int)
		p.ChoiceAltCnt[choiceIdent] = m
	}
	// We increment altI by 1, so the keys do not start at 0
	alt := strconv.Itoa(altI + 1)
	if altI == choiceNoMatch {
		alt = p.choiceNoMatch
	}
	m[alt]++
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for altI, alt := range ch.alternatives {
		// dummy assignment to prevent compile error if optimized
		_ = altI

		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			p.incChoiceAltCnt(ch, altI)
			return val, ok
		}
	}
	p.incChoiceAltCnt(ch, choiceNoMatch)
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

func (p *parser) parseRecoveryExpr(recover *recoveryExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRecoveryExpr (" + strings.Join(recover.failureLabel, ",") + ")"))
	}

	p.pushRecovery(recover.failureLabel, recover.recoverExpr)
	val, ok := p.parseExpr(recover.expr)
	p.popRecovery()

	return val, ok
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

func (p *parser) parseThrowExpr(expr *throwExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseThrowExpr"))
	}

	for i := len(p.recoveryStack) - 1; i >= 0; i-- {
		if recoverExpr, ok := p.recoveryStack[i][expr.label]; ok {
			if val, ok := p.parseExpr(recoverExpr); ok {
				return val, ok
			}
		}
	}

	return nil, false
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
