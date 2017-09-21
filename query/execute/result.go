package execute

type Result interface {
	Blocks() BlockIterator
}

// resultSink implements both the Transformation and Result interfaces,
// mapping the pushed based Transformation API to the pull based Result interface.
type resultSink struct {
	blocks chan Block
	closed bool

	parents []DatasetID
}

func newResultSink() *resultSink {
	return &resultSink{
		// TODO(nathanielc): Currently this buffer needs to be big enough hold all result blocks :(
		blocks: make(chan Block, 1000),
	}
}

func (s *resultSink) RetractBlock(DatasetID, BlockMetadata) {
	//TODO implement
}

func (s *resultSink) Process(id DatasetID, b Block) {
	s.blocks <- b
}

func (s *resultSink) Blocks() BlockIterator {
	return s
}

func (s *resultSink) Do(f func(Block)) {
	for b := range s.blocks {
		f(b)
	}
}

func (s *resultSink) UpdateWatermark(id DatasetID, mark Time) {
	//Nothing to do
}
func (s *resultSink) UpdateProcessingTime(id DatasetID, t Time) {
	//Nothing to do
}
func (s *resultSink) setParents(ids []DatasetID) {
	s.parents = ids
}

func (s *resultSink) setTrigger(Trigger) {
	//TODO: Change interfaces so that resultSink, does not need to implement this method.
}

func (s *resultSink) Finish(id DatasetID) {
	close(s.blocks)
}
