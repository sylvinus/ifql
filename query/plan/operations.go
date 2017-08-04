package plan

import (
	"fmt"
	"reflect"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

type OperationID uuid.UUID

func (oid OperationID) String() string {
	return uuid.UUID(oid).String()
}

var InvalidOperationID OperationID

type Operation struct {
	ID       OperationID
	Parents  []DatasetID
	Children []DatasetID
	Spec     OperationSpec
}

// OperationSpec specifies an operation as part of a query.
type OperationSpec interface {
	// Kind returns the kind of the operation.
	Kind() OperationKind

	// SetSpec applies the query.OperationSpec to this OperationSpec
	SetSpec(query.OperationSpec) error
}

type NarrowOperationSpec interface {
	NewChild(*Dataset)
}
type WideOperationSpec interface {
	DetermineChildren() []*Dataset
}

// OperationKind denotes the kind of operations.
type OperationKind int

func (k OperationKind) String() string {
	return kindToNames[k]
}

type PredicateSpec interface{}

// TODO actually design implement this idea
func ExpressionToPredicate(query.ExpressionSpec) PredicateSpec {
	return nil
}

const (
	SelectKind OperationKind = iota
	RangeKind
	ClearKind
	WindowKind
	MergeKind
	KeysKind
	ValuesKind
	CardinalityKind
	LimitKind
	ShiftKind
	InterpolateKind
	JoinKind
	UnionKind
	FilterKind
	SortKind
	RateKind
	CountKind
	SumKind
	MeanKind
	PercentileKind
	StddevKind
	MinKind
	MaxKind
	TopKind
	DifferenceKind
	// NumberOfKinds is the of Kind values
	NumberOfKinds int = iota
)

var kindToNames [NumberOfKinds]string
var namesToKind = make(map[string]OperationKind)
var kindToGoType [NumberOfKinds]reflect.Type
var queryOpToOpKind [query.NumberOfKinds]OperationKind

var operationSpecType = reflect.TypeOf((*OperationSpec)(nil)).Elem()

// registerOpSpec populates package vars with lookup information about the OperationSpec.
func registerOpSpec(k OperationKind, qk query.OperationKind, name string, s OperationSpec) {
	if kindToGoType[k] != nil {
		panic(fmt.Errorf("duplicate registration for Kind %v", k))
	}
	t := reflect.Indirect(reflect.ValueOf(s)).Type()
	if !reflect.PtrTo(t).Implements(operationSpecType) {
		panic(fmt.Errorf("type %T does not implement OperationSpec as a pointer receiver", s))
	}
	kindToGoType[k] = t

	kindToNames[k] = name
	_, ok := namesToKind[name]
	if ok {
		panic(fmt.Errorf("duplicate kind name found %q", name))
	}
	namesToKind[name] = k

	queryOpToOpKind[qk] = k
}

type SelectOpSpec struct {
	Database string        `json:"database"`
	Where    PredicateSpec `json:"where"`
}

func init() {
	registerOpSpec(SelectKind, query.SelectKind, "select", new(SelectOpSpec))
}

func (s *SelectOpSpec) Kind() OperationKind {
	return SelectKind
}

func (s *SelectOpSpec) SetSpec(qs query.OperationSpec) error {
	spec, ok := qs.(*query.SelectOpSpec)
	if !ok {
		return fmt.Errorf("invalid spec type %T", qs)
	}
	s.Database = spec.Database
	s.Where = ExpressionToPredicate(spec.Where)
	return nil
}

func (s *SelectOpSpec) DetermineChildren() []*Dataset {
	return []*Dataset{new(Dataset)}
}

type RangeOpSpec struct {
	Bounds BoundsSpec
}

func init() {
	registerOpSpec(RangeKind, query.RangeKind, "range", new(RangeOpSpec))
}

func (s *RangeOpSpec) Kind() OperationKind {
	return RangeKind
}

func (s *RangeOpSpec) SetSpec(qs query.OperationSpec) error {
	spec, ok := qs.(*query.RangeOpSpec)
	if !ok {
		return fmt.Errorf("invalid spec type %T", qs)
	}
	s.Bounds = BoundsSpec{
		Start: spec.Start,
		Stop:  spec.Stop,
	}
	return nil
}
func (s *RangeOpSpec) NewChild(ds *Dataset) {
	ds.Bounds = s.Bounds
}

type ClearOpSpec struct {
}

func init() {
	registerOpSpec(ClearKind, query.ClearKind, "clear", new(ClearOpSpec))
}

func (s *ClearOpSpec) Kind() OperationKind {
	return ClearKind
}

func (s *ClearOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type WindowOpSpec struct {
	Window WindowSpec
}

func init() {
	registerOpSpec(WindowKind, query.WindowKind, "window", new(WindowOpSpec))
}

func (s *WindowOpSpec) Kind() OperationKind {
	return WindowKind
}

func (s *WindowOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MergeOpSpec struct {
	Keys []string `json:"keys"`
	Keep []string `json:"keep"`
}

func init() {
	registerOpSpec(MergeKind, query.MergeKind, "merge", new(MergeOpSpec))
}

func (s *MergeOpSpec) Kind() OperationKind {
	return MergeKind
}

func (s *MergeOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type KeysOpSpec struct {
}

func init() {
	registerOpSpec(KeysKind, query.KeysKind, "keys", new(KeysOpSpec))
}

func (s *KeysOpSpec) Kind() OperationKind {
	return KeysKind
}

func (s *KeysOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type ValuesOpSpec struct {
}

func init() {
	registerOpSpec(ValuesKind, query.ValuesKind, "values", new(ValuesOpSpec))
}

func (s *ValuesOpSpec) Kind() OperationKind {
	return ValuesKind
}

func (s *ValuesOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type CardinalityOpSpec struct {
}

func init() {
	registerOpSpec(CardinalityKind, query.CardinalityKind, "cardinality", new(CardinalityOpSpec))
}

func (s *CardinalityOpSpec) Kind() OperationKind {
	return CardinalityKind
}

func (s *CardinalityOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type LimitOpSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func init() {
	registerOpSpec(LimitKind, query.LimitKind, "limit", new(LimitOpSpec))
}

func (s *LimitOpSpec) Kind() OperationKind {
	return LimitKind
}

func (s *LimitOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type ShiftOpSpec struct {
	Duration query.Duration `json:"duration"`
}

func init() {
	registerOpSpec(ShiftKind, query.ShiftKind, "shift", new(ShiftOpSpec))
}

func (s *ShiftOpSpec) Kind() OperationKind {
	return ShiftKind
}

func (s *ShiftOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type InterpolateOpSpec struct {
	Start query.Time     `json:"start"`
	Every query.Duration `json:"every"`
	// TODO define this spec
	Value interface{} `json:"value"`
}

func init() {
	registerOpSpec(InterpolateKind, query.InterpolateKind, "interpolate", new(InterpolateOpSpec))
}

func (s *InterpolateOpSpec) Kind() OperationKind {
	return InterpolateKind
}

func (s *InterpolateOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type JoinOpSpec struct {
	Keys      []string      `json:"keys"`
	Predicate PredicateSpec `json:"predicate"`
}

func init() {
	registerOpSpec(JoinKind, query.JoinKind, "join", new(JoinOpSpec))
}

func (s *JoinOpSpec) Kind() OperationKind {
	return JoinKind
}

func (s *JoinOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type UnionOpSpec struct {
}

func init() {
	registerOpSpec(UnionKind, query.UnionKind, "union", new(UnionOpSpec))
}

func (s *UnionOpSpec) Kind() OperationKind {
	return UnionKind
}

func (s *UnionOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type FilterOpSpec struct {
	Predicate PredicateSpec `json:"predicate"`
}

func init() {
	registerOpSpec(FilterKind, query.FilterKind, "filter", new(FilterOpSpec))
}

func (s *FilterOpSpec) Kind() OperationKind {
	return FilterKind
}

func (s *FilterOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type SortOpSpec struct {
	Predicate PredicateSpec `json:"predicate"`
	Keys      []string      `json:"keys"`
	// TODO define Order type
	Order string `json:"order"` //asc or desc
}

func init() {
	registerOpSpec(SortKind, query.SortKind, "sort", new(SortOpSpec))
}

func (s *SortOpSpec) Kind() OperationKind {
	return SortKind
}

func (s *SortOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type RateOpSpec struct {
	Unit query.Duration `json:"unit"`
}

func init() {
	registerOpSpec(RateKind, query.RateKind, "rate", new(RateOpSpec))
}

func (s *RateOpSpec) Kind() OperationKind {
	return RateKind
}

func (s *RateOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type CountOpSpec struct {
}

func init() {
	registerOpSpec(CountKind, query.CountKind, "count", new(CountOpSpec))
}

func (s *CountOpSpec) Kind() OperationKind {
	return CountKind
}

func (s *CountOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

func (s *CountOpSpec) NewChild(ds *Dataset) {}

type SumOpSpec struct {
}

func init() {
	registerOpSpec(SumKind, query.SumKind, "sum", new(SumOpSpec))
}

func (s *SumOpSpec) Kind() OperationKind {
	return SumKind
}

func (s *SumOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}
func (s *SumOpSpec) NewChild(ds *Dataset) {}

type MeanOpSpec struct {
}

func init() {
	registerOpSpec(MeanKind, query.MeanKind, "mean", new(MeanOpSpec))
}

func (s *MeanOpSpec) Kind() OperationKind {
	return MeanKind
}

func (s *MeanOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}
func (s *MeanOpSpec) NewChild(ds *Dataset) {}

type PercentileOpSpec struct {
}

func init() {
	registerOpSpec(PercentileKind, query.PercentileKind, "percentile", new(PercentileOpSpec))
}

func (s *PercentileOpSpec) Kind() OperationKind {
	return PercentileKind
}

func (s *PercentileOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type StddevOpSpec struct {
}

func init() {
	registerOpSpec(StddevKind, query.StddevKind, "stddev", new(StddevOpSpec))
}

func (s *StddevOpSpec) Kind() OperationKind {
	return StddevKind
}

func (s *StddevOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MinOpSpec struct {
}

func init() {
	registerOpSpec(MinKind, query.MinKind, "min", new(MinOpSpec))
}

func (s *MinOpSpec) Kind() OperationKind {
	return MinKind
}

func (s *MinOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MaxOpSpec struct {
}

func init() {
	registerOpSpec(MaxKind, query.MaxKind, "max", new(MaxOpSpec))
}

func (s *MaxOpSpec) Kind() OperationKind {
	return MaxKind
}

func (s *MaxOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type TopOpSpec struct {
}

func init() {
	registerOpSpec(TopKind, query.TopKind, "top", new(TopOpSpec))
}

func (s *TopOpSpec) Kind() OperationKind {
	return TopKind
}

func (s *TopOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type DifferenceOpSpec struct {
}

func init() {
	registerOpSpec(DifferenceKind, query.DifferenceKind, "difference", new(DifferenceOpSpec))
}

func (s *DifferenceOpSpec) Kind() OperationKind {
	return DifferenceKind
}
func (s *DifferenceOpSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}
