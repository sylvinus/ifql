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

	"github.com/influxdata/ifql/query/execute/storage"
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
							label: "function",
							expr: &ruleRefExpr{
								pos:  position{line: 7, col: 23, offset: 82},
								name: "Function",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 7, col: 32, offset: 91},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Tests",
			pos:  position{line: 11, col: 1, offset: 137},
			expr: &choiceExpr{
				pos: position{line: 11, col: 10, offset: 146},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 11, col: 10, offset: 146},
						name: "Function",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 21, offset: 157},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 37, offset: 173},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 48, offset: 184},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 11, col: 59, offset: 195},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 13, col: 1, offset: 203},
			expr: &actionExpr{
				pos: position{line: 13, col: 12, offset: 214},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 13, col: 12, offset: 214},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 13, col: 12, offset: 214},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 13, col: 17, offset: 219},
								name: "FunctionName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 30, offset: 232},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 13, col: 34, offset: 236},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 38, offset: 240},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 13, col: 41, offset: 243},
							label: "args",
							expr: &zeroOrMoreExpr{
								pos: position{line: 13, col: 46, offset: 248},
								expr: &ruleRefExpr{
									pos:  position{line: 13, col: 46, offset: 248},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 13, col: 60, offset: 262},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 13, col: 63, offset: 265},
							val:        ")",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 13, col: 67, offset: 269},
							label: "children",
							expr: &zeroOrMoreExpr{
								pos: position{line: 13, col: 76, offset: 278},
								expr: &ruleRefExpr{
									pos:  position{line: 13, col: 76, offset: 278},
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
			pos:  position{line: 17, col: 1, offset: 343},
			expr: &actionExpr{
				pos: position{line: 17, col: 17, offset: 359},
				run: (*parser).callonFunctionChain1,
				expr: &seqExpr{
					pos: position{line: 17, col: 17, offset: 359},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 17, col: 17, offset: 359},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 17, col: 20, offset: 362},
							val:        ".",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 17, col: 24, offset: 366},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 17, col: 27, offset: 369},
							label: "child",
							expr: &ruleRefExpr{
								pos:  position{line: 17, col: 33, offset: 375},
								name: "Function",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 21, col: 1, offset: 411},
			expr: &ruleRefExpr{
				pos:  position{line: 21, col: 16, offset: 426},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 22, col: 1, offset: 433},
			expr: &actionExpr{
				pos: position{line: 22, col: 16, offset: 448},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 22, col: 16, offset: 448},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 22, col: 16, offset: 448},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 22, col: 22, offset: 454},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 22, col: 34, offset: 466},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 22, col: 37, offset: 469},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 22, col: 42, offset: 474},
								expr: &ruleRefExpr{
									pos:  position{line: 22, col: 42, offset: 474},
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
			pos:  position{line: 26, col: 1, offset: 537},
			expr: &actionExpr{
				pos: position{line: 26, col: 20, offset: 556},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 26, col: 20, offset: 556},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 26, col: 20, offset: 556},
							val:        ",",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 26, col: 24, offset: 560},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 26, col: 28, offset: 564},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 30, col: 1, offset: 601},
			expr: &actionExpr{
				pos: position{line: 30, col: 15, offset: 615},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 30, col: 15, offset: 615},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 30, col: 15, offset: 615},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 20, offset: 620},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 27, offset: 627},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 30, col: 31, offset: 631},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 35, offset: 635},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 30, col: 38, offset: 638},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 42, offset: 642},
								name: "FunctionArgValues",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 60, offset: 660},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 37, col: 1, offset: 757},
			expr: &choiceExpr{
				pos: position{line: 37, col: 22, offset: 778},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 37, col: 22, offset: 778},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 34, offset: 790},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 50, offset: 806},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 61, offset: 817},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 72, offset: 828},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 39, col: 1, offset: 836},
			expr: &actionExpr{
				pos: position{line: 39, col: 13, offset: 848},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 39, col: 13, offset: 848},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 39, col: 13, offset: 848},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 39, col: 17, offset: 852},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 39, col: 20, offset: 855},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 39, col: 25, offset: 860},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 39, col: 30, offset: 865},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 39, col: 34, offset: 869},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 43, col: 1, offset: 927},
			expr: &actionExpr{
				pos: position{line: 43, col: 8, offset: 934},
				run: (*parser).callonExpr1,
				expr: &seqExpr{
					pos: position{line: 43, col: 8, offset: 934},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 43, col: 8, offset: 934},
							label: "lhs",
							expr: &ruleRefExpr{
								pos:  position{line: 43, col: 12, offset: 938},
								name: "Primary",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 43, col: 20, offset: 946},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 43, col: 23, offset: 949},
							label: "rhs",
							expr: &oneOrMoreExpr{
								pos: position{line: 43, col: 27, offset: 953},
								expr: &ruleRefExpr{
									pos:  position{line: 43, col: 27, offset: 953},
									name: "BinaryExpr",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BinaryExpr",
			pos:  position{line: 47, col: 1, offset: 999},
			expr: &actionExpr{
				pos: position{line: 47, col: 14, offset: 1012},
				run: (*parser).callonBinaryExpr1,
				expr: &seqExpr{
					pos: position{line: 47, col: 14, offset: 1012},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 47, col: 14, offset: 1012},
							label: "op",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 17, offset: 1015},
								name: "Operators",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 47, col: 27, offset: 1025},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 47, col: 30, offset: 1028},
							label: "rhs",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 34, offset: 1032},
								name: "Primary",
							},
						},
					},
				},
			},
		},
		{
			name: "Primary",
			pos:  position{line: 50, col: 1, offset: 1071},
			expr: &choiceExpr{
				pos: position{line: 50, col: 11, offset: 1081},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 50, col: 11, offset: 1081},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 50, col: 11, offset: 1081},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 50, col: 11, offset: 1081},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 50, col: 15, offset: 1085},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 50, col: 18, offset: 1088},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 50, col: 23, offset: 1093},
										name: "Expr",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 50, col: 28, offset: 1098},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 50, col: 31, offset: 1101},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 5, offset: 1148},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 21, offset: 1164},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 32, offset: 1175},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 43, offset: 1186},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 55, col: 1, offset: 1247},
			expr: &choiceExpr{
				pos: position{line: 55, col: 14, offset: 1260},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 55, col: 14, offset: 1260},
						name: "ComparisonOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 36, offset: 1282},
						name: "LogicalOperators",
					},
				},
			},
		},
		{
			name: "ComparisonOperators",
			pos:  position{line: 56, col: 1, offset: 1300},
			expr: &actionExpr{
				pos: position{line: 56, col: 24, offset: 1323},
				run: (*parser).callonComparisonOperators1,
				expr: &choiceExpr{
					pos: position{line: 56, col: 25, offset: 1324},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 56, col: 25, offset: 1324},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 32, offset: 1331},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 38, offset: 1337},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 45, offset: 1344},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 51, offset: 1350},
							val:        "=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 57, offset: 1356},
							val:        "!=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 64, offset: 1363},
							val:        "startsWith",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 79, offset: 1378},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 56, col: 87, offset: 1386},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 56, col: 102, offset: 1401},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 59, col: 1, offset: 1456},
			expr: &actionExpr{
				pos: position{line: 59, col: 21, offset: 1476},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 59, col: 22, offset: 1477},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 59, col: 22, offset: 1477},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 59, col: 30, offset: 1485},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 63, col: 1, offset: 1536},
			expr: &seqExpr{
				pos: position{line: 63, col: 16, offset: 1551},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 63, col: 16, offset: 1551},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 22, offset: 1557},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 28, offset: 1563},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 34, offset: 1569},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 65, col: 1, offset: 1576},
			expr: &seqExpr{
				pos: position{line: 67, col: 5, offset: 1601},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 67, col: 5, offset: 1601},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 67, col: 11, offset: 1607},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 69, col: 1, offset: 1614},
			expr: &seqExpr{
				pos: position{line: 72, col: 5, offset: 1684},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 72, col: 5, offset: 1684},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 72, col: 11, offset: 1690},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 74, col: 1, offset: 1697},
			expr: &seqExpr{
				pos: position{line: 76, col: 5, offset: 1721},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 76, col: 5, offset: 1721},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 76, col: 11, offset: 1727},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 78, col: 1, offset: 1734},
			expr: &seqExpr{
				pos: position{line: 80, col: 5, offset: 1760},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 80, col: 5, offset: 1760},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 80, col: 11, offset: 1766},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 82, col: 1, offset: 1773},
			expr: &seqExpr{
				pos: position{line: 85, col: 5, offset: 1845},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 85, col: 5, offset: 1845},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 11, offset: 1851},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 87, col: 1, offset: 1858},
			expr: &seqExpr{
				pos: position{line: 87, col: 15, offset: 1872},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 87, col: 15, offset: 1872},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 87, col: 19, offset: 1876},
						expr: &ruleRefExpr{
							pos:  position{line: 87, col: 19, offset: 1876},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 88, col: 1, offset: 1883},
			expr: &seqExpr{
				pos: position{line: 88, col: 17, offset: 1899},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 88, col: 18, offset: 1900},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 88, col: 18, offset: 1900},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 88, col: 24, offset: 1906},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 29, offset: 1911},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 88, col: 38, offset: 1920},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 42, offset: 1924},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 89, col: 1, offset: 1935},
			expr: &choiceExpr{
				pos: position{line: 89, col: 15, offset: 1949},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 89, col: 15, offset: 1949},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 22, offset: 1956},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 90, col: 1, offset: 1971},
			expr: &seqExpr{
				pos: position{line: 90, col: 15, offset: 1985},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 90, col: 15, offset: 1985},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 90, col: 24, offset: 1994},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 28, offset: 1998},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 90, col: 39, offset: 2009},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 43, offset: 2013},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 90, col: 54, offset: 2024},
						expr: &ruleRefExpr{
							pos:  position{line: 90, col: 54, offset: 2024},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 91, col: 1, offset: 2037},
			expr: &seqExpr{
				pos: position{line: 91, col: 12, offset: 2048},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 91, col: 12, offset: 2048},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 91, col: 25, offset: 2061},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 91, col: 29, offset: 2065},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 91, col: 39, offset: 2075},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 91, col: 43, offset: 2079},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 92, col: 1, offset: 2088},
			expr: &seqExpr{
				pos: position{line: 92, col: 12, offset: 2099},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 92, col: 12, offset: 2099},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 92, col: 24, offset: 2111},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 93, col: 1, offset: 2122},
			expr: &actionExpr{
				pos: position{line: 93, col: 12, offset: 2133},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 93, col: 12, offset: 2133},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 93, col: 12, offset: 2133},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 93, col: 21, offset: 2142},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 93, col: 26, offset: 2147},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 101, col: 1, offset: 2299},
			expr: &litMatcher{
				pos:        position{line: 101, col: 19, offset: 2317},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 102, col: 1, offset: 2322},
			expr: &choiceExpr{
				pos: position{line: 102, col: 21, offset: 2342},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 102, col: 21, offset: 2342},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 102, col: 28, offset: 2349},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 102, col: 35, offset: 2357},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 103, col: 1, offset: 2364},
			expr: &litMatcher{
				pos:        position{line: 103, col: 20, offset: 2383},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 104, col: 1, offset: 2388},
			expr: &litMatcher{
				pos:        position{line: 104, col: 15, offset: 2402},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 105, col: 1, offset: 2406},
			expr: &litMatcher{
				pos:        position{line: 105, col: 15, offset: 2420},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 106, col: 1, offset: 2424},
			expr: &litMatcher{
				pos:        position{line: 106, col: 13, offset: 2436},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 107, col: 1, offset: 2440},
			expr: &choiceExpr{
				pos: position{line: 107, col: 18, offset: 2457},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 18, offset: 2457},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 36, offset: 2475},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 55, offset: 2494},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 74, offset: 2513},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 88, offset: 2527},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 102, offset: 2541},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 109, col: 1, offset: 2553},
			expr: &seqExpr{
				pos: position{line: 109, col: 18, offset: 2570},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 109, col: 18, offset: 2570},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 25, offset: 2577},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 110, col: 1, offset: 2591},
			expr: &actionExpr{
				pos: position{line: 110, col: 12, offset: 2602},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 110, col: 12, offset: 2602},
					expr: &ruleRefExpr{
						pos:  position{line: 110, col: 12, offset: 2602},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 118, col: 1, offset: 2751},
			expr: &choiceExpr{
				pos: position{line: 118, col: 17, offset: 2769},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 118, col: 17, offset: 2769},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 118, col: 19, offset: 2771},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 118, col: 19, offset: 2771},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 118, col: 23, offset: 2775},
									expr: &ruleRefExpr{
										pos:  position{line: 118, col: 23, offset: 2775},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 118, col: 41, offset: 2793},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 124, col: 5, offset: 2935},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 124, col: 7, offset: 2937},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 124, col: 7, offset: 2937},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 124, col: 11, offset: 2941},
									expr: &ruleRefExpr{
										pos:  position{line: 124, col: 11, offset: 2941},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 124, col: 31, offset: 2961},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 124, col: 31, offset: 2961},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 124, col: 37, offset: 2967},
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
			pos:  position{line: 128, col: 1, offset: 3039},
			expr: &choiceExpr{
				pos: position{line: 128, col: 20, offset: 3060},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 128, col: 20, offset: 3060},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 128, col: 20, offset: 3060},
								expr: &choiceExpr{
									pos: position{line: 128, col: 23, offset: 3063},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 128, col: 23, offset: 3063},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 128, col: 29, offset: 3069},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 128, col: 36, offset: 3076},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 128, col: 42, offset: 3082},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 128, col: 55, offset: 3095},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 128, col: 55, offset: 3095},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 128, col: 60, offset: 3100},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 130, col: 1, offset: 3120},
			expr: &choiceExpr{
				pos: position{line: 130, col: 23, offset: 3144},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 130, col: 23, offset: 3144},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 130, col: 29, offset: 3150},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 130, col: 31, offset: 3152},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 130, col: 31, offset: 3152},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 130, col: 44, offset: 3165},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 130, col: 50, offset: 3171},
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
			pos:  position{line: 134, col: 1, offset: 3237},
			expr: &actionExpr{
				pos: position{line: 134, col: 10, offset: 3246},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 134, col: 10, offset: 3246},
					expr: &ruleRefExpr{
						pos:  position{line: 134, col: 10, offset: 3246},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 138, col: 1, offset: 3294},
			expr: &seqExpr{
				pos: position{line: 138, col: 14, offset: 3307},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 138, col: 14, offset: 3307},
						expr: &choiceExpr{
							pos: position{line: 138, col: 16, offset: 3309},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 138, col: 16, offset: 3309},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 22, offset: 3315},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 28, offset: 3321},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 34, offset: 3327},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 40, offset: 3333},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 46, offset: 3339},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 52, offset: 3345},
									val:        ",",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 138, col: 57, offset: 3350},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 140, col: 1, offset: 3362},
			expr: &actionExpr{
				pos: position{line: 140, col: 10, offset: 3373},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 140, col: 10, offset: 3373},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 140, col: 10, offset: 3373},
							expr: &litMatcher{
								pos:        position{line: 140, col: 10, offset: 3373},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 140, col: 15, offset: 3378},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 140, col: 23, offset: 3386},
							expr: &seqExpr{
								pos: position{line: 140, col: 25, offset: 3388},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 140, col: 25, offset: 3388},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 140, col: 29, offset: 3392},
										expr: &ruleRefExpr{
											pos:  position{line: 140, col: 29, offset: 3392},
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
			pos:  position{line: 148, col: 1, offset: 3537},
			expr: &choiceExpr{
				pos: position{line: 148, col: 11, offset: 3549},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 148, col: 11, offset: 3549},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 148, col: 17, offset: 3555},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 148, col: 17, offset: 3555},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 148, col: 30, offset: 3568},
								expr: &ruleRefExpr{
									pos:  position{line: 148, col: 30, offset: 3568},
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
			pos:  position{line: 149, col: 1, offset: 3575},
			expr: &charClassMatcher{
				pos:        position{line: 149, col: 16, offset: 3592},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 150, col: 1, offset: 3598},
			expr: &charClassMatcher{
				pos:        position{line: 150, col: 9, offset: 3608},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 152, col: 1, offset: 3615},
			expr: &anyMatcher{
				line: 152, col: 14, offset: 3630,
			},
		},
		{
			name: "__",
			pos:  position{line: 154, col: 1, offset: 3633},
			expr: &zeroOrMoreExpr{
				pos: position{line: 154, col: 6, offset: 3640},
				expr: &choiceExpr{
					pos: position{line: 154, col: 8, offset: 3642},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 154, col: 8, offset: 3642},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 154, col: 13, offset: 3647},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 155, col: 1, offset: 3654},
			expr: &charClassMatcher{
				pos:        position{line: 155, col: 6, offset: 3661},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 157, col: 1, offset: 3672},
			expr: &litMatcher{
				pos:        position{line: 157, col: 7, offset: 3680},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 158, col: 1, offset: 3685},
			expr: &notExpr{
				pos: position{line: 158, col: 7, offset: 3693},
				expr: &anyMatcher{
					line: 158, col: 8, offset: 3694,
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
	return NewFunction(name, args, children)
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
	return &WhereExpr{expr.(*storage.Node)}, nil
}

func (p *parser) callonWhereExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWhereExpr1(stack["expr"])
}

func (c *current) onExpr1(lhs, rhs interface{}) (interface{}, error) {
	return NewExpr(lhs, rhs)
}

func (p *parser) callonExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onExpr1(stack["lhs"], stack["rhs"])
}

func (c *current) onBinaryExpr1(op, rhs interface{}) (interface{}, error) {
	return NewRHS(op, rhs)
}

func (p *parser) callonBinaryExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBinaryExpr1(stack["op"], stack["rhs"])
}

func (c *current) onPrimary2(expr interface{}) (interface{}, error) {
	return expr.(*storage.Node), nil
}

func (p *parser) callonPrimary2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrimary2(stack["expr"])
}

func (c *current) onComparisonOperators1() (interface{}, error) {
	return NewComparisonOperator(c.text)
}

func (p *parser) callonComparisonOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onComparisonOperators1()
}

func (c *current) onLogicalOperators1() (interface{}, error) {
	return NewLogicalOperator(c.text)
}

func (p *parser) callonLogicalOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLogicalOperators1()
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
