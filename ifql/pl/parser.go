package pl

import (
	"github.com/influxdata/ifql/ast"
)

func Parse(script string) (*ast.Program, error) {
	p := &parser{Buffer: script, Pretty: true}
	p.Init()
	if err := p.Parse(); err != nil {
		return nil, err
	}
	p.Execute()
	prog := p.pop().(*ast.Program)
	return prog, nil
}

type types struct {
	vstack []ast.Node
}

func (t *types) pop() ast.Node {
	l := len(t.vstack)
	n := t.vstack[l-1]
	t.vstack = t.vstack[:l-1]
	return n
}

func (t *types) push(n ast.Node) {
	t.vstack = append(t.vstack, n)
}

func (t *types) PushProgram() {
	body := t.pop().(ast.Statement)
	t.push(&ast.Program{
		Body: []ast.Statement{body},
	})
}
func (t *types) PushVariableDeclaration() {
	vd := t.pop().(*ast.VariableDeclarator)

	t.push(&ast.VariableDeclaration{
		Declarations: []*ast.VariableDeclarator{
			vd,
		},
	})
}

func (t *types) PushVariableDeclarator() {
	init := t.pop().(ast.Expression)
	id := t.pop().(*ast.Identifier)
	t.push(&ast.VariableDeclarator{
		ID:   id,
		Init: init,
	})
}
func (t *types) PushStringLiteral(v string) {
	t.vstack = append(t.vstack, &ast.StringLiteral{
		Value: v,
	})
}
func (t *types) PushIdentifier(name string) {
	t.vstack = append(t.vstack, &ast.Identifier{
		Name: name,
	})
}
