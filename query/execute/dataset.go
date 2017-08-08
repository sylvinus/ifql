package execute

import (
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
)

type Dataset interface {
	Window() Window
	Frames() FrameIterator
}

type Window interface {
	Every() query.Duration
	Period() query.Duration
	Round() query.Duration
	Start() query.Time
}

type window struct {
	every  query.Duration
	period query.Duration
	round  query.Duration
	start  query.Time
}

func (w window) Every() query.Duration {
	return w.every
}
func (w window) Period() query.Duration {
	return w.period
}
func (w window) Round() query.Duration {
	return w.round
}
func (w window) Start() query.Time {
	return w.start
}

type FrameIterator interface {
	NextFrame() (DataFrame, bool)
}

type dataset struct {
	parent FrameIterator
	op     Operation
	window Window
}

func (d *dataset) Window() Window {
	return d.window
}

func (d *dataset) Frames() FrameIterator {
	return d
}

func (d *dataset) NextFrame() (DataFrame, bool) {
	src, ok := d.parent.NextFrame()
	if !ok {
		return nil, false
	}
	return d.op.Do(src)
}

type readDataset struct {
	reader StorageReader
	spec   *plan.SelectOpSpec
}

func (d *readDataset) Window() Window {
	return nil
}
func (d *readDataset) Frames() FrameIterator {
	return d
}

func (d *readDataset) NextFrame() (DataFrame, bool) {
	//TODO push down bounds into readDataset
	return d.reader.Read(d.spec.Database, 0, 1e10)
}
