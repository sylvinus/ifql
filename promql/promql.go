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
									name: "VectorSelector",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 41, offset: 7736},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 223, col: 1, offset: 7776},
			expr: &anyMatcher{
				line: 223, col: 14, offset: 7789,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 225, col: 1, offset: 7792},
			expr: &seqExpr{
				pos: position{line: 225, col: 11, offset: 7802},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 225, col: 11, offset: 7802},
						val:        "#",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 225, col: 15, offset: 7806},
						expr: &seqExpr{
							pos: position{line: 225, col: 17, offset: 7808},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 225, col: 17, offset: 7808},
									expr: &ruleRefExpr{
										pos:  position{line: 225, col: 18, offset: 7809},
										name: "EOL",
									},
								},
								&ruleRefExpr{
									pos:  position{line: 225, col: 22, offset: 7813},
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
			pos:  position{line: 227, col: 1, offset: 7828},
			expr: &actionExpr{
				pos: position{line: 227, col: 14, offset: 7841},
				run: (*parser).callonIdentifier1,
				expr: &labeledExpr{
					pos:   position{line: 227, col: 14, offset: 7841},
					label: "ident",
					expr: &ruleRefExpr{
						pos:  position{line: 227, col: 20, offset: 7847},
						name: "IdentifierName",
					},
				},
			},
		},
		{
			name: "IdentifierName",
			pos:  position{line: 235, col: 1, offset: 8009},
			expr: &actionExpr{
				pos: position{line: 235, col: 18, offset: 8026},
				run: (*parser).callonIdentifierName1,
				expr: &seqExpr{
					pos: position{line: 235, col: 18, offset: 8026},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 235, col: 18, offset: 8026},
							name: "IdentifierStart",
						},
						&zeroOrMoreExpr{
							pos: position{line: 235, col: 34, offset: 8042},
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 34, offset: 8042},
								name: "IdentifierPart",
							},
						},
					},
				},
			},
		},
		{
			name: "IdentifierStart",
			pos:  position{line: 238, col: 1, offset: 8093},
			expr: &charClassMatcher{
				pos:        position{line: 238, col: 19, offset: 8111},
				val:        "[\\pL_]",
				chars:      []rune{'_'},
				classes:    []*unicode.RangeTable{rangeTable("L")},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "IdentifierPart",
			pos:  position{line: 239, col: 1, offset: 8118},
			expr: &choiceExpr{
				pos: position{line: 239, col: 18, offset: 8135},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 239, col: 18, offset: 8135},
						name: "IdentifierStart",
					},
					&charClassMatcher{
						pos:        position{line: 239, col: 36, offset: 8153},
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
			pos:  position{line: 241, col: 1, offset: 8163},
			expr: &choiceExpr{
				pos: position{line: 241, col: 17, offset: 8179},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 241, col: 17, offset: 8179},
						run: (*parser).callonStringLiteral2,
						expr: &choiceExpr{
							pos: position{line: 241, col: 19, offset: 8181},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 241, col: 19, offset: 8181},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 19, offset: 8181},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 241, col: 23, offset: 8185},
											expr: &ruleRefExpr{
												pos:  position{line: 241, col: 23, offset: 8185},
												name: "DoubleStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 241, col: 41, offset: 8203},
											val:        "\"",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 241, col: 47, offset: 8209},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 47, offset: 8209},
											val:        "'",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 241, col: 51, offset: 8213},
											name: "SingleStringChar",
										},
										&litMatcher{
											pos:        position{line: 241, col: 68, offset: 8230},
											val:        "'",
											ignoreCase: false,
										},
									},
								},
								&seqExpr{
									pos: position{line: 241, col: 74, offset: 8236},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 241, col: 74, offset: 8236},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 241, col: 78, offset: 8240},
											expr: &ruleRefExpr{
												pos:  position{line: 241, col: 78, offset: 8240},
												name: "RawStringChar",
											},
										},
										&litMatcher{
											pos:        position{line: 241, col: 93, offset: 8255},
											val:        "`",
											ignoreCase: false,
										},
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 243, col: 5, offset: 8298},
						run: (*parser).callonStringLiteral18,
						expr: &choiceExpr{
							pos: position{line: 243, col: 7, offset: 8300},
							alternatives: []interface{}{
								&seqExpr{
									pos: position{line: 243, col: 9, offset: 8302},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 9, offset: 8302},
											val:        "\"",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 243, col: 13, offset: 8306},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 13, offset: 8306},
												name: "DoubleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 243, col: 33, offset: 8326},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 243, col: 33, offset: 8326},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 243, col: 39, offset: 8332},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 243, col: 51, offset: 8344},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 51, offset: 8344},
											val:        "'",
											ignoreCase: false,
										},
										&zeroOrOneExpr{
											pos: position{line: 243, col: 55, offset: 8348},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 55, offset: 8348},
												name: "SingleStringChar",
											},
										},
										&choiceExpr{
											pos: position{line: 243, col: 75, offset: 8368},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 243, col: 75, offset: 8368},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 243, col: 81, offset: 8374},
													name: "EOF",
												},
											},
										},
									},
								},
								&seqExpr{
									pos: position{line: 243, col: 91, offset: 8384},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 243, col: 91, offset: 8384},
											val:        "`",
											ignoreCase: false,
										},
										&zeroOrMoreExpr{
											pos: position{line: 243, col: 95, offset: 8388},
											expr: &ruleRefExpr{
												pos:  position{line: 243, col: 95, offset: 8388},
												name: "RawStringChar",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 243, col: 110, offset: 8403},
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
			pos:  position{line: 247, col: 1, offset: 8474},
			expr: &choiceExpr{
				pos: position{line: 247, col: 20, offset: 8493},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 247, col: 20, offset: 8493},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 247, col: 20, offset: 8493},
								expr: &choiceExpr{
									pos: position{line: 247, col: 23, offset: 8496},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 247, col: 23, offset: 8496},
											val:        "\"",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 247, col: 29, offset: 8502},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 247, col: 36, offset: 8509},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 247, col: 42, offset: 8515},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 247, col: 55, offset: 8528},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 247, col: 55, offset: 8528},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 247, col: 60, offset: 8533},
								name: "DoubleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "SingleStringChar",
			pos:  position{line: 248, col: 1, offset: 8552},
			expr: &choiceExpr{
				pos: position{line: 248, col: 20, offset: 8571},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 248, col: 20, offset: 8571},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 248, col: 20, offset: 8571},
								expr: &choiceExpr{
									pos: position{line: 248, col: 23, offset: 8574},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 248, col: 23, offset: 8574},
											val:        "'",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 248, col: 29, offset: 8580},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 36, offset: 8587},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 248, col: 42, offset: 8593},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 248, col: 55, offset: 8606},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 248, col: 55, offset: 8606},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 248, col: 60, offset: 8611},
								name: "SingleStringEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "RawStringChar",
			pos:  position{line: 249, col: 1, offset: 8630},
			expr: &seqExpr{
				pos: position{line: 249, col: 17, offset: 8646},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 249, col: 17, offset: 8646},
						expr: &litMatcher{
							pos:        position{line: 249, col: 18, offset: 8647},
							val:        "`",
							ignoreCase: false,
						},
					},
					&ruleRefExpr{
						pos:  position{line: 249, col: 22, offset: 8651},
						name: "SourceChar",
					},
				},
			},
		},
		{
			name: "DoubleStringEscape",
			pos:  position{line: 251, col: 1, offset: 8663},
			expr: &choiceExpr{
				pos: position{line: 251, col: 22, offset: 8684},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 251, col: 24, offset: 8686},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 251, col: 24, offset: 8686},
								val:        "\"",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 251, col: 30, offset: 8692},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 252, col: 7, offset: 8721},
						run: (*parser).callonDoubleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 252, col: 9, offset: 8723},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 252, col: 9, offset: 8723},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 252, col: 22, offset: 8736},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 252, col: 28, offset: 8742},
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
			pos:  position{line: 255, col: 1, offset: 8807},
			expr: &choiceExpr{
				pos: position{line: 255, col: 22, offset: 8828},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 255, col: 24, offset: 8830},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 255, col: 24, offset: 8830},
								val:        "'",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 255, col: 30, offset: 8836},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 256, col: 7, offset: 8865},
						run: (*parser).callonSingleStringEscape5,
						expr: &choiceExpr{
							pos: position{line: 256, col: 9, offset: 8867},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 256, col: 9, offset: 8867},
									name: "SourceChar",
								},
								&ruleRefExpr{
									pos:  position{line: 256, col: 22, offset: 8880},
									name: "EOL",
								},
								&ruleRefExpr{
									pos:  position{line: 256, col: 28, offset: 8886},
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
			pos:  position{line: 260, col: 1, offset: 8952},
			expr: &choiceExpr{
				pos: position{line: 260, col: 24, offset: 8975},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 260, col: 24, offset: 8975},
						name: "SingleCharEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 43, offset: 8994},
						name: "OctalEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 57, offset: 9008},
						name: "HexEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 69, offset: 9020},
						name: "LongUnicodeEscape",
					},
					&ruleRefExpr{
						pos:  position{line: 260, col: 89, offset: 9040},
						name: "ShortUnicodeEscape",
					},
				},
			},
		},
		{
			name: "SingleCharEscape",
			pos:  position{line: 261, col: 1, offset: 9059},
			expr: &choiceExpr{
				pos: position{line: 261, col: 20, offset: 9078},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 261, col: 20, offset: 9078},
						val:        "a",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 26, offset: 9084},
						val:        "b",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 32, offset: 9090},
						val:        "n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 38, offset: 9096},
						val:        "f",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 44, offset: 9102},
						val:        "r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 50, offset: 9108},
						val:        "t",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 56, offset: 9114},
						val:        "v",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 261, col: 62, offset: 9120},
						val:        "\\",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "OctalEscape",
			pos:  position{line: 262, col: 1, offset: 9125},
			expr: &choiceExpr{
				pos: position{line: 262, col: 15, offset: 9139},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 262, col: 15, offset: 9139},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 262, col: 15, offset: 9139},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 262, col: 26, offset: 9150},
								name: "OctalDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 262, col: 37, offset: 9161},
								name: "OctalDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 263, col: 7, offset: 9178},
						run: (*parser).callonOctalEscape6,
						expr: &seqExpr{
							pos: position{line: 263, col: 7, offset: 9178},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 263, col: 7, offset: 9178},
									name: "OctalDigit",
								},
								&choiceExpr{
									pos: position{line: 263, col: 20, offset: 9191},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 263, col: 20, offset: 9191},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 263, col: 33, offset: 9204},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 263, col: 39, offset: 9210},
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
			pos:  position{line: 266, col: 1, offset: 9271},
			expr: &choiceExpr{
				pos: position{line: 266, col: 13, offset: 9283},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 266, col: 13, offset: 9283},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 266, col: 13, offset: 9283},
								val:        "x",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 266, col: 17, offset: 9287},
								name: "HexDigit",
							},
							&ruleRefExpr{
								pos:  position{line: 266, col: 26, offset: 9296},
								name: "HexDigit",
							},
						},
					},
					&actionExpr{
						pos: position{line: 267, col: 7, offset: 9311},
						run: (*parser).callonHexEscape6,
						expr: &seqExpr{
							pos: position{line: 267, col: 7, offset: 9311},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 267, col: 7, offset: 9311},
									val:        "x",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 267, col: 13, offset: 9317},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 267, col: 13, offset: 9317},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 267, col: 26, offset: 9330},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 267, col: 32, offset: 9336},
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
			pos:  position{line: 270, col: 1, offset: 9403},
			expr: &choiceExpr{
				pos: position{line: 271, col: 5, offset: 9428},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 271, col: 5, offset: 9428},
						run: (*parser).callonLongUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 271, col: 5, offset: 9428},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 271, col: 5, offset: 9428},
									val:        "U",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 9, offset: 9432},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 18, offset: 9441},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 27, offset: 9450},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 36, offset: 9459},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 45, offset: 9468},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 54, offset: 9477},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 63, offset: 9486},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 72, offset: 9495},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 274, col: 7, offset: 9597},
						run: (*parser).callonLongUnicodeEscape13,
						expr: &seqExpr{
							pos: position{line: 274, col: 7, offset: 9597},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 274, col: 7, offset: 9597},
									val:        "U",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 274, col: 13, offset: 9603},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 274, col: 13, offset: 9603},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 274, col: 26, offset: 9616},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 274, col: 32, offset: 9622},
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
			pos:  position{line: 277, col: 1, offset: 9685},
			expr: &choiceExpr{
				pos: position{line: 278, col: 5, offset: 9711},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 278, col: 5, offset: 9711},
						run: (*parser).callonShortUnicodeEscape2,
						expr: &seqExpr{
							pos: position{line: 278, col: 5, offset: 9711},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 278, col: 5, offset: 9711},
									val:        "u",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 9, offset: 9715},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 18, offset: 9724},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 27, offset: 9733},
									name: "HexDigit",
								},
								&ruleRefExpr{
									pos:  position{line: 278, col: 36, offset: 9742},
									name: "HexDigit",
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 281, col: 7, offset: 9844},
						run: (*parser).callonShortUnicodeEscape9,
						expr: &seqExpr{
							pos: position{line: 281, col: 7, offset: 9844},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 281, col: 7, offset: 9844},
									val:        "u",
									ignoreCase: false,
								},
								&choiceExpr{
									pos: position{line: 281, col: 13, offset: 9850},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 281, col: 13, offset: 9850},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 281, col: 26, offset: 9863},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 281, col: 32, offset: 9869},
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
			pos:  position{line: 285, col: 1, offset: 9933},
			expr: &charClassMatcher{
				pos:        position{line: 285, col: 14, offset: 9946},
				val:        "[0-7]",
				ranges:     []rune{'0', '7'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 286, col: 1, offset: 9952},
			expr: &charClassMatcher{
				pos:        position{line: 286, col: 16, offset: 9967},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "HexDigit",
			pos:  position{line: 287, col: 1, offset: 9973},
			expr: &charClassMatcher{
				pos:        position{line: 287, col: 12, offset: 9984},
				val:        "[0-9a-f]i",
				ranges:     []rune{'0', '9', 'a', 'f'},
				ignoreCase: true,
				inverted:   false,
			},
		},
		{
			name: "CharClassMatcher",
			pos:  position{line: 289, col: 1, offset: 9995},
			expr: &choiceExpr{
				pos: position{line: 289, col: 20, offset: 10014},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 289, col: 20, offset: 10014},
						run: (*parser).callonCharClassMatcher2,
						expr: &seqExpr{
							pos: position{line: 289, col: 20, offset: 10014},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 289, col: 20, offset: 10014},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 289, col: 24, offset: 10018},
									expr: &choiceExpr{
										pos: position{line: 289, col: 26, offset: 10020},
										alternatives: []interface{}{
											&ruleRefExpr{
												pos:  position{line: 289, col: 26, offset: 10020},
												name: "ClassCharRange",
											},
											&ruleRefExpr{
												pos:  position{line: 289, col: 43, offset: 10037},
												name: "ClassChar",
											},
											&seqExpr{
												pos: position{line: 289, col: 55, offset: 10049},
												exprs: []interface{}{
													&litMatcher{
														pos:        position{line: 289, col: 55, offset: 10049},
														val:        "\\",
														ignoreCase: false,
													},
													&ruleRefExpr{
														pos:  position{line: 289, col: 60, offset: 10054},
														name: "UnicodeClassEscape",
													},
												},
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 289, col: 82, offset: 10076},
									val:        "]",
									ignoreCase: false,
								},
								&zeroOrOneExpr{
									pos: position{line: 289, col: 86, offset: 10080},
									expr: &litMatcher{
										pos:        position{line: 289, col: 86, offset: 10080},
										val:        "i",
										ignoreCase: false,
									},
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 291, col: 5, offset: 10122},
						run: (*parser).callonCharClassMatcher15,
						expr: &seqExpr{
							pos: position{line: 291, col: 5, offset: 10122},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 291, col: 5, offset: 10122},
									val:        "[",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 291, col: 9, offset: 10126},
									expr: &seqExpr{
										pos: position{line: 291, col: 11, offset: 10128},
										exprs: []interface{}{
											&notExpr{
												pos: position{line: 291, col: 11, offset: 10128},
												expr: &ruleRefExpr{
													pos:  position{line: 291, col: 14, offset: 10131},
													name: "EOL",
												},
											},
											&ruleRefExpr{
												pos:  position{line: 291, col: 20, offset: 10137},
												name: "SourceChar",
											},
										},
									},
								},
								&choiceExpr{
									pos: position{line: 291, col: 36, offset: 10153},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 291, col: 36, offset: 10153},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 291, col: 42, offset: 10159},
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
			pos:  position{line: 295, col: 1, offset: 10231},
			expr: &seqExpr{
				pos: position{line: 295, col: 18, offset: 10248},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 295, col: 18, offset: 10248},
						name: "ClassChar",
					},
					&litMatcher{
						pos:        position{line: 295, col: 28, offset: 10258},
						val:        "-",
						ignoreCase: false,
					},
					&ruleRefExpr{
						pos:  position{line: 295, col: 32, offset: 10262},
						name: "ClassChar",
					},
				},
			},
		},
		{
			name: "ClassChar",
			pos:  position{line: 296, col: 1, offset: 10272},
			expr: &choiceExpr{
				pos: position{line: 296, col: 13, offset: 10284},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 296, col: 13, offset: 10284},
						exprs: []interface{}{
							&notExpr{
								pos: position{line: 296, col: 13, offset: 10284},
								expr: &choiceExpr{
									pos: position{line: 296, col: 16, offset: 10287},
									alternatives: []interface{}{
										&litMatcher{
											pos:        position{line: 296, col: 16, offset: 10287},
											val:        "]",
											ignoreCase: false,
										},
										&litMatcher{
											pos:        position{line: 296, col: 22, offset: 10293},
											val:        "\\",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 296, col: 29, offset: 10300},
											name: "EOL",
										},
									},
								},
							},
							&ruleRefExpr{
								pos:  position{line: 296, col: 35, offset: 10306},
								name: "SourceChar",
							},
						},
					},
					&seqExpr{
						pos: position{line: 296, col: 48, offset: 10319},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 296, col: 48, offset: 10319},
								val:        "\\",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 296, col: 53, offset: 10324},
								name: "CharClassEscape",
							},
						},
					},
				},
			},
		},
		{
			name: "CharClassEscape",
			pos:  position{line: 297, col: 1, offset: 10340},
			expr: &choiceExpr{
				pos: position{line: 297, col: 19, offset: 10358},
				alternatives: []interface{}{
					&choiceExpr{
						pos: position{line: 297, col: 21, offset: 10360},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 297, col: 21, offset: 10360},
								val:        "]",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 297, col: 27, offset: 10366},
								name: "CommonEscapeSequence",
							},
						},
					},
					&actionExpr{
						pos: position{line: 298, col: 7, offset: 10395},
						run: (*parser).callonCharClassEscape5,
						expr: &seqExpr{
							pos: position{line: 298, col: 7, offset: 10395},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 298, col: 7, offset: 10395},
									expr: &litMatcher{
										pos:        position{line: 298, col: 8, offset: 10396},
										val:        "p",
										ignoreCase: false,
									},
								},
								&choiceExpr{
									pos: position{line: 298, col: 14, offset: 10402},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 298, col: 14, offset: 10402},
											name: "SourceChar",
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 27, offset: 10415},
											name: "EOL",
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 33, offset: 10421},
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
			pos:  position{line: 302, col: 1, offset: 10487},
			expr: &seqExpr{
				pos: position{line: 302, col: 22, offset: 10508},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 302, col: 22, offset: 10508},
						val:        "p",
						ignoreCase: false,
					},
					&choiceExpr{
						pos: position{line: 303, col: 7, offset: 10521},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 303, col: 7, offset: 10521},
								name: "SingleCharUnicodeClass",
							},
							&actionExpr{
								pos: position{line: 304, col: 7, offset: 10550},
								run: (*parser).callonUnicodeClassEscape5,
								expr: &seqExpr{
									pos: position{line: 304, col: 7, offset: 10550},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 304, col: 7, offset: 10550},
											expr: &litMatcher{
												pos:        position{line: 304, col: 8, offset: 10551},
												val:        "{",
												ignoreCase: false,
											},
										},
										&choiceExpr{
											pos: position{line: 304, col: 14, offset: 10557},
											alternatives: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 304, col: 14, offset: 10557},
													name: "SourceChar",
												},
												&ruleRefExpr{
													pos:  position{line: 304, col: 27, offset: 10570},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 304, col: 33, offset: 10576},
													name: "EOF",
												},
											},
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 305, col: 7, offset: 10647},
								run: (*parser).callonUnicodeClassEscape13,
								expr: &seqExpr{
									pos: position{line: 305, col: 7, offset: 10647},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 305, col: 7, offset: 10647},
											val:        "{",
											ignoreCase: false,
										},
										&labeledExpr{
											pos:   position{line: 305, col: 11, offset: 10651},
											label: "ident",
											expr: &ruleRefExpr{
												pos:  position{line: 305, col: 17, offset: 10657},
												name: "IdentifierName",
											},
										},
										&litMatcher{
											pos:        position{line: 305, col: 32, offset: 10672},
											val:        "}",
											ignoreCase: false,
										},
									},
								},
							},
							&actionExpr{
								pos: position{line: 311, col: 7, offset: 10836},
								run: (*parser).callonUnicodeClassEscape19,
								expr: &seqExpr{
									pos: position{line: 311, col: 7, offset: 10836},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 311, col: 7, offset: 10836},
											val:        "{",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 311, col: 11, offset: 10840},
											name: "IdentifierName",
										},
										&choiceExpr{
											pos: position{line: 311, col: 28, offset: 10857},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 311, col: 28, offset: 10857},
													val:        "]",
													ignoreCase: false,
												},
												&ruleRefExpr{
													pos:  position{line: 311, col: 34, offset: 10863},
													name: "EOL",
												},
												&ruleRefExpr{
													pos:  position{line: 311, col: 40, offset: 10869},
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
			pos:  position{line: 316, col: 1, offset: 10949},
			expr: &charClassMatcher{
				pos:        position{line: 316, col: 26, offset: 10974},
				val:        "[LMNCPZS]",
				chars:      []rune{'L', 'M', 'N', 'C', 'P', 'Z', 'S'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Number",
			pos:  position{line: 319, col: 1, offset: 10986},
			expr: &actionExpr{
				pos: position{line: 319, col: 10, offset: 10995},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 319, col: 10, offset: 10995},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 319, col: 10, offset: 10995},
							expr: &litMatcher{
								pos:        position{line: 319, col: 10, offset: 10995},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 319, col: 15, offset: 11000},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 319, col: 23, offset: 11008},
							expr: &seqExpr{
								pos: position{line: 319, col: 25, offset: 11010},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 319, col: 25, offset: 11010},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 319, col: 29, offset: 11014},
										expr: &ruleRefExpr{
											pos:  position{line: 319, col: 29, offset: 11014},
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
			pos:  position{line: 323, col: 1, offset: 11079},
			expr: &choiceExpr{
				pos: position{line: 323, col: 11, offset: 11089},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 323, col: 11, offset: 11089},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 323, col: 17, offset: 11095},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 323, col: 17, offset: 11095},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 323, col: 17, offset: 11095},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 323, col: 30, offset: 11108},
									expr: &ruleRefExpr{
										pos:  position{line: 323, col: 30, offset: 11108},
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
			pos:  position{line: 327, col: 1, offset: 11172},
			expr: &charClassMatcher{
				pos:        position{line: 327, col: 16, offset: 11187},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 328, col: 1, offset: 11193},
			expr: &charClassMatcher{
				pos:        position{line: 328, col: 9, offset: 11201},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "CodeBlock",
			pos:  position{line: 330, col: 1, offset: 11208},
			expr: &choiceExpr{
				pos: position{line: 330, col: 13, offset: 11220},
				alternatives: []interface{}{
					&actionExpr{
						pos: position{line: 330, col: 13, offset: 11220},
						run: (*parser).callonCodeBlock2,
						expr: &seqExpr{
							pos: position{line: 330, col: 13, offset: 11220},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 330, col: 13, offset: 11220},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 330, col: 17, offset: 11224},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 330, col: 22, offset: 11229},
									val:        "}",
									ignoreCase: false,
								},
							},
						},
					},
					&actionExpr{
						pos: position{line: 332, col: 5, offset: 11270},
						run: (*parser).callonCodeBlock7,
						expr: &seqExpr{
							pos: position{line: 332, col: 5, offset: 11270},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 332, col: 5, offset: 11270},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 332, col: 9, offset: 11274},
									name: "Code",
								},
								&ruleRefExpr{
									pos:  position{line: 332, col: 14, offset: 11279},
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
			pos:  position{line: 336, col: 1, offset: 11344},
			expr: &actionExpr{
				pos: position{line: 336, col: 19, offset: 11362},
				run: (*parser).callonNanoSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 336, col: 19, offset: 11362},
					val:        "ns",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MicroSecondUnits",
			pos:  position{line: 341, col: 1, offset: 11467},
			expr: &actionExpr{
				pos: position{line: 341, col: 20, offset: 11486},
				run: (*parser).callonMicroSecondUnits1,
				expr: &choiceExpr{
					pos: position{line: 341, col: 21, offset: 11487},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 21, offset: 11487},
							val:        "us",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 28, offset: 11494},
							val:        "µs",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 35, offset: 11502},
							val:        "μs",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "MilliSecondUnits",
			pos:  position{line: 346, col: 1, offset: 11611},
			expr: &actionExpr{
				pos: position{line: 346, col: 20, offset: 11630},
				run: (*parser).callonMilliSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 346, col: 20, offset: 11630},
					val:        "ms",
					ignoreCase: false,
				},
			},
		},
		{
			name: "SecondUnits",
			pos:  position{line: 351, col: 1, offset: 11737},
			expr: &actionExpr{
				pos: position{line: 351, col: 15, offset: 11751},
				run: (*parser).callonSecondUnits1,
				expr: &litMatcher{
					pos:        position{line: 351, col: 15, offset: 11751},
					val:        "s",
					ignoreCase: false,
				},
			},
		},
		{
			name: "MinuteUnits",
			pos:  position{line: 355, col: 1, offset: 11788},
			expr: &actionExpr{
				pos: position{line: 355, col: 15, offset: 11802},
				run: (*parser).callonMinuteUnits1,
				expr: &litMatcher{
					pos:        position{line: 355, col: 15, offset: 11802},
					val:        "m",
					ignoreCase: false,
				},
			},
		},
		{
			name: "HourUnits",
			pos:  position{line: 359, col: 1, offset: 11839},
			expr: &actionExpr{
				pos: position{line: 359, col: 13, offset: 11851},
				run: (*parser).callonHourUnits1,
				expr: &litMatcher{
					pos:        position{line: 359, col: 13, offset: 11851},
					val:        "h",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DayUnits",
			pos:  position{line: 363, col: 1, offset: 11886},
			expr: &actionExpr{
				pos: position{line: 363, col: 12, offset: 11897},
				run: (*parser).callonDayUnits1,
				expr: &litMatcher{
					pos:        position{line: 363, col: 12, offset: 11897},
					val:        "d",
					ignoreCase: false,
				},
			},
		},
		{
			name: "WeekUnits",
			pos:  position{line: 369, col: 1, offset: 12105},
			expr: &actionExpr{
				pos: position{line: 369, col: 13, offset: 12117},
				run: (*parser).callonWeekUnits1,
				expr: &litMatcher{
					pos:        position{line: 369, col: 13, offset: 12117},
					val:        "w",
					ignoreCase: false,
				},
			},
		},
		{
			name: "YearUnits",
			pos:  position{line: 375, col: 1, offset: 12328},
			expr: &actionExpr{
				pos: position{line: 375, col: 13, offset: 12340},
				run: (*parser).callonYearUnits1,
				expr: &litMatcher{
					pos:        position{line: 375, col: 13, offset: 12340},
					val:        "y",
					ignoreCase: false,
				},
			},
		},
		{
			name: "DurationUnits",
			pos:  position{line: 381, col: 1, offset: 12537},
			expr: &choiceExpr{
				pos: position{line: 381, col: 18, offset: 12554},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 381, col: 18, offset: 12554},
						name: "NanoSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 36, offset: 12572},
						name: "MicroSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 55, offset: 12591},
						name: "MilliSecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 74, offset: 12610},
						name: "SecondUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 88, offset: 12624},
						name: "MinuteUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 102, offset: 12638},
						name: "HourUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 114, offset: 12650},
						name: "DayUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 125, offset: 12661},
						name: "WeekUnits",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 137, offset: 12673},
						name: "YearUnits",
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 383, col: 1, offset: 12685},
			expr: &actionExpr{
				pos: position{line: 383, col: 12, offset: 12696},
				run: (*parser).callonDuration1,
				expr: &seqExpr{
					pos: position{line: 383, col: 12, offset: 12696},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 383, col: 12, offset: 12696},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 16, offset: 12700},
								name: "Integer",
							},
						},
						&labeledExpr{
							pos:   position{line: 383, col: 24, offset: 12708},
							label: "units",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 30, offset: 12714},
								name: "DurationUnits",
							},
						},
					},
				},
			},
		},
		{
			name: "Code",
			pos:  position{line: 389, col: 1, offset: 12863},
			expr: &zeroOrMoreExpr{
				pos: position{line: 389, col: 8, offset: 12870},
				expr: &choiceExpr{
					pos: position{line: 389, col: 10, offset: 12872},
					alternatives: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 389, col: 10, offset: 12872},
							expr: &seqExpr{
								pos: position{line: 389, col: 12, offset: 12874},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 389, col: 12, offset: 12874},
										expr: &charClassMatcher{
											pos:        position{line: 389, col: 13, offset: 12875},
											val:        "[{}]",
											chars:      []rune{'{', '}'},
											ignoreCase: false,
											inverted:   false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 389, col: 18, offset: 12880},
										name: "LabelMatches",
									},
								},
							},
						},
						&seqExpr{
							pos: position{line: 389, col: 36, offset: 12898},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 389, col: 36, offset: 12898},
									val:        "{",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 389, col: 40, offset: 12902},
									name: "Code",
								},
								&litMatcher{
									pos:        position{line: 389, col: 45, offset: 12907},
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
			pos:  position{line: 391, col: 1, offset: 12915},
			expr: &choiceExpr{
				pos: position{line: 391, col: 13, offset: 12927},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 391, col: 13, offset: 12927},
						val:        "-",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 19, offset: 12933},
						val:        "+",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 25, offset: 12939},
						val:        "*",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 31, offset: 12945},
						val:        "%",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 37, offset: 12951},
						val:        "/",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 43, offset: 12957},
						val:        "==",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 50, offset: 12964},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 57, offset: 12971},
						val:        "<=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 64, offset: 12978},
						val:        "<",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 70, offset: 12984},
						val:        ">=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 77, offset: 12991},
						val:        ">",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 83, offset: 12997},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 90, offset: 13004},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 97, offset: 13011},
						val:        "^",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 391, col: 103, offset: 13017},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "LabelOperators",
			pos:  position{line: 392, col: 1, offset: 13021},
			expr: &choiceExpr{
				pos: position{line: 392, col: 19, offset: 13039},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 392, col: 19, offset: 13039},
						val:        "!=",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 392, col: 26, offset: 13046},
						val:        "=~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 392, col: 33, offset: 13053},
						val:        "!~",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 392, col: 40, offset: 13060},
						val:        "=",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "LabelMatch",
			pos:  position{line: 394, col: 1, offset: 13065},
			expr: &seqExpr{
				pos: position{line: 394, col: 14, offset: 13078},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 394, col: 14, offset: 13078},
						label: "label",
						expr: &ruleRefExpr{
							pos:  position{line: 394, col: 20, offset: 13084},
							name: "Identifier",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 394, col: 31, offset: 13095},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 394, col: 34, offset: 13098},
						label: "op",
						expr: &ruleRefExpr{
							pos:  position{line: 394, col: 37, offset: 13101},
							name: "LabelOperators",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 394, col: 52, offset: 13116},
						name: "__",
					},
					&labeledExpr{
						pos:   position{line: 394, col: 55, offset: 13119},
						label: "match",
						expr: &choiceExpr{
							pos: position{line: 394, col: 63, offset: 13127},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 394, col: 63, offset: 13127},
									name: "StringLiteral",
								},
								&ruleRefExpr{
									pos:  position{line: 394, col: 79, offset: 13143},
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
			pos:  position{line: 395, col: 1, offset: 13152},
			expr: &seqExpr{
				pos: position{line: 395, col: 16, offset: 13167},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 395, col: 16, offset: 13167},
						name: "LabelMatch",
					},
					&ruleRefExpr{
						pos:  position{line: 395, col: 27, offset: 13178},
						name: "__",
					},
					&zeroOrMoreExpr{
						pos: position{line: 395, col: 30, offset: 13181},
						expr: &seqExpr{
							pos: position{line: 395, col: 32, offset: 13183},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 395, col: 32, offset: 13183},
									val:        ",",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 395, col: 36, offset: 13187},
									name: "__",
								},
								&ruleRefExpr{
									pos:  position{line: 395, col: 39, offset: 13190},
									name: "LabelMatch",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "VectorSelector",
			pos:  position{line: 397, col: 1, offset: 13204},
			expr: &seqExpr{
				pos: position{line: 397, col: 18, offset: 13221},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 397, col: 18, offset: 13221},
						label: "metric",
						expr: &ruleRefExpr{
							pos:  position{line: 397, col: 25, offset: 13228},
							name: "Identifier",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 397, col: 36, offset: 13239},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 397, col: 40, offset: 13243},
						expr: &ruleRefExpr{
							pos:  position{line: 397, col: 40, offset: 13243},
							name: "CodeBlock",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 397, col: 51, offset: 13254},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 397, col: 54, offset: 13257},
						expr: &ruleRefExpr{
							pos:  position{line: 397, col: 54, offset: 13257},
							name: "Range",
						},
					},
					&ruleRefExpr{
						pos:  position{line: 397, col: 61, offset: 13264},
						name: "__",
					},
					&zeroOrOneExpr{
						pos: position{line: 397, col: 64, offset: 13267},
						expr: &ruleRefExpr{
							pos:  position{line: 397, col: 64, offset: 13267},
							name: "Offset",
						},
					},
				},
			},
		},
		{
			name: "Range",
			pos:  position{line: 399, col: 1, offset: 13276},
			expr: &actionExpr{
				pos: position{line: 399, col: 9, offset: 13284},
				run: (*parser).callonRange1,
				expr: &seqExpr{
					pos: position{line: 399, col: 9, offset: 13284},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 399, col: 9, offset: 13284},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 13, offset: 13288},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 16, offset: 13291},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 20, offset: 13295},
								name: "Duration",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 29, offset: 13304},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 399, col: 32, offset: 13307},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Offset",
			pos:  position{line: 403, col: 1, offset: 13336},
			expr: &actionExpr{
				pos: position{line: 403, col: 10, offset: 13345},
				run: (*parser).callonOffset1,
				expr: &seqExpr{
					pos: position{line: 403, col: 10, offset: 13345},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 403, col: 10, offset: 13345},
							val:        "offset",
							ignoreCase: true,
						},
						&ruleRefExpr{
							pos:  position{line: 403, col: 20, offset: 13355},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 403, col: 23, offset: 13358},
							label: "dur",
							expr: &ruleRefExpr{
								pos:  position{line: 403, col: 27, offset: 13362},
								name: "Duration",
							},
						},
					},
				},
			},
		},
		{
			name: "__",
			pos:  position{line: 407, col: 1, offset: 13396},
			expr: &zeroOrMoreExpr{
				pos: position{line: 407, col: 6, offset: 13401},
				expr: &choiceExpr{
					pos: position{line: 407, col: 8, offset: 13403},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 407, col: 8, offset: 13403},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 407, col: 21, offset: 13416},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 407, col: 27, offset: 13422},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 408, col: 1, offset: 13433},
			expr: &zeroOrMoreExpr{
				pos: position{line: 408, col: 5, offset: 13437},
				expr: &ruleRefExpr{
					pos:  position{line: 408, col: 5, offset: 13437},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 410, col: 1, offset: 13450},
			expr: &charClassMatcher{
				pos:        position{line: 410, col: 14, offset: 13463},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 411, col: 1, offset: 13471},
			expr: &litMatcher{
				pos:        position{line: 411, col: 7, offset: 13477},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 412, col: 1, offset: 13482},
			expr: &choiceExpr{
				pos: position{line: 412, col: 7, offset: 13488},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 412, col: 7, offset: 13488},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 412, col: 7, offset: 13488},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 412, col: 10, offset: 13491},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 412, col: 16, offset: 13497},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 412, col: 16, offset: 13497},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 412, col: 18, offset: 13499},
								expr: &ruleRefExpr{
									pos:  position{line: 412, col: 18, offset: 13499},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 412, col: 37, offset: 13518},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 412, col: 43, offset: 13524},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 412, col: 43, offset: 13524},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 412, col: 46, offset: 13527},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 414, col: 1, offset: 13532},
			expr: &notExpr{
				pos: position{line: 414, col: 7, offset: 13538},
				expr: &anyMatcher{
					line: 414, col: 8, offset: 13539,
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
