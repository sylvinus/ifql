package execute

import (
	"context"
	"io"

	"github.com/gogo/protobuf/codec"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/yarpc"
)

type StorageReader interface {
	Read(database string, predicate *storage.Predicate, limit int64, desc bool, start, stop Time) (BlockIterator, error)
	Close()
}

func NewStorageReader() (StorageReader, error) {
	opts := []yarpc.DialOption{yarpc.WithCodec(codec.New(1000))}
	conn, err := yarpc.Dial("localhost:8082", opts...)
	if err != nil {
		return nil, err
	}

	c := storage.NewStorageClient(conn)
	return &storageReader{
		conn: conn,
		c:    c,
	}, nil
}

type storageReader struct {
	conn *yarpc.ClientConn
	c    storage.StorageClient
}

func (sr *storageReader) Read(database string, predicate *storage.Predicate, limit int64, desc bool, start, stop Time) (BlockIterator, error) {
	var req storage.ReadRequest
	req.Database = database
	req.Predicate = predicate
	//	req.Limit = limit
	req.Descending = desc
	req.TimestampRange.Start = int64(start)
	req.TimestampRange.End = int64(stop)

	stream, err := sr.c.Read(context.Background(), &req)
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
	sr.conn.Close()
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

		builder := newRowListBlockBuilder()
		builder.SetBounds(bi.bounds)

		for _, frame := range rep.Frames {
			if s := frame.GetSeries(); s != nil {
				tags := make(Tags)
				for _, t := range s.Tags {
					tags[string(t.Key)] = string(t.Value)
				}
				builder.AddRow(tags)
				builder.SetTags(tags)
			} else if p := frame.GetIntegerPoints(); p != nil {
				panic("ints not supported")
			} else if p := frame.GetFloatPoints(); p != nil {
				for _, c := range p.Timestamps {
					builder.AddCol(Time(c))
				}
				for j, v := range p.Values {
					builder.Set(0, j, v)
				}

				// Each row is its own block
				bi.blocks <- builder.Block()
				builder = newRowListBlockBuilder()
				builder.SetBounds(bi.bounds)
			}
		}
	}
}
