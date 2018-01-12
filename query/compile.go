package query

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Pass through interpreter types so that consumers of the query package need not know about the interpreter package.
const (
	TInvalid  = interpreter.TInvalid
	TString   = interpreter.TString
	TInt      = interpreter.TInt
	TUInt     = interpreter.TUInt
	TFloat    = interpreter.TFloat
	TBool     = interpreter.TBool
	TTime     = interpreter.TTime
	TDuration = interpreter.TDuration
	TFunction = interpreter.TFunction
	TArray    = interpreter.TArray
	TMap      = interpreter.TMap
)

// Compile parses IFQL into an AST; validates and checks the AST; and produces a query Spec.
func Compile(ctx context.Context, q string) (*Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	program, err := parser.NewAST(q)
	if err != nil {
		return nil, err
	}
	s.Finish()
	s, _ = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()

	// Create top-level builtin scope
	scope := builtinScope.Nest()

	// Create new query domain
	d := new(queryDomain)

	if err := interpreter.Eval(program, scope, d, interpreter.NewFileImporter()); err != nil {
		return nil, err
	}
	spec := d.ToSpec()
	log.Println(Formatted(spec, FmtJSON))
	return spec, nil
}

type CreateOperationSpec func(args Arguments, a *Administration) (OperationSpec, error)

var functionsMap = make(map[string]function)

// RegisterFunction adds a new builtin top level function.
func RegisterFunction(name string, c CreateOperationSpec) {
	registerFunction(name, c, false)
}

// RegisterMethod adds a new builtin method of the TTable type.
func RegisterMethod(name string, c CreateOperationSpec) {
	registerFunction(name, c, true)
}

func registerFunction(name string, c CreateOperationSpec, chainable bool) {
	if finalized {
		panic(fmt.Errorf("already finalized, cannot register function %q", name))
	}
	if _, ok := functionsMap[name]; ok {
		panic(fmt.Errorf("duplicate registration for function %q", name))
	}
	f := function{
		name:         name,
		createOpSpec: c,
		chainable:    chainable,
	}
	functionsMap[name] = f
	if !chainable {
		builtinScope.Set(name, f)
	}
}

var builtinScope = interpreter.NewScope()

// list of builtin scripts
var builtins []string
var finalized bool

// RegisterBuiltIn adds any variable declarations in the script to the builtin scope.
func RegisterBuiltIn(script string) {
	if finalized {
		panic(errors.New("already finalized, cannot register builtin"))
	}
	builtins = append(builtins, script)
}

// FinalizeRegistration must be called to complete registration.
// Future calls to RegisterFunction, RegisterMethod or RegisterBuiltIn will panic.
func FinalizeRegistration() {
	finalized = true
	for _, script := range builtins {
		program, err := parser.NewAST(script)
		if err != nil {
			panic(err)
		}

		// Create new query domain
		d := new(queryDomain)

		if err := interpreter.Eval(program, builtinScope, d, nil); err != nil {
			panic(err)
		}
	}
	// free builtins list
	builtins = nil
}

type Administration struct {
	id      OperationID
	parents []OperationID
	d       *queryDomain
}

// AddParent instructs the evaluation Context that a new edge should be created from the parent to the current operation.
// Duplicate parents will be removed, so the caller need not concern itself with which parents have already been added.
func (a *Administration) AddParent(id OperationID) {
	// Check for duplicates
	for _, p := range a.parents {
		if p == id {
			return
		}
	}
	a.parents = append(a.parents, id)
}

func (a *Administration) finalize() {
	// Add parents
	a.d.AddParentEdges(a.id, a.parents...)
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

func (d *queryDomain) ToSpec() *Spec {
	return &Spec{
		Operations: d.operations,
		Edges:      d.edges,
	}
}

var TTable = interpreter.RegisterType("table")

// Table represents a table created via a function call.
type Table struct {
	ID OperationID
}

func (t Table) Type() interpreter.Type {
	return TTable
}

func (t Table) Value() interface{} {
	return t
}

func (t Table) Property(name string) (interpreter.Value, error) {
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

func (f function) Type() interpreter.Type {
	return interpreter.TFunction
}

func (f function) String() string {
	return fmt.Sprintf("%s()", f.name)
}

func (f function) Value() interface{} {
	return f
}
func (f function) Property(name string) (interpreter.Value, error) {
	return nil, fmt.Errorf("property %q does not exist", name)
}

func (f function) Call(args interpreter.Arguments, d interpreter.Domain) (interpreter.Value, error) {
	qd := d.(*queryDomain)
	o := qd.AddOperation(f.name)

	a := &Administration{
		id: o.ID,
		d:  qd,
	}
	if f.chainable {
		a.parents = append(a.parents, f.parentID)
	}

	spec, err := f.createOpSpec(Arguments{Arguments: args}, a)
	if err != nil {
		return nil, err
	}
	o.Spec = spec

	a.finalize()

	return Table{
		ID: o.ID,
	}, nil
}

func (f function) Resolve() (*ast.ArrowFunctionExpression, error) {
	return nil, fmt.Errorf("function %q cannot be resolved", f.name)
}

type Arguments struct {
	interpreter.Arguments
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
func (a Arguments) GetTable(name string) (Table, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return Table{}, false, nil
	}
	return v.Value().(Table), ok, nil
}

func (a Arguments) GetRequiredTable(name string) (Table, error) {
	d, ok, err := a.GetTable(name)
	if err != nil {
		return Table{}, err
	}
	if !ok {
		return Table{}, fmt.Errorf("missing required keyword argument %q", name)
	}
	return d, nil
}

func ToQueryTime(value interpreter.Value) (Time, error) {
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
		return Time{}, fmt.Errorf("value is not a time, got %v", value.Type())
	}
}
