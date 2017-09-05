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
			pos:  position{line: 22, col: 1, offset: 305},
			expr: &ruleRefExpr{
				pos:  position{line: 23, col: 5, offset: 323},
				name: "Statement",
			},
		},
		{
			name: "Statement",
			pos:  position{line: 25, col: 1, offset: 334},
			expr: &choiceExpr{
				pos: position{line: 26, col: 5, offset: 348},
				alternatives: []interface{}{
					&labeledExpr{
						pos:   position{line: 26, col: 5, offset: 348},
						label: "varstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 26, col: 13, offset: 356},
							name: "VariableStatement",
						},
					},
					&labeledExpr{
						pos:   position{line: 27, col: 5, offset: 378},
						label: "exprstmt",
						expr: &ruleRefExpr{
							pos:  position{line: 27, col: 14, offset: 387},
							name: "ExpressionStatement",
						},
					},
				},
			},
		},
		{
			name: "VariableStatement",
			pos:  position{line: 29, col: 1, offset: 408},
			expr: &actionExpr{
				pos: position{line: 30, col: 5, offset: 430},
				run: (*parser).callonVariableStatement1,
				expr: &seqExpr{
					pos: position{line: 30, col: 5, offset: 430},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 30, col: 5, offset: 430},
							name: "VarToken",
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 14, offset: 439},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 30, col: 17, offset: 442},
							label: "declarations",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 30, offset: 455},
								name: "VariableDeclarationList",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableDeclarationList",
			pos:  position{line: 34, col: 1, offset: 538},
			expr: &actionExpr{
				pos: position{line: 35, col: 5, offset: 566},
				run: (*parser).callonVariableDeclarationList1,
				expr: &seqExpr{
					pos: position{line: 35, col: 5, offset: 566},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 35, col: 5, offset: 566},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 35, col: 10, offset: 571},
								name: "VariableDeclaration",
							},
						},
						&labeledExpr{
							pos:   position{line: 35, col: 30, offset: 591},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 35, col: 35, offset: 596},
								expr: &seqExpr{
									pos: position{line: 35, col: 36, offset: 597},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 35, col: 36, offset: 597},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 35, col: 39, offset: 600},
											val:        ",",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 35, col: 43, offset: 604},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 35, col: 46, offset: 607},
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
			pos:  position{line: 39, col: 1, offset: 674},
			expr: &litMatcher{
				pos:        position{line: 39, col: 12, offset: 685},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 41, col: 1, offset: 692},
			expr: &actionExpr{
				pos: position{line: 42, col: 5, offset: 716},
				run: (*parser).callonVariableDeclaration1,
				expr: &seqExpr{
					pos: position{line: 42, col: 5, offset: 716},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 42, col: 5, offset: 716},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 8, offset: 719},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 42, col: 15, offset: 726},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 42, col: 18, offset: 729},
							label: "initExpr",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 27, offset: 738},
								name: "Initializer",
							},
						},
					},
				},
			},
		},
		{
			name: "Initializer",
			pos:  position{line: 46, col: 1, offset: 809},
			expr: &actionExpr{
				pos: position{line: 47, col: 5, offset: 825},
				run: (*parser).callonInitializer1,
				expr: &seqExpr{
					pos: position{line: 47, col: 5, offset: 825},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 47, col: 5, offset: 825},
							val:        "=",
							ignoreCase: false,
						},
						&notExpr{
							pos: position{line: 47, col: 9, offset: 829},
							expr: &litMatcher{
								pos:        position{line: 47, col: 10, offset: 830},
								val:        "=",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 47, col: 14, offset: 834},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 47, col: 17, offset: 837},
							label: "expression",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 28, offset: 848},
								name: "VariableExpression",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 52, col: 1, offset: 975},
			expr: &choiceExpr{
				pos: position{line: 53, col: 5, offset: 998},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 53, col: 5, offset: 998},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 54, col: 5, offset: 1017},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 5, offset: 1035},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 56, col: 5, offset: 1064},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 5, offset: 1077},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 58, col: 5, offset: 1090},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 59, col: 5, offset: 1101},
						name: "Field",
					},
				},
			},
		},
		{
			name: "ExpressionStatement",
			pos:  position{line: 63, col: 1, offset: 1164},
			expr: &actionExpr{
				pos: position{line: 64, col: 5, offset: 1188},
				run: (*parser).callonExpressionStatement1,
				expr: &labeledExpr{
					pos:   position{line: 64, col: 5, offset: 1188},
					label: "call",
					expr: &ruleRefExpr{
						pos:  position{line: 64, col: 10, offset: 1193},
						name: "CallExpression",
					},
				},
			},
		},
		{
			name: "CallExpression",
			pos:  position{line: 68, col: 1, offset: 1262},
			expr: &actionExpr{
				pos: position{line: 69, col: 5, offset: 1281},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 69, col: 5, offset: 1281},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 69, col: 5, offset: 1281},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 69, col: 12, offset: 1288},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 69, col: 19, offset: 1295},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 69, col: 22, offset: 1298},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 69, col: 27, offset: 1303},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 69, col: 37, offset: 1313},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 69, col: 45, offset: 1321},
								expr: &seqExpr{
									pos: position{line: 69, col: 47, offset: 1323},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 69, col: 47, offset: 1323},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 69, col: 50, offset: 1326},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 54, offset: 1330},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 57, offset: 1333},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 64, offset: 1340},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 67, offset: 1343},
											name: "Arguments",
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
			pos:  position{line: 73, col: 1, offset: 1425},
			expr: &actionExpr{
				pos: position{line: 74, col: 5, offset: 1439},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 74, col: 5, offset: 1439},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 74, col: 5, offset: 1439},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 74, col: 9, offset: 1443},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 74, col: 12, offset: 1446},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 74, col: 17, offset: 1451},
								expr: &ruleRefExpr{
									pos:  position{line: 74, col: 18, offset: 1452},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 74, col: 33, offset: 1467},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 74, col: 36, offset: 1470},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 78, col: 1, offset: 1506},
			expr: &actionExpr{
				pos: position{line: 79, col: 5, offset: 1523},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 79, col: 5, offset: 1523},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 79, col: 5, offset: 1523},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 79, col: 11, offset: 1529},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 79, col: 23, offset: 1541},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 79, col: 26, offset: 1544},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 79, col: 31, offset: 1549},
								expr: &ruleRefExpr{
									pos:  position{line: 79, col: 31, offset: 1549},
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
			pos:  position{line: 83, col: 1, offset: 1624},
			expr: &actionExpr{
				pos: position{line: 84, col: 5, offset: 1645},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 84, col: 5, offset: 1645},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 84, col: 5, offset: 1645},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 84, col: 9, offset: 1649},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 84, col: 13, offset: 1653},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 84, col: 17, offset: 1657},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 88, col: 1, offset: 1700},
			expr: &actionExpr{
				pos: position{line: 89, col: 5, offset: 1716},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 89, col: 5, offset: 1716},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 89, col: 5, offset: 1716},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 9, offset: 1720},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 89, col: 16, offset: 1727},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 89, col: 20, offset: 1731},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 89, col: 24, offset: 1735},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 89, col: 27, offset: 1738},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 33, offset: 1744},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 93, col: 1, offset: 1820},
			expr: &choiceExpr{
				pos: position{line: 94, col: 5, offset: 1842},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 94, col: 5, offset: 1842},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 95, col: 5, offset: 1856},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 96, col: 5, offset: 1874},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 97, col: 5, offset: 1903},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 98, col: 5, offset: 1916},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 99, col: 5, offset: 1929},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 100, col: 5, offset: 1940},
						name: "String",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 102, col: 1, offset: 1948},
			expr: &actionExpr{
				pos: position{line: 103, col: 5, offset: 1962},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 103, col: 5, offset: 1962},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 103, col: 5, offset: 1962},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 103, col: 9, offset: 1966},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 103, col: 12, offset: 1969},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 103, col: 17, offset: 1974},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 103, col: 22, offset: 1979},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 103, col: 26, offset: 1983},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 114, col: 1, offset: 2206},
			expr: &ruleRefExpr{
				pos:  position{line: 115, col: 5, offset: 2215},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 117, col: 1, offset: 2224},
			expr: &actionExpr{
				pos: position{line: 118, col: 5, offset: 2245},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 118, col: 6, offset: 2246},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 118, col: 6, offset: 2246},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 118, col: 14, offset: 2254},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 122, col: 1, offset: 2306},
			expr: &actionExpr{
				pos: position{line: 123, col: 5, offset: 2318},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 123, col: 5, offset: 2318},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 123, col: 5, offset: 2318},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 123, col: 10, offset: 2323},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 123, col: 19, offset: 2332},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 123, col: 24, offset: 2337},
								expr: &seqExpr{
									pos: position{line: 123, col: 26, offset: 2339},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 123, col: 26, offset: 2339},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 123, col: 30, offset: 2343},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 123, col: 47, offset: 2360},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 123, col: 51, offset: 2364},
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
			pos:  position{line: 127, col: 1, offset: 2443},
			expr: &actionExpr{
				pos: position{line: 128, col: 5, offset: 2465},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 128, col: 6, offset: 2466},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 128, col: 6, offset: 2466},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 128, col: 13, offset: 2473},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 132, col: 1, offset: 2519},
			expr: &actionExpr{
				pos: position{line: 133, col: 5, offset: 2532},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 133, col: 5, offset: 2532},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 133, col: 5, offset: 2532},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 133, col: 10, offset: 2537},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 133, col: 21, offset: 2548},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 133, col: 26, offset: 2553},
								expr: &seqExpr{
									pos: position{line: 133, col: 28, offset: 2555},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 133, col: 28, offset: 2555},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 133, col: 31, offset: 2558},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 133, col: 49, offset: 2576},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 133, col: 52, offset: 2579},
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
			pos:  position{line: 137, col: 1, offset: 2659},
			expr: &actionExpr{
				pos: position{line: 138, col: 5, offset: 2683},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 138, col: 9, offset: 2687},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 138, col: 9, offset: 2687},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 139, col: 9, offset: 2700},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 140, col: 9, offset: 2712},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 141, col: 9, offset: 2725},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 142, col: 9, offset: 2737},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 143, col: 9, offset: 2759},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 144, col: 9, offset: 2773},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 145, col: 9, offset: 2794},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 150, col: 1, offset: 2852},
			expr: &actionExpr{
				pos: position{line: 151, col: 5, offset: 2867},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 151, col: 5, offset: 2867},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 151, col: 5, offset: 2867},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 151, col: 10, offset: 2872},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 151, col: 19, offset: 2881},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 151, col: 24, offset: 2886},
								expr: &seqExpr{
									pos: position{line: 151, col: 26, offset: 2888},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 151, col: 26, offset: 2888},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 29, offset: 2891},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 49, offset: 2911},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 52, offset: 2914},
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
			pos:  position{line: 155, col: 1, offset: 2992},
			expr: &actionExpr{
				pos: position{line: 156, col: 5, offset: 3013},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 156, col: 6, offset: 3014},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 156, col: 6, offset: 3014},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 156, col: 12, offset: 3020},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 160, col: 1, offset: 3068},
			expr: &actionExpr{
				pos: position{line: 161, col: 5, offset: 3081},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 161, col: 5, offset: 3081},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 161, col: 5, offset: 3081},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 161, col: 10, offset: 3086},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 161, col: 25, offset: 3101},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 161, col: 30, offset: 3106},
								expr: &seqExpr{
									pos: position{line: 161, col: 32, offset: 3108},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 161, col: 32, offset: 3108},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 161, col: 35, offset: 3111},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 161, col: 52, offset: 3128},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 161, col: 55, offset: 3131},
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
			pos:  position{line: 165, col: 1, offset: 3216},
			expr: &actionExpr{
				pos: position{line: 166, col: 5, offset: 3243},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 166, col: 6, offset: 3244},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 6, offset: 3244},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 166, col: 12, offset: 3250},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 170, col: 1, offset: 3294},
			expr: &actionExpr{
				pos: position{line: 171, col: 5, offset: 3313},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 171, col: 5, offset: 3313},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 171, col: 5, offset: 3313},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 10, offset: 3318},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 171, col: 18, offset: 3326},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 171, col: 23, offset: 3331},
								expr: &seqExpr{
									pos: position{line: 171, col: 25, offset: 3333},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 171, col: 25, offset: 3333},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 171, col: 28, offset: 3336},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 171, col: 51, offset: 3359},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 171, col: 54, offset: 3362},
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
			pos:  position{line: 175, col: 1, offset: 3439},
			expr: &choiceExpr{
				pos: position{line: 176, col: 5, offset: 3451},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 176, col: 5, offset: 3451},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 176, col: 5, offset: 3451},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 176, col: 5, offset: 3451},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 176, col: 9, offset: 3455},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 176, col: 12, offset: 3458},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 176, col: 17, offset: 3463},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 176, col: 25, offset: 3471},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 176, col: 28, offset: 3474},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 179, col: 5, offset: 3513},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 180, col: 5, offset: 3531},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 181, col: 5, offset: 3560},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 182, col: 5, offset: 3573},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 183, col: 5, offset: 3586},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 184, col: 5, offset: 3597},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 185, col: 5, offset: 3607},
						name: "String",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 187, col: 1, offset: 3615},
			expr: &seqExpr{
				pos: position{line: 188, col: 5, offset: 3632},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 188, col: 5, offset: 3632},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 188, col: 11, offset: 3638},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 188, col: 17, offset: 3644},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 188, col: 23, offset: 3650},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 190, col: 1, offset: 3657},
			expr: &seqExpr{
				pos: position{line: 192, col: 5, offset: 3682},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 192, col: 5, offset: 3682},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 192, col: 11, offset: 3688},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 194, col: 1, offset: 3695},
			expr: &seqExpr{
				pos: position{line: 197, col: 5, offset: 3765},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 197, col: 5, offset: 3765},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 197, col: 11, offset: 3771},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 199, col: 1, offset: 3778},
			expr: &seqExpr{
				pos: position{line: 201, col: 5, offset: 3802},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 201, col: 5, offset: 3802},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 201, col: 11, offset: 3808},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 203, col: 1, offset: 3815},
			expr: &seqExpr{
				pos: position{line: 205, col: 5, offset: 3841},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 205, col: 5, offset: 3841},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 11, offset: 3847},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 207, col: 1, offset: 3854},
			expr: &seqExpr{
				pos: position{line: 210, col: 5, offset: 3926},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 210, col: 5, offset: 3926},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 210, col: 11, offset: 3932},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 212, col: 1, offset: 3939},
			expr: &seqExpr{
				pos: position{line: 213, col: 5, offset: 3955},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 213, col: 5, offset: 3955},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 213, col: 9, offset: 3959},
						expr: &ruleRefExpr{
							pos:  position{line: 213, col: 9, offset: 3959},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 215, col: 1, offset: 3967},
			expr: &seqExpr{
				pos: position{line: 216, col: 5, offset: 3985},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 216, col: 6, offset: 3986},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 216, col: 6, offset: 3986},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 216, col: 12, offset: 3992},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 216, col: 17, offset: 3997},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 216, col: 26, offset: 4006},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 216, col: 30, offset: 4010},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 218, col: 1, offset: 4022},
			expr: &choiceExpr{
				pos: position{line: 219, col: 6, offset: 4038},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 219, col: 6, offset: 4038},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 219, col: 12, offset: 4044},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 221, col: 1, offset: 4060},
			expr: &seqExpr{
				pos: position{line: 222, col: 5, offset: 4076},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 222, col: 5, offset: 4076},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 222, col: 14, offset: 4085},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 222, col: 18, offset: 4089},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 222, col: 29, offset: 4100},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 222, col: 33, offset: 4104},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 222, col: 44, offset: 4115},
						expr: &ruleRefExpr{
							pos:  position{line: 222, col: 44, offset: 4115},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 224, col: 1, offset: 4129},
			expr: &seqExpr{
				pos: position{line: 225, col: 5, offset: 4142},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 225, col: 5, offset: 4142},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 225, col: 18, offset: 4155},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 225, col: 22, offset: 4159},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 225, col: 32, offset: 4169},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 225, col: 36, offset: 4173},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 227, col: 1, offset: 4183},
			expr: &seqExpr{
				pos: position{line: 228, col: 5, offset: 4196},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 228, col: 5, offset: 4196},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 228, col: 17, offset: 4208},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 230, col: 1, offset: 4220},
			expr: &actionExpr{
				pos: position{line: 231, col: 5, offset: 4233},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 231, col: 5, offset: 4233},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 231, col: 5, offset: 4233},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 231, col: 14, offset: 4242},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 18, offset: 4246},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 235, col: 1, offset: 4301},
			expr: &litMatcher{
				pos:        position{line: 236, col: 5, offset: 4321},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 238, col: 1, offset: 4327},
			expr: &choiceExpr{
				pos: position{line: 239, col: 6, offset: 4349},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 239, col: 6, offset: 4349},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 239, col: 13, offset: 4356},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 239, col: 20, offset: 4364},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 241, col: 1, offset: 4372},
			expr: &litMatcher{
				pos:        position{line: 242, col: 5, offset: 4393},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 244, col: 1, offset: 4399},
			expr: &litMatcher{
				pos:        position{line: 245, col: 5, offset: 4415},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 247, col: 1, offset: 4420},
			expr: &litMatcher{
				pos:        position{line: 248, col: 5, offset: 4436},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 250, col: 1, offset: 4441},
			expr: &litMatcher{
				pos:        position{line: 251, col: 5, offset: 4455},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 253, col: 1, offset: 4460},
			expr: &choiceExpr{
				pos: position{line: 255, col: 9, offset: 4488},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 255, col: 9, offset: 4488},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 256, col: 9, offset: 4512},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 257, col: 9, offset: 4537},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 258, col: 9, offset: 4562},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 259, col: 9, offset: 4582},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 9, offset: 4602},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 263, col: 1, offset: 4619},
			expr: &seqExpr{
				pos: position{line: 264, col: 5, offset: 4638},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 264, col: 5, offset: 4638},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 264, col: 12, offset: 4645},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 266, col: 1, offset: 4660},
			expr: &actionExpr{
				pos: position{line: 267, col: 5, offset: 4673},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 267, col: 5, offset: 4673},
					expr: &ruleRefExpr{
						pos:  position{line: 267, col: 5, offset: 4673},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 271, col: 1, offset: 4742},
			expr: &choiceExpr{
				pos: position{line: 272, col: 5, offset: 4760},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 272, col: 5, offset: 4760},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 272, col: 7, offset: 4762},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 272, col: 7, offset: 4762},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 272, col: 11, offset: 4766},
									expr: &ruleRefExpr{
										pos:  position{line: 272, col: 11, offset: 4766},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 272, col: 29, offset: 4784},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 275, col: 5, offset: 4848},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 275, col: 7, offset: 4850},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 275, col: 7, offset: 4850},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 275, col: 11, offset: 4854},
									expr: &ruleRefExpr{
										pos:  position{line: 275, col: 11, offset: 4854},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 275, col: 31, offset: 4874},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 275, col: 31, offset: 4874},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 275, col: 37, offset: 4880},
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
			pos:  position{line: 279, col: 1, offset: 4962},
			expr: &choiceExpr{
				pos: position{line: 280, col: 5, offset: 4983},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 280, col: 5, offset: 4983},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 280, col: 5, offset: 4983},
								expr: &choiceExpr{
									pos: position{line: 280, col: 8, offset: 4986},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 280, col: 8, offset: 4986},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 280, col: 14, offset: 4992},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 21, offset: 4999},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 280, col: 27, offset: 5005},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 281, col: 5, offset: 5020},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 281, col: 5, offset: 5020},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 281, col: 10, offset: 5025},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 283, col: 1, offset: 5045},
			expr: &choiceExpr{
				pos: position{line: 284, col: 5, offset: 5068},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 284, col: 5, offset: 5068},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 285, col: 5, offset: 5076},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 285, col: 7, offset: 5078},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 285, col: 7, offset: 5078},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 285, col: 20, offset: 5091},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 285, col: 26, offset: 5097},
									name: "EOF",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "String",
			pos:  position{line: 289, col: 1, offset: 5169},
			expr: &actionExpr{
				pos: position{line: 290, col: 5, offset: 5180},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 290, col: 5, offset: 5180},
					expr: &ruleRefExpr{
						pos:  position{line: 290, col: 5, offset: 5180},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 294, col: 1, offset: 5240},
			expr: &seqExpr{
				pos: position{line: 295, col: 5, offset: 5255},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 295, col: 5, offset: 5255},
						expr: &choiceExpr{
							pos: position{line: 295, col: 7, offset: 5257},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 295, col: 7, offset: 5257},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 13, offset: 5263},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 19, offset: 5269},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 25, offset: 5275},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 31, offset: 5281},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 37, offset: 5287},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 43, offset: 5293},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 295, col: 49, offset: 5299},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 295, col: 54, offset: 5304},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 297, col: 1, offset: 5316},
			expr: &actionExpr{
				pos: position{line: 298, col: 5, offset: 5327},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 298, col: 5, offset: 5327},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 298, col: 5, offset: 5327},
							expr: &litMatcher{
								pos:        position{line: 298, col: 5, offset: 5327},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 298, col: 10, offset: 5332},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 298, col: 18, offset: 5340},
							expr: &seqExpr{
								pos: position{line: 298, col: 20, offset: 5342},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 298, col: 20, offset: 5342},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 298, col: 24, offset: 5346},
										expr: &ruleRefExpr{
											pos:  position{line: 298, col: 24, offset: 5346},
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
			pos:  position{line: 302, col: 1, offset: 5407},
			expr: &choiceExpr{
				pos: position{line: 303, col: 5, offset: 5419},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 303, col: 5, offset: 5419},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 304, col: 5, offset: 5427},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 304, col: 5, offset: 5427},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 304, col: 18, offset: 5440},
								expr: &ruleRefExpr{
									pos:  position{line: 304, col: 18, offset: 5440},
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
			pos:  position{line: 306, col: 1, offset: 5448},
			expr: &charClassMatcher{
				pos:        position{line: 307, col: 5, offset: 5465},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 309, col: 1, offset: 5472},
			expr: &charClassMatcher{
				pos:        position{line: 310, col: 5, offset: 5482},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 312, col: 1, offset: 5489},
			expr: &actionExpr{
				pos: position{line: 313, col: 5, offset: 5499},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 313, col: 5, offset: 5499},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 313, col: 11, offset: 5505},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 317, col: 1, offset: 5559},
			expr: &actionExpr{
				pos: position{line: 318, col: 5, offset: 5609},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 318, col: 5, offset: 5609},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 318, col: 5, offset: 5609},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 318, col: 9, offset: 5613},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 318, col: 17, offset: 5621},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 318, col: 39, offset: 5643},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 322, col: 1, offset: 5685},
			expr: &actionExpr{
				pos: position{line: 323, col: 5, offset: 5711},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 323, col: 5, offset: 5711},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 323, col: 11, offset: 5717},
						expr: &ruleRefExpr{
							pos:  position{line: 323, col: 11, offset: 5717},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 327, col: 1, offset: 5795},
			expr: &choiceExpr{
				pos: position{line: 328, col: 5, offset: 5821},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 328, col: 5, offset: 5821},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 328, col: 5, offset: 5821},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 328, col: 5, offset: 5821},
									expr: &charClassMatcher{
										pos:        position{line: 328, col: 6, offset: 5822},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 328, col: 12, offset: 5828},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 328, col: 15, offset: 5831},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 331, col: 5, offset: 5893},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 333, col: 1, offset: 5929},
			expr: &choiceExpr{
				pos: position{line: 334, col: 5, offset: 5968},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 334, col: 5, offset: 5968},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 334, col: 5, offset: 5968},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 337, col: 5, offset: 6006},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 337, col: 5, offset: 6006},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 337, col: 10, offset: 6011},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 339, col: 1, offset: 6043},
			expr: &actionExpr{
				pos: position{line: 340, col: 5, offset: 6078},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 340, col: 5, offset: 6078},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 340, col: 5, offset: 6078},
							expr: &ruleRefExpr{
								pos:  position{line: 340, col: 6, offset: 6079},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 340, col: 21, offset: 6094},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 344, col: 1, offset: 6145},
			expr: &anyMatcher{
				line: 345, col: 5, offset: 6160,
			},
		},
		{
			name: "__",
			pos:  position{line: 347, col: 1, offset: 6163},
			expr: &zeroOrMoreExpr{
				pos: position{line: 348, col: 5, offset: 6170},
				expr: &choiceExpr{
					pos: position{line: 348, col: 7, offset: 6172},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 348, col: 7, offset: 6172},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 348, col: 12, offset: 6177},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 350, col: 1, offset: 6185},
			expr: &charClassMatcher{
				pos:        position{line: 351, col: 5, offset: 6192},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 353, col: 1, offset: 6203},
			expr: &charClassMatcher{
				pos:        position{line: 354, col: 5, offset: 6222},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 356, col: 1, offset: 6230},
			expr: &litMatcher{
				pos:        position{line: 357, col: 5, offset: 6238},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 359, col: 1, offset: 6244},
			expr: &notExpr{
				pos: position{line: 360, col: 5, offset: 6252},
				expr: &anyMatcher{
					line: 360, col: 6, offset: 6253,
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
	return head, nil

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

func (c *current) onCallExpression1(callee, args, members interface{}) (interface{}, error) {
	return callchain(callee, args, members, c.text, c.pos)

}

func (p *parser) callonCallExpression1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCallExpression1(stack["callee"], stack["args"], stack["members"])
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

func (c *current) onString1() (interface{}, error) {
	return identifier(c.text, c.pos)

}

func (p *parser) callonString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onString1()
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
