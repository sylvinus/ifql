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
				pos: position{line: 8, col: 5, offset: 72},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 8, col: 5, offset: 72},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 8, col: 5, offset: 72},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 8, col: 8, offset: 75},
							label: "call",
							expr: &ruleRefExpr{
								pos:  position{line: 8, col: 13, offset: 80},
								name: "CallExpression",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 8, col: 28, offset: 95},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Program",
			pos:  position{line: 12, col: 1, offset: 155},
			expr: &labeledExpr{
				pos:   position{line: 13, col: 5, offset: 167},
				label: "body",
				expr: &zeroOrOneExpr{
					pos: position{line: 13, col: 10, offset: 172},
					expr: &ruleRefExpr{
						pos:  position{line: 13, col: 10, offset: 172},
						name: "SourceElements",
					},
				},
			},
		},
		{
			name: "SourceElements",
			pos:  position{line: 15, col: 1, offset: 189},
			expr: &seqExpr{
				pos: position{line: 16, col: 5, offset: 208},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 16, col: 5, offset: 208},
						label: "head",
						expr: &ruleRefExpr{
							pos:  position{line: 16, col: 10, offset: 213},
							name: "SourceElement",
						},
					},
					&labeledExpr{
						pos:   position{line: 16, col: 24, offset: 227},
						label: "tail",
						expr: &zeroOrMoreExpr{
							pos: position{line: 16, col: 29, offset: 232},
							expr: &seqExpr{
								pos: position{line: 16, col: 30, offset: 233},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 16, col: 30, offset: 233},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 16, col: 33, offset: 236},
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
			pos:  position{line: 18, col: 1, offset: 253},
			expr: &ruleRefExpr{
				pos:  position{line: 19, col: 5, offset: 271},
				name: "Statement",
			},
		},
		{
			name: "Statement",
			pos:  position{line: 21, col: 1, offset: 282},
			expr: &choiceExpr{
				pos: position{line: 22, col: 5, offset: 296},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 22, col: 5, offset: 296},
						name: "VariableStatement",
					},
					&ruleRefExpr{
						pos:  position{line: 23, col: 5, offset: 318},
						name: "ExpressionStatement",
					},
				},
			},
		},
		{
			name: "VariableStatement",
			pos:  position{line: 25, col: 1, offset: 339},
			expr: &seqExpr{
				pos: position{line: 26, col: 5, offset: 361},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 26, col: 5, offset: 361},
						name: "VarToken",
					},
					&ruleRefExpr{
						pos:  position{line: 26, col: 14, offset: 370},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 26, col: 17, offset: 373},
						label: "declarations",
						expr: &ruleRefExpr{
							pos:  position{line: 26, col: 30, offset: 386},
							name: "VariableDeclarationList",
						},
					},
				},
			},
		},
		{
			name: "VariableDeclarationList",
			pos:  position{line: 28, col: 1, offset: 411},
			expr: &seqExpr{
				pos: position{line: 29, col: 5, offset: 439},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 29, col: 5, offset: 439},
						label: "head",
						expr: &ruleRefExpr{
							pos:  position{line: 29, col: 10, offset: 444},
							name: "VariableDeclaration",
						},
					},
					&labeledExpr{
						pos:   position{line: 29, col: 30, offset: 464},
						label: "tail",
						expr: &zeroOrMoreExpr{
							pos: position{line: 29, col: 35, offset: 469},
							expr: &seqExpr{
								pos: position{line: 29, col: 36, offset: 470},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 29, col: 36, offset: 470},
										name: "__",
									},
									&litMatcher{
										pos:        position{line: 29, col: 39, offset: 473},
										val:        ",",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 43, offset: 477},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 29, col: 46, offset: 480},
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
			pos:  position{line: 31, col: 1, offset: 503},
			expr: &litMatcher{
				pos:        position{line: 31, col: 12, offset: 514},
				val:        "var",
				ignoreCase: false,
			},
		},
		{
			name: "VariableDeclaration",
			pos:  position{line: 33, col: 1, offset: 521},
			expr: &seqExpr{
				pos: position{line: 34, col: 5, offset: 545},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 34, col: 5, offset: 545},
						label: "id",
						expr: &ruleRefExpr{
							pos:  position{line: 34, col: 8, offset: 548},
							name: "String",
						},
					},
					&labeledExpr{
						pos:   position{line: 34, col: 15, offset: 555},
						label: "initExpr",
						expr: &zeroOrOneExpr{
							pos: position{line: 34, col: 24, offset: 564},
							expr: &seqExpr{
								pos: position{line: 34, col: 26, offset: 566},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 34, col: 26, offset: 566},
										name: "__",
									},
									&ruleRefExpr{
										pos:  position{line: 34, col: 29, offset: 569},
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
			pos:  position{line: 36, col: 1, offset: 584},
			expr: &seqExpr{
				pos: position{line: 37, col: 5, offset: 600},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 37, col: 5, offset: 600},
						val:        "=",
						ignoreCase: false,
					},
					&notExpr{
						pos: position{line: 37, col: 9, offset: 604},
						expr: &litMatcher{
							pos:        position{line: 37, col: 10, offset: 605},
							val:        "=",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 37, col: 14, offset: 609},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 37, col: 17, offset: 612},
						label: "expression",
						expr: &ruleRefExpr{
							pos:  position{line: 37, col: 28, offset: 623},
							name: "VariableExpression",
						},
					},
				},
			},
		},
		{
			name: "VariableExpression",
			pos:  position{line: 40, col: 1, offset: 713},
			expr: &choiceExpr{
				pos: position{line: 41, col: 5, offset: 736},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 41, col: 5, offset: 736},
						name: "CallExpression",
					},
					&ruleRefExpr{
						pos:  position{line: 42, col: 5, offset: 755},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 43, col: 5, offset: 773},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 5, offset: 802},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 45, col: 5, offset: 815},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 46, col: 5, offset: 828},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 47, col: 5, offset: 839},
						name: "Field",
					},
				},
			},
		},
		{
			name: "ExpressionStatement",
			pos:  position{line: 51, col: 1, offset: 902},
			expr: &ruleRefExpr{
				pos:  position{line: 52, col: 5, offset: 926},
				name: "CallExpression",
			},
		},
		{
			name: "CallExpression",
			pos:  position{line: 54, col: 1, offset: 942},
			expr: &actionExpr{
				pos: position{line: 55, col: 5, offset: 961},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 55, col: 5, offset: 961},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 55, col: 5, offset: 961},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 55, col: 12, offset: 968},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 55, col: 19, offset: 975},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 55, col: 22, offset: 978},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 55, col: 27, offset: 983},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 55, col: 37, offset: 993},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 55, col: 45, offset: 1001},
								expr: &seqExpr{
									pos: position{line: 55, col: 47, offset: 1003},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 55, col: 47, offset: 1003},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 55, col: 50, offset: 1006},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 55, col: 54, offset: 1010},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 55, col: 57, offset: 1013},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 55, col: 64, offset: 1020},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 55, col: 67, offset: 1023},
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
			pos:  position{line: 59, col: 1, offset: 1105},
			expr: &actionExpr{
				pos: position{line: 60, col: 5, offset: 1119},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 60, col: 5, offset: 1119},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 60, col: 5, offset: 1119},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 60, col: 9, offset: 1123},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 60, col: 12, offset: 1126},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 60, col: 17, offset: 1131},
								expr: &ruleRefExpr{
									pos:  position{line: 60, col: 18, offset: 1132},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 60, col: 33, offset: 1147},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 60, col: 36, offset: 1150},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 64, col: 1, offset: 1186},
			expr: &actionExpr{
				pos: position{line: 65, col: 5, offset: 1203},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 65, col: 5, offset: 1203},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 65, col: 5, offset: 1203},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 65, col: 11, offset: 1209},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 65, col: 23, offset: 1221},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 65, col: 26, offset: 1224},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 65, col: 31, offset: 1229},
								expr: &ruleRefExpr{
									pos:  position{line: 65, col: 31, offset: 1229},
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
			pos:  position{line: 69, col: 1, offset: 1304},
			expr: &actionExpr{
				pos: position{line: 70, col: 5, offset: 1325},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 70, col: 5, offset: 1325},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 70, col: 5, offset: 1325},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 70, col: 9, offset: 1329},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 70, col: 13, offset: 1333},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 70, col: 17, offset: 1337},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 74, col: 1, offset: 1380},
			expr: &actionExpr{
				pos: position{line: 75, col: 5, offset: 1396},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 75, col: 5, offset: 1396},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 75, col: 5, offset: 1396},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 75, col: 9, offset: 1400},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 75, col: 16, offset: 1407},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 75, col: 20, offset: 1411},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 75, col: 24, offset: 1415},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 75, col: 27, offset: 1418},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 75, col: 33, offset: 1424},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 79, col: 1, offset: 1500},
			expr: &choiceExpr{
				pos: position{line: 80, col: 5, offset: 1522},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 80, col: 5, offset: 1522},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 81, col: 5, offset: 1536},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 82, col: 5, offset: 1554},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 1583},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 84, col: 5, offset: 1596},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 85, col: 5, offset: 1609},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 86, col: 5, offset: 1620},
						name: "String",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 88, col: 1, offset: 1628},
			expr: &actionExpr{
				pos: position{line: 89, col: 5, offset: 1642},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 89, col: 5, offset: 1642},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 89, col: 5, offset: 1642},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 89, col: 9, offset: 1646},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 89, col: 12, offset: 1649},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 17, offset: 1654},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 89, col: 22, offset: 1659},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 89, col: 26, offset: 1663},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 100, col: 1, offset: 1886},
			expr: &ruleRefExpr{
				pos:  position{line: 101, col: 5, offset: 1895},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 103, col: 1, offset: 1904},
			expr: &actionExpr{
				pos: position{line: 104, col: 5, offset: 1925},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 104, col: 6, offset: 1926},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 104, col: 6, offset: 1926},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 104, col: 14, offset: 1934},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 108, col: 1, offset: 1986},
			expr: &actionExpr{
				pos: position{line: 109, col: 5, offset: 1998},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 109, col: 5, offset: 1998},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 109, col: 5, offset: 1998},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 10, offset: 2003},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 109, col: 19, offset: 2012},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 109, col: 24, offset: 2017},
								expr: &seqExpr{
									pos: position{line: 109, col: 26, offset: 2019},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 109, col: 26, offset: 2019},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 109, col: 30, offset: 2023},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 109, col: 47, offset: 2040},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 109, col: 51, offset: 2044},
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
			pos:  position{line: 113, col: 1, offset: 2123},
			expr: &actionExpr{
				pos: position{line: 114, col: 5, offset: 2145},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 114, col: 6, offset: 2146},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 114, col: 6, offset: 2146},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 114, col: 13, offset: 2153},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 118, col: 1, offset: 2199},
			expr: &actionExpr{
				pos: position{line: 119, col: 5, offset: 2212},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 119, col: 5, offset: 2212},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 119, col: 5, offset: 2212},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 119, col: 10, offset: 2217},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 119, col: 21, offset: 2228},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 119, col: 26, offset: 2233},
								expr: &seqExpr{
									pos: position{line: 119, col: 28, offset: 2235},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 119, col: 28, offset: 2235},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 119, col: 31, offset: 2238},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 119, col: 49, offset: 2256},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 119, col: 52, offset: 2259},
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
			pos:  position{line: 123, col: 1, offset: 2339},
			expr: &actionExpr{
				pos: position{line: 124, col: 5, offset: 2363},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 124, col: 9, offset: 2367},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 124, col: 9, offset: 2367},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 125, col: 9, offset: 2380},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 126, col: 9, offset: 2392},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 127, col: 9, offset: 2405},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 128, col: 9, offset: 2417},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 129, col: 9, offset: 2439},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 130, col: 9, offset: 2453},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 131, col: 9, offset: 2474},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 136, col: 1, offset: 2532},
			expr: &actionExpr{
				pos: position{line: 137, col: 5, offset: 2547},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 137, col: 5, offset: 2547},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 137, col: 5, offset: 2547},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 137, col: 10, offset: 2552},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 137, col: 19, offset: 2561},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 137, col: 24, offset: 2566},
								expr: &seqExpr{
									pos: position{line: 137, col: 26, offset: 2568},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 137, col: 26, offset: 2568},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 29, offset: 2571},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 49, offset: 2591},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 52, offset: 2594},
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
			pos:  position{line: 141, col: 1, offset: 2672},
			expr: &actionExpr{
				pos: position{line: 142, col: 5, offset: 2693},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 142, col: 6, offset: 2694},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 142, col: 6, offset: 2694},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 142, col: 12, offset: 2700},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 146, col: 1, offset: 2748},
			expr: &actionExpr{
				pos: position{line: 147, col: 5, offset: 2761},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 147, col: 5, offset: 2761},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 147, col: 5, offset: 2761},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 147, col: 10, offset: 2766},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 147, col: 25, offset: 2781},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 147, col: 30, offset: 2786},
								expr: &seqExpr{
									pos: position{line: 147, col: 32, offset: 2788},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 147, col: 32, offset: 2788},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 147, col: 35, offset: 2791},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 147, col: 52, offset: 2808},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 147, col: 55, offset: 2811},
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
			pos:  position{line: 151, col: 1, offset: 2896},
			expr: &actionExpr{
				pos: position{line: 152, col: 5, offset: 2923},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 152, col: 6, offset: 2924},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 152, col: 6, offset: 2924},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 152, col: 12, offset: 2930},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 156, col: 1, offset: 2974},
			expr: &actionExpr{
				pos: position{line: 157, col: 5, offset: 2993},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 157, col: 5, offset: 2993},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 157, col: 5, offset: 2993},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 157, col: 10, offset: 2998},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 157, col: 18, offset: 3006},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 157, col: 23, offset: 3011},
								expr: &seqExpr{
									pos: position{line: 157, col: 25, offset: 3013},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 157, col: 25, offset: 3013},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 28, offset: 3016},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 51, offset: 3039},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 54, offset: 3042},
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
			pos:  position{line: 161, col: 1, offset: 3119},
			expr: &choiceExpr{
				pos: position{line: 162, col: 5, offset: 3131},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 162, col: 5, offset: 3131},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 162, col: 5, offset: 3131},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 162, col: 5, offset: 3131},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 162, col: 9, offset: 3135},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 162, col: 12, offset: 3138},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 162, col: 17, offset: 3143},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 162, col: 25, offset: 3151},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 162, col: 28, offset: 3154},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 5, offset: 3193},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 5, offset: 3211},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 5, offset: 3240},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 5, offset: 3253},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 5, offset: 3266},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 5, offset: 3277},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 171, col: 5, offset: 3287},
						name: "String",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 173, col: 1, offset: 3295},
			expr: &seqExpr{
				pos: position{line: 174, col: 5, offset: 3312},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 174, col: 5, offset: 3312},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 11, offset: 3318},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 17, offset: 3324},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 23, offset: 3330},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 176, col: 1, offset: 3337},
			expr: &seqExpr{
				pos: position{line: 178, col: 5, offset: 3362},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 178, col: 5, offset: 3362},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 11, offset: 3368},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 180, col: 1, offset: 3375},
			expr: &seqExpr{
				pos: position{line: 183, col: 5, offset: 3445},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 183, col: 5, offset: 3445},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 183, col: 11, offset: 3451},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 185, col: 1, offset: 3458},
			expr: &seqExpr{
				pos: position{line: 187, col: 5, offset: 3482},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 187, col: 5, offset: 3482},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 187, col: 11, offset: 3488},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 189, col: 1, offset: 3495},
			expr: &seqExpr{
				pos: position{line: 191, col: 5, offset: 3521},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 191, col: 5, offset: 3521},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 191, col: 11, offset: 3527},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 193, col: 1, offset: 3534},
			expr: &seqExpr{
				pos: position{line: 196, col: 5, offset: 3606},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 196, col: 5, offset: 3606},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 196, col: 11, offset: 3612},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 198, col: 1, offset: 3619},
			expr: &seqExpr{
				pos: position{line: 199, col: 5, offset: 3635},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 199, col: 5, offset: 3635},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 199, col: 9, offset: 3639},
						expr: &ruleRefExpr{
							pos:  position{line: 199, col: 9, offset: 3639},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 201, col: 1, offset: 3647},
			expr: &seqExpr{
				pos: position{line: 202, col: 5, offset: 3665},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 202, col: 6, offset: 3666},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 202, col: 6, offset: 3666},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 202, col: 12, offset: 3672},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 17, offset: 3677},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 202, col: 26, offset: 3686},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 30, offset: 3690},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 204, col: 1, offset: 3702},
			expr: &choiceExpr{
				pos: position{line: 205, col: 6, offset: 3718},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 205, col: 6, offset: 3718},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 12, offset: 3724},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 207, col: 1, offset: 3740},
			expr: &seqExpr{
				pos: position{line: 208, col: 5, offset: 3756},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 208, col: 5, offset: 3756},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 208, col: 14, offset: 3765},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 208, col: 18, offset: 3769},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 208, col: 29, offset: 3780},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 208, col: 33, offset: 3784},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 208, col: 44, offset: 3795},
						expr: &ruleRefExpr{
							pos:  position{line: 208, col: 44, offset: 3795},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 210, col: 1, offset: 3809},
			expr: &seqExpr{
				pos: position{line: 211, col: 5, offset: 3822},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 211, col: 5, offset: 3822},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 211, col: 18, offset: 3835},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 211, col: 22, offset: 3839},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 211, col: 32, offset: 3849},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 211, col: 36, offset: 3853},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 213, col: 1, offset: 3863},
			expr: &seqExpr{
				pos: position{line: 214, col: 5, offset: 3876},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 214, col: 5, offset: 3876},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 214, col: 17, offset: 3888},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 216, col: 1, offset: 3900},
			expr: &actionExpr{
				pos: position{line: 217, col: 5, offset: 3913},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 217, col: 5, offset: 3913},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 217, col: 5, offset: 3913},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 217, col: 14, offset: 3922},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 217, col: 18, offset: 3926},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 221, col: 1, offset: 3981},
			expr: &litMatcher{
				pos:        position{line: 222, col: 5, offset: 4001},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 224, col: 1, offset: 4007},
			expr: &choiceExpr{
				pos: position{line: 225, col: 6, offset: 4029},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 225, col: 6, offset: 4029},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 225, col: 13, offset: 4036},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 225, col: 20, offset: 4044},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 227, col: 1, offset: 4052},
			expr: &litMatcher{
				pos:        position{line: 228, col: 5, offset: 4073},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 230, col: 1, offset: 4079},
			expr: &litMatcher{
				pos:        position{line: 231, col: 5, offset: 4095},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 233, col: 1, offset: 4100},
			expr: &litMatcher{
				pos:        position{line: 234, col: 5, offset: 4116},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 236, col: 1, offset: 4121},
			expr: &litMatcher{
				pos:        position{line: 237, col: 5, offset: 4135},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 239, col: 1, offset: 4140},
			expr: &choiceExpr{
				pos: position{line: 241, col: 9, offset: 4168},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 241, col: 9, offset: 4168},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 242, col: 9, offset: 4192},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 243, col: 9, offset: 4217},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 244, col: 9, offset: 4242},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 245, col: 9, offset: 4262},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 246, col: 9, offset: 4282},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 249, col: 1, offset: 4299},
			expr: &seqExpr{
				pos: position{line: 250, col: 5, offset: 4318},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 250, col: 5, offset: 4318},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 250, col: 12, offset: 4325},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 252, col: 1, offset: 4340},
			expr: &actionExpr{
				pos: position{line: 253, col: 5, offset: 4353},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 253, col: 5, offset: 4353},
					expr: &ruleRefExpr{
						pos:  position{line: 253, col: 5, offset: 4353},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 257, col: 1, offset: 4422},
			expr: &choiceExpr{
				pos: position{line: 258, col: 5, offset: 4440},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 258, col: 5, offset: 4440},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 258, col: 7, offset: 4442},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 258, col: 7, offset: 4442},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 258, col: 11, offset: 4446},
									expr: &ruleRefExpr{
										pos:  position{line: 258, col: 11, offset: 4446},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 258, col: 29, offset: 4464},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 261, col: 5, offset: 4528},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 261, col: 7, offset: 4530},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 261, col: 7, offset: 4530},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 261, col: 11, offset: 4534},
									expr: &ruleRefExpr{
										pos:  position{line: 261, col: 11, offset: 4534},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 261, col: 31, offset: 4554},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 261, col: 31, offset: 4554},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 261, col: 37, offset: 4560},
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
			pos:  position{line: 265, col: 1, offset: 4642},
			expr: &choiceExpr{
				pos: position{line: 266, col: 5, offset: 4663},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 266, col: 5, offset: 4663},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 266, col: 5, offset: 4663},
								expr: &choiceExpr{
									pos: position{line: 266, col: 8, offset: 4666},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 266, col: 8, offset: 4666},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 266, col: 14, offset: 4672},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 266, col: 21, offset: 4679},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 266, col: 27, offset: 4685},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 267, col: 5, offset: 4700},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 267, col: 5, offset: 4700},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 267, col: 10, offset: 4705},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 269, col: 1, offset: 4725},
			expr: &choiceExpr{
				pos: position{line: 270, col: 5, offset: 4748},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 270, col: 5, offset: 4748},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 271, col: 5, offset: 4756},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 271, col: 7, offset: 4758},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 271, col: 7, offset: 4758},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 20, offset: 4771},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 26, offset: 4777},
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
			pos:  position{line: 275, col: 1, offset: 4849},
			expr: &actionExpr{
				pos: position{line: 276, col: 5, offset: 4860},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 276, col: 5, offset: 4860},
					expr: &ruleRefExpr{
						pos:  position{line: 276, col: 5, offset: 4860},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 280, col: 1, offset: 4920},
			expr: &seqExpr{
				pos: position{line: 281, col: 5, offset: 4935},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 281, col: 5, offset: 4935},
						expr: &choiceExpr{
							pos: position{line: 281, col: 7, offset: 4937},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 281, col: 7, offset: 4937},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 13, offset: 4943},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 19, offset: 4949},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 25, offset: 4955},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 31, offset: 4961},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 37, offset: 4967},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 43, offset: 4973},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 281, col: 49, offset: 4979},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 281, col: 54, offset: 4984},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 283, col: 1, offset: 4996},
			expr: &actionExpr{
				pos: position{line: 284, col: 5, offset: 5007},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 284, col: 5, offset: 5007},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 284, col: 5, offset: 5007},
							expr: &litMatcher{
								pos:        position{line: 284, col: 5, offset: 5007},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 284, col: 10, offset: 5012},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 284, col: 18, offset: 5020},
							expr: &seqExpr{
								pos: position{line: 284, col: 20, offset: 5022},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 284, col: 20, offset: 5022},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 284, col: 24, offset: 5026},
										expr: &ruleRefExpr{
											pos:  position{line: 284, col: 24, offset: 5026},
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
			pos:  position{line: 288, col: 1, offset: 5087},
			expr: &choiceExpr{
				pos: position{line: 289, col: 5, offset: 5099},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 289, col: 5, offset: 5099},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 290, col: 5, offset: 5107},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 290, col: 5, offset: 5107},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 290, col: 18, offset: 5120},
								expr: &ruleRefExpr{
									pos:  position{line: 290, col: 18, offset: 5120},
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
			pos:  position{line: 292, col: 1, offset: 5128},
			expr: &charClassMatcher{
				pos:        position{line: 293, col: 5, offset: 5145},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 295, col: 1, offset: 5152},
			expr: &charClassMatcher{
				pos:        position{line: 296, col: 5, offset: 5162},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 298, col: 1, offset: 5169},
			expr: &actionExpr{
				pos: position{line: 299, col: 5, offset: 5179},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 299, col: 5, offset: 5179},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 299, col: 11, offset: 5185},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 303, col: 1, offset: 5239},
			expr: &actionExpr{
				pos: position{line: 304, col: 5, offset: 5289},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 304, col: 5, offset: 5289},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 304, col: 5, offset: 5289},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 304, col: 9, offset: 5293},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 304, col: 17, offset: 5301},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 304, col: 39, offset: 5323},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 308, col: 1, offset: 5365},
			expr: &actionExpr{
				pos: position{line: 309, col: 5, offset: 5391},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 309, col: 5, offset: 5391},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 309, col: 11, offset: 5397},
						expr: &ruleRefExpr{
							pos:  position{line: 309, col: 11, offset: 5397},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 313, col: 1, offset: 5475},
			expr: &choiceExpr{
				pos: position{line: 314, col: 5, offset: 5501},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 314, col: 5, offset: 5501},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 314, col: 5, offset: 5501},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 314, col: 5, offset: 5501},
									expr: &charClassMatcher{
										pos:        position{line: 314, col: 6, offset: 5502},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 314, col: 12, offset: 5508},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 314, col: 15, offset: 5511},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 317, col: 5, offset: 5573},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 319, col: 1, offset: 5609},
			expr: &choiceExpr{
				pos: position{line: 320, col: 5, offset: 5648},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 320, col: 5, offset: 5648},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 320, col: 5, offset: 5648},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 323, col: 5, offset: 5686},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 323, col: 5, offset: 5686},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 323, col: 10, offset: 5691},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 325, col: 1, offset: 5723},
			expr: &actionExpr{
				pos: position{line: 326, col: 5, offset: 5758},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 326, col: 5, offset: 5758},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 326, col: 5, offset: 5758},
							expr: &ruleRefExpr{
								pos:  position{line: 326, col: 6, offset: 5759},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 326, col: 21, offset: 5774},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 330, col: 1, offset: 5825},
			expr: &anyMatcher{
				line: 331, col: 5, offset: 5840,
			},
		},
		{
			name: "__",
			pos:  position{line: 333, col: 1, offset: 5843},
			expr: &zeroOrMoreExpr{
				pos: position{line: 334, col: 5, offset: 5850},
				expr: &choiceExpr{
					pos: position{line: 334, col: 7, offset: 5852},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 334, col: 7, offset: 5852},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 334, col: 12, offset: 5857},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 336, col: 1, offset: 5865},
			expr: &charClassMatcher{
				pos:        position{line: 337, col: 5, offset: 5872},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 339, col: 1, offset: 5883},
			expr: &charClassMatcher{
				pos:        position{line: 340, col: 5, offset: 5902},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 342, col: 1, offset: 5910},
			expr: &litMatcher{
				pos:        position{line: 343, col: 5, offset: 5918},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 345, col: 1, offset: 5924},
			expr: &notExpr{
				pos: position{line: 346, col: 5, offset: 5932},
				expr: &anyMatcher{
					line: 346, col: 6, offset: 5933,
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
