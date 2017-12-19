package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/ifql"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	parentTableArg = "table"
)

// Compile parses IFQL into an AST; validates and checks the AST; and produces a QuerySpec.
func Compile(ctx context.Context, q string, opts ...ifql.Option) (*QuerySpec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	program, err := ifql.NewAST(q, opts...)
	if err != nil {
		return nil, err
	}
	s.Finish()
	s, _ = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()

	// Create top-level builtin scope
	scope := newBuiltInScope()

	// Create new query domain
	d := new(queryDomain)

	if err := ifql.Eval(program, scope, d); err != nil {
		return nil, err
	}
	return d.ToSpec(), nil
}

type CreateOperationSpec func(args Arguments, ctx *Administration) (OperationSpec, error)

var functionsMap = make(map[string]function)

// RegisterFunction adds a new top level builtin function.
func RegisterFunction(name string, c CreateOperationSpec) {
	registerFunction(name, c, false)
}

// RegisterMethod adds a new builtin method.
func RegisterMethod(name string, c CreateOperationSpec) {
	registerFunction(name, c, true)
}

func registerFunction(name string, c CreateOperationSpec, chainable bool) {
	if _, ok := functionsMap[name]; ok {
		panic(fmt.Errorf("duplicate registration for function %q", name))
	}
	functionsMap[name] = function{
		name:         name,
		createOpSpec: c,
		chainable:    chainable,
	}
}

// newBuiltInScope returns a scope populated with the builtin values.
func newBuiltInScope() *ifql.Scope {
	s := ifql.NewScope()
	for n, f := range functionsMap {
		s.Set(n, f)
	}
	return s
}

type Administration struct {
	parents []OperationID
}

// AddParent instructs the evaluation Context that a new edge should be created from the parent to the current operation.
// Duplicate parents will be removed, so the caller need not concern itself with which parents have already been added.
func (c *Administration) AddParent(id OperationID) {
	// Check for duplicates
	for _, p := range c.parents {
		if p == id {
			return
		}
	}
	c.parents = append(c.parents, id)
}

type queryDomain struct {
	id int

	operations []*Operation
	edges      []Edge
}

func (d *queryDomain) AddOperation(name string) *Operation {
	o := &Operation{
		ID: OperationID(fmt.Sprintf("%s%d", name, d.nextID())),
	}
	d.operations = append(d.operations, o)
	return o
}

func (d *queryDomain) nextID() int {
	id := d.id
	d.id++
	return id
}

func (d *queryDomain) AddParentEdges(id OperationID, parents ...OperationID) {
	if len(parents) > 1 {
		// Always add parents in a consistent order
		sort.Slice(parents, func(i, j int) bool { return parents[i] < parents[j] })
	}
	for _, p := range parents {
		if p != id {
			d.edges = append(d.edges, Edge{
				Parent: p,
				Child:  id,
			})
		}
	}
}

func (d *queryDomain) ToSpec() *QuerySpec {
	return &QuerySpec{
		Operations: d.operations,
		Edges:      d.edges,
	}
}

var TTable = ifql.RegisterType("table")

// Table represents a table created via a function call.
type Table struct {
	ID OperationID
}

func (t Table) Type() ifql.Type {
	return TTable
}

func (t Table) Value() interface{} {
	return t
}

func (t Table) Property(name string) (ifql.Value, error) {
	// All chainable methods are properties of all tables
	f, ok := functionsMap[name]
	if !ok || !f.chainable {
		return nil, fmt.Errorf("unknown property %s", name)
	}
	f.parentID = t.ID
	return f, nil
}

type function struct {
	name         string
	createOpSpec CreateOperationSpec
	chainable    bool
	parentID     OperationID
}

func (f function) Type() ifql.Type {
	return ifql.TFunction
}

func (f function) Value() interface{} {
	return f
}
func (f function) Property(name string) (ifql.Value, error) {
	return nil, fmt.Errorf("property %q does not exist", name)
}

func (f function) Call(args ifql.Arguments, d ifql.Domain) (ifql.Value, error) {
	qd := d.(*queryDomain)
	o := qd.AddOperation(f.name)

	if f.chainable {
		qd.AddParentEdges(o.ID, f.parentID)
	}

	ctx := new(Administration)
	spec, err := f.createOpSpec(Arguments{Arguments: args}, ctx)
	if err != nil {
		return nil, err
	}
	o.Spec = spec

	// Add any additional parents
	qd.AddParentEdges(o.ID, ctx.parents...)

	return Table{
		ID: o.ID,
	}, nil
}

func (f function) Resolve() (*ast.ArrowFunctionExpression, error) {
	return nil, fmt.Errorf("function %q cannot be resolved", f.name)
}

type Arguments struct {
	ifql.Arguments
}

func (a Arguments) GetTime(name string) (Time, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return Time{}, false, nil
	}
	qt, err := ToQueryTime(v)
	if err != nil {
		return Time{}, ok, err
	}
	return qt, ok, nil
}

func (a Arguments) GetRequiredTime(name string) (Time, error) {
	qt, ok, err := a.GetTime(name)
	if err != nil {
		return Time{}, err
	}
	if !ok {
		return Time{}, fmt.Errorf("missing required keyword argument %q", name)
	}
	return qt, nil
}

func (a Arguments) GetDuration(name string) (Duration, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return 0, false, nil
	}
	return (Duration)(v.Value().(time.Duration)), ok, nil
}

func (a Arguments) GetRequiredDuration(name string) (Duration, error) {
	d, ok, err := a.GetDuration(name)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("missing required keyword argument %q", name)
	}
	return d, nil
}

func ToQueryTime(value ifql.Value) (Time, error) {
	switch v := value.Value().(type) {
	case time.Time:
		return Time{
			Absolute: v,
		}, nil
	case time.Duration:
		return Time{
			Relative:   v,
			IsRelative: true,
		}, nil
	case int64:
		return Time{
			Absolute: time.Unix(v, 0),
		}, nil
	default:
		return Time{}, fmt.Errorf("value is not a time, got %v", value.Type)
	}
}
