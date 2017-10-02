package execute

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
}

type finishMsg struct {
	srcMessage
}

func (m *finishMsg) Type() MessageType {
	return FinishType
}

type TransformationTransport struct {
	t Transformation

	buf    chan Message
	cancel <-chan struct{}
}

func newTransformationTransport(t Transformation, cancel <-chan struct{}) *TransformationTransport {
	return &TransformationTransport{
		t:      t,
		buf:    make(chan Message, 100),
		cancel: cancel,
	}
}

func (t *TransformationTransport) Run() {
	for {
		select {
		case <-t.cancel:
			return
		case m := <-t.buf:
			switch msg := m.(type) {
			case RetractBlockMsg:
				t.t.RetractBlock(msg.SrcDatasetID(), msg.BlockMetadata())
			case ProcessMsg:
				t.t.Process(msg.SrcDatasetID(), msg.Block())
			case UpdateWatermarkMsg:
				t.t.UpdateWatermark(msg.SrcDatasetID(), msg.WatermarkTime())
			case UpdateProcessingTimeMsg:
				t.t.UpdateProcessingTime(msg.SrcDatasetID(), msg.ProcessingTime())
			case FinishMsg:
				t.t.Finish(msg.SrcDatasetID())
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

func (t *TransformationTransport) Finish(id DatasetID) {
	select {
	case t.buf <- &finishMsg{
		srcMessage: srcMessage(id),
	}:
	case <-t.cancel:
	}
}

func (t *TransformationTransport) SetParents(ids []DatasetID) {
	//TODO(nathanielc): Need a better mechanism to inform transformations of their parent ids.
	t.t.SetParents(ids)
}
