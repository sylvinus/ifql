package plan

import (
	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

type DatasetID uuid.UUID

func (did DatasetID) String() string {
	return uuid.UUID(did).String()
}

var ZeroDatasetID DatasetID

type Dataset struct {
	ID          DatasetID
	Bounds      BoundsSpec
	Window      WindowSpec
	Source      ProcedureID
	Destination ProcedureID
}

func (d *Dataset) MakeNarrowChild(pid ProcedureID, name string) *Dataset {
	c := new(Dataset)
	*c = *d
	c.ID = CreateDatasetID(pid, name)
	c.Source = pid
	return c
}

type BoundsSpec struct {
	Start query.Time
	Stop  query.Time
}

type WindowSpec struct {
	Every  query.Duration
	Period query.Duration
	Round  query.Duration
	Start  query.Time
}
