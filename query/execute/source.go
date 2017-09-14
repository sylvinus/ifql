package execute

import (
	"log"
	"time"

	"github.com/influxdata/ifql/query/plan"
)

type Node interface {
	addTransformation(t Transformation)
}

type Source interface {
	Node
	Run()
}

// storageSource performs storage reads
type storageSource struct {
	id     DatasetID
	reader StorageReader
	spec   *plan.SelectProcedureSpec
	window Window
	bounds Bounds

	ts []Transformation

	currentTime Time
}

func newStorageSource(id DatasetID, r StorageReader, spec *plan.SelectProcedureSpec, now time.Time) Source {
	var w Window
	if spec.WindowSet {
		w = Window{
			Every:  Duration(spec.Window.Every),
			Period: Duration(spec.Window.Period),
			Round:  Duration(spec.Window.Round),
			Start:  Time(spec.Window.Start.Time(now).UnixNano()),
		}
	} else {
		w = Window{
			Every:  Duration(spec.Bounds.Stop.Time(now).UnixNano() - spec.Bounds.Start.Time(now).UnixNano()),
			Period: Duration(spec.Bounds.Stop.Time(now).UnixNano() - spec.Bounds.Start.Time(now).UnixNano()),
			Start:  Time(spec.Bounds.Start.Time(now).UnixNano()),
		}
	}
	currentTime := w.Start + Time(w.Period)
	return &storageSource{
		id:     id,
		reader: r,
		spec:   spec,
		bounds: Bounds{
			Start: Time(spec.Bounds.Start.Time(now).UnixNano()),
			Stop:  Time(spec.Bounds.Stop.Time(now).UnixNano()),
		},
		window:      w,
		currentTime: currentTime,
	}
}

func (s *storageSource) addTransformation(t Transformation) {
	s.ts = append(s.ts, t)
}

func (s *storageSource) Run() {
	for blocks, mark, ok := s.Next(); ok; blocks, mark, ok = s.Next() {
		blocks.Do(func(b Block) {
			for _, t := range s.ts {
				t.Process(s.id, b)
				//TODO(nathanielc): Also add mechanism to send UpdateProcessingTime calls, when no data is arriving.
				// This is probably not needed for this source, but other sources should do so.
				t.UpdateProcessingTime(s.id, Now())
			}
		})
		for _, t := range s.ts {
			t.UpdateWatermark(s.id, mark)
		}
	}
	for _, t := range s.ts {
		t.Finish(s.id)
	}
}

func (s *storageSource) Next() (BlockIterator, Time, bool) {
	start := s.currentTime - Time(s.window.Period)
	stop := s.currentTime

	s.currentTime = s.currentTime + Time(s.window.Every)
	if stop > s.bounds.Stop {
		return nil, 0, false
	}
	bi, err := s.reader.Read(
		s.spec.Database,
		s.spec.Where,
		s.spec.Limit,
		s.spec.Desc,
		start,
		stop,
	)
	if err != nil {
		log.Println("E!", err)
		return nil, 0, false
	}
	return bi, stop, true
}
