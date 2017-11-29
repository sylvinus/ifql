package execute

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type CentralTransport struct {
	buf chan Message

	mu      sync.Mutex
	closed  bool
	closing chan struct{}
	wg      sync.WaitGroup
	err     error

	transformations map[transformationID]Transformation
	finishedCount   int32
}

func newCentralTransport() *CentralTransport {
	return &CentralTransport{
		buf:             make(chan Message, 100),
		closing:         make(chan struct{}),
		transformations: make(map[transformationID]Transformation),
	}
}

type transformationID uint32

func (ct *CentralTransport) Wrap(t Transformation) Transformation {
	id := transformationID(len(ct.transformations))
	ct.transformations[id] = t
	return newTransformationTransport(ct, id)
}

func (ct *CentralTransport) Start(n int, ctx context.Context) {
	ct.wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer ct.wg.Done()
			// Setup panic handling on the worker goroutines
			defer func() {
				if e := recover(); e != nil {
					// We had a panic, abort the entire execution.
					var err error
					switch e := e.(type) {
					case error:
						err = e
					default:
						err = fmt.Errorf("%v", e)
					}
					ct.setErr(fmt.Errorf("panic: %v\n%s", err, debug.Stack()))
				}
			}()
			ct.run(ctx)
		}()
	}
}

func (ct *CentralTransport) Err() error {
	ct.wg.Wait()
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.err
}

func (ct *CentralTransport) setErr(err error) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	// TODO(nathanielc): Collect all error information.
	if ct.err == nil {
		ct.err = err
	}
}

func (ct *CentralTransport) stop() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	if ct.closed {
		return
	}
	ct.closed = true
	close(ct.closing)
}

func (ct *CentralTransport) incFinished() {
	atomic.AddInt32(&ct.finishedCount, 1)
}
func (ct *CentralTransport) finished() int {
	return int(atomic.LoadInt32(&ct.finishedCount))
}

func (ct *CentralTransport) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Immediately return, do not process any buffered messages
			return
		case <-ct.closing:
			// We are done, nothing left to do.
			return
		case m := <-ct.buf:
			t := ct.transformations[m.TransformationID()]
			switch m := m.(type) {
			case RetractBlockMsg:
				t.RetractBlock(m.SrcDatasetID(), m.BlockMetadata())
			case ProcessMsg:
				t.Process(m.SrcDatasetID(), m.Block())
			case UpdateWatermarkMsg:
				t.UpdateWatermark(m.SrcDatasetID(), m.WatermarkTime())
			case UpdateProcessingTimeMsg:
				t.UpdateProcessingTime(m.SrcDatasetID(), m.ProcessingTime())
			case FinishMsg:
				t.Finish(m.SrcDatasetID(), m.Error())
				ct.incFinished()

				// Check if all transformations are finished
				if ct.finished() == len(ct.transformations) {
					ct.stop()
					return
				}
			}
		}
	}
}

type TransformationTransport struct {
	ct  *CentralTransport
	tID transformationID
}

func newTransformationTransport(ct *CentralTransport, tID transformationID) *TransformationTransport {
	return &TransformationTransport{
		ct:  ct,
		tID: tID,
	}
}

func (t *TransformationTransport) RetractBlock(id DatasetID, meta BlockMetadata) {
	select {
	case t.ct.buf <- &retractBlockMsg{
		srcMessage: srcMessage{
			dsID: id,
			tID:  t.tID,
		},
		blockMetadata: meta,
	}:
	case <-t.ct.closing:
	}
}

func (t *TransformationTransport) Process(id DatasetID, b Block) {
	select {
	case t.ct.buf <- &processMsg{
		srcMessage: srcMessage{
			dsID: id,
			tID:  t.tID,
		},
		block: b,
	}:
	case <-t.ct.closing:
	}
}

func (t *TransformationTransport) UpdateWatermark(id DatasetID, time Time) {
	select {
	case t.ct.buf <- &updateWatermarkMsg{
		srcMessage: srcMessage{
			dsID: id,
			tID:  t.tID,
		},
		time: time,
	}:
	case <-t.ct.closing:
	}
}

func (t *TransformationTransport) UpdateProcessingTime(id DatasetID, time Time) {
	select {
	case t.ct.buf <- &updateProcessingTimeMsg{
		srcMessage: srcMessage{
			dsID: id,
			tID:  t.tID,
		},
		time: time,
	}:
	case <-t.ct.closing:
	}
}

func (t *TransformationTransport) Finish(id DatasetID, err error) {
	select {
	case t.ct.buf <- &finishMsg{
		srcMessage: srcMessage{
			dsID: id,
			tID:  t.tID,
		},
		err: err,
	}:
	case <-t.ct.closing:
	}
}

func (t *TransformationTransport) SetParents(ids []DatasetID) {
	//TODO(nathanielc): Need a better mechanism to inform transformations of their parent ids.
	t.ct.transformations[t.tID].SetParents(ids)
}

type Message interface {
	Type() MessageType
	SrcDatasetID() DatasetID
	TransformationID() transformationID
}

type MessageType int

const (
	RetractBlockType MessageType = iota
	ProcessType
	UpdateWatermarkType
	UpdateProcessingTimeType
	FinishType
)

type srcMessage struct {
	dsID DatasetID
	tID  transformationID
}

func (m srcMessage) SrcDatasetID() DatasetID {
	return m.dsID
}
func (m srcMessage) TransformationID() transformationID {
	return m.tID
}

type RetractBlockMsg interface {
	Message
	BlockMetadata() BlockMetadata
}

type retractBlockMsg struct {
	srcMessage
	blockMetadata BlockMetadata
}

func (m *retractBlockMsg) Type() MessageType {
	return RetractBlockType
}
func (m *retractBlockMsg) BlockMetadata() BlockMetadata {
	return m.blockMetadata
}

type ProcessMsg interface {
	Message
	Block() Block
}

type processMsg struct {
	srcMessage
	block Block
}

func (m *processMsg) Type() MessageType {
	return ProcessType
}
func (m *processMsg) Block() Block {
	return m.block
}

type UpdateWatermarkMsg interface {
	Message
	WatermarkTime() Time
}

type updateWatermarkMsg struct {
	srcMessage
	time Time
}

func (m *updateWatermarkMsg) Type() MessageType {
	return UpdateWatermarkType
}
func (m *updateWatermarkMsg) WatermarkTime() Time {
	return m.time
}

type UpdateProcessingTimeMsg interface {
	Message
	ProcessingTime() Time
}

type updateProcessingTimeMsg struct {
	srcMessage
	time Time
}

func (m *updateProcessingTimeMsg) Type() MessageType {
	return UpdateProcessingTimeType
}
func (m *updateProcessingTimeMsg) ProcessingTime() Time {
	return m.time
}

type FinishMsg interface {
	Message
	Error() error
}

type finishMsg struct {
	srcMessage
	err error
}

func (m *finishMsg) Type() MessageType {
	return FinishType
}
func (m *finishMsg) Error() error {
	return m.err
}
