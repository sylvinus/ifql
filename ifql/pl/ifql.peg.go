package pl

//go:generate peg -switch -inline -strict ifql.peg

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleStart
	ruleProgram
	ruleSourceElements
	ruleSourceElement
	ruleStatement
	ruleVariableDeclaration
	ruleVariableDeclarator
	ruleExpr
	ruleStringLiteral
	ruleIdentifier
	rule__
	rulews
	ruleEOL
	ruleEOF
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
)

var rul3s = [...]string{
	"Unknown",
	"Start",
	"Program",
	"SourceElements",
	"SourceElement",
	"Statement",
	"VariableDeclaration",
	"VariableDeclarator",
	"Expr",
	"StringLiteral",
	"Identifier",
	"__",
	"ws",
	"EOL",
	"EOF",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type parser struct {
	types

	Buffer string
	buffer []rune
	rules  [21]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *parser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *parser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *parser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *parser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:

			p.PushProgram()

		case ruleAction1:

			p.PushVariableDeclaration()

		case ruleAction2:

			p.PushVariableDeclarator()

		case ruleAction3:

			p.PushStringLiteral(buffer[begin:end])

		case ruleAction4:

			p.PushIdentifier(buffer[begin:end])

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *parser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Start <- <(__ Program __ EOF Action0)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rule__]() {
					goto l0
				}
				{
					position2 := position
					{
						position3 := position
						if !_rules[rule__]() {
							goto l0
						}
						{
							position6 := position
							{
								position7 := position
								{
									position8 := position
									if !_rules[rule__]() {
										goto l0
									}
									if buffer[position] != rune('v') {
										goto l0
									}
									position++
									if buffer[position] != rune('a') {
										goto l0
									}
									position++
									if buffer[position] != rune('r') {
										goto l0
									}
									position++
									if !_rules[rule__]() {
										goto l0
									}
									{
										position9 := position
										{
											position10 := position
											{
												position11 := position
												if buffer[position] != rune('a') {
													goto l0
												}
												position++
												add(rulePegText, position11)
											}
											{
												add(ruleAction4, position)
											}
											add(ruleIdentifier, position10)
										}
										if !_rules[rule__]() {
											goto l0
										}
										if buffer[position] != rune('=') {
											goto l0
										}
										position++
										if !_rules[rule__]() {
											goto l0
										}
										{
											position13 := position
											{
												position14 := position
												{
													position15 := position
													{
														position18, tokenIndex18 := position, tokenIndex
														if c := buffer[position]; c < rune('0') || c > rune('9') {
															goto l19
														}
														position++
														goto l18
													l19:
														position, tokenIndex = position18, tokenIndex18
														if c := buffer[position]; c < rune('a') || c > rune('z') {
															goto l0
														}
														position++
													}
												l18:
												l16:
													{
														position17, tokenIndex17 := position, tokenIndex
														{
															position20, tokenIndex20 := position, tokenIndex
															if c := buffer[position]; c < rune('0') || c > rune('9') {
																goto l21
															}
															position++
															goto l20
														l21:
															position, tokenIndex = position20, tokenIndex20
															if c := buffer[position]; c < rune('a') || c > rune('z') {
																goto l17
															}
															position++
														}
													l20:
														goto l16
													l17:
														position, tokenIndex = position17, tokenIndex17
													}
													add(rulePegText, position15)
												}
												{
													add(ruleAction3, position)
												}
												add(ruleStringLiteral, position14)
											}
											add(ruleExpr, position13)
										}
										{
											add(ruleAction2, position)
										}
										add(ruleVariableDeclarator, position9)
									}
									if !_rules[rule__]() {
										goto l0
									}
									{
										add(ruleAction1, position)
									}
									add(ruleVariableDeclaration, position8)
								}
								add(ruleStatement, position7)
							}
							add(ruleSourceElement, position6)
						}
					l4:
						{
							position5, tokenIndex5 := position, tokenIndex
							if !_rules[rule__]() {
								goto l5
							}
							{
								position25 := position
								{
									position26 := position
									{
										position27 := position
										if !_rules[rule__]() {
											goto l5
										}
										if buffer[position] != rune('v') {
											goto l5
										}
										position++
										if buffer[position] != rune('a') {
											goto l5
										}
										position++
										if buffer[position] != rune('r') {
											goto l5
										}
										position++
										if !_rules[rule__]() {
											goto l5
										}
										{
											position28 := position
											{
												position29 := position
												{
													position30 := position
													if buffer[position] != rune('a') {
														goto l5
													}
													position++
													add(rulePegText, position30)
												}
												{
													add(ruleAction4, position)
												}
												add(ruleIdentifier, position29)
											}
											if !_rules[rule__]() {
												goto l5
											}
											if buffer[position] != rune('=') {
												goto l5
											}
											position++
											if !_rules[rule__]() {
												goto l5
											}
											{
												position32 := position
												{
													position33 := position
													{
														position34 := position
														{
															position37, tokenIndex37 := position, tokenIndex
															if c := buffer[position]; c < rune('0') || c > rune('9') {
																goto l38
															}
															position++
															goto l37
														l38:
															position, tokenIndex = position37, tokenIndex37
															if c := buffer[position]; c < rune('a') || c > rune('z') {
																goto l5
															}
															position++
														}
													l37:
													l35:
														{
															position36, tokenIndex36 := position, tokenIndex
															{
																position39, tokenIndex39 := position, tokenIndex
																if c := buffer[position]; c < rune('0') || c > rune('9') {
																	goto l40
																}
																position++
																goto l39
															l40:
																position, tokenIndex = position39, tokenIndex39
																if c := buffer[position]; c < rune('a') || c > rune('z') {
																	goto l36
																}
																position++
															}
														l39:
															goto l35
														l36:
															position, tokenIndex = position36, tokenIndex36
														}
														add(rulePegText, position34)
													}
													{
														add(ruleAction3, position)
													}
													add(ruleStringLiteral, position33)
												}
												add(ruleExpr, position32)
											}
											{
												add(ruleAction2, position)
											}
											add(ruleVariableDeclarator, position28)
										}
										if !_rules[rule__]() {
											goto l5
										}
										{
											add(ruleAction1, position)
										}
										add(ruleVariableDeclaration, position27)
									}
									add(ruleStatement, position26)
								}
								add(ruleSourceElement, position25)
							}
							goto l4
						l5:
							position, tokenIndex = position5, tokenIndex5
						}
						add(ruleSourceElements, position3)
					}
					add(ruleProgram, position2)
				}
				if !_rules[rule__]() {
					goto l0
				}
				{
					position44 := position
					{
						position45, tokenIndex45 := position, tokenIndex
						if !matchDot() {
							goto l45
						}
						goto l0
					l45:
						position, tokenIndex = position45, tokenIndex45
					}
					add(ruleEOF, position44)
				}
				{
					add(ruleAction0, position)
				}
				add(ruleStart, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Program <- <SourceElements> */
		nil,
		/* 2 SourceElements <- <(__ SourceElement)+> */
		nil,
		/* 3 SourceElement <- <Statement> */
		nil,
		/* 4 Statement <- <VariableDeclaration> */
		nil,
		/* 5 VariableDeclaration <- <(__ ('v' 'a' 'r') __ VariableDeclarator __ Action1)> */
		nil,
		/* 6 VariableDeclarator <- <(Identifier __ '=' __ Expr Action2)> */
		nil,
		/* 7 Expr <- <StringLiteral> */
		nil,
		/* 8 StringLiteral <- <(<([0-9] / [a-z])+> Action3)> */
		nil,
		/* 9 Identifier <- <(<'a'> Action4)> */
		nil,
		/* 10 __ <- <(ws / EOL)*> */
		func() bool {
			{
				position57 := position
			l58:
				{
					position59, tokenIndex59 := position, tokenIndex
					{
						position60, tokenIndex60 := position, tokenIndex
						{
							position62 := position
							{
								switch buffer[position] {
								case '\n':
									if buffer[position] != rune('\n') {
										goto l61
									}
									position++
									break
								case '\r':
									if buffer[position] != rune('\r') {
										goto l61
									}
									position++
									break
								case '\t':
									if buffer[position] != rune('\t') {
										goto l61
									}
									position++
									break
								default:
									if buffer[position] != rune(' ') {
										goto l61
									}
									position++
									break
								}
							}

							add(rulews, position62)
						}
						goto l60
					l61:
						position, tokenIndex = position60, tokenIndex60
						{
							position64 := position
							if buffer[position] != rune('\n') {
								goto l59
							}
							position++
							add(ruleEOL, position64)
						}
					}
				l60:
					goto l58
				l59:
					position, tokenIndex = position59, tokenIndex59
				}
				add(rule__, position57)
			}
			return true
		},
		/* 11 ws <- <((&('\n') '\n') | (&('\r') '\r') | (&('\t') '\t') | (&(' ') ' '))> */
		nil,
		/* 12 EOL <- <'\n'> */
		nil,
		/* 13 EOF <- <!.> */
		nil,
		/* 15 Action0 <- <{
		   p.PushProgram()
		 }> */
		nil,
		/* 16 Action1 <- <{
		    p.PushVariableDeclaration()
		}> */
		nil,
		/* 17 Action2 <- <{
		   p.PushVariableDeclarator()
		 }> */
		nil,
		nil,
		/* 19 Action3 <- <{
		   p.PushStringLiteral(buffer[begin:end])
		 }> */
		nil,
		/* 20 Action4 <- <{
		   p.PushIdentifier(buffer[begin:end])
		 }> */
		nil,
	}
	p.rules = _rules
}
