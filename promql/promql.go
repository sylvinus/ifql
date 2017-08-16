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
						&choiceExpr{
							pos: position{line: 10, col: 14, offset: 119},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 10, col: 14, offset: 119},
									name: "Comment",
								},
								&ruleRefExpr{
									pos:  position{line: 10, col: 24, offset: 129},
									name: "AggregateExpression",
								},
								&ruleRefExpr{
									pos:  position{line: 10, col: 46, offset: 151},
									name: "VectorSelector",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 10, col: 63, offset: 168},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 14, col: 1, offset: 208},
			expr: &anyMatcher{
				line: 14, col: 14, offset: 221,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 16, col: 1, offset: 224},
			expr: &seqExpr{
				pos: position{line: 16, col: 11, offset: 234},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 16, col: 11, offset: 234},
						val:        "#",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 16, col: 15, offset: 238},
						expr: &seqExpr{
							pos: position{line: 16, col: 17, offset: 240},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 16, col: 17, offset: 240},
									expr: &ruleRefExpr{
										pos:  position{line: 16, col: 18, offset: 241},
										name: "EOL",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 16, col: 22, offset: 245},
									name: "SourceChar",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Identifier",
			pos:  position{line: 18, col: 1, offset: 260},
			expr: &actionExpr{
				pos: position{line: 18, col: 14, offset: 273},
				run: (*parser).callonIdentifier1,
				expr: &labeledExpr{
					pos:   position{line: 18, col: 14, offset: 273},
					label: "ident",
					expr: &ruleRefExpr{
						pos:  position{line: 18, col: 20, offset: 279},
						name: "IdentifierName",
					},
				},
			},
		},
		{
			name: "IdentifierName",
			pos:  position{line: 26, col: 1, offset: 463},
			expr: &actionExpr{
				pos: position{line: 26, col: 18, offset: 480},
				run: (*parser).callonIdentifierName1,
				expr: &seqExpr{
					pos: position{line: 26, col: 18, offset: 480},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 26, col: 18, offset: 480},
							name: "IdentifierStart",
						},
						&zeroOrMoreExpr{
							pos: position{line: 26, col: 34, offset: 496},
							expr: &ruleRefExpr{
								pos:  position{line: 26, col: 34, offset: 496},
								name: "IdentifierPart",
							},
						},
					},
				},
			},
		},
		{
			name: "IdentifierStart",
			pos:  position{line: 29, col: 1, offset: 547},
			expr: &charClassMatcher{
				pos:        position{line: 29, col: 19, offset: 565},
				val:        "[\\pL_]",
				chars:      []rune{'_'},
				classes:    []*unicode.RangeTable{rangeTable("L")},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "IdentifierPart",
			pos:  position{line: 30, col: 1, offset: 572},
			expr: &choiceExpr{
				pos: position{line: 30, col: 18, offset: 589},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 30, col: 18, offset: 589},
						name: "IdentifierStart",
					},
					&charClassMatcher{
						pos:        position{line: 30, col: 36, offset: 607},
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
			pos:  position{line: 32, col: 1, offset: 617},
			expr: &choiceExpr{
				pos: position{line: 32, col: 17, offset: 633},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 32, col: 17, offset: 633},
						run: (*parser).callonStringLiteral2,
						expr: &choiceExpr{
							pos: position{line: 32, col: 19, offset: 635},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 32, col: 19, offset: 635},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 32, col: 19, offset: 635},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 32, col: 23, offset: 639},
											expr: &ruleRefExpr{
												pos:  position{line: 32, col: 23, offset: 639},
												name: "DoubleStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 32, col: 41, offset: 657},
											val:        "\"",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 32, col: 47, offset: 663},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 32, col: 47, offset: 663},
											val:        "'",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 32, col: 51, offset: 667},
											name: "SingleStringChar",
										},
										&litMatcher{
											pos:        position{line: 32, col: 68, offset: 684},
											val:        "'",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 32, col: 74, offset: 690},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 32, col: 74, offset: 690},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 32, col: 78, offset: 694},
											expr: &ruleRefExpr{
												pos:  position{line: 32, col: 78, offset: 694},
												name: "RawStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 32, col: 93, offset: 709},
											val:        "`",
											ignoreCase: false,
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 38, col: 5, offset: 855},
						run: (*parser).callonStringLiteral18,
						expr: &choiceExpr{
							pos: position{line: 38, col: 7, offset: 857},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 38, col: 9, offset: 859},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 38, col: 9, offset: 859},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 38, col: 13, offset: 863},
											expr: &ruleRefExpr{
												pos:  position{line: 38, col: 13, offset: 863},
												name: "DoubleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 38, col: 33, offset: 883},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 38, col: 33, offset: 883},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 38, col: 39, offset: 889},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 38, col: 51, offset: 901},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 38, col: 51, offset: 901},
											val:        "'",
											ignoreCase: false,
										},
										&zeroOrOneExpr{
											pos: position{line: 38, col: 55, offset: 905},
											expr: &ruleRefExpr{
												pos:  position{line: 38, col: 55, offset: 905},
												name: "SingleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 38, col: 75, offset: 925},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 38, col: 75, offset: 925},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 38, col: 81, offset: 931},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 38, col: 91, offset: 941},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 38, col: 91, offset: 941},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 38, col: 95, offset: 945},
											expr: &ruleRefExpr{
												pos:  position{line: 38, col: 95, offset: 945},
												name: "RawStringChar",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 38, col: 110, offset: 960},
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
			pos:  position{line: 42, col: 1, offset: 1031},
			expr: &choiceExpr{
				pos: position{line: 42, col: 20, offset: 1050},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 42, col: 20, offset: 1050},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 42, col: 20, offset: 1050},
								expr: &choiceExpr{
									pos: position{line: 42, col: 23, offset: 1053},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 42, col: 23, offset: 1053},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 42, col: 29, offset: 1059},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 42, col: 36, offset: 1066},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 42, col: 42, offset: 1072},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 42, col: 55, offset: 1085},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 42, col: 55, offset: 1085},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 42, col: 60, offset: 1090},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "SingleStringChar",
			pos:  position{line: 43, col: 1, offset: 1109},
			expr: &choiceExpr{
				pos: position{line: 43, col: 20, offset: 1128},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 43, col: 20, offset: 1128},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 43, col: 20, offset: 1128},
								expr: &choiceExpr{
									pos: position{line: 43, col: 23, offset: 1131},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 43, col: 23, offset: 1131},
											val:        "'",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 43, col: 29, offset: 1137},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 43, col: 36, offset: 1144},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 43, col: 42, offset: 1150},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 43, col: 55, offset: 1163},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 43, col: 55, offset: 1163},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 43, col: 60, offset: 1168},
								name: "SingleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "RawStringChar",
			pos:  position{line: 44, col: 1, offset: 1187},
			expr: &seqExpr{
				pos: position{line: 44, col: 17, offset: 1203},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 44, col: 17, offset: 1203},
						expr: &litMatcher{
							pos:        position{line: 44, col: 18, offset: 1204},
							val:        "`",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 44, col: 22, offset: 1208},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 46, col: 1, offset: 1220},
			expr: &choiceExpr{
				pos: position{line: 46, col: 22, offset: 1241},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 46, col: 24, offset: 1243},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 46, col: 24, offset: 1243},
								val:        "\"",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 46, col: 30, offset: 1249},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 47, col: 7, offset: 1278},
						run: (*parser).callonDoubleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 47, col: 9, offset: 1280},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 47, col: 9, offset: 1280},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 47, col: 22, offset: 1293},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 47, col: 28, offset: 1299},
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
			pos:  position{line: 50, col: 1, offset: 1364},
			expr: &choiceExpr{
				pos: position{line: 50, col: 22, offset: 1385},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 50, col: 24, offset: 1387},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 50, col: 24, offset: 1387},
								val:        "'",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 50, col: 30, offset: 1393},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 51, col: 7, offset: 1422},
						run: (*parser).callonSingleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 51, col: 9, offset: 1424},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 51, col: 9, offset: 1424},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 51, col: 22, offset: 1437},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 51, col: 28, offset: 1443},
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
			pos:  position{line: 55, col: 1, offset: 1509},
			expr: &choiceExpr{
				pos: position{line: 55, col: 24, offset: 1532},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 55, col: 24, offset: 1532},
						name: "SingleCharEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 43, offset: 1551},
						name: "OctalEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 57, offset: 1565},
						name: "HexEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 69, offset: 1577},
						name: "LongUnicodeEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 55, col: 89, offset: 1597},
						name: "ShortUnicodeEscape",
					},
				},
			},
		},
		{
			name: "SingleCharEscape",
			pos:  position{line: 56, col: 1, offset: 1616},
			expr: &choiceExpr{
				pos: position{line: 56, col: 20, offset: 1635},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 56, col: 20, offset: 1635},
						val:        "a",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 26, offset: 1641},
						val:        "b",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 32, offset: 1647},
						val:        "n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 38, offset: 1653},
						val:        "f",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 44, offset: 1659},
						val:        "r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 50, offset: 1665},
						val:        "t",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 56, offset: 1671},
						val:        "v",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 56, col: 62, offset: 1677},
						val:        "\\",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "OctalEscape",
			pos:  position{line: 57, col: 1, offset: 1682},
			expr: &choiceExpr{
				pos: position{line: 57, col: 15, offset: 1696},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 57, col: 15, offset: 1696},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 57, col: 15, offset: 1696},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 57, col: 26, offset: 1707},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 57, col: 37, offset: 1718},
								name: "OctalDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 58, col: 7, offset: 1735},
						run: (*parser).callonOctalEscape6,
						expr: &seqExpr{
							pos: position{line: 58, col: 7, offset: 1735},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 58, col: 7, offset: 1735},
									name: "OctalDigit",
								},
								&choiceExpr{
									pos: position{line: 58, col: 20, offset: 1748},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 58, col: 20, offset: 1748},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 58, col: 33, offset: 1761},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 58, col: 39, offset: 1767},
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
			pos:  position{line: 61, col: 1, offset: 1828},
			expr: &choiceExpr{
				pos: position{line: 61, col: 13, offset: 1840},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 61, col: 13, offset: 1840},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 61, col: 13, offset: 1840},
								val:        "x",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 61, col: 17, offset: 1844},
								name: "HexDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 61, col: 26, offset: 1853},
								name: "HexDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 62, col: 7, offset: 1868},
						run: (*parser).callonHexEscape6,
						expr: &seqExpr{
							pos: position{line: 62, col: 7, offset: 1868},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 62, col: 7, offset: 1868},
									val:        "x",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 62, col: 13, offset: 1874},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 62, col: 13, offset: 1874},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 62, col: 26, offset: 1887},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 62, col: 32, offset: 1893},
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
			pos:  position{line: 65, col: 1, offset: 1960},
			expr: &choiceExpr{
				pos: position{line: 66, col: 5, offset: 1985},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 66, col: 5, offset: 1985},
						run: (*parser).callonLongUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 66, col: 5, offset: 1985},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 66, col: 5, offset: 1985},
									val:        "U",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 9, offset: 1989},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 18, offset: 1998},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 27, offset: 2007},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 36, offset: 2016},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 45, offset: 2025},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 54, offset: 2034},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 63, offset: 2043},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 66, col: 72, offset: 2052},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 69, col: 7, offset: 2154},
						run: (*parser).callonLongUnicodeEscape13,
						expr: &seqExpr{
							pos: position{line: 69, col: 7, offset: 2154},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 69, col: 7, offset: 2154},
									val:        "U",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 69, col: 13, offset: 2160},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 69, col: 13, offset: 2160},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 26, offset: 2173},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 69, col: 32, offset: 2179},
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
			pos:  position{line: 72, col: 1, offset: 2242},
			expr: &choiceExpr{
				pos: position{line: 73, col: 5, offset: 2268},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 73, col: 5, offset: 2268},
						run: (*parser).callonShortUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 73, col: 5, offset: 2268},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 73, col: 5, offset: 2268},
									val:        "u",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 73, col: 9, offset: 2272},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 73, col: 18, offset: 2281},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 73, col: 27, offset: 2290},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 73, col: 36, offset: 2299},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 76, col: 7, offset: 2401},
						run: (*parser).callonShortUnicodeEscape9,
						expr: &seqExpr{
							pos: position{line: 76, col: 7, offset: 2401},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 76, col: 7, offset: 2401},
									val:        "u",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 76, col: 13, offset: 2407},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 76, col: 13, offset: 2407},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 76, col: 26, offset: 2420},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 76, col: 32, offset: 2426},
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
			pos:  position{line: 80, col: 1, offset: 2490},
			expr: &charClassMatcher{
				pos:        position{line: 80, col: 14, offset: 2503},
				val:        "[0-7]",
				ranges:     []rune{'0', '7'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 81, col: 1, offset: 2509},
			expr: &charClassMatcher{
				pos:        position{line: 81, col: 16, offset: 2524},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "HexDigit",
			pos:  position{line: 82, col: 1, offset: 2530},
			expr: &charClassMatcher{
				pos:        position{line: 82, col: 12, offset: 2541},
				val:        "[0-9a-f]i",
				ranges:     []rune{'0', '9', 'a', 'f'},
				ignoreCase: true,
				inverted:   false,
			},
		},
		{
			name: "CharClassMatcher",
			pos:  position{line: 84, col: 1, offset: 2552},
			expr: &choiceExpr{
				pos: position{line: 84, col: 20, offset: 2571},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 84, col: 20, offset: 2571},
						run: (*parser).callonCharClassMatcher2,
						expr: &seqExpr{
							pos: position{line: 84, col: 20, offset: 2571},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 84, col: 20, offset: 2571},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 84, col: 24, offset: 2575},
									expr: &choiceExpr{
										pos: position{line: 84, col: 26, offset: 2577},
										alternatives: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 84, col: 26, offset: 2577},
												name: "ClassCharRange",
											},
											&ruleRefExpr{
												pos:  position{line: 84, col: 43, offset: 2594},
												name: "ClassChar",
											},
											&seqExpr{
												pos: position{line: 84, col: 55, offset: 2606},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 84, col: 55, offset: 2606},
														val:        "\\",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 84, col: 60, offset: 2611},
														name: "UnicodeClassEscape",
													},
												},
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 84, col: 82, offset: 2633},
									val:        "]",
									ignoreCase: false,
								},
								&zeroOrOneExpr{
									pos: position{line: 84, col: 86, offset: 2637},
									expr: &litMatcher{
										pos:        position{line: 84, col: 86, offset: 2637},
										val:        "i",
										ignoreCase: false,
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 86, col: 5, offset: 2679},
						run: (*parser).callonCharClassMatcher15,
						expr: &seqExpr{
							pos: position{line: 86, col: 5, offset: 2679},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 86, col: 5, offset: 2679},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 86, col: 9, offset: 2683},
									expr: &seqExpr{
										pos: position{line: 86, col: 11, offset: 2685},
										exprs: []interface{}{
											&notExpr{
												pos: position{line: 86, col: 11, offset: 2685},
												expr: &ruleRefExpr{
													pos:  position{line: 86, col: 14, offset: 2688},
													name: "EOL",
												},
											},
											&ruleRefExpr{
												pos:  position{line: 86, col: 20, offset: 2694},
												name: "SourceChar",
											},
										},
									},
								},
								&choiceExpr{
									pos: position{line: 86, col: 36, offset: 2710},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 86, col: 36, offset: 2710},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 86, col: 42, offset: 2716},
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
			pos:  position{line: 90, col: 1, offset: 2788},
			expr: &seqExpr{
				pos: position{line: 90, col: 18, offset: 2805},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 90, col: 18, offset: 2805},
						name: "ClassChar",
					},
					&litMatcher{
						pos:        position{line: 90, col: 28, offset: 2815},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 90, col: 32, offset: 2819},
						name: "ClassChar",
					},
				},
			},
		},
		{
			name: "ClassChar",
			pos:  position{line: 91, col: 1, offset: 2829},
			expr: &choiceExpr{
				pos: position{line: 91, col: 13, offset: 2841},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 91, col: 13, offset: 2841},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 91, col: 13, offset: 2841},
								expr: &choiceExpr{
									pos: position{line: 91, col: 16, offset: 2844},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 91, col: 16, offset: 2844},
											val:        "]",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 91, col: 22, offset: 2850},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 91, col: 29, offset: 2857},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 91, col: 35, offset: 2863},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 91, col: 48, offset: 2876},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 91, col: 48, offset: 2876},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 91, col: 53, offset: 2881},
								name: "CharClassEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "CharClassEscape",
			pos:  position{line: 92, col: 1, offset: 2897},
			expr: &choiceExpr{
				pos: position{line: 92, col: 19, offset: 2915},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 92, col: 21, offset: 2917},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 92, col: 21, offset: 2917},
								val:        "]",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 92, col: 27, offset: 2923},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 93, col: 7, offset: 2952},
						run: (*parser).callonCharClassEscape5,
						expr: &seqExpr{
							pos: position{line: 93, col: 7, offset: 2952},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 93, col: 7, offset: 2952},
									expr: &litMatcher{
										pos:        position{line: 93, col: 8, offset: 2953},
										val:        "p",
										ignoreCase: false,
									},
								},
								&choiceExpr{
									pos: position{line: 93, col: 14, offset: 2959},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 93, col: 14, offset: 2959},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 93, col: 27, offset: 2972},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 93, col: 33, offset: 2978},
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
			pos:  position{line: 97, col: 1, offset: 3044},
			expr: &seqExpr{
				pos: position{line: 97, col: 22, offset: 3065},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 97, col: 22, offset: 3065},
						val:        "p",
						ignoreCase: false,
					},
					&choiceExpr{
						pos: position{line: 98, col: 7, offset: 3078},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 98, col: 7, offset: 3078},
								name: "SingleCharUnicodeClass",
							},
							&actionExpr{
								pos: position{line: 99, col: 7, offset: 3107},
								run: (*parser).callonUnicodeClassEscape5,
								expr: &seqExpr{
									pos: position{line: 99, col: 7, offset: 3107},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 99, col: 7, offset: 3107},
											expr: &litMatcher{
												pos:        position{line: 99, col: 8, offset: 3108},
												val:        "{",
												ignoreCase: false,
											},
										},
										&choiceExpr{
											pos: position{line: 99, col: 14, offset: 3114},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 99, col: 14, offset: 3114},
													name: "SourceChar",
												},
												&ruleRefExpr{
													pos:  position{line: 99, col: 27, offset: 3127},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 99, col: 33, offset: 3133},
													name: "EOF",
												},
											},
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 100, col: 7, offset: 3204},
								run: (*parser).callonUnicodeClassEscape13,
								expr: &seqExpr{
									pos: position{line: 100, col: 7, offset: 3204},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 100, col: 7, offset: 3204},
											val:        "{",
											ignoreCase: false,
										},
										&labeledExpr{
											pos:   position{line: 100, col: 11, offset: 3208},
											label: "ident",
											expr: &ruleRefExpr{
												pos:  position{line: 100, col: 17, offset: 3214},
												name: "IdentifierName",
											},
										},
										&litMatcher{
											pos:        position{line: 100, col: 32, offset: 3229},
											val:        "}",
											ignoreCase: false,
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 106, col: 7, offset: 3393},
								run: (*parser).callonUnicodeClassEscape19,
								expr: &seqExpr{
									pos: position{line: 106, col: 7, offset: 3393},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 106, col: 7, offset: 3393},
											val:        "{",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 11, offset: 3397},
											name: "IdentifierName",
										},
										&choiceExpr{
											pos: position{line: 106, col: 28, offset: 3414},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 106, col: 28, offset: 3414},
													val:        "]",
													ignoreCase: false,
												},
												&ruleRefExpr{
													pos:  position{line: 106, col: 34, offset: 3420},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 106, col: 40, offset: 3426},
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
			pos:  position{line: 111, col: 1, offset: 3506},
			expr: &charClassMatcher{
				pos:        position{line: 111, col: 26, offset: 3531},
				val:        "[LMNCPZS]",
				chars:      []rune{'L', 'M', 'N', 'C', 'P', 'Z', 'S'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Number",
			pos:  position{line: 114, col: 1, offset: 3543},
			expr: &actionExpr{
				pos: position{line: 114, col: 10, offset: 3552},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 114, col: 10, offset: 3552},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 114, col: 10, offset: 3552},
							expr: &litMatcher{
								pos:        position{line: 114, col: 10, offset: 3552},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 114, col: 15, offset: 3557},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 114, col: 23, offset: 3565},
							expr: &seqExpr{
								pos: position{line: 114, col: 25, offset: 3567},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 114, col: 25, offset: 3567},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 114, col: 29, offset: 3571},
										expr: &ruleRefExpr{
											pos:  position{line: 114, col: 29, offset: 3571},
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
			pos:  position{line: 118, col: 1, offset: 3623},
			expr: &choiceExpr{
				pos: position{line: 118, col: 11, offset: 3633},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 118, col: 11, offset: 3633},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 118, col: 17, offset: 3639},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 118, col: 17, offset: 3639},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 118, col: 17, offset: 3639},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 118, col: 30, offset: 3652},
									expr: &ruleRefExpr{
										pos:  position{line: 118, col: 30, offset: 3652},
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
			pos:  position{line: 122, col: 1, offset: 3716},
			expr: &charClassMatcher{
				pos:        position{line: 122, col: 16, offset: 3731},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 123, col: 1, offset: 3737},
			expr: &charClassMatcher{
				pos:        position{line: 123, col: 9, offset: 3745},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "CodeBlock",
			pos:  position{line: 125, col: 1, offset: 3752},
			expr: &choiceExpr{
				pos: position{line: 125, col: 13, offset: 3764},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 125, col: 13, offset: 3764},
						run: (*parser).callonCodeBlock2,
						expr: &seqExpr{
							pos: position{line: 125, col: 13, offset: 3764},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 125, col: 13, offset: 3764},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 125, col: 17, offset: 3768},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 125, col: 22, offset: 3773},
									val:        "}",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 127, col: 5, offset: 3814},
						run: (*parser).callonCodeBlock7,
						expr: &seqExpr{
							pos: position{line: 127, col: 5, offset: 3814},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 127, col: 5, offset: 3814},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 127, col: 9, offset: 3818},
									name: "Code",
								},
								&ruleRefExpr{
									pos:  position{line: 127, col: 14, offset: 3823},
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
			pos:  position{line: 131, col: 1, offset: 3888},
			expr: &actionExpr{
				pos: position{line: 131, col: 19, offset: 3906},
				run: (*parser).callonNanoSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 131, col: 19, offset: 3906},
					val:        "ns",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 136, col: 1, offset: 4011},
			expr: &actionExpr{
				pos: position{line: 136, col: 20, offset: 4030},
				run: (*parser).callonMicroSecondUnits1,
				expr: &choiceExpr{
					pos: position{line: 136, col: 21, offset: 4031},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 136, col: 21, offset: 4031},
							val:        "us",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 136, col: 28, offset: 4038},
							val:        "µs",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 136, col: 35, offset: 4046},
							val:        "μs",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 141, col: 1, offset: 4155},
			expr: &actionExpr{
				pos: position{line: 141, col: 20, offset: 4174},
				run: (*parser).callonMilliSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 141, col: 20, offset: 4174},
					val:        "ms",
					ignoreCase: false,
				},
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 146, col: 1, offset: 4281},
			expr: &actionExpr{
				pos: position{line: 146, col: 15, offset: 4295},
				run: (*parser).callonSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 146, col: 15, offset: 4295},
					val:        "s",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 150, col: 1, offset: 4332},
			expr: &actionExpr{
				pos: position{line: 150, col: 15, offset: 4346},
				run: (*parser).callonMinuteUnits1,
				expr: &litMatcher{
					pos:        position{line: 150, col: 15, offset: 4346},
					val:        "m",
					ignoreCase: false,
				},
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 154, col: 1, offset: 4383},
			expr: &actionExpr{
				pos: position{line: 154, col: 13, offset: 4395},
				run: (*parser).callonHourUnits1,
				expr: &litMatcher{
					pos:        position{line: 154, col: 13, offset: 4395},
					val:        "h",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DayUnits",
			pos:  position{line: 158, col: 1, offset: 4430},
			expr: &actionExpr{
				pos: position{line: 158, col: 12, offset: 4441},
				run: (*parser).callonDayUnits1,
				expr: &litMatcher{
					pos:        position{line: 158, col: 12, offset: 4441},
					val:        "d",
					ignoreCase: false,
				},
			},
		},
		{
			name: "WeekUnits",
			pos:  position{line: 164, col: 1, offset: 4649},
			expr: &actionExpr{
				pos: position{line: 164, col: 13, offset: 4661},
				run: (*parser).callonWeekUnits1,
				expr: &litMatcher{
					pos:        position{line: 164, col: 13, offset: 4661},
					val:        "w",
					ignoreCase: false,
				},
			},
		},
		{
			name: "YearUnits",
			pos:  position{line: 170, col: 1, offset: 4872},
			expr: &actionExpr{
				pos: position{line: 170, col: 13, offset: 4884},
				run: (*parser).callonYearUnits1,
				expr: &litMatcher{
					pos:        position{line: 170, col: 13, offset: 4884},
					val:        "y",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 176, col: 1, offset: 5081},
			expr: &choiceExpr{
				pos: position{line: 176, col: 18, offset: 5098},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 176, col: 18, offset: 5098},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 36, offset: 5116},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 55, offset: 5135},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 74, offset: 5154},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 88, offset: 5168},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 102, offset: 5182},
						name: "HourUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 114, offset: 5194},
						name: "DayUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 125, offset: 5205},
						name: "WeekUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 176, col: 137, offset: 5217},
						name: "YearUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 178, col: 1, offset: 5229},
			expr: &actionExpr{
				pos: position{line: 178, col: 12, offset: 5240},
				run: (*parser).callonDuration1,
				expr: &seqExpr{
					pos: position{line: 178, col: 12, offset: 5240},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 178, col: 12, offset: 5240},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 16, offset: 5244},
								name: "Integer",
							},
						},
						&labeledExpr{
							pos:   position{line: 178, col: 24, offset: 5252},
							label: "units",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 30, offset: 5258},
								name: "DurationUnits",
							},
						},
					},
				},
			},
		},
		{
			name: "Code",
			pos:  position{line: 184, col: 1, offset: 5407},
			expr: &zeroOrMoreExpr{
				pos: position{line: 184, col: 8, offset: 5414},
				expr: &choiceExpr{
					pos: position{line: 184, col: 10, offset: 5416},
					alternatives: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 184, col: 10, offset: 5416},
							expr: &seqExpr{
								pos: position{line: 184, col: 12, offset: 5418},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 184, col: 12, offset: 5418},
										expr: &charClassMatcher{
											pos:        position{line: 184, col: 13, offset: 5419},
											val:        "[{}]",
											chars:      []rune{'{', '}'},
											ignoreCase: false,
											inverted:   false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 184, col: 18, offset: 5424},
										name: "LabelMatches",
									},
								},
							},
						},
						&seqExpr{
							pos: position{line: 184, col: 36, offset: 5442},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 184, col: 36, offset: 5442},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 184, col: 40, offset: 5446},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 184, col: 45, offset: 5451},
									val:        "}",
									ignoreCase: false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Operators",
			pos:  position{line: 186, col: 1, offset: 5459},
			expr: &choiceExpr{
				pos: position{line: 186, col: 13, offset: 5471},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 186, col: 13, offset: 5471},
						val:        "-",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 19, offset: 5477},
						val:        "+",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 25, offset: 5483},
						val:        "*",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 31, offset: 5489},
						val:        "%",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 37, offset: 5495},
						val:        "/",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 43, offset: 5501},
						val:        "==",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 50, offset: 5508},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 57, offset: 5515},
						val:        "<=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 64, offset: 5522},
						val:        "<",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 70, offset: 5528},
						val:        ">=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 77, offset: 5535},
						val:        ">",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 83, offset: 5541},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 90, offset: 5548},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 97, offset: 5555},
						val:        "^",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 186, col: 103, offset: 5561},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "LabelOperators",
			pos:  position{line: 188, col: 1, offset: 5566},
			expr: &choiceExpr{
				pos: position{line: 188, col: 19, offset: 5584},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 188, col: 19, offset: 5584},
						run: (*parser).callonLabelOperators2,
						expr: &litMatcher{
							pos:        position{line: 188, col: 19, offset: 5584},
							val:        "!=",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 190, col: 5, offset: 5620},
						run: (*parser).callonLabelOperators4,
						expr: &litMatcher{
							pos:        position{line: 190, col: 5, offset: 5620},
							val:        "=~",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 192, col: 5, offset: 5658},
						run: (*parser).callonLabelOperators6,
						expr: &litMatcher{
							pos:        position{line: 192, col: 5, offset: 5658},
							val:        "!~",
							ignoreCase: false,
						},
					},
					&actionExpr{
						pos: position{line: 194, col: 5, offset: 5698},
						run: (*parser).callonLabelOperators8,
						expr: &litMatcher{
							pos:        position{line: 194, col: 5, offset: 5698},
							val:        "=",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Label",
			pos:  position{line: 198, col: 1, offset: 5729},
			expr: &ruleRefExpr{
				pos:  position{line: 198, col: 9, offset: 5737},
				name: "Identifier",
			},
		},
		{
			name: "LabelMatch",
			pos:  position{line: 199, col: 1, offset: 5748},
			expr: &actionExpr{
				pos: position{line: 199, col: 14, offset: 5761},
				run: (*parser).callonLabelMatch1,
				expr: &seqExpr{
					pos: position{line: 199, col: 14, offset: 5761},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 199, col: 14, offset: 5761},
							label: "label",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 20, offset: 5767},
								name: "Label",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 26, offset: 5773},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 29, offset: 5776},
							label: "op",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 32, offset: 5779},
								name: "LabelOperators",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 47, offset: 5794},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 50, offset: 5797},
							label: "match",
							expr: &choiceExpr{
								pos: position{line: 199, col: 58, offset: 5805},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 199, col: 58, offset: 5805},
										name: "StringLiteral",
									},
									&ruleRefExpr{
										pos:  position{line: 199, col: 74, offset: 5821},
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
			pos:  position{line: 202, col: 1, offset: 5911},
			expr: &actionExpr{
				pos: position{line: 202, col: 16, offset: 5926},
				run: (*parser).callonLabelMatches1,
				expr: &seqExpr{
					pos: position{line: 202, col: 16, offset: 5926},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 202, col: 16, offset: 5926},
							label: "first",
							expr: &ruleRefExpr{
								pos:  position{line: 202, col: 22, offset: 5932},
								name: "LabelMatch",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 202, col: 33, offset: 5943},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 202, col: 36, offset: 5946},
							label: "rest",
							expr: &zeroOrMoreExpr{
								pos: position{line: 202, col: 41, offset: 5951},
								expr: &ruleRefExpr{
									pos:  position{line: 202, col: 41, offset: 5951},
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
			pos:  position{line: 206, col: 1, offset: 6030},
			expr: &actionExpr{
				pos: position{line: 206, col: 21, offset: 6050},
				run: (*parser).callonLabelMatchesRest1,
				expr: &seqExpr{
					pos: position{line: 206, col: 21, offset: 6050},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 206, col: 21, offset: 6050},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 206, col: 25, offset: 6054},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 206, col: 28, offset: 6057},
							label: "match",
							expr: &ruleRefExpr{
								pos:  position{line: 206, col: 34, offset: 6063},
								name: "LabelMatch",
							},
						},
					},
				},
			},
		},
		{
			name: "LabelList",
			pos:  position{line: 210, col: 1, offset: 6101},
			expr: &choiceExpr{
				pos: position{line: 210, col: 13, offset: 6113},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 210, col: 14, offset: 6114},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 210, col: 14, offset: 6114},
								val:        "(",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 210, col: 18, offset: 6118},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 210, col: 21, offset: 6121},
								val:        ")",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 210, col: 29, offset: 6129},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 210, col: 29, offset: 6129},
								val:        "(",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 210, col: 33, offset: 6133},
								name: "__",
							},
							&labeledExpr{
								pos:   position{line: 210, col: 36, offset: 6136},
								label: "label",
								expr: &ruleRefExpr{
									pos:  position{line: 210, col: 42, offset: 6142},
									name: "Label",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 210, col: 48, offset: 6148},
								name: "__",
							},
							&labeledExpr{
								pos:   position{line: 210, col: 51, offset: 6151},
								label: "rest",
								expr: &zeroOrMoreExpr{
									pos: position{line: 210, col: 56, offset: 6156},
									expr: &seqExpr{
										pos: position{line: 210, col: 58, offset: 6158},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 210, col: 58, offset: 6158},
												val:        ",",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 210, col: 62, offset: 6162},
												name: "__",
											},
											&ruleRefExpr{
												pos:  position{line: 210, col: 65, offset: 6165},
												name: "Label",
											},
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 210, col: 74, offset: 6174},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 210, col: 77, offset: 6177},
								val:        ")",
								ignoreCase: false,
							},
						},
					},
				},
			},
		},
		{
			name: "VectorSelector",
			pos:  position{line: 212, col: 1, offset: 6183},
			expr: &actionExpr{
				pos: position{line: 212, col: 18, offset: 6200},
				run: (*parser).callonVectorSelector1,
				expr: &seqExpr{
					pos: position{line: 212, col: 18, offset: 6200},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 212, col: 18, offset: 6200},
							label: "metric",
							expr: &ruleRefExpr{
								pos:  position{line: 212, col: 25, offset: 6207},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 212, col: 36, offset: 6218},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 212, col: 40, offset: 6222},
							label: "block",
							expr: &zeroOrOneExpr{
								pos: position{line: 212, col: 46, offset: 6228},
								expr: &ruleRefExpr{
									pos:  position{line: 212, col: 46, offset: 6228},
									name: "CodeBlock",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 212, col: 57, offset: 6239},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 212, col: 60, offset: 6242},
							label: "rng",
							expr: &zeroOrOneExpr{
								pos: position{line: 212, col: 64, offset: 6246},
								expr: &ruleRefExpr{
									pos:  position{line: 212, col: 64, offset: 6246},
									name: "Range",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 212, col: 71, offset: 6253},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 212, col: 74, offset: 6256},
							label: "offset",
							expr: &zeroOrOneExpr{
								pos: position{line: 212, col: 81, offset: 6263},
								expr: &ruleRefExpr{
									pos:  position{line: 212, col: 81, offset: 6263},
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
			pos:  position{line: 216, col: 1, offset: 6341},
			expr: &actionExpr{
				pos: position{line: 216, col: 9, offset: 6349},
				run: (*parser).callonRange1,
				expr: &seqExpr{
					pos: position{line: 216, col: 9, offset: 6349},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 216, col: 9, offset: 6349},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 216, col: 13, offset: 6353},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 216, col: 16, offset: 6356},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 216, col: 20, offset: 6360},
								name: "Duration",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 216, col: 29, offset: 6369},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 216, col: 32, offset: 6372},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Offset",
			pos:  position{line: 220, col: 1, offset: 6401},
			expr: &actionExpr{
				pos: position{line: 220, col: 10, offset: 6410},
				run: (*parser).callonOffset1,
				expr: &seqExpr{
					pos: position{line: 220, col: 10, offset: 6410},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 220, col: 10, offset: 6410},
							val:        "offset",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 20, offset: 6420},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 23, offset: 6423},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 27, offset: 6427},
								name: "Duration",
							},
						},
					},
				},
			},
		},
		{
			name: "BinaryAggregateOperators",
			pos:  position{line: 224, col: 1, offset: 6461},
			expr: &choiceExpr{
				pos: position{line: 224, col: 28, offset: 6488},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 224, col: 28, offset: 6488},
						val:        "count_values",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 224, col: 46, offset: 6506},
						val:        "topk",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 224, col: 56, offset: 6516},
						val:        "bottomk",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 224, col: 69, offset: 6529},
						val:        "quantile",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "UnaryAggregateOperators",
			pos:  position{line: 225, col: 1, offset: 6541},
			expr: &choiceExpr{
				pos: position{line: 225, col: 27, offset: 6567},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 225, col: 27, offset: 6567},
						val:        "sum",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 36, offset: 6576},
						val:        "min",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 45, offset: 6585},
						val:        "max",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 54, offset: 6594},
						val:        "avg",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 63, offset: 6603},
						val:        "stddev",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 75, offset: 6615},
						val:        "stdvar",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 225, col: 87, offset: 6627},
						val:        "count",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "AggregateOperators",
			pos:  position{line: 226, col: 1, offset: 6636},
			expr: &choiceExpr{
				pos: position{line: 226, col: 22, offset: 6657},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 226, col: 22, offset: 6657},
						name: "BinaryAggregateOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 226, col: 49, offset: 6684},
						name: "UnaryAggregateOperators",
					},
				},
			},
		},
		{
			name: "AggregateBy",
			pos:  position{line: 228, col: 1, offset: 6709},
			expr: &actionExpr{
				pos: position{line: 228, col: 15, offset: 6723},
				run: (*parser).callonAggregateBy1,
				expr: &seqExpr{
					pos: position{line: 228, col: 15, offset: 6723},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 15, offset: 6723},
							val:        "by",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 6729},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 24, offset: 6732},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 31, offset: 6739},
								name: "LabelList",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 41, offset: 6749},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 44, offset: 6752},
							label: "keep",
							expr: &zeroOrOneExpr{
								pos: position{line: 228, col: 49, offset: 6757},
								expr: &litMatcher{
									pos:        position{line: 228, col: 49, offset: 6757},
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
			pos:  position{line: 233, col: 1, offset: 6833},
			expr: &actionExpr{
				pos: position{line: 233, col: 20, offset: 6852},
				run: (*parser).callonAggregateWithout1,
				expr: &seqExpr{
					pos: position{line: 233, col: 20, offset: 6852},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 233, col: 20, offset: 6852},
							val:        "without",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 31, offset: 6863},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 233, col: 34, offset: 6866},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 41, offset: 6873},
								name: "LabelList",
							},
						},
					},
				},
			},
		},
		{
			name: "AggregateGroup",
			pos:  position{line: 237, col: 1, offset: 6911},
			expr: &choiceExpr{
				pos: position{line: 237, col: 18, offset: 6928},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 237, col: 18, offset: 6928},
						name: "AggregateBy",
					},
					&ruleRefExpr{
						pos:  position{line: 237, col: 32, offset: 6942},
						name: "AggregateWithout",
					},
				},
			},
		},
		{
			name: "AggregateExpression",
			pos:  position{line: 239, col: 1, offset: 6960},
			expr: &choiceExpr{
				pos: position{line: 239, col: 23, offset: 6982},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 239, col: 23, offset: 6982},
						run: (*parser).callonAggregateExpression2,
						expr: &seqExpr{
							pos: position{line: 239, col: 23, offset: 6982},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 239, col: 23, offset: 6982},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 239, col: 26, offset: 6985},
										name: "AggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 45, offset: 7004},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 239, col: 48, offset: 7007},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 52, offset: 7011},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 239, col: 55, offset: 7014},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 239, col: 62, offset: 7021},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 77, offset: 7036},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 239, col: 80, offset: 7039},
									val:        ")",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 84, offset: 7043},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 239, col: 87, offset: 7046},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 239, col: 93, offset: 7052},
										expr: &ruleRefExpr{
											pos:  position{line: 239, col: 93, offset: 7052},
											name: "AggregateGroup",
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 241, col: 5, offset: 7097},
						run: (*parser).callonAggregateExpression17,
						expr: &seqExpr{
							pos: position{line: 241, col: 5, offset: 7097},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 241, col: 5, offset: 7097},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 241, col: 8, offset: 7100},
										name: "AggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 241, col: 27, offset: 7119},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 241, col: 30, offset: 7122},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 241, col: 36, offset: 7128},
										expr: &ruleRefExpr{
											pos:  position{line: 241, col: 36, offset: 7128},
											name: "AggregateGroup",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 241, col: 52, offset: 7144},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 241, col: 55, offset: 7147},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 241, col: 59, offset: 7151},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 241, col: 62, offset: 7154},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 241, col: 69, offset: 7161},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 241, col: 84, offset: 7176},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 241, col: 87, offset: 7179},
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
			pos:  position{line: 245, col: 1, offset: 7211},
			expr: &zeroOrMoreExpr{
				pos: position{line: 245, col: 6, offset: 7216},
				expr: &choiceExpr{
					pos: position{line: 245, col: 8, offset: 7218},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 245, col: 8, offset: 7218},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 21, offset: 7231},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 27, offset: 7237},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 246, col: 1, offset: 7248},
			expr: &zeroOrMoreExpr{
				pos: position{line: 246, col: 5, offset: 7252},
				expr: &ruleRefExpr{
					pos:  position{line: 246, col: 5, offset: 7252},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 248, col: 1, offset: 7265},
			expr: &charClassMatcher{
				pos:        position{line: 248, col: 14, offset: 7278},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 249, col: 1, offset: 7286},
			expr: &litMatcher{
				pos:        position{line: 249, col: 7, offset: 7292},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 250, col: 1, offset: 7297},
			expr: &choiceExpr{
				pos: position{line: 250, col: 7, offset: 7303},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 250, col: 7, offset: 7303},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 250, col: 7, offset: 7303},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 250, col: 10, offset: 7306},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 250, col: 16, offset: 7312},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 250, col: 16, offset: 7312},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 250, col: 18, offset: 7314},
								expr: &ruleRefExpr{
									pos:  position{line: 250, col: 18, offset: 7314},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 250, col: 37, offset: 7333},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 250, col: 43, offset: 7339},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 250, col: 43, offset: 7339},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 250, col: 46, offset: 7342},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 252, col: 1, offset: 7347},
			expr: &notExpr{
				pos: position{line: 252, col: 7, offset: 7353},
				expr: &anyMatcher{
					line: 252, col: 8, offset: 7354,
				},
			},
		},
	},
}

func (c *current) onGrammar1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1()
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

func (c *current) onCodeBlock2() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonCodeBlock2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCodeBlock2()
}

func (c *current) onCodeBlock7() (interface{}, error) {
	return nil, errors.New("code block not terminated")
}

func (p *parser) callonCodeBlock7() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCodeBlock7()
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

func (c *current) onAggregateBy1(labels, keep interface{}) (interface{}, error) {
	// TODO: handle keep_common
	return labels, nil
}

func (p *parser) callonAggregateBy1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateBy1(stack["labels"], stack["keep"])
}

func (c *current) onAggregateWithout1(labels interface{}) (interface{}, error) {
	return labels, nil
}

func (p *parser) callonAggregateWithout1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateWithout1(stack["labels"])
}

func (c *current) onAggregateExpression2(op, vector, group interface{}) (interface{}, error) {
	return vector, nil
}

func (p *parser) callonAggregateExpression2() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression2(stack["op"], stack["vector"], stack["group"])
}

func (c *current) onAggregateExpression17(op, group, vector interface{}) (interface{}, error) {
	return vector, nil
}

func (p *parser) callonAggregateExpression17() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAggregateExpression17(stack["op"], stack["group"], stack["vector"])
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
