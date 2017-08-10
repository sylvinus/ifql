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

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 5, col: 1, offset: 18},
			expr: &actionExpr{
				pos: position{line: 5, col: 11, offset: 28},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 5, col: 11, offset: 28},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 5, col: 11, offset: 28},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 5, col: 14, offset: 31},
							label: "function",
							expr: &ruleRefExpr{
								pos:  position{line: 5, col: 23, offset: 40},
								name: "Function",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 5, col: 32, offset: 49},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Tests",
			pos:  position{line: 9, col: 1, offset: 95},
			expr: &choiceExpr{
				pos: position{line: 9, col: 10, offset: 104},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 9, col: 10, offset: 104},
						name: "Function",
					},
					&ruleRefExpr{
						pos:  position{line: 9, col: 21, offset: 115},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 9, col: 37, offset: 131},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 9, col: 48, offset: 142},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 9, col: 59, offset: 153},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 11, col: 1, offset: 161},
			expr: &actionExpr{
				pos: position{line: 11, col: 12, offset: 172},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 11, col: 12, offset: 172},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 11, col: 12, offset: 172},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 11, col: 17, offset: 177},
								name: "FunctionName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 11, col: 30, offset: 190},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 11, col: 34, offset: 194},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 11, col: 38, offset: 198},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 11, col: 41, offset: 201},
							label: "args",
							expr: &zeroOrMoreExpr{
								pos: position{line: 11, col: 46, offset: 206},
								expr: &ruleRefExpr{
									pos:  position{line: 11, col: 46, offset: 206},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 11, col: 60, offset: 220},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 11, col: 63, offset: 223},
							val:        ")",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 11, col: 67, offset: 227},
							label: "children",
							expr: &zeroOrMoreExpr{
								pos: position{line: 11, col: 76, offset: 236},
								expr: &ruleRefExpr{
									pos:  position{line: 11, col: 76, offset: 236},
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
			pos:  position{line: 15, col: 1, offset: 301},
			expr: &actionExpr{
				pos: position{line: 15, col: 17, offset: 317},
				run: (*parser).callonFunctionChain1,
				expr: &seqExpr{
					pos: position{line: 15, col: 17, offset: 317},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 15, col: 17, offset: 317},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 15, col: 20, offset: 320},
							val:        ".",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 15, col: 24, offset: 324},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 15, col: 27, offset: 327},
							label: "child",
							expr: &ruleRefExpr{
								pos:  position{line: 15, col: 33, offset: 333},
								name: "Function",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 19, col: 1, offset: 369},
			expr: &ruleRefExpr{
				pos:  position{line: 19, col: 16, offset: 384},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 20, col: 1, offset: 391},
			expr: &actionExpr{
				pos: position{line: 20, col: 16, offset: 406},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 20, col: 16, offset: 406},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 20, col: 16, offset: 406},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 20, col: 22, offset: 412},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 20, col: 34, offset: 424},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 20, col: 37, offset: 427},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 20, col: 42, offset: 432},
								expr: &ruleRefExpr{
									pos:  position{line: 20, col: 42, offset: 432},
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
			pos:  position{line: 24, col: 1, offset: 495},
			expr: &actionExpr{
				pos: position{line: 24, col: 20, offset: 514},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 24, col: 20, offset: 514},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 24, col: 20, offset: 514},
							val:        ",",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 24, col: 24, offset: 518},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 24, col: 28, offset: 522},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 28, col: 1, offset: 559},
			expr: &actionExpr{
				pos: position{line: 28, col: 15, offset: 573},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 28, col: 15, offset: 573},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 28, col: 15, offset: 573},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 28, col: 20, offset: 578},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 28, col: 27, offset: 585},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 28, col: 31, offset: 589},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 28, col: 35, offset: 593},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 28, col: 38, offset: 596},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 28, col: 42, offset: 600},
								name: "FunctionArgValues",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 28, col: 60, offset: 618},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 35, col: 1, offset: 715},
			expr: &choiceExpr{
				pos: position{line: 35, col: 22, offset: 736},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 35, col: 22, offset: 736},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 35, col: 34, offset: 748},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 35, col: 50, offset: 764},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 35, col: 61, offset: 775},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 35, col: 72, offset: 786},
						name: "Number",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 37, col: 1, offset: 794},
			expr: &actionExpr{
				pos: position{line: 37, col: 13, offset: 806},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 37, col: 13, offset: 806},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 37, col: 13, offset: 806},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 17, offset: 810},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 37, col: 20, offset: 813},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 37, col: 25, offset: 818},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 37, col: 30, offset: 823},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 37, col: 34, offset: 827},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 41, col: 1, offset: 885},
			expr: &actionExpr{
				pos: position{line: 41, col: 8, offset: 892},
				run: (*parser).callonExpr1,
				expr: &seqExpr{
					pos: position{line: 41, col: 8, offset: 892},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 41, col: 8, offset: 892},
							label: "lhs",
							expr: &ruleRefExpr{
								pos:  position{line: 41, col: 12, offset: 896},
								name: "Primary",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 41, col: 20, offset: 904},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 41, col: 23, offset: 907},
							label: "rhs",
							expr: &oneOrMoreExpr{
								pos: position{line: 41, col: 27, offset: 911},
								expr: &ruleRefExpr{
									pos:  position{line: 41, col: 27, offset: 911},
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
			pos:  position{line: 45, col: 1, offset: 957},
			expr: &actionExpr{
				pos: position{line: 45, col: 14, offset: 970},
				run: (*parser).callonBinaryExpr1,
				expr: &seqExpr{
					pos: position{line: 45, col: 14, offset: 970},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 45, col: 14, offset: 970},
							label: "op",
							expr: &ruleRefExpr{
								pos:  position{line: 45, col: 17, offset: 973},
								name: "Operators",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 45, col: 27, offset: 983},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 45, col: 30, offset: 986},
							label: "rhs",
							expr: &ruleRefExpr{
								pos:  position{line: 45, col: 34, offset: 990},
								name: "Primary",
							},
						},
					},
				},
			},
		},
		{
			name: "Primary",
			pos:  position{line: 48, col: 1, offset: 1029},
			expr: &choiceExpr{
				pos: position{line: 48, col: 11, offset: 1039},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 48, col: 11, offset: 1039},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 48, col: 11, offset: 1039},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 48, col: 11, offset: 1039},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 48, col: 15, offset: 1043},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 48, col: 18, offset: 1046},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 48, col: 23, offset: 1051},
										name: "Expr",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 48, col: 28, offset: 1056},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 48, col: 31, offset: 1059},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 50, col: 5, offset: 1106},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 50, col: 21, offset: 1122},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 50, col: 32, offset: 1133},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 50, col: 43, offset: 1144},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 53, col: 1, offset: 1205},
			expr: &choiceExpr{
				pos: position{line: 53, col: 14, offset: 1218},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 53, col: 14, offset: 1218},
						name: "ComparisonOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 53, col: 36, offset: 1240},
						name: "LogicalOperators",
					},
				},
			},
		},
		{
			name: "ComparisonOperators",
			pos:  position{line: 54, col: 1, offset: 1258},
			expr: &actionExpr{
				pos: position{line: 54, col: 24, offset: 1281},
				run: (*parser).callonComparisonOperators1,
				expr: &choiceExpr{
					pos: position{line: 54, col: 25, offset: 1282},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 54, col: 25, offset: 1282},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 32, offset: 1289},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 38, offset: 1295},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 45, offset: 1302},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 51, offset: 1308},
							val:        "=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 57, offset: 1314},
							val:        "!=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 64, offset: 1321},
							val:        "startsWith",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 54, col: 79, offset: 1336},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 54, col: 87, offset: 1344},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 54, col: 102, offset: 1359},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 57, col: 1, offset: 1414},
			expr: &actionExpr{
				pos: position{line: 57, col: 21, offset: 1434},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 57, col: 22, offset: 1435},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 57, col: 22, offset: 1435},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 57, col: 30, offset: 1443},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 61, col: 1, offset: 1494},
			expr: &seqExpr{
				pos: position{line: 61, col: 16, offset: 1509},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 61, col: 16, offset: 1509},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 22, offset: 1515},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 28, offset: 1521},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 34, offset: 1527},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 63, col: 1, offset: 1534},
			expr: &seqExpr{
				pos: position{line: 65, col: 5, offset: 1559},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 65, col: 5, offset: 1559},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 65, col: 11, offset: 1565},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 67, col: 1, offset: 1572},
			expr: &seqExpr{
				pos: position{line: 70, col: 5, offset: 1642},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 70, col: 5, offset: 1642},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 70, col: 11, offset: 1648},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 72, col: 1, offset: 1655},
			expr: &seqExpr{
				pos: position{line: 74, col: 5, offset: 1679},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 74, col: 5, offset: 1679},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 74, col: 11, offset: 1685},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 76, col: 1, offset: 1692},
			expr: &seqExpr{
				pos: position{line: 78, col: 5, offset: 1718},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 78, col: 5, offset: 1718},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 78, col: 11, offset: 1724},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 80, col: 1, offset: 1731},
			expr: &seqExpr{
				pos: position{line: 83, col: 5, offset: 1803},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 1803},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 11, offset: 1809},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 85, col: 1, offset: 1816},
			expr: &seqExpr{
				pos: position{line: 85, col: 15, offset: 1830},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 85, col: 15, offset: 1830},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 85, col: 19, offset: 1834},
						expr: &ruleRefExpr{
							pos:  position{line: 85, col: 19, offset: 1834},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 86, col: 1, offset: 1841},
			expr: &seqExpr{
				pos: position{line: 86, col: 17, offset: 1857},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 86, col: 18, offset: 1858},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 86, col: 18, offset: 1858},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 86, col: 24, offset: 1864},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 29, offset: 1869},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 86, col: 38, offset: 1878},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 42, offset: 1882},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 87, col: 1, offset: 1893},
			expr: &choiceExpr{
				pos: position{line: 87, col: 15, offset: 1907},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 87, col: 15, offset: 1907},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 87, col: 22, offset: 1914},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 88, col: 1, offset: 1929},
			expr: &seqExpr{
				pos: position{line: 88, col: 15, offset: 1943},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 88, col: 15, offset: 1943},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 88, col: 24, offset: 1952},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 28, offset: 1956},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 88, col: 39, offset: 1967},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 88, col: 43, offset: 1971},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 88, col: 54, offset: 1982},
						expr: &ruleRefExpr{
							pos:  position{line: 88, col: 54, offset: 1982},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 89, col: 1, offset: 1995},
			expr: &seqExpr{
				pos: position{line: 89, col: 12, offset: 2006},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 89, col: 12, offset: 2006},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 89, col: 25, offset: 2019},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 29, offset: 2023},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 89, col: 39, offset: 2033},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 43, offset: 2037},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 90, col: 1, offset: 2046},
			expr: &seqExpr{
				pos: position{line: 90, col: 12, offset: 2057},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 90, col: 12, offset: 2057},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 24, offset: 2069},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 91, col: 1, offset: 2080},
			expr: &actionExpr{
				pos: position{line: 91, col: 12, offset: 2091},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 91, col: 12, offset: 2091},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 91, col: 12, offset: 2091},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 91, col: 21, offset: 2100},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 91, col: 26, offset: 2105},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 99, col: 1, offset: 2257},
			expr: &litMatcher{
				pos:        position{line: 99, col: 19, offset: 2275},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 100, col: 1, offset: 2280},
			expr: &choiceExpr{
				pos: position{line: 100, col: 21, offset: 2300},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 100, col: 21, offset: 2300},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 100, col: 28, offset: 2307},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 100, col: 35, offset: 2315},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 101, col: 1, offset: 2322},
			expr: &litMatcher{
				pos:        position{line: 101, col: 20, offset: 2341},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 102, col: 1, offset: 2346},
			expr: &litMatcher{
				pos:        position{line: 102, col: 15, offset: 2360},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 103, col: 1, offset: 2364},
			expr: &litMatcher{
				pos:        position{line: 103, col: 15, offset: 2378},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 104, col: 1, offset: 2382},
			expr: &litMatcher{
				pos:        position{line: 104, col: 13, offset: 2394},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 105, col: 1, offset: 2398},
			expr: &choiceExpr{
				pos: position{line: 105, col: 18, offset: 2415},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 105, col: 18, offset: 2415},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 36, offset: 2433},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 55, offset: 2452},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 74, offset: 2471},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 88, offset: 2485},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 105, col: 102, offset: 2499},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 107, col: 1, offset: 2511},
			expr: &seqExpr{
				pos: position{line: 107, col: 18, offset: 2528},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 107, col: 18, offset: 2528},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 107, col: 25, offset: 2535},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 108, col: 1, offset: 2549},
			expr: &actionExpr{
				pos: position{line: 108, col: 12, offset: 2560},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 108, col: 12, offset: 2560},
					expr: &ruleRefExpr{
						pos:  position{line: 108, col: 12, offset: 2560},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 116, col: 1, offset: 2709},
			expr: &choiceExpr{
				pos: position{line: 116, col: 17, offset: 2727},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 116, col: 17, offset: 2727},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 116, col: 19, offset: 2729},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 116, col: 19, offset: 2729},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 116, col: 23, offset: 2733},
									expr: &ruleRefExpr{
										pos:  position{line: 116, col: 23, offset: 2733},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 116, col: 41, offset: 2751},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 122, col: 5, offset: 2893},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 122, col: 7, offset: 2895},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 122, col: 7, offset: 2895},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 122, col: 11, offset: 2899},
									expr: &ruleRefExpr{
										pos:  position{line: 122, col: 11, offset: 2899},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 122, col: 31, offset: 2919},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 122, col: 31, offset: 2919},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 122, col: 37, offset: 2925},
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
			pos:  position{line: 126, col: 1, offset: 2997},
			expr: &choiceExpr{
				pos: position{line: 126, col: 20, offset: 3018},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 126, col: 20, offset: 3018},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 126, col: 20, offset: 3018},
								expr: &choiceExpr{
									pos: position{line: 126, col: 23, offset: 3021},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 126, col: 23, offset: 3021},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 126, col: 29, offset: 3027},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 126, col: 36, offset: 3034},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 126, col: 42, offset: 3040},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 126, col: 55, offset: 3053},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 126, col: 55, offset: 3053},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 126, col: 60, offset: 3058},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 128, col: 1, offset: 3078},
			expr: &choiceExpr{
				pos: position{line: 128, col: 23, offset: 3102},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 128, col: 23, offset: 3102},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 128, col: 29, offset: 3108},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 128, col: 31, offset: 3110},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 128, col: 31, offset: 3110},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 128, col: 44, offset: 3123},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 128, col: 50, offset: 3129},
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
			pos:  position{line: 132, col: 1, offset: 3195},
			expr: &actionExpr{
				pos: position{line: 132, col: 10, offset: 3204},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 132, col: 10, offset: 3204},
					expr: &ruleRefExpr{
						pos:  position{line: 132, col: 10, offset: 3204},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 136, col: 1, offset: 3252},
			expr: &seqExpr{
				pos: position{line: 136, col: 14, offset: 3265},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 136, col: 14, offset: 3265},
						expr: &choiceExpr{
							pos: position{line: 136, col: 16, offset: 3267},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 136, col: 16, offset: 3267},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 22, offset: 3273},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 28, offset: 3279},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 34, offset: 3285},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 40, offset: 3291},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 46, offset: 3297},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 136, col: 52, offset: 3303},
									val:        ",",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 57, offset: 3308},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 138, col: 1, offset: 3320},
			expr: &actionExpr{
				pos: position{line: 138, col: 10, offset: 3331},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 138, col: 10, offset: 3331},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 138, col: 10, offset: 3331},
							expr: &litMatcher{
								pos:        position{line: 138, col: 10, offset: 3331},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 138, col: 15, offset: 3336},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 138, col: 23, offset: 3344},
							expr: &seqExpr{
								pos: position{line: 138, col: 25, offset: 3346},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 138, col: 25, offset: 3346},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 138, col: 29, offset: 3350},
										expr: &ruleRefExpr{
											pos:  position{line: 138, col: 29, offset: 3350},
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
			pos:  position{line: 146, col: 1, offset: 3495},
			expr: &choiceExpr{
				pos: position{line: 146, col: 11, offset: 3507},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 146, col: 11, offset: 3507},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 146, col: 17, offset: 3513},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 146, col: 17, offset: 3513},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 146, col: 30, offset: 3526},
								expr: &ruleRefExpr{
									pos:  position{line: 146, col: 30, offset: 3526},
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
			pos:  position{line: 147, col: 1, offset: 3533},
			expr: &charClassMatcher{
				pos:        position{line: 147, col: 16, offset: 3550},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 148, col: 1, offset: 3556},
			expr: &charClassMatcher{
				pos:        position{line: 148, col: 9, offset: 3566},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 150, col: 1, offset: 3573},
			expr: &anyMatcher{
				line: 150, col: 14, offset: 3588,
			},
		},
		{
			name: "__",
			pos:  position{line: 152, col: 1, offset: 3591},
			expr: &zeroOrMoreExpr{
				pos: position{line: 152, col: 6, offset: 3598},
				expr: &choiceExpr{
					pos: position{line: 152, col: 8, offset: 3600},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 152, col: 8, offset: 3600},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 152, col: 13, offset: 3605},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 153, col: 1, offset: 3612},
			expr: &charClassMatcher{
				pos:        position{line: 153, col: 6, offset: 3619},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 155, col: 1, offset: 3630},
			expr: &litMatcher{
				pos:        position{line: 155, col: 7, offset: 3638},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 156, col: 1, offset: 3643},
			expr: &notExpr{
				pos: position{line: 156, col: 7, offset: 3651},
				expr: &anyMatcher{
					line: 156, col: 8, offset: 3652,
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
