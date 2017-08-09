package peg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// select(db:"foo").where(exp:{"t1"="val1" and "t2"="val2"}).range(start:-1h)
// select(db:"foo").where(exp:{"t1"="val1" and "t2"="val2"}).range(start:-1h).sum()
// select(db:"foo").where(exp:{"t1"="val1" and "t2"="val2"}).range(start:-1h).window(period:1m).count()

func toInterface(vals interface{}) (interface{}, error) {
	if vals == nil {
		return nil, nil
	}
	switch v := vals.(type) {
	case float64:
		return v, nil
	case string:
		return v, nil
	case time.Time:
		return v.String(), nil
	case time.Duration:
		return v.String(), nil
	default:
		log.Printf("UNKNOWN TYPE %t", v)
	}
	return nil, fmt.Errorf("Unknown types for now")
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 36, col: 1, offset: 847},
			expr: &actionExpr{
				pos: position{line: 36, col: 11, offset: 857},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 36, col: 11, offset: 857},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 36, col: 11, offset: 857},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 36, col: 14, offset: 860},
							label: "vals",
							expr: &ruleRefExpr{
								pos:  position{line: 36, col: 19, offset: 865},
								name: "Tests",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 36, col: 25, offset: 871},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Tests",
			pos:  position{line: 41, col: 1, offset: 974},
			expr: &choiceExpr{
				pos: position{line: 41, col: 10, offset: 983},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 41, col: 10, offset: 983},
						name: "Function",
					},
					&ruleRefExpr{
						pos:  position{line: 41, col: 21, offset: 994},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 41, col: 37, offset: 1010},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 41, col: 48, offset: 1021},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 41, col: 59, offset: 1032},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 43, col: 1, offset: 1040},
			expr: &seqExpr{
				pos: position{line: 43, col: 12, offset: 1051},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 43, col: 12, offset: 1051},
						name: "FunctionName",
					},
					&ruleRefExpr{
						pos:  position{line: 43, col: 25, offset: 1064},
						name: "__",
					},
					&litMatcher{
						pos:        position{line: 43, col: 29, offset: 1068},
						val:        "(",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 43, col: 33, offset: 1072},
						name: "__",
					},
					&zeroOrMoreExpr{
						pos: position{line: 43, col: 36, offset: 1075},
						expr: &ruleRefExpr{
							pos:  position{line: 43, col: 36, offset: 1075},
							name: "FunctionArgs",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 43, col: 50, offset: 1089},
						name: "__",
					},
					&litMatcher{
						pos:        position{line: 43, col: 53, offset: 1092},
						val:        ")",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 43, col: 57, offset: 1096},
						expr: &ruleRefExpr{
							pos:  position{line: 43, col: 57, offset: 1096},
							name: "FunctionChain",
						},
					},
				},
			},
		},
		{
			name: "FunctionChain",
			pos:  position{line: 44, col: 1, offset: 1111},
			expr: &seqExpr{
				pos: position{line: 44, col: 17, offset: 1127},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 44, col: 17, offset: 1127},
						name: "__",
					},
					&litMatcher{
						pos:        position{line: 44, col: 20, offset: 1130},
						val:        ".",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 24, offset: 1134},
						name: "__",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 27, offset: 1137},
						name: "Function",
					},
				},
			},
		},
		{
			name: "FunctionName",
			pos:  position{line: 45, col: 1, offset: 1146},
			expr: &ruleRefExpr{
				pos:  position{line: 45, col: 16, offset: 1161},
				name: "String",
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 46, col: 1, offset: 1168},
			expr: &seqExpr{
				pos: position{line: 46, col: 16, offset: 1183},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 46, col: 16, offset: 1183},
						name: "FunctionArg",
					},
					&ruleRefExpr{
						pos:  position{line: 46, col: 28, offset: 1195},
						name: "__",
					},
					&zeroOrMoreExpr{
						pos: position{line: 46, col: 31, offset: 1198},
						expr: &seqExpr{
							pos: position{line: 46, col: 33, offset: 1200},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 46, col: 33, offset: 1200},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 46, col: 37, offset: 1204},
									name: "FunctionArg",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 47, col: 1, offset: 1219},
			expr: &seqExpr{
				pos: position{line: 47, col: 15, offset: 1233},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 47, col: 15, offset: 1233},
						name: "String",
					},
					&ruleRefExpr{
						pos:  position{line: 47, col: 22, offset: 1240},
						name: "__",
					},
					&litMatcher{
						pos:        position{line: 47, col: 26, offset: 1244},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 47, col: 30, offset: 1248},
						name: "__",
					},
					&ruleRefExpr{
						pos:  position{line: 47, col: 33, offset: 1251},
						name: "FunctionArgValues",
					},
					&ruleRefExpr{
						pos:  position{line: 47, col: 51, offset: 1269},
						name: "__",
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 48, col: 1, offset: 1272},
			expr: &choiceExpr{
				pos: position{line: 48, col: 21, offset: 1292},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 48, col: 21, offset: 1292},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 48, col: 37, offset: 1308},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 48, col: 48, offset: 1319},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 48, col: 59, offset: 1330},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 48, col: 68, offset: 1339},
						name: "Expr",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 50, col: 1, offset: 1345},
			expr: &actionExpr{
				pos: position{line: 50, col: 13, offset: 1357},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 50, col: 13, offset: 1357},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 50, col: 13, offset: 1357},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 50, col: 17, offset: 1361},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 50, col: 20, offset: 1364},
							name: "Expr",
						},
						&ruleRefExpr{
							pos:  position{line: 50, col: 25, offset: 1369},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 50, col: 29, offset: 1373},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 55, col: 1, offset: 1429},
			expr: &seqExpr{
				pos: position{line: 55, col: 8, offset: 1436},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 55, col: 8, offset: 1436},
						name: "Primary",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 16, offset: 1444},
						name: "__",
					},
					&zeroOrMoreExpr{
						pos: position{line: 55, col: 19, offset: 1447},
						expr: &ruleRefExpr{
							pos:  position{line: 55, col: 19, offset: 1447},
							name: "BinaryExpr",
						},
					},
				},
			},
		},
		{
			name: "BinaryExpr",
			pos:  position{line: 56, col: 1, offset: 1459},
			expr: &seqExpr{
				pos: position{line: 56, col: 14, offset: 1472},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 56, col: 14, offset: 1472},
						name: "Operators",
					},
					&ruleRefExpr{
						pos:  position{line: 56, col: 24, offset: 1482},
						name: "__",
					},
					&ruleRefExpr{
						pos:  position{line: 56, col: 27, offset: 1485},
						name: "Primary",
					},
				},
			},
		},
		{
			name: "Primary",
			pos:  position{line: 57, col: 1, offset: 1493},
			expr: &choiceExpr{
				pos: position{line: 57, col: 11, offset: 1503},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 57, col: 11, offset: 1503},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 57, col: 11, offset: 1503},
								val:        "(",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 57, col: 15, offset: 1507},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 57, col: 18, offset: 1510},
								name: "Expr",
							},
							&ruleRefExpr{
								pos:  position{line: 57, col: 23, offset: 1515},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 57, col: 26, offset: 1518},
								val:        ")",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 32, offset: 1524},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 48, offset: 1540},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 59, offset: 1551},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 70, offset: 1562},
						name: "Number",
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 59, col: 1, offset: 1622},
			expr: &choiceExpr{
				pos: position{line: 59, col: 13, offset: 1634},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 59, col: 13, offset: 1634},
						val:        "<=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 20, offset: 1641},
						val:        "<",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 26, offset: 1647},
						val:        ">=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 33, offset: 1654},
						val:        ">",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 39, offset: 1660},
						val:        "=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 45, offset: 1666},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 59, col: 52, offset: 1673},
						val:        "or",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 59, col: 60, offset: 1681},
						val:        "and",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 59, col: 69, offset: 1690},
						val:        "in",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 59, col: 77, offset: 1698},
						val:        "not",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 59, col: 86, offset: 1707},
						val:        "empty",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 59, col: 97, offset: 1718},
						val:        "startswith",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 61, col: 1, offset: 1733},
			expr: &seqExpr{
				pos: position{line: 61, col: 16, offset: 1748},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 61, col: 16, offset: 1748},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 22, offset: 1754},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 28, offset: 1760},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 61, col: 34, offset: 1766},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 63, col: 1, offset: 1773},
			expr: &seqExpr{
				pos: position{line: 65, col: 5, offset: 1798},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 65, col: 5, offset: 1798},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 65, col: 11, offset: 1804},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 67, col: 1, offset: 1811},
			expr: &seqExpr{
				pos: position{line: 70, col: 5, offset: 1881},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 70, col: 5, offset: 1881},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 70, col: 11, offset: 1887},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 72, col: 1, offset: 1894},
			expr: &seqExpr{
				pos: position{line: 74, col: 5, offset: 1918},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 74, col: 5, offset: 1918},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 74, col: 11, offset: 1924},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 76, col: 1, offset: 1931},
			expr: &seqExpr{
				pos: position{line: 78, col: 5, offset: 1957},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 78, col: 5, offset: 1957},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 78, col: 11, offset: 1963},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 80, col: 1, offset: 1970},
			expr: &seqExpr{
				pos: position{line: 83, col: 5, offset: 2042},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 2042},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 11, offset: 2048},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 85, col: 1, offset: 2055},
			expr: &seqExpr{
				pos: position{line: 86, col: 5, offset: 2071},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 86, col: 5, offset: 2071},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 86, col: 9, offset: 2075},
						expr: &ruleRefExpr{
							pos:  position{line: 86, col: 9, offset: 2075},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 88, col: 1, offset: 2083},
			expr: &seqExpr{
				pos: position{line: 89, col: 5, offset: 2101},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 89, col: 6, offset: 2102},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 89, col: 6, offset: 2102},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 89, col: 12, offset: 2108},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 17, offset: 2113},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 89, col: 26, offset: 2122},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 89, col: 30, offset: 2126},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 91, col: 1, offset: 2138},
			expr: &choiceExpr{
				pos: position{line: 91, col: 15, offset: 2152},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 91, col: 15, offset: 2152},
						val:        "z",
						ignoreCase: true,
					},
					&ruleRefExpr{
						pos:  position{line: 91, col: 22, offset: 2159},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 93, col: 1, offset: 2175},
			expr: &seqExpr{
				pos: position{line: 94, col: 5, offset: 2191},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 94, col: 5, offset: 2191},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 94, col: 14, offset: 2200},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 94, col: 18, offset: 2204},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 94, col: 29, offset: 2215},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 94, col: 33, offset: 2219},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 94, col: 44, offset: 2230},
						expr: &ruleRefExpr{
							pos:  position{line: 94, col: 44, offset: 2230},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 96, col: 1, offset: 2244},
			expr: &seqExpr{
				pos: position{line: 97, col: 5, offset: 2257},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 97, col: 5, offset: 2257},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 97, col: 18, offset: 2270},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 97, col: 22, offset: 2274},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 97, col: 32, offset: 2284},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 97, col: 36, offset: 2288},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 99, col: 1, offset: 2298},
			expr: &seqExpr{
				pos: position{line: 100, col: 5, offset: 2311},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 100, col: 5, offset: 2311},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 100, col: 17, offset: 2323},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 102, col: 1, offset: 2335},
			expr: &actionExpr{
				pos: position{line: 102, col: 12, offset: 2346},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 102, col: 12, offset: 2346},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 102, col: 12, offset: 2346},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 102, col: 21, offset: 2355},
							val:        "t",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 102, col: 26, offset: 2360},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 106, col: 1, offset: 2430},
			expr: &litMatcher{
				pos:        position{line: 106, col: 19, offset: 2448},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 107, col: 1, offset: 2453},
			expr: &choiceExpr{
				pos: position{line: 107, col: 21, offset: 2473},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 107, col: 21, offset: 2473},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 107, col: 28, offset: 2480},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 107, col: 35, offset: 2488},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 108, col: 1, offset: 2495},
			expr: &litMatcher{
				pos:        position{line: 108, col: 20, offset: 2514},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 109, col: 1, offset: 2519},
			expr: &litMatcher{
				pos:        position{line: 109, col: 15, offset: 2533},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 110, col: 1, offset: 2537},
			expr: &litMatcher{
				pos:        position{line: 110, col: 15, offset: 2551},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 111, col: 1, offset: 2555},
			expr: &litMatcher{
				pos:        position{line: 111, col: 13, offset: 2567},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 112, col: 1, offset: 2571},
			expr: &choiceExpr{
				pos: position{line: 112, col: 18, offset: 2588},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 112, col: 18, offset: 2588},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 36, offset: 2606},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 55, offset: 2625},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 74, offset: 2644},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 88, offset: 2658},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 112, col: 102, offset: 2672},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 114, col: 1, offset: 2684},
			expr: &seqExpr{
				pos: position{line: 114, col: 18, offset: 2701},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 114, col: 18, offset: 2701},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 114, col: 25, offset: 2708},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 115, col: 1, offset: 2722},
			expr: &actionExpr{
				pos: position{line: 115, col: 12, offset: 2733},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 115, col: 12, offset: 2733},
					expr: &ruleRefExpr{
						pos:  position{line: 115, col: 12, offset: 2733},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 119, col: 1, offset: 2800},
			expr: &choiceExpr{
				pos: position{line: 119, col: 17, offset: 2818},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 119, col: 17, offset: 2818},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 119, col: 19, offset: 2820},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 119, col: 19, offset: 2820},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 119, col: 23, offset: 2824},
									expr: &ruleRefExpr{
										pos:  position{line: 119, col: 23, offset: 2824},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 119, col: 41, offset: 2842},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 121, col: 5, offset: 2897},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 121, col: 7, offset: 2899},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 121, col: 7, offset: 2899},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 121, col: 11, offset: 2903},
									expr: &ruleRefExpr{
										pos:  position{line: 121, col: 11, offset: 2903},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 121, col: 31, offset: 2923},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 121, col: 31, offset: 2923},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 121, col: 37, offset: 2929},
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
			pos:  position{line: 125, col: 1, offset: 3001},
			expr: &choiceExpr{
				pos: position{line: 125, col: 20, offset: 3022},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 125, col: 20, offset: 3022},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 125, col: 20, offset: 3022},
								expr: &choiceExpr{
									pos: position{line: 125, col: 23, offset: 3025},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 125, col: 23, offset: 3025},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 125, col: 29, offset: 3031},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 125, col: 36, offset: 3038},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 125, col: 42, offset: 3044},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 125, col: 55, offset: 3057},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 125, col: 55, offset: 3057},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 125, col: 60, offset: 3062},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 127, col: 1, offset: 3082},
			expr: &choiceExpr{
				pos: position{line: 127, col: 23, offset: 3106},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 127, col: 23, offset: 3106},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 127, col: 29, offset: 3112},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 127, col: 31, offset: 3114},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 127, col: 31, offset: 3114},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 127, col: 44, offset: 3127},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 127, col: 50, offset: 3133},
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
			pos:  position{line: 131, col: 1, offset: 3199},
			expr: &oneOrMoreExpr{
				pos: position{line: 131, col: 10, offset: 3208},
				expr: &ruleRefExpr{
					pos:  position{line: 131, col: 10, offset: 3208},
					name: "StringChar",
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 132, col: 1, offset: 3220},
			expr: &seqExpr{
				pos: position{line: 132, col: 14, offset: 3233},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 132, col: 14, offset: 3233},
						expr: &choiceExpr{
							pos: position{line: 132, col: 16, offset: 3235},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 132, col: 16, offset: 3235},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 22, offset: 3241},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 28, offset: 3247},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 34, offset: 3253},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 40, offset: 3259},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 46, offset: 3265},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 132, col: 52, offset: 3271},
									val:        ",",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 132, col: 57, offset: 3276},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 134, col: 1, offset: 3288},
			expr: &actionExpr{
				pos: position{line: 134, col: 10, offset: 3299},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 134, col: 10, offset: 3299},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 134, col: 10, offset: 3299},
							expr: &litMatcher{
								pos:        position{line: 134, col: 10, offset: 3299},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 134, col: 15, offset: 3304},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 134, col: 23, offset: 3312},
							expr: &seqExpr{
								pos: position{line: 134, col: 25, offset: 3314},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 134, col: 25, offset: 3314},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 134, col: 29, offset: 3318},
										expr: &ruleRefExpr{
											pos:  position{line: 134, col: 29, offset: 3318},
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
			pos:  position{line: 138, col: 1, offset: 3383},
			expr: &choiceExpr{
				pos: position{line: 138, col: 11, offset: 3395},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 138, col: 11, offset: 3395},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 138, col: 17, offset: 3401},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 138, col: 17, offset: 3401},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 138, col: 30, offset: 3414},
								expr: &ruleRefExpr{
									pos:  position{line: 138, col: 30, offset: 3414},
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
			pos:  position{line: 139, col: 1, offset: 3421},
			expr: &charClassMatcher{
				pos:        position{line: 139, col: 16, offset: 3438},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 140, col: 1, offset: 3444},
			expr: &charClassMatcher{
				pos:        position{line: 140, col: 9, offset: 3454},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 142, col: 1, offset: 3461},
			expr: &anyMatcher{
				line: 142, col: 14, offset: 3476,
			},
		},
		{
			name: "__",
			pos:  position{line: 144, col: 1, offset: 3479},
			expr: &zeroOrMoreExpr{
				pos: position{line: 144, col: 6, offset: 3486},
				expr: &choiceExpr{
					pos: position{line: 144, col: 8, offset: 3488},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 144, col: 8, offset: 3488},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 144, col: 21, offset: 3501},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 145, col: 1, offset: 3508},
			expr: &charClassMatcher{
				pos:        position{line: 145, col: 14, offset: 3523},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 147, col: 1, offset: 3532},
			expr: &litMatcher{
				pos:        position{line: 147, col: 7, offset: 3540},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 148, col: 1, offset: 3545},
			expr: &notExpr{
				pos: position{line: 148, col: 7, offset: 3553},
				expr: &anyMatcher{
					line: 148, col: 8, offset: 3554,
				},
			},
		},
	},
}

func (c *current) onGrammar1(vals interface{}) (interface{}, error) {
	return toInterface(vals)
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["vals"])
}

func (c *current) onWhereExpr1() (interface{}, error) {
	log.Printf("howdy")
	return c.text, nil
}

func (p *parser) callonWhereExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWhereExpr1()
}

func (c *current) onDateTime1() (interface{}, error) {
	return time.Parse(time.RFC3339Nano, string(c.text))
}

func (p *parser) callonDateTime1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDateTime1()
}

func (c *current) onDuration1() (interface{}, error) {
	return time.ParseDuration(string(c.text))
}

func (p *parser) callonDuration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDuration1()
}

func (c *current) onStringLiteral2() (interface{}, error) {
	return strconv.Unquote(string(c.text))
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

func (c *current) onNumber1() (interface{}, error) {
	return strconv.ParseFloat(string(c.text), 64)
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
