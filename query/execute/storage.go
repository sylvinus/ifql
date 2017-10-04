package execute

import (
	"context"
	"io"
	"sync"

	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/yarpc"
)

type StorageReader interface {
	Read(rs ReadSpec, start, stop Time) (BlockIterator, error)
	Close()
}

type ReadSpec struct {
	Database   string
	Predicate  *storage.Predicate
	Limit      int64
	Descending bool
}

func NewStorageReader() (StorageReader, error) {
	return &storageReader{}, nil
}

type storageReader struct {
	mu          sync.Mutex
	connections []*yarpc.ClientConn
}

func (sr *storageReader) connect() (*yarpc.ClientConn, error) {
	conn, err := yarpc.Dial("localhost:8082")
	if err != nil {
		return nil, err
	}
	sr.mu.Lock()
	sr.connections = append(sr.connections, conn)
	sr.mu.Unlock()
	return conn, nil
}

func (sr *storageReader) Read(readSpec ReadSpec, start, stop Time) (BlockIterator, error) {
	conn, err := sr.connect()
	if err != nil {
		return nil, err
	}
	c := storage.NewStorageClient(conn)

	var req storage.ReadRequest
	req.Database = readSpec.Database
	req.Predicate = readSpec.Predicate
	//	req.Limit = limit
	req.Descending = readSpec.Descending
	req.TimestampRange.Start = int64(start)
	req.TimestampRange.End = int64(stop)

	stream, err := c.Read(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	bi := &storageBlockIterator{
		bounds: Bounds{
			Start: start,
			Stop:  stop,
		},
		data: &readState{
			stream: stream,
		},
	}
	return bi, nil
}

func (sr *storageReader) Close() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, c := range sr.connections {
		c.Close()
	}
}

type storageBlockIterator struct {
	bounds Bounds
	data   *readState
}

type readState struct {
	stream storage.Storage_ReadClient
	rep    storage.ReadResponse
}

type responseType int

const (
	seriesType responseType = iota
	integerPointsType
	floatPointsType
)

func (s *readState) peek() responseType {
	frame := s.rep.Frames[0]
	switch {
	case frame.GetSeries() != nil:
		return seriesType
	case frame.GetIntegerPoints() != nil:
		return integerPointsType
	case frame.GetFloatPoints() != nil:
		return floatPointsType
	default:
		panic("read response frame should have one of series, integerPoints, or floatPoints")
	}
}

func (s *readState) more() bool {
	if len(s.rep.Frames) > 0 {
		return true
	}
	if err := s.stream.RecvMsg(&s.rep); err != nil {
		if err == io.EOF {
			// We are done
			return false
		}
		//TODO add proper error handling
		return false
	}
	return true
}

func (s *readState) next() storage.ReadResponse_Frame {
	frame := s.rep.Frames[0]
	s.rep.Frames = s.rep.Frames[1:]
	return frame
}

func (bi *storageBlockIterator) Do(f func(Block)) {
	for bi.data.more() {
		if p := bi.data.peek(); p != seriesType {
			//This means the consumer didn't read all the data off the block
			continue
		}
		frame := bi.data.next()
		s := frame.GetSeries()
		tags := make(Tags)
		for _, t := range s.Tags {
			tags[string(t.Key)] = string(t.Value)
		}
		block := &storageBlock{
			bounds:  bi.bounds,
			tags:    tags,
			colMeta: []ColMeta{TimeCol, ValueCol},
			data:    bi.data,
			done:    make(chan struct{}),
		}
		f(block)
		// Wait until the block has been read.
		block.wait()
	}
}

type storageBlock struct {
	bounds  Bounds
	tags    Tags
	colMeta []ColMeta

	done chan struct{}

	data *readState
}

func (b *storageBlock) wait() {
	<-b.done
}

func (b *storageBlock) Bounds() Bounds {
	return b.bounds
}
func (b *storageBlock) Tags() Tags {
	return b.tags
}
func (b *storageBlock) Cols() []ColMeta {
	return b.colMeta
}

func (b *storageBlock) Col(c int) ValueIterator {
	return &storageBlockValueIterator{
		data: b.data,
		col:  b.colMeta[c],
		done: b.done,
	}
}

func (b *storageBlock) Values() ValueIterator {
	valueIdx := ValueIdx(b)
	return b.Col(valueIdx)
}
func (b *storageBlock) Times() ValueIterator {
	timeIdx := TimeIdx(b)
	return b.Col(timeIdx)
}

// TODO(nathanielc): Maybe change this API to be part of the ValueIterator so you can scan a single column, and then get data horizonatally as needed.
func (b *storageBlock) AtFloat(i, j int) float64 {
	panic("not implemented")
}
func (b *storageBlock) AtString(i, j int) string {
	panic("not implemented")
}
func (b *storageBlock) AtTime(i, j int) Time {
	panic("not implemented")
}

type storageBlockValueIterator struct {
	data *readState
	col  ColMeta

	done chan<- struct{}

	floatBuf  []float64
	stringBuf []string
	timeBuf   []Time
}

func (b *storageBlockValueIterator) DoFloat(f func([]float64)) {
	checkColType(b.col, TFloat)
	for b.advance() {
		f(b.floatBuf)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoString(f func([]string)) {
	checkColType(b.col, TString)
	for b.advance() {
		f(b.stringBuf)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoTime(f func([]Time)) {
	checkColType(b.col, TTime)
	for b.advance() {
		f(b.timeBuf)
	}
	close(b.done)
}

func (b *storageBlockValueIterator) advance() bool {
	for b.data.more() {
		switch b.data.peek() {
		case seriesType:
			return false
		case integerPointsType:
			panic("integers not supported")
		case floatPointsType:
			// read next frame
			frame := b.data.next()
			p := frame.GetFloatPoints()

			//reset buffers
			b.timeBuf = b.timeBuf[0:0]
			b.floatBuf = b.floatBuf[0:0]
			b.stringBuf = b.stringBuf[0:0]

			for i, c := range p.Timestamps {
				b.timeBuf = append(b.timeBuf, Time(c))
				b.floatBuf = append(b.floatBuf, p.Values[i])
			}
			return true
		}
	}
	return false
}
