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

	"github.com/influxdata/ifql/ast"
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
			pos:  position{line: 11, col: 1, offset: 145},
			expr: &actionExpr{
				pos: position{line: 11, col: 18, offset: 162},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 11, col: 18, offset: 162},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 11, col: 18, offset: 162},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 11, col: 25, offset: 169},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 11, col: 32, offset: 176},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 11, col: 35, offset: 179},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 11, col: 40, offset: 184},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 11, col: 50, offset: 194},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 11, col: 58, offset: 202},
								expr: &seqExpr{
									pos: position{line: 11, col: 60, offset: 204},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 11, col: 60, offset: 204},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 11, col: 63, offset: 207},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 67, offset: 211},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 70, offset: 214},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 77, offset: 221},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 11, col: 80, offset: 224},
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
			pos:  position{line: 15, col: 1, offset: 300},
			expr: &actionExpr{
				pos: position{line: 15, col: 13, offset: 312},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 15, col: 13, offset: 312},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 15, col: 13, offset: 312},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 15, col: 17, offset: 316},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 15, col: 20, offset: 319},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 15, col: 25, offset: 324},
								expr: &ruleRefExpr{
									pos:  position{line: 15, col: 26, offset: 325},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 15, col: 41, offset: 340},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 15, col: 44, offset: 343},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 19, col: 1, offset: 373},
			expr: &actionExpr{
				pos: position{line: 19, col: 16, offset: 388},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 19, col: 16, offset: 388},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 19, col: 16, offset: 388},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 19, col: 22, offset: 394},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 19, col: 34, offset: 406},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 19, col: 37, offset: 409},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 19, col: 42, offset: 414},
								expr: &ruleRefExpr{
									pos:  position{line: 19, col: 42, offset: 414},
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
			pos:  position{line: 23, col: 1, offset: 483},
			expr: &actionExpr{
				pos: position{line: 23, col: 20, offset: 502},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 23, col: 20, offset: 502},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 23, col: 20, offset: 502},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 23, col: 24, offset: 506},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 23, col: 28, offset: 510},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 23, col: 32, offset: 514},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 27, col: 1, offset: 551},
			expr: &actionExpr{
				pos: position{line: 27, col: 15, offset: 565},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 27, col: 15, offset: 565},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 27, col: 15, offset: 565},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 27, col: 19, offset: 569},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 27, col: 26, offset: 576},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 27, col: 30, offset: 580},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 27, col: 34, offset: 584},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 27, col: 37, offset: 587},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 27, col: 43, offset: 593},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 31, col: 1, offset: 663},
			expr: &choiceExpr{
				pos: position{line: 31, col: 22, offset: 684},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 31, col: 22, offset: 684},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 34, offset: 696},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 50, offset: 712},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 77, offset: 739},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 88, offset: 750},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 31, col: 99, offset: 761},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 33, col: 1, offset: 769},
			expr: &actionExpr{
				pos: position{line: 33, col: 13, offset: 781},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 33, col: 13, offset: 781},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 33, col: 13, offset: 781},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 17, offset: 785},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 33, col: 20, offset: 788},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 33, col: 25, offset: 793},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 30, offset: 798},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 33, col: 34, offset: 802},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 44, col: 1, offset: 1038},
			expr: &ruleRefExpr{
				pos:  position{line: 44, col: 8, offset: 1045},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 46, col: 1, offset: 1054},
			expr: &actionExpr{
				pos: position{line: 46, col: 21, offset: 1074},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 46, col: 22, offset: 1075},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 46, col: 22, offset: 1075},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 46, col: 30, offset: 1083},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 49, col: 1, offset: 1124},
			expr: &actionExpr{
				pos: position{line: 49, col: 11, offset: 1134},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 49, col: 11, offset: 1134},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 49, col: 11, offset: 1134},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 49, col: 16, offset: 1139},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 49, col: 25, offset: 1148},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 49, col: 30, offset: 1153},
								expr: &seqExpr{
									pos: position{line: 49, col: 32, offset: 1155},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 49, col: 32, offset: 1155},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 36, offset: 1159},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 53, offset: 1176},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 49, col: 57, offset: 1180},
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
			pos:  position{line: 53, col: 1, offset: 1258},
			expr: &actionExpr{
				pos: position{line: 53, col: 22, offset: 1279},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 53, col: 23, offset: 1280},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 53, col: 23, offset: 1280},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 53, col: 30, offset: 1287},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 56, col: 1, offset: 1326},
			expr: &actionExpr{
				pos: position{line: 56, col: 12, offset: 1337},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 56, col: 12, offset: 1337},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 56, col: 12, offset: 1337},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 56, col: 17, offset: 1342},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 56, col: 28, offset: 1353},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 56, col: 33, offset: 1358},
								expr: &seqExpr{
									pos: position{line: 56, col: 35, offset: 1360},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 56, col: 35, offset: 1360},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 38, offset: 1363},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 56, offset: 1381},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 56, col: 59, offset: 1384},
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
			pos:  position{line: 60, col: 1, offset: 1462},
			expr: &actionExpr{
				pos: position{line: 60, col: 24, offset: 1485},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 60, col: 26, offset: 1487},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 60, col: 26, offset: 1487},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 33, offset: 1494},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 39, offset: 1500},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 46, offset: 1507},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 60, col: 52, offset: 1513},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 68, offset: 1529},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 76, offset: 1537},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 60, col: 91, offset: 1552},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 63, col: 1, offset: 1595},
			expr: &actionExpr{
				pos: position{line: 63, col: 14, offset: 1608},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 63, col: 14, offset: 1608},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 63, col: 14, offset: 1608},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 63, col: 19, offset: 1613},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 63, col: 28, offset: 1622},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 63, col: 33, offset: 1627},
								expr: &seqExpr{
									pos: position{line: 63, col: 35, offset: 1629},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 63, col: 35, offset: 1629},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 38, offset: 1632},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 58, offset: 1652},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 61, offset: 1655},
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
			pos:  position{line: 67, col: 1, offset: 1731},
			expr: &actionExpr{
				pos: position{line: 67, col: 20, offset: 1750},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 67, col: 21, offset: 1751},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 67, col: 21, offset: 1751},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 67, col: 27, offset: 1757},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 70, col: 1, offset: 1794},
			expr: &actionExpr{
				pos: position{line: 70, col: 12, offset: 1805},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 70, col: 12, offset: 1805},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 70, col: 12, offset: 1805},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 70, col: 17, offset: 1810},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 70, col: 32, offset: 1825},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 70, col: 37, offset: 1830},
								expr: &seqExpr{
									pos: position{line: 70, col: 39, offset: 1832},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 70, col: 39, offset: 1832},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 42, offset: 1835},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 59, offset: 1852},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 70, col: 62, offset: 1855},
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
			pos:  position{line: 74, col: 1, offset: 1938},
			expr: &actionExpr{
				pos: position{line: 74, col: 26, offset: 1963},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 74, col: 27, offset: 1964},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 74, col: 27, offset: 1964},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 74, col: 33, offset: 1970},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 77, col: 1, offset: 2007},
			expr: &actionExpr{
				pos: position{line: 77, col: 18, offset: 2024},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 77, col: 18, offset: 2024},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 77, col: 18, offset: 2024},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 77, col: 23, offset: 2029},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 77, col: 31, offset: 2037},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 36, offset: 2042},
								expr: &seqExpr{
									pos: position{line: 77, col: 38, offset: 2044},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 77, col: 38, offset: 2044},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 41, offset: 2047},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 64, offset: 2070},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 67, offset: 2073},
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
			pos:  position{line: 81, col: 1, offset: 2148},
			expr: &choiceExpr{
				pos: position{line: 81, col: 11, offset: 2158},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 81, col: 11, offset: 2158},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 81, col: 11, offset: 2158},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 81, col: 11, offset: 2158},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 81, col: 15, offset: 2162},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 81, col: 18, offset: 2165},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 81, col: 23, offset: 2170},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 81, col: 31, offset: 2178},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 81, col: 34, offset: 2181},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 2229},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 21, offset: 2245},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 48, offset: 2272},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 59, offset: 2283},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 70, offset: 2294},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 79, offset: 2303},
						name: "Field",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 85, col: 1, offset: 2310},
			expr: &seqExpr{
				pos: position{line: 85, col: 16, offset: 2325},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 85, col: 16, offset: 2325},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 22, offset: 2331},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 28, offset: 2337},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 34, offset: 2343},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 87, col: 1, offset: 2350},
			expr: &seqExpr{
				pos: position{line: 89, col: 5, offset: 2375},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 89, col: 5, offset: 2375},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 11, offset: 2381},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 91, col: 1, offset: 2388},
			expr: &seqExpr{
				pos: position{line: 94, col: 5, offset: 2458},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 94, col: 5, offset: 2458},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 94, col: 11, offset: 2464},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 96, col: 1, offset: 2471},
			expr: &seqExpr{
				pos: position{line: 98, col: 5, offset: 2495},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 98, col: 5, offset: 2495},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 98, col: 11, offset: 2501},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 100, col: 1, offset: 2508},
			expr: &seqExpr{
				pos: position{line: 102, col: 5, offset: 2534},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 102, col: 5, offset: 2534},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 102, col: 11, offset: 2540},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 104, col: 1, offset: 2547},
			expr: &seqExpr{
				pos: position{line: 107, col: 5, offset: 2619},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 5, offset: 2619},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 11, offset: 2625},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 109, col: 1, offset: 2632},
			expr: &seqExpr{
				pos: position{line: 109, col: 15, offset: 2646},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 109, col: 15, offset: 2646},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 109, col: 19, offset: 2650},
						expr: &ruleRefExpr{
							pos:  position{line: 109, col: 19, offset: 2650},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 110, col: 1, offset: 2657},
			expr: &seqExpr{
				pos: position{line: 110, col: 17, offset: 2673},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 110, col: 18, offset: 2674},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 110, col: 18, offset: 2674},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 110, col: 24, offset: 2680},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 110, col: 29, offset: 2685},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 110, col: 38, offset: 2694},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 110, col: 42, offset: 2698},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 111, col: 1, offset: 2709},
			expr: &choiceExpr{
				pos: position{line: 111, col: 15, offset: 2723},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 111, col: 15, offset: 2723},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 111, col: 22, offset: 2730},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 112, col: 1, offset: 2745},
			expr: &seqExpr{
				pos: position{line: 112, col: 15, offset: 2759},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 112, col: 15, offset: 2759},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 112, col: 24, offset: 2768},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 28, offset: 2772},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 112, col: 39, offset: 2783},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 43, offset: 2787},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 112, col: 54, offset: 2798},
						expr: &ruleRefExpr{
							pos:  position{line: 112, col: 54, offset: 2798},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 113, col: 1, offset: 2811},
			expr: &seqExpr{
				pos: position{line: 113, col: 12, offset: 2822},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 113, col: 12, offset: 2822},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 113, col: 25, offset: 2835},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 29, offset: 2839},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 113, col: 39, offset: 2849},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 113, col: 43, offset: 2853},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 114, col: 1, offset: 2862},
			expr: &seqExpr{
				pos: position{line: 114, col: 12, offset: 2873},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 114, col: 12, offset: 2873},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 114, col: 24, offset: 2885},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 115, col: 1, offset: 2896},
			expr: &actionExpr{
				pos: position{line: 115, col: 12, offset: 2907},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 115, col: 12, offset: 2907},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 115, col: 12, offset: 2907},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 115, col: 21, offset: 2916},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 115, col: 25, offset: 2920},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 119, col: 1, offset: 2969},
			expr: &litMatcher{
				pos:        position{line: 119, col: 19, offset: 2987},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 120, col: 1, offset: 2992},
			expr: &choiceExpr{
				pos: position{line: 120, col: 21, offset: 3012},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 120, col: 21, offset: 3012},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 120, col: 28, offset: 3019},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 120, col: 35, offset: 3027},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 121, col: 1, offset: 3034},
			expr: &litMatcher{
				pos:        position{line: 121, col: 20, offset: 3053},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 122, col: 1, offset: 3058},
			expr: &litMatcher{
				pos:        position{line: 122, col: 15, offset: 3072},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 123, col: 1, offset: 3076},
			expr: &litMatcher{
				pos:        position{line: 123, col: 15, offset: 3090},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 124, col: 1, offset: 3094},
			expr: &litMatcher{
				pos:        position{line: 124, col: 13, offset: 3106},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 125, col: 1, offset: 3110},
			expr: &choiceExpr{
				pos: position{line: 125, col: 18, offset: 3127},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 125, col: 18, offset: 3127},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 36, offset: 3145},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 55, offset: 3164},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 74, offset: 3183},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 88, offset: 3197},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 102, offset: 3211},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 127, col: 1, offset: 3223},
			expr: &seqExpr{
				pos: position{line: 127, col: 18, offset: 3240},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 127, col: 18, offset: 3240},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 127, col: 25, offset: 3247},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 128, col: 1, offset: 3261},
			expr: &actionExpr{
				pos: position{line: 128, col: 12, offset: 3272},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 128, col: 12, offset: 3272},
					expr: &ruleRefExpr{
						pos:  position{line: 128, col: 12, offset: 3272},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 132, col: 1, offset: 3335},
			expr: &choiceExpr{
				pos: position{line: 132, col: 17, offset: 3351},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 132, col: 17, offset: 3351},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 132, col: 19, offset: 3353},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 132, col: 19, offset: 3353},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 132, col: 23, offset: 3357},
									expr: &ruleRefExpr{
										pos:  position{line: 132, col: 23, offset: 3357},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 132, col: 41, offset: 3375},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 134, col: 5, offset: 3427},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 134, col: 7, offset: 3429},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 134, col: 7, offset: 3429},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 134, col: 11, offset: 3433},
									expr: &ruleRefExpr{
										pos:  position{line: 134, col: 11, offset: 3433},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 134, col: 31, offset: 3453},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 134, col: 31, offset: 3453},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 134, col: 37, offset: 3459},
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
			pos:  position{line: 138, col: 1, offset: 3531},
			expr: &choiceExpr{
				pos: position{line: 138, col: 20, offset: 3550},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 138, col: 20, offset: 3550},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 138, col: 20, offset: 3550},
								expr: &choiceExpr{
									pos: position{line: 138, col: 23, offset: 3553},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 138, col: 23, offset: 3553},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 138, col: 29, offset: 3559},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 138, col: 36, offset: 3566},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 138, col: 42, offset: 3572},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 138, col: 55, offset: 3585},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 138, col: 55, offset: 3585},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 138, col: 60, offset: 3590},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 140, col: 1, offset: 3610},
			expr: &choiceExpr{
				pos: position{line: 140, col: 23, offset: 3632},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 140, col: 23, offset: 3632},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 140, col: 29, offset: 3638},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 140, col: 31, offset: 3640},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 140, col: 31, offset: 3640},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 140, col: 44, offset: 3653},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 140, col: 50, offset: 3659},
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
			pos:  position{line: 144, col: 1, offset: 3725},
			expr: &actionExpr{
				pos: position{line: 144, col: 10, offset: 3734},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 144, col: 10, offset: 3734},
					expr: &ruleRefExpr{
						pos:  position{line: 144, col: 10, offset: 3734},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 148, col: 1, offset: 3788},
			expr: &seqExpr{
				pos: position{line: 148, col: 14, offset: 3801},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 148, col: 14, offset: 3801},
						expr: &choiceExpr{
							pos: position{line: 148, col: 16, offset: 3803},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 148, col: 16, offset: 3803},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 22, offset: 3809},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 28, offset: 3815},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 34, offset: 3821},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 40, offset: 3827},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 46, offset: 3833},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 52, offset: 3839},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 148, col: 58, offset: 3845},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 148, col: 63, offset: 3850},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 150, col: 1, offset: 3862},
			expr: &actionExpr{
				pos: position{line: 150, col: 10, offset: 3871},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 150, col: 10, offset: 3871},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 150, col: 10, offset: 3871},
							expr: &litMatcher{
								pos:        position{line: 150, col: 10, offset: 3871},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 150, col: 15, offset: 3876},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 150, col: 23, offset: 3884},
							expr: &seqExpr{
								pos: position{line: 150, col: 25, offset: 3886},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 150, col: 25, offset: 3886},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 150, col: 29, offset: 3890},
										expr: &ruleRefExpr{
											pos:  position{line: 150, col: 29, offset: 3890},
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
			pos:  position{line: 154, col: 1, offset: 3945},
			expr: &choiceExpr{
				pos: position{line: 154, col: 11, offset: 3955},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 154, col: 11, offset: 3955},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 154, col: 17, offset: 3961},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 154, col: 17, offset: 3961},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 154, col: 30, offset: 3974},
								expr: &ruleRefExpr{
									pos:  position{line: 154, col: 30, offset: 3974},
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
			pos:  position{line: 155, col: 1, offset: 3981},
			expr: &charClassMatcher{
				pos:        position{line: 155, col: 16, offset: 3996},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 156, col: 1, offset: 4002},
			expr: &charClassMatcher{
				pos:        position{line: 156, col: 9, offset: 4010},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 158, col: 1, offset: 4017},
			expr: &actionExpr{
				pos: position{line: 158, col: 9, offset: 4025},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 158, col: 9, offset: 4025},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 158, col: 15, offset: 4031},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 162, col: 1, offset: 4079},
			expr: &actionExpr{
				pos: position{line: 163, col: 5, offset: 4129},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 163, col: 5, offset: 4129},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 163, col: 5, offset: 4129},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 163, col: 9, offset: 4133},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 163, col: 17, offset: 4141},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 163, col: 39, offset: 4163},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 167, col: 1, offset: 4205},
			expr: &actionExpr{
				pos: position{line: 168, col: 5, offset: 4231},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 168, col: 5, offset: 4231},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 168, col: 11, offset: 4237},
						expr: &ruleRefExpr{
							pos:  position{line: 168, col: 11, offset: 4237},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 172, col: 1, offset: 4315},
			expr: &choiceExpr{
				pos: position{line: 173, col: 5, offset: 4341},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 173, col: 5, offset: 4341},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 173, col: 5, offset: 4341},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 173, col: 5, offset: 4341},
									expr: &charClassMatcher{
										pos:        position{line: 173, col: 6, offset: 4342},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 173, col: 12, offset: 4348},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 173, col: 15, offset: 4351},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 5, offset: 4413},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 178, col: 1, offset: 4449},
			expr: &choiceExpr{
				pos: position{line: 179, col: 5, offset: 4488},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 179, col: 5, offset: 4488},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 179, col: 5, offset: 4488},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 182, col: 5, offset: 4526},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 182, col: 5, offset: 4526},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 182, col: 10, offset: 4531},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 184, col: 1, offset: 4563},
			expr: &actionExpr{
				pos: position{line: 185, col: 5, offset: 4598},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 185, col: 5, offset: 4598},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 185, col: 5, offset: 4598},
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 6, offset: 4599},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 21, offset: 4614},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 189, col: 1, offset: 4665},
			expr: &anyMatcher{
				line: 189, col: 14, offset: 4678,
			},
		},
		{
			name: "__",
			pos:  position{line: 191, col: 1, offset: 4681},
			expr: &zeroOrMoreExpr{
				pos: position{line: 191, col: 6, offset: 4686},
				expr: &choiceExpr{
					pos: position{line: 191, col: 8, offset: 4688},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 191, col: 8, offset: 4688},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 13, offset: 4693},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 192, col: 1, offset: 4700},
			expr: &charClassMatcher{
				pos:        position{line: 192, col: 6, offset: 4705},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 193, col: 1, offset: 4715},
			expr: &charClassMatcher{
				pos:        position{line: 194, col: 5, offset: 4734},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 196, col: 1, offset: 4742},
			expr: &litMatcher{
				pos:        position{line: 196, col: 7, offset: 4748},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 197, col: 1, offset: 4753},
			expr: &notExpr{
				pos: position{line: 197, col: 7, offset: 4759},
				expr: &anyMatcher{
					line: 197, col: 8, offset: 4760,
				},
			},
		},
	},
}

func (c *current) onGrammar1(call interface{}) (interface{}, error) {
	return call.(*ast.CallExpression), nil
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
