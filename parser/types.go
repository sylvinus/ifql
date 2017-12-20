package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/pkg/errors"
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func program(pkg, imports, body interface{}, text []byte, pos position) (*ast.Program, error) {
	var pkgDecl *ast.PackageDeclaration
	if pkg != nil {
		pkgDecl = pkg.(*ast.PackageDeclaration)
	}
	return &ast.Program{
		Package:  pkgDecl,
		Imports:  imports.([]*ast.ImportDeclaration),
		Body:     body.([]ast.Statement),
		BaseNode: base(text, pos),
	}, nil
}

func packageDecl(name interface{}, text []byte, pos position) (*ast.PackageDeclaration, error) {
	return &ast.PackageDeclaration{
		ID:       name.(*ast.Identifier),
		BaseNode: base(text, pos),
	}, nil
}

func imports(imprts interface{}) ([]*ast.ImportDeclaration, error) {
	list := toIfaceSlice(imprts)
	if len(list) == 0 {
		return nil, nil
	}
	elems := make([]*ast.ImportDeclaration, len(list))
	for i, s := range list {
		elems[i] = toIfaceSlice(s)[1].(*ast.ImportDeclaration)
	}
	return elems, nil
}

func importDecl(path, version, as interface{}, text []byte, pos position) (*ast.ImportDeclaration, error) {
	var asIdent *ast.Identifier
	if as != nil {
		asIdent = toIfaceSlice(as)[2].(*ast.Identifier)
	}
	var versionDecl *ast.VersionDeclaration
	if version != nil {
		versionDecl = version.(*ast.VersionDeclaration)
	}
	return &ast.ImportDeclaration{
		Path:     path.(*ast.StringLiteral),
		Version:  versionDecl,
		As:       asIdent,
		BaseNode: base(text, pos),
	}, nil
}

func versionDecl(op, num interface{}, text []byte, pos position) (*ast.VersionDeclaration, error) {
	return &ast.VersionDeclaration{
		Operator: op.(ast.VersionOperatorKind),
		Number:   num.(*ast.VersionNumber),
		BaseNode: base(text, pos),
	}, nil
}

func versionNumber(text []byte, pos position) (*ast.VersionNumber, error) {
	vStr := string(text)
	if vStr[0] == 'v' {
		vStr = vStr[1:]
	}
	var version [3]int
	for i := range version {
		n := strings.IndexRune(vStr, '.')
		if n == -1 {
			n = len(vStr)
		}
		v, err := strconv.ParseInt(vStr[:n], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "invalid version")
		}
		version[i] = int(v)
		if n < len(vStr) {
			vStr = vStr[n+1:]
		}
	}

	return &ast.VersionNumber{
		Literal:  string(text),
		Major:    version[0],
		Minor:    version[1],
		Patch:    version[2],
		BaseNode: base(text, pos),
	}, nil
}

func versionOp(text []byte) (ast.VersionOperatorKind, error) {
	return ast.VersionOperatorLookup(string(text)), nil
}

func statements(stmts interface{}) ([]ast.Statement, error) {
	list := toIfaceSlice(stmts)
	elems := make([]ast.Statement, len(list))
	for i, s := range list {
		elems[i] = toIfaceSlice(s)[1].(ast.Statement)
	}
	return elems, nil
}

func blockstmt(body interface{}, text []byte, pos position) (*ast.BlockStatement, error) {
	bodySlice := toIfaceSlice(body)
	statements := make([]ast.Statement, len(bodySlice))
	for i, s := range bodySlice {
		stmt := toIfaceSlice(s)[1] // Skip whitespace
		statements[i] = stmt.(ast.Statement)
	}
	return &ast.BlockStatement{
		BaseNode: base(text, pos),
		Body:     statements,
	}, nil
}

func varstmt(declaration interface{}, text []byte, pos position) (*ast.VariableDeclaration, error) {
	return &ast.VariableDeclaration{
		Declarations: []*ast.VariableDeclarator{declaration.(*ast.VariableDeclarator)},
		BaseNode:     base(text, pos),
	}, nil
}

func vardecl(id, initializer interface{}, text []byte, pos position) (*ast.VariableDeclarator, error) {
	return &ast.VariableDeclarator{
		ID:   id.(*ast.Identifier),
		Init: initializer.(ast.Expression),
	}, nil
}

func exprstmt(expr interface{}, text []byte, pos position) (*ast.ExpressionStatement, error) {
	return &ast.ExpressionStatement{
		Expression: expr.(ast.Expression),
		BaseNode:   base(text, pos),
	}, nil
}

func returnstmt(argument interface{}, text []byte, pos position) (*ast.ReturnStatement, error) {
	return &ast.ReturnStatement{
		BaseNode: base(text, pos),
		Argument: argument.(ast.Expression),
	}, nil
}

func memberexprs(head, tail interface{}, text []byte, pos position) (ast.Expression, error) {
	res := head.(ast.Expression)
	for _, prop := range toIfaceSlice(tail) {
		res = &ast.MemberExpression{
			Object:   res,
			Property: prop.(ast.Expression),
			BaseNode: base(text, pos),
		}
	}
	return res, nil
}

func memberexpr(object, property interface{}, text []byte, pos position) (*ast.MemberExpression, error) {
	m := &ast.MemberExpression{
		BaseNode: base(text, pos),
	}

	if object != nil {
		m.Object = object.(ast.Expression)
	}

	if property != nil {
		m.Property = property.(*ast.Identifier)
	}

	return m, nil
}

func callexpr(callee, args interface{}, text []byte, pos position) (*ast.CallExpression, error) {
	c := &ast.CallExpression{
		BaseNode: base(text, pos),
	}

	if callee != nil {
		c.Callee = callee.(ast.Expression)
	}

	if args != nil {
		c.Arguments = []ast.Expression{args.(*ast.ObjectExpression)}
	}
	return c, nil
}

func callexprs(head, tail interface{}, text []byte, pos position) (ast.Expression, error) {
	expr := head.(ast.Expression)
	for _, i := range toIfaceSlice(tail) {
		switch elem := i.(type) {
		case *ast.CallExpression:
			elem.Callee = expr
			expr = elem
		case *ast.MemberExpression:
			elem.Object = expr
			expr = elem
		}
	}
	return expr, nil
}

func arrowfunc(params interface{}, body interface{}, text []byte, pos position) *ast.ArrowFunctionExpression {
	paramsSlice := toIfaceSlice(params)
	paramsList := make([]*ast.Property, len(paramsSlice))
	for i, p := range paramsSlice {
		paramsList[i] = p.(*ast.Property)
	}
	return &ast.ArrowFunctionExpression{
		BaseNode: base(text, pos),
		Params:   paramsList,
		Body:     body.(ast.Node),
	}
}

func objectexpr(first, rest interface{}, text []byte, pos position) (*ast.ObjectExpression, error) {
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
	var v ast.Expression
	if value != nil {
		v = value.(ast.Expression)
	}
	return &ast.Property{
		Key:      key.(*ast.Identifier),
		Value:    v,
		BaseNode: base(text, pos),
	}, nil
}

func identifier(text []byte, pos position) (*ast.Identifier, error) {
	return &ast.Identifier{
		Name:     string(text),
		BaseNode: base(text, pos),
	}, nil
}

func array(first, rest interface{}, text []byte, pos position) *ast.ArrayExpression {
	var elements []ast.Expression
	if first != nil {
		elements = append(elements, first.(ast.Expression))
	}
	if rest != nil {
		for _, el := range rest.([]interface{}) {
			elements = append(elements, el.(ast.Expression))
		}
	}
	return &ast.ArrayExpression{
		Elements: elements,
		BaseNode: base(text, pos),
	}
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

func unaryExpression(op, argument interface{}, text []byte, pos position) (*ast.UnaryExpression, error) {
	return &ast.UnaryExpression{
		Operator: op.(ast.OperatorKind),
		Argument: argument.(ast.Expression),
		BaseNode: base(text, pos),
	}, nil
}

func operator(text []byte) (ast.OperatorKind, error) {
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

func booleanLiteral(b bool, text []byte, pos position) (*ast.BooleanLiteral, error) {
	return &ast.BooleanLiteral{
		BaseNode: base(text, pos),
		Value:    b,
	}, nil
}

func integerLiteral(text []byte, pos position) (*ast.IntegerLiteral, error) {
	n, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return nil, err
	}
	return &ast.IntegerLiteral{
		BaseNode: base(text, pos),
		Value:    n,
	}, nil
}

func numberLiteral(text []byte, pos position) (*ast.FloatLiteral, error) {
	n, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return nil, err
	}
	return &ast.FloatLiteral{
		BaseNode: base(text, pos),
		Value:    n,
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
