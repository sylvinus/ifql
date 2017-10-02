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

var g = &grammar{
	rules: []*rule{
		{
			name: "Start",
			pos:  position{line: 6, col: 1, offset: 19},
			expr: &actionExpr{
				pos: position{line: 7, col: 5, offset: 29},
				run: (*parser).callonStart1,
				expr: &seqExpr{
					pos: position{line: 7, col: 5, offset: 29},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 7, col: 5, offset: 29},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 7, col: 8, offset: 32},
							label: "program",
							expr: &ruleRefExpr{
								pos:  position{line: 7, col: 16, offset: 40},
								name: "Program",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 7, col: 24, offset: 48},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Program",
			pos:  position{line: 11, col: 1, offset: 86},
			expr: &actionExpr{
				pos: position{line: 12, col: 5, offset: 98},
				run: (*parser).callonProgram1,
				expr: &labeledExpr{
					pos:   position{line: 12, col: 5, offset: 98},
					label: "body",
					expr: &ruleRefExpr{
						pos:  position{line: 12, col: 10, offset: 103},
						name: "SourceElements",
					},
				},
			},
		},
		{
			name: "SourceElements",
			pos:  position{line: 16, col: 1, offset: 169},
			expr: &actionExpr{
				pos: position{line: 17, col: 5, offset: 188},
				run: (*parser).callonSourceElements1,
				expr: &seqExpr{
					pos: position{line: 17, col: 5, offset: 188},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 17, col: 5, offset: 188},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 17, col: 10, offset: 193},
								name: "SourceElement",
							},
						},
						&labeledExpr{
							pos:   position{line: 17, col: 24, offset: 207},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 17, col: 29, offset: 212},
								expr: &seqExpr{
									pos: position{line: 17, col: 30, offset: 213},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 17, col: 30, offset: 213},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 17, col: 33, offset: 216},
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
			pos:  position{line: 21, col: 1, offset: 275},
			expr: &ruleRefExpr{
				pos:  position{line: 22, col: 5, offset: 293},
				name: "Statement",
			},
		},
		{
			name: "Statement",
			pos:  position{line: 24, col: 1, offset: 304},
			expr: &choiceExpr{
				pos: position{line: 25, col: 5, offset: 318},
				alternatives: []interface{}{
					&labeledExpr{
						pos:   position{line: 25, col: 5, offset: 318},
						label: "varstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 25, col: 13, offset: 326},
							name: "VariableStatement",
						},
					},
					&labeledExpr{
						pos:   position{line: 26, col: 5, offset: 348},
						label: "exprstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 26, col: 14, offset: 357},
							name: "ExpressionStatement",
						},
					},
				},
			},
		},
		{
			name: "VariableStatement",
			pos:  position{line: 28, col: 1, offset: 378},
			expr: &actionExpr{
				pos: position{line: 29, col: 5, offset: 400},
				run: (*parser).callonVariableStatement1,
				expr: &seqExpr{
					pos: position{line: 29, col: 5, offset: 400},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 29, col: 5, offset: 400},
							name: "VarToken",
						},
						&ruleRefExpr{
							pos:  position{line: 29, col: 14, offset: 409},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 29, col: 17, offset: 412},
							label: "declaration",
							expr: &ruleRefExpr{
								pos:  position{line: 29, col: 29, offset: 424},
								name: "VariableDeclaration",
							},
						},
					},
				},
			},
		},
		{
			name: "VarToken",
			pos:  position{line: 33, col: 1, offset: 502},
			expr: &litMatcher{
				pos:        position{line: 33, col: 12, offset: 513},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 35, col: 1, offset: 520},
			expr: &actionExpr{
				pos: position{line: 36, col: 5, offset: 544},
				run: (*parser).callonVariableDeclaration1,
				expr: &seqExpr{
					pos: position{line: 36, col: 5, offset: 544},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 36, col: 5, offset: 544},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 36, col: 8, offset: 547},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 36, col: 19, offset: 558},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 36, col: 22, offset: 561},
							label: "initExpr",
							expr: &ruleRefExpr{
								pos:  position{line: 36, col: 31, offset: 570},
								name: "Initializer",
							},
						},
					},
				},
			},
		},
		{
			name: "Initializer",
			pos:  position{line: 40, col: 1, offset: 641},
			expr: &actionExpr{
				pos: position{line: 41, col: 5, offset: 657},
				run: (*parser).callonInitializer1,
				expr: &seqExpr{
					pos: position{line: 41, col: 5, offset: 657},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 41, col: 5, offset: 657},
							val:        "=",
							ignoreCase: false,
						},
						&notExpr{
							pos: position{line: 41, col: 9, offset: 661},
							expr: &litMatcher{
								pos:        position{line: 41, col: 10, offset: 662},
								val:        "=",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 41, col: 14, offset: 666},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 41, col: 17, offset: 669},
							label: "expression",
							expr: &ruleRefExpr{
								pos:  position{line: 41, col: 28, offset: 680},
								name: "VariableExpression",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 46, col: 1, offset: 807},
			expr: &choiceExpr{
				pos: position{line: 47, col: 5, offset: 830},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 47, col: 5, offset: 830},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 48, col: 5, offset: 849},
						name: "Primary",
					},
				},
			},
		},
		{
			name: "ExpressionStatement",
			pos:  position{line: 50, col: 1, offset: 858},
			expr: &actionExpr{
				pos: position{line: 51, col: 5, offset: 882},
				run: (*parser).callonExpressionStatement1,
				expr: &labeledExpr{
					pos:   position{line: 51, col: 5, offset: 882},
					label: "call",
					expr: &ruleRefExpr{
						pos:  position{line: 51, col: 10, offset: 887},
						name: "CallExpression",
					},
				},
			},
		},
		{
			name: "MemberExpression",
			pos:  position{line: 55, col: 1, offset: 956},
			expr: &actionExpr{
				pos: position{line: 56, col: 5, offset: 977},
				run: (*parser).callonMemberExpression1,
				expr: &seqExpr{
					pos: position{line: 56, col: 5, offset: 977},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 56, col: 5, offset: 977},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 56, col: 10, offset: 982},
								name: "Identifier",
							},
						},
						&labeledExpr{
							pos:   position{line: 57, col: 5, offset: 1024},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 57, col: 10, offset: 1029},
								expr: &actionExpr{
									pos: position{line: 58, col: 9, offset: 1039},
									run: (*parser).callonMemberExpression7,
									expr: &seqExpr{
										pos: position{line: 58, col: 9, offset: 1039},
										exprs: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 58, col: 9, offset: 1039},
												name: "__",
											},
											&litMatcher{
												pos:        position{line: 58, col: 12, offset: 1042},
												val:        ".",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 58, col: 16, offset: 1046},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 58, col: 19, offset: 1049},
												label: "property",
												expr: &ruleRefExpr{
													pos:  position{line: 58, col: 28, offset: 1058},
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
			pos:  position{line: 66, col: 1, offset: 1184},
			expr: &actionExpr{
				pos: position{line: 67, col: 5, offset: 1203},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 67, col: 5, offset: 1203},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 67, col: 5, offset: 1203},
							label: "head",
							expr: &actionExpr{
								pos: position{line: 68, col: 7, offset: 1216},
								run: (*parser).callonCallExpression4,
								expr: &seqExpr{
									pos: position{line: 68, col: 7, offset: 1216},
									exprs: []interface{}{
										&labeledExpr{
											pos:   position{line: 68, col: 7, offset: 1216},
											label: "callee",
											expr: &ruleRefExpr{
												pos:  position{line: 68, col: 14, offset: 1223},
												name: "MemberExpression",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 68, col: 31, offset: 1240},
											name: "__",
										},
										&labeledExpr{
											pos:   position{line: 68, col: 34, offset: 1243},
											label: "args",
											expr: &ruleRefExpr{
												pos:  position{line: 68, col: 39, offset: 1248},
												name: "Arguments",
											},
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 72, col: 5, offset: 1331},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 72, col: 10, offset: 1336},
								expr: &choiceExpr{
									pos: position{line: 73, col: 9, offset: 1346},
									alternatives: []interface{}{
										&actionExpr{
											pos: position{line: 73, col: 9, offset: 1346},
											run: (*parser).callonCallExpression14,
											expr: &seqExpr{
												pos: position{line: 73, col: 9, offset: 1346},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 73, col: 9, offset: 1346},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 73, col: 12, offset: 1349},
														label: "args",
														expr: &ruleRefExpr{
															pos:  position{line: 73, col: 17, offset: 1354},
															name: "Arguments",
														},
													},
												},
											},
										},
										&actionExpr{
											pos: position{line: 76, col: 9, offset: 1436},
											run: (*parser).callonCallExpression19,
											expr: &seqExpr{
												pos: position{line: 76, col: 9, offset: 1436},
												exprs: []interface{}{
													&ruleRefExpr{
														pos:  position{line: 76, col: 9, offset: 1436},
														name: "__",
													},
													&litMatcher{
														pos:        position{line: 76, col: 12, offset: 1439},
														val:        ".",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 76, col: 16, offset: 1443},
														name: "__",
													},
													&labeledExpr{
														pos:   position{line: 76, col: 19, offset: 1446},
														label: "property",
														expr: &ruleRefExpr{
															pos:  position{line: 76, col: 28, offset: 1455},
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
			pos:  position{line: 84, col: 1, offset: 1606},
			expr: &actionExpr{
				pos: position{line: 85, col: 5, offset: 1620},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 85, col: 5, offset: 1620},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 85, col: 5, offset: 1620},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 9, offset: 1624},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 12, offset: 1627},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 85, col: 17, offset: 1632},
								expr: &ruleRefExpr{
									pos:  position{line: 85, col: 18, offset: 1633},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 33, offset: 1648},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 85, col: 36, offset: 1651},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 89, col: 1, offset: 1687},
			expr: &actionExpr{
				pos: position{line: 90, col: 5, offset: 1704},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 90, col: 5, offset: 1704},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 90, col: 5, offset: 1704},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 90, col: 11, offset: 1710},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 90, col: 23, offset: 1722},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 90, col: 26, offset: 1725},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 90, col: 31, offset: 1730},
								expr: &ruleRefExpr{
									pos:  position{line: 90, col: 31, offset: 1730},
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
			pos:  position{line: 94, col: 1, offset: 1805},
			expr: &actionExpr{
				pos: position{line: 95, col: 5, offset: 1826},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 95, col: 5, offset: 1826},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 95, col: 5, offset: 1826},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 95, col: 9, offset: 1830},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 95, col: 13, offset: 1834},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 95, col: 17, offset: 1838},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 99, col: 1, offset: 1881},
			expr: &actionExpr{
				pos: position{line: 100, col: 5, offset: 1897},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 100, col: 5, offset: 1897},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 100, col: 5, offset: 1897},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 100, col: 9, offset: 1901},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 100, col: 20, offset: 1912},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 100, col: 24, offset: 1916},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 100, col: 28, offset: 1920},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 100, col: 31, offset: 1923},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 100, col: 37, offset: 1929},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 104, col: 1, offset: 2005},
			expr: &choiceExpr{
				pos: position{line: 105, col: 5, offset: 2027},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 105, col: 5, offset: 2027},
						name: "LambdaExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 106, col: 5, offset: 2048},
						name: "Primary",
					},
				},
			},
		},
		{
			name: "LambdaExpression",
			pos:  position{line: 108, col: 1, offset: 2057},
			expr: &actionExpr{
				pos: position{line: 109, col: 5, offset: 2078},
				run: (*parser).callonLambdaExpression1,
				expr: &seqExpr{
					pos: position{line: 109, col: 5, offset: 2078},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 109, col: 5, offset: 2078},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 109, col: 9, offset: 2082},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 109, col: 12, offset: 2085},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 17, offset: 2090},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 109, col: 22, offset: 2095},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 109, col: 26, offset: 2099},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 120, col: 1, offset: 2332},
			expr: &ruleRefExpr{
				pos:  position{line: 121, col: 5, offset: 2341},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 123, col: 1, offset: 2350},
			expr: &actionExpr{
				pos: position{line: 124, col: 5, offset: 2371},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 124, col: 6, offset: 2372},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 124, col: 6, offset: 2372},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 124, col: 14, offset: 2380},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 128, col: 1, offset: 2432},
			expr: &actionExpr{
				pos: position{line: 129, col: 5, offset: 2444},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 129, col: 5, offset: 2444},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 129, col: 5, offset: 2444},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 129, col: 10, offset: 2449},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 129, col: 19, offset: 2458},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 129, col: 24, offset: 2463},
								expr: &seqExpr{
									pos: position{line: 129, col: 26, offset: 2465},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 129, col: 26, offset: 2465},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 129, col: 30, offset: 2469},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 129, col: 47, offset: 2486},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 129, col: 51, offset: 2490},
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
			pos:  position{line: 133, col: 1, offset: 2569},
			expr: &actionExpr{
				pos: position{line: 134, col: 5, offset: 2591},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 134, col: 6, offset: 2592},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 134, col: 6, offset: 2592},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 134, col: 13, offset: 2599},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 138, col: 1, offset: 2645},
			expr: &actionExpr{
				pos: position{line: 139, col: 5, offset: 2658},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 139, col: 5, offset: 2658},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 139, col: 5, offset: 2658},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 139, col: 10, offset: 2663},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 139, col: 21, offset: 2674},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 139, col: 26, offset: 2679},
								expr: &seqExpr{
									pos: position{line: 139, col: 28, offset: 2681},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 139, col: 28, offset: 2681},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 139, col: 31, offset: 2684},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 139, col: 49, offset: 2702},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 139, col: 52, offset: 2705},
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
			pos:  position{line: 143, col: 1, offset: 2785},
			expr: &actionExpr{
				pos: position{line: 144, col: 5, offset: 2809},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 144, col: 9, offset: 2813},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 144, col: 9, offset: 2813},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 145, col: 9, offset: 2826},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 146, col: 9, offset: 2838},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 147, col: 9, offset: 2851},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 148, col: 9, offset: 2863},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 149, col: 9, offset: 2885},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 150, col: 9, offset: 2899},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 151, col: 9, offset: 2920},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 156, col: 1, offset: 2978},
			expr: &actionExpr{
				pos: position{line: 157, col: 5, offset: 2993},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 157, col: 5, offset: 2993},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 157, col: 5, offset: 2993},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 157, col: 10, offset: 2998},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 157, col: 19, offset: 3007},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 157, col: 24, offset: 3012},
								expr: &seqExpr{
									pos: position{line: 157, col: 26, offset: 3014},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 157, col: 26, offset: 3014},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 29, offset: 3017},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 49, offset: 3037},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 52, offset: 3040},
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
			pos:  position{line: 161, col: 1, offset: 3118},
			expr: &actionExpr{
				pos: position{line: 162, col: 5, offset: 3139},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 162, col: 6, offset: 3140},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 162, col: 6, offset: 3140},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 162, col: 12, offset: 3146},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 166, col: 1, offset: 3194},
			expr: &actionExpr{
				pos: position{line: 167, col: 5, offset: 3207},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 167, col: 5, offset: 3207},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 167, col: 5, offset: 3207},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 167, col: 10, offset: 3212},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 167, col: 25, offset: 3227},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 167, col: 30, offset: 3232},
								expr: &seqExpr{
									pos: position{line: 167, col: 32, offset: 3234},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 167, col: 32, offset: 3234},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 167, col: 35, offset: 3237},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 167, col: 52, offset: 3254},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 167, col: 55, offset: 3257},
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
			pos:  position{line: 171, col: 1, offset: 3341},
			expr: &actionExpr{
				pos: position{line: 172, col: 5, offset: 3368},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 172, col: 6, offset: 3369},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 172, col: 6, offset: 3369},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 172, col: 12, offset: 3375},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 176, col: 1, offset: 3419},
			expr: &actionExpr{
				pos: position{line: 177, col: 5, offset: 3438},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 177, col: 5, offset: 3438},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 177, col: 5, offset: 3438},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 10, offset: 3443},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 177, col: 18, offset: 3451},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 177, col: 23, offset: 3456},
								expr: &seqExpr{
									pos: position{line: 177, col: 25, offset: 3458},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 177, col: 25, offset: 3458},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 177, col: 28, offset: 3461},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 177, col: 51, offset: 3484},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 177, col: 54, offset: 3487},
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
			pos:  position{line: 181, col: 1, offset: 3564},
			expr: &choiceExpr{
				pos: position{line: 182, col: 5, offset: 3576},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 182, col: 5, offset: 3576},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 182, col: 5, offset: 3576},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 182, col: 5, offset: 3576},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 182, col: 9, offset: 3580},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 182, col: 12, offset: 3583},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 182, col: 17, offset: 3588},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 182, col: 25, offset: 3596},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 182, col: 28, offset: 3599},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 185, col: 5, offset: 3638},
						name: "Array",
					},
					&ruleRefExpr{
						pos:  position{line: 186, col: 5, offset: 3648},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 187, col: 5, offset: 3666},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 188, col: 5, offset: 3695},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 189, col: 5, offset: 3708},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 190, col: 5, offset: 3721},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 191, col: 5, offset: 3732},
						name: "Integer",
					},
					&ruleRefExpr{
						pos:  position{line: 192, col: 5, offset: 3744},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 193, col: 5, offset: 3754},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "Array",
			pos:  position{line: 196, col: 1, offset: 3767},
			expr: &actionExpr{
				pos: position{line: 197, col: 5, offset: 3777},
				run: (*parser).callonArray1,
				expr: &seqExpr{
					pos: position{line: 197, col: 5, offset: 3777},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 197, col: 5, offset: 3777},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 197, col: 9, offset: 3781},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 197, col: 12, offset: 3784},
							label: "elements",
							expr: &zeroOrOneExpr{
								pos: position{line: 197, col: 21, offset: 3793},
								expr: &ruleRefExpr{
									pos:  position{line: 197, col: 21, offset: 3793},
									name: "ArrayElements",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 197, col: 36, offset: 3808},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 197, col: 39, offset: 3811},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ArrayElements",
			pos:  position{line: 201, col: 1, offset: 3851},
			expr: &actionExpr{
				pos: position{line: 202, col: 5, offset: 3869},
				run: (*parser).callonArrayElements1,
				expr: &seqExpr{
					pos: position{line: 202, col: 5, offset: 3869},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 202, col: 5, offset: 3869},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 202, col: 11, offset: 3875},
								name: "Primary",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 202, col: 19, offset: 3883},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 202, col: 22, offset: 3886},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 202, col: 27, offset: 3891},
								expr: &ruleRefExpr{
									pos:  position{line: 202, col: 27, offset: 3891},
									name: "ArrayRest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ArrayRest",
			pos:  position{line: 206, col: 1, offset: 3963},
			expr: &actionExpr{
				pos: position{line: 207, col: 5, offset: 3977},
				run: (*parser).callonArrayRest1,
				expr: &seqExpr{
					pos: position{line: 207, col: 5, offset: 3977},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 207, col: 5, offset: 3977},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 207, col: 9, offset: 3981},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 207, col: 12, offset: 3984},
							label: "element",
							expr: &ruleRefExpr{
								pos:  position{line: 207, col: 20, offset: 3992},
								name: "Primary",
							},
						},
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 211, col: 1, offset: 4035},
			expr: &seqExpr{
				pos: position{line: 212, col: 5, offset: 4052},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 212, col: 5, offset: 4052},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 212, col: 11, offset: 4058},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 212, col: 17, offset: 4064},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 212, col: 23, offset: 4070},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 214, col: 1, offset: 4077},
			expr: &seqExpr{
				pos: position{line: 216, col: 5, offset: 4102},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 216, col: 5, offset: 4102},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 216, col: 11, offset: 4108},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 218, col: 1, offset: 4115},
			expr: &seqExpr{
				pos: position{line: 221, col: 5, offset: 4185},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 221, col: 5, offset: 4185},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 221, col: 11, offset: 4191},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 223, col: 1, offset: 4198},
			expr: &seqExpr{
				pos: position{line: 225, col: 5, offset: 4222},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 225, col: 5, offset: 4222},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 225, col: 11, offset: 4228},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 227, col: 1, offset: 4235},
			expr: &seqExpr{
				pos: position{line: 229, col: 5, offset: 4261},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 229, col: 5, offset: 4261},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 229, col: 11, offset: 4267},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 231, col: 1, offset: 4274},
			expr: &seqExpr{
				pos: position{line: 234, col: 5, offset: 4346},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 234, col: 5, offset: 4346},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 234, col: 11, offset: 4352},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 236, col: 1, offset: 4359},
			expr: &seqExpr{
				pos: position{line: 237, col: 5, offset: 4375},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 237, col: 5, offset: 4375},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 237, col: 9, offset: 4379},
						expr: &ruleRefExpr{
							pos:  position{line: 237, col: 9, offset: 4379},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 239, col: 1, offset: 4387},
			expr: &seqExpr{
				pos: position{line: 240, col: 5, offset: 4405},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 240, col: 6, offset: 4406},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 240, col: 6, offset: 4406},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 240, col: 12, offset: 4412},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 240, col: 17, offset: 4417},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 240, col: 26, offset: 4426},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 240, col: 30, offset: 4430},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 242, col: 1, offset: 4442},
			expr: &choiceExpr{
				pos: position{line: 243, col: 6, offset: 4458},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 243, col: 6, offset: 4458},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 243, col: 12, offset: 4464},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 245, col: 1, offset: 4480},
			expr: &seqExpr{
				pos: position{line: 246, col: 5, offset: 4496},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 246, col: 5, offset: 4496},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 246, col: 14, offset: 4505},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 246, col: 18, offset: 4509},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 246, col: 29, offset: 4520},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 246, col: 33, offset: 4524},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 246, col: 44, offset: 4535},
						expr: &ruleRefExpr{
							pos:  position{line: 246, col: 44, offset: 4535},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 248, col: 1, offset: 4549},
			expr: &seqExpr{
				pos: position{line: 249, col: 5, offset: 4562},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 249, col: 5, offset: 4562},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 249, col: 18, offset: 4575},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 249, col: 22, offset: 4579},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 249, col: 32, offset: 4589},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 249, col: 36, offset: 4593},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 251, col: 1, offset: 4603},
			expr: &seqExpr{
				pos: position{line: 252, col: 5, offset: 4616},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 252, col: 5, offset: 4616},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 252, col: 17, offset: 4628},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 254, col: 1, offset: 4640},
			expr: &actionExpr{
				pos: position{line: 255, col: 5, offset: 4653},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 255, col: 5, offset: 4653},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 255, col: 5, offset: 4653},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 255, col: 14, offset: 4662},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 255, col: 18, offset: 4666},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 259, col: 1, offset: 4721},
			expr: &litMatcher{
				pos:        position{line: 260, col: 5, offset: 4741},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 262, col: 1, offset: 4747},
			expr: &choiceExpr{
				pos: position{line: 263, col: 6, offset: 4769},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 263, col: 6, offset: 4769},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 263, col: 13, offset: 4776},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 263, col: 20, offset: 4784},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 265, col: 1, offset: 4792},
			expr: &litMatcher{
				pos:        position{line: 266, col: 5, offset: 4813},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 268, col: 1, offset: 4819},
			expr: &litMatcher{
				pos:        position{line: 269, col: 5, offset: 4835},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 271, col: 1, offset: 4840},
			expr: &litMatcher{
				pos:        position{line: 272, col: 5, offset: 4856},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 274, col: 1, offset: 4861},
			expr: &litMatcher{
				pos:        position{line: 275, col: 5, offset: 4875},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 277, col: 1, offset: 4880},
			expr: &choiceExpr{
				pos: position{line: 279, col: 9, offset: 4908},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 279, col: 9, offset: 4908},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 280, col: 9, offset: 4932},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 281, col: 9, offset: 4957},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 282, col: 9, offset: 4982},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 283, col: 9, offset: 5002},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 284, col: 9, offset: 5022},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 287, col: 1, offset: 5039},
			expr: &seqExpr{
				pos: position{line: 288, col: 5, offset: 5058},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 288, col: 5, offset: 5058},
						name: "Integer",
					},
					&ruleRefExpr{
						pos:  position{line: 288, col: 13, offset: 5066},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 290, col: 1, offset: 5081},
			expr: &actionExpr{
				pos: position{line: 291, col: 5, offset: 5094},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 291, col: 5, offset: 5094},
					expr: &ruleRefExpr{
						pos:  position{line: 291, col: 5, offset: 5094},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 295, col: 1, offset: 5163},
			expr: &choiceExpr{
				pos: position{line: 296, col: 5, offset: 5181},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 296, col: 5, offset: 5181},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 296, col: 7, offset: 5183},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 296, col: 7, offset: 5183},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 296, col: 11, offset: 5187},
									expr: &ruleRefExpr{
										pos:  position{line: 296, col: 11, offset: 5187},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 296, col: 29, offset: 5205},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 299, col: 5, offset: 5269},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 299, col: 7, offset: 5271},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 299, col: 7, offset: 5271},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 299, col: 11, offset: 5275},
									expr: &ruleRefExpr{
										pos:  position{line: 299, col: 11, offset: 5275},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 299, col: 31, offset: 5295},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 299, col: 31, offset: 5295},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 299, col: 37, offset: 5301},
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
			pos:  position{line: 303, col: 1, offset: 5383},
			expr: &choiceExpr{
				pos: position{line: 304, col: 5, offset: 5404},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 304, col: 5, offset: 5404},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 304, col: 5, offset: 5404},
								expr: &choiceExpr{
									pos: position{line: 304, col: 8, offset: 5407},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 304, col: 8, offset: 5407},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 304, col: 14, offset: 5413},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 304, col: 21, offset: 5420},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 304, col: 27, offset: 5426},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 305, col: 5, offset: 5441},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 305, col: 5, offset: 5441},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 305, col: 10, offset: 5446},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 307, col: 1, offset: 5466},
			expr: &choiceExpr{
				pos: position{line: 308, col: 5, offset: 5489},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 308, col: 5, offset: 5489},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 309, col: 5, offset: 5497},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 309, col: 7, offset: 5499},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 309, col: 7, offset: 5499},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 309, col: 20, offset: 5512},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 309, col: 26, offset: 5518},
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
			pos:  position{line: 313, col: 1, offset: 5590},
			expr: &actionExpr{
				pos: position{line: 315, col: 5, offset: 5635},
				run: (*parser).callonIdentifier1,
				expr: &oneOrMoreExpr{
					pos: position{line: 315, col: 5, offset: 5635},
					expr: &charClassMatcher{
						pos:        position{line: 315, col: 5, offset: 5635},
						val:        "[\\pL]",
						classes:    []*unicode.RangeTable{rangeTable("L")},
						ignoreCase: false,
						inverted:   false,
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 319, col: 1, offset: 5690},
			expr: &actionExpr{
				pos: position{line: 320, col: 5, offset: 5701},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 320, col: 5, offset: 5701},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 320, col: 5, offset: 5701},
							name: "Integer",
						},
						&litMatcher{
							pos:        position{line: 320, col: 13, offset: 5709},
							val:        ".",
							ignoreCase: false,
						},
						&oneOrMoreExpr{
							pos: position{line: 320, col: 17, offset: 5713},
							expr: &ruleRefExpr{
								pos:  position{line: 320, col: 17, offset: 5713},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "Integer",
			pos:  position{line: 324, col: 1, offset: 5770},
			expr: &actionExpr{
				pos: position{line: 325, col: 5, offset: 5782},
				run: (*parser).callonInteger1,
				expr: &choiceExpr{
					pos: position{line: 325, col: 6, offset: 5783},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 325, col: 6, offset: 5783},
							val:        "0",
							ignoreCase: false,
						},
						&seqExpr{
							pos: position{line: 325, col: 12, offset: 5789},
							exprs: []interface{}{
								&zeroOrOneExpr{
									pos: position{line: 325, col: 12, offset: 5789},
									expr: &litMatcher{
										pos:        position{line: 325, col: 12, offset: 5789},
										val:        "-",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 325, col: 17, offset: 5794},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 325, col: 30, offset: 5807},
									expr: &ruleRefExpr{
										pos:  position{line: 325, col: 30, offset: 5807},
										name: "Digit",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "NonZeroDigit",
			pos:  position{line: 330, col: 1, offset: 5867},
			expr: &charClassMatcher{
				pos:        position{line: 331, col: 5, offset: 5884},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 333, col: 1, offset: 5891},
			expr: &charClassMatcher{
				pos:        position{line: 334, col: 5, offset: 5901},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 336, col: 1, offset: 5908},
			expr: &actionExpr{
				pos: position{line: 337, col: 5, offset: 5918},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 337, col: 5, offset: 5918},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 337, col: 11, offset: 5924},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 341, col: 1, offset: 5978},
			expr: &actionExpr{
				pos: position{line: 342, col: 5, offset: 6028},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 342, col: 5, offset: 6028},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 342, col: 5, offset: 6028},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 342, col: 9, offset: 6032},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 342, col: 17, offset: 6040},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 342, col: 39, offset: 6062},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 346, col: 1, offset: 6104},
			expr: &actionExpr{
				pos: position{line: 347, col: 5, offset: 6130},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 347, col: 5, offset: 6130},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 347, col: 11, offset: 6136},
						expr: &ruleRefExpr{
							pos:  position{line: 347, col: 11, offset: 6136},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 351, col: 1, offset: 6214},
			expr: &choiceExpr{
				pos: position{line: 352, col: 5, offset: 6240},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 352, col: 5, offset: 6240},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 352, col: 5, offset: 6240},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 352, col: 5, offset: 6240},
									expr: &charClassMatcher{
										pos:        position{line: 352, col: 6, offset: 6241},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 352, col: 12, offset: 6247},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 352, col: 15, offset: 6250},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 355, col: 5, offset: 6312},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 357, col: 1, offset: 6348},
			expr: &choiceExpr{
				pos: position{line: 358, col: 5, offset: 6387},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 358, col: 5, offset: 6387},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 358, col: 5, offset: 6387},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 361, col: 5, offset: 6425},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 361, col: 5, offset: 6425},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 361, col: 10, offset: 6430},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 363, col: 1, offset: 6462},
			expr: &actionExpr{
				pos: position{line: 364, col: 5, offset: 6497},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 364, col: 5, offset: 6497},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 364, col: 5, offset: 6497},
							expr: &ruleRefExpr{
								pos:  position{line: 364, col: 6, offset: 6498},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 364, col: 21, offset: 6513},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 368, col: 1, offset: 6564},
			expr: &anyMatcher{
				line: 369, col: 5, offset: 6579,
			},
		},
		{
			name: "__",
			pos:  position{line: 371, col: 1, offset: 6582},
			expr: &zeroOrMoreExpr{
				pos: position{line: 372, col: 5, offset: 6589},
				expr: &choiceExpr{
					pos: position{line: 372, col: 7, offset: 6591},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 372, col: 7, offset: 6591},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 372, col: 12, offset: 6596},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 374, col: 1, offset: 6604},
			expr: &charClassMatcher{
				pos:        position{line: 375, col: 5, offset: 6611},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 377, col: 1, offset: 6622},
			expr: &charClassMatcher{
				pos:        position{line: 378, col: 5, offset: 6641},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 380, col: 1, offset: 6649},
			expr: &litMatcher{
				pos:        position{line: 381, col: 5, offset: 6657},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 383, col: 1, offset: 6663},
			expr: &notExpr{
				pos: position{line: 384, col: 5, offset: 6671},
				expr: &anyMatcher{
					line: 384, col: 6, offset: 6672,
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

func (c *current) onLambdaExpression1(expr interface{}) (interface{}, error) {
	return function(expr), nil

}

func (p *parser) callonLambdaExpression1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLambdaExpression1(stack["expr"])
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

func (c *current) onArray1(elements interface{}) (interface{}, error) {
	return elements, nil

}

func (p *parser) callonArray1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onArray1(stack["elements"])
}

func (c *current) onArrayElements1(first, rest interface{}) (interface{}, error) {
	return array(first, rest, c.text, c.pos), nil

}

func (p *parser) callonArrayElements1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onArrayElements1(stack["first"], stack["rest"])
}

func (c *current) onArrayRest1(element interface{}) (interface{}, error) {
	return element, nil

}

func (p *parser) callonArrayRest1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onArrayRest1(stack["element"])
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

func (c *current) onInteger1() (interface{}, error) {
	return integerLiteral(c.text, c.pos)

}

func (p *parser) callonInteger1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInteger1()
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

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
