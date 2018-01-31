package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
	"github.com/influxdata/ifql/semantic"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

const (
	TableParameter = "table"
	tableIDKey     = "id"
)

// Compile evaluates an IFQL script producing a query Spec.
func Compile(ctx context.Context, q string) (*Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	astProg, err := parser.NewAST(q)
	if err != nil {
		return nil, err
	}
	s.Finish()
	s, _ = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()

	// Convert AST program to a semantic program
	semProg, err := semantic.New(astProg, builtinDeclarations)
	if err != nil {
		return nil, err
	}

	// Create top-level builtin scope
	scope := builtinScope.Nest()

	// Create new query domain
	d := new(queryDomain)

	if err := interpreter.Eval(semProg, scope, d); err != nil {
		return nil, err
	}
	spec := d.ToSpec()
	//log.Println(Formatted(spec, FmtJSON))
	return spec, nil
}

type CreateOperationSpec func(args Arguments, a *Administration) (OperationSpec, error)

var functionsMap = make(map[string]function)

// RegisterFunction adds a new builtin top level function.
func RegisterFunction(name string, c CreateOperationSpec, sig semantic.FunctionSignature) {
	if finalized {
		panic(errors.New("already finalized, cannot register function"))
	}
	if _, ok := functionsMap[name]; ok {
		panic(fmt.Errorf("duplicate registration for function %q", name))
	}
	f := function{
		name:         name,
		createOpSpec: c,
	}
	functionsMap[name] = f
	builtinScope.Set(name, f)
	builtinDeclarations[name] = semantic.NewExternalVariableDeclaration(
		name,
		semantic.NewFunctionType(sig),
	)
}

var TableObjectType = semantic.NewObjectType(map[string]semantic.Type{tableIDKey: semantic.String})

// DefaultFunctionSignature returns a FunctionSignature for standard functions which accept a table piped argument.
// It is safe to modify the returned signature.
func DefaultFunctionSignature() semantic.FunctionSignature {
	return semantic.FunctionSignature{
		Params: map[string]semantic.Type{
			TableParameter: TableObjectType,
		},
		ReturnType:   TableObjectType,
		PipeArgument: TableParameter,
	}
}

var builtinScope = interpreter.NewScope()
var builtinDeclarations = make(map[string]semantic.VariableDeclaration)

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
// Future calls to RegisterFunction or RegisterBuiltIn will panic.
func FinalizeRegistration() {
	finalized = true
	for _, script := range builtins {
		astProg, err := parser.NewAST(script)
		if err != nil {
			panic(err)
		}
		semProg, err := semantic.New(astProg, builtinDeclarations)
		if err != nil {
			panic(err)
		}

		// Create new query domain
		d := new(queryDomain)

		if err := interpreter.Eval(semProg, builtinScope, d); err != nil {
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

// AddParentFromArgs reads the args for the `table` argument and adds the value as a parent.
func (a *Administration) AddParentFromArgs(args Arguments) error {
	parent, err := args.GetRequiredObject(TableParameter)
	if err != nil {
		return err
	}
	a.AddParent(GetIDFromObject(parent))
	return nil
}

func GetIDFromObject(obj interpreter.Object) OperationID {
	return OperationID(obj.Properties[tableIDKey].Value().(string))
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

type function struct {
	name         string
	createOpSpec CreateOperationSpec
}

func (f function) Type() semantic.Type {
	//TODO(nathanielc): Return a complete function type
	return semantic.Function
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

	spec, err := f.createOpSpec(Arguments{Arguments: args}, a)
	if err != nil {
		return nil, err
	}
	o.Spec = spec

	a.finalize()

	return interpreter.Object{
		Properties: map[string]interpreter.Value{
			tableIDKey: interpreter.NewStringValue(string(o.ID)),
		},
	}, nil
}

func (f function) Resolve() (*semantic.FunctionExpression, error) {
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
