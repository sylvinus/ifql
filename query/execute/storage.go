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

	newBuilder := func() BlockBuilder {
		builder := NewColListBlockBuilder()
		builder.SetBounds(bi.bounds)
		builder.AddCol(TimeCol)
		builder.AddCol(ValueCol)
		return builder
	}

	var builder BlockBuilder

	for {
		// Recv the next response
		var rep storage.ReadResponse
		if err := bi.stream.RecvMsg(&rep); err != nil {
			if err == io.EOF {
				if builder != nil {
					b := builder.Block()
					bi.blocks <- b
				}
				return
			}
			//TODO add proper error handling
			return
		}

		for _, frame := range rep.Frames {
			if s := frame.GetSeries(); s != nil {
				if builder != nil {
					b := builder.Block()
					bi.blocks <- b
				}

				builder = newBuilder()

				tags := make(Tags)
				for _, t := range s.Tags {
					tags[string(t.Key)] = string(t.Value)
				}
				builder.SetTags(tags)
			} else if p := frame.GetIntegerPoints(); p != nil {
				panic("ints not supported")
			} else if p := frame.GetFloatPoints(); p != nil {
				for i, c := range p.Timestamps {
					builder.AppendTime(0, Time(c))
					builder.AppendFloat(1, p.Values[i])
				}
			}
		}
	}
}
