package execute

type Result interface {
	Blocks() BlockIterator
}

// resultSink implements both the Transformation and Result interfaces,
// mapping the pushed based Transformation API to the pull based Result interface.
type resultSink struct {
	blocks chan Block
}

func newResultSink() *resultSink {
	return &resultSink{
		// TODO(nathanielc): Currently this buffer needs to be big enough hold all result blocks :(
		blocks: make(chan Block, 1000),
	}
}

func (s *resultSink) RetractBlock(BlockMetadata) {
	//TODO implement
}

func (s *resultSink) Process(b Block) {
	s.blocks <- b
}

func (s *resultSink) Blocks() BlockIterator {
	return s
}

func (s *resultSink) NextBlock() (Block, bool) {
	b, ok := <-s.blocks
	return b, ok
}

func (s *resultSink) UpdateWatermark(mark Time) {
	//Nothing to do
}
func (s *resultSink) UpdateProcessingTime(t Time) {
	//Nothing to do
}

func (s *resultSink) setTrigger(Trigger) {
	//TODO: Change interfaces so that resultSink, does not need to implement this method.
}

func (s *resultSink) Finish() {
	close(s.blocks)
}
