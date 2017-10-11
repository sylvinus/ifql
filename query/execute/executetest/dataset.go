package executetest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	uuid "github.com/satori/go.uuid"
)

func RandomDatasetID() execute.DatasetID {
	return execute.DatasetID(uuid.NewV4())
}

type Dataset struct {
	ID                    execute.DatasetID
	Retractions           []execute.BlockKey
	ProcessingTimeUpdates []execute.Time
	WatermarkUpdates      []execute.Time
	Finished              bool
}

func NewDataset(id execute.DatasetID) *Dataset {
	return &Dataset{
		ID: id,
	}
}

func (d *Dataset) AddTransformation(t execute.Transformation) {
	panic("not implemented")
}

func (d *Dataset) RetractBlock(key execute.BlockKey) {
	d.Retractions = append(d.Retractions, key)
}

func (d *Dataset) UpdateProcessingTime(t execute.Time) {
	d.ProcessingTimeUpdates = append(d.ProcessingTimeUpdates, t)
}

func (d *Dataset) UpdateWatermark(mark execute.Time) {
	d.WatermarkUpdates = append(d.WatermarkUpdates, mark)
}

func (d *Dataset) Finish() {
	if d.Finished {
		panic("finish has already been called")
	}
	d.Finished = true
}

func (d *Dataset) SetTriggerSpec(t query.TriggerSpec) {
	panic("not implemented")
}

type NewTransformation func(execute.Dataset, execute.BlockBuilderCache) execute.Transformation

func TransformationPassThroughTestHelper(t *testing.T, newTr NewTransformation) {
	t.Helper()

	now := execute.Now()
	d := NewDataset(RandomDatasetID())
	c := execute.NewBlockBuilderCache()
	c.SetTriggerSpec(execute.DefaultTriggerSpec)

	parentID := RandomDatasetID()
	tr := newTr(d, c)
	tr.UpdateWatermark(parentID, now)
	tr.UpdateProcessingTime(parentID, now)
	tr.Finish(parentID)

	exp := &Dataset{
		ID: d.ID,
		ProcessingTimeUpdates: []execute.Time{now},
		WatermarkUpdates:      []execute.Time{now},
		Finished:              true,
	}
	if !cmp.Equal(d, exp) {
		t.Errorf("unexpected dataset -want/+got\n%s", cmp.Diff(exp, d))
	}
}
