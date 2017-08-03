package query

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// Operation denotes a single operation in a query.
type Operation struct {
	ID   OperationID   `json:"id"`
	Spec OperationSpec `json:"spec"`
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	type operationJSON struct {
		ID   OperationID     `json:"id"`
		Kind OperationKind   `json:"kind"`
		Spec json.RawMessage `json:"spec"`
	}
	oj := operationJSON{}
	err := json.Unmarshal(data, &oj)
	if err != nil {
		return err
	}
	o.ID = oj.ID
	spec, err := unmarshalOpSpec(oj.ID, oj.Kind, oj.Spec)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal operation %q", oj.ID)
	}
	o.Spec = spec
	return nil
}

func unmarshalOpSpec(id OperationID, k OperationKind, data []byte) (OperationSpec, error) {
	if int(k) > len(kindToGoType) {
		return nil, fmt.Errorf("unknown operation spec kind %v", k)
	}
	t := kindToGoType[k]
	if t == nil {
		return nil, fmt.Errorf("unknown operation spec kind %v", k)
	}
	spec := reflect.New(t).Interface()
	op := spec.(OperationSpec)

	if len(data) > 0 {
		err := json.Unmarshal(data, op)
		if err != nil {
			return nil, err
		}
	}
	return op, nil
}

func (o Operation) MarshalJSON() ([]byte, error) {
	type operationJSON struct {
		ID   OperationID   `json:"id"`
		Kind OperationKind `json:"kind"`
		Spec OperationSpec `json:"spec"`
	}
	oj := operationJSON{
		ID:   o.ID,
		Kind: o.Spec.Kind(),
		Spec: o.Spec,
	}
	return json.Marshal(oj)
}

// OperationSpec specifies an operation as part of a query.
type OperationSpec interface {
	// Kind returns the kind of the operation.
	Kind() OperationKind
}

// OperationID is a unique ID within a query for the operation.
type OperationID string

// OperationKind denotes the kind of operations.
type OperationKind int

func (k OperationKind) String() string {
	return kindToNames[k]
}
func (k *OperationKind) UnmarshalText(data []byte) error {
	kind, ok := namesToKind[string(data)]
	if !ok {
		return fmt.Errorf("unknown kind %q", string(data))
	}
	*k = kind
	return nil
}

func (k OperationKind) MarshalText() ([]byte, error) {
	return []byte(kindToNames[k]), nil
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

var operationSpecType = reflect.TypeOf((*OperationSpec)(nil)).Elem()

// registerOpSpec populates package vars with lookup information about the OperationSpec.
func registerOpSpec(k OperationKind, name string, s OperationSpec) {
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
}

type SelectOpSpec struct {
	Database string         `json:"database"`
	Where    ExpressionSpec `json:"where"`
}

func init() {
	registerOpSpec(SelectKind, "select", new(SelectOpSpec))
}

func (s *SelectOpSpec) Kind() OperationKind {
	return SelectKind
}

type RangeOpSpec struct {
	Start Time `json:"start"`
	Stop  Time `json:"stop"`
}

func init() {
	registerOpSpec(RangeKind, "range", new(RangeOpSpec))
}

func (s *RangeOpSpec) Kind() OperationKind {
	return RangeKind
}

type ClearOpSpec struct {
}

func init() {
	registerOpSpec(ClearKind, "clear", new(ClearOpSpec))
}

func (s *ClearOpSpec) Kind() OperationKind {
	return ClearKind
}

type WindowOpSpec struct {
	Every       Duration `json:"every"`
	Period      Duration `json:"period"`
	EveryCount  int64    `json:"every_count"`
	PeriodCount int64    `json:"period_count"`
	Start       Time     `json:"start"`
	Round       Duration `json:"round"`
}

func init() {
	registerOpSpec(WindowKind, "window", new(WindowOpSpec))
}

func (s *WindowOpSpec) Kind() OperationKind {
	return WindowKind
}

type MergeOpSpec struct {
	Keys []string `json:"keys"`
	Keep []string `json:"keep"`
}

func init() {
	registerOpSpec(MergeKind, "merge", new(MergeOpSpec))
}

func (s *MergeOpSpec) Kind() OperationKind {
	return MergeKind
}

type KeysOpSpec struct {
}

func init() {
	registerOpSpec(KeysKind, "keys", new(KeysOpSpec))
}

func (s *KeysOpSpec) Kind() OperationKind {
	return KeysKind
}

type ValuesOpSpec struct {
}

func init() {
	registerOpSpec(ValuesKind, "values", new(ValuesOpSpec))
}

func (s *ValuesOpSpec) Kind() OperationKind {
	return ValuesKind
}

type CardinalityOpSpec struct {
}

func init() {
	registerOpSpec(CardinalityKind, "cardinality", new(CardinalityOpSpec))
}

func (s *CardinalityOpSpec) Kind() OperationKind {
	return CardinalityKind
}

type LimitOpSpec struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func init() {
	registerOpSpec(LimitKind, "limit", new(LimitOpSpec))
}

func (s *LimitOpSpec) Kind() OperationKind {
	return LimitKind
}

type ShiftOpSpec struct {
	Duration Duration `json:"duration"`
}

func init() {
	registerOpSpec(ShiftKind, "shift", new(ShiftOpSpec))
}

func (s *ShiftOpSpec) Kind() OperationKind {
	return ShiftKind
}

type InterpolateOpSpec struct {
	Start Time     `json:"start"`
	Every Duration `json:"every"`
	// TODO define this spec
	Value interface{} `json:"value"`
}

func init() {
	registerOpSpec(InterpolateKind, "interpolate", new(InterpolateOpSpec))
}

func (s *InterpolateOpSpec) Kind() OperationKind {
	return InterpolateKind
}

type JoinOpSpec struct {
	Keys       []string       `json:"keys"`
	Expression ExpressionSpec `json:"expression"`
}

func init() {
	registerOpSpec(JoinKind, "join", new(JoinOpSpec))
}

func (s *JoinOpSpec) Kind() OperationKind {
	return JoinKind
}

type UnionOpSpec struct {
}

func init() {
	registerOpSpec(UnionKind, "union", new(UnionOpSpec))
}

func (s *UnionOpSpec) Kind() OperationKind {
	return UnionKind
}

type FilterOpSpec struct {
	Expression ExpressionSpec `json:"expression"`
}

func init() {
	registerOpSpec(FilterKind, "filter", new(FilterOpSpec))
}

func (s *FilterOpSpec) Kind() OperationKind {
	return FilterKind
}

type SortOpSpec struct {
	Expression ExpressionSpec `json:"expression"`
	Keys       []string       `json:"keys"`
	// TODO define Order type
	Order string `json:"order"` //asc or desc
}

func init() {
	registerOpSpec(SortKind, "sort", new(SortOpSpec))
}

func (s *SortOpSpec) Kind() OperationKind {
	return SortKind
}

type RateOpSpec struct {
	Unit Duration `json:"unit"`
}

func init() {
	registerOpSpec(RateKind, "rate", new(RateOpSpec))
}

func (s *RateOpSpec) Kind() OperationKind {
	return RateKind
}

type CountOpSpec struct {
}

func init() {
	registerOpSpec(CountKind, "count", new(CountOpSpec))
}

func (s *CountOpSpec) Kind() OperationKind {
	return CountKind
}

type SumOpSpec struct {
}

func init() {
	registerOpSpec(SumKind, "sum", new(SumOpSpec))
}

func (s *SumOpSpec) Kind() OperationKind {
	return SumKind
}

type MeanOpSpec struct {
}

func init() {
	registerOpSpec(MeanKind, "mean", new(MeanOpSpec))
}

func (s *MeanOpSpec) Kind() OperationKind {
	return MeanKind
}

type PercentileOpSpec struct {
}

func init() {
	registerOpSpec(PercentileKind, "percentile", new(PercentileOpSpec))
}

func (s *PercentileOpSpec) Kind() OperationKind {
	return PercentileKind
}

type StddevOpSpec struct {
}

func init() {
	registerOpSpec(StddevKind, "stddev", new(StddevOpSpec))
}

func (s *StddevOpSpec) Kind() OperationKind {
	return StddevKind
}

type MinOpSpec struct {
}

func init() {
	registerOpSpec(MinKind, "min", new(MinOpSpec))
}

func (s *MinOpSpec) Kind() OperationKind {
	return MinKind
}

type MaxOpSpec struct {
}

func init() {
	registerOpSpec(MaxKind, "max", new(MaxOpSpec))
}

func (s *MaxOpSpec) Kind() OperationKind {
	return MaxKind
}

type TopOpSpec struct {
}

func init() {
	registerOpSpec(TopKind, "top", new(TopOpSpec))
}

func (s *TopOpSpec) Kind() OperationKind {
	return TopKind
}

type DifferenceOpSpec struct {
}

func init() {
	registerOpSpec(DifferenceKind, "difference", new(DifferenceOpSpec))
}

func (s *DifferenceOpSpec) Kind() OperationKind {
	return DifferenceKind
}
