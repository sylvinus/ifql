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

	"github.com/mna/pigeon/ast"
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
							label: "fn",
							expr: &ruleRefExpr{
								pos:  position{line: 7, col: 17, offset: 76},
								name: "Function",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 7, col: 26, offset: 85},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Tests",
			pos:  position{line: 11, col: 1, offset: 125},
			expr: &choiceExpr{
				pos: position{line: 11, col: 10, offset: 134},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 11, col: 10, offset: 134},
						name: "Function",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 21, offset: 145},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 37, offset: 161},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 48, offset: 172},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 59, offset: 183},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 13, col: 1, offset: 191},
			expr: &actionExpr{
				pos: position{line: 13, col: 12, offset: 202},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 13, col: 12, offset: 202},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 13, col: 12, offset: 202},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 13, col: 17, offset: 207},
								name: "FunctionName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 30, offset: 220},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 13, col: 34, offset: 224},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 38, offset: 228},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 13, col: 41, offset: 231},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 13, col: 46, offset: 236},
								expr: &ruleRefExpr{
									pos:  position{line: 13, col: 46, offset: 236},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 60, offset: 250},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 13, col: 63, offset: 253},
							val:        ")",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 13, col: 67, offset: 257},
							label: "children",
							expr: &zeroOrMoreExpr{
								pos: position{line: 13, col: 76, offset: 266},
								expr: &ruleRefExpr{
									pos:  position{line: 13, col: 76, offset: 266},
									name: "FunctionChain",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionChain",
			pos:  position{line: 17, col: 1, offset: 340},
			expr: &actionExpr{
				pos: position{line: 17, col: 17, offset: 356},
				run: (*parser).callonFunctionChain1,
				expr: &seqExpr{
					pos: position{line: 17, col: 17, offset: 356},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 17, col: 17, offset: 356},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 17, col: 20, offset: 359},
							val:        ".",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 17, col: 24, offset: 363},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 17, col: 27, offset: 366},
							label: "child",
							expr: &ruleRefExpr{
								pos:  position{line: 17, col: 33, offset: 372},
								name: "Function",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 21, col: 1, offset: 408},
			expr: &ruleRefExpr{
				pos:  position{line: 21, col: 16, offset: 423},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 22, col: 1, offset: 430},
			expr: &actionExpr{
				pos: position{line: 22, col: 16, offset: 445},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 22, col: 16, offset: 445},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 22, col: 16, offset: 445},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 22, col: 22, offset: 451},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 22, col: 34, offset: 463},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 22, col: 37, offset: 466},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 22, col: 42, offset: 471},
								expr: &ruleRefExpr{
									pos:  position{line: 22, col: 42, offset: 471},
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
			pos:  position{line: 26, col: 1, offset: 540},
			expr: &actionExpr{
				pos: position{line: 26, col: 20, offset: 559},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 26, col: 20, offset: 559},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 26, col: 20, offset: 559},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 26, col: 24, offset: 563},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 26, col: 28, offset: 567},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 26, col: 32, offset: 571},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 30, col: 1, offset: 608},
			expr: &actionExpr{
				pos: position{line: 30, col: 15, offset: 622},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 30, col: 15, offset: 622},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 30, col: 15, offset: 622},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 19, offset: 626},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 26, offset: 633},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 30, col: 30, offset: 637},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 34, offset: 641},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 30, col: 37, offset: 644},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 43, offset: 650},
								name: "FunctionArgValue",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 34, col: 1, offset: 719},
			expr: &choiceExpr{
				pos: position{line: 34, col: 22, offset: 740},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 34, col: 22, offset: 740},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 34, col: 34, offset: 752},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 34, col: 50, offset: 768},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 34, col: 77, offset: 795},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 34, col: 88, offset: 806},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 34, col: 99, offset: 817},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 36, col: 1, offset: 825},
			expr: &actionExpr{
				pos: position{line: 36, col: 13, offset: 837},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 36, col: 13, offset: 837},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 36, col: 13, offset: 837},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 36, col: 17, offset: 841},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 36, col: 20, offset: 844},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 36, col: 25, offset: 849},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 36, col: 30, offset: 854},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 36, col: 34, offset: 858},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 47, col: 1, offset: 1094},
			expr: &ruleRefExpr{
				pos:  position{line: 47, col: 8, offset: 1101},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 49, col: 1, offset: 1110},
			expr: &actionExpr{
				pos: position{line: 49, col: 21, offset: 1130},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 49, col: 22, offset: 1131},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 49, col: 22, offset: 1131},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 49, col: 30, offset: 1139},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 52, col: 1, offset: 1180},
			expr: &actionExpr{
				pos: position{line: 52, col: 11, offset: 1190},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 52, col: 11, offset: 1190},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 52, col: 11, offset: 1190},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 52, col: 16, offset: 1195},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 52, col: 25, offset: 1204},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 52, col: 30, offset: 1209},
								expr: &seqExpr{
									pos: position{line: 52, col: 32, offset: 1211},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 52, col: 32, offset: 1211},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 36, offset: 1215},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 53, offset: 1232},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 57, offset: 1236},
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
			pos:  position{line: 56, col: 1, offset: 1314},
			expr: &actionExpr{
				pos: position{line: 56, col: 22, offset: 1335},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 56, col: 23, offset: 1336},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 56, col: 23, offset: 1336},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 30, offset: 1343},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 59, col: 1, offset: 1382},
			expr: &actionExpr{
				pos: position{line: 59, col: 12, offset: 1393},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 59, col: 12, offset: 1393},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 59, col: 12, offset: 1393},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 59, col: 17, offset: 1398},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 59, col: 28, offset: 1409},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 59, col: 33, offset: 1414},
								expr: &seqExpr{
									pos: position{line: 59, col: 35, offset: 1416},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 59, col: 35, offset: 1416},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 59, col: 38, offset: 1419},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 59, col: 56, offset: 1437},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 59, col: 59, offset: 1440},
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
			pos:  position{line: 63, col: 1, offset: 1518},
			expr: &actionExpr{
				pos: position{line: 63, col: 24, offset: 1541},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 63, col: 26, offset: 1543},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 63, col: 26, offset: 1543},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 63, col: 33, offset: 1550},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 63, col: 39, offset: 1556},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 63, col: 46, offset: 1563},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 63, col: 52, offset: 1569},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 63, col: 68, offset: 1585},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 63, col: 76, offset: 1593},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 63, col: 91, offset: 1608},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 66, col: 1, offset: 1651},
			expr: &actionExpr{
				pos: position{line: 66, col: 14, offset: 1664},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 66, col: 14, offset: 1664},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 66, col: 14, offset: 1664},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 66, col: 19, offset: 1669},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 66, col: 28, offset: 1678},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 66, col: 33, offset: 1683},
								expr: &seqExpr{
									pos: position{line: 66, col: 35, offset: 1685},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 66, col: 35, offset: 1685},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 38, offset: 1688},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 58, offset: 1708},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 61, offset: 1711},
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
			pos:  position{line: 70, col: 1, offset: 1787},
			expr: &actionExpr{
				pos: position{line: 70, col: 20, offset: 1806},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 70, col: 21, offset: 1807},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 70, col: 21, offset: 1807},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 70, col: 27, offset: 1813},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 73, col: 1, offset: 1850},
			expr: &actionExpr{
				pos: position{line: 73, col: 12, offset: 1861},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 73, col: 12, offset: 1861},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 73, col: 12, offset: 1861},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 73, col: 17, offset: 1866},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 73, col: 32, offset: 1881},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 73, col: 37, offset: 1886},
								expr: &seqExpr{
									pos: position{line: 73, col: 39, offset: 1888},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 73, col: 39, offset: 1888},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 73, col: 42, offset: 1891},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 73, col: 59, offset: 1908},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 73, col: 62, offset: 1911},
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
			pos:  position{line: 77, col: 1, offset: 1994},
			expr: &actionExpr{
				pos: position{line: 77, col: 26, offset: 2019},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 77, col: 27, offset: 2020},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 77, col: 27, offset: 2020},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 77, col: 33, offset: 2026},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 80, col: 1, offset: 2063},
			expr: &actionExpr{
				pos: position{line: 80, col: 18, offset: 2080},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 80, col: 18, offset: 2080},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 80, col: 18, offset: 2080},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 80, col: 23, offset: 2085},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 80, col: 31, offset: 2093},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 80, col: 36, offset: 2098},
								expr: &seqExpr{
									pos: position{line: 80, col: 38, offset: 2100},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 80, col: 38, offset: 2100},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 80, col: 41, offset: 2103},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 80, col: 64, offset: 2126},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 80, col: 67, offset: 2129},
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
			pos:  position{line: 84, col: 1, offset: 2204},
			expr: &choiceExpr{
				pos: position{line: 84, col: 11, offset: 2214},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 84, col: 11, offset: 2214},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 84, col: 11, offset: 2214},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 84, col: 11, offset: 2214},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 84, col: 15, offset: 2218},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 84, col: 18, offset: 2221},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 84, col: 23, offset: 2226},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 84, col: 31, offset: 2234},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 84, col: 34, offset: 2237},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 5, offset: 2285},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 21, offset: 2301},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 48, offset: 2328},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 59, offset: 2339},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 70, offset: 2350},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 79, offset: 2359},
						name: "Field",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 88, col: 1, offset: 2366},
			expr: &seqExpr{
				pos: position{line: 88, col: 16, offset: 2381},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 88, col: 16, offset: 2381},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 22, offset: 2387},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 28, offset: 2393},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 34, offset: 2399},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 90, col: 1, offset: 2406},
			expr: &seqExpr{
				pos: position{line: 92, col: 5, offset: 2431},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 92, col: 5, offset: 2431},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 92, col: 11, offset: 2437},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 94, col: 1, offset: 2444},
			expr: &seqExpr{
				pos: position{line: 97, col: 5, offset: 2514},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 97, col: 5, offset: 2514},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 97, col: 11, offset: 2520},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 99, col: 1, offset: 2527},
			expr: &seqExpr{
				pos: position{line: 101, col: 5, offset: 2551},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 101, col: 5, offset: 2551},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 101, col: 11, offset: 2557},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 103, col: 1, offset: 2564},
			expr: &seqExpr{
				pos: position{line: 105, col: 5, offset: 2590},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 105, col: 5, offset: 2590},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 11, offset: 2596},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 107, col: 1, offset: 2603},
			expr: &seqExpr{
				pos: position{line: 110, col: 5, offset: 2675},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 110, col: 5, offset: 2675},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 110, col: 11, offset: 2681},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 112, col: 1, offset: 2688},
			expr: &seqExpr{
				pos: position{line: 112, col: 15, offset: 2702},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 112, col: 15, offset: 2702},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 112, col: 19, offset: 2706},
						expr: &ruleRefExpr{
							pos:  position{line: 112, col: 19, offset: 2706},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 113, col: 1, offset: 2713},
			expr: &seqExpr{
				pos: position{line: 113, col: 17, offset: 2729},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 113, col: 18, offset: 2730},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 113, col: 18, offset: 2730},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 113, col: 24, offset: 2736},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 29, offset: 2741},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 113, col: 38, offset: 2750},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 42, offset: 2754},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 114, col: 1, offset: 2765},
			expr: &choiceExpr{
				pos: position{line: 114, col: 15, offset: 2779},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 114, col: 15, offset: 2779},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 114, col: 22, offset: 2786},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 115, col: 1, offset: 2801},
			expr: &seqExpr{
				pos: position{line: 115, col: 15, offset: 2815},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 115, col: 15, offset: 2815},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 115, col: 24, offset: 2824},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 115, col: 28, offset: 2828},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 115, col: 39, offset: 2839},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 115, col: 43, offset: 2843},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 115, col: 54, offset: 2854},
						expr: &ruleRefExpr{
							pos:  position{line: 115, col: 54, offset: 2854},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 116, col: 1, offset: 2867},
			expr: &seqExpr{
				pos: position{line: 116, col: 12, offset: 2878},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 116, col: 12, offset: 2878},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 116, col: 25, offset: 2891},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 116, col: 29, offset: 2895},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 116, col: 39, offset: 2905},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 116, col: 43, offset: 2909},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 117, col: 1, offset: 2918},
			expr: &seqExpr{
				pos: position{line: 117, col: 12, offset: 2929},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 117, col: 12, offset: 2929},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 117, col: 24, offset: 2941},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 118, col: 1, offset: 2952},
			expr: &actionExpr{
				pos: position{line: 118, col: 12, offset: 2963},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 118, col: 12, offset: 2963},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 118, col: 12, offset: 2963},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 118, col: 21, offset: 2972},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 118, col: 26, offset: 2977},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 122, col: 1, offset: 3026},
			expr: &litMatcher{
				pos:        position{line: 122, col: 19, offset: 3044},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 123, col: 1, offset: 3049},
			expr: &choiceExpr{
				pos: position{line: 123, col: 21, offset: 3069},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 123, col: 21, offset: 3069},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 123, col: 28, offset: 3076},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 123, col: 35, offset: 3084},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 124, col: 1, offset: 3091},
			expr: &litMatcher{
				pos:        position{line: 124, col: 20, offset: 3110},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 125, col: 1, offset: 3115},
			expr: &litMatcher{
				pos:        position{line: 125, col: 15, offset: 3129},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 126, col: 1, offset: 3133},
			expr: &litMatcher{
				pos:        position{line: 126, col: 15, offset: 3147},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 127, col: 1, offset: 3151},
			expr: &litMatcher{
				pos:        position{line: 127, col: 13, offset: 3163},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 128, col: 1, offset: 3167},
			expr: &choiceExpr{
				pos: position{line: 128, col: 18, offset: 3184},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 128, col: 18, offset: 3184},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 36, offset: 3202},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 55, offset: 3221},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 74, offset: 3240},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 88, offset: 3254},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 128, col: 102, offset: 3268},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 130, col: 1, offset: 3280},
			expr: &seqExpr{
				pos: position{line: 130, col: 18, offset: 3297},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 130, col: 18, offset: 3297},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 130, col: 25, offset: 3304},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 131, col: 1, offset: 3318},
			expr: &actionExpr{
				pos: position{line: 131, col: 12, offset: 3329},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 131, col: 12, offset: 3329},
					expr: &ruleRefExpr{
						pos:  position{line: 131, col: 12, offset: 3329},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 135, col: 1, offset: 3392},
			expr: &choiceExpr{
				pos: position{line: 135, col: 17, offset: 3408},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 135, col: 17, offset: 3408},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 135, col: 19, offset: 3410},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 135, col: 19, offset: 3410},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 135, col: 23, offset: 3414},
									expr: &ruleRefExpr{
										pos:  position{line: 135, col: 23, offset: 3414},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 135, col: 41, offset: 3432},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 137, col: 5, offset: 3484},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 137, col: 7, offset: 3486},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 137, col: 7, offset: 3486},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 137, col: 11, offset: 3490},
									expr: &ruleRefExpr{
										pos:  position{line: 137, col: 11, offset: 3490},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 137, col: 31, offset: 3510},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 137, col: 31, offset: 3510},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 37, offset: 3516},
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
			pos:  position{line: 141, col: 1, offset: 3588},
			expr: &choiceExpr{
				pos: position{line: 141, col: 20, offset: 3607},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 141, col: 20, offset: 3607},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 141, col: 20, offset: 3607},
								expr: &choiceExpr{
									pos: position{line: 141, col: 23, offset: 3610},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 141, col: 23, offset: 3610},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 141, col: 29, offset: 3616},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 141, col: 36, offset: 3623},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 141, col: 42, offset: 3629},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 141, col: 55, offset: 3642},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 141, col: 55, offset: 3642},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 141, col: 60, offset: 3647},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 143, col: 1, offset: 3667},
			expr: &choiceExpr{
				pos: position{line: 143, col: 23, offset: 3689},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 143, col: 23, offset: 3689},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 143, col: 29, offset: 3695},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 143, col: 31, offset: 3697},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 143, col: 31, offset: 3697},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 143, col: 44, offset: 3710},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 143, col: 50, offset: 3716},
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
			pos:  position{line: 147, col: 1, offset: 3782},
			expr: &actionExpr{
				pos: position{line: 147, col: 10, offset: 3791},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 147, col: 10, offset: 3791},
					expr: &ruleRefExpr{
						pos:  position{line: 147, col: 10, offset: 3791},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 151, col: 1, offset: 3845},
			expr: &seqExpr{
				pos: position{line: 151, col: 14, offset: 3858},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 151, col: 14, offset: 3858},
						expr: &choiceExpr{
							pos: position{line: 151, col: 16, offset: 3860},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 151, col: 16, offset: 3860},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 22, offset: 3866},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 28, offset: 3872},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 34, offset: 3878},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 40, offset: 3884},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 46, offset: 3890},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 52, offset: 3896},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 151, col: 58, offset: 3902},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 151, col: 63, offset: 3907},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 153, col: 1, offset: 3919},
			expr: &actionExpr{
				pos: position{line: 153, col: 10, offset: 3928},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 153, col: 10, offset: 3928},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 153, col: 10, offset: 3928},
							expr: &litMatcher{
								pos:        position{line: 153, col: 10, offset: 3928},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 153, col: 15, offset: 3933},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 153, col: 23, offset: 3941},
							expr: &seqExpr{
								pos: position{line: 153, col: 25, offset: 3943},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 153, col: 25, offset: 3943},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 153, col: 29, offset: 3947},
										expr: &ruleRefExpr{
											pos:  position{line: 153, col: 29, offset: 3947},
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
			pos:  position{line: 157, col: 1, offset: 4002},
			expr: &choiceExpr{
				pos: position{line: 157, col: 11, offset: 4012},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 157, col: 11, offset: 4012},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 157, col: 17, offset: 4018},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 157, col: 17, offset: 4018},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 157, col: 30, offset: 4031},
								expr: &ruleRefExpr{
									pos:  position{line: 157, col: 30, offset: 4031},
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
			pos:  position{line: 158, col: 1, offset: 4038},
			expr: &charClassMatcher{
				pos:        position{line: 158, col: 16, offset: 4053},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 159, col: 1, offset: 4059},
			expr: &charClassMatcher{
				pos:        position{line: 159, col: 9, offset: 4067},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 161, col: 1, offset: 4074},
			expr: &actionExpr{
				pos: position{line: 161, col: 9, offset: 4082},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 161, col: 9, offset: 4082},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 161, col: 15, offset: 4088},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 165, col: 1, offset: 4136},
			expr: &actionExpr{
				pos: position{line: 166, col: 5, offset: 4186},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 166, col: 5, offset: 4186},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 5, offset: 4186},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 166, col: 9, offset: 4190},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 166, col: 17, offset: 4198},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 166, col: 39, offset: 4220},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 170, col: 1, offset: 4262},
			expr: &actionExpr{
				pos: position{line: 171, col: 5, offset: 4288},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 171, col: 5, offset: 4288},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 171, col: 11, offset: 4294},
						expr: &ruleRefExpr{
							pos:  position{line: 171, col: 11, offset: 4294},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 175, col: 1, offset: 4372},
			expr: &choiceExpr{
				pos: position{line: 176, col: 5, offset: 4398},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 176, col: 5, offset: 4398},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 176, col: 5, offset: 4398},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 176, col: 5, offset: 4398},
									expr: &charClassMatcher{
										pos:        position{line: 176, col: 6, offset: 4399},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 176, col: 12, offset: 4405},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 176, col: 15, offset: 4408},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 179, col: 5, offset: 4470},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 181, col: 1, offset: 4506},
			expr: &choiceExpr{
				pos: position{line: 182, col: 5, offset: 4545},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 182, col: 5, offset: 4545},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 182, col: 5, offset: 4545},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 185, col: 5, offset: 4583},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 185, col: 5, offset: 4583},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 185, col: 10, offset: 4588},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 187, col: 1, offset: 4620},
			expr: &actionExpr{
				pos: position{line: 188, col: 5, offset: 4655},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 188, col: 5, offset: 4655},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 188, col: 5, offset: 4655},
							expr: &ruleRefExpr{
								pos:  position{line: 188, col: 6, offset: 4656},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 21, offset: 4671},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 192, col: 1, offset: 4722},
			expr: &anyMatcher{
				line: 192, col: 14, offset: 4735,
			},
		},
		{
			name: "__",
			pos:  position{line: 194, col: 1, offset: 4738},
			expr: &zeroOrMoreExpr{
				pos: position{line: 194, col: 6, offset: 4743},
				expr: &choiceExpr{
					pos: position{line: 194, col: 8, offset: 4745},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 194, col: 8, offset: 4745},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 194, col: 13, offset: 4750},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 195, col: 1, offset: 4757},
			expr: &charClassMatcher{
				pos:        position{line: 195, col: 6, offset: 4762},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 196, col: 1, offset: 4772},
			expr: &charClassMatcher{
				pos:        position{line: 197, col: 5, offset: 4791},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 199, col: 1, offset: 4799},
			expr: &litMatcher{
				pos:        position{line: 199, col: 7, offset: 4805},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 200, col: 1, offset: 4810},
			expr: &notExpr{
				pos: position{line: 200, col: 7, offset: 4816},
				expr: &anyMatcher{
					line: 200, col: 8, offset: 4817,
				},
			},
		},
	},
}

func (c *current) onGrammar1(fn interface{}) (interface{}, error) {
	return fn.(*Function), nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["fn"])
}

func (c *current) onFunction1(name, args, children interface{}) (interface{}, error) {
	return NewFunction(name.(string), args, children)
}

func (p *parser) callonFunction1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunction1(stack["name"], stack["args"], stack["children"])
}

func (c *current) onFunctionChain1(child interface{}) (interface{}, error) {
	return child, nil
}

func (p *parser) callonFunctionChain1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionChain1(stack["child"])
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
	return expr.(ast.Expression), nil
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
	return logicalExpression(head, tail) // TODO: Add source
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
	return binaryExpression(head, tail) // TODO: Add source
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
	return binaryExpression(head, tail) // TODO: Add source
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

	return binaryExpression(head, tail) // TODO: Add source
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
	return binaryExpression(head, tail) // TODO: Add source
}

func (p *parser) callonMultiplicative1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiplicative1(stack["head"], stack["tail"])
}

func (c *current) onPrimary2(expr interface{}) (interface{}, error) {
	return expr.(ast.Expression), nil
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
