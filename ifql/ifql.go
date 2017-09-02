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
			name: "CallExpression",
			pos:  position{line: 11, col: 1, offset: 147},
			expr: &actionExpr{
				pos: position{line: 11, col: 18, offset: 164},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 11, col: 18, offset: 164},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 11, col: 18, offset: 164},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 11, col: 25, offset: 171},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 11, col: 32, offset: 178},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 11, col: 35, offset: 181},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 11, col: 40, offset: 186},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 11, col: 50, offset: 196},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 11, col: 58, offset: 204},
								expr: &seqExpr{
									pos: position{line: 11, col: 60, offset: 206},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 11, col: 60, offset: 206},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 11, col: 63, offset: 209},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 67, offset: 213},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 70, offset: 216},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 77, offset: 223},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 80, offset: 226},
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
			pos:  position{line: 15, col: 1, offset: 302},
			expr: &actionExpr{
				pos: position{line: 15, col: 13, offset: 314},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 15, col: 13, offset: 314},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 15, col: 13, offset: 314},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 15, col: 17, offset: 318},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 15, col: 20, offset: 321},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 15, col: 25, offset: 326},
								expr: &ruleRefExpr{
									pos:  position{line: 15, col: 26, offset: 327},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 15, col: 41, offset: 342},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 15, col: 44, offset: 345},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 19, col: 1, offset: 375},
			expr: &actionExpr{
				pos: position{line: 19, col: 16, offset: 390},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 19, col: 16, offset: 390},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 19, col: 16, offset: 390},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 19, col: 22, offset: 396},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 19, col: 34, offset: 408},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 19, col: 37, offset: 411},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 19, col: 42, offset: 416},
								expr: &ruleRefExpr{
									pos:  position{line: 19, col: 42, offset: 416},
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
			pos:  position{line: 23, col: 1, offset: 485},
			expr: &actionExpr{
				pos: position{line: 23, col: 20, offset: 504},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 23, col: 20, offset: 504},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 23, col: 20, offset: 504},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 23, col: 24, offset: 508},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 23, col: 28, offset: 512},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 23, col: 32, offset: 516},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 27, col: 1, offset: 553},
			expr: &actionExpr{
				pos: position{line: 27, col: 15, offset: 567},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 27, col: 15, offset: 567},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 27, col: 15, offset: 567},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 27, col: 19, offset: 571},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 27, col: 26, offset: 578},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 27, col: 30, offset: 582},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 27, col: 34, offset: 586},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 27, col: 37, offset: 589},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 27, col: 43, offset: 595},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 31, col: 1, offset: 665},
			expr: &choiceExpr{
				pos: position{line: 31, col: 22, offset: 686},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 31, col: 22, offset: 686},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 34, offset: 698},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 50, offset: 714},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 77, offset: 741},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 88, offset: 752},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 99, offset: 763},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 33, col: 1, offset: 771},
			expr: &actionExpr{
				pos: position{line: 33, col: 13, offset: 783},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 33, col: 13, offset: 783},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 33, col: 13, offset: 783},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 17, offset: 787},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 33, col: 20, offset: 790},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 33, col: 25, offset: 795},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 30, offset: 800},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 33, col: 34, offset: 804},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 44, col: 1, offset: 1023},
			expr: &ruleRefExpr{
				pos:  position{line: 44, col: 8, offset: 1030},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 46, col: 1, offset: 1039},
			expr: &actionExpr{
				pos: position{line: 46, col: 21, offset: 1059},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 46, col: 22, offset: 1060},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 46, col: 22, offset: 1060},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 46, col: 30, offset: 1068},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 49, col: 1, offset: 1109},
			expr: &actionExpr{
				pos: position{line: 49, col: 11, offset: 1119},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 49, col: 11, offset: 1119},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 49, col: 11, offset: 1119},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 49, col: 16, offset: 1124},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 49, col: 25, offset: 1133},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 49, col: 30, offset: 1138},
								expr: &seqExpr{
									pos: position{line: 49, col: 32, offset: 1140},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 49, col: 32, offset: 1140},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 36, offset: 1144},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 53, offset: 1161},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 57, offset: 1165},
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
			pos:  position{line: 53, col: 1, offset: 1238},
			expr: &actionExpr{
				pos: position{line: 53, col: 22, offset: 1259},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 53, col: 23, offset: 1260},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 53, col: 23, offset: 1260},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 53, col: 30, offset: 1267},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 56, col: 1, offset: 1306},
			expr: &actionExpr{
				pos: position{line: 56, col: 12, offset: 1317},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 56, col: 12, offset: 1317},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 56, col: 12, offset: 1317},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 56, col: 17, offset: 1322},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 56, col: 28, offset: 1333},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 56, col: 33, offset: 1338},
								expr: &seqExpr{
									pos: position{line: 56, col: 35, offset: 1340},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 56, col: 35, offset: 1340},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 38, offset: 1343},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 56, offset: 1361},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 59, offset: 1364},
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
			pos:  position{line: 60, col: 1, offset: 1438},
			expr: &actionExpr{
				pos: position{line: 60, col: 24, offset: 1461},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 60, col: 26, offset: 1463},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 60, col: 26, offset: 1463},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 33, offset: 1470},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 39, offset: 1476},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 46, offset: 1483},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 52, offset: 1489},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 68, offset: 1505},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 76, offset: 1513},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 91, offset: 1528},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 63, col: 1, offset: 1571},
			expr: &actionExpr{
				pos: position{line: 63, col: 14, offset: 1584},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 63, col: 14, offset: 1584},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 63, col: 14, offset: 1584},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 63, col: 19, offset: 1589},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 63, col: 28, offset: 1598},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 63, col: 33, offset: 1603},
								expr: &seqExpr{
									pos: position{line: 63, col: 35, offset: 1605},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 63, col: 35, offset: 1605},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 38, offset: 1608},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 58, offset: 1628},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 61, offset: 1631},
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
			pos:  position{line: 67, col: 1, offset: 1703},
			expr: &actionExpr{
				pos: position{line: 67, col: 20, offset: 1722},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 67, col: 21, offset: 1723},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 67, col: 21, offset: 1723},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 67, col: 27, offset: 1729},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 70, col: 1, offset: 1766},
			expr: &actionExpr{
				pos: position{line: 70, col: 12, offset: 1777},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 70, col: 12, offset: 1777},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 70, col: 12, offset: 1777},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 70, col: 17, offset: 1782},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 70, col: 32, offset: 1797},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 70, col: 37, offset: 1802},
								expr: &seqExpr{
									pos: position{line: 70, col: 39, offset: 1804},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 70, col: 39, offset: 1804},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 42, offset: 1807},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 59, offset: 1824},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 62, offset: 1827},
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
			pos:  position{line: 74, col: 1, offset: 1906},
			expr: &actionExpr{
				pos: position{line: 74, col: 26, offset: 1931},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 74, col: 27, offset: 1932},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 74, col: 27, offset: 1932},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 74, col: 33, offset: 1938},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 77, col: 1, offset: 1975},
			expr: &actionExpr{
				pos: position{line: 77, col: 18, offset: 1992},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 77, col: 18, offset: 1992},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 77, col: 18, offset: 1992},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 77, col: 23, offset: 1997},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 77, col: 31, offset: 2005},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 36, offset: 2010},
								expr: &seqExpr{
									pos: position{line: 77, col: 38, offset: 2012},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 77, col: 38, offset: 2012},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 41, offset: 2015},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 64, offset: 2038},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 67, offset: 2041},
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
			pos:  position{line: 81, col: 1, offset: 2112},
			expr: &choiceExpr{
				pos: position{line: 81, col: 11, offset: 2122},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 81, col: 11, offset: 2122},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 81, col: 11, offset: 2122},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 81, col: 11, offset: 2122},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 81, col: 15, offset: 2126},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 81, col: 18, offset: 2129},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 81, col: 23, offset: 2134},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 81, col: 31, offset: 2142},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 81, col: 34, offset: 2145},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 2176},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 21, offset: 2192},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 48, offset: 2219},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 59, offset: 2230},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 70, offset: 2241},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 79, offset: 2250},
						name: "Field",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 85, col: 1, offset: 2257},
			expr: &seqExpr{
				pos: position{line: 85, col: 16, offset: 2272},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 85, col: 16, offset: 2272},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 22, offset: 2278},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 28, offset: 2284},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 34, offset: 2290},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 87, col: 1, offset: 2297},
			expr: &seqExpr{
				pos: position{line: 89, col: 5, offset: 2322},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 89, col: 5, offset: 2322},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 11, offset: 2328},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 91, col: 1, offset: 2335},
			expr: &seqExpr{
				pos: position{line: 94, col: 5, offset: 2405},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 94, col: 5, offset: 2405},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 94, col: 11, offset: 2411},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 96, col: 1, offset: 2418},
			expr: &seqExpr{
				pos: position{line: 98, col: 5, offset: 2442},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 98, col: 5, offset: 2442},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 98, col: 11, offset: 2448},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 100, col: 1, offset: 2455},
			expr: &seqExpr{
				pos: position{line: 102, col: 5, offset: 2481},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 102, col: 5, offset: 2481},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 102, col: 11, offset: 2487},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 104, col: 1, offset: 2494},
			expr: &seqExpr{
				pos: position{line: 107, col: 5, offset: 2566},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 5, offset: 2566},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 11, offset: 2572},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 109, col: 1, offset: 2579},
			expr: &seqExpr{
				pos: position{line: 109, col: 15, offset: 2593},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 109, col: 15, offset: 2593},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 109, col: 19, offset: 2597},
						expr: &ruleRefExpr{
							pos:  position{line: 109, col: 19, offset: 2597},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 110, col: 1, offset: 2604},
			expr: &seqExpr{
				pos: position{line: 110, col: 17, offset: 2620},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 110, col: 18, offset: 2621},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 110, col: 18, offset: 2621},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 110, col: 24, offset: 2627},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 110, col: 29, offset: 2632},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 110, col: 38, offset: 2641},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 110, col: 42, offset: 2645},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 111, col: 1, offset: 2656},
			expr: &choiceExpr{
				pos: position{line: 111, col: 15, offset: 2670},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 111, col: 15, offset: 2670},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 111, col: 22, offset: 2677},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 112, col: 1, offset: 2692},
			expr: &seqExpr{
				pos: position{line: 112, col: 15, offset: 2706},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 112, col: 15, offset: 2706},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 112, col: 24, offset: 2715},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 28, offset: 2719},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 112, col: 39, offset: 2730},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 43, offset: 2734},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 112, col: 54, offset: 2745},
						expr: &ruleRefExpr{
							pos:  position{line: 112, col: 54, offset: 2745},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 113, col: 1, offset: 2758},
			expr: &seqExpr{
				pos: position{line: 113, col: 12, offset: 2769},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 113, col: 12, offset: 2769},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 113, col: 25, offset: 2782},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 29, offset: 2786},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 113, col: 39, offset: 2796},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 43, offset: 2800},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 114, col: 1, offset: 2809},
			expr: &seqExpr{
				pos: position{line: 114, col: 12, offset: 2820},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 114, col: 12, offset: 2820},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 114, col: 24, offset: 2832},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 115, col: 1, offset: 2843},
			expr: &actionExpr{
				pos: position{line: 115, col: 12, offset: 2854},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 115, col: 12, offset: 2854},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 115, col: 12, offset: 2854},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 115, col: 21, offset: 2863},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 115, col: 25, offset: 2867},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 119, col: 1, offset: 2916},
			expr: &litMatcher{
				pos:        position{line: 119, col: 19, offset: 2934},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 120, col: 1, offset: 2939},
			expr: &choiceExpr{
				pos: position{line: 120, col: 21, offset: 2959},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 120, col: 21, offset: 2959},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 120, col: 28, offset: 2966},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 120, col: 35, offset: 2974},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 121, col: 1, offset: 2981},
			expr: &litMatcher{
				pos:        position{line: 121, col: 20, offset: 3000},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 122, col: 1, offset: 3005},
			expr: &litMatcher{
				pos:        position{line: 122, col: 15, offset: 3019},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 123, col: 1, offset: 3023},
			expr: &litMatcher{
				pos:        position{line: 123, col: 15, offset: 3037},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 124, col: 1, offset: 3041},
			expr: &litMatcher{
				pos:        position{line: 124, col: 13, offset: 3053},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 125, col: 1, offset: 3057},
			expr: &choiceExpr{
				pos: position{line: 125, col: 18, offset: 3074},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 125, col: 18, offset: 3074},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 36, offset: 3092},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 55, offset: 3111},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 74, offset: 3130},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 88, offset: 3144},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 102, offset: 3158},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 127, col: 1, offset: 3170},
			expr: &seqExpr{
				pos: position{line: 127, col: 18, offset: 3187},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 127, col: 18, offset: 3187},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 127, col: 25, offset: 3194},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 128, col: 1, offset: 3208},
			expr: &actionExpr{
				pos: position{line: 128, col: 12, offset: 3219},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 128, col: 12, offset: 3219},
					expr: &ruleRefExpr{
						pos:  position{line: 128, col: 12, offset: 3219},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 132, col: 1, offset: 3282},
			expr: &choiceExpr{
				pos: position{line: 132, col: 17, offset: 3298},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 132, col: 17, offset: 3298},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 132, col: 19, offset: 3300},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 132, col: 19, offset: 3300},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 132, col: 23, offset: 3304},
									expr: &ruleRefExpr{
										pos:  position{line: 132, col: 23, offset: 3304},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 132, col: 41, offset: 3322},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 134, col: 5, offset: 3374},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 134, col: 7, offset: 3376},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 134, col: 7, offset: 3376},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 134, col: 11, offset: 3380},
									expr: &ruleRefExpr{
										pos:  position{line: 134, col: 11, offset: 3380},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 134, col: 31, offset: 3400},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 134, col: 31, offset: 3400},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 134, col: 37, offset: 3406},
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
			pos:  position{line: 138, col: 1, offset: 3478},
			expr: &choiceExpr{
				pos: position{line: 138, col: 20, offset: 3497},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 138, col: 20, offset: 3497},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 138, col: 20, offset: 3497},
								expr: &choiceExpr{
									pos: position{line: 138, col: 23, offset: 3500},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 138, col: 23, offset: 3500},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 138, col: 29, offset: 3506},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 138, col: 36, offset: 3513},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 138, col: 42, offset: 3519},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 138, col: 55, offset: 3532},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 138, col: 55, offset: 3532},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 138, col: 60, offset: 3537},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 140, col: 1, offset: 3557},
			expr: &choiceExpr{
				pos: position{line: 140, col: 23, offset: 3579},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 140, col: 23, offset: 3579},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 140, col: 29, offset: 3585},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 140, col: 31, offset: 3587},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 140, col: 31, offset: 3587},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 140, col: 44, offset: 3600},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 140, col: 50, offset: 3606},
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
			pos:  position{line: 144, col: 1, offset: 3672},
			expr: &actionExpr{
				pos: position{line: 144, col: 10, offset: 3681},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 144, col: 10, offset: 3681},
					expr: &ruleRefExpr{
						pos:  position{line: 144, col: 10, offset: 3681},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 148, col: 1, offset: 3735},
			expr: &seqExpr{
				pos: position{line: 148, col: 14, offset: 3748},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 148, col: 14, offset: 3748},
						expr: &choiceExpr{
							pos: position{line: 148, col: 16, offset: 3750},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 148, col: 16, offset: 3750},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 22, offset: 3756},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 28, offset: 3762},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 34, offset: 3768},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 40, offset: 3774},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 46, offset: 3780},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 52, offset: 3786},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 58, offset: 3792},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 148, col: 63, offset: 3797},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 150, col: 1, offset: 3809},
			expr: &actionExpr{
				pos: position{line: 150, col: 10, offset: 3818},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 150, col: 10, offset: 3818},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 150, col: 10, offset: 3818},
							expr: &litMatcher{
								pos:        position{line: 150, col: 10, offset: 3818},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 150, col: 15, offset: 3823},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 150, col: 23, offset: 3831},
							expr: &seqExpr{
								pos: position{line: 150, col: 25, offset: 3833},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 150, col: 25, offset: 3833},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 150, col: 29, offset: 3837},
										expr: &ruleRefExpr{
											pos:  position{line: 150, col: 29, offset: 3837},
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
			pos:  position{line: 154, col: 1, offset: 3892},
			expr: &choiceExpr{
				pos: position{line: 154, col: 11, offset: 3902},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 154, col: 11, offset: 3902},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 154, col: 17, offset: 3908},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 154, col: 17, offset: 3908},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 154, col: 30, offset: 3921},
								expr: &ruleRefExpr{
									pos:  position{line: 154, col: 30, offset: 3921},
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
			pos:  position{line: 155, col: 1, offset: 3928},
			expr: &charClassMatcher{
				pos:        position{line: 155, col: 16, offset: 3943},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 156, col: 1, offset: 3949},
			expr: &charClassMatcher{
				pos:        position{line: 156, col: 9, offset: 3957},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 158, col: 1, offset: 3964},
			expr: &actionExpr{
				pos: position{line: 158, col: 9, offset: 3972},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 158, col: 9, offset: 3972},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 158, col: 15, offset: 3978},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 162, col: 1, offset: 4026},
			expr: &actionExpr{
				pos: position{line: 163, col: 5, offset: 4076},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 163, col: 5, offset: 4076},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 163, col: 5, offset: 4076},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 163, col: 9, offset: 4080},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 163, col: 17, offset: 4088},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 163, col: 39, offset: 4110},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 167, col: 1, offset: 4152},
			expr: &actionExpr{
				pos: position{line: 168, col: 5, offset: 4178},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 168, col: 5, offset: 4178},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 168, col: 11, offset: 4184},
						expr: &ruleRefExpr{
							pos:  position{line: 168, col: 11, offset: 4184},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 172, col: 1, offset: 4262},
			expr: &choiceExpr{
				pos: position{line: 173, col: 5, offset: 4288},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 173, col: 5, offset: 4288},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 173, col: 5, offset: 4288},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 173, col: 5, offset: 4288},
									expr: &charClassMatcher{
										pos:        position{line: 173, col: 6, offset: 4289},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 173, col: 12, offset: 4295},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 173, col: 15, offset: 4298},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 5, offset: 4360},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 178, col: 1, offset: 4396},
			expr: &choiceExpr{
				pos: position{line: 179, col: 5, offset: 4435},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 179, col: 5, offset: 4435},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 179, col: 5, offset: 4435},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 182, col: 5, offset: 4473},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 182, col: 5, offset: 4473},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 182, col: 10, offset: 4478},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 184, col: 1, offset: 4510},
			expr: &actionExpr{
				pos: position{line: 185, col: 5, offset: 4545},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 185, col: 5, offset: 4545},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 185, col: 5, offset: 4545},
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 6, offset: 4546},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 21, offset: 4561},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 189, col: 1, offset: 4612},
			expr: &anyMatcher{
				line: 189, col: 14, offset: 4625,
			},
		},
		{
			name: "__",
			pos:  position{line: 191, col: 1, offset: 4628},
			expr: &zeroOrMoreExpr{
				pos: position{line: 191, col: 6, offset: 4633},
				expr: &choiceExpr{
					pos: position{line: 191, col: 8, offset: 4635},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 191, col: 8, offset: 4635},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 13, offset: 4640},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 192, col: 1, offset: 4647},
			expr: &charClassMatcher{
				pos:        position{line: 192, col: 6, offset: 4652},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 193, col: 1, offset: 4662},
			expr: &charClassMatcher{
				pos:        position{line: 194, col: 5, offset: 4681},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 196, col: 1, offset: 4689},
			expr: &litMatcher{
				pos:        position{line: 196, col: 7, offset: 4695},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 197, col: 1, offset: 4700},
			expr: &notExpr{
				pos: position{line: 197, col: 7, offset: 4706},
				expr: &anyMatcher{
					line: 197, col: 8, offset: 4707,
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
