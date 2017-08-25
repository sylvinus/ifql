package promql

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

//go:generate pigeon -o promql.go promql.peg

var reservedWords = map[string]bool{}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 10, col: 1, offset: 106},
			expr: &actionExpr{
				pos: position{line: 10, col: 12, offset: 117},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 10, col: 12, offset: 117},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 10, col: 12, offset: 117},
							label: "grammar",
							expr: &choiceExpr{
								pos: position{line: 10, col: 22, offset: 127},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 10, col: 22, offset: 127},
										name: "Comment",
									},
									&ruleRefExpr{
										pos:  position{line: 10, col: 32, offset: 137},
										name: "AggregateExpression",
									},
									&ruleRefExpr{
										pos:  position{line: 10, col: 54, offset: 159},
										name: "VectorSelector",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 10, col: 71, offset: 176},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 14, col: 1, offset: 209},
			expr: &anyMatcher{
				line: 14, col: 14, offset: 222,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 16, col: 1, offset: 225},
			expr: &actionExpr{
				pos: position{line: 16, col: 11, offset: 235},
				run: (*parser).callonComment1,
				expr: &seqExpr{
					pos: position{line: 16, col: 11, offset: 235},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 16, col: 11, offset: 235},
							val:        "#",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 16, col: 15, offset: 239},
							expr: &seqExpr{
								pos: position{line: 16, col: 17, offset: 241},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 16, col: 17, offset: 241},
										expr: &ruleRefExpr{
											pos:  position{line: 16, col: 18, offset: 242},
											name: "EOL",
										},
									},
									&ruleRefExpr{
										pos:  position{line: 16, col: 22, offset: 246},
										name: "SourceChar",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Identifier",
			pos:  position{line: 20, col: 1, offset: 306},
			expr: &actionExpr{
				pos: position{line: 20, col: 14, offset: 319},
				run: (*parser).callonIdentifier1,
				expr: &labeledExpr{
					pos:   position{line: 20, col: 14, offset: 319},
					label: "ident",
					expr: &ruleRefExpr{
						pos:  position{line: 20, col: 20, offset: 325},
						name: "IdentifierName",
					},
				},
			},
		},
		{
			name: "IdentifierName",
			pos:  position{line: 28, col: 1, offset: 509},
			expr: &actionExpr{
				pos: position{line: 28, col: 18, offset: 526},
				run: (*parser).callonIdentifierName1,
				expr: &seqExpr{
					pos: position{line: 28, col: 18, offset: 526},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 28, col: 18, offset: 526},
							name: "IdentifierStart",
						},
						&zeroOrMoreExpr{
							pos: position{line: 28, col: 34, offset: 542},
							expr: &ruleRefExpr{
								pos:  position{line: 28, col: 34, offset: 542},
								name: "IdentifierPart",
							},
						},
					},
				},
			},
		},
		{
			name: "IdentifierStart",
			pos:  position{line: 31, col: 1, offset: 593},
			expr: &charClassMatcher{
				pos:        position{line: 31, col: 19, offset: 611},
				val:        "[\\pL_]",
				chars:      []rune{'_'},
				classes:    []*unicode.RangeTable{rangeTable("L")},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "IdentifierPart",
			pos:  position{line: 32, col: 1, offset: 618},
			expr: &choiceExpr{
				pos: position{line: 32, col: 18, offset: 635},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 32, col: 18, offset: 635},
						name: "IdentifierStart",
					},
					&charClassMatcher{
						pos:        position{line: 32, col: 36, offset: 653},
						val:        "[\\p{Nd}]",
						classes:    []*unicode.RangeTable{rangeTable("Nd")},
						ignoreCase: false,
						inverted:   false,
					},
				},
			},
		},
		{
			name: "StringLiteral",
			pos:  position{line: 34, col: 1, offset: 663},
			expr: &choiceExpr{
				pos: position{line: 34, col: 17, offset: 679},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 34, col: 17, offset: 679},
						run: (*parser).callonStringLiteral2,
						expr: &choiceExpr{
							pos: position{line: 34, col: 19, offset: 681},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 34, col: 19, offset: 681},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 34, col: 19, offset: 681},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 34, col: 23, offset: 685},
											expr: &ruleRefExpr{
												pos:  position{line: 34, col: 23, offset: 685},
												name: "DoubleStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 34, col: 41, offset: 703},
											val:        "\"",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 34, col: 47, offset: 709},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 34, col: 47, offset: 709},
											val:        "'",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 34, col: 51, offset: 713},
											name: "SingleStringChar",
										},
										&litMatcher{
											pos:        position{line: 34, col: 68, offset: 730},
											val:        "'",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 34, col: 74, offset: 736},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 34, col: 74, offset: 736},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 34, col: 78, offset: 740},
											expr: &ruleRefExpr{
												pos:  position{line: 34, col: 78, offset: 740},
												name: "RawStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 34, col: 93, offset: 755},
											val:        "`",
											ignoreCase: false,
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 40, col: 5, offset: 901},
						run: (*parser).callonStringLiteral18,
						expr: &choiceExpr{
							pos: position{line: 40, col: 7, offset: 903},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 40, col: 9, offset: 905},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 40, col: 9, offset: 905},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 40, col: 13, offset: 909},
											expr: &ruleRefExpr{
												pos:  position{line: 40, col: 13, offset: 909},
												name: "DoubleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 40, col: 33, offset: 929},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 40, col: 33, offset: 929},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 40, col: 39, offset: 935},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 40, col: 51, offset: 947},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 40, col: 51, offset: 947},
											val:        "'",
											ignoreCase: false,
										},
										&zeroOrOneExpr{
											pos: position{line: 40, col: 55, offset: 951},
											expr: &ruleRefExpr{
												pos:  position{line: 40, col: 55, offset: 951},
												name: "SingleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 40, col: 75, offset: 971},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 40, col: 75, offset: 971},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 40, col: 81, offset: 977},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 40, col: 91, offset: 987},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 40, col: 91, offset: 987},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 40, col: 95, offset: 991},
											expr: &ruleRefExpr{
												pos:  position{line: 40, col: 95, offset: 991},
												name: "RawStringChar",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 40, col: 110, offset: 1006},
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
			pos:  position{line: 44, col: 1, offset: 1077},
			expr: &choiceExpr{
				pos: position{line: 44, col: 20, offset: 1096},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 44, col: 20, offset: 1096},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 44, col: 20, offset: 1096},
								expr: &choiceExpr{
									pos: position{line: 44, col: 23, offset: 1099},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 44, col: 23, offset: 1099},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 44, col: 29, offset: 1105},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 44, col: 36, offset: 1112},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 44, col: 42, offset: 1118},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 44, col: 55, offset: 1131},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 44, col: 55, offset: 1131},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 44, col: 60, offset: 1136},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "SingleStringChar",
			pos:  position{line: 45, col: 1, offset: 1155},
			expr: &choiceExpr{
				pos: position{line: 45, col: 20, offset: 1174},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 45, col: 20, offset: 1174},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 45, col: 20, offset: 1174},
								expr: &choiceExpr{
									pos: position{line: 45, col: 23, offset: 1177},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 45, col: 23, offset: 1177},
											val:        "'",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 45, col: 29, offset: 1183},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 45, col: 36, offset: 1190},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 45, col: 42, offset: 1196},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 45, col: 55, offset: 1209},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 45, col: 55, offset: 1209},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 45, col: 60, offset: 1214},
								name: "SingleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "RawStringChar",
			pos:  position{line: 46, col: 1, offset: 1233},
			expr: &seqExpr{
				pos: position{line: 46, col: 17, offset: 1249},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 46, col: 17, offset: 1249},
						expr: &litMatcher{
							pos:        position{line: 46, col: 18, offset: 1250},
							val:        "`",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 46, col: 22, offset: 1254},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 48, col: 1, offset: 1266},
			expr: &choiceExpr{
				pos: position{line: 48, col: 22, offset: 1287},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 48, col: 24, offset: 1289},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 48, col: 24, offset: 1289},
								val:        "\"",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 48, col: 30, offset: 1295},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 49, col: 7, offset: 1324},
						run: (*parser).callonDoubleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 49, col: 9, offset: 1326},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 49, col: 9, offset: 1326},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 49, col: 22, offset: 1339},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 49, col: 28, offset: 1345},
									name: "EOF",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SingleStringEscape",
			pos:  position{line: 52, col: 1, offset: 1410},
			expr: &choiceExpr{
				pos: position{line: 52, col: 22, offset: 1431},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 52, col: 24, offset: 1433},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 52, col: 24, offset: 1433},
								val:        "'",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 52, col: 30, offset: 1439},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 53, col: 7, offset: 1468},
						run: (*parser).callonSingleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 53, col: 9, offset: 1470},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 53, col: 9, offset: 1470},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 53, col: 22, offset: 1483},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 53, col: 28, offset: 1489},
									name: "EOF",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CommonEscapeSequence",
			pos:  position{line: 57, col: 1, offset: 1555},
			expr: &choiceExpr{
				pos: position{line: 57, col: 24, offset: 1578},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 57, col: 24, offset: 1578},
						name: "SingleCharEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 43, offset: 1597},
						name: "OctalEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 57, offset: 1611},
						name: "HexEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 69, offset: 1623},
						name: "LongUnicodeEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 57, col: 89, offset: 1643},
						name: "ShortUnicodeEscape",
					},
				},
			},
		},
		{
			name: "SingleCharEscape",
			pos:  position{line: 58, col: 1, offset: 1662},
			expr: &choiceExpr{
				pos: position{line: 58, col: 20, offset: 1681},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 58, col: 20, offset: 1681},
						val:        "a",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 26, offset: 1687},
						val:        "b",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 32, offset: 1693},
						val:        "n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 38, offset: 1699},
						val:        "f",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 44, offset: 1705},
						val:        "r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 50, offset: 1711},
						val:        "t",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 56, offset: 1717},
						val:        "v",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 58, col: 62, offset: 1723},
						val:        "\\",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "OctalEscape",
			pos:  position{line: 59, col: 1, offset: 1728},
			expr: &choiceExpr{
				pos: position{line: 59, col: 15, offset: 1742},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 59, col: 15, offset: 1742},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 59, col: 15, offset: 1742},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 59, col: 26, offset: 1753},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 59, col: 37, offset: 1764},
								name: "OctalDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 60, col: 7, offset: 1781},
						run: (*parser).callonOctalEscape6,
						expr: &seqExpr{
							pos: position{line: 60, col: 7, offset: 1781},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 60, col: 7, offset: 1781},
									name: "OctalDigit",
								},
								&choiceExpr{
									pos: position{line: 60, col: 20, offset: 1794},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 60, col: 20, offset: 1794},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 60, col: 33, offset: 1807},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 60, col: 39, offset: 1813},
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
			name: "HexEscape",
			pos:  position{line: 63, col: 1, offset: 1874},
			expr: &choiceExpr{
				pos: position{line: 63, col: 13, offset: 1886},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 63, col: 13, offset: 1886},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 63, col: 13, offset: 1886},
								val:        "x",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 63, col: 17, offset: 1890},
								name: "HexDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 63, col: 26, offset: 1899},
								name: "HexDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 64, col: 7, offset: 1914},
						run: (*parser).callonHexEscape6,
						expr: &seqExpr{
							pos: position{line: 64, col: 7, offset: 1914},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 64, col: 7, offset: 1914},
									val:        "x",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 64, col: 13, offset: 1920},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 64, col: 13, offset: 1920},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 64, col: 26, offset: 1933},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 64, col: 32, offset: 1939},
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
			name: "LongUnicodeEscape",
			pos:  position{line: 67, col: 1, offset: 2006},
			expr: &choiceExpr{
				pos: position{line: 68, col: 5, offset: 2031},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 68, col: 5, offset: 2031},
						run: (*parser).callonLongUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 68, col: 5, offset: 2031},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 68, col: 5, offset: 2031},
									val:        "U",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 9, offset: 2035},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 18, offset: 2044},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 27, offset: 2053},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 36, offset: 2062},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 45, offset: 2071},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 54, offset: 2080},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 63, offset: 2089},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 68, col: 72, offset: 2098},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 71, col: 7, offset: 2200},
						run: (*parser).callonLongUnicodeEscape13,
						expr: &seqExpr{
							pos: position{line: 71, col: 7, offset: 2200},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 71, col: 7, offset: 2200},
									val:        "U",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 71, col: 13, offset: 2206},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 71, col: 13, offset: 2206},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 71, col: 26, offset: 2219},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 71, col: 32, offset: 2225},
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
			name: "ShortUnicodeEscape",
			pos:  position{line: 74, col: 1, offset: 2288},
			expr: &choiceExpr{
				pos: position{line: 75, col: 5, offset: 2314},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 75, col: 5, offset: 2314},
						run: (*parser).callonShortUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 75, col: 5, offset: 2314},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 75, col: 5, offset: 2314},
									val:        "u",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 9, offset: 2318},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 18, offset: 2327},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 27, offset: 2336},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 36, offset: 2345},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 78, col: 7, offset: 2447},
						run: (*parser).callonShortUnicodeEscape9,
						expr: &seqExpr{
							pos: position{line: 78, col: 7, offset: 2447},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 78, col: 7, offset: 2447},
									val:        "u",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 78, col: 13, offset: 2453},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 78, col: 13, offset: 2453},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 78, col: 26, offset: 2466},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 78, col: 32, offset: 2472},
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
			name: "OctalDigit",
			pos:  position{line: 82, col: 1, offset: 2536},
			expr: &charClassMatcher{
				pos:        position{line: 82, col: 14, offset: 2549},
				val:        "[0-7]",
				ranges:     []rune{'0', '7'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 83, col: 1, offset: 2555},
			expr: &charClassMatcher{
				pos:        position{line: 83, col: 16, offset: 2570},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "HexDigit",
			pos:  position{line: 84, col: 1, offset: 2576},
			expr: &charClassMatcher{
				pos:        position{line: 84, col: 12, offset: 2587},
				val:        "[0-9a-f]i",
				ranges:     []rune{'0', '9', 'a', 'f'},
				ignoreCase: true,
				inverted:   false,
			},
		},
		{
			name: "CharClassMatcher",
			pos:  position{line: 86, col: 1, offset: 2598},
			expr: &choiceExpr{
				pos: position{line: 86, col: 20, offset: 2617},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 86, col: 20, offset: 2617},
						run: (*parser).callonCharClassMatcher2,
						expr: &seqExpr{
							pos: position{line: 86, col: 20, offset: 2617},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 86, col: 20, offset: 2617},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 86, col: 24, offset: 2621},
									expr: &choiceExpr{
										pos: position{line: 86, col: 26, offset: 2623},
										alternatives: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 86, col: 26, offset: 2623},
												name: "ClassCharRange",
											},
											&ruleRefExpr{
												pos:  position{line: 86, col: 43, offset: 2640},
												name: "ClassChar",
											},
											&seqExpr{
												pos: position{line: 86, col: 55, offset: 2652},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 86, col: 55, offset: 2652},
														val:        "\\",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 86, col: 60, offset: 2657},
														name: "UnicodeClassEscape",
													},
												},
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 86, col: 82, offset: 2679},
									val:        "]",
									ignoreCase: false,
								},
								&zeroOrOneExpr{
									pos: position{line: 86, col: 86, offset: 2683},
									expr: &litMatcher{
										pos:        position{line: 86, col: 86, offset: 2683},
										val:        "i",
										ignoreCase: false,
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 88, col: 5, offset: 2725},
						run: (*parser).callonCharClassMatcher15,
						expr: &seqExpr{
							pos: position{line: 88, col: 5, offset: 2725},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 88, col: 5, offset: 2725},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 88, col: 9, offset: 2729},
									expr: &seqExpr{
										pos: position{line: 88, col: 11, offset: 2731},
										exprs: []interface{}{
											&notExpr{
												pos: position{line: 88, col: 11, offset: 2731},
												expr: &ruleRefExpr{
													pos:  position{line: 88, col: 14, offset: 2734},
													name: "EOL",
												},
											},
											&ruleRefExpr{
												pos:  position{line: 88, col: 20, offset: 2740},
												name: "SourceChar",
											},
										},
									},
								},
								&choiceExpr{
									pos: position{line: 88, col: 36, offset: 2756},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 88, col: 36, offset: 2756},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 88, col: 42, offset: 2762},
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
			name: "ClassCharRange",
			pos:  position{line: 92, col: 1, offset: 2834},
			expr: &seqExpr{
				pos: position{line: 92, col: 18, offset: 2851},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 92, col: 18, offset: 2851},
						name: "ClassChar",
					},
					&litMatcher{
						pos:        position{line: 92, col: 28, offset: 2861},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 92, col: 32, offset: 2865},
						name: "ClassChar",
					},
				},
			},
		},
		{
			name: "ClassChar",
			pos:  position{line: 93, col: 1, offset: 2875},
			expr: &choiceExpr{
				pos: position{line: 93, col: 13, offset: 2887},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 93, col: 13, offset: 2887},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 93, col: 13, offset: 2887},
								expr: &choiceExpr{
									pos: position{line: 93, col: 16, offset: 2890},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 93, col: 16, offset: 2890},
											val:        "]",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 93, col: 22, offset: 2896},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 93, col: 29, offset: 2903},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 93, col: 35, offset: 2909},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 93, col: 48, offset: 2922},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 93, col: 48, offset: 2922},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 93, col: 53, offset: 2927},
								name: "CharClassEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "CharClassEscape",
			pos:  position{line: 94, col: 1, offset: 2943},
			expr: &choiceExpr{
				pos: position{line: 94, col: 19, offset: 2961},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 94, col: 21, offset: 2963},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 94, col: 21, offset: 2963},
								val:        "]",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 94, col: 27, offset: 2969},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 95, col: 7, offset: 2998},
						run: (*parser).callonCharClassEscape5,
						expr: &seqExpr{
							pos: position{line: 95, col: 7, offset: 2998},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 95, col: 7, offset: 2998},
									expr: &litMatcher{
										pos:        position{line: 95, col: 8, offset: 2999},
										val:        "p",
										ignoreCase: false,
									},
								},
								&choiceExpr{
									pos: position{line: 95, col: 14, offset: 3005},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 95, col: 14, offset: 3005},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 95, col: 27, offset: 3018},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 95, col: 33, offset: 3024},
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
			name: "UnicodeClassEscape",
			pos:  position{line: 99, col: 1, offset: 3090},
			expr: &seqExpr{
				pos: position{line: 99, col: 22, offset: 3111},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 99, col: 22, offset: 3111},
						val:        "p",
						ignoreCase: false,
					},
					&choiceExpr{
						pos: position{line: 100, col: 7, offset: 3124},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 100, col: 7, offset: 3124},
								name: "SingleCharUnicodeClass",
							},
							&actionExpr{
								pos: position{line: 101, col: 7, offset: 3153},
								run: (*parser).callonUnicodeClassEscape5,
								expr: &seqExpr{
									pos: position{line: 101, col: 7, offset: 3153},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 101, col: 7, offset: 3153},
											expr: &litMatcher{
												pos:        position{line: 101, col: 8, offset: 3154},
												val:        "{",
												ignoreCase: false,
											},
										},
										&choiceExpr{
											pos: position{line: 101, col: 14, offset: 3160},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 101, col: 14, offset: 3160},
													name: "SourceChar",
												},
												&ruleRefExpr{
													pos:  position{line: 101, col: 27, offset: 3173},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 101, col: 33, offset: 3179},
													name: "EOF",
												},
											},
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 102, col: 7, offset: 3250},
								run: (*parser).callonUnicodeClassEscape13,
								expr: &seqExpr{
									pos: position{line: 102, col: 7, offset: 3250},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 102, col: 7, offset: 3250},
											val:        "{",
											ignoreCase: false,
										},
										&labeledExpr{
											pos:   position{line: 102, col: 11, offset: 3254},
											label: "ident",
											expr: &ruleRefExpr{
												pos:  position{line: 102, col: 17, offset: 3260},
												name: "IdentifierName",
											},
										},
										&litMatcher{
											pos:        position{line: 102, col: 32, offset: 3275},
											val:        "}",
											ignoreCase: false,
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 108, col: 7, offset: 3439},
								run: (*parser).callonUnicodeClassEscape19,
								expr: &seqExpr{
									pos: position{line: 108, col: 7, offset: 3439},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 108, col: 7, offset: 3439},
											val:        "{",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 108, col: 11, offset: 3443},
											name: "IdentifierName",
										},
										&choiceExpr{
											pos: position{line: 108, col: 28, offset: 3460},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 108, col: 28, offset: 3460},
													val:        "]",
													ignoreCase: false,
												},
												&ruleRefExpr{
													pos:  position{line: 108, col: 34, offset: 3466},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 108, col: 40, offset: 3472},
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
			},
		},
		{
			name: "SingleCharUnicodeClass",
			pos:  position{line: 113, col: 1, offset: 3552},
			expr: &charClassMatcher{
				pos:        position{line: 113, col: 26, offset: 3577},
				val:        "[LMNCPZS]",
				chars:      []rune{'L', 'M', 'N', 'C', 'P', 'Z', 'S'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Number",
			pos:  position{line: 116, col: 1, offset: 3589},
			expr: &actionExpr{
				pos: position{line: 116, col: 10, offset: 3598},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 116, col: 10, offset: 3598},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 116, col: 10, offset: 3598},
							expr: &litMatcher{
								pos:        position{line: 116, col: 10, offset: 3598},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 116, col: 15, offset: 3603},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 116, col: 23, offset: 3611},
							expr: &seqExpr{
								pos: position{line: 116, col: 25, offset: 3613},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 116, col: 25, offset: 3613},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 116, col: 29, offset: 3617},
										expr: &ruleRefExpr{
											pos:  position{line: 116, col: 29, offset: 3617},
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
			pos:  position{line: 120, col: 1, offset: 3669},
			expr: &choiceExpr{
				pos: position{line: 120, col: 11, offset: 3679},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 120, col: 11, offset: 3679},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 120, col: 17, offset: 3685},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 120, col: 17, offset: 3685},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 120, col: 17, offset: 3685},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 120, col: 30, offset: 3698},
									expr: &ruleRefExpr{
										pos:  position{line: 120, col: 30, offset: 3698},
										name: "Digit",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "NonZeroDigit",
			pos:  position{line: 124, col: 1, offset: 3762},
			expr: &charClassMatcher{
				pos:        position{line: 124, col: 16, offset: 3777},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 125, col: 1, offset: 3783},
			expr: &charClassMatcher{
				pos:        position{line: 125, col: 9, offset: 3791},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "LabelBlock",
			pos:  position{line: 127, col: 1, offset: 3798},
			expr: &choiceExpr{
				pos: position{line: 127, col: 14, offset: 3811},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 127, col: 14, offset: 3811},
						run: (*parser).callonLabelBlock2,
						expr: &seqExpr{
							pos: position{line: 127, col: 14, offset: 3811},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 127, col: 14, offset: 3811},
									val:        "{",
									ignoreCase: false,
								},
								&labeledExpr{
									pos:   position{line: 127, col: 18, offset: 3815},
									label: "block",
									expr: &ruleRefExpr{
										pos:  position{line: 127, col: 24, offset: 3821},
										name: "LabelMatches",
									},
								},
								&litMatcher{
									pos:        position{line: 127, col: 37, offset: 3834},
									val:        "}",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 129, col: 5, offset: 3866},
						run: (*parser).callonLabelBlock8,
						expr: &seqExpr{
							pos: position{line: 129, col: 5, offset: 3866},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 129, col: 5, offset: 3866},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 129, col: 9, offset: 3870},
									name: "LabelMatches",
								},
								&ruleRefExpr{
									pos:  position{line: 129, col: 22, offset: 3883},
									name: "EOF",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "NanoSecondUnits",
			pos:  position{line: 133, col: 1, offset: 3948},
			expr: &actionExpr{
				pos: position{line: 133, col: 19, offset: 3966},
				run: (*parser).callonNanoSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 133, col: 19, offset: 3966},
					val:        "ns",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 138, col: 1, offset: 4071},
			expr: &actionExpr{
				pos: position{line: 138, col: 20, offset: 4090},
				run: (*parser).callonMicroSecondUnits1,
				expr: &choiceExpr{
					pos: position{line: 138, col: 21, offset: 4091},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 138, col: 21, offset: 4091},
							val:        "us",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 138, col: 28, offset: 4098},
							val:        "µs",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 138, col: 35, offset: 4106},
							val:        "μs",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 143, col: 1, offset: 4215},
			expr: &actionExpr{
				pos: position{line: 143, col: 20, offset: 4234},
				run: (*parser).callonMilliSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 143, col: 20, offset: 4234},
					val:        "ms",
					ignoreCase: false,
				},
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 148, col: 1, offset: 4341},
			expr: &actionExpr{
				pos: position{line: 148, col: 15, offset: 4355},
				run: (*parser).callonSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 148, col: 15, offset: 4355},
					val:        "s",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 152, col: 1, offset: 4392},
			expr: &actionExpr{
				pos: position{line: 152, col: 15, offset: 4406},
				run: (*parser).callonMinuteUnits1,
				expr: &litMatcher{
					pos:        position{line: 152, col: 15, offset: 4406},
					val:        "m",
					ignoreCase: false,
				},
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 156, col: 1, offset: 4443},
			expr: &actionExpr{
				pos: position{line: 156, col: 13, offset: 4455},
				run: (*parser).callonHourUnits1,
				expr: &litMatcher{
					pos:        position{line: 156, col: 13, offset: 4455},
					val:        "h",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DayUnits",
			pos:  position{line: 160, col: 1, offset: 4490},
			expr: &actionExpr{
				pos: position{line: 160, col: 12, offset: 4501},
				run: (*parser).callonDayUnits1,
				expr: &litMatcher{
					pos:        position{line: 160, col: 12, offset: 4501},
					val:        "d",
					ignoreCase: false,
				},
			},
		},
		{
			name: "WeekUnits",
			pos:  position{line: 166, col: 1, offset: 4709},
			expr: &actionExpr{
				pos: position{line: 166, col: 13, offset: 4721},
				run: (*parser).callonWeekUnits1,
				expr: &litMatcher{
					pos:        position{line: 166, col: 13, offset: 4721},
					val:        "w",
					ignoreCase: false,
				},
			},
		},
		{
			name: "YearUnits",
			pos:  position{line: 172, col: 1, offset: 4932},
			expr: &actionExpr{
				pos: position{line: 172, col: 13, offset: 4944},
				run: (*parser).callonYearUnits1,
				expr: &litMatcher{
					pos:        position{line: 172, col: 13, offset: 4944},
					val:        "y",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 178, col: 1, offset: 5141},
			expr: &choiceExpr{
				pos: position{line: 178, col: 18, offset: 5158},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 178, col: 18, offset: 5158},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 36, offset: 5176},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 55, offset: 5195},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 74, offset: 5214},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 88, offset: 5228},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 102, offset: 5242},
						name: "HourUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 114, offset: 5254},
						name: "DayUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 125, offset: 5265},
						name: "WeekUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 178, col: 137, offset: 5277},
						name: "YearUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 180, col: 1, offset: 5289},
			expr: &actionExpr{
				pos: position{line: 180, col: 12, offset: 5300},
				run: (*parser).callonDuration1,
				expr: &seqExpr{
					pos: position{line: 180, col: 12, offset: 5300},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 180, col: 12, offset: 5300},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 180, col: 16, offset: 5304},
								name: "Integer",
							},
						},
						&labeledExpr{
							pos:   position{line: 180, col: 24, offset: 5312},
							label: "units",
							expr: &ruleRefExpr{
								pos:  position{line: 180, col: 30, offset: 5318},
								name: "DurationUnits",
							},
						},
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 186, col: 1, offset: 5467},
			expr: &choiceExpr{
				pos: position{line: 186, col: 13, offset: 5479},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 186, col: 13, offset: 5479},
						val:        "-",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 19, offset: 5485},
						val:        "+",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 25, offset: 5491},
						val:        "*",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 31, offset: 5497},
						val:        "%",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 37, offset: 5503},
						val:        "/",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 43, offset: 5509},
						val:        "==",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 50, offset: 5516},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 57, offset: 5523},
						val:        "<=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 64, offset: 5530},
						val:        "<",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 70, offset: 5536},
						val:        ">=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 77, offset: 5543},
						val:        ">",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 83, offset: 5549},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 90, offset: 5556},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 97, offset: 5563},
						val:        "^",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 103, offset: 5569},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "LabelOperators",
			pos:  position{line: 188, col: 1, offset: 5574},
			expr: &choiceExpr{
				pos: position{line: 188, col: 19, offset: 5592},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 188, col: 19, offset: 5592},
						run: (*parser).callonLabelOperators2,
						expr: &litMatcher{
							pos:        position{line: 188, col: 19, offset: 5592},
							val:        "!=",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 190, col: 5, offset: 5628},
						run: (*parser).callonLabelOperators4,
						expr: &litMatcher{
							pos:        position{line: 190, col: 5, offset: 5628},
							val:        "=~",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 192, col: 5, offset: 5666},
						run: (*parser).callonLabelOperators6,
						expr: &litMatcher{
							pos:        position{line: 192, col: 5, offset: 5666},
							val:        "!~",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 194, col: 5, offset: 5706},
						run: (*parser).callonLabelOperators8,
						expr: &litMatcher{
							pos:        position{line: 194, col: 5, offset: 5706},
							val:        "=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Label",
			pos:  position{line: 198, col: 1, offset: 5737},
			expr: &ruleRefExpr{
				pos:  position{line: 198, col: 9, offset: 5745},
				name: "Identifier",
			},
		},
		{
			name: "LabelMatch",
			pos:  position{line: 199, col: 1, offset: 5756},
			expr: &actionExpr{
				pos: position{line: 199, col: 14, offset: 5769},
				run: (*parser).callonLabelMatch1,
				expr: &seqExpr{
					pos: position{line: 199, col: 14, offset: 5769},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 199, col: 14, offset: 5769},
							label: "label",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 20, offset: 5775},
								name: "Label",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 26, offset: 5781},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 29, offset: 5784},
							label: "op",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 32, offset: 5787},
								name: "LabelOperators",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 47, offset: 5802},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 50, offset: 5805},
							label: "match",
							expr: &choiceExpr{
								pos: position{line: 199, col: 58, offset: 5813},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 199, col: 58, offset: 5813},
										name: "StringLiteral",
									},
									&ruleRefExpr{
										pos:  position{line: 199, col: 74, offset: 5829},
										name: "Number",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "LabelMatches",
			pos:  position{line: 202, col: 1, offset: 5919},
			expr: &actionExpr{
				pos: position{line: 202, col: 16, offset: 5934},
				run: (*parser).callonLabelMatches1,
				expr: &seqExpr{
					pos: position{line: 202, col: 16, offset: 5934},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 202, col: 16, offset: 5934},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 202, col: 22, offset: 5940},
								name: "LabelMatch",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 202, col: 33, offset: 5951},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 202, col: 36, offset: 5954},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 202, col: 41, offset: 5959},
								expr: &ruleRefExpr{
									pos:  position{line: 202, col: 41, offset: 5959},
									name: "LabelMatchesRest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "LabelMatchesRest",
			pos:  position{line: 206, col: 1, offset: 6038},
			expr: &actionExpr{
				pos: position{line: 206, col: 21, offset: 6058},
				run: (*parser).callonLabelMatchesRest1,
				expr: &seqExpr{
					pos: position{line: 206, col: 21, offset: 6058},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 206, col: 21, offset: 6058},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 206, col: 25, offset: 6062},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 206, col: 28, offset: 6065},
							label: "match",
							expr: &ruleRefExpr{
								pos:  position{line: 206, col: 34, offset: 6071},
								name: "LabelMatch",
							},
						},
					},
				},
			},
		},
		{
			name: "LabelList",
			pos:  position{line: 210, col: 1, offset: 6109},
			expr: &choiceExpr{
				pos: position{line: 210, col: 13, offset: 6121},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 210, col: 13, offset: 6121},
						run: (*parser).callonLabelList2,
						expr: &seqExpr{
							pos: position{line: 210, col: 14, offset: 6122},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 210, col: 14, offset: 6122},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 210, col: 18, offset: 6126},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 210, col: 21, offset: 6129},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 212, col: 6, offset: 6161},
						run: (*parser).callonLabelList7,
						expr: &seqExpr{
							pos: position{line: 212, col: 6, offset: 6161},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 212, col: 6, offset: 6161},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 212, col: 10, offset: 6165},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 212, col: 13, offset: 6168},
									label: "label",
									expr: &ruleRefExpr{
										pos:  position{line: 212, col: 19, offset: 6174},
										name: "Label",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 212, col: 25, offset: 6180},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 212, col: 28, offset: 6183},
									label: "rest",
									expr: &zeroOrMoreExpr{
										pos: position{line: 212, col: 33, offset: 6188},
										expr: &ruleRefExpr{
											pos:  position{line: 212, col: 33, offset: 6188},
											name: "LabelListRest",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 212, col: 48, offset: 6203},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 212, col: 51, offset: 6206},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "LabelListRest",
			pos:  position{line: 216, col: 1, offset: 6272},
			expr: &actionExpr{
				pos: position{line: 216, col: 18, offset: 6289},
				run: (*parser).callonLabelListRest1,
				expr: &seqExpr{
					pos: position{line: 216, col: 18, offset: 6289},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 216, col: 18, offset: 6289},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 216, col: 22, offset: 6293},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 216, col: 25, offset: 6296},
							label: "label",
							expr: &ruleRefExpr{
								pos:  position{line: 216, col: 31, offset: 6302},
								name: "Label",
							},
						},
					},
				},
			},
		},
		{
			name: "VectorSelector",
			pos:  position{line: 220, col: 1, offset: 6335},
			expr: &actionExpr{
				pos: position{line: 220, col: 18, offset: 6352},
				run: (*parser).callonVectorSelector1,
				expr: &seqExpr{
					pos: position{line: 220, col: 18, offset: 6352},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 220, col: 18, offset: 6352},
							label: "metric",
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 25, offset: 6359},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 36, offset: 6370},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 40, offset: 6374},
							label: "block",
							expr: &zeroOrOneExpr{
								pos: position{line: 220, col: 46, offset: 6380},
								expr: &ruleRefExpr{
									pos:  position{line: 220, col: 46, offset: 6380},
									name: "LabelBlock",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 58, offset: 6392},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 61, offset: 6395},
							label: "rng",
							expr: &zeroOrOneExpr{
								pos: position{line: 220, col: 65, offset: 6399},
								expr: &ruleRefExpr{
									pos:  position{line: 220, col: 65, offset: 6399},
									name: "Range",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 72, offset: 6406},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 75, offset: 6409},
							label: "offset",
							expr: &zeroOrOneExpr{
								pos: position{line: 220, col: 82, offset: 6416},
								expr: &ruleRefExpr{
									pos:  position{line: 220, col: 82, offset: 6416},
									name: "Offset",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Range",
			pos:  position{line: 224, col: 1, offset: 6494},
			expr: &actionExpr{
				pos: position{line: 224, col: 9, offset: 6502},
				run: (*parser).callonRange1,
				expr: &seqExpr{
					pos: position{line: 224, col: 9, offset: 6502},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 224, col: 9, offset: 6502},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 13, offset: 6506},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 224, col: 16, offset: 6509},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 224, col: 20, offset: 6513},
								name: "Duration",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 29, offset: 6522},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 224, col: 32, offset: 6525},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Offset",
			pos:  position{line: 228, col: 1, offset: 6554},
			expr: &actionExpr{
				pos: position{line: 228, col: 10, offset: 6563},
				run: (*parser).callonOffset1,
				expr: &seqExpr{
					pos: position{line: 228, col: 10, offset: 6563},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 10, offset: 6563},
							val:        "offset",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 20, offset: 6573},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 6576},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 6580},
								name: "Duration",
							},
						},
					},
				},
			},
		},
		{
			name: "CountValueOperator",
			pos:  position{line: 232, col: 1, offset: 6614},
			expr: &actionExpr{
				pos: position{line: 232, col: 22, offset: 6635},
				run: (*parser).callonCountValueOperator1,
				expr: &litMatcher{
					pos:        position{line: 232, col: 22, offset: 6635},
					val:        "count_values",
					ignoreCase: true,
				},
			},
		},
		{
			name: "BinaryAggregateOperators",
			pos:  position{line: 238, col: 1, offset: 6720},
			expr: &actionExpr{
				pos: position{line: 238, col: 29, offset: 6748},
				run: (*parser).callonBinaryAggregateOperators1,
				expr: &labeledExpr{
					pos:   position{line: 238, col: 29, offset: 6748},
					label: "op",
					expr: &choiceExpr{
						pos: position{line: 238, col: 33, offset: 6752},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 238, col: 33, offset: 6752},
								val:        "topk",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 238, col: 43, offset: 6762},
								val:        "bottomk",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 238, col: 56, offset: 6775},
								val:        "quantile",
								ignoreCase: true,
							},
						},
					},
				},
			},
		},
		{
			name: "UnaryAggregateOperators",
			pos:  position{line: 244, col: 1, offset: 6877},
			expr: &actionExpr{
				pos: position{line: 244, col: 27, offset: 6903},
				run: (*parser).callonUnaryAggregateOperators1,
				expr: &labeledExpr{
					pos:   position{line: 244, col: 27, offset: 6903},
					label: "op",
					expr: &choiceExpr{
						pos: position{line: 244, col: 31, offset: 6907},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 244, col: 31, offset: 6907},
								val:        "sum",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 40, offset: 6916},
								val:        "min",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 49, offset: 6925},
								val:        "max",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 58, offset: 6934},
								val:        "avg",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 67, offset: 6943},
								val:        "stddev",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 79, offset: 6955},
								val:        "stdvar",
								ignoreCase: true,
							},
							&litMatcher{
								pos:        position{line: 244, col: 91, offset: 6967},
								val:        "count",
								ignoreCase: true,
							},
						},
					},
				},
			},
		},
		{
			name: "AggregateOperators",
			pos:  position{line: 250, col: 1, offset: 7066},
			expr: &choiceExpr{
				pos: position{line: 250, col: 22, offset: 7087},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 250, col: 22, offset: 7087},
						name: "CountValueOperator",
					},
					&ruleRefExpr{
						pos:  position{line: 250, col: 43, offset: 7108},
						name: "BinaryAggregateOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 250, col: 70, offset: 7135},
						name: "UnaryAggregateOperators",
					},
				},
			},
		},
		{
			name: "AggregateBy",
			pos:  position{line: 252, col: 1, offset: 7160},
			expr: &actionExpr{
				pos: position{line: 252, col: 15, offset: 7174},
				run: (*parser).callonAggregateBy1,
				expr: &seqExpr{
					pos: position{line: 252, col: 15, offset: 7174},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 252, col: 15, offset: 7174},
							val:        "by",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 21, offset: 7180},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 24, offset: 7183},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 31, offset: 7190},
								name: "LabelList",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 41, offset: 7200},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 44, offset: 7203},
							label: "keep",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 49, offset: 7208},
								expr: &litMatcher{
									pos:        position{line: 252, col: 49, offset: 7208},
									val:        "keep_common",
									ignoreCase: true,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "AggregateWithout",
			pos:  position{line: 260, col: 1, offset: 7354},
			expr: &actionExpr{
				pos: position{line: 260, col: 20, offset: 7373},
				run: (*parser).callonAggregateWithout1,
				expr: &seqExpr{
					pos: position{line: 260, col: 20, offset: 7373},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 260, col: 20, offset: 7373},
							val:        "without",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 260, col: 31, offset: 7384},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 260, col: 34, offset: 7387},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 260, col: 41, offset: 7394},
								name: "LabelList",
							},
						},
					},
				},
			},
		},
		{
			name: "AggregateGroup",
			pos:  position{line: 267, col: 1, offset: 7506},
			expr: &choiceExpr{
				pos: position{line: 267, col: 18, offset: 7523},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 267, col: 18, offset: 7523},
						name: "AggregateBy",
					},
					&ruleRefExpr{
						pos:  position{line: 267, col: 32, offset: 7537},
						name: "AggregateWithout",
					},
				},
			},
		},
		{
			name: "AggregateExpression",
			pos:  position{line: 269, col: 1, offset: 7555},
			expr: &choiceExpr{
				pos: position{line: 270, col: 1, offset: 7577},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 270, col: 1, offset: 7577},
						run: (*parser).callonAggregateExpression2,
						expr: &seqExpr{
							pos: position{line: 270, col: 1, offset: 7577},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 270, col: 1, offset: 7577},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 270, col: 4, offset: 7580},
										name: "CountValueOperator",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 24, offset: 7600},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 270, col: 27, offset: 7603},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 31, offset: 7607},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 270, col: 34, offset: 7610},
									label: "param",
									expr: &ruleRefExpr{
										pos:  position{line: 270, col: 40, offset: 7616},
										name: "StringLiteral",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 54, offset: 7630},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 270, col: 57, offset: 7633},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 61, offset: 7637},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 270, col: 64, offset: 7640},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 270, col: 71, offset: 7647},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 86, offset: 7662},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 270, col: 89, offset: 7665},
									val:        ")",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 270, col: 93, offset: 7669},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 270, col: 96, offset: 7672},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 270, col: 102, offset: 7678},
										expr: &ruleRefExpr{
											pos:  position{line: 270, col: 102, offset: 7678},
											name: "AggregateGroup",
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 276, col: 1, offset: 7826},
						run: (*parser).callonAggregateExpression22,
						expr: &seqExpr{
							pos: position{line: 276, col: 1, offset: 7826},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 276, col: 1, offset: 7826},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 276, col: 4, offset: 7829},
										name: "CountValueOperator",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 24, offset: 7849},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 276, col: 27, offset: 7852},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 276, col: 33, offset: 7858},
										expr: &ruleRefExpr{
											pos:  position{line: 276, col: 33, offset: 7858},
											name: "AggregateGroup",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 49, offset: 7874},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 276, col: 52, offset: 7877},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 56, offset: 7881},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 276, col: 59, offset: 7884},
									label: "param",
									expr: &ruleRefExpr{
										pos:  position{line: 276, col: 65, offset: 7890},
										name: "StringLiteral",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 79, offset: 7904},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 276, col: 82, offset: 7907},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 86, offset: 7911},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 276, col: 89, offset: 7914},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 276, col: 96, offset: 7921},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 276, col: 111, offset: 7936},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 276, col: 114, offset: 7939},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 282, col: 1, offset: 8075},
						run: (*parser).callonAggregateExpression42,
						expr: &seqExpr{
							pos: position{line: 282, col: 1, offset: 8075},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 282, col: 1, offset: 8075},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 282, col: 4, offset: 8078},
										name: "BinaryAggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 30, offset: 8104},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 282, col: 33, offset: 8107},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 37, offset: 8111},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 282, col: 41, offset: 8115},
									label: "param",
									expr: &ruleRefExpr{
										pos:  position{line: 282, col: 47, offset: 8121},
										name: "Number",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 54, offset: 8128},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 282, col: 57, offset: 8131},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 61, offset: 8135},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 282, col: 64, offset: 8138},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 282, col: 71, offset: 8145},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 86, offset: 8160},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 282, col: 89, offset: 8163},
									val:        ")",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 93, offset: 8167},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 282, col: 96, offset: 8170},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 282, col: 102, offset: 8176},
										expr: &ruleRefExpr{
											pos:  position{line: 282, col: 102, offset: 8176},
											name: "AggregateGroup",
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 288, col: 1, offset: 8317},
						run: (*parser).callonAggregateExpression62,
						expr: &seqExpr{
							pos: position{line: 288, col: 1, offset: 8317},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 288, col: 1, offset: 8317},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 288, col: 4, offset: 8320},
										name: "BinaryAggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 30, offset: 8346},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 288, col: 33, offset: 8349},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 288, col: 39, offset: 8355},
										expr: &ruleRefExpr{
											pos:  position{line: 288, col: 39, offset: 8355},
											name: "AggregateGroup",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 55, offset: 8371},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 288, col: 58, offset: 8374},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 62, offset: 8378},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 288, col: 66, offset: 8382},
									label: "param",
									expr: &ruleRefExpr{
										pos:  position{line: 288, col: 72, offset: 8388},
										name: "Number",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 79, offset: 8395},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 288, col: 82, offset: 8398},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 86, offset: 8402},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 288, col: 89, offset: 8405},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 288, col: 96, offset: 8412},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 288, col: 111, offset: 8427},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 288, col: 114, offset: 8430},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 294, col: 1, offset: 8559},
						run: (*parser).callonAggregateExpression82,
						expr: &seqExpr{
							pos: position{line: 294, col: 1, offset: 8559},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 294, col: 1, offset: 8559},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 294, col: 4, offset: 8562},
										name: "UnaryAggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 294, col: 29, offset: 8587},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 294, col: 32, offset: 8590},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 294, col: 36, offset: 8594},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 294, col: 39, offset: 8597},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 294, col: 46, offset: 8604},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 294, col: 61, offset: 8619},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 294, col: 64, offset: 8622},
									val:        ")",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 294, col: 68, offset: 8626},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 294, col: 71, offset: 8629},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 294, col: 77, offset: 8635},
										expr: &ruleRefExpr{
											pos:  position{line: 294, col: 77, offset: 8635},
											name: "AggregateGroup",
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 298, col: 1, offset: 8728},
						run: (*parser).callonAggregateExpression97,
						expr: &seqExpr{
							pos: position{line: 298, col: 1, offset: 8728},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 298, col: 1, offset: 8728},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 298, col: 4, offset: 8731},
										name: "UnaryAggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 298, col: 29, offset: 8756},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 298, col: 32, offset: 8759},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 298, col: 38, offset: 8765},
										expr: &ruleRefExpr{
											pos:  position{line: 298, col: 38, offset: 8765},
											name: "AggregateGroup",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 298, col: 54, offset: 8781},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 298, col: 57, offset: 8784},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 298, col: 61, offset: 8788},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 298, col: 64, offset: 8791},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 298, col: 71, offset: 8798},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 298, col: 86, offset: 8813},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 298, col: 89, offset: 8816},
									val:        ")",
									ignoreCase: false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "__",
			pos:  position{line: 302, col: 1, offset: 8896},
			expr: &zeroOrMoreExpr{
				pos: position{line: 302, col: 6, offset: 8901},
				expr: &choiceExpr{
					pos: position{line: 302, col: 8, offset: 8903},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 302, col: 8, offset: 8903},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 21, offset: 8916},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 27, offset: 8922},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 303, col: 1, offset: 8933},
			expr: &zeroOrMoreExpr{
				pos: position{line: 303, col: 5, offset: 8937},
				expr: &ruleRefExpr{
					pos:  position{line: 303, col: 5, offset: 8937},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 305, col: 1, offset: 8950},
			expr: &charClassMatcher{
				pos:        position{line: 305, col: 14, offset: 8963},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 306, col: 1, offset: 8971},
			expr: &litMatcher{
				pos:        position{line: 306, col: 7, offset: 8977},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 307, col: 1, offset: 8982},
			expr: &choiceExpr{
				pos: position{line: 307, col: 7, offset: 8988},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 307, col: 7, offset: 8988},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 307, col: 7, offset: 8988},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 307, col: 10, offset: 8991},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 307, col: 16, offset: 8997},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 307, col: 16, offset: 8997},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 307, col: 18, offset: 8999},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 18, offset: 8999},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 307, col: 37, offset: 9018},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 307, col: 43, offset: 9024},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 307, col: 43, offset: 9024},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 307, col: 46, offset: 9027},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 309, col: 1, offset: 9032},
			expr: &notExpr{
				pos: position{line: 309, col: 7, offset: 9038},
				expr: &anyMatcher{
					line: 309, col: 8, offset: 9039,
				},
			},
		},
	},
}

func (c *current) onGrammar1(grammar interface{}) (interface{}, error) {
	return grammar, nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["grammar"])
}

func (c *current) onComment1() (interface{}, error) {
	return &Comment{string(c.text)}, nil
}

func (p *parser) callonComment1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onComment1()
}

func (c *current) onIdentifier1(ident interface{}) (interface{}, error) {
	i := string(c.text)
	if reservedWords[i] {
		return nil, errors.New("identifier is a reserved word")
	}
	return &Identifier{ident.(string)}, nil
}

func (p *parser) callonIdentifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifier1(stack["ident"])
}

func (c *current) onIdentifierName1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonIdentifierName1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifierName1()
}

func (c *current) onStringLiteral2() (interface{}, error) {
	str, err := strconv.Unquote(string(c.text))
	if err != nil {
		return nil, err
	}
	return &StringLiteral{str}, nil
}

func (p *parser) callonStringLiteral2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStringLiteral2()
}

func (c *current) onStringLiteral18() (interface{}, error) {
	return nil, errors.New("string literal not terminated")
}

func (p *parser) callonStringLiteral18() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStringLiteral18()
}

func (c *current) onDoubleStringEscape5() (interface{}, error) {
	return nil, errors.New("invalid escape character")
}

func (p *parser) callonDoubleStringEscape5() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDoubleStringEscape5()
}

func (c *current) onSingleStringEscape5() (interface{}, error) {
	return nil, errors.New("invalid escape character")
}

func (p *parser) callonSingleStringEscape5() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSingleStringEscape5()
}

func (c *current) onOctalEscape6() (interface{}, error) {
	return nil, errors.New("invalid octal escape")
}

func (p *parser) callonOctalEscape6() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOctalEscape6()
}

func (c *current) onHexEscape6() (interface{}, error) {
	return nil, errors.New("invalid hexadecimal escape")
}

func (p *parser) callonHexEscape6() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onHexEscape6()
}

func (c *current) onLongUnicodeEscape2() (interface{}, error) {
	return validateUnicodeEscape(string(c.text), "invalid Unicode escape")

}

func (p *parser) callonLongUnicodeEscape2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLongUnicodeEscape2()
}

func (c *current) onLongUnicodeEscape13() (interface{}, error) {
	return nil, errors.New("invalid Unicode escape")
}

func (p *parser) callonLongUnicodeEscape13() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLongUnicodeEscape13()
}

func (c *current) onShortUnicodeEscape2() (interface{}, error) {
	return validateUnicodeEscape(string(c.text), "invalid Unicode escape")

}

func (p *parser) callonShortUnicodeEscape2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onShortUnicodeEscape2()
}

func (c *current) onShortUnicodeEscape9() (interface{}, error) {
	return nil, errors.New("invalid Unicode escape")
}

func (p *parser) callonShortUnicodeEscape9() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onShortUnicodeEscape9()
}

func (c *current) onCharClassMatcher2() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonCharClassMatcher2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCharClassMatcher2()
}

func (c *current) onCharClassMatcher15() (interface{}, error) {
	return nil, errors.New("character class not terminated")
}

func (p *parser) callonCharClassMatcher15() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCharClassMatcher15()
}

func (c *current) onCharClassEscape5() (interface{}, error) {
	return nil, errors.New("invalid escape character")
}

func (p *parser) callonCharClassEscape5() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCharClassEscape5()
}

func (c *current) onUnicodeClassEscape5() (interface{}, error) {
	return nil, errors.New("invalid Unicode class escape")
}

func (p *parser) callonUnicodeClassEscape5() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnicodeClassEscape5()
}

func (c *current) onUnicodeClassEscape13(ident interface{}) (interface{}, error) {
	if !unicodeClasses[ident.(string)] {
		return nil, errors.New("invalid Unicode class escape")
	}
	return nil, nil

}

func (p *parser) callonUnicodeClassEscape13() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnicodeClassEscape13(stack["ident"])
}

func (c *current) onUnicodeClassEscape19() (interface{}, error) {
	return nil, errors.New("Unicode class not terminated")

}

func (p *parser) callonUnicodeClassEscape19() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnicodeClassEscape19()
}

func (c *current) onNumber1() (interface{}, error) {
	return NewNumber(string(c.text))
}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onInteger3() (interface{}, error) {
	return strconv.ParseInt(string(c.text), 10, 64)
}

func (p *parser) callonInteger3() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInteger3()
}

func (c *current) onLabelBlock2(block interface{}) (interface{}, error) {
	return block, nil
}

func (p *parser) callonLabelBlock2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelBlock2(stack["block"])
}

func (c *current) onLabelBlock8() (interface{}, error) {
	return nil, errors.New("code block not terminated")
}

func (p *parser) callonLabelBlock8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelBlock8()
}

func (c *current) onNanoSecondUnits1() (interface{}, error) {
	// Prometheus doesn't support nanoseconds, but, influx does
	return time.Nanosecond, nil
}

func (p *parser) callonNanoSecondUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNanoSecondUnits1()
}

func (c *current) onMicroSecondUnits1() (interface{}, error) {
	// Prometheus doesn't support nanoseconds, but, influx does
	return time.Microsecond, nil
}

func (p *parser) callonMicroSecondUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMicroSecondUnits1()
}

func (c *current) onMilliSecondUnits1() (interface{}, error) {
	// Prometheus doesn't support nanoseconds, but, influx does
	return time.Millisecond, nil
}

func (p *parser) callonMilliSecondUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMilliSecondUnits1()
}

func (c *current) onSecondUnits1() (interface{}, error) {
	return time.Second, nil
}

func (p *parser) callonSecondUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSecondUnits1()
}

func (c *current) onMinuteUnits1() (interface{}, error) {
	return time.Minute, nil
}

func (p *parser) callonMinuteUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMinuteUnits1()
}

func (c *current) onHourUnits1() (interface{}, error) {
	return time.Hour, nil
}

func (p *parser) callonHourUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onHourUnits1()
}

func (c *current) onDayUnits1() (interface{}, error) {
	// Prometheus always assumes exactly 24 hours in a day
	// https://github.com/prometheus/common/blob/61f87aac8082fa8c3c5655c7608d7478d46ac2ad/model/time.go#L180
	return time.Hour * 24, nil
}

func (p *parser) callonDayUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDayUnits1()
}

func (c *current) onWeekUnits1() (interface{}, error) {
	// Prometheus always assumes exactly 7 days in a week
	// https://github.com/prometheus/common/blob/61f87aac8082fa8c3c5655c7608d7478d46ac2ad/model/time.go#L180
	return time.Hour * 24 * 7, nil
}

func (p *parser) callonWeekUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWeekUnits1()
}

func (c *current) onYearUnits1() (interface{}, error) {
	// Prometheus always assumes 365 days
	// https://github.com/prometheus/common/blob/61f87aac8082fa8c3c5655c7608d7478d46ac2ad/model/time.go#L180
	return time.Hour * 24 * 365, nil
}

func (p *parser) callonYearUnits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onYearUnits1()
}

func (c *current) onDuration1(dur, units interface{}) (interface{}, error) {
	nanos := time.Duration(dur.(int64))
	conversion := units.(time.Duration)
	return time.Duration(nanos) * conversion, nil
}

func (p *parser) callonDuration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDuration1(stack["dur"], stack["units"])
}

func (c *current) onLabelOperators2() (interface{}, error) {
	return NotEqual, nil
}

func (p *parser) callonLabelOperators2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelOperators2()
}

func (c *current) onLabelOperators4() (interface{}, error) {
	return RegexMatch, nil
}

func (p *parser) callonLabelOperators4() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelOperators4()
}

func (c *current) onLabelOperators6() (interface{}, error) {
	return RegexNoMatch, nil
}

func (p *parser) callonLabelOperators6() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelOperators6()
}

func (c *current) onLabelOperators8() (interface{}, error) {
	return Equal, nil
}

func (p *parser) callonLabelOperators8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelOperators8()
}

func (c *current) onLabelMatch1(label, op, match interface{}) (interface{}, error) {
	return NewLabelMatcher(label.(*Identifier), op.(MatchKind), match.(Arg))
}

func (p *parser) callonLabelMatch1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelMatch1(stack["label"], stack["op"], stack["match"])
}

func (c *current) onLabelMatches1(first, rest interface{}) (interface{}, error) {
	return NewLabelMatches(first.(*LabelMatcher), rest)
}

func (p *parser) callonLabelMatches1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelMatches1(stack["first"], stack["rest"])
}

func (c *current) onLabelMatchesRest1(match interface{}) (interface{}, error) {
	return match, nil
}

func (p *parser) callonLabelMatchesRest1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelMatchesRest1(stack["match"])
}

func (c *current) onLabelList2() (interface{}, error) {
	return nil, nil
}

func (p *parser) callonLabelList2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelList2()
}

func (c *current) onLabelList7(label, rest interface{}) (interface{}, error) {
	return NewIdentifierList(label.(*Identifier), rest)
}

func (p *parser) callonLabelList7() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelList7(stack["label"], stack["rest"])
}

func (c *current) onLabelListRest1(label interface{}) (interface{}, error) {
	return label, nil
}

func (p *parser) callonLabelListRest1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLabelListRest1(stack["label"])
}

func (c *current) onVectorSelector1(metric, block, rng, offset interface{}) (interface{}, error) {
	return NewSelector(metric.(*Identifier), block, rng, offset)
}

func (p *parser) callonVectorSelector1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onVectorSelector1(stack["metric"], stack["block"], stack["rng"], stack["offset"])
}

func (c *current) onRange1(dur interface{}) (interface{}, error) {
	return dur, nil
}

func (p *parser) callonRange1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRange1(stack["dur"])
}

func (c *current) onOffset1(dur interface{}) (interface{}, error) {
	return dur, nil
}

func (p *parser) callonOffset1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOffset1(stack["dur"])
}

func (c *current) onCountValueOperator1() (interface{}, error) {
	return &Operator{
		Kind: CountValuesKind,
	}, nil
}

func (p *parser) callonCountValueOperator1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCountValueOperator1()
}

func (c *current) onBinaryAggregateOperators1(op interface{}) (interface{}, error) {
	return &Operator{
		Kind: ToOperatorKind(string(op.([]byte))),
	}, nil
}

func (p *parser) callonBinaryAggregateOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBinaryAggregateOperators1(stack["op"])
}

func (c *current) onUnaryAggregateOperators1(op interface{}) (interface{}, error) {
	return &Operator{
		Kind: ToOperatorKind(string(op.([]byte))),
	}, nil
}

func (p *parser) callonUnaryAggregateOperators1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnaryAggregateOperators1(stack["op"])
}

func (c *current) onAggregateBy1(labels, keep interface{}) (interface{}, error) {
	return &Aggregate{
		By:         true,
		KeepCommon: keep != nil,
		Labels:     labels.([]*Identifier),
	}, nil
}

func (p *parser) callonAggregateBy1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateBy1(stack["labels"], stack["keep"])
}

func (c *current) onAggregateWithout1(labels interface{}) (interface{}, error) {
	return &Aggregate{
		Without: true,
		Labels:  labels.([]*Identifier),
	}, nil
}

func (p *parser) callonAggregateWithout1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateWithout1(stack["labels"])
}

func (c *current) onAggregateExpression2(op, param, vector, group interface{}) (interface{}, error) {
	oper := op.(*Operator)
	oper.Arg = param.(*StringLiteral)
	return NewAggregateExpr(oper, vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression2(stack["op"], stack["param"], stack["vector"], stack["group"])
}

func (c *current) onAggregateExpression22(op, group, param, vector interface{}) (interface{}, error) {
	oper := op.(*Operator)
	oper.Arg = param.(*StringLiteral)
	return NewAggregateExpr(oper, vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression22() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression22(stack["op"], stack["group"], stack["param"], stack["vector"])
}

func (c *current) onAggregateExpression42(op, param, vector, group interface{}) (interface{}, error) {
	oper := op.(*Operator)
	oper.Arg = param.(*Number)
	return NewAggregateExpr(oper, vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression42() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression42(stack["op"], stack["param"], stack["vector"], stack["group"])
}

func (c *current) onAggregateExpression62(op, group, param, vector interface{}) (interface{}, error) {
	oper := op.(*Operator)
	oper.Arg = param.(*Number)
	return NewAggregateExpr(oper, vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression62() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression62(stack["op"], stack["group"], stack["param"], stack["vector"])
}

func (c *current) onAggregateExpression82(op, vector, group interface{}) (interface{}, error) {
	return NewAggregateExpr(op.(*Operator), vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression82() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression82(stack["op"], stack["vector"], stack["group"])
}

func (c *current) onAggregateExpression97(op, group, vector interface{}) (interface{}, error) {
	return NewAggregateExpr(op.(*Operator), vector.(*Selector), group)
}

func (p *parser) callonAggregateExpression97() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression97(stack["op"], stack["group"], stack["vector"])
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

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
