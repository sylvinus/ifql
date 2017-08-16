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
							expr: &zeroOrOneExpr{
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
			pos:  position{line: 17, col: 1, offset: 352},
			expr: &actionExpr{
				pos: position{line: 17, col: 17, offset: 368},
				run: (*parser).callonFunctionChain1,
				expr: &seqExpr{
					pos: position{line: 17, col: 17, offset: 368},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 17, col: 17, offset: 368},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 17, col: 20, offset: 371},
							val:        ".",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 17, col: 24, offset: 375},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 17, col: 27, offset: 378},
							label: "child",
							expr: &ruleRefExpr{
								pos:  position{line: 17, col: 33, offset: 384},
								name: "Function",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 21, col: 1, offset: 420},
			expr: &ruleRefExpr{
				pos:  position{line: 21, col: 16, offset: 435},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 22, col: 1, offset: 442},
			expr: &actionExpr{
				pos: position{line: 22, col: 16, offset: 457},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 22, col: 16, offset: 457},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 22, col: 16, offset: 457},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 22, col: 22, offset: 463},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 22, col: 34, offset: 475},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 22, col: 37, offset: 478},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 22, col: 42, offset: 483},
								expr: &ruleRefExpr{
									pos:  position{line: 22, col: 42, offset: 483},
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
			pos:  position{line: 26, col: 1, offset: 546},
			expr: &actionExpr{
				pos: position{line: 26, col: 20, offset: 565},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 26, col: 20, offset: 565},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 26, col: 20, offset: 565},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 26, col: 24, offset: 569},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 26, col: 28, offset: 573},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 26, col: 32, offset: 577},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 30, col: 1, offset: 614},
			expr: &actionExpr{
				pos: position{line: 30, col: 15, offset: 628},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 30, col: 15, offset: 628},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 30, col: 15, offset: 628},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 20, offset: 633},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 27, offset: 640},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 30, col: 31, offset: 644},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 30, col: 35, offset: 648},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 30, col: 38, offset: 651},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 30, col: 42, offset: 655},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 37, col: 1, offset: 767},
			expr: &choiceExpr{
				pos: position{line: 37, col: 22, offset: 788},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 37, col: 22, offset: 788},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 34, offset: 800},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 50, offset: 816},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 61, offset: 827},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 72, offset: 838},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 39, col: 1, offset: 846},
			expr: &actionExpr{
				pos: position{line: 39, col: 13, offset: 858},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 39, col: 13, offset: 858},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 39, col: 13, offset: 858},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 39, col: 17, offset: 862},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 39, col: 20, offset: 865},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 39, col: 25, offset: 870},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 39, col: 30, offset: 875},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 39, col: 34, offset: 879},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 43, col: 1, offset: 937},
			expr: &actionExpr{
				pos: position{line: 43, col: 8, offset: 944},
				run: (*parser).callonExpr1,
				expr: &seqExpr{
					pos: position{line: 43, col: 8, offset: 944},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 43, col: 8, offset: 944},
							label: "lhs",
							expr: &ruleRefExpr{
								pos:  position{line: 43, col: 12, offset: 948},
								name: "Primary",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 43, col: 20, offset: 956},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 43, col: 23, offset: 959},
							label: "rhs",
							expr: &oneOrMoreExpr{
								pos: position{line: 43, col: 27, offset: 963},
								expr: &ruleRefExpr{
									pos:  position{line: 43, col: 27, offset: 963},
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
			pos:  position{line: 47, col: 1, offset: 1009},
			expr: &actionExpr{
				pos: position{line: 47, col: 14, offset: 1022},
				run: (*parser).callonBinaryExpr1,
				expr: &seqExpr{
					pos: position{line: 47, col: 14, offset: 1022},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 47, col: 14, offset: 1022},
							label: "op",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 17, offset: 1025},
								name: "Operators",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 47, col: 27, offset: 1035},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 47, col: 30, offset: 1038},
							label: "rhs",
							expr: &ruleRefExpr{
								pos:  position{line: 47, col: 34, offset: 1042},
								name: "Primary",
							},
						},
					},
				},
			},
		},
		{
			name: "Primary",
			pos:  position{line: 50, col: 1, offset: 1081},
			expr: &choiceExpr{
				pos: position{line: 50, col: 11, offset: 1091},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 50, col: 11, offset: 1091},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 50, col: 11, offset: 1091},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 50, col: 11, offset: 1091},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 50, col: 15, offset: 1095},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 50, col: 18, offset: 1098},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 50, col: 23, offset: 1103},
										name: "Expr",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 50, col: 28, offset: 1108},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 50, col: 31, offset: 1111},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 5, offset: 1158},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 21, offset: 1174},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 32, offset: 1185},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 43, offset: 1196},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 52, col: 52, offset: 1205},
						name: "Field",
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 55, col: 1, offset: 1265},
			expr: &choiceExpr{
				pos: position{line: 55, col: 14, offset: 1278},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 55, col: 14, offset: 1278},
						name: "ComparisonOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 36, offset: 1300},
						name: "LogicalOperators",
					},
				},
			},
		},
		{
			name: "ComparisonOperators",
			pos:  position{line: 56, col: 1, offset: 1318},
			expr: &actionExpr{
				pos: position{line: 56, col: 24, offset: 1341},
				run: (*parser).callonComparisonOperators1,
				expr: &choiceExpr{
					pos: position{line: 56, col: 25, offset: 1342},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 56, col: 25, offset: 1342},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 32, offset: 1349},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 38, offset: 1355},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 45, offset: 1362},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 51, offset: 1368},
							val:        "=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 57, offset: 1374},
							val:        "!=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 64, offset: 1381},
							val:        "startsWith",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 56, col: 79, offset: 1396},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 56, col: 87, offset: 1404},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 56, col: 102, offset: 1419},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 59, col: 1, offset: 1474},
			expr: &actionExpr{
				pos: position{line: 59, col: 21, offset: 1494},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 59, col: 22, offset: 1495},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 59, col: 22, offset: 1495},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 59, col: 30, offset: 1503},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 63, col: 1, offset: 1554},
			expr: &seqExpr{
				pos: position{line: 63, col: 16, offset: 1569},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 63, col: 16, offset: 1569},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 22, offset: 1575},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 28, offset: 1581},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 63, col: 34, offset: 1587},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 65, col: 1, offset: 1594},
			expr: &seqExpr{
				pos: position{line: 67, col: 5, offset: 1619},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 67, col: 5, offset: 1619},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 67, col: 11, offset: 1625},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 69, col: 1, offset: 1632},
			expr: &seqExpr{
				pos: position{line: 72, col: 5, offset: 1702},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 72, col: 5, offset: 1702},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 72, col: 11, offset: 1708},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 74, col: 1, offset: 1715},
			expr: &seqExpr{
				pos: position{line: 76, col: 5, offset: 1739},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 76, col: 5, offset: 1739},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 76, col: 11, offset: 1745},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 78, col: 1, offset: 1752},
			expr: &seqExpr{
				pos: position{line: 80, col: 5, offset: 1778},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 80, col: 5, offset: 1778},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 80, col: 11, offset: 1784},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 82, col: 1, offset: 1791},
			expr: &seqExpr{
				pos: position{line: 85, col: 5, offset: 1863},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 85, col: 5, offset: 1863},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 11, offset: 1869},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 87, col: 1, offset: 1876},
			expr: &seqExpr{
				pos: position{line: 87, col: 15, offset: 1890},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 87, col: 15, offset: 1890},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 87, col: 19, offset: 1894},
						expr: &ruleRefExpr{
							pos:  position{line: 87, col: 19, offset: 1894},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 88, col: 1, offset: 1901},
			expr: &seqExpr{
				pos: position{line: 88, col: 17, offset: 1917},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 88, col: 18, offset: 1918},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 88, col: 18, offset: 1918},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 88, col: 24, offset: 1924},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 29, offset: 1929},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 88, col: 38, offset: 1938},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 42, offset: 1942},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 89, col: 1, offset: 1953},
			expr: &choiceExpr{
				pos: position{line: 89, col: 15, offset: 1967},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 89, col: 15, offset: 1967},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 22, offset: 1974},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 90, col: 1, offset: 1989},
			expr: &seqExpr{
				pos: position{line: 90, col: 15, offset: 2003},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 90, col: 15, offset: 2003},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 90, col: 24, offset: 2012},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 28, offset: 2016},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 90, col: 39, offset: 2027},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 43, offset: 2031},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 90, col: 54, offset: 2042},
						expr: &ruleRefExpr{
							pos:  position{line: 90, col: 54, offset: 2042},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 91, col: 1, offset: 2055},
			expr: &seqExpr{
				pos: position{line: 91, col: 12, offset: 2066},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 91, col: 12, offset: 2066},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 91, col: 25, offset: 2079},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 91, col: 29, offset: 2083},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 91, col: 39, offset: 2093},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 91, col: 43, offset: 2097},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 92, col: 1, offset: 2106},
			expr: &seqExpr{
				pos: position{line: 92, col: 12, offset: 2117},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 92, col: 12, offset: 2117},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 92, col: 24, offset: 2129},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 93, col: 1, offset: 2140},
			expr: &actionExpr{
				pos: position{line: 93, col: 12, offset: 2151},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 93, col: 12, offset: 2151},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 93, col: 12, offset: 2151},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 93, col: 21, offset: 2160},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 93, col: 26, offset: 2165},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 101, col: 1, offset: 2317},
			expr: &litMatcher{
				pos:        position{line: 101, col: 19, offset: 2335},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 102, col: 1, offset: 2340},
			expr: &choiceExpr{
				pos: position{line: 102, col: 21, offset: 2360},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 102, col: 21, offset: 2360},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 102, col: 28, offset: 2367},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 102, col: 35, offset: 2375},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 103, col: 1, offset: 2382},
			expr: &litMatcher{
				pos:        position{line: 103, col: 20, offset: 2401},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 104, col: 1, offset: 2406},
			expr: &litMatcher{
				pos:        position{line: 104, col: 15, offset: 2420},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 105, col: 1, offset: 2424},
			expr: &litMatcher{
				pos:        position{line: 105, col: 15, offset: 2438},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 106, col: 1, offset: 2442},
			expr: &litMatcher{
				pos:        position{line: 106, col: 13, offset: 2454},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 107, col: 1, offset: 2458},
			expr: &choiceExpr{
				pos: position{line: 107, col: 18, offset: 2475},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 18, offset: 2475},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 36, offset: 2493},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 55, offset: 2512},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 74, offset: 2531},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 88, offset: 2545},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 102, offset: 2559},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 109, col: 1, offset: 2571},
			expr: &seqExpr{
				pos: position{line: 109, col: 18, offset: 2588},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 109, col: 18, offset: 2588},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 25, offset: 2595},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 110, col: 1, offset: 2609},
			expr: &actionExpr{
				pos: position{line: 110, col: 12, offset: 2620},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 110, col: 12, offset: 2620},
					expr: &ruleRefExpr{
						pos:  position{line: 110, col: 12, offset: 2620},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 118, col: 1, offset: 2769},
			expr: &choiceExpr{
				pos: position{line: 118, col: 17, offset: 2787},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 118, col: 17, offset: 2787},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 118, col: 19, offset: 2789},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 118, col: 19, offset: 2789},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 118, col: 23, offset: 2793},
									expr: &ruleRefExpr{
										pos:  position{line: 118, col: 23, offset: 2793},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 118, col: 41, offset: 2811},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 124, col: 5, offset: 2953},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 124, col: 7, offset: 2955},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 124, col: 7, offset: 2955},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 124, col: 11, offset: 2959},
									expr: &ruleRefExpr{
										pos:  position{line: 124, col: 11, offset: 2959},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 124, col: 31, offset: 2979},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 124, col: 31, offset: 2979},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 124, col: 37, offset: 2985},
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
			pos:  position{line: 128, col: 1, offset: 3057},
			expr: &choiceExpr{
				pos: position{line: 128, col: 20, offset: 3078},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 128, col: 20, offset: 3078},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 128, col: 20, offset: 3078},
								expr: &choiceExpr{
									pos: position{line: 128, col: 23, offset: 3081},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 128, col: 23, offset: 3081},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 128, col: 29, offset: 3087},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 128, col: 36, offset: 3094},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 128, col: 42, offset: 3100},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 128, col: 55, offset: 3113},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 128, col: 55, offset: 3113},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 128, col: 60, offset: 3118},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 130, col: 1, offset: 3138},
			expr: &choiceExpr{
				pos: position{line: 130, col: 23, offset: 3162},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 130, col: 23, offset: 3162},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 130, col: 29, offset: 3168},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 130, col: 31, offset: 3170},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 130, col: 31, offset: 3170},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 130, col: 44, offset: 3183},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 130, col: 50, offset: 3189},
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
			pos:  position{line: 134, col: 1, offset: 3255},
			expr: &actionExpr{
				pos: position{line: 134, col: 10, offset: 3264},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 134, col: 10, offset: 3264},
					expr: &ruleRefExpr{
						pos:  position{line: 134, col: 10, offset: 3264},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 138, col: 1, offset: 3312},
			expr: &seqExpr{
				pos: position{line: 138, col: 14, offset: 3325},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 138, col: 14, offset: 3325},
						expr: &choiceExpr{
							pos: position{line: 138, col: 16, offset: 3327},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 138, col: 16, offset: 3327},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 22, offset: 3333},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 28, offset: 3339},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 34, offset: 3345},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 40, offset: 3351},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 46, offset: 3357},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 52, offset: 3363},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 138, col: 58, offset: 3369},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 138, col: 63, offset: 3374},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 140, col: 1, offset: 3386},
			expr: &actionExpr{
				pos: position{line: 140, col: 10, offset: 3397},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 140, col: 10, offset: 3397},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 140, col: 10, offset: 3397},
							expr: &litMatcher{
								pos:        position{line: 140, col: 10, offset: 3397},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 140, col: 15, offset: 3402},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 140, col: 23, offset: 3410},
							expr: &seqExpr{
								pos: position{line: 140, col: 25, offset: 3412},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 140, col: 25, offset: 3412},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 140, col: 29, offset: 3416},
										expr: &ruleRefExpr{
											pos:  position{line: 140, col: 29, offset: 3416},
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
			pos:  position{line: 148, col: 1, offset: 3561},
			expr: &choiceExpr{
				pos: position{line: 148, col: 11, offset: 3573},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 148, col: 11, offset: 3573},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 148, col: 17, offset: 3579},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 148, col: 17, offset: 3579},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 148, col: 30, offset: 3592},
								expr: &ruleRefExpr{
									pos:  position{line: 148, col: 30, offset: 3592},
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
			pos:  position{line: 149, col: 1, offset: 3599},
			expr: &charClassMatcher{
				pos:        position{line: 149, col: 16, offset: 3616},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 150, col: 1, offset: 3622},
			expr: &charClassMatcher{
				pos:        position{line: 150, col: 9, offset: 3632},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 152, col: 1, offset: 3639},
			expr: &actionExpr{
				pos: position{line: 152, col: 9, offset: 3647},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 152, col: 9, offset: 3647},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 152, col: 15, offset: 3653},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 156, col: 1, offset: 3687},
			expr: &anyMatcher{
				line: 156, col: 14, offset: 3702,
			},
		},
		{
			name: "__",
			pos:  position{line: 158, col: 1, offset: 3705},
			expr: &zeroOrMoreExpr{
				pos: position{line: 158, col: 6, offset: 3712},
				expr: &choiceExpr{
					pos: position{line: 158, col: 8, offset: 3714},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 158, col: 8, offset: 3714},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 158, col: 13, offset: 3719},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 159, col: 1, offset: 3726},
			expr: &charClassMatcher{
				pos:        position{line: 159, col: 6, offset: 3733},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 161, col: 1, offset: 3744},
			expr: &litMatcher{
				pos:        position{line: 161, col: 7, offset: 3752},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 162, col: 1, offset: 3757},
			expr: &notExpr{
				pos: position{line: 162, col: 7, offset: 3765},
				expr: &anyMatcher{
					line: 162, col: 8, offset: 3766,
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

func (c *current) onField1(field interface{}) (interface{}, error) {
	return &Field{}, nil
}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["field"])
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
