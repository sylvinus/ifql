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
		stream: stream,
		blocks: make(chan Block),
	}
	go bi.readBlocks()
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
	stream storage.Storage_ReadClient
	blocks chan Block
}

func (bi *storageBlockIterator) Do(f func(Block)) {
	for b := range bi.blocks {
		f(b)
	}
}

func (bi *storageBlockIterator) readBlocks() {
	defer close(bi.blocks)

	for {
		// Recv the next response
		var rep storage.ReadResponse
		if err := bi.stream.RecvMsg(&rep); err != nil {
			if err == io.EOF {
				return
			}
			//TODO add proper error handling
			return
		}

		builder := NewColListBlockBuilder()
		builder.SetBounds(bi.bounds)

		for _, frame := range rep.Frames {
			if s := frame.GetSeries(); s != nil {
				tags := make(Tags)
				for _, t := range s.Tags {
					tags[string(t.Key)] = string(t.Value)
				}
				builder.SetTags(tags)
				builder.AddCol(TimeCol)
				builder.AddCol(ValueCol)
			} else if p := frame.GetIntegerPoints(); p != nil {
				panic("ints not supported")
			} else if p := frame.GetFloatPoints(); p != nil {
				for i, c := range p.Timestamps {
					builder.AddRow()
					builder.SetTime(0, 0, Time(c))
					builder.SetFloat(0, 1, p.Values[i])
					b := builder.Block()
					bi.blocks <- b
					builder.ClearData()
				}

				// Series is complete, create a new builder with new bound and tags
				builder = NewColListBlockBuilder()
				builder.SetBounds(bi.bounds)
			}
		}
	}
}
