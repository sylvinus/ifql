package semantic

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/ifql/ast"
)

type Node interface {
	node()
	NodeType() string
	Copy() Node

	json.Marshaler
}

func (*Program) node() {}

func (*BlockStatement) node()      {}
func (*ExpressionStatement) node() {}
func (*ReturnStatement) node()     {}
func (*VariableDeclaration) node() {}

func (*ArrayExpression) node()         {}
func (*ArrowFunctionExpression) node() {}
func (*BinaryExpression) node()        {}
func (*CallExpression) node()          {}
func (*ConditionalExpression) node()   {}
func (*IdentifierExpression) node()    {}
func (*LogicalExpression) node()       {}
func (*MemberExpression) node()        {}
func (*ObjectExpression) node()        {}
func (*UnaryExpression) node()         {}

func (*Identifier) node()    {}
func (*Property) node()      {}
func (*FunctionParam) node() {}

func (*BooleanLiteral) node()         {}
func (*DateTimeLiteral) node()        {}
func (*DurationLiteral) node()        {}
func (*FloatLiteral) node()           {}
func (*IntegerLiteral) node()         {}
func (*RegexpLiteral) node()          {}
func (*StringLiteral) node()          {}
func (*UnsignedIntegerLiteral) node() {}

type Statement interface {
	Node
	stmt()
}

func (*BlockStatement) stmt()      {}
func (*ExpressionStatement) stmt() {}
func (*ReturnStatement) stmt()     {}
func (*VariableDeclaration) stmt() {}

type Expression interface {
	Node
	Type() Type
	expression()
}

func (*ArrayExpression) expression()         {}
func (*ArrowFunctionExpression) expression() {}
func (*BinaryExpression) expression()        {}
func (*BooleanLiteral) expression()          {}
func (*CallExpression) expression()          {}
func (*ConditionalExpression) expression()   {}
func (*DateTimeLiteral) expression()         {}
func (*DurationLiteral) expression()         {}
func (*FloatLiteral) expression()            {}
func (*IdentifierExpression) expression()    {}
func (*IntegerLiteral) expression()          {}
func (*LogicalExpression) expression()       {}
func (*MemberExpression) expression()        {}
func (*ObjectExpression) expression()        {}
func (*RegexpLiteral) expression()           {}
func (*StringLiteral) expression()           {}
func (*UnaryExpression) expression()         {}
func (*UnsignedIntegerLiteral) expression()  {}

type Literal interface {
	Expression
	literal()
}

func (*BooleanLiteral) literal()         {}
func (*DateTimeLiteral) literal()        {}
func (*DurationLiteral) literal()        {}
func (*FloatLiteral) literal()           {}
func (*IntegerLiteral) literal()         {}
func (*RegexpLiteral) literal()          {}
func (*StringLiteral) literal()          {}
func (*UnsignedIntegerLiteral) literal() {}

type Program struct {
	Body []Statement `json:"body"`
}

func (*Program) NodeType() string { return "Program" }

func (p *Program) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Program)
	*np = *p

	if len(p.Body) > 0 {
		np.Body = make([]Statement, len(p.Body))
		for i, s := range p.Body {
			np.Body[i] = s.Copy().(Statement)
		}
	}

	return np
}

type BlockStatement struct {
	Body []Statement `json:"body"`
}

func (*BlockStatement) NodeType() string { return "BlockStatement" }

func (s *BlockStatement) ReturnStatement() *ReturnStatement {
	return s.Body[len(s.Body)-1].(*ReturnStatement)
}

func (s *BlockStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(BlockStatement)
	*ns = *s

	if len(s.Body) > 0 {
		ns.Body = make([]Statement, len(s.Body))
		for i, stmt := range s.Body {
			ns.Body[i] = stmt.Copy().(Statement)
		}
	}

	return ns
}

type ExpressionStatement struct {
	Expression Expression `json:"expression"`
}

func (*ExpressionStatement) NodeType() string { return "ExpressionStatement" }

func (s *ExpressionStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ExpressionStatement)
	*ns = *s

	ns.Expression = s.Expression.Copy().(Expression)

	return ns
}

type ReturnStatement struct {
	Argument Expression `json:"argument"`
}

func (*ReturnStatement) NodeType() string { return "ReturnStatement" }

func (s *ReturnStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ReturnStatement)
	*ns = *s

	ns.Argument = s.Argument.Copy().(Expression)

	return ns
}

type VariableDeclaration struct {
	ID   *Identifier `json:"id"`
	Init Expression  `json:"init"`
	// Uses is a list of every place this identifier is used.
	// TODO(nathanielc): Do we need this?
	//Uses []Expression
}

func (*VariableDeclaration) NodeType() string { return "VariableDeclaration" }

func (s *VariableDeclaration) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(VariableDeclaration)
	*ns = *s

	if s.Init != nil {
		ns.Init = s.Init.Copy().(Expression)
	}

	return ns
}

type ArrayExpression struct {
	Elements []Expression `json:"elements"`
}

func (*ArrayExpression) NodeType() string { return "ArrayExpression" }
func (e *ArrayExpression) Type() Type     { return e }
func (e *ArrayExpression) Kind() Kind     { return KArray }
func (e *ArrayExpression) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get property of kind %s", e.Kind()))
}
func (e *ArrayExpression) ElementType() Type {
	if len(e.Elements) == 0 {
		// TODO(nathanielc): How do we do type inference for an empty array?
		return KInvalid
	}
	return e.Elements[0].Type()
}

func (e *ArrayExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ArrayExpression)
	*ne = *e

	if len(e.Elements) > 0 {
		ne.Elements = make([]Expression, len(e.Elements))
		for i, elem := range e.Elements {
			ne.Elements[i] = elem.Copy().(Expression)
		}
	}

	return ne
}

type ArrowFunctionExpression struct {
	Params []*FunctionParam `json:"params"`
	Body   Node             `json:"body"`
}

func (*ArrowFunctionExpression) NodeType() string { return "ArrowFunctionExpression" }
func (e *ArrowFunctionExpression) Type() Type {
	switch b := e.Body.(type) {
	case Expression:
		return b.Type()
	case *BlockStatement:
		rs := b.ReturnStatement()
		return rs.Argument.Type()
	default:
		return KInvalid
	}
}

func (e *ArrowFunctionExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ArrowFunctionExpression)
	*ne = *e

	if len(e.Params) > 0 {
		ne.Params = make([]*FunctionParam, len(e.Params))
		for i, p := range e.Params {
			ne.Params[i] = p.Copy().(*FunctionParam)
		}
	}
	ne.Body = e.Body.Copy()

	return ne
}

type FunctionParam struct {
	Key     *Identifier `json:"key"`
	Default Literal     `json:"default"`
}

func (*FunctionParam) NodeType() string { return "FunctionParam" }

func (p *FunctionParam) Copy() Node {
	if p == nil {
		return p
	}
	np := new(FunctionParam)
	*np = *p

	np.Key = p.Key.Copy().(*Identifier)
	if np.Default != nil {
		np.Default = p.Default.Copy().(Literal)
	}

	return np
}

type BinaryExpression struct {
	Operator ast.OperatorKind `json:"operator"`
	Left     Expression       `json:"left"`
	Right    Expression       `json:"right"`
}

func (*BinaryExpression) NodeType() string { return "BinaryExpression" }
func (e *BinaryExpression) Type() Type {
	return binaryTypesLookup[binarySignature{
		operator: e.Operator,
		left:     e.Left.Type().Kind(),
		right:    e.Right.Type().Kind(),
	}]
}

func (e *BinaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(BinaryExpression)
	*ne = *e

	ne.Left = e.Left.Copy().(Expression)
	ne.Right = e.Right.Copy().(Expression)

	return ne
}

type CallExpression struct {
	Callee    Expression        `json:"callee"`
	Arguments *ObjectExpression `json:"arguments"`
}

func (*CallExpression) NodeType() string { return "CallExpression" }
func (e *CallExpression) Type() Type {
	return e.Callee.Type()
}

func (e *CallExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(CallExpression)
	*ne = *e

	ne.Callee = e.Callee.Copy().(Expression)
	ne.Arguments = e.Arguments.Copy().(*ObjectExpression)

	return ne
}

type ConditionalExpression struct {
	Test       Expression `json:"test"`
	Alternate  Expression `json:"alternate"`
	Consequent Expression `json:"consequent"`
}

func (*ConditionalExpression) NodeType() string { return "ConditionalExpression" }

func (e *ConditionalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ConditionalExpression)
	*ne = *e

	ne.Test = e.Test.Copy().(Expression)
	ne.Alternate = e.Alternate.Copy().(Expression)
	ne.Consequent = e.Consequent.Copy().(Expression)

	return ne
}

type LogicalExpression struct {
	Operator ast.LogicalOperatorKind `json:"operator"`
	Left     Expression              `json:"left"`
	Right    Expression              `json:"right"`
}

func (*LogicalExpression) NodeType() string { return "LogicalExpression" }
func (*LogicalExpression) Type() Type       { return KBool }

func (e *LogicalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(LogicalExpression)
	*ne = *e

	ne.Left = e.Left.Copy().(Expression)
	ne.Right = e.Right.Copy().(Expression)

	return ne
}

type MemberExpression struct {
	Object   Expression `json:"object"`
	Property string     `json:"property"`
}

func (*MemberExpression) NodeType() string { return "MemberExpression" }

func (e *MemberExpression) Type() Type {
	return e.Object.Type().PropertyType(e.Property)
}

func (e *MemberExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(MemberExpression)
	*ne = *e

	ne.Object = e.Object.Copy().(Expression)

	return ne
}

type ObjectExpression struct {
	Properties []*Property `json:"properties"`
}

func (*ObjectExpression) NodeType() string { return "ObjectExpression" }
func (e *ObjectExpression) Type() Type     { return e }
func (e *ObjectExpression) Kind() Kind     { return KMap }
func (e *ObjectExpression) PropertyType(name string) Type {
	for _, p := range e.Properties {
		if p.Key.Name == name {
			return p.Value.Type()
		}
	}
	return KInvalid
}
func (e *ObjectExpression) ElementType() Type {
	panic(errors.New("map expressions do not have a single element type"))
}

func (e *ObjectExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ObjectExpression)
	*ne = *e

	if len(e.Properties) > 0 {
		ne.Properties = make([]*Property, len(e.Properties))
		for i, prop := range e.Properties {
			ne.Properties[i] = prop.Copy().(*Property)
		}
	}

	return ne
}

type UnaryExpression struct {
	Operator ast.OperatorKind `json:"operator"`
	Argument Expression       `json:"argument"`
}

func (*UnaryExpression) NodeType() string { return "UnaryExpression" }
func (e *UnaryExpression) Type() Type {
	return e.Argument.Type()
}

func (e *UnaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(UnaryExpression)
	*ne = *e

	ne.Argument = e.Argument.Copy().(Expression)

	return ne
}

type Property struct {
	Key   *Identifier `json:"key"`
	Value Expression  `json:"value"`
}

func (*Property) NodeType() string { return "Property" }

func (p *Property) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Property)
	*np = *p

	np.Value = p.Value.Copy().(Expression)

	return np
}

type IdentifierExpression struct {
	Name string `json:"name"`
	// Declaration is the node that declares this identifier
	Declaration *VariableDeclaration `json:"declaration,omitempty"`
}

func (*IdentifierExpression) NodeType() string { return "IdentifierExpression" }

func (e *IdentifierExpression) Type() Type {
	return e.Declaration.Init.Type()
}

func (e *IdentifierExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(IdentifierExpression)
	*ne = *e

	ne.Declaration = e.Declaration.Copy().(*VariableDeclaration)

	return ne
}

type Identifier struct {
	Name string `json:"name"`
}

func (*Identifier) NodeType() string { return "Identifier" }

func (i *Identifier) Copy() Node {
	if i == nil {
		return i
	}
	ni := new(Identifier)
	*ni = *i

	return ni
}

type BooleanLiteral struct {
	Value bool `json:"value"`
}

func (*BooleanLiteral) NodeType() string { return "BooleanLiteral" }
func (*BooleanLiteral) Type() Type       { return KBool }

func (l *BooleanLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(BooleanLiteral)
	*nl = *l

	return nl
}

type DateTimeLiteral struct {
	Value time.Time `json:"value"`
}

func (*DateTimeLiteral) NodeType() string { return "DateTimeLiteral" }
func (*DateTimeLiteral) Type() Type       { return KTime }

func (l *DateTimeLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DateTimeLiteral)
	*nl = *l

	return nl
}

type DurationLiteral struct {
	Value time.Duration `json:"value"`
}

func (*DurationLiteral) NodeType() string { return "DurationLiteral" }
func (*DurationLiteral) Type() Type       { return KDuration }

func (l *DurationLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DurationLiteral)
	*nl = *l

	return nl
}

type IntegerLiteral struct {
	Value int64 `json:"value"`
}

func (*IntegerLiteral) NodeType() string { return "IntegerLiteral" }
func (*IntegerLiteral) Type() Type       { return KInt }

func (l *IntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(IntegerLiteral)
	*nl = *l

	return nl
}

type FloatLiteral struct {
	Value float64 `json:"value"`
}

func (*FloatLiteral) NodeType() string { return "FloatLiteral" }
func (*FloatLiteral) Type() Type       { return KFloat }

func (l *FloatLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(FloatLiteral)
	*nl = *l

	return nl
}

type RegexpLiteral struct {
	Value *regexp.Regexp `json:"value"`
}

func (*RegexpLiteral) NodeType() string { return "RegexpLiteral" }
func (*RegexpLiteral) Type() Type       { return KRegex }

func (l *RegexpLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(RegexpLiteral)
	*nl = *l

	nl.Value = l.Value.Copy()

	return nl
}

type StringLiteral struct {
	Value string `json:"value"`
}

func (*StringLiteral) NodeType() string { return "StringLiteral" }
func (*StringLiteral) Type() Type       { return KString }

func (l *StringLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(StringLiteral)
	*nl = *l

	return nl
}

type UnsignedIntegerLiteral struct {
	Value uint64 `json:"value"`
}

func (*UnsignedIntegerLiteral) NodeType() string { return "UnsignedIntegerLiteral" }
func (*UnsignedIntegerLiteral) Type() Type       { return KUInt }

func (l *UnsignedIntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(UnsignedIntegerLiteral)
	*nl = *l

	return nl
}

func New(prog *ast.Program) (*Program, error) {
	return analyzeProgram(prog)
}

type declarationScope map[string]*VariableDeclaration

func (s declarationScope) Copy() declarationScope {
	cpy := make(declarationScope, len(s))
	for k, v := range s {
		cpy[k] = v
	}
	return cpy
}

func analyzeProgram(prog *ast.Program) (*Program, error) {
	declarations := make(declarationScope)
	p := &Program{
		Body: make([]Statement, len(prog.Body)),
	}
	for i, s := range prog.Body {
		n, err := analyzeStatment(s, declarations)
		if err != nil {
			return nil, err
		}
		p.Body[i] = n
	}
	return p, nil
}

func analyzeNode(n ast.Node, declarations declarationScope) (Node, error) {
	switch n := n.(type) {
	case ast.Statement:
		return analyzeStatment(n, declarations)
	case ast.Expression:
		return analyzeExpression(n, declarations)
	default:
		return nil, fmt.Errorf("unsupported node %T", n)
	}
}

func analyzeStatment(s ast.Statement, declarations declarationScope) (Statement, error) {
	switch s := s.(type) {
	case *ast.BlockStatement:
		return analyzeBlockStatement(s, declarations)
	case *ast.ExpressionStatement:
		return analyzeExpressionStatement(s, declarations)
	case *ast.ReturnStatement:
		return analyzeReturnStatement(s, declarations)
	case *ast.VariableDeclaration:
		// Expect a single declaration
		if len(s.Declarations) != 1 {
			return nil, fmt.Errorf("only single variable declarations are supported, found %d declarations", len(s.Declarations))
		}
		return analyzeVariableDeclaration(s.Declarations[0], declarations)
	default:
		return nil, fmt.Errorf("unsupported statement %T", s)
	}
}

func analyzeBlockStatement(block *ast.BlockStatement, declarations declarationScope) (*BlockStatement, error) {
	declarations = declarations.Copy()
	b := &BlockStatement{
		Body: make([]Statement, len(block.Body)),
	}
	for i, s := range block.Body {
		n, err := analyzeStatment(s, declarations)
		if err != nil {
			return nil, err
		}
		b.Body[i] = n
	}
	last := len(b.Body) - 1
	if _, ok := b.Body[last].(*ReturnStatement); !ok {
		return nil, errors.New("missing return statement in block")
	}
	return b, nil
}

func analyzeExpressionStatement(expr *ast.ExpressionStatement, declarations declarationScope) (*ExpressionStatement, error) {
	e, err := analyzeExpression(expr.Expression, declarations)
	if err != nil {
		return nil, err
	}
	return &ExpressionStatement{
		Expression: e,
	}, nil
}

func analyzeReturnStatement(ret *ast.ReturnStatement, declarations declarationScope) (*ReturnStatement, error) {
	arg, err := analyzeExpression(ret.Argument, declarations)
	if err != nil {
		return nil, err
	}
	return &ReturnStatement{
		Argument: arg,
	}, nil
}

func analyzeVariableDeclaration(decl *ast.VariableDeclarator, declarations declarationScope) (*VariableDeclaration, error) {
	id, err := analyzeIdentifier(decl.ID, declarations)
	if err != nil {
		return nil, err
	}
	init, err := analyzeExpression(decl.Init, declarations)
	if err != nil {
		return nil, err
	}
	vd := &VariableDeclaration{
		ID:   id,
		Init: init,
	}
	declarations[vd.ID.Name] = vd
	return vd, nil
}

func analyzeExpression(expr ast.Expression, declarations declarationScope) (Expression, error) {
	switch expr := expr.(type) {
	case *ast.ArrowFunctionExpression:
		return analyzeArrowFunctionExpression(expr, declarations)
	case *ast.CallExpression:
		return analyzeCallExpression(expr, declarations)
	case *ast.MemberExpression:
		return analyzeMemberExpression(expr, declarations)
	case *ast.BinaryExpression:
		return analyzeBinaryExpression(expr, declarations)
	case *ast.UnaryExpression:
		return analyzeUnaryExpression(expr, declarations)
	case *ast.LogicalExpression:
		return analyzeLogicalExpression(expr, declarations)
	case *ast.ObjectExpression:
		return analyzeObjectExpression(expr, declarations)
	case *ast.ArrayExpression:
		return analyzeArrayExpression(expr, declarations)
	case *ast.Identifier:
		return analyzeIdentifierExpression(expr, declarations)
	case ast.Literal:
		return analyzeLiteral(expr, declarations)
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func analyzeLiteral(lit ast.Literal, declarations declarationScope) (Literal, error) {
	switch lit := lit.(type) {
	case *ast.StringLiteral:
		return analyzeStringLiteral(lit, declarations)
	case *ast.BooleanLiteral:
		return analyzeBooleanLiteral(lit, declarations)
	case *ast.FloatLiteral:
		return analyzeFloatLiteral(lit, declarations)
	case *ast.IntegerLiteral:
		return analyzeIntegerLiteral(lit, declarations)
	case *ast.UnsignedIntegerLiteral:
		return analyzeUnsignedIntegerLiteral(lit, declarations)
	case *ast.RegexpLiteral:
		return analyzeRegexpLiteral(lit, declarations)
	case *ast.DurationLiteral:
		return analyzeDurationLiteral(lit, declarations)
	case *ast.DateTimeLiteral:
		return analyzeDateTimeLiteral(lit, declarations)
	default:
		return nil, fmt.Errorf("unsupported literal %T", lit)
	}
}

func analyzeArrowFunctionExpression(arrow *ast.ArrowFunctionExpression, declarations declarationScope) (*ArrowFunctionExpression, error) {
	declarations = declarations.Copy()
	b, err := analyzeNode(arrow.Body, declarations)
	if err != nil {
		return nil, err
	}
	f := &ArrowFunctionExpression{
		Params: make([]*FunctionParam, len(arrow.Params)),
		Body:   b,
	}
	for i, p := range arrow.Params {
		key, err := analyzeIdentifier(p.Key, declarations)
		if err != nil {
			return nil, err
		}
		var def Literal
		if p.Value != nil {
			lit, ok := p.Value.(ast.Literal)
			if !ok {
				return nil, fmt.Errorf("function parameter %q default value is not a literal", p.Key.Name)
			}
			var err error
			def, err = analyzeLiteral(lit, declarations)
			if err != nil {
				return nil, err
			}
		}

		f.Params[i] = &FunctionParam{
			Key:     key,
			Default: def,
		}
	}
	return f, nil
}

func analyzeCallExpression(call *ast.CallExpression, declarations declarationScope) (*CallExpression, error) {
	callee, err := analyzeExpression(call.Callee, declarations)
	if err != nil {
		return nil, err
	}
	var args *ObjectExpression
	if l := len(call.Arguments); l > 1 {
		return nil, fmt.Errorf("arguments are not a single object expression %v", args)
	} else if l == 1 {
		obj, ok := call.Arguments[0].(*ast.ObjectExpression)
		if !ok {
			return nil, fmt.Errorf("arguments not an object expression")
		}
		var err error
		args, err = analyzeObjectExpression(obj, declarations)
		if err != nil {
			return nil, err
		}
	}

	expr := &CallExpression{
		Callee:    callee,
		Arguments: args,
	}

	// Traverse Callee and update IdentifierExpressions with declarations
	declarations = declarations.Copy()
	for _, arg := range args.Properties {
		declarations[arg.Key.Name] = &VariableDeclaration{
			ID:   arg.Key,
			Init: arg.Value,
		}
	}

	v := &callExpressionVisitor{
		expr:         expr,
		declarations: declarations,
	}
	Walk(v, expr.Callee)
	return expr, nil
}

type callExpressionVisitor struct {
	expr         *CallExpression
	declarations declarationScope
}

func (v *callExpressionVisitor) Visit(n Node) Visitor {
	if ident, ok := n.(*IdentifierExpression); ok {
		if ident.Declaration == nil {
			ident.Declaration = v.declarations[ident.Name]
			// No need to walk further down this branch
			return nil
		}
	}
	return v
}
func (v *callExpressionVisitor) Done() {}

func analyzeMemberExpression(member *ast.MemberExpression, declarations declarationScope) (*MemberExpression, error) {
	obj, err := analyzeExpression(member.Object, declarations)
	if err != nil {
		return nil, err
	}

	var propertyName string
	switch p := member.Property.(type) {
	case *ast.Identifier:
		propertyName = p.Name
	case *ast.StringLiteral:
		propertyName = p.Value
	default:
		return nil, fmt.Errorf("unsupported member property expression of type %T", member.Property)
	}

	return &MemberExpression{
		Object:   obj,
		Property: propertyName,
	}, nil
}

func analyzeBinaryExpression(binary *ast.BinaryExpression, declarations declarationScope) (*BinaryExpression, error) {
	left, err := analyzeExpression(binary.Left, declarations)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(binary.Right, declarations)
	if err != nil {
		return nil, err
	}
	return &BinaryExpression{
		Operator: binary.Operator,
		Left:     left,
		Right:    right,
	}, nil
}
func analyzeUnaryExpression(unary *ast.UnaryExpression, declarations declarationScope) (*UnaryExpression, error) {
	arg, err := analyzeExpression(unary.Argument, declarations)
	if err != nil {
		return nil, err
	}
	return &UnaryExpression{
		Operator: unary.Operator,
		Argument: arg,
	}, nil
}
func analyzeLogicalExpression(logical *ast.LogicalExpression, declarations declarationScope) (*LogicalExpression, error) {
	left, err := analyzeExpression(logical.Left, declarations)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(logical.Right, declarations)
	if err != nil {
		return nil, err
	}
	return &LogicalExpression{
		Operator: logical.Operator,
		Left:     left,
		Right:    right,
	}, nil
}
func analyzeObjectExpression(obj *ast.ObjectExpression, declarations declarationScope) (*ObjectExpression, error) {
	o := &ObjectExpression{
		Properties: make([]*Property, len(obj.Properties)),
	}
	for i, p := range obj.Properties {
		n, err := analyzeProperty(p, declarations)
		if err != nil {
			return nil, err
		}
		o.Properties[i] = n
	}
	return o, nil
}
func analyzeArrayExpression(array *ast.ArrayExpression, declarations declarationScope) (*ArrayExpression, error) {
	a := &ArrayExpression{
		Elements: make([]Expression, len(array.Elements)),
	}
	for i, e := range array.Elements {
		n, err := analyzeExpression(e, declarations)
		if err != nil {
			return nil, err
		}
		a.Elements[i] = n
	}
	return a, nil
}

func analyzeIdentifier(ident *ast.Identifier, declarations declarationScope) (*Identifier, error) {
	return &Identifier{
		Name: ident.Name,
	}, nil
}

func analyzeIdentifierExpression(ident *ast.Identifier, declarations declarationScope) (*IdentifierExpression, error) {
	return &IdentifierExpression{
		Name:        ident.Name,
		Declaration: declarations[ident.Name],
	}, nil
}

func analyzeProperty(property *ast.Property, declarations declarationScope) (*Property, error) {
	key, err := analyzeIdentifier(property.Key, declarations)
	if err != nil {
		return nil, err
	}
	value, err := analyzeExpression(property.Value, declarations)
	if err != nil {
		return nil, err
	}
	return &Property{
		Key:   key,
		Value: value,
	}, nil
}

func analyzeDateTimeLiteral(lit *ast.DateTimeLiteral, declarations declarationScope) (*DateTimeLiteral, error) {
	return &DateTimeLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeDurationLiteral(lit *ast.DurationLiteral, declarations declarationScope) (*DurationLiteral, error) {
	return &DurationLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeFloatLiteral(lit *ast.FloatLiteral, declarations declarationScope) (*FloatLiteral, error) {
	return &FloatLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeIntegerLiteral(lit *ast.IntegerLiteral, declarations declarationScope) (*IntegerLiteral, error) {
	return &IntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeUnsignedIntegerLiteral(lit *ast.UnsignedIntegerLiteral, declarations declarationScope) (*UnsignedIntegerLiteral, error) {
	return &UnsignedIntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeStringLiteral(lit *ast.StringLiteral, declarations declarationScope) (*StringLiteral, error) {
	return &StringLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeBooleanLiteral(lit *ast.BooleanLiteral, declarations declarationScope) (*BooleanLiteral, error) {
	return &BooleanLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeRegexpLiteral(lit *ast.RegexpLiteral, declarations declarationScope) (*RegexpLiteral, error) {
	return &RegexpLiteral{
		Value: lit.Value,
	}, nil
}
