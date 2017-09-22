package execute

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Transformation interface {
	RetractBlock(id DatasetID, meta BlockMetadata)
	Process(id DatasetID, b Block)
	UpdateWatermark(id DatasetID, t Time)
	UpdateProcessingTime(id DatasetID, t Time)
	Finish(id DatasetID)
	SetParents(ids []DatasetID)
}

type CreateTransformation func(id DatasetID, mode AccumulationMode, spec plan.ProcedureSpec, now time.Time) (Transformation, Dataset, error)

var procedureToTransformation = make(map[plan.ProcedureKind]CreateTransformation)

func RegisterTransformation(k plan.ProcedureKind, c CreateTransformation) {
	if procedureToTransformation[k] != nil {
		panic(fmt.Errorf("duplicate registration for transformation with procedure kind %v", k))
	}
	procedureToTransformation[k] = c
}
