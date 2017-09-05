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
			name: "Grammar",
			pos:  position{line: 7, col: 1, offset: 60},
			expr: &actionExpr{
				pos: position{line: 7, col: 11, offset: 70},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 7, col: 11, offset: 70},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 7, col: 11, offset: 70},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 7, col: 14, offset: 73},
							label: "call",
							expr: &ruleRefExpr{
								pos:  position{line: 7, col: 19, offset: 78},
								name: "CallExpression",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 7, col: 34, offset: 93},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Program",
			pos:  position{line: 11, col: 1, offset: 147},
			expr: &labeledExpr{
				pos:   position{line: 11, col: 11, offset: 157},
				label: "body",
				expr: &zeroOrOneExpr{
					pos: position{line: 11, col: 16, offset: 162},
					expr: &ruleRefExpr{
						pos:  position{line: 11, col: 16, offset: 162},
						name: "SourceElements",
					},
				},
			},
		},
		{
			name: "SourceElements",
			pos:  position{line: 12, col: 1, offset: 178},
			expr: &seqExpr{
				pos: position{line: 12, col: 18, offset: 195},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 12, col: 18, offset: 195},
						label: "head",
						expr: &ruleRefExpr{
							pos:  position{line: 12, col: 23, offset: 200},
							name: "SourceElement",
						},
					},
					&labeledExpr{
						pos:   position{line: 12, col: 37, offset: 214},
						label: "tail",
						expr: &zeroOrMoreExpr{
							pos: position{line: 12, col: 42, offset: 219},
							expr: &seqExpr{
								pos: position{line: 12, col: 43, offset: 220},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 12, col: 43, offset: 220},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 12, col: 46, offset: 223},
										name: "SourceElement",
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
			pos:  position{line: 13, col: 1, offset: 239},
			expr: &ruleRefExpr{
				pos:  position{line: 13, col: 17, offset: 255},
				name: "Statement",
			},
		},
		{
			name: "Statement",
			pos:  position{line: 15, col: 1, offset: 266},
			expr: &choiceExpr{
				pos: position{line: 16, col: 5, offset: 280},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 16, col: 5, offset: 280},
						name: "VariableStatement",
					},
					&ruleRefExpr{
						pos:  position{line: 17, col: 5, offset: 302},
						name: "ExpressionStatement",
					},
				},
			},
		},
		{
			name: "VariableStatement",
			pos:  position{line: 19, col: 1, offset: 323},
			expr: &seqExpr{
				pos: position{line: 20, col: 5, offset: 345},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 20, col: 5, offset: 345},
						name: "VarToken",
					},
					&ruleRefExpr{
						pos:  position{line: 20, col: 14, offset: 354},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 20, col: 17, offset: 357},
						label: "declarations",
						expr: &ruleRefExpr{
							pos:  position{line: 20, col: 30, offset: 370},
							name: "VariableDeclarationList",
						},
					},
				},
			},
		},
		{
			name: "VariableDeclarationList",
			pos:  position{line: 22, col: 1, offset: 395},
			expr: &seqExpr{
				pos: position{line: 23, col: 5, offset: 423},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 23, col: 5, offset: 423},
						label: "head",
						expr: &ruleRefExpr{
							pos:  position{line: 23, col: 10, offset: 428},
							name: "VariableDeclaration",
						},
					},
					&labeledExpr{
						pos:   position{line: 23, col: 30, offset: 448},
						label: "tail",
						expr: &zeroOrMoreExpr{
							pos: position{line: 23, col: 35, offset: 453},
							expr: &seqExpr{
								pos: position{line: 23, col: 36, offset: 454},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 23, col: 36, offset: 454},
										name: "__",
									},
									&litMatcher{
										pos:        position{line: 23, col: 39, offset: 457},
										val:        ",",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 23, col: 43, offset: 461},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 23, col: 46, offset: 464},
										name: "VariableDeclaration",
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
			pos:  position{line: 25, col: 1, offset: 487},
			expr: &litMatcher{
				pos:        position{line: 25, col: 12, offset: 498},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 27, col: 1, offset: 505},
			expr: &seqExpr{
				pos: position{line: 28, col: 5, offset: 529},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 28, col: 5, offset: 529},
						label: "id",
						expr: &ruleRefExpr{
							pos:  position{line: 28, col: 8, offset: 532},
							name: "String",
						},
					},
					&labeledExpr{
						pos:   position{line: 28, col: 15, offset: 539},
						label: "initExpr",
						expr: &zeroOrOneExpr{
							pos: position{line: 28, col: 24, offset: 548},
							expr: &seqExpr{
								pos: position{line: 28, col: 26, offset: 550},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 28, col: 26, offset: 550},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 28, col: 29, offset: 553},
										name: "Initializer",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Initializer",
			pos:  position{line: 30, col: 1, offset: 568},
			expr: &seqExpr{
				pos: position{line: 31, col: 5, offset: 584},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 31, col: 5, offset: 584},
						val:        "=",
						ignoreCase: false,
					},
					&notExpr{
						pos: position{line: 31, col: 9, offset: 588},
						expr: &litMatcher{
							pos:        position{line: 31, col: 10, offset: 589},
							val:        "=",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 14, offset: 593},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 31, col: 17, offset: 596},
						label: "expression",
						expr: &ruleRefExpr{
							pos:  position{line: 31, col: 28, offset: 607},
							name: "VariableExpression",
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 33, col: 1, offset: 627},
			expr: &choiceExpr{
				pos: position{line: 33, col: 22, offset: 648},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 33, col: 22, offset: 648},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 39, offset: 665},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 55, offset: 681},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 82, offset: 708},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 93, offset: 719},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 104, offset: 730},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 33, col: 113, offset: 739},
						name: "Field",
					},
				},
			},
		},
		{
			name: "CallExpression",
			pos:  position{line: 37, col: 1, offset: 802},
			expr: &actionExpr{
				pos: position{line: 37, col: 18, offset: 819},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 37, col: 18, offset: 819},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 37, col: 18, offset: 819},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 25, offset: 826},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 32, offset: 833},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 37, col: 35, offset: 836},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 40, offset: 841},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 37, col: 50, offset: 851},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 37, col: 58, offset: 859},
								expr: &seqExpr{
									pos: position{line: 37, col: 60, offset: 861},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 37, col: 60, offset: 861},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 37, col: 63, offset: 864},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 37, col: 67, offset: 868},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 37, col: 70, offset: 871},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 37, col: 77, offset: 878},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 37, col: 80, offset: 881},
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
			pos:  position{line: 41, col: 1, offset: 957},
			expr: &actionExpr{
				pos: position{line: 41, col: 13, offset: 969},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 41, col: 13, offset: 969},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 41, col: 13, offset: 969},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 41, col: 17, offset: 973},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 41, col: 20, offset: 976},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 41, col: 25, offset: 981},
								expr: &ruleRefExpr{
									pos:  position{line: 41, col: 26, offset: 982},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 41, col: 41, offset: 997},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 41, col: 44, offset: 1000},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 45, col: 1, offset: 1030},
			expr: &actionExpr{
				pos: position{line: 45, col: 16, offset: 1045},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 45, col: 16, offset: 1045},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 45, col: 16, offset: 1045},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 45, col: 22, offset: 1051},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 45, col: 34, offset: 1063},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 45, col: 37, offset: 1066},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 45, col: 42, offset: 1071},
								expr: &ruleRefExpr{
									pos:  position{line: 45, col: 42, offset: 1071},
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
			pos:  position{line: 49, col: 1, offset: 1140},
			expr: &actionExpr{
				pos: position{line: 49, col: 20, offset: 1159},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 49, col: 20, offset: 1159},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 49, col: 20, offset: 1159},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 49, col: 24, offset: 1163},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 49, col: 28, offset: 1167},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 49, col: 32, offset: 1171},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 53, col: 1, offset: 1208},
			expr: &actionExpr{
				pos: position{line: 53, col: 15, offset: 1222},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 53, col: 15, offset: 1222},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 53, col: 15, offset: 1222},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 53, col: 19, offset: 1226},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 53, col: 26, offset: 1233},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 53, col: 30, offset: 1237},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 53, col: 34, offset: 1241},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 53, col: 37, offset: 1244},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 53, col: 43, offset: 1250},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 57, col: 1, offset: 1320},
			expr: &choiceExpr{
				pos: position{line: 57, col: 22, offset: 1341},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 57, col: 22, offset: 1341},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 34, offset: 1353},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 50, offset: 1369},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 77, offset: 1396},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 88, offset: 1407},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 99, offset: 1418},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 108, offset: 1427},
						name: "String",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 59, col: 1, offset: 1435},
			expr: &actionExpr{
				pos: position{line: 59, col: 13, offset: 1447},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 59, col: 13, offset: 1447},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 59, col: 13, offset: 1447},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 59, col: 17, offset: 1451},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 59, col: 20, offset: 1454},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 59, col: 25, offset: 1459},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 59, col: 30, offset: 1464},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 59, col: 34, offset: 1468},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 70, col: 1, offset: 1685},
			expr: &ruleRefExpr{
				pos:  position{line: 70, col: 8, offset: 1692},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 72, col: 1, offset: 1701},
			expr: &actionExpr{
				pos: position{line: 72, col: 21, offset: 1721},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 72, col: 22, offset: 1722},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 72, col: 22, offset: 1722},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 72, col: 30, offset: 1730},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 75, col: 1, offset: 1771},
			expr: &actionExpr{
				pos: position{line: 75, col: 11, offset: 1781},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 75, col: 11, offset: 1781},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 75, col: 11, offset: 1781},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 75, col: 16, offset: 1786},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 75, col: 25, offset: 1795},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 75, col: 30, offset: 1800},
								expr: &seqExpr{
									pos: position{line: 75, col: 32, offset: 1802},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 75, col: 32, offset: 1802},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 36, offset: 1806},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 53, offset: 1823},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 57, offset: 1827},
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
			pos:  position{line: 79, col: 1, offset: 1900},
			expr: &actionExpr{
				pos: position{line: 79, col: 22, offset: 1921},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 79, col: 23, offset: 1922},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 79, col: 23, offset: 1922},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 79, col: 30, offset: 1929},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 82, col: 1, offset: 1968},
			expr: &actionExpr{
				pos: position{line: 82, col: 12, offset: 1979},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 82, col: 12, offset: 1979},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 82, col: 12, offset: 1979},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 82, col: 17, offset: 1984},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 82, col: 28, offset: 1995},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 82, col: 33, offset: 2000},
								expr: &seqExpr{
									pos: position{line: 82, col: 35, offset: 2002},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 82, col: 35, offset: 2002},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 82, col: 38, offset: 2005},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 82, col: 56, offset: 2023},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 82, col: 59, offset: 2026},
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
			pos:  position{line: 86, col: 1, offset: 2100},
			expr: &actionExpr{
				pos: position{line: 86, col: 24, offset: 2123},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 86, col: 26, offset: 2125},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 86, col: 26, offset: 2125},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 86, col: 33, offset: 2132},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 86, col: 39, offset: 2138},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 86, col: 46, offset: 2145},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 86, col: 52, offset: 2151},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 86, col: 68, offset: 2167},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 86, col: 76, offset: 2175},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 86, col: 91, offset: 2190},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 89, col: 1, offset: 2233},
			expr: &actionExpr{
				pos: position{line: 89, col: 14, offset: 2246},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 89, col: 14, offset: 2246},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 89, col: 14, offset: 2246},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 19, offset: 2251},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 89, col: 28, offset: 2260},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 89, col: 33, offset: 2265},
								expr: &seqExpr{
									pos: position{line: 89, col: 35, offset: 2267},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 89, col: 35, offset: 2267},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 89, col: 38, offset: 2270},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 89, col: 58, offset: 2290},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 89, col: 61, offset: 2293},
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
			pos:  position{line: 93, col: 1, offset: 2365},
			expr: &actionExpr{
				pos: position{line: 93, col: 20, offset: 2384},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 93, col: 21, offset: 2385},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 93, col: 21, offset: 2385},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 93, col: 27, offset: 2391},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 96, col: 1, offset: 2428},
			expr: &actionExpr{
				pos: position{line: 96, col: 12, offset: 2439},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 96, col: 12, offset: 2439},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 96, col: 12, offset: 2439},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 96, col: 17, offset: 2444},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 96, col: 32, offset: 2459},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 96, col: 37, offset: 2464},
								expr: &seqExpr{
									pos: position{line: 96, col: 39, offset: 2466},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 96, col: 39, offset: 2466},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 96, col: 42, offset: 2469},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 96, col: 59, offset: 2486},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 96, col: 62, offset: 2489},
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
			pos:  position{line: 100, col: 1, offset: 2568},
			expr: &actionExpr{
				pos: position{line: 100, col: 26, offset: 2593},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 100, col: 27, offset: 2594},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 100, col: 27, offset: 2594},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 100, col: 33, offset: 2600},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 103, col: 1, offset: 2637},
			expr: &actionExpr{
				pos: position{line: 103, col: 18, offset: 2654},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 103, col: 18, offset: 2654},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 103, col: 18, offset: 2654},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 103, col: 23, offset: 2659},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 103, col: 31, offset: 2667},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 103, col: 36, offset: 2672},
								expr: &seqExpr{
									pos: position{line: 103, col: 38, offset: 2674},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 103, col: 38, offset: 2674},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 103, col: 41, offset: 2677},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 103, col: 64, offset: 2700},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 103, col: 67, offset: 2703},
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
			pos:  position{line: 107, col: 1, offset: 2774},
			expr: &choiceExpr{
				pos: position{line: 107, col: 11, offset: 2784},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 107, col: 11, offset: 2784},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 107, col: 11, offset: 2784},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 107, col: 11, offset: 2784},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 107, col: 15, offset: 2788},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 107, col: 18, offset: 2791},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 107, col: 23, offset: 2796},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 107, col: 31, offset: 2804},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 107, col: 34, offset: 2807},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 5, offset: 2838},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 21, offset: 2854},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 48, offset: 2881},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 59, offset: 2892},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 70, offset: 2903},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 79, offset: 2912},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 87, offset: 2920},
						name: "String",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 111, col: 1, offset: 2928},
			expr: &seqExpr{
				pos: position{line: 111, col: 16, offset: 2943},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 111, col: 16, offset: 2943},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 111, col: 22, offset: 2949},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 111, col: 28, offset: 2955},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 111, col: 34, offset: 2961},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 113, col: 1, offset: 2968},
			expr: &seqExpr{
				pos: position{line: 115, col: 5, offset: 2993},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 115, col: 5, offset: 2993},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 115, col: 11, offset: 2999},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 117, col: 1, offset: 3006},
			expr: &seqExpr{
				pos: position{line: 120, col: 5, offset: 3076},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 120, col: 5, offset: 3076},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 120, col: 11, offset: 3082},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 122, col: 1, offset: 3089},
			expr: &seqExpr{
				pos: position{line: 124, col: 5, offset: 3113},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 124, col: 5, offset: 3113},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 124, col: 11, offset: 3119},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 126, col: 1, offset: 3126},
			expr: &seqExpr{
				pos: position{line: 128, col: 5, offset: 3152},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 128, col: 5, offset: 3152},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 11, offset: 3158},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 130, col: 1, offset: 3165},
			expr: &seqExpr{
				pos: position{line: 133, col: 5, offset: 3237},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 133, col: 5, offset: 3237},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 133, col: 11, offset: 3243},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 135, col: 1, offset: 3250},
			expr: &seqExpr{
				pos: position{line: 135, col: 15, offset: 3264},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 135, col: 15, offset: 3264},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 135, col: 19, offset: 3268},
						expr: &ruleRefExpr{
							pos:  position{line: 135, col: 19, offset: 3268},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 136, col: 1, offset: 3275},
			expr: &seqExpr{
				pos: position{line: 136, col: 17, offset: 3291},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 136, col: 18, offset: 3292},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 136, col: 18, offset: 3292},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 136, col: 24, offset: 3298},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 29, offset: 3303},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 136, col: 38, offset: 3312},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 42, offset: 3316},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 137, col: 1, offset: 3327},
			expr: &choiceExpr{
				pos: position{line: 137, col: 15, offset: 3341},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 137, col: 15, offset: 3341},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 137, col: 22, offset: 3348},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 138, col: 1, offset: 3363},
			expr: &seqExpr{
				pos: position{line: 138, col: 15, offset: 3377},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 138, col: 15, offset: 3377},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 138, col: 24, offset: 3386},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 138, col: 28, offset: 3390},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 138, col: 39, offset: 3401},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 138, col: 43, offset: 3405},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 138, col: 54, offset: 3416},
						expr: &ruleRefExpr{
							pos:  position{line: 138, col: 54, offset: 3416},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 139, col: 1, offset: 3429},
			expr: &seqExpr{
				pos: position{line: 139, col: 12, offset: 3440},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 139, col: 12, offset: 3440},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 139, col: 25, offset: 3453},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 139, col: 29, offset: 3457},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 139, col: 39, offset: 3467},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 139, col: 43, offset: 3471},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 140, col: 1, offset: 3480},
			expr: &seqExpr{
				pos: position{line: 140, col: 12, offset: 3491},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 140, col: 12, offset: 3491},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 140, col: 24, offset: 3503},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 141, col: 1, offset: 3514},
			expr: &actionExpr{
				pos: position{line: 141, col: 12, offset: 3525},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 141, col: 12, offset: 3525},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 141, col: 12, offset: 3525},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 141, col: 21, offset: 3534},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 141, col: 25, offset: 3538},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 145, col: 1, offset: 3587},
			expr: &litMatcher{
				pos:        position{line: 145, col: 19, offset: 3605},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 146, col: 1, offset: 3610},
			expr: &choiceExpr{
				pos: position{line: 146, col: 21, offset: 3630},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 146, col: 21, offset: 3630},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 146, col: 28, offset: 3637},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 146, col: 35, offset: 3645},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 147, col: 1, offset: 3652},
			expr: &litMatcher{
				pos:        position{line: 147, col: 20, offset: 3671},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 148, col: 1, offset: 3676},
			expr: &litMatcher{
				pos:        position{line: 148, col: 15, offset: 3690},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 149, col: 1, offset: 3694},
			expr: &litMatcher{
				pos:        position{line: 149, col: 15, offset: 3708},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 150, col: 1, offset: 3712},
			expr: &litMatcher{
				pos:        position{line: 150, col: 13, offset: 3724},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 151, col: 1, offset: 3728},
			expr: &choiceExpr{
				pos: position{line: 151, col: 18, offset: 3745},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 151, col: 18, offset: 3745},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 36, offset: 3763},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 55, offset: 3782},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 74, offset: 3801},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 88, offset: 3815},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 102, offset: 3829},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 153, col: 1, offset: 3841},
			expr: &seqExpr{
				pos: position{line: 153, col: 18, offset: 3858},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 153, col: 18, offset: 3858},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 153, col: 25, offset: 3865},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 154, col: 1, offset: 3879},
			expr: &actionExpr{
				pos: position{line: 154, col: 12, offset: 3890},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 154, col: 12, offset: 3890},
					expr: &ruleRefExpr{
						pos:  position{line: 154, col: 12, offset: 3890},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 158, col: 1, offset: 3953},
			expr: &choiceExpr{
				pos: position{line: 158, col: 17, offset: 3969},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 158, col: 17, offset: 3969},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 158, col: 19, offset: 3971},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 158, col: 19, offset: 3971},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 158, col: 23, offset: 3975},
									expr: &ruleRefExpr{
										pos:  position{line: 158, col: 23, offset: 3975},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 158, col: 41, offset: 3993},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 160, col: 5, offset: 4045},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 160, col: 7, offset: 4047},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 160, col: 7, offset: 4047},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 160, col: 11, offset: 4051},
									expr: &ruleRefExpr{
										pos:  position{line: 160, col: 11, offset: 4051},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 160, col: 31, offset: 4071},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 160, col: 31, offset: 4071},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 160, col: 37, offset: 4077},
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
			pos:  position{line: 164, col: 1, offset: 4149},
			expr: &choiceExpr{
				pos: position{line: 164, col: 20, offset: 4168},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 164, col: 20, offset: 4168},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 164, col: 20, offset: 4168},
								expr: &choiceExpr{
									pos: position{line: 164, col: 23, offset: 4171},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 164, col: 23, offset: 4171},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 164, col: 29, offset: 4177},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 164, col: 36, offset: 4184},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 164, col: 42, offset: 4190},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 164, col: 55, offset: 4203},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 164, col: 55, offset: 4203},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 164, col: 60, offset: 4208},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 166, col: 1, offset: 4228},
			expr: &choiceExpr{
				pos: position{line: 166, col: 23, offset: 4250},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 166, col: 23, offset: 4250},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 166, col: 29, offset: 4256},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 166, col: 31, offset: 4258},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 166, col: 31, offset: 4258},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 166, col: 44, offset: 4271},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 166, col: 50, offset: 4277},
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
			pos:  position{line: 170, col: 1, offset: 4343},
			expr: &actionExpr{
				pos: position{line: 170, col: 10, offset: 4352},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 170, col: 10, offset: 4352},
					expr: &ruleRefExpr{
						pos:  position{line: 170, col: 10, offset: 4352},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 174, col: 1, offset: 4406},
			expr: &seqExpr{
				pos: position{line: 174, col: 14, offset: 4419},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 174, col: 14, offset: 4419},
						expr: &choiceExpr{
							pos: position{line: 174, col: 16, offset: 4421},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 174, col: 16, offset: 4421},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 22, offset: 4427},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 28, offset: 4433},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 34, offset: 4439},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 40, offset: 4445},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 46, offset: 4451},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 52, offset: 4457},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 174, col: 58, offset: 4463},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 63, offset: 4468},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 176, col: 1, offset: 4480},
			expr: &actionExpr{
				pos: position{line: 176, col: 10, offset: 4489},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 176, col: 10, offset: 4489},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 176, col: 10, offset: 4489},
							expr: &litMatcher{
								pos:        position{line: 176, col: 10, offset: 4489},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 176, col: 15, offset: 4494},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 176, col: 23, offset: 4502},
							expr: &seqExpr{
								pos: position{line: 176, col: 25, offset: 4504},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 176, col: 25, offset: 4504},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 176, col: 29, offset: 4508},
										expr: &ruleRefExpr{
											pos:  position{line: 176, col: 29, offset: 4508},
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
			pos:  position{line: 180, col: 1, offset: 4563},
			expr: &choiceExpr{
				pos: position{line: 180, col: 11, offset: 4573},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 180, col: 11, offset: 4573},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 180, col: 17, offset: 4579},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 180, col: 17, offset: 4579},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 180, col: 30, offset: 4592},
								expr: &ruleRefExpr{
									pos:  position{line: 180, col: 30, offset: 4592},
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
			pos:  position{line: 181, col: 1, offset: 4599},
			expr: &charClassMatcher{
				pos:        position{line: 181, col: 16, offset: 4614},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 182, col: 1, offset: 4620},
			expr: &charClassMatcher{
				pos:        position{line: 182, col: 9, offset: 4628},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 184, col: 1, offset: 4635},
			expr: &actionExpr{
				pos: position{line: 184, col: 9, offset: 4643},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 184, col: 9, offset: 4643},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 184, col: 15, offset: 4649},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 188, col: 1, offset: 4697},
			expr: &actionExpr{
				pos: position{line: 189, col: 5, offset: 4747},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 189, col: 5, offset: 4747},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 189, col: 5, offset: 4747},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 189, col: 9, offset: 4751},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 17, offset: 4759},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 189, col: 39, offset: 4781},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 193, col: 1, offset: 4823},
			expr: &actionExpr{
				pos: position{line: 194, col: 5, offset: 4849},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 194, col: 5, offset: 4849},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 194, col: 11, offset: 4855},
						expr: &ruleRefExpr{
							pos:  position{line: 194, col: 11, offset: 4855},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 198, col: 1, offset: 4933},
			expr: &choiceExpr{
				pos: position{line: 199, col: 5, offset: 4959},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 199, col: 5, offset: 4959},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 199, col: 5, offset: 4959},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 199, col: 5, offset: 4959},
									expr: &charClassMatcher{
										pos:        position{line: 199, col: 6, offset: 4960},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 199, col: 12, offset: 4966},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 199, col: 15, offset: 4969},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 5, offset: 5031},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 204, col: 1, offset: 5067},
			expr: &choiceExpr{
				pos: position{line: 205, col: 5, offset: 5106},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 205, col: 5, offset: 5106},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 205, col: 5, offset: 5106},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 208, col: 5, offset: 5144},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 208, col: 5, offset: 5144},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 208, col: 10, offset: 5149},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 210, col: 1, offset: 5181},
			expr: &actionExpr{
				pos: position{line: 211, col: 5, offset: 5216},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 211, col: 5, offset: 5216},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 211, col: 5, offset: 5216},
							expr: &ruleRefExpr{
								pos:  position{line: 211, col: 6, offset: 5217},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 211, col: 21, offset: 5232},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 215, col: 1, offset: 5283},
			expr: &anyMatcher{
				line: 215, col: 14, offset: 5296,
			},
		},
		{
			name: "__",
			pos:  position{line: 217, col: 1, offset: 5299},
			expr: &zeroOrMoreExpr{
				pos: position{line: 217, col: 6, offset: 5304},
				expr: &choiceExpr{
					pos: position{line: 217, col: 8, offset: 5306},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 217, col: 8, offset: 5306},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 217, col: 13, offset: 5311},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 218, col: 1, offset: 5318},
			expr: &charClassMatcher{
				pos:        position{line: 218, col: 6, offset: 5323},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 219, col: 1, offset: 5333},
			expr: &charClassMatcher{
				pos:        position{line: 220, col: 5, offset: 5352},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 222, col: 1, offset: 5360},
			expr: &litMatcher{
				pos:        position{line: 222, col: 7, offset: 5366},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 223, col: 1, offset: 5371},
			expr: &notExpr{
				pos: position{line: 223, col: 7, offset: 5377},
				expr: &anyMatcher{
					line: 223, col: 8, offset: 5378,
				},
			},
		},
	},
}

func (c *current) onGrammar1(call interface{}) (interface{}, error) {
	return buildProgram(call, c.text, c.pos)
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["call"])
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
