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

// validateUnicodeEscape checks that the provided escape sequence is a
// valid Unicode escape sequence.
func validateUnicodeEscape(escape, errMsg string) (interface{}, error) {
	r, _, _, err := strconv.UnquoteChar("\\"+escape, '"')
	if err != nil {
		return nil, errors.New(errMsg)
	}
	if 0xD800 <= r && r <= 0xDFFF {
		return nil, errors.New(errMsg)
	}
	return nil, nil
}

var unicodeClasses = map[string]bool{
	"ASCII_Hex_Digit": true,
	"Arabic":          true,
	"Armenian":        true,
	"Avestan":         true,
	"Balinese":        true,
	"Bamum":           true,
	"Bassa_Vah":       true,
	"Batak":           true,
	"Bengali":         true,
	"Bidi_Control":    true,
	"Bopomofo":        true,
	"Brahmi":          true,
	"Braille":         true,
	"Buginese":        true,
	"Buhid":           true,
	"C":               true,
	"Canadian_Aboriginal":    true,
	"Carian":                 true,
	"Caucasian_Albanian":     true,
	"Cc":                     true,
	"Cf":                     true,
	"Chakma":                 true,
	"Cham":                   true,
	"Cherokee":               true,
	"Co":                     true,
	"Common":                 true,
	"Coptic":                 true,
	"Cs":                     true,
	"Cuneiform":              true,
	"Cypriot":                true,
	"Cyrillic":               true,
	"Dash":                   true,
	"Deprecated":             true,
	"Deseret":                true,
	"Devanagari":             true,
	"Diacritic":              true,
	"Duployan":               true,
	"Egyptian_Hieroglyphs":   true,
	"Elbasan":                true,
	"Ethiopic":               true,
	"Extender":               true,
	"Georgian":               true,
	"Glagolitic":             true,
	"Gothic":                 true,
	"Grantha":                true,
	"Greek":                  true,
	"Gujarati":               true,
	"Gurmukhi":               true,
	"Han":                    true,
	"Hangul":                 true,
	"Hanunoo":                true,
	"Hebrew":                 true,
	"Hex_Digit":              true,
	"Hiragana":               true,
	"Hyphen":                 true,
	"IDS_Binary_Operator":    true,
	"IDS_Trinary_Operator":   true,
	"Ideographic":            true,
	"Imperial_Aramaic":       true,
	"Inherited":              true,
	"Inscriptional_Pahlavi":  true,
	"Inscriptional_Parthian": true,
	"Javanese":               true,
	"Join_Control":           true,
	"Kaithi":                 true,
	"Kannada":                true,
	"Katakana":               true,
	"Kayah_Li":               true,
	"Kharoshthi":             true,
	"Khmer":                  true,
	"Khojki":                 true,
	"Khudawadi":              true,
	"L":                      true,
	"Lao":                    true,
	"Latin":                  true,
	"Lepcha":                 true,
	"Limbu":                  true,
	"Linear_A":               true,
	"Linear_B":               true,
	"Lisu":                   true,
	"Ll":                     true,
	"Lm":                     true,
	"Lo":                     true,
	"Logical_Order_Exception": true,
	"Lt":                   true,
	"Lu":                   true,
	"Lycian":               true,
	"Lydian":               true,
	"M":                    true,
	"Mahajani":             true,
	"Malayalam":            true,
	"Mandaic":              true,
	"Manichaean":           true,
	"Mc":                   true,
	"Me":                   true,
	"Meetei_Mayek":         true,
	"Mende_Kikakui":        true,
	"Meroitic_Cursive":     true,
	"Meroitic_Hieroglyphs": true,
	"Miao":                 true,
	"Mn":                   true,
	"Modi":                 true,
	"Mongolian":            true,
	"Mro":                  true,
	"Myanmar":              true,
	"N":                    true,
	"Nabataean":            true,
	"Nd":                   true,
	"New_Tai_Lue":          true,
	"Nko":                  true,
	"Nl":                   true,
	"No":                   true,
	"Noncharacter_Code_Point":            true,
	"Ogham":                              true,
	"Ol_Chiki":                           true,
	"Old_Italic":                         true,
	"Old_North_Arabian":                  true,
	"Old_Permic":                         true,
	"Old_Persian":                        true,
	"Old_South_Arabian":                  true,
	"Old_Turkic":                         true,
	"Oriya":                              true,
	"Osmanya":                            true,
	"Other_Alphabetic":                   true,
	"Other_Default_Ignorable_Code_Point": true,
	"Other_Grapheme_Extend":              true,
	"Other_ID_Continue":                  true,
	"Other_ID_Start":                     true,
	"Other_Lowercase":                    true,
	"Other_Math":                         true,
	"Other_Uppercase":                    true,
	"P":                                  true,
	"Pahawh_Hmong":                       true,
	"Palmyrene":                          true,
	"Pattern_Syntax":                     true,
	"Pattern_White_Space":                true,
	"Pau_Cin_Hau":                        true,
	"Pc":                                 true,
	"Pd":                                 true,
	"Pe":                                 true,
	"Pf":                                 true,
	"Phags_Pa":                           true,
	"Phoenician":                         true,
	"Pi":                                 true,
	"Po":                                 true,
	"Ps":                                 true,
	"Psalter_Pahlavi":                    true,
	"Quotation_Mark":                     true,
	"Radical":                            true,
	"Rejang":                             true,
	"Runic":                              true,
	"S":                                  true,
	"STerm":                              true,
	"Samaritan":                          true,
	"Saurashtra":                         true,
	"Sc":                                 true,
	"Sharada":                            true,
	"Shavian":                            true,
	"Siddham":                            true,
	"Sinhala":                            true,
	"Sk":                                 true,
	"Sm":                                 true,
	"So":                                 true,
	"Soft_Dotted":                        true,
	"Sora_Sompeng":                       true,
	"Sundanese":                          true,
	"Syloti_Nagri":                       true,
	"Syriac":                             true,
	"Tagalog":                            true,
	"Tagbanwa":                           true,
	"Tai_Le":                             true,
	"Tai_Tham":                           true,
	"Tai_Viet":                           true,
	"Takri":                              true,
	"Tamil":                              true,
	"Telugu":                             true,
	"Terminal_Punctuation":               true,
	"Thaana":                             true,
	"Thai":                               true,
	"Tibetan":                            true,
	"Tifinagh":                           true,
	"Tirhuta":                            true,
	"Ugaritic":                           true,
	"Unified_Ideograph":                  true,
	"Vai":                                true,
	"Variation_Selector":                 true,
	"Warang_Citi":                        true,
	"White_Space":                        true,
	"Yi":                                 true,
	"Z":                                  true,
	"Zl":                                 true,
	"Zp":                                 true,
	"Zs":                                 true,
}

var reservedWords = map[string]bool{}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 219, col: 1, offset: 7696},
			expr: &actionExpr{
				pos: position{line: 219, col: 12, offset: 7707},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 219, col: 12, offset: 7707},
					exprs: []interface{}{
						&choiceExpr{
							pos: position{line: 219, col: 14, offset: 7709},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 219, col: 14, offset: 7709},
									name: "Comment",
								},
								&ruleRefExpr{
									pos:  position{line: 219, col: 24, offset: 7719},
									name: "AggregateExpression",
								},
								&ruleRefExpr{
									pos:  position{line: 219, col: 46, offset: 7741},
									name: "VectorSelector",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 63, offset: 7758},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 223, col: 1, offset: 7798},
			expr: &anyMatcher{
				line: 223, col: 14, offset: 7811,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 225, col: 1, offset: 7814},
			expr: &seqExpr{
				pos: position{line: 225, col: 11, offset: 7824},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 225, col: 11, offset: 7824},
						val:        "#",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 225, col: 15, offset: 7828},
						expr: &seqExpr{
							pos: position{line: 225, col: 17, offset: 7830},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 225, col: 17, offset: 7830},
									expr: &ruleRefExpr{
										pos:  position{line: 225, col: 18, offset: 7831},
										name: "EOL",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 225, col: 22, offset: 7835},
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
			pos:  position{line: 227, col: 1, offset: 7850},
			expr: &actionExpr{
				pos: position{line: 227, col: 14, offset: 7863},
				run: (*parser).callonIdentifier1,
				expr: &labeledExpr{
					pos:   position{line: 227, col: 14, offset: 7863},
					label: "ident",
					expr: &ruleRefExpr{
						pos:  position{line: 227, col: 20, offset: 7869},
						name: "IdentifierName",
					},
				},
			},
		},
		{
			name: "IdentifierName",
			pos:  position{line: 235, col: 1, offset: 8031},
			expr: &actionExpr{
				pos: position{line: 235, col: 18, offset: 8048},
				run: (*parser).callonIdentifierName1,
				expr: &seqExpr{
					pos: position{line: 235, col: 18, offset: 8048},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 235, col: 18, offset: 8048},
							name: "IdentifierStart",
						},
						&zeroOrMoreExpr{
							pos: position{line: 235, col: 34, offset: 8064},
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 34, offset: 8064},
								name: "IdentifierPart",
							},
						},
					},
				},
			},
		},
		{
			name: "IdentifierStart",
			pos:  position{line: 238, col: 1, offset: 8115},
			expr: &charClassMatcher{
				pos:        position{line: 238, col: 19, offset: 8133},
				val:        "[\\pL_]",
				chars:      []rune{'_'},
				classes:    []*unicode.RangeTable{rangeTable("L")},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "IdentifierPart",
			pos:  position{line: 239, col: 1, offset: 8140},
			expr: &choiceExpr{
				pos: position{line: 239, col: 18, offset: 8157},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 239, col: 18, offset: 8157},
						name: "IdentifierStart",
					},
					&charClassMatcher{
						pos:        position{line: 239, col: 36, offset: 8175},
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
			pos:  position{line: 241, col: 1, offset: 8185},
			expr: &choiceExpr{
				pos: position{line: 241, col: 17, offset: 8201},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 241, col: 17, offset: 8201},
						run: (*parser).callonStringLiteral2,
						expr: &choiceExpr{
							pos: position{line: 241, col: 19, offset: 8203},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 241, col: 19, offset: 8203},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 19, offset: 8203},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 241, col: 23, offset: 8207},
											expr: &ruleRefExpr{
												pos:  position{line: 241, col: 23, offset: 8207},
												name: "DoubleStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 241, col: 41, offset: 8225},
											val:        "\"",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 241, col: 47, offset: 8231},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 47, offset: 8231},
											val:        "'",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 241, col: 51, offset: 8235},
											name: "SingleStringChar",
										},
										&litMatcher{
											pos:        position{line: 241, col: 68, offset: 8252},
											val:        "'",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 241, col: 74, offset: 8258},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 74, offset: 8258},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 241, col: 78, offset: 8262},
											expr: &ruleRefExpr{
												pos:  position{line: 241, col: 78, offset: 8262},
												name: "RawStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 241, col: 93, offset: 8277},
											val:        "`",
											ignoreCase: false,
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 243, col: 5, offset: 8320},
						run: (*parser).callonStringLiteral18,
						expr: &choiceExpr{
							pos: position{line: 243, col: 7, offset: 8322},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 243, col: 9, offset: 8324},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 9, offset: 8324},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 243, col: 13, offset: 8328},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 13, offset: 8328},
												name: "DoubleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 243, col: 33, offset: 8348},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 243, col: 33, offset: 8348},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 243, col: 39, offset: 8354},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 243, col: 51, offset: 8366},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 51, offset: 8366},
											val:        "'",
											ignoreCase: false,
										},
										&zeroOrOneExpr{
											pos: position{line: 243, col: 55, offset: 8370},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 55, offset: 8370},
												name: "SingleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 243, col: 75, offset: 8390},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 243, col: 75, offset: 8390},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 243, col: 81, offset: 8396},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 243, col: 91, offset: 8406},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 91, offset: 8406},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 243, col: 95, offset: 8410},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 95, offset: 8410},
												name: "RawStringChar",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 243, col: 110, offset: 8425},
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
			pos:  position{line: 247, col: 1, offset: 8496},
			expr: &choiceExpr{
				pos: position{line: 247, col: 20, offset: 8515},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 247, col: 20, offset: 8515},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 247, col: 20, offset: 8515},
								expr: &choiceExpr{
									pos: position{line: 247, col: 23, offset: 8518},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 247, col: 23, offset: 8518},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 247, col: 29, offset: 8524},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 247, col: 36, offset: 8531},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 247, col: 42, offset: 8537},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 247, col: 55, offset: 8550},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 247, col: 55, offset: 8550},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 247, col: 60, offset: 8555},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "SingleStringChar",
			pos:  position{line: 248, col: 1, offset: 8574},
			expr: &choiceExpr{
				pos: position{line: 248, col: 20, offset: 8593},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 248, col: 20, offset: 8593},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 248, col: 20, offset: 8593},
								expr: &choiceExpr{
									pos: position{line: 248, col: 23, offset: 8596},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 248, col: 23, offset: 8596},
											val:        "'",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 248, col: 29, offset: 8602},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 36, offset: 8609},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 248, col: 42, offset: 8615},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 248, col: 55, offset: 8628},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 248, col: 55, offset: 8628},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 248, col: 60, offset: 8633},
								name: "SingleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "RawStringChar",
			pos:  position{line: 249, col: 1, offset: 8652},
			expr: &seqExpr{
				pos: position{line: 249, col: 17, offset: 8668},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 249, col: 17, offset: 8668},
						expr: &litMatcher{
							pos:        position{line: 249, col: 18, offset: 8669},
							val:        "`",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 249, col: 22, offset: 8673},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 251, col: 1, offset: 8685},
			expr: &choiceExpr{
				pos: position{line: 251, col: 22, offset: 8706},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 251, col: 24, offset: 8708},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 251, col: 24, offset: 8708},
								val:        "\"",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 251, col: 30, offset: 8714},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 252, col: 7, offset: 8743},
						run: (*parser).callonDoubleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 252, col: 9, offset: 8745},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 252, col: 9, offset: 8745},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 252, col: 22, offset: 8758},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 252, col: 28, offset: 8764},
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
			pos:  position{line: 255, col: 1, offset: 8829},
			expr: &choiceExpr{
				pos: position{line: 255, col: 22, offset: 8850},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 255, col: 24, offset: 8852},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 255, col: 24, offset: 8852},
								val:        "'",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 255, col: 30, offset: 8858},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 256, col: 7, offset: 8887},
						run: (*parser).callonSingleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 256, col: 9, offset: 8889},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 256, col: 9, offset: 8889},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 256, col: 22, offset: 8902},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 256, col: 28, offset: 8908},
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
			pos:  position{line: 260, col: 1, offset: 8974},
			expr: &choiceExpr{
				pos: position{line: 260, col: 24, offset: 8997},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 260, col: 24, offset: 8997},
						name: "SingleCharEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 43, offset: 9016},
						name: "OctalEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 57, offset: 9030},
						name: "HexEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 69, offset: 9042},
						name: "LongUnicodeEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 89, offset: 9062},
						name: "ShortUnicodeEscape",
					},
				},
			},
		},
		{
			name: "SingleCharEscape",
			pos:  position{line: 261, col: 1, offset: 9081},
			expr: &choiceExpr{
				pos: position{line: 261, col: 20, offset: 9100},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 261, col: 20, offset: 9100},
						val:        "a",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 26, offset: 9106},
						val:        "b",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 32, offset: 9112},
						val:        "n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 38, offset: 9118},
						val:        "f",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 44, offset: 9124},
						val:        "r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 50, offset: 9130},
						val:        "t",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 56, offset: 9136},
						val:        "v",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 62, offset: 9142},
						val:        "\\",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "OctalEscape",
			pos:  position{line: 262, col: 1, offset: 9147},
			expr: &choiceExpr{
				pos: position{line: 262, col: 15, offset: 9161},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 262, col: 15, offset: 9161},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 262, col: 15, offset: 9161},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 262, col: 26, offset: 9172},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 262, col: 37, offset: 9183},
								name: "OctalDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 263, col: 7, offset: 9200},
						run: (*parser).callonOctalEscape6,
						expr: &seqExpr{
							pos: position{line: 263, col: 7, offset: 9200},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 263, col: 7, offset: 9200},
									name: "OctalDigit",
								},
								&choiceExpr{
									pos: position{line: 263, col: 20, offset: 9213},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 263, col: 20, offset: 9213},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 263, col: 33, offset: 9226},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 263, col: 39, offset: 9232},
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
			pos:  position{line: 266, col: 1, offset: 9293},
			expr: &choiceExpr{
				pos: position{line: 266, col: 13, offset: 9305},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 266, col: 13, offset: 9305},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 266, col: 13, offset: 9305},
								val:        "x",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 266, col: 17, offset: 9309},
								name: "HexDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 266, col: 26, offset: 9318},
								name: "HexDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 267, col: 7, offset: 9333},
						run: (*parser).callonHexEscape6,
						expr: &seqExpr{
							pos: position{line: 267, col: 7, offset: 9333},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 267, col: 7, offset: 9333},
									val:        "x",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 267, col: 13, offset: 9339},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 267, col: 13, offset: 9339},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 267, col: 26, offset: 9352},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 267, col: 32, offset: 9358},
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
			pos:  position{line: 270, col: 1, offset: 9425},
			expr: &choiceExpr{
				pos: position{line: 271, col: 5, offset: 9450},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 271, col: 5, offset: 9450},
						run: (*parser).callonLongUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 271, col: 5, offset: 9450},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 271, col: 5, offset: 9450},
									val:        "U",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 9, offset: 9454},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 18, offset: 9463},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 27, offset: 9472},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 36, offset: 9481},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 45, offset: 9490},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 54, offset: 9499},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 63, offset: 9508},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 72, offset: 9517},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 274, col: 7, offset: 9619},
						run: (*parser).callonLongUnicodeEscape13,
						expr: &seqExpr{
							pos: position{line: 274, col: 7, offset: 9619},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 274, col: 7, offset: 9619},
									val:        "U",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 274, col: 13, offset: 9625},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 274, col: 13, offset: 9625},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 274, col: 26, offset: 9638},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 274, col: 32, offset: 9644},
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
			pos:  position{line: 277, col: 1, offset: 9707},
			expr: &choiceExpr{
				pos: position{line: 278, col: 5, offset: 9733},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 278, col: 5, offset: 9733},
						run: (*parser).callonShortUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 278, col: 5, offset: 9733},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 278, col: 5, offset: 9733},
									val:        "u",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 9, offset: 9737},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 18, offset: 9746},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 27, offset: 9755},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 36, offset: 9764},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 281, col: 7, offset: 9866},
						run: (*parser).callonShortUnicodeEscape9,
						expr: &seqExpr{
							pos: position{line: 281, col: 7, offset: 9866},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 281, col: 7, offset: 9866},
									val:        "u",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 281, col: 13, offset: 9872},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 281, col: 13, offset: 9872},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 281, col: 26, offset: 9885},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 281, col: 32, offset: 9891},
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
			pos:  position{line: 285, col: 1, offset: 9955},
			expr: &charClassMatcher{
				pos:        position{line: 285, col: 14, offset: 9968},
				val:        "[0-7]",
				ranges:     []rune{'0', '7'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 286, col: 1, offset: 9974},
			expr: &charClassMatcher{
				pos:        position{line: 286, col: 16, offset: 9989},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "HexDigit",
			pos:  position{line: 287, col: 1, offset: 9995},
			expr: &charClassMatcher{
				pos:        position{line: 287, col: 12, offset: 10006},
				val:        "[0-9a-f]i",
				ranges:     []rune{'0', '9', 'a', 'f'},
				ignoreCase: true,
				inverted:   false,
			},
		},
		{
			name: "CharClassMatcher",
			pos:  position{line: 289, col: 1, offset: 10017},
			expr: &choiceExpr{
				pos: position{line: 289, col: 20, offset: 10036},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 289, col: 20, offset: 10036},
						run: (*parser).callonCharClassMatcher2,
						expr: &seqExpr{
							pos: position{line: 289, col: 20, offset: 10036},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 289, col: 20, offset: 10036},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 289, col: 24, offset: 10040},
									expr: &choiceExpr{
										pos: position{line: 289, col: 26, offset: 10042},
										alternatives: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 289, col: 26, offset: 10042},
												name: "ClassCharRange",
											},
											&ruleRefExpr{
												pos:  position{line: 289, col: 43, offset: 10059},
												name: "ClassChar",
											},
											&seqExpr{
												pos: position{line: 289, col: 55, offset: 10071},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 289, col: 55, offset: 10071},
														val:        "\\",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 289, col: 60, offset: 10076},
														name: "UnicodeClassEscape",
													},
												},
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 289, col: 82, offset: 10098},
									val:        "]",
									ignoreCase: false,
								},
								&zeroOrOneExpr{
									pos: position{line: 289, col: 86, offset: 10102},
									expr: &litMatcher{
										pos:        position{line: 289, col: 86, offset: 10102},
										val:        "i",
										ignoreCase: false,
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 291, col: 5, offset: 10144},
						run: (*parser).callonCharClassMatcher15,
						expr: &seqExpr{
							pos: position{line: 291, col: 5, offset: 10144},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 291, col: 5, offset: 10144},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 291, col: 9, offset: 10148},
									expr: &seqExpr{
										pos: position{line: 291, col: 11, offset: 10150},
										exprs: []interface{}{
											&notExpr{
												pos: position{line: 291, col: 11, offset: 10150},
												expr: &ruleRefExpr{
													pos:  position{line: 291, col: 14, offset: 10153},
													name: "EOL",
												},
											},
											&ruleRefExpr{
												pos:  position{line: 291, col: 20, offset: 10159},
												name: "SourceChar",
											},
										},
									},
								},
								&choiceExpr{
									pos: position{line: 291, col: 36, offset: 10175},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 291, col: 36, offset: 10175},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 291, col: 42, offset: 10181},
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
			pos:  position{line: 295, col: 1, offset: 10253},
			expr: &seqExpr{
				pos: position{line: 295, col: 18, offset: 10270},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 295, col: 18, offset: 10270},
						name: "ClassChar",
					},
					&litMatcher{
						pos:        position{line: 295, col: 28, offset: 10280},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 295, col: 32, offset: 10284},
						name: "ClassChar",
					},
				},
			},
		},
		{
			name: "ClassChar",
			pos:  position{line: 296, col: 1, offset: 10294},
			expr: &choiceExpr{
				pos: position{line: 296, col: 13, offset: 10306},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 296, col: 13, offset: 10306},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 296, col: 13, offset: 10306},
								expr: &choiceExpr{
									pos: position{line: 296, col: 16, offset: 10309},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 296, col: 16, offset: 10309},
											val:        "]",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 296, col: 22, offset: 10315},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 296, col: 29, offset: 10322},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 296, col: 35, offset: 10328},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 296, col: 48, offset: 10341},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 296, col: 48, offset: 10341},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 296, col: 53, offset: 10346},
								name: "CharClassEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "CharClassEscape",
			pos:  position{line: 297, col: 1, offset: 10362},
			expr: &choiceExpr{
				pos: position{line: 297, col: 19, offset: 10380},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 297, col: 21, offset: 10382},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 297, col: 21, offset: 10382},
								val:        "]",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 297, col: 27, offset: 10388},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 298, col: 7, offset: 10417},
						run: (*parser).callonCharClassEscape5,
						expr: &seqExpr{
							pos: position{line: 298, col: 7, offset: 10417},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 298, col: 7, offset: 10417},
									expr: &litMatcher{
										pos:        position{line: 298, col: 8, offset: 10418},
										val:        "p",
										ignoreCase: false,
									},
								},
								&choiceExpr{
									pos: position{line: 298, col: 14, offset: 10424},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 298, col: 14, offset: 10424},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 27, offset: 10437},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 33, offset: 10443},
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
			pos:  position{line: 302, col: 1, offset: 10509},
			expr: &seqExpr{
				pos: position{line: 302, col: 22, offset: 10530},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 302, col: 22, offset: 10530},
						val:        "p",
						ignoreCase: false,
					},
					&choiceExpr{
						pos: position{line: 303, col: 7, offset: 10543},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 303, col: 7, offset: 10543},
								name: "SingleCharUnicodeClass",
							},
							&actionExpr{
								pos: position{line: 304, col: 7, offset: 10572},
								run: (*parser).callonUnicodeClassEscape5,
								expr: &seqExpr{
									pos: position{line: 304, col: 7, offset: 10572},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 304, col: 7, offset: 10572},
											expr: &litMatcher{
												pos:        position{line: 304, col: 8, offset: 10573},
												val:        "{",
												ignoreCase: false,
											},
										},
										&choiceExpr{
											pos: position{line: 304, col: 14, offset: 10579},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 304, col: 14, offset: 10579},
													name: "SourceChar",
												},
												&ruleRefExpr{
													pos:  position{line: 304, col: 27, offset: 10592},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 304, col: 33, offset: 10598},
													name: "EOF",
												},
											},
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 305, col: 7, offset: 10669},
								run: (*parser).callonUnicodeClassEscape13,
								expr: &seqExpr{
									pos: position{line: 305, col: 7, offset: 10669},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 305, col: 7, offset: 10669},
											val:        "{",
											ignoreCase: false,
										},
										&labeledExpr{
											pos:   position{line: 305, col: 11, offset: 10673},
											label: "ident",
											expr: &ruleRefExpr{
												pos:  position{line: 305, col: 17, offset: 10679},
												name: "IdentifierName",
											},
										},
										&litMatcher{
											pos:        position{line: 305, col: 32, offset: 10694},
											val:        "}",
											ignoreCase: false,
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 311, col: 7, offset: 10858},
								run: (*parser).callonUnicodeClassEscape19,
								expr: &seqExpr{
									pos: position{line: 311, col: 7, offset: 10858},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 311, col: 7, offset: 10858},
											val:        "{",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 311, col: 11, offset: 10862},
											name: "IdentifierName",
										},
										&choiceExpr{
											pos: position{line: 311, col: 28, offset: 10879},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 311, col: 28, offset: 10879},
													val:        "]",
													ignoreCase: false,
												},
												&ruleRefExpr{
													pos:  position{line: 311, col: 34, offset: 10885},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 311, col: 40, offset: 10891},
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
			pos:  position{line: 316, col: 1, offset: 10971},
			expr: &charClassMatcher{
				pos:        position{line: 316, col: 26, offset: 10996},
				val:        "[LMNCPZS]",
				chars:      []rune{'L', 'M', 'N', 'C', 'P', 'Z', 'S'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Number",
			pos:  position{line: 319, col: 1, offset: 11008},
			expr: &actionExpr{
				pos: position{line: 319, col: 10, offset: 11017},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 319, col: 10, offset: 11017},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 319, col: 10, offset: 11017},
							expr: &litMatcher{
								pos:        position{line: 319, col: 10, offset: 11017},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 319, col: 15, offset: 11022},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 319, col: 23, offset: 11030},
							expr: &seqExpr{
								pos: position{line: 319, col: 25, offset: 11032},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 319, col: 25, offset: 11032},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 319, col: 29, offset: 11036},
										expr: &ruleRefExpr{
											pos:  position{line: 319, col: 29, offset: 11036},
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
			pos:  position{line: 323, col: 1, offset: 11101},
			expr: &choiceExpr{
				pos: position{line: 323, col: 11, offset: 11111},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 323, col: 11, offset: 11111},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 323, col: 17, offset: 11117},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 323, col: 17, offset: 11117},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 323, col: 17, offset: 11117},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 323, col: 30, offset: 11130},
									expr: &ruleRefExpr{
										pos:  position{line: 323, col: 30, offset: 11130},
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
			pos:  position{line: 327, col: 1, offset: 11194},
			expr: &charClassMatcher{
				pos:        position{line: 327, col: 16, offset: 11209},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 328, col: 1, offset: 11215},
			expr: &charClassMatcher{
				pos:        position{line: 328, col: 9, offset: 11223},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "CodeBlock",
			pos:  position{line: 330, col: 1, offset: 11230},
			expr: &choiceExpr{
				pos: position{line: 330, col: 13, offset: 11242},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 330, col: 13, offset: 11242},
						run: (*parser).callonCodeBlock2,
						expr: &seqExpr{
							pos: position{line: 330, col: 13, offset: 11242},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 330, col: 13, offset: 11242},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 330, col: 17, offset: 11246},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 330, col: 22, offset: 11251},
									val:        "}",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 332, col: 5, offset: 11292},
						run: (*parser).callonCodeBlock7,
						expr: &seqExpr{
							pos: position{line: 332, col: 5, offset: 11292},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 332, col: 5, offset: 11292},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 332, col: 9, offset: 11296},
									name: "Code",
								},
								&ruleRefExpr{
									pos:  position{line: 332, col: 14, offset: 11301},
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
			pos:  position{line: 336, col: 1, offset: 11366},
			expr: &actionExpr{
				pos: position{line: 336, col: 19, offset: 11384},
				run: (*parser).callonNanoSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 336, col: 19, offset: 11384},
					val:        "ns",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 341, col: 1, offset: 11489},
			expr: &actionExpr{
				pos: position{line: 341, col: 20, offset: 11508},
				run: (*parser).callonMicroSecondUnits1,
				expr: &choiceExpr{
					pos: position{line: 341, col: 21, offset: 11509},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 21, offset: 11509},
							val:        "us",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 28, offset: 11516},
							val:        "µs",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 35, offset: 11524},
							val:        "μs",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 346, col: 1, offset: 11633},
			expr: &actionExpr{
				pos: position{line: 346, col: 20, offset: 11652},
				run: (*parser).callonMilliSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 346, col: 20, offset: 11652},
					val:        "ms",
					ignoreCase: false,
				},
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 351, col: 1, offset: 11759},
			expr: &actionExpr{
				pos: position{line: 351, col: 15, offset: 11773},
				run: (*parser).callonSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 351, col: 15, offset: 11773},
					val:        "s",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 355, col: 1, offset: 11810},
			expr: &actionExpr{
				pos: position{line: 355, col: 15, offset: 11824},
				run: (*parser).callonMinuteUnits1,
				expr: &litMatcher{
					pos:        position{line: 355, col: 15, offset: 11824},
					val:        "m",
					ignoreCase: false,
				},
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 359, col: 1, offset: 11861},
			expr: &actionExpr{
				pos: position{line: 359, col: 13, offset: 11873},
				run: (*parser).callonHourUnits1,
				expr: &litMatcher{
					pos:        position{line: 359, col: 13, offset: 11873},
					val:        "h",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DayUnits",
			pos:  position{line: 363, col: 1, offset: 11908},
			expr: &actionExpr{
				pos: position{line: 363, col: 12, offset: 11919},
				run: (*parser).callonDayUnits1,
				expr: &litMatcher{
					pos:        position{line: 363, col: 12, offset: 11919},
					val:        "d",
					ignoreCase: false,
				},
			},
		},
		{
			name: "WeekUnits",
			pos:  position{line: 369, col: 1, offset: 12127},
			expr: &actionExpr{
				pos: position{line: 369, col: 13, offset: 12139},
				run: (*parser).callonWeekUnits1,
				expr: &litMatcher{
					pos:        position{line: 369, col: 13, offset: 12139},
					val:        "w",
					ignoreCase: false,
				},
			},
		},
		{
			name: "YearUnits",
			pos:  position{line: 375, col: 1, offset: 12350},
			expr: &actionExpr{
				pos: position{line: 375, col: 13, offset: 12362},
				run: (*parser).callonYearUnits1,
				expr: &litMatcher{
					pos:        position{line: 375, col: 13, offset: 12362},
					val:        "y",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 381, col: 1, offset: 12559},
			expr: &choiceExpr{
				pos: position{line: 381, col: 18, offset: 12576},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 381, col: 18, offset: 12576},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 36, offset: 12594},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 55, offset: 12613},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 74, offset: 12632},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 88, offset: 12646},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 102, offset: 12660},
						name: "HourUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 114, offset: 12672},
						name: "DayUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 125, offset: 12683},
						name: "WeekUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 137, offset: 12695},
						name: "YearUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 383, col: 1, offset: 12707},
			expr: &actionExpr{
				pos: position{line: 383, col: 12, offset: 12718},
				run: (*parser).callonDuration1,
				expr: &seqExpr{
					pos: position{line: 383, col: 12, offset: 12718},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 383, col: 12, offset: 12718},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 16, offset: 12722},
								name: "Integer",
							},
						},
						&labeledExpr{
							pos:   position{line: 383, col: 24, offset: 12730},
							label: "units",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 30, offset: 12736},
								name: "DurationUnits",
							},
						},
					},
				},
			},
		},
		{
			name: "Code",
			pos:  position{line: 389, col: 1, offset: 12885},
			expr: &zeroOrMoreExpr{
				pos: position{line: 389, col: 8, offset: 12892},
				expr: &choiceExpr{
					pos: position{line: 389, col: 10, offset: 12894},
					alternatives: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 389, col: 10, offset: 12894},
							expr: &seqExpr{
								pos: position{line: 389, col: 12, offset: 12896},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 389, col: 12, offset: 12896},
										expr: &charClassMatcher{
											pos:        position{line: 389, col: 13, offset: 12897},
											val:        "[{}]",
											chars:      []rune{'{', '}'},
											ignoreCase: false,
											inverted:   false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 389, col: 18, offset: 12902},
										name: "LabelMatches",
									},
								},
							},
						},
						&seqExpr{
							pos: position{line: 389, col: 36, offset: 12920},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 389, col: 36, offset: 12920},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 389, col: 40, offset: 12924},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 389, col: 45, offset: 12929},
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
			pos:  position{line: 391, col: 1, offset: 12937},
			expr: &choiceExpr{
				pos: position{line: 391, col: 13, offset: 12949},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 391, col: 13, offset: 12949},
						val:        "-",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 19, offset: 12955},
						val:        "+",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 25, offset: 12961},
						val:        "*",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 31, offset: 12967},
						val:        "%",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 37, offset: 12973},
						val:        "/",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 43, offset: 12979},
						val:        "==",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 50, offset: 12986},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 57, offset: 12993},
						val:        "<=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 64, offset: 13000},
						val:        "<",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 70, offset: 13006},
						val:        ">=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 77, offset: 13013},
						val:        ">",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 83, offset: 13019},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 90, offset: 13026},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 97, offset: 13033},
						val:        "^",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 103, offset: 13039},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "LabelOperators",
			pos:  position{line: 393, col: 1, offset: 13044},
			expr: &choiceExpr{
				pos: position{line: 393, col: 19, offset: 13062},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 393, col: 19, offset: 13062},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 393, col: 26, offset: 13069},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 393, col: 33, offset: 13076},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 393, col: 40, offset: 13083},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "Label",
			pos:  position{line: 394, col: 1, offset: 13087},
			expr: &ruleRefExpr{
				pos:  position{line: 394, col: 9, offset: 13095},
				name: "Identifier",
			},
		},
		{
			name: "LabelMatch",
			pos:  position{line: 395, col: 1, offset: 13106},
			expr: &seqExpr{
				pos: position{line: 395, col: 14, offset: 13119},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 395, col: 14, offset: 13119},
						label: "label",
						expr: &ruleRefExpr{
							pos:  position{line: 395, col: 20, offset: 13125},
							name: "Label",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 395, col: 26, offset: 13131},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 395, col: 29, offset: 13134},
						label: "op",
						expr: &ruleRefExpr{
							pos:  position{line: 395, col: 32, offset: 13137},
							name: "LabelOperators",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 395, col: 47, offset: 13152},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 395, col: 50, offset: 13155},
						label: "match",
						expr: &choiceExpr{
							pos: position{line: 395, col: 58, offset: 13163},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 395, col: 58, offset: 13163},
									name: "StringLiteral",
								},
								&ruleRefExpr{
									pos:  position{line: 395, col: 74, offset: 13179},
									name: "Number",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "LabelMatches",
			pos:  position{line: 396, col: 1, offset: 13188},
			expr: &seqExpr{
				pos: position{line: 396, col: 16, offset: 13203},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 396, col: 16, offset: 13203},
						name: "LabelMatch",
					},
					&ruleRefExpr{
						pos:  position{line: 396, col: 27, offset: 13214},
						name: "__",
					},
					&zeroOrMoreExpr{
						pos: position{line: 396, col: 30, offset: 13217},
						expr: &seqExpr{
							pos: position{line: 396, col: 32, offset: 13219},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 396, col: 32, offset: 13219},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 396, col: 36, offset: 13223},
									name: "__",
								},
								&ruleRefExpr{
									pos:  position{line: 396, col: 39, offset: 13226},
									name: "LabelMatch",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "LabelList",
			pos:  position{line: 397, col: 1, offset: 13239},
			expr: &choiceExpr{
				pos: position{line: 397, col: 13, offset: 13251},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 397, col: 14, offset: 13252},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 397, col: 14, offset: 13252},
								val:        "(",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 397, col: 18, offset: 13256},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 397, col: 21, offset: 13259},
								val:        ")",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 397, col: 29, offset: 13267},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 397, col: 29, offset: 13267},
								val:        "(",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 397, col: 33, offset: 13271},
								name: "__",
							},
							&labeledExpr{
								pos:   position{line: 397, col: 36, offset: 13274},
								label: "label",
								expr: &ruleRefExpr{
									pos:  position{line: 397, col: 42, offset: 13280},
									name: "Label",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 397, col: 48, offset: 13286},
								name: "__",
							},
							&labeledExpr{
								pos:   position{line: 397, col: 51, offset: 13289},
								label: "rest",
								expr: &zeroOrMoreExpr{
									pos: position{line: 397, col: 56, offset: 13294},
									expr: &seqExpr{
										pos: position{line: 397, col: 58, offset: 13296},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 397, col: 58, offset: 13296},
												val:        ",",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 397, col: 62, offset: 13300},
												name: "__",
											},
											&ruleRefExpr{
												pos:  position{line: 397, col: 65, offset: 13303},
												name: "Label",
											},
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 397, col: 74, offset: 13312},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 397, col: 77, offset: 13315},
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
			pos:  position{line: 399, col: 1, offset: 13321},
			expr: &seqExpr{
				pos: position{line: 399, col: 18, offset: 13338},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 399, col: 18, offset: 13338},
						label: "metric",
						expr: &ruleRefExpr{
							pos:  position{line: 399, col: 25, offset: 13345},
							name: "Identifier",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 399, col: 36, offset: 13356},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 399, col: 40, offset: 13360},
						expr: &ruleRefExpr{
							pos:  position{line: 399, col: 40, offset: 13360},
							name: "CodeBlock",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 399, col: 51, offset: 13371},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 399, col: 54, offset: 13374},
						expr: &ruleRefExpr{
							pos:  position{line: 399, col: 54, offset: 13374},
							name: "Range",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 399, col: 61, offset: 13381},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 399, col: 64, offset: 13384},
						expr: &ruleRefExpr{
							pos:  position{line: 399, col: 64, offset: 13384},
							name: "Offset",
						},
					},
				},
			},
		},
		{
			name: "Range",
			pos:  position{line: 401, col: 1, offset: 13393},
			expr: &actionExpr{
				pos: position{line: 401, col: 9, offset: 13401},
				run: (*parser).callonRange1,
				expr: &seqExpr{
					pos: position{line: 401, col: 9, offset: 13401},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 401, col: 9, offset: 13401},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 401, col: 13, offset: 13405},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 401, col: 16, offset: 13408},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 401, col: 20, offset: 13412},
								name: "Duration",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 401, col: 29, offset: 13421},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 401, col: 32, offset: 13424},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Offset",
			pos:  position{line: 405, col: 1, offset: 13453},
			expr: &actionExpr{
				pos: position{line: 405, col: 10, offset: 13462},
				run: (*parser).callonOffset1,
				expr: &seqExpr{
					pos: position{line: 405, col: 10, offset: 13462},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 405, col: 10, offset: 13462},
							val:        "offset",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 405, col: 20, offset: 13472},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 405, col: 23, offset: 13475},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 405, col: 27, offset: 13479},
								name: "Duration",
							},
						},
					},
				},
			},
		},
		{
			name: "BinaryAggregateOperators",
			pos:  position{line: 409, col: 1, offset: 13513},
			expr: &choiceExpr{
				pos: position{line: 409, col: 28, offset: 13540},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 409, col: 28, offset: 13540},
						val:        "count_values",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 409, col: 46, offset: 13558},
						val:        "topk",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 409, col: 56, offset: 13568},
						val:        "bottomk",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 409, col: 69, offset: 13581},
						val:        "quantile",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "UnaryAggregateOperators",
			pos:  position{line: 410, col: 1, offset: 13593},
			expr: &choiceExpr{
				pos: position{line: 410, col: 27, offset: 13619},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 410, col: 27, offset: 13619},
						val:        "sum",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 36, offset: 13628},
						val:        "min",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 45, offset: 13637},
						val:        "max",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 54, offset: 13646},
						val:        "avg",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 63, offset: 13655},
						val:        "stddev",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 75, offset: 13667},
						val:        "stdvar",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 410, col: 87, offset: 13679},
						val:        "count",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "AggregateOperators",
			pos:  position{line: 411, col: 1, offset: 13688},
			expr: &choiceExpr{
				pos: position{line: 411, col: 22, offset: 13709},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 411, col: 22, offset: 13709},
						name: "BinaryAggregateOperators",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 49, offset: 13736},
						name: "UnaryAggregateOperators",
					},
				},
			},
		},
		{
			name: "AggregateBy",
			pos:  position{line: 413, col: 1, offset: 13761},
			expr: &actionExpr{
				pos: position{line: 413, col: 15, offset: 13775},
				run: (*parser).callonAggregateBy1,
				expr: &seqExpr{
					pos: position{line: 413, col: 15, offset: 13775},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 413, col: 15, offset: 13775},
							val:        "by",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 413, col: 21, offset: 13781},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 413, col: 24, offset: 13784},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 413, col: 31, offset: 13791},
								name: "LabelList",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 413, col: 41, offset: 13801},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 413, col: 44, offset: 13804},
							label: "keep",
							expr: &zeroOrOneExpr{
								pos: position{line: 413, col: 49, offset: 13809},
								expr: &litMatcher{
									pos:        position{line: 413, col: 49, offset: 13809},
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
			pos:  position{line: 418, col: 1, offset: 13885},
			expr: &actionExpr{
				pos: position{line: 418, col: 20, offset: 13904},
				run: (*parser).callonAggregateWithout1,
				expr: &seqExpr{
					pos: position{line: 418, col: 20, offset: 13904},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 418, col: 20, offset: 13904},
							val:        "without",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 31, offset: 13915},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 418, col: 34, offset: 13918},
							label: "labels",
							expr: &ruleRefExpr{
								pos:  position{line: 418, col: 41, offset: 13925},
								name: "LabelList",
							},
						},
					},
				},
			},
		},
		{
			name: "AggregateGroup",
			pos:  position{line: 422, col: 1, offset: 13963},
			expr: &choiceExpr{
				pos: position{line: 422, col: 18, offset: 13980},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 422, col: 18, offset: 13980},
						name: "AggregateBy",
					},
					&ruleRefExpr{
						pos:  position{line: 422, col: 32, offset: 13994},
						name: "AggregateWithout",
					},
				},
			},
		},
		{
			name: "AggregateExpression",
			pos:  position{line: 424, col: 1, offset: 14012},
			expr: &choiceExpr{
				pos: position{line: 424, col: 23, offset: 14034},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 424, col: 23, offset: 14034},
						run: (*parser).callonAggregateExpression2,
						expr: &seqExpr{
							pos: position{line: 424, col: 23, offset: 14034},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 424, col: 23, offset: 14034},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 424, col: 26, offset: 14037},
										name: "AggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 424, col: 45, offset: 14056},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 424, col: 48, offset: 14059},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 424, col: 52, offset: 14063},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 424, col: 55, offset: 14066},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 424, col: 62, offset: 14073},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 424, col: 77, offset: 14088},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 424, col: 80, offset: 14091},
									val:        ")",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 424, col: 84, offset: 14095},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 424, col: 87, offset: 14098},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 424, col: 93, offset: 14104},
										expr: &ruleRefExpr{
											pos:  position{line: 424, col: 93, offset: 14104},
											name: "AggregateGroup",
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 426, col: 5, offset: 14149},
						run: (*parser).callonAggregateExpression17,
						expr: &seqExpr{
							pos: position{line: 426, col: 5, offset: 14149},
							exprs: []interface{}{
								&labeledExpr{
									pos:   position{line: 426, col: 5, offset: 14149},
									label: "op",
									expr: &ruleRefExpr{
										pos:  position{line: 426, col: 8, offset: 14152},
										name: "AggregateOperators",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 426, col: 27, offset: 14171},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 426, col: 30, offset: 14174},
									label: "group",
									expr: &zeroOrOneExpr{
										pos: position{line: 426, col: 36, offset: 14180},
										expr: &ruleRefExpr{
											pos:  position{line: 426, col: 36, offset: 14180},
											name: "AggregateGroup",
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 426, col: 52, offset: 14196},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 426, col: 55, offset: 14199},
									val:        "(",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 426, col: 59, offset: 14203},
									name: "__",
								},
								&labeledExpr{
									pos:   position{line: 426, col: 62, offset: 14206},
									label: "vector",
									expr: &ruleRefExpr{
										pos:  position{line: 426, col: 69, offset: 14213},
										name: "VectorSelector",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 426, col: 84, offset: 14228},
									name: "__",
								},
								&litMatcher{
									pos:        position{line: 426, col: 87, offset: 14231},
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
			pos:  position{line: 430, col: 1, offset: 14263},
			expr: &zeroOrMoreExpr{
				pos: position{line: 430, col: 6, offset: 14268},
				expr: &choiceExpr{
					pos: position{line: 430, col: 8, offset: 14270},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 430, col: 8, offset: 14270},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 430, col: 21, offset: 14283},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 430, col: 27, offset: 14289},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 431, col: 1, offset: 14300},
			expr: &zeroOrMoreExpr{
				pos: position{line: 431, col: 5, offset: 14304},
				expr: &ruleRefExpr{
					pos:  position{line: 431, col: 5, offset: 14304},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 433, col: 1, offset: 14317},
			expr: &charClassMatcher{
				pos:        position{line: 433, col: 14, offset: 14330},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 434, col: 1, offset: 14338},
			expr: &litMatcher{
				pos:        position{line: 434, col: 7, offset: 14344},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 435, col: 1, offset: 14349},
			expr: &choiceExpr{
				pos: position{line: 435, col: 7, offset: 14355},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 435, col: 7, offset: 14355},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 435, col: 7, offset: 14355},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 435, col: 10, offset: 14358},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 435, col: 16, offset: 14364},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 435, col: 16, offset: 14364},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 435, col: 18, offset: 14366},
								expr: &ruleRefExpr{
									pos:  position{line: 435, col: 18, offset: 14366},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 435, col: 37, offset: 14385},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 435, col: 43, offset: 14391},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 435, col: 43, offset: 14391},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 435, col: 46, offset: 14394},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 437, col: 1, offset: 14399},
			expr: &notExpr{
				pos: position{line: 437, col: 7, offset: 14405},
				expr: &anyMatcher{
					line: 437, col: 8, offset: 14406,
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
	return ident, nil
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
	return string(c.text), nil
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
	return strconv.ParseFloat(string(c.text), 64)
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
