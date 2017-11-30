package execute

import (
	"sync/atomic"
)

type Transport interface {
	Transformation
	// Finished reports when the Transport has completed and there is no more work to do.
	Finished() <-chan struct{}
}

// consecutiveTransport implements Transport by transporting data consecutively to the downstream Transformation.
type consecutiveTransport struct {
	dispatcher Dispatcher

	t        Transformation
	messages MessageQueue

	finished chan struct{}

	schedulerState int32
	inflight       int32
}

func newConescutiveTransport(dispatcher Dispatcher, t Transformation) *consecutiveTransport {
	return &consecutiveTransport{
		dispatcher: dispatcher,
		t:          t,
		// TODO(nathanielc): Have planner specify message queue initial buffer size.
		messages: newMessageQueue(64),
		finished: make(chan struct{}),
	}
}

func (t *consecutiveTransport) Finished() <-chan struct{} {
	return t.finished
}

func (t *consecutiveTransport) RetractBlock(id DatasetID, meta BlockMetadata) {
	t.pushMsg(&retractBlockMsg{
		srcMessage:    srcMessage(id),
		blockMetadata: meta,
	})
}

func (t *consecutiveTransport) Process(id DatasetID, b Block) {
	t.pushMsg(&processMsg{
		srcMessage: srcMessage(id),
		block:      b,
	})
}

func (t *consecutiveTransport) UpdateWatermark(id DatasetID, time Time) {
	t.pushMsg(&updateWatermarkMsg{
		srcMessage: srcMessage(id),
		time:       time,
	})
}

func (t *consecutiveTransport) UpdateProcessingTime(id DatasetID, time Time) {
	t.pushMsg(&updateProcessingTimeMsg{
		srcMessage: srcMessage(id),
		time:       time,
	})
}

func (t *consecutiveTransport) Finish(id DatasetID, err error) {
	t.pushMsg(&finishMsg{
		srcMessage: srcMessage(id),
		err:        err,
	})
}

func (t *consecutiveTransport) SetParents(ids []DatasetID) {
	//TODO(nathanielc): Need a better mechanism to inform transformations of their parent ids.
	t.t.SetParents(ids)
}

func (t *consecutiveTransport) pushMsg(m Message) {
	t.messages.Push(m)
	atomic.AddInt32(&t.inflight, 1)
	t.schedule()
}

const (
	// consecutiveTransport schedule states
	idle int32 = iota
	running
	finished
)

// schedule indicates that there is work available to schedule.
func (t *consecutiveTransport) schedule() {
	if t.tryTransition(idle, running) {
		t.dispatcher.Schedule(t.processMessages)
	}
}

// tryTransition attempts to transition into the new state and returns true on success.
func (t *consecutiveTransport) tryTransition(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&t.schedulerState, old, new)
}

// transition sets the new state.
func (t *consecutiveTransport) transition(new int32) {
	atomic.StoreInt32(&t.schedulerState, new)
}

func (t *consecutiveTransport) processMessages(throughput int) {
PROCESS:
	i := 0
	for m := t.messages.Pop(); m != nil; m = t.messages.Pop() {
		atomic.AddInt32(&t.inflight, -1)
		if processMessage(t.t, m) {
			// Transition to the finished state.
			if t.tryTransition(running, finished) {
				// We are finished
				close(t.finished)
				return
			}
		}
		i++
		if i >= throughput {
			// We have done enough work.
			// Transition to the idle state and reschedule for later.
			t.transition(idle)
			t.schedule()
			return
		}
	}

	t.transition(idle)
	// Check if more messages arrived after the above loop finished.
	// This check must happen in the idle state.
	if atomic.LoadInt32(&t.inflight) > 0 {
		if t.tryTransition(idle, running) {
			goto PROCESS
		} // else we have already been scheduled again, we can return
	}
}

// processMessage processes the message on t.
// The return value is true if the message was a FinishMsg.
func processMessage(t Transformation, m Message) bool {
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
		return true
	}
	return false
}

type Message interface {
	Type() MessageType
	SrcDatasetID() DatasetID
}

type MessageType int

const (
	RetractBlockType MessageType = iota
	ProcessType
	UpdateWatermarkType
	UpdateProcessingTimeType
	FinishType
)

type srcMessage DatasetID

func (m srcMessage) SrcDatasetID() DatasetID {
	return DatasetID(m)
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
