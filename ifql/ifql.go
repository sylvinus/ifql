package ifql

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

//go:generate pigeon -o ifql.go ifql.peg

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 14, col: 1, offset: 199},
			expr: &actionExpr{
				pos: position{line: 14, col: 11, offset: 209},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 14, col: 11, offset: 209},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 14, col: 11, offset: 209},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 14, col: 14, offset: 212},
							label: "function",
							expr: &ruleRefExpr{
								pos:  position{line: 14, col: 23, offset: 221},
								name: "Function",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 14, col: 32, offset: 230},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Tests",
			pos:  position{line: 18, col: 1, offset: 276},
			expr: &choiceExpr{
				pos: position{line: 18, col: 10, offset: 285},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 18, col: 10, offset: 285},
						name: "Function",
					},
					&ruleRefExpr{
						pos:  position{line: 18, col: 21, offset: 296},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 18, col: 37, offset: 312},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 18, col: 48, offset: 323},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 18, col: 59, offset: 334},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 20, col: 1, offset: 342},
			expr: &actionExpr{
				pos: position{line: 20, col: 12, offset: 353},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 20, col: 12, offset: 353},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 20, col: 12, offset: 353},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 20, col: 17, offset: 358},
								name: "FunctionName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 20, col: 30, offset: 371},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 20, col: 34, offset: 375},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 20, col: 38, offset: 379},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 20, col: 41, offset: 382},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 20, col: 46, offset: 387},
								expr: &ruleRefExpr{
									pos:  position{line: 20, col: 46, offset: 387},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 20, col: 60, offset: 401},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 20, col: 63, offset: 404},
							val:        ")",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 20, col: 67, offset: 408},
							label: "children",
							expr: &zeroOrMoreExpr{
								pos: position{line: 20, col: 76, offset: 417},
								expr: &ruleRefExpr{
									pos:  position{line: 20, col: 76, offset: 417},
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
			pos:  position{line: 24, col: 1, offset: 491},
			expr: &actionExpr{
				pos: position{line: 24, col: 17, offset: 507},
				run: (*parser).callonFunctionChain1,
				expr: &seqExpr{
					pos: position{line: 24, col: 17, offset: 507},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 24, col: 17, offset: 507},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 24, col: 20, offset: 510},
							val:        ".",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 24, col: 24, offset: 514},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 24, col: 27, offset: 517},
							label: "child",
							expr: &ruleRefExpr{
								pos:  position{line: 24, col: 33, offset: 523},
								name: "Function",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 28, col: 1, offset: 559},
			expr: &ruleRefExpr{
				pos:  position{line: 28, col: 16, offset: 574},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 29, col: 1, offset: 581},
			expr: &actionExpr{
				pos: position{line: 29, col: 16, offset: 596},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 29, col: 16, offset: 596},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 29, col: 16, offset: 596},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 29, col: 22, offset: 602},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 29, col: 34, offset: 614},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 29, col: 37, offset: 617},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 29, col: 42, offset: 622},
								expr: &ruleRefExpr{
									pos:  position{line: 29, col: 42, offset: 622},
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
			pos:  position{line: 33, col: 1, offset: 685},
			expr: &actionExpr{
				pos: position{line: 33, col: 20, offset: 704},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 33, col: 20, offset: 704},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 33, col: 20, offset: 704},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 33, col: 24, offset: 708},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 33, col: 28, offset: 712},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 33, col: 32, offset: 716},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 37, col: 1, offset: 753},
			expr: &actionExpr{
				pos: position{line: 37, col: 15, offset: 767},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 37, col: 15, offset: 767},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 37, col: 15, offset: 767},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 20, offset: 772},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 27, offset: 779},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 37, col: 31, offset: 783},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 35, offset: 787},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 37, col: 38, offset: 790},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 42, offset: 794},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 44, col: 1, offset: 906},
			expr: &choiceExpr{
				pos: position{line: 44, col: 22, offset: 927},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 44, col: 22, offset: 927},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 34, offset: 939},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 50, offset: 955},
						name: "Regex",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 58, offset: 963},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 69, offset: 974},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 80, offset: 985},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 46, col: 1, offset: 993},
			expr: &actionExpr{
				pos: position{line: 46, col: 13, offset: 1005},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 46, col: 13, offset: 1005},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 46, col: 13, offset: 1005},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 46, col: 17, offset: 1009},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 46, col: 20, offset: 1012},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 46, col: 25, offset: 1017},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 46, col: 30, offset: 1022},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 46, col: 34, offset: 1026},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 57, col: 1, offset: 1282},
			expr: &ruleRefExpr{
				pos:  position{line: 57, col: 8, offset: 1289},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 59, col: 1, offset: 1298},
			expr: &actionExpr{
				pos: position{line: 59, col: 21, offset: 1318},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 59, col: 22, offset: 1319},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 59, col: 22, offset: 1319},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 59, col: 30, offset: 1327},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 60, col: 1, offset: 1383},
			expr: &actionExpr{
				pos: position{line: 60, col: 11, offset: 1393},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 60, col: 11, offset: 1393},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 60, col: 11, offset: 1393},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 60, col: 16, offset: 1398},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 60, col: 25, offset: 1407},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 60, col: 30, offset: 1412},
								expr: &seqExpr{
									pos: position{line: 60, col: 32, offset: 1414},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 60, col: 32, offset: 1414},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 60, col: 36, offset: 1418},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 60, col: 53, offset: 1435},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 60, col: 57, offset: 1439},
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
			pos:  position{line: 62, col: 1, offset: 1495},
			expr: &actionExpr{
				pos: position{line: 62, col: 22, offset: 1516},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 62, col: 23, offset: 1517},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 62, col: 23, offset: 1517},
							val:        "=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 62, col: 29, offset: 1523},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 63, col: 1, offset: 1561},
			expr: &actionExpr{
				pos: position{line: 63, col: 12, offset: 1572},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 63, col: 12, offset: 1572},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 63, col: 12, offset: 1572},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 63, col: 17, offset: 1577},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 63, col: 28, offset: 1588},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 63, col: 33, offset: 1593},
								expr: &seqExpr{
									pos: position{line: 63, col: 35, offset: 1595},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 63, col: 35, offset: 1595},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 38, offset: 1598},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 56, offset: 1616},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 63, col: 59, offset: 1619},
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
			pos:  position{line: 65, col: 1, offset: 1677},
			expr: &actionExpr{
				pos: position{line: 65, col: 24, offset: 1700},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 65, col: 26, offset: 1702},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 65, col: 26, offset: 1702},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 65, col: 33, offset: 1709},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 65, col: 39, offset: 1715},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 65, col: 46, offset: 1722},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 65, col: 52, offset: 1728},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 65, col: 68, offset: 1744},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 65, col: 76, offset: 1752},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 65, col: 91, offset: 1767},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 66, col: 1, offset: 1826},
			expr: &actionExpr{
				pos: position{line: 66, col: 14, offset: 1839},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 66, col: 14, offset: 1839},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 66, col: 14, offset: 1839},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 66, col: 19, offset: 1844},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 66, col: 28, offset: 1853},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 66, col: 33, offset: 1858},
								expr: &seqExpr{
									pos: position{line: 66, col: 35, offset: 1860},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 66, col: 35, offset: 1860},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 38, offset: 1863},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 58, offset: 1883},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 66, col: 61, offset: 1886},
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
			pos:  position{line: 68, col: 1, offset: 1942},
			expr: &actionExpr{
				pos: position{line: 68, col: 20, offset: 1961},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 68, col: 21, offset: 1962},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 68, col: 21, offset: 1962},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 68, col: 27, offset: 1968},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 69, col: 1, offset: 2004},
			expr: &actionExpr{
				pos: position{line: 69, col: 12, offset: 2015},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 69, col: 12, offset: 2015},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 69, col: 12, offset: 2015},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 69, col: 17, offset: 2020},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 69, col: 32, offset: 2035},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 69, col: 37, offset: 2040},
								expr: &seqExpr{
									pos: position{line: 69, col: 39, offset: 2042},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 69, col: 39, offset: 2042},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 42, offset: 2045},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 59, offset: 2062},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 62, offset: 2065},
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
			pos:  position{line: 71, col: 1, offset: 2127},
			expr: &actionExpr{
				pos: position{line: 71, col: 26, offset: 2152},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 71, col: 27, offset: 2153},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 71, col: 27, offset: 2153},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 71, col: 33, offset: 2159},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 72, col: 1, offset: 2196},
			expr: &actionExpr{
				pos: position{line: 72, col: 18, offset: 2213},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 72, col: 18, offset: 2213},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 72, col: 18, offset: 2213},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 72, col: 23, offset: 2218},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 72, col: 31, offset: 2226},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 72, col: 36, offset: 2231},
								expr: &seqExpr{
									pos: position{line: 72, col: 38, offset: 2233},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 72, col: 38, offset: 2233},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 72, col: 41, offset: 2236},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 72, col: 64, offset: 2259},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 72, col: 67, offset: 2262},
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
			pos:  position{line: 74, col: 1, offset: 2317},
			expr: &choiceExpr{
				pos: position{line: 74, col: 11, offset: 2327},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 74, col: 11, offset: 2327},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 74, col: 11, offset: 2327},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 74, col: 11, offset: 2327},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 74, col: 15, offset: 2331},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 74, col: 18, offset: 2334},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 74, col: 23, offset: 2339},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 74, col: 31, offset: 2347},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 74, col: 34, offset: 2350},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 7, offset: 2413},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 23, offset: 2429},
						name: "Regex",
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 31, offset: 2437},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 42, offset: 2448},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 53, offset: 2459},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 77, col: 62, offset: 2468},
						name: "Field",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 79, col: 1, offset: 2475},
			expr: &seqExpr{
				pos: position{line: 79, col: 16, offset: 2490},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 79, col: 16, offset: 2490},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 79, col: 22, offset: 2496},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 79, col: 28, offset: 2502},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 79, col: 34, offset: 2508},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 81, col: 1, offset: 2515},
			expr: &seqExpr{
				pos: position{line: 83, col: 5, offset: 2540},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 2540},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 11, offset: 2546},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 85, col: 1, offset: 2553},
			expr: &seqExpr{
				pos: position{line: 88, col: 5, offset: 2623},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 88, col: 5, offset: 2623},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 11, offset: 2629},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 90, col: 1, offset: 2636},
			expr: &seqExpr{
				pos: position{line: 92, col: 5, offset: 2660},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 92, col: 5, offset: 2660},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 92, col: 11, offset: 2666},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 94, col: 1, offset: 2673},
			expr: &seqExpr{
				pos: position{line: 96, col: 5, offset: 2699},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 96, col: 5, offset: 2699},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 96, col: 11, offset: 2705},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 98, col: 1, offset: 2712},
			expr: &seqExpr{
				pos: position{line: 101, col: 5, offset: 2784},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 101, col: 5, offset: 2784},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 101, col: 11, offset: 2790},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 103, col: 1, offset: 2797},
			expr: &seqExpr{
				pos: position{line: 103, col: 15, offset: 2811},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 103, col: 15, offset: 2811},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 103, col: 19, offset: 2815},
						expr: &ruleRefExpr{
							pos:  position{line: 103, col: 19, offset: 2815},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 104, col: 1, offset: 2822},
			expr: &seqExpr{
				pos: position{line: 104, col: 17, offset: 2838},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 104, col: 18, offset: 2839},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 104, col: 18, offset: 2839},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 104, col: 24, offset: 2845},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 104, col: 29, offset: 2850},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 104, col: 38, offset: 2859},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 104, col: 42, offset: 2863},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 105, col: 1, offset: 2874},
			expr: &choiceExpr{
				pos: position{line: 105, col: 15, offset: 2888},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 105, col: 15, offset: 2888},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 22, offset: 2895},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 106, col: 1, offset: 2910},
			expr: &seqExpr{
				pos: position{line: 106, col: 15, offset: 2924},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 106, col: 15, offset: 2924},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 106, col: 24, offset: 2933},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 106, col: 28, offset: 2937},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 106, col: 39, offset: 2948},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 106, col: 43, offset: 2952},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 106, col: 54, offset: 2963},
						expr: &ruleRefExpr{
							pos:  position{line: 106, col: 54, offset: 2963},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 107, col: 1, offset: 2976},
			expr: &seqExpr{
				pos: position{line: 107, col: 12, offset: 2987},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 12, offset: 2987},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 107, col: 25, offset: 3000},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 29, offset: 3004},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 107, col: 39, offset: 3014},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 43, offset: 3018},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 108, col: 1, offset: 3027},
			expr: &seqExpr{
				pos: position{line: 108, col: 12, offset: 3038},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 108, col: 12, offset: 3038},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 108, col: 24, offset: 3050},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 109, col: 1, offset: 3061},
			expr: &actionExpr{
				pos: position{line: 109, col: 12, offset: 3072},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 109, col: 12, offset: 3072},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 109, col: 12, offset: 3072},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 109, col: 21, offset: 3081},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 109, col: 26, offset: 3086},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 117, col: 1, offset: 3238},
			expr: &litMatcher{
				pos:        position{line: 117, col: 19, offset: 3256},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 118, col: 1, offset: 3261},
			expr: &choiceExpr{
				pos: position{line: 118, col: 21, offset: 3281},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 118, col: 21, offset: 3281},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 118, col: 28, offset: 3288},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 118, col: 35, offset: 3296},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 119, col: 1, offset: 3303},
			expr: &litMatcher{
				pos:        position{line: 119, col: 20, offset: 3322},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 120, col: 1, offset: 3327},
			expr: &litMatcher{
				pos:        position{line: 120, col: 15, offset: 3341},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 121, col: 1, offset: 3345},
			expr: &litMatcher{
				pos:        position{line: 121, col: 15, offset: 3359},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 122, col: 1, offset: 3363},
			expr: &litMatcher{
				pos:        position{line: 122, col: 13, offset: 3375},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 123, col: 1, offset: 3379},
			expr: &choiceExpr{
				pos: position{line: 123, col: 18, offset: 3396},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 123, col: 18, offset: 3396},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 123, col: 36, offset: 3414},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 123, col: 55, offset: 3433},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 123, col: 74, offset: 3452},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 123, col: 88, offset: 3466},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 123, col: 102, offset: 3480},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 125, col: 1, offset: 3492},
			expr: &seqExpr{
				pos: position{line: 125, col: 18, offset: 3509},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 125, col: 18, offset: 3509},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 125, col: 25, offset: 3516},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 126, col: 1, offset: 3530},
			expr: &actionExpr{
				pos: position{line: 126, col: 12, offset: 3541},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 126, col: 12, offset: 3541},
					expr: &ruleRefExpr{
						pos:  position{line: 126, col: 12, offset: 3541},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 134, col: 1, offset: 3690},
			expr: &choiceExpr{
				pos: position{line: 134, col: 17, offset: 3706},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 134, col: 17, offset: 3706},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 134, col: 19, offset: 3708},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 134, col: 19, offset: 3708},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 134, col: 23, offset: 3712},
									expr: &ruleRefExpr{
										pos:  position{line: 134, col: 23, offset: 3712},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 134, col: 41, offset: 3730},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 140, col: 5, offset: 3872},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 140, col: 7, offset: 3874},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 140, col: 7, offset: 3874},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 140, col: 11, offset: 3878},
									expr: &ruleRefExpr{
										pos:  position{line: 140, col: 11, offset: 3878},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 140, col: 31, offset: 3898},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 140, col: 31, offset: 3898},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 140, col: 37, offset: 3904},
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
			pos:  position{line: 144, col: 1, offset: 3976},
			expr: &choiceExpr{
				pos: position{line: 144, col: 20, offset: 3995},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 144, col: 20, offset: 3995},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 144, col: 20, offset: 3995},
								expr: &choiceExpr{
									pos: position{line: 144, col: 23, offset: 3998},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 144, col: 23, offset: 3998},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 144, col: 29, offset: 4004},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 144, col: 36, offset: 4011},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 144, col: 42, offset: 4017},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 144, col: 55, offset: 4030},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 144, col: 55, offset: 4030},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 144, col: 60, offset: 4035},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 146, col: 1, offset: 4055},
			expr: &choiceExpr{
				pos: position{line: 146, col: 23, offset: 4077},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 146, col: 23, offset: 4077},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 146, col: 29, offset: 4083},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 146, col: 31, offset: 4085},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 146, col: 31, offset: 4085},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 146, col: 44, offset: 4098},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 146, col: 50, offset: 4104},
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
			pos:  position{line: 150, col: 1, offset: 4170},
			expr: &actionExpr{
				pos: position{line: 150, col: 10, offset: 4179},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 150, col: 10, offset: 4179},
					expr: &ruleRefExpr{
						pos:  position{line: 150, col: 10, offset: 4179},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 154, col: 1, offset: 4227},
			expr: &seqExpr{
				pos: position{line: 154, col: 14, offset: 4240},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 154, col: 14, offset: 4240},
						expr: &choiceExpr{
							pos: position{line: 154, col: 16, offset: 4242},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 154, col: 16, offset: 4242},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 22, offset: 4248},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 28, offset: 4254},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 34, offset: 4260},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 40, offset: 4266},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 46, offset: 4272},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 52, offset: 4278},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 154, col: 58, offset: 4284},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 154, col: 63, offset: 4289},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 156, col: 1, offset: 4301},
			expr: &actionExpr{
				pos: position{line: 156, col: 10, offset: 4310},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 156, col: 10, offset: 4310},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 156, col: 10, offset: 4310},
							expr: &litMatcher{
								pos:        position{line: 156, col: 10, offset: 4310},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 156, col: 15, offset: 4315},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 156, col: 23, offset: 4323},
							expr: &seqExpr{
								pos: position{line: 156, col: 25, offset: 4325},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 156, col: 25, offset: 4325},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 156, col: 29, offset: 4329},
										expr: &ruleRefExpr{
											pos:  position{line: 156, col: 29, offset: 4329},
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
			pos:  position{line: 164, col: 1, offset: 4474},
			expr: &choiceExpr{
				pos: position{line: 164, col: 11, offset: 4484},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 164, col: 11, offset: 4484},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 164, col: 17, offset: 4490},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 164, col: 17, offset: 4490},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 164, col: 30, offset: 4503},
								expr: &ruleRefExpr{
									pos:  position{line: 164, col: 30, offset: 4503},
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
			pos:  position{line: 165, col: 1, offset: 4510},
			expr: &charClassMatcher{
				pos:        position{line: 165, col: 16, offset: 4525},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 166, col: 1, offset: 4531},
			expr: &charClassMatcher{
				pos:        position{line: 166, col: 9, offset: 4539},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 168, col: 1, offset: 4546},
			expr: &actionExpr{
				pos: position{line: 168, col: 9, offset: 4554},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 168, col: 9, offset: 4554},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 168, col: 15, offset: 4560},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "Regex",
			pos:  position{line: 172, col: 1, offset: 4594},
			expr: &actionExpr{
				pos: position{line: 172, col: 9, offset: 4602},
				run: (*parser).callonRegex1,
				expr: &seqExpr{
					pos: position{line: 172, col: 9, offset: 4602},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 172, col: 9, offset: 4602},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 172, col: 13, offset: 4606},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 172, col: 18, offset: 4611},
								name: "StringLiteral",
							},
						},
						&litMatcher{
							pos:        position{line: 172, col: 32, offset: 4625},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 177, col: 1, offset: 4730},
			expr: &anyMatcher{
				line: 177, col: 14, offset: 4743,
			},
		},
		{
			name: "__",
			pos:  position{line: 179, col: 1, offset: 4746},
			expr: &zeroOrMoreExpr{
				pos: position{line: 179, col: 6, offset: 4751},
				expr: &choiceExpr{
					pos: position{line: 179, col: 8, offset: 4753},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 179, col: 8, offset: 4753},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 13, offset: 4758},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 180, col: 1, offset: 4765},
			expr: &charClassMatcher{
				pos:        position{line: 180, col: 6, offset: 4770},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 182, col: 1, offset: 4781},
			expr: &litMatcher{
				pos:        position{line: 182, col: 7, offset: 4787},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 183, col: 1, offset: 4792},
			expr: &notExpr{
				pos: position{line: 183, col: 7, offset: 4798},
				expr: &anyMatcher{
					line: 183, col: 8, offset: 4799,
				},
			},
		},
	},
}

func (c *current) onGrammar1(function interface{}) (interface{}, error) {
	return function.(*Function), nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["function"])
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
	return NewFunctionArgs(first, rest)
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

func (c *current) onFunctionArg1(name, arg interface{}) (interface{}, error) {
	return &FunctionArg{
		Name: name.(string),
		Arg:  arg.(Arg),
	}, nil
}

func (p *parser) callonFunctionArg1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionArg1(stack["name"], stack["arg"])
}

func (c *current) onWhereExpr1(expr interface{}) (interface{}, error) {
	return &WhereExpr{Expr: expr.(*BinaryExpression)}, nil
}

func (p *parser) callonWhereExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWhereExpr1(stack["expr"])
}

func (c *current) onLogicalOperators1() (interface{}, error) {
	return strings.ToLower(string(c.text)), nil
}

func (p *parser) callonLogicalOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLogicalOperators1()
}

func (c *current) onLogical1(head, tail interface{}) (interface{}, error) {
	return NewBinaryExpression(head, tail)
}

func (p *parser) callonLogical1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLogical1(stack["head"], stack["tail"])
}

func (c *current) onEqualityOperators1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonEqualityOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEqualityOperators1()
}

func (c *current) onEquality1(head, tail interface{}) (interface{}, error) {
	return NewBinaryExpression(head, tail)
}

func (p *parser) callonEquality1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEquality1(stack["head"], stack["tail"])
}

func (c *current) onRelationalOperators1() (interface{}, error) {
	return strings.ToLower(string(c.text)), nil
}

func (p *parser) callonRelationalOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRelationalOperators1()
}

func (c *current) onRelational1(head, tail interface{}) (interface{}, error) {
	return NewBinaryExpression(head, tail)
}

func (p *parser) callonRelational1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRelational1(stack["head"], stack["tail"])
}

func (c *current) onAdditiveOperator1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonAdditiveOperator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAdditiveOperator1()
}

func (c *current) onAdditive1(head, tail interface{}) (interface{}, error) {
	return NewBinaryExpression(head, tail)
}

func (p *parser) callonAdditive1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAdditive1(stack["head"], stack["tail"])
}

func (c *current) onMultiplicativeOperator1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonMultiplicativeOperator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiplicativeOperator1()
}

func (c *current) onMultiplicative1(head, tail interface{}) (interface{}, error) {
	return NewBinaryExpression(head, tail)
}

func (p *parser) callonMultiplicative1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiplicative1(stack["head"], stack["tail"])
}

func (c *current) onPrimary2(expr interface{}) (interface{}, error) {
	return expr.(*BinaryExpression), nil

}

func (p *parser) callonPrimary2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrimary2(stack["expr"])
}

func (c *current) onDateTime1() (interface{}, error) {
	t, err := time.Parse(time.RFC3339Nano, string(c.text))
	if err != nil {
		return nil, err
	}
	return &DateTime{t}, nil
}

func (p *parser) callonDateTime1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDateTime1()
}

func (c *current) onDuration1() (interface{}, error) {
	d, err := time.ParseDuration(string(c.text))
	if err != nil {
		return nil, err
	}
	return &Duration{d}, nil
}

func (p *parser) callonDuration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDuration1()
}

func (c *current) onStringLiteral2() (interface{}, error) {
	s, err := strconv.Unquote(string(c.text))
	if err != nil {
		return nil, err
	}
	return &StringLiteral{s}, nil
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
	return string(c.text), nil
}

func (p *parser) callonString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onString1()
}

func (c *current) onNumber1() (interface{}, error) {
	n, err := strconv.ParseFloat(string(c.text), 64)
	if err != nil {
		return nil, err
	}
	return &Number{n}, nil
}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onField1(field interface{}) (interface{}, error) {
	return &Field{}, nil
}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["field"])
}

func (c *current) onRegex1(expr interface{}) (interface{}, error) {
	// TODO: perhaps we should not check regex here?
	return NewRegex(expr.(*StringLiteral))
}

func (p *parser) callonRegex1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRegex1(stack["expr"])
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
