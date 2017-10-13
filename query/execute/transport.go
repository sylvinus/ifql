package execute

import (
	"context"
)

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

type TransformationTransport struct {
	t Transformation

	buf    chan Message
	cancel chan struct{}
}

func newTransformationTransport(t Transformation) *TransformationTransport {
	return &TransformationTransport{
		t:      t,
		buf:    make(chan Message, 100),
		cancel: make(chan struct{}),
	}
}

func (t *TransformationTransport) Run(ctx context.Context) {
	defer close(t.cancel)
	for {
		select {
		case <-ctx.Done():
			// Immediately return, do not process any buffered messages
			return
		case m := <-t.buf:
			switch m := m.(type) {
			case RetractBlockMsg:
				t.t.RetractBlock(m.SrcDatasetID(), m.BlockMetadata())
			case ProcessMsg:
				t.t.Process(m.SrcDatasetID(), m.Block())
			case UpdateWatermarkMsg:
				t.t.UpdateWatermark(m.SrcDatasetID(), m.WatermarkTime())
			case UpdateProcessingTimeMsg:
				t.t.UpdateProcessingTime(m.SrcDatasetID(), m.ProcessingTime())
			case FinishMsg:
				t.t.Finish(m.SrcDatasetID(), m.Error())
				// Stop processing messages
				return
			}
		}
	}
}

func (t *TransformationTransport) RetractBlock(id DatasetID, meta BlockMetadata) {
	select {
	case t.buf <- &retractBlockMsg{
		srcMessage:    srcMessage(id),
		blockMetadata: meta,
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) Process(id DatasetID, b Block) {
	select {
	case t.buf <- &processMsg{
		srcMessage: srcMessage(id),
		block:      b,
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) UpdateWatermark(id DatasetID, time Time) {
	select {
	case t.buf <- &updateWatermarkMsg{
		srcMessage: srcMessage(id),
		time:       time,
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) UpdateProcessingTime(id DatasetID, time Time) {
	select {
	case t.buf <- &updateProcessingTimeMsg{
		srcMessage: srcMessage(id),
		time:       time,
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) Finish(id DatasetID, err error) {
	select {
	case t.buf <- &finishMsg{
		srcMessage: srcMessage(id),
		err:        err,
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) SetParents(ids []DatasetID) {
	//TODO(nathanielc): Need a better mechanism to inform transformations of their parent ids.
	t.t.SetParents(ids)
}
