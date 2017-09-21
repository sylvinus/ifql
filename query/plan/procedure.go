package plan

import (
	"fmt"
	"reflect"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute/storage"
	uuid "github.com/satori/go.uuid"
)

type ProcedureID uuid.UUID

func (id ProcedureID) String() string {
	return uuid.UUID(id).String()
}

var ZeroProcedureID ProcedureID

type Procedure struct {
	ID      ProcedureID
	Parents []DatasetID
	Child   DatasetID
	Spec    ProcedureSpec
}

// ProcedureSpec specifies an operation as part of a query.
type ProcedureSpec interface {
	// Kind returns the kind of the operation.
	Kind() ProcedureKind

	// SetSpec applies the query.OperationSpec to this OperationSpec
	SetSpec(query.OperationSpec) error
}

// ProcedureKind denotes the kind of operations.
type ProcedureKind int

func (k ProcedureKind) String() string {
	return kindToNames[k]
}

type PredicateSpec interface{}

// TODO actually design implement this idea
func ExpressionToPredicate(query.ExpressionSpec) PredicateSpec {
	return nil
}

const (
	SelectKind ProcedureKind = iota
	RangeKind
	WhereKind
	LimitKind
	ClearKind
	WindowKind
	MergeKind
	KeysKind
	ValuesKind
	CardinalityKind
	ShiftKind
	InterpolateKind
	MergeJoinKind
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
	// NumberOfProcedures is the of Kind values
	NumberOfProcedures int = iota
)

var kindToNames [NumberOfProcedures]string
var namesToKind = make(map[string]ProcedureKind)
var kindToGoType [NumberOfProcedures]reflect.Type
var opToProcedureKind [query.NumberOfOperations]ProcedureKind

var procedureSpecType = reflect.TypeOf((*ProcedureSpec)(nil)).Elem()

// registerProcedureSpec populates package vars with lookup information about the ProcedureSpeq.
func registerProcedureSpec(k ProcedureKind, qk query.OperationKind, name string, s ProcedureSpec) {
	if kindToGoType[k] != nil {
		panic(fmt.Errorf("duplicate registration for Kind %v", k))
	}
	t := reflect.Indirect(reflect.ValueOf(s)).Type()
	if !reflect.PtrTo(t).Implements(procedureSpecType) {
		panic(fmt.Errorf("type %T does not implement OperationSpec as a pointer receiver", s))
	}
	kindToGoType[k] = t

	kindToNames[k] = name
	_, ok := namesToKind[name]
	if ok {
		panic(fmt.Errorf("duplicate kind name found %q", name))
	}
	namesToKind[name] = k

	opToProcedureKind[qk] = k
}

type SelectProcedureSpec struct {
	Database string

	BoundsSet bool
	Bounds    BoundsSpec

	WhereSet bool
	Where    *storage.Predicate

	DescSet bool
	Desc    bool

	LimitSet bool
	Limit    int64
	Offset   int64

	WindowSet bool
	Window    WindowSpec
}

func init() {
	registerProcedureSpec(SelectKind, query.SelectKind, "select", new(SelectProcedureSpec))
}

func (s *SelectProcedureSpec) Kind() ProcedureKind {
	return SelectKind
}

func (s *SelectProcedureSpec) SetSpec(qs query.OperationSpec) error {
	spec, ok := qs.(*query.SelectOpSpec)
	if !ok {
		return fmt.Errorf("invalid spec type %T", qs)
	}
	s.Database = spec.Database
	return nil
}

func (s *SelectProcedureSpec) DetermineChildren() []*Dataset {
	return []*Dataset{new(Dataset)}
}

type WhereProcedureSpec struct {
	Exp *query.WhereOpSpec `json:"exp"`
}

func init() {
	registerProcedureSpec(WhereKind, query.WhereKind, "where", new(WhereProcedureSpec))
}

func (w *WhereProcedureSpec) Kind() ProcedureKind {
	return WhereKind
}

func (w *WhereProcedureSpec) SetSpec(qs query.OperationSpec) error {
	spec, ok := qs.(*query.WhereOpSpec)
	if !ok {
		return fmt.Errorf("invalid spec type %T", qs)
	}
	w.Exp = spec
	return nil
}

type RangeProcedureSpec struct {
	Bounds BoundsSpec
}

func init() {
	registerProcedureSpec(RangeKind, query.RangeKind, "range", new(RangeProcedureSpec))
}

func (s *RangeProcedureSpec) Kind() ProcedureKind {
	return RangeKind
}

func (s *RangeProcedureSpec) SetSpec(qs query.OperationSpec) error {
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

type ClearProcedureSpec struct {
}

func init() {
	registerProcedureSpec(ClearKind, query.ClearKind, "clear", new(ClearProcedureSpec))
}

func (s *ClearProcedureSpec) Kind() ProcedureKind {
	return ClearKind
}

func (s *ClearProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type WindowProcedureSpec struct {
	Window     WindowSpec
	Triggering query.TriggerSpec
}

func init() {
	registerProcedureSpec(WindowKind, query.WindowKind, "window", new(WindowProcedureSpec))
}

func (s *WindowProcedureSpec) Kind() ProcedureKind {
	return WindowKind
}

func (s *WindowProcedureSpec) SetSpec(qs query.OperationSpec) error {
	ws := qs.(*query.WindowOpSpec)
	s.Window = WindowSpec{
		Every:  ws.Every,
		Period: ws.Period,
		Round:  ws.Round,
		Start:  ws.Start,
	}
	s.Triggering = ws.Triggering
	if s.Triggering == nil {
		s.Triggering = query.DefaultTrigger
	}
	return nil
}

type MergeProcedureSpec struct {
	Keys []string `json:"keys"`
	Keep []string `json:"keep"`
}

func init() {
	registerProcedureSpec(MergeKind, query.MergeKind, "merge", new(MergeProcedureSpec))
}

func (s *MergeProcedureSpec) Kind() ProcedureKind {
	return MergeKind
}

func (s *MergeProcedureSpec) SetSpec(qs query.OperationSpec) error {
	mq := qs.(*query.MergeOpSpec)
	s.Keys = mq.Keys
	s.Keep = mq.Keep
	return nil
}

type KeysProcedureSpec struct {
}

func init() {
	registerProcedureSpec(KeysKind, query.KeysKind, "keys", new(KeysProcedureSpec))
}

func (s *KeysProcedureSpec) Kind() ProcedureKind {
	return KeysKind
}

func (s *KeysProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type ValuesProcedureSpec struct {
}

func init() {
	registerProcedureSpec(ValuesKind, query.ValuesKind, "values", new(ValuesProcedureSpec))
}

func (s *ValuesProcedureSpec) Kind() ProcedureKind {
	return ValuesKind
}

func (s *ValuesProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type CardinalityProcedureSpec struct {
}

func init() {
	registerProcedureSpec(CardinalityKind, query.CardinalityKind, "cardinality", new(CardinalityProcedureSpec))
}

func (s *CardinalityProcedureSpec) Kind() ProcedureKind {
	return CardinalityKind
}

func (s *CardinalityProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type LimitProcedureSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func init() {
	registerProcedureSpec(LimitKind, query.LimitKind, "limit", new(LimitProcedureSpec))
}

func (s *LimitProcedureSpec) Kind() ProcedureKind {
	return LimitKind
}

func (s *LimitProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type ShiftProcedureSpec struct {
	Duration query.Duration `json:"duration"`
}

func init() {
	registerProcedureSpec(ShiftKind, query.ShiftKind, "shift", new(ShiftProcedureSpec))
}

func (s *ShiftProcedureSpec) Kind() ProcedureKind {
	return ShiftKind
}

func (s *ShiftProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type InterpolateProcedureSpec struct {
	Start query.Time     `json:"start"`
	Every query.Duration `json:"every"`
	// TODO define this spec
	Value interface{} `json:"value"`
}

func init() {
	registerProcedureSpec(InterpolateKind, query.InterpolateKind, "interpolate", new(InterpolateProcedureSpec))
}

func (s *InterpolateProcedureSpec) Kind() ProcedureKind {
	return InterpolateKind
}

func (s *InterpolateProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MergeJoinProcedureSpec struct {
	Keys      []string      `json:"keys"`
	Predicate PredicateSpec `json:"predicate"`
}

func init() {
	registerProcedureSpec(MergeJoinKind, query.JoinKind, "merge-join", new(MergeJoinProcedureSpec))
}

func (s *MergeJoinProcedureSpec) Kind() ProcedureKind {
	return MergeJoinKind
}

func (s *MergeJoinProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type UnionProcedureSpec struct {
}

func init() {
	registerProcedureSpec(UnionKind, query.UnionKind, "union", new(UnionProcedureSpec))
}

func (s *UnionProcedureSpec) Kind() ProcedureKind {
	return UnionKind
}

func (s *UnionProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type FilterProcedureSpec struct {
	Predicate PredicateSpec `json:"predicate"`
}

func init() {
	registerProcedureSpec(FilterKind, query.FilterKind, "filter", new(FilterProcedureSpec))
}

func (s *FilterProcedureSpec) Kind() ProcedureKind {
	return FilterKind
}

func (s *FilterProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type SortProcedureSpec struct {
	Predicate PredicateSpec `json:"predicate"`
	Keys      []string      `json:"keys"`
	// TODO define Order type
	Order string `json:"order"` //asc or desc
}

func init() {
	registerProcedureSpec(SortKind, query.SortKind, "sort", new(SortProcedureSpec))
}

func (s *SortProcedureSpec) Kind() ProcedureKind {
	return SortKind
}

func (s *SortProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type RateProcedureSpec struct {
	Unit query.Duration `json:"unit"`
}

func init() {
	registerProcedureSpec(RateKind, query.RateKind, "rate", new(RateProcedureSpec))
}

func (s *RateProcedureSpec) Kind() ProcedureKind {
	return RateKind
}

func (s *RateProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type CountProcedureSpec struct {
}

func init() {
	registerProcedureSpec(CountKind, query.CountKind, "count", new(CountProcedureSpec))
}

func (s *CountProcedureSpec) Kind() ProcedureKind {
	return CountKind
}

func (s *CountProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type SumProcedureSpec struct {
}

func init() {
	registerProcedureSpec(SumKind, query.SumKind, "sum", new(SumProcedureSpec))
}

func (s *SumProcedureSpec) Kind() ProcedureKind {
	return SumKind
}

func (s *SumProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MeanProcedureSpec struct {
}

func init() {
	registerProcedureSpec(MeanKind, query.MeanKind, "mean", new(MeanProcedureSpec))
}

func (s *MeanProcedureSpec) Kind() ProcedureKind {
	return MeanKind
}

func (s *MeanProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type PercentileProcedureSpec struct {
}

func init() {
	registerProcedureSpec(PercentileKind, query.PercentileKind, "percentile", new(PercentileProcedureSpec))
}

func (s *PercentileProcedureSpec) Kind() ProcedureKind {
	return PercentileKind
}

func (s *PercentileProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type StddevProcedureSpec struct {
}

func init() {
	registerProcedureSpec(StddevKind, query.StddevKind, "stddev", new(StddevProcedureSpec))
}

func (s *StddevProcedureSpec) Kind() ProcedureKind {
	return StddevKind
}

func (s *StddevProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MinProcedureSpec struct {
}

func init() {
	registerProcedureSpec(MinKind, query.MinKind, "min", new(MinProcedureSpec))
}

func (s *MinProcedureSpec) Kind() ProcedureKind {
	return MinKind
}

func (s *MinProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type MaxProcedureSpec struct {
}

func init() {
	registerProcedureSpec(MaxKind, query.MaxKind, "max", new(MaxProcedureSpec))
}

func (s *MaxProcedureSpec) Kind() ProcedureKind {
	return MaxKind
}

func (s *MaxProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type TopProcedureSpec struct {
}

func init() {
	registerProcedureSpec(TopKind, query.TopKind, "top", new(TopProcedureSpec))
}

func (s *TopProcedureSpec) Kind() ProcedureKind {
	return TopKind
}

func (s *TopProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}

type DifferenceProcedureSpec struct {
}

func init() {
	registerProcedureSpec(DifferenceKind, query.DifferenceKind, "difference", new(DifferenceProcedureSpec))
}

func (s *DifferenceProcedureSpec) Kind() ProcedureKind {
	return DifferenceKind
}
func (s *DifferenceProcedureSpec) SetSpec(qs query.OperationSpec) error {
	return nil
}
