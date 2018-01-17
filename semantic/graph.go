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
	Type() string
	Copy() Node

	json.Marshaler
}

func (*Program) node() {}

func (*BlockStatement) node()      {}
func (*ExpressionStatement) node() {}
func (*ReturnStatement) node()     {}
func (*VariableDeclaration) node() {}
func (*VariableDeclarator) node()  {}

func (*ArrayExpression) node()         {}
func (*ArrowFunctionExpression) node() {}
func (*BinaryExpression) node()        {}
func (*CallExpression) node()          {}
func (*ConditionalExpression) node()   {}
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
func (*Identifier) expression()              {}
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

func (*Program) Type() string { return "Program" }

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

func (*BlockStatement) Type() string { return "BlockStatement" }

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

func (*ExpressionStatement) Type() string { return "ExpressionStatement" }

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

func (*ReturnStatement) Type() string { return "ReturnStatement" }

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
	Declarations []*VariableDeclarator `json:"declarations"`
}

func (*VariableDeclaration) Type() string { return "VariableDeclaration" }

func (s *VariableDeclaration) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(VariableDeclaration)
	*ns = *s

	if len(s.Declarations) > 0 {
		ns.Declarations = make([]*VariableDeclarator, len(s.Declarations))
		for i, decl := range s.Declarations {
			ns.Declarations[i] = decl.Copy().(*VariableDeclarator)
		}
	}

	return ns
}

type VariableDeclarator struct {
	ID   *Identifier `json:"id"`
	Init Expression  `json:"init"`
	// Uses is a list of every place this identifier is used.
	// TODO(nathanielc): Do we need this?
	//Uses []Expression
}

func (*VariableDeclarator) Type() string { return "VariableDeclarator" }

func (s *VariableDeclarator) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(VariableDeclarator)
	*ns = *s

	ns.ID = s.ID.Copy().(*Identifier)
	if s.Init != nil {
		ns.Init = s.Init.Copy().(Expression)
	}

	return ns
}

type ArrayExpression struct {
	Elements []Expression `json:"elements"`
}

func (*ArrayExpression) Type() string { return "ArrayExpression" }

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

func (*ArrowFunctionExpression) Type() string { return "ArrowFunctionExpression" }

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

func (*FunctionParam) Type() string { return "FunctionParam" }

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

func (*BinaryExpression) Type() string { return "BinaryExpression" }

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

func (*CallExpression) Type() string { return "CallExpression" }

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

func (*ConditionalExpression) Type() string { return "ConditionalExpression" }

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

func (*LogicalExpression) Type() string { return "LogicalExpression" }

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

func (*MemberExpression) Type() string { return "MemberExpression" }

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

func (*ObjectExpression) Type() string { return "ObjectExpression" }

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

func (*UnaryExpression) Type() string { return "UnaryExpression" }

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

func (*Property) Type() string { return "Property" }

func (p *Property) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Property)
	*np = *p

	np.Key = p.Key.Copy().(*Identifier)
	np.Value = p.Value.Copy().(Expression)

	return np
}

type Identifier struct {
	Name string `json:"name"`
	// Declaration is the node that declares this identifier
	// TODO(nathanielc): Do we need this?
	//Declaration *VariableDeclarator
}

func (*Identifier) Type() string { return "Identifier" }

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

func (*BooleanLiteral) Type() string { return "BooleanLiteral" }

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

func (*DateTimeLiteral) Type() string { return "DateTimeLiteral" }

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

func (*DurationLiteral) Type() string { return "DurationLiteral" }

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

func (*IntegerLiteral) Type() string { return "IntegerLiteral" }

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

func (*FloatLiteral) Type() string { return "FloatLiteral" }

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

func (*RegexpLiteral) Type() string { return "RegexpLiteral" }

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

func (*StringLiteral) Type() string { return "StringLiteral" }

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

func (*UnsignedIntegerLiteral) Type() string { return "UnsignedIntegerLiteral" }

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

func analyzeNode(n ast.Node) (Node, error) {
	switch n := n.(type) {
	case ast.Statement:
		return analyzeStatment(n)
	case ast.Expression:
		return analyzeExpression(n)
	default:
		return nil, fmt.Errorf("unsupported node %T", n)
	}
}

func analyzeProgram(prog *ast.Program) (*Program, error) {
	p := &Program{
		Body: make([]Statement, len(prog.Body)),
	}
	for i, s := range prog.Body {
		n, err := analyzeStatment(s)
		if err != nil {
			return nil, err
		}
		p.Body[i] = n
	}
	return p, nil
}

func analyzeStatment(s ast.Statement) (Statement, error) {
	switch s := s.(type) {
	case *ast.BlockStatement:
		return analyzeBlockStatement(s)
	case *ast.ExpressionStatement:
		return analyzeExpressionStatement(s)
	case *ast.ReturnStatement:
		return analyzeReturnStatement(s)
	case *ast.VariableDeclaration:
		return analyzeVariableDeclaration(s)
	default:
		return nil, fmt.Errorf("unsupported statement %T", s)
	}
}

func analyzeBlockStatement(block *ast.BlockStatement) (*BlockStatement, error) {
	b := &BlockStatement{
		Body: make([]Statement, len(block.Body)),
	}
	for i, s := range block.Body {
		n, err := analyzeStatment(s)
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

func analyzeExpressionStatement(expr *ast.ExpressionStatement) (*ExpressionStatement, error) {
	e, err := analyzeExpression(expr.Expression)
	if err != nil {
		return nil, err
	}
	return &ExpressionStatement{
		Expression: e,
	}, nil
}

func analyzeReturnStatement(ret *ast.ReturnStatement) (*ReturnStatement, error) {
	arg, err := analyzeExpression(ret.Argument)
	if err != nil {
		return nil, err
	}
	return &ReturnStatement{
		Argument: arg,
	}, nil
}

func analyzeVariableDeclaration(decl *ast.VariableDeclaration) (*VariableDeclaration, error) {
	vd := &VariableDeclaration{
		Declarations: make([]*VariableDeclarator, len(decl.Declarations)),
	}
	for i, d := range decl.Declarations {
		n, err := analyzeVariableDeclarator(d)
		if err != nil {
			return nil, err
		}
		vd.Declarations[i] = n
	}
	return vd, nil
}
func analyzeVariableDeclarator(decl *ast.VariableDeclarator) (*VariableDeclarator, error) {
	id, err := analyzeIdentifier(decl.ID)
	if err != nil {
		return nil, err
	}
	init, err := analyzeExpression(decl.Init)
	if err != nil {
		return nil, err
	}
	return &VariableDeclarator{
		ID:   id,
		Init: init,
	}, nil
}

func analyzeExpression(expr ast.Expression) (Expression, error) {
	switch expr := expr.(type) {
	case *ast.ArrowFunctionExpression:
		return analyzeArrowFunctionExpression(expr)
	case *ast.CallExpression:
		return analyzeCallExpression(expr)
	case *ast.MemberExpression:
		return analyzeMemberExpression(expr)
	case *ast.BinaryExpression:
		return analyzeBinaryExpression(expr)
	case *ast.UnaryExpression:
		return analyzeUnaryExpression(expr)
	case *ast.LogicalExpression:
		return analyzeLogicalExpression(expr)
	case *ast.ObjectExpression:
		return analyzeObjectExpression(expr)
	case *ast.ArrayExpression:
		return analyzeArrayExpression(expr)
	case *ast.Identifier:
		return analyzeIdentifier(expr)
	case ast.Literal:
		return analyzeLiteral(expr)
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func analyzeLiteral(lit ast.Literal) (Literal, error) {
	switch lit := lit.(type) {
	case *ast.StringLiteral:
		return analyzeStringLiteral(lit)
	case *ast.BooleanLiteral:
		return analyzeBooleanLiteral(lit)
	case *ast.FloatLiteral:
		return analyzeFloatLiteral(lit)
	case *ast.IntegerLiteral:
		return analyzeIntegerLiteral(lit)
	case *ast.UnsignedIntegerLiteral:
		return analyzeUnsignedIntegerLiteral(lit)
	case *ast.RegexpLiteral:
		return analyzeRegexpLiteral(lit)
	case *ast.DurationLiteral:
		return analyzeDurationLiteral(lit)
	case *ast.DateTimeLiteral:
		return analyzeDateTimeLiteral(lit)
	default:
		return nil, fmt.Errorf("unsupported literal %T", lit)
	}
}

func analyzeArrowFunctionExpression(arrow *ast.ArrowFunctionExpression) (*ArrowFunctionExpression, error) {
	b, err := analyzeNode(arrow.Body)
	if err != nil {
		return nil, err
	}
	f := &ArrowFunctionExpression{
		Params: make([]*FunctionParam, len(arrow.Params)),
		Body:   b,
	}
	for i, p := range arrow.Params {
		key, err := analyzeIdentifier(p.Key)
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
			def, err = analyzeLiteral(lit)
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

func analyzeCallExpression(call *ast.CallExpression) (*CallExpression, error) {
	callee, err := analyzeExpression(call.Callee)
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
		args, err = analyzeObjectExpression(obj)
		if err != nil {
			return nil, err
		}
	}
	return &CallExpression{
		Callee:    callee,
		Arguments: args,
	}, nil
}

func analyzeMemberExpression(member *ast.MemberExpression) (*MemberExpression, error) {
	obj, err := analyzeExpression(member.Object)
	if err != nil {
		return nil, err
	}
	property, err := analyzeExpression(member.Property)
	if err != nil {
		return nil, err
	}

	var propertyName string
	switch p := property.(type) {
	case *Identifier:
		propertyName = p.Name
	case *StringLiteral:
		propertyName = p.Value
	default:
		return nil, fmt.Errorf("unsupported member property expression of type %T", property)
	}

	return &MemberExpression{
		Object:   obj,
		Property: propertyName,
	}, nil
}

func analyzeBinaryExpression(binary *ast.BinaryExpression) (*BinaryExpression, error) {
	left, err := analyzeExpression(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(binary.Right)
	if err != nil {
		return nil, err
	}
	return &BinaryExpression{
		Operator: binary.Operator,
		Left:     left,
		Right:    right,
	}, nil
}
func analyzeUnaryExpression(unary *ast.UnaryExpression) (*UnaryExpression, error) {
	arg, err := analyzeExpression(unary.Argument)
	if err != nil {
		return nil, err
	}
	return &UnaryExpression{
		Operator: unary.Operator,
		Argument: arg,
	}, nil
}
func analyzeLogicalExpression(logical *ast.LogicalExpression) (*LogicalExpression, error) {
	left, err := analyzeExpression(logical.Left)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(logical.Right)
	if err != nil {
		return nil, err
	}
	return &LogicalExpression{
		Operator: logical.Operator,
		Left:     left,
		Right:    right,
	}, nil
}
func analyzeObjectExpression(obj *ast.ObjectExpression) (*ObjectExpression, error) {
	o := &ObjectExpression{
		Properties: make([]*Property, len(obj.Properties)),
	}
	for i, p := range obj.Properties {
		n, err := analyzeProperty(p)
		if err != nil {
			return nil, err
		}
		o.Properties[i] = n
	}
	return o, nil
}
func analyzeArrayExpression(array *ast.ArrayExpression) (*ArrayExpression, error) {
	a := &ArrayExpression{
		Elements: make([]Expression, len(array.Elements)),
	}
	for i, e := range array.Elements {
		n, err := analyzeExpression(e)
		if err != nil {
			return nil, err
		}
		a.Elements[i] = n
	}
	return a, nil
}
func analyzeIdentifier(ident *ast.Identifier) (*Identifier, error) {
	return &Identifier{
		Name: ident.Name,
	}, nil
}
func analyzeProperty(property *ast.Property) (*Property, error) {
	key, err := analyzeIdentifier(property.Key)
	if err != nil {
		return nil, err
	}
	value, err := analyzeExpression(property.Value)
	if err != nil {
		return nil, err
	}
	return &Property{
		Key:   key,
		Value: value,
	}, nil
}

func analyzeDateTimeLiteral(lit *ast.DateTimeLiteral) (*DateTimeLiteral, error) {
	return &DateTimeLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeDurationLiteral(lit *ast.DurationLiteral) (*DurationLiteral, error) {
	return &DurationLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeFloatLiteral(lit *ast.FloatLiteral) (*FloatLiteral, error) {
	return &FloatLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeIntegerLiteral(lit *ast.IntegerLiteral) (*IntegerLiteral, error) {
	return &IntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeUnsignedIntegerLiteral(lit *ast.UnsignedIntegerLiteral) (*UnsignedIntegerLiteral, error) {
	return &UnsignedIntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeStringLiteral(lit *ast.StringLiteral) (*StringLiteral, error) {
	return &StringLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeBooleanLiteral(lit *ast.BooleanLiteral) (*BooleanLiteral, error) {
	return &BooleanLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeRegexpLiteral(lit *ast.RegexpLiteral) (*RegexpLiteral, error) {
	return &RegexpLiteral{
		Value: lit.Value,
	}, nil
}
