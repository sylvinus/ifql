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
			name: "CallExpression",
			pos:  position{line: 51, col: 1, offset: 902},
			expr: &actionExpr{
				pos: position{line: 52, col: 5, offset: 921},
				run: (*parser).callonCallExpression1,
				expr: &seqExpr{
					pos: position{line: 52, col: 5, offset: 921},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 52, col: 5, offset: 921},
							label: "callee",
							expr: &ruleRefExpr{
								pos:  position{line: 52, col: 12, offset: 928},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 52, col: 19, offset: 935},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 52, col: 22, offset: 938},
							label: "args",
							expr: &ruleRefExpr{
								pos:  position{line: 52, col: 27, offset: 943},
								name: "Arguments",
							},
						},
						&labeledExpr{
							pos:   position{line: 52, col: 37, offset: 953},
							label: "members",
							expr: &zeroOrMoreExpr{
								pos: position{line: 52, col: 45, offset: 961},
								expr: &seqExpr{
									pos: position{line: 52, col: 47, offset: 963},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 52, col: 47, offset: 963},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 52, col: 50, offset: 966},
											val:        ".",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 54, offset: 970},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 57, offset: 973},
											name: "String",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 64, offset: 980},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 52, col: 67, offset: 983},
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
			pos:  position{line: 56, col: 1, offset: 1065},
			expr: &actionExpr{
				pos: position{line: 57, col: 5, offset: 1079},
				run: (*parser).callonArguments1,
				expr: &seqExpr{
					pos: position{line: 57, col: 5, offset: 1079},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 57, col: 5, offset: 1079},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 57, col: 9, offset: 1083},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 57, col: 12, offset: 1086},
							label: "args",
							expr: &zeroOrOneExpr{
								pos: position{line: 57, col: 17, offset: 1091},
								expr: &ruleRefExpr{
									pos:  position{line: 57, col: 18, offset: 1092},
									name: "FunctionArgs",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 57, col: 33, offset: 1107},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 57, col: 36, offset: 1110},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FunctionArgs",
			pos:  position{line: 61, col: 1, offset: 1146},
			expr: &actionExpr{
				pos: position{line: 62, col: 5, offset: 1163},
				run: (*parser).callonFunctionArgs1,
				expr: &seqExpr{
					pos: position{line: 62, col: 5, offset: 1163},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 62, col: 5, offset: 1163},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 62, col: 11, offset: 1169},
								name: "FunctionArg",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 62, col: 23, offset: 1181},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 62, col: 26, offset: 1184},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 62, col: 31, offset: 1189},
								expr: &ruleRefExpr{
									pos:  position{line: 62, col: 31, offset: 1189},
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
			pos:  position{line: 66, col: 1, offset: 1264},
			expr: &actionExpr{
				pos: position{line: 67, col: 5, offset: 1285},
				run: (*parser).callonFunctionArgsRest1,
				expr: &seqExpr{
					pos: position{line: 67, col: 5, offset: 1285},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 67, col: 5, offset: 1285},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 67, col: 9, offset: 1289},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 67, col: 13, offset: 1293},
							label: "arg",
							expr: &ruleRefExpr{
								pos:  position{line: 67, col: 17, offset: 1297},
								name: "FunctionArg",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArg",
			pos:  position{line: 71, col: 1, offset: 1340},
			expr: &actionExpr{
				pos: position{line: 72, col: 5, offset: 1356},
				run: (*parser).callonFunctionArg1,
				expr: &seqExpr{
					pos: position{line: 72, col: 5, offset: 1356},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 72, col: 5, offset: 1356},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 72, col: 9, offset: 1360},
								name: "String",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 72, col: 16, offset: 1367},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 72, col: 20, offset: 1371},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 72, col: 24, offset: 1375},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 72, col: 27, offset: 1378},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 72, col: 33, offset: 1384},
								name: "FunctionArgValues",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionArgValues",
			pos:  position{line: 76, col: 1, offset: 1460},
			expr: &choiceExpr{
				pos: position{line: 77, col: 5, offset: 1482},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 77, col: 5, offset: 1482},
						name: "WhereExpr",
					},
					&ruleRefExpr{
						pos:  position{line: 78, col: 5, offset: 1496},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 79, col: 5, offset: 1514},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 80, col: 5, offset: 1543},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 81, col: 5, offset: 1556},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 82, col: 5, offset: 1569},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 83, col: 5, offset: 1580},
						name: "String",
					},
				},
			},
		},
		{
			name: "WhereExpr",
			pos:  position{line: 85, col: 1, offset: 1588},
			expr: &actionExpr{
				pos: position{line: 86, col: 5, offset: 1602},
				run: (*parser).callonWhereExpr1,
				expr: &seqExpr{
					pos: position{line: 86, col: 5, offset: 1602},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 86, col: 5, offset: 1602},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 86, col: 9, offset: 1606},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 86, col: 12, offset: 1609},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 86, col: 17, offset: 1614},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 86, col: 22, offset: 1619},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 86, col: 26, offset: 1623},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 97, col: 1, offset: 1846},
			expr: &ruleRefExpr{
				pos:  position{line: 98, col: 5, offset: 1855},
				name: "Logical",
			},
		},
		{
			name: "LogicalOperators",
			pos:  position{line: 100, col: 1, offset: 1864},
			expr: &actionExpr{
				pos: position{line: 101, col: 5, offset: 1885},
				run: (*parser).callonLogicalOperators1,
				expr: &choiceExpr{
					pos: position{line: 101, col: 6, offset: 1886},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 101, col: 6, offset: 1886},
							val:        "or",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 101, col: 14, offset: 1894},
							val:        "and",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Logical",
			pos:  position{line: 105, col: 1, offset: 1946},
			expr: &actionExpr{
				pos: position{line: 106, col: 5, offset: 1958},
				run: (*parser).callonLogical1,
				expr: &seqExpr{
					pos: position{line: 106, col: 5, offset: 1958},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 106, col: 5, offset: 1958},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 106, col: 10, offset: 1963},
								name: "Equality",
							},
						},
						&labeledExpr{
							pos:   position{line: 106, col: 19, offset: 1972},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 106, col: 24, offset: 1977},
								expr: &seqExpr{
									pos: position{line: 106, col: 26, offset: 1979},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 106, col: 26, offset: 1979},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 30, offset: 1983},
											name: "LogicalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 47, offset: 2000},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 51, offset: 2004},
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
			pos:  position{line: 110, col: 1, offset: 2083},
			expr: &actionExpr{
				pos: position{line: 111, col: 5, offset: 2105},
				run: (*parser).callonEqualityOperators1,
				expr: &choiceExpr{
					pos: position{line: 111, col: 6, offset: 2106},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 111, col: 6, offset: 2106},
							val:        "==",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 111, col: 13, offset: 2113},
							val:        "!=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Equality",
			pos:  position{line: 115, col: 1, offset: 2159},
			expr: &actionExpr{
				pos: position{line: 116, col: 5, offset: 2172},
				run: (*parser).callonEquality1,
				expr: &seqExpr{
					pos: position{line: 116, col: 5, offset: 2172},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 116, col: 5, offset: 2172},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 116, col: 10, offset: 2177},
								name: "Relational",
							},
						},
						&labeledExpr{
							pos:   position{line: 116, col: 21, offset: 2188},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 116, col: 26, offset: 2193},
								expr: &seqExpr{
									pos: position{line: 116, col: 28, offset: 2195},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 116, col: 28, offset: 2195},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 116, col: 31, offset: 2198},
											name: "EqualityOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 116, col: 49, offset: 2216},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 116, col: 52, offset: 2219},
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
			pos:  position{line: 120, col: 1, offset: 2299},
			expr: &actionExpr{
				pos: position{line: 121, col: 5, offset: 2323},
				run: (*parser).callonRelationalOperators1,
				expr: &choiceExpr{
					pos: position{line: 121, col: 9, offset: 2327},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 121, col: 9, offset: 2327},
							val:        "<=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 122, col: 9, offset: 2340},
							val:        "<",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 123, col: 9, offset: 2352},
							val:        ">=",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 124, col: 9, offset: 2365},
							val:        ">",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 125, col: 9, offset: 2377},
							val:        "startswith",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 126, col: 9, offset: 2399},
							val:        "in",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 127, col: 9, offset: 2413},
							val:        "not empty",
							ignoreCase: true,
						},
						&litMatcher{
							pos:        position{line: 128, col: 9, offset: 2434},
							val:        "empty",
							ignoreCase: true,
						},
					},
				},
			},
		},
		{
			name: "Relational",
			pos:  position{line: 133, col: 1, offset: 2492},
			expr: &actionExpr{
				pos: position{line: 134, col: 5, offset: 2507},
				run: (*parser).callonRelational1,
				expr: &seqExpr{
					pos: position{line: 134, col: 5, offset: 2507},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 134, col: 5, offset: 2507},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 134, col: 10, offset: 2512},
								name: "Additive",
							},
						},
						&labeledExpr{
							pos:   position{line: 134, col: 19, offset: 2521},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 134, col: 24, offset: 2526},
								expr: &seqExpr{
									pos: position{line: 134, col: 26, offset: 2528},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 134, col: 26, offset: 2528},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 134, col: 29, offset: 2531},
											name: "RelationalOperators",
										},
										&ruleRefExpr{
											pos:  position{line: 134, col: 49, offset: 2551},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 134, col: 52, offset: 2554},
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
			pos:  position{line: 138, col: 1, offset: 2632},
			expr: &actionExpr{
				pos: position{line: 139, col: 5, offset: 2653},
				run: (*parser).callonAdditiveOperator1,
				expr: &choiceExpr{
					pos: position{line: 139, col: 6, offset: 2654},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 139, col: 6, offset: 2654},
							val:        "+",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 139, col: 12, offset: 2660},
							val:        "-",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Additive",
			pos:  position{line: 143, col: 1, offset: 2708},
			expr: &actionExpr{
				pos: position{line: 144, col: 5, offset: 2721},
				run: (*parser).callonAdditive1,
				expr: &seqExpr{
					pos: position{line: 144, col: 5, offset: 2721},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 144, col: 5, offset: 2721},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 144, col: 10, offset: 2726},
								name: "Multiplicative",
							},
						},
						&labeledExpr{
							pos:   position{line: 144, col: 25, offset: 2741},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 144, col: 30, offset: 2746},
								expr: &seqExpr{
									pos: position{line: 144, col: 32, offset: 2748},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 144, col: 32, offset: 2748},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 144, col: 35, offset: 2751},
											name: "AdditiveOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 144, col: 52, offset: 2768},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 144, col: 55, offset: 2771},
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
			pos:  position{line: 148, col: 1, offset: 2856},
			expr: &actionExpr{
				pos: position{line: 149, col: 5, offset: 2883},
				run: (*parser).callonMultiplicativeOperator1,
				expr: &choiceExpr{
					pos: position{line: 149, col: 6, offset: 2884},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 149, col: 6, offset: 2884},
							val:        "*",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 149, col: 12, offset: 2890},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Multiplicative",
			pos:  position{line: 153, col: 1, offset: 2934},
			expr: &actionExpr{
				pos: position{line: 154, col: 5, offset: 2953},
				run: (*parser).callonMultiplicative1,
				expr: &seqExpr{
					pos: position{line: 154, col: 5, offset: 2953},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 154, col: 5, offset: 2953},
							label: "head",
							expr: &ruleRefExpr{
								pos:  position{line: 154, col: 10, offset: 2958},
								name: "Primary",
							},
						},
						&labeledExpr{
							pos:   position{line: 154, col: 18, offset: 2966},
							label: "tail",
							expr: &zeroOrMoreExpr{
								pos: position{line: 154, col: 23, offset: 2971},
								expr: &seqExpr{
									pos: position{line: 154, col: 25, offset: 2973},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 154, col: 25, offset: 2973},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 154, col: 28, offset: 2976},
											name: "MultiplicativeOperator",
										},
										&ruleRefExpr{
											pos:  position{line: 154, col: 51, offset: 2999},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 154, col: 54, offset: 3002},
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
			pos:  position{line: 158, col: 1, offset: 3079},
			expr: &choiceExpr{
				pos: position{line: 159, col: 5, offset: 3091},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 159, col: 5, offset: 3091},
						run: (*parser).callonPrimary2,
						expr: &seqExpr{
							pos: position{line: 159, col: 5, offset: 3091},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 159, col: 5, offset: 3091},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 159, col: 9, offset: 3095},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 159, col: 12, offset: 3098},
									label: "expr",
									expr: &ruleRefExpr{
										pos:  position{line: 159, col: 17, offset: 3103},
										name: "Logical",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 159, col: 25, offset: 3111},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 159, col: 28, offset: 3114},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 162, col: 5, offset: 3153},
						name: "StringLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 163, col: 5, offset: 3171},
						name: "RegularExpressionLiteral",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 5, offset: 3200},
						name: "Duration",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 5, offset: 3213},
						name: "DateTime",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 5, offset: 3226},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 5, offset: 3237},
						name: "Field",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 5, offset: 3247},
						name: "String",
					},
				},
			},
		},
		{
			name: "DateFullYear",
			pos:  position{line: 170, col: 1, offset: 3255},
			expr: &seqExpr{
				pos: position{line: 171, col: 5, offset: 3272},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 171, col: 5, offset: 3272},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 171, col: 11, offset: 3278},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 171, col: 17, offset: 3284},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 171, col: 23, offset: 3290},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMonth",
			pos:  position{line: 173, col: 1, offset: 3297},
			expr: &seqExpr{
				pos: position{line: 175, col: 5, offset: 3322},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 175, col: 5, offset: 3322},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 175, col: 11, offset: 3328},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "DateMDay",
			pos:  position{line: 177, col: 1, offset: 3335},
			expr: &seqExpr{
				pos: position{line: 180, col: 5, offset: 3405},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 180, col: 5, offset: 3405},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 180, col: 11, offset: 3411},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeHour",
			pos:  position{line: 182, col: 1, offset: 3418},
			expr: &seqExpr{
				pos: position{line: 184, col: 5, offset: 3442},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 184, col: 5, offset: 3442},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 184, col: 11, offset: 3448},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeMinute",
			pos:  position{line: 186, col: 1, offset: 3455},
			expr: &seqExpr{
				pos: position{line: 188, col: 5, offset: 3481},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 188, col: 5, offset: 3481},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 188, col: 11, offset: 3487},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecond",
			pos:  position{line: 190, col: 1, offset: 3494},
			expr: &seqExpr{
				pos: position{line: 193, col: 5, offset: 3566},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 193, col: 5, offset: 3566},
						name: "Digit",
					},
					&ruleRefExpr{
						pos:  position{line: 193, col: 11, offset: 3572},
						name: "Digit",
					},
				},
			},
		},
		{
			name: "TimeSecFrac",
			pos:  position{line: 195, col: 1, offset: 3579},
			expr: &seqExpr{
				pos: position{line: 196, col: 5, offset: 3595},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 196, col: 5, offset: 3595},
						val:        ".",
						ignoreCase: false,
					},
					&oneOrMoreExpr{
						pos: position{line: 196, col: 9, offset: 3599},
						expr: &ruleRefExpr{
							pos:  position{line: 196, col: 9, offset: 3599},
							name: "Digit",
						},
					},
				},
			},
		},
		{
			name: "TimeNumOffset",
			pos:  position{line: 198, col: 1, offset: 3607},
			expr: &seqExpr{
				pos: position{line: 199, col: 5, offset: 3625},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 199, col: 6, offset: 3626},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 199, col: 6, offset: 3626},
								val:        "+",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 199, col: 12, offset: 3632},
								val:        "-",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 199, col: 17, offset: 3637},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 199, col: 26, offset: 3646},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 199, col: 30, offset: 3650},
						name: "TimeMinute",
					},
				},
			},
		},
		{
			name: "TimeOffset",
			pos:  position{line: 201, col: 1, offset: 3662},
			expr: &choiceExpr{
				pos: position{line: 202, col: 6, offset: 3678},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 202, col: 6, offset: 3678},
						val:        "Z",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 202, col: 12, offset: 3684},
						name: "TimeNumOffset",
					},
				},
			},
		},
		{
			name: "PartialTime",
			pos:  position{line: 204, col: 1, offset: 3700},
			expr: &seqExpr{
				pos: position{line: 205, col: 5, offset: 3716},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 205, col: 5, offset: 3716},
						name: "TimeHour",
					},
					&litMatcher{
						pos:        position{line: 205, col: 14, offset: 3725},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 18, offset: 3729},
						name: "TimeMinute",
					},
					&litMatcher{
						pos:        position{line: 205, col: 29, offset: 3740},
						val:        ":",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 205, col: 33, offset: 3744},
						name: "TimeSecond",
					},
					&zeroOrOneExpr{
						pos: position{line: 205, col: 44, offset: 3755},
						expr: &ruleRefExpr{
							pos:  position{line: 205, col: 44, offset: 3755},
							name: "TimeSecFrac",
						},
					},
				},
			},
		},
		{
			name: "FullDate",
			pos:  position{line: 207, col: 1, offset: 3769},
			expr: &seqExpr{
				pos: position{line: 208, col: 5, offset: 3782},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 208, col: 5, offset: 3782},
						name: "DateFullYear",
					},
					&litMatcher{
						pos:        position{line: 208, col: 18, offset: 3795},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 208, col: 22, offset: 3799},
						name: "DateMonth",
					},
					&litMatcher{
						pos:        position{line: 208, col: 32, offset: 3809},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 208, col: 36, offset: 3813},
						name: "DateMDay",
					},
				},
			},
		},
		{
			name: "FullTime",
			pos:  position{line: 210, col: 1, offset: 3823},
			expr: &seqExpr{
				pos: position{line: 211, col: 5, offset: 3836},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 211, col: 5, offset: 3836},
						name: "PartialTime",
					},
					&ruleRefExpr{
						pos:  position{line: 211, col: 17, offset: 3848},
						name: "TimeOffset",
					},
				},
			},
		},
		{
			name: "DateTime",
			pos:  position{line: 213, col: 1, offset: 3860},
			expr: &actionExpr{
				pos: position{line: 214, col: 5, offset: 3873},
				run: (*parser).callonDateTime1,
				expr: &seqExpr{
					pos: position{line: 214, col: 5, offset: 3873},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 214, col: 5, offset: 3873},
							name: "FullDate",
						},
						&litMatcher{
							pos:        position{line: 214, col: 14, offset: 3882},
							val:        "T",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 214, col: 18, offset: 3886},
							name: "FullTime",
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 218, col: 1, offset: 3941},
			expr: &litMatcher{
				pos:        position{line: 219, col: 5, offset: 3961},
				val:        "ns",
				ignoreCase: false,
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 221, col: 1, offset: 3967},
			expr: &choiceExpr{
				pos: position{line: 222, col: 6, offset: 3989},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 222, col: 6, offset: 3989},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 222, col: 13, offset: 3996},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 222, col: 20, offset: 4004},
						val:        "μs",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 224, col: 1, offset: 4012},
			expr: &litMatcher{
				pos:        position{line: 225, col: 5, offset: 4033},
				val:        "ms",
				ignoreCase: false,
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 227, col: 1, offset: 4039},
			expr: &litMatcher{
				pos:        position{line: 228, col: 5, offset: 4055},
				val:        "s",
				ignoreCase: false,
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 230, col: 1, offset: 4060},
			expr: &litMatcher{
				pos:        position{line: 231, col: 5, offset: 4076},
				val:        "m",
				ignoreCase: false,
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 233, col: 1, offset: 4081},
			expr: &litMatcher{
				pos:        position{line: 234, col: 5, offset: 4095},
				val:        "h",
				ignoreCase: false,
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 236, col: 1, offset: 4100},
			expr: &choiceExpr{
				pos: position{line: 238, col: 9, offset: 4128},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 238, col: 9, offset: 4128},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 239, col: 9, offset: 4152},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 240, col: 9, offset: 4177},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 241, col: 9, offset: 4202},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 242, col: 9, offset: 4222},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 243, col: 9, offset: 4242},
						name: "HourUnits",
					},
				},
			},
		},
		{
			name: "SingleDuration",
			pos:  position{line: 246, col: 1, offset: 4259},
			expr: &seqExpr{
				pos: position{line: 247, col: 5, offset: 4278},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 247, col: 5, offset: 4278},
						name: "Number",
					},
					&ruleRefExpr{
						pos:  position{line: 247, col: 12, offset: 4285},
						name: "DurationUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 249, col: 1, offset: 4300},
			expr: &actionExpr{
				pos: position{line: 250, col: 5, offset: 4313},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 250, col: 5, offset: 4313},
					expr: &ruleRefExpr{
						pos:  position{line: 250, col: 5, offset: 4313},
						name: "SingleDuration",
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 254, col: 1, offset: 4382},
			expr: &choiceExpr{
				pos: position{line: 255, col: 5, offset: 4400},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 255, col: 5, offset: 4400},
						run: (*parser).callonStringLiteral2,
						expr: &seqExpr{
							pos: position{line: 255, col: 7, offset: 4402},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 255, col: 7, offset: 4402},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 255, col: 11, offset: 4406},
									expr: &ruleRefExpr{
										pos:  position{line: 255, col: 11, offset: 4406},
										name: "DoubleStringChar",
									},
								},
								&litMatcher{
									pos:        position{line: 255, col: 29, offset: 4424},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 258, col: 5, offset: 4488},
						run: (*parser).callonStringLiteral8,
						expr: &seqExpr{
							pos: position{line: 258, col: 7, offset: 4490},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 258, col: 7, offset: 4490},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 258, col: 11, offset: 4494},
									expr: &ruleRefExpr{
										pos:  position{line: 258, col: 11, offset: 4494},
										name: "DoubleStringChar",
									},
								},
								&choiceExpr{
									pos: position{line: 258, col: 31, offset: 4514},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 258, col: 31, offset: 4514},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 258, col: 37, offset: 4520},
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
			pos:  position{line: 262, col: 1, offset: 4602},
			expr: &choiceExpr{
				pos: position{line: 263, col: 5, offset: 4623},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 263, col: 5, offset: 4623},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 263, col: 5, offset: 4623},
								expr: &choiceExpr{
									pos: position{line: 263, col: 8, offset: 4626},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 263, col: 8, offset: 4626},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 263, col: 14, offset: 4632},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 263, col: 21, offset: 4639},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 263, col: 27, offset: 4645},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 264, col: 5, offset: 4660},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 264, col: 5, offset: 4660},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 264, col: 10, offset: 4665},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 266, col: 1, offset: 4685},
			expr: &choiceExpr{
				pos: position{line: 267, col: 5, offset: 4708},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 267, col: 5, offset: 4708},
						val:        "\"",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 268, col: 5, offset: 4716},
						run: (*parser).callonDoubleStringEscape3,
						expr: &choiceExpr{
							pos: position{line: 268, col: 7, offset: 4718},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 268, col: 7, offset: 4718},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 268, col: 20, offset: 4731},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 268, col: 26, offset: 4737},
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
			pos:  position{line: 272, col: 1, offset: 4809},
			expr: &actionExpr{
				pos: position{line: 273, col: 5, offset: 4820},
				run: (*parser).callonString1,
				expr: &oneOrMoreExpr{
					pos: position{line: 273, col: 5, offset: 4820},
					expr: &ruleRefExpr{
						pos:  position{line: 273, col: 5, offset: 4820},
						name: "StringChar",
					},
				},
			},
		},
		{
			name: "StringChar",
			pos:  position{line: 277, col: 1, offset: 4880},
			expr: &seqExpr{
				pos: position{line: 278, col: 5, offset: 4895},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 278, col: 5, offset: 4895},
						expr: &choiceExpr{
							pos: position{line: 278, col: 7, offset: 4897},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 278, col: 7, offset: 4897},
									val:        "\"",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 13, offset: 4903},
									val:        "(",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 19, offset: 4909},
									val:        ")",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 25, offset: 4915},
									val:        ":",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 31, offset: 4921},
									val:        "{",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 37, offset: 4927},
									val:        "}",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 43, offset: 4933},
									val:        ",",
									ignoreCase: false,
								},
								&litMatcher{
									pos:        position{line: 278, col: 49, offset: 4939},
									val:        "$",
									ignoreCase: false,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 278, col: 54, offset: 4944},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 280, col: 1, offset: 4956},
			expr: &actionExpr{
				pos: position{line: 281, col: 5, offset: 4967},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 281, col: 5, offset: 4967},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 281, col: 5, offset: 4967},
							expr: &litMatcher{
								pos:        position{line: 281, col: 5, offset: 4967},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 281, col: 10, offset: 4972},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 281, col: 18, offset: 4980},
							expr: &seqExpr{
								pos: position{line: 281, col: 20, offset: 4982},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 281, col: 20, offset: 4982},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 281, col: 24, offset: 4986},
										expr: &ruleRefExpr{
											pos:  position{line: 281, col: 24, offset: 4986},
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
			pos:  position{line: 285, col: 1, offset: 5047},
			expr: &choiceExpr{
				pos: position{line: 286, col: 5, offset: 5059},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 286, col: 5, offset: 5059},
						val:        "0",
						ignoreCase: false,
					},
					&seqExpr{
						pos: position{line: 287, col: 5, offset: 5067},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 287, col: 5, offset: 5067},
								name: "NonZeroDigit",
							},
							&zeroOrMoreExpr{
								pos: position{line: 287, col: 18, offset: 5080},
								expr: &ruleRefExpr{
									pos:  position{line: 287, col: 18, offset: 5080},
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
			pos:  position{line: 289, col: 1, offset: 5088},
			expr: &charClassMatcher{
				pos:        position{line: 290, col: 5, offset: 5105},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 292, col: 1, offset: 5112},
			expr: &charClassMatcher{
				pos:        position{line: 293, col: 5, offset: 5122},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Field",
			pos:  position{line: 295, col: 1, offset: 5129},
			expr: &actionExpr{
				pos: position{line: 296, col: 5, offset: 5139},
				run: (*parser).callonField1,
				expr: &labeledExpr{
					pos:   position{line: 296, col: 5, offset: 5139},
					label: "field",
					expr: &litMatcher{
						pos:        position{line: 296, col: 11, offset: 5145},
						val:        "$",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "RegularExpressionLiteral",
			displayName: "\"regular expression\"",
			pos:         position{line: 300, col: 1, offset: 5199},
			expr: &actionExpr{
				pos: position{line: 301, col: 5, offset: 5249},
				run: (*parser).callonRegularExpressionLiteral1,
				expr: &seqExpr{
					pos: position{line: 301, col: 5, offset: 5249},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 301, col: 5, offset: 5249},
							val:        "/",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 301, col: 9, offset: 5253},
							label: "pattern",
							expr: &ruleRefExpr{
								pos:  position{line: 301, col: 17, offset: 5261},
								name: "RegularExpressionBody",
							},
						},
						&litMatcher{
							pos:        position{line: 301, col: 39, offset: 5283},
							val:        "/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionBody",
			pos:  position{line: 305, col: 1, offset: 5325},
			expr: &actionExpr{
				pos: position{line: 306, col: 5, offset: 5351},
				run: (*parser).callonRegularExpressionBody1,
				expr: &labeledExpr{
					pos:   position{line: 306, col: 5, offset: 5351},
					label: "chars",
					expr: &oneOrMoreExpr{
						pos: position{line: 306, col: 11, offset: 5357},
						expr: &ruleRefExpr{
							pos:  position{line: 306, col: 11, offset: 5357},
							name: "RegularExpressionChar",
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionChar",
			pos:  position{line: 310, col: 1, offset: 5435},
			expr: &choiceExpr{
				pos: position{line: 311, col: 5, offset: 5461},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 311, col: 5, offset: 5461},
						run: (*parser).callonRegularExpressionChar2,
						expr: &seqExpr{
							pos: position{line: 311, col: 5, offset: 5461},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 311, col: 5, offset: 5461},
									expr: &charClassMatcher{
										pos:        position{line: 311, col: 6, offset: 5462},
										val:        "[\\\\/]",
										chars:      []rune{'\\', '/'},
										ignoreCase: false,
										inverted:   false,
									},
								},
								&labeledExpr{
									pos:   position{line: 311, col: 12, offset: 5468},
									label: "re",
									expr: &ruleRefExpr{
										pos:  position{line: 311, col: 15, offset: 5471},
										name: "RegularExpressionNonTerminator",
									},
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 314, col: 5, offset: 5533},
						name: "RegularExpressionBackslashSequence",
					},
				},
			},
		},
		{
			name: "RegularExpressionBackslashSequence",
			pos:  position{line: 316, col: 1, offset: 5569},
			expr: &choiceExpr{
				pos: position{line: 317, col: 5, offset: 5608},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 317, col: 5, offset: 5608},
						run: (*parser).callonRegularExpressionBackslashSequence2,
						expr: &litMatcher{
							pos:        position{line: 317, col: 5, offset: 5608},
							val:        "\\/",
							ignoreCase: false,
						},
					},
					&seqExpr{
						pos: position{line: 320, col: 5, offset: 5646},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 320, col: 5, offset: 5646},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 320, col: 10, offset: 5651},
								name: "RegularExpressionNonTerminator",
							},
						},
					},
				},
			},
		},
		{
			name: "RegularExpressionNonTerminator",
			pos:  position{line: 322, col: 1, offset: 5683},
			expr: &actionExpr{
				pos: position{line: 323, col: 5, offset: 5718},
				run: (*parser).callonRegularExpressionNonTerminator1,
				expr: &seqExpr{
					pos: position{line: 323, col: 5, offset: 5718},
					exprs: []interface{}{
						&notExpr{
							pos: position{line: 323, col: 5, offset: 5718},
							expr: &ruleRefExpr{
								pos:  position{line: 323, col: 6, offset: 5719},
								name: "LineTerminator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 21, offset: 5734},
							name: "SourceChar",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 327, col: 1, offset: 5785},
			expr: &anyMatcher{
				line: 328, col: 5, offset: 5800,
			},
		},
		{
			name: "__",
			pos:  position{line: 330, col: 1, offset: 5803},
			expr: &zeroOrMoreExpr{
				pos: position{line: 331, col: 5, offset: 5810},
				expr: &choiceExpr{
					pos: position{line: 331, col: 7, offset: 5812},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 331, col: 7, offset: 5812},
							name: "ws",
						},
						&ruleRefExpr{
							pos:  position{line: 331, col: 12, offset: 5817},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "ws",
			pos:  position{line: 333, col: 1, offset: 5825},
			expr: &charClassMatcher{
				pos:        position{line: 334, col: 5, offset: 5832},
				val:        "[ \\t\\r\\n]",
				chars:      []rune{' ', '\t', '\r', '\n'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LineTerminator",
			pos:  position{line: 336, col: 1, offset: 5843},
			expr: &charClassMatcher{
				pos:        position{line: 337, col: 5, offset: 5862},
				val:        "[\\n\\r]",
				chars:      []rune{'\n', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 339, col: 1, offset: 5870},
			expr: &litMatcher{
				pos:        position{line: 340, col: 5, offset: 5878},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOF",
			pos:  position{line: 342, col: 1, offset: 5884},
			expr: &notExpr{
				pos: position{line: 343, col: 5, offset: 5892},
				expr: &anyMatcher{
					line: 343, col: 6, offset: 5893,
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
