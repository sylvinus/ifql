package ifql

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/ifql/ast"
)

func program(body interface{}, text []byte, pos position) (*ast.Program, error) {
	return &ast.Program{
		Body:     body.([]ast.Statement),
		BaseNode: base(text, pos),
	}, nil
}

func srcElems(head, tails interface{}) ([]ast.Statement, error) {
	elems := []ast.Statement{head.(ast.Statement)}
	for _, tail := range toIfaceSlice(tails) {
		elem := toIfaceSlice(tail)[1] // Skip whitespace
		elems = append(elems, elem.(ast.Statement))
	}
	return elems, nil
}

func varstmt(declarations interface{}, text []byte, pos position) (*ast.VariableDeclaration, error) {
	return &ast.VariableDeclaration{
		Declarations: declarations.([]*ast.VariableDeclarator),
		BaseNode:     base(text, pos),
	}, nil
}

func vardecls(head, tails interface{}) ([]*ast.VariableDeclarator, error) {
	decls := []*ast.VariableDeclarator{head.(*ast.VariableDeclarator)}
	for _, tail := range toIfaceSlice(tails) {
		decl := toIfaceSlice(tail)[3] // Skip whitespace and comma
		decls = append(decls, decl.(*ast.VariableDeclarator))
	}
	return decls, nil
}

func vardecl(id, initializer interface{}, text []byte, pos position) (*ast.VariableDeclarator, error) {
	return &ast.VariableDeclarator{
		ID:   id.(*ast.Identifier),
		Init: initializer.(ast.Expression),
	}, nil
}

func exprstmt(call interface{}, text []byte, pos position) (*ast.ExpressionStatement, error) {
	return &ast.ExpressionStatement{
		Expression: call.(ast.Expression),
		BaseNode:   base(text, pos),
	}, nil
}

func callchain(callee, args, members interface{}, text []byte, pos position) (*ast.CallExpression, error) {
	res, err := call(callee, args, text, pos)
	if err != nil {
		return nil, err
	}

	for _, memexprs := range toIfaceSlice(members) {
		memexpr := toIfaceSlice(memexprs)
		if memexpr == nil {
			continue
		}

		m, err := member(res, memexpr[3], text, pos)
		if err != nil {
			return nil, err
		}

		res, err = call(m, memexpr[5], text, pos)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func call(callee, args interface{}, text []byte, pos position) (*ast.CallExpression, error) {
	c := &ast.CallExpression{
		Callee:   callee.(ast.Expression),
		BaseNode: base(text, pos),
	}
	if args != nil {
		c.Arguments = []ast.Expression{args.(*ast.ObjectExpression)}
	}
	return c, nil
}

func member(callee, property interface{}, text []byte, pos position) (*ast.MemberExpression, error) {
	return &ast.MemberExpression{
		Object:   callee.(*ast.CallExpression),
		Property: property.(*ast.Identifier),
		BaseNode: base(text, pos),
	}, nil
}

func object(first, rest interface{}, text []byte, pos position) (*ast.ObjectExpression, error) {
	props := []*ast.Property{first.(*ast.Property)}
	if rest != nil {
		for _, prop := range toIfaceSlice(rest) {
			props = append(props, prop.(*ast.Property))
		}
	}

	return &ast.ObjectExpression{
		Properties: props,
		BaseNode:   base(text, pos),
	}, nil
}

func property(key, value interface{}, text []byte, pos position) (*ast.Property, error) {
	return &ast.Property{
		Key:      key.(*ast.Identifier),
		Value:    value.(ast.Expression),
		BaseNode: base(text, pos),
	}, nil
}

func identifier(text []byte, pos position) (*ast.Identifier, error) {
	return &ast.Identifier{
		Name:     string(text),
		BaseNode: base(text, pos),
	}, nil
}

func logicalExpression(head, tails interface{}, text []byte, pos position) (ast.Expression, error) {
	res := head.(ast.Expression)
	for _, tail := range toIfaceSlice(tails) {
		right := toIfaceSlice(tail)
		res = &ast.LogicalExpression{
			Left:     res,
			Right:    right[3].(ast.Expression),
			Operator: right[1].(ast.LogicalOperatorKind),
			BaseNode: base(text, pos),
		}
	}
	return res, nil
}

func logicalOp(text []byte) (ast.LogicalOperatorKind, error) {
	return ast.LogicalOperatorLookup(strings.ToLower(string(text))), nil
}

func binaryExpression(head, tails interface{}, text []byte, pos position) (ast.Expression, error) {
	res := head.(ast.Expression)
	for _, tail := range toIfaceSlice(tails) {
		right := toIfaceSlice(tail)
		res = &ast.BinaryExpression{
			Left:     res,
			Right:    right[3].(ast.Expression),
			Operator: right[1].(ast.OperatorKind),
			BaseNode: base(text, pos),
		}
	}
	return res, nil
}

func binaryOp(text []byte) (ast.OperatorKind, error) {
	return ast.OperatorLookup(strings.ToLower(string(text))), nil
}

func stringLiteral(text []byte, pos position) (*ast.StringLiteral, error) {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		return nil, err
	}
	return &ast.StringLiteral{
		BaseNode: base(text, pos),
		Value:    s,
	}, nil
}

func numberLiteral(text []byte, pos position) (*ast.NumberLiteral, error) {
	n, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return nil, err
	}
	return &ast.NumberLiteral{
		BaseNode: base(text, pos),
		Value:    n,
	}, nil
}

func fieldLiteral(text []byte, pos position) (*ast.FieldLiteral, error) {
	return &ast.FieldLiteral{
		BaseNode: base(text, pos),
		Value:    "_field",
	}, nil
}

func regexLiteral(chars interface{}, text []byte, pos position) (*ast.RegexpLiteral, error) {
	var regex string
	for _, char := range toIfaceSlice(chars) {
		regex += char.(string)
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return &ast.RegexpLiteral{
		BaseNode: base(text, pos),
		Value:    r,
	}, nil
}

func durationLiteral(text []byte, pos position) (*ast.DurationLiteral, error) {
	d, err := time.ParseDuration(string(text))
	if err != nil {
		return nil, err
	}
	return &ast.DurationLiteral{
		BaseNode: base(text, pos),
		Value:    d,
	}, nil
}

func datetime(text []byte, pos position) (*ast.DateTimeLiteral, error) {
	t, err := time.Parse(time.RFC3339Nano, string(text))
	if err != nil {
		return nil, err
	}
	return &ast.DateTimeLiteral{
		BaseNode: base(text, pos),
		Value:    t,
	}, nil
}

func base(text []byte, pos position) *ast.BaseNode {
	return nil
	return &ast.BaseNode{
		Loc: &ast.SourceLocation{
			Start: ast.Position{
				Line:   pos.line,
				Column: pos.col,
			},
			End: ast.Position{
				Line:   pos.line,
				Column: pos.col + len(text),
			},
			Source: source(text),
		},
	}
}

func source(text []byte) *string {
	str := string(text)
	return &str
}
