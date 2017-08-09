package execute

import (
	"context"
	"io"
	"log"

	"github.com/gogo/protobuf/codec"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/yarpc"
)

type StorageReader interface {
	Read(database string, start, stop Time) (DataFrame, bool)
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

func (sr *storageReader) Read(database string, start, stop Time) (DataFrame, bool) {

	var req storage.ReadRequest
	req.Database = database
	req.TimestampRange.Start = int64(start)
	req.TimestampRange.End = int64(stop)
	stream, err := sr.c.Read(context.Background(), &req)
	if err != nil {
		log.Println("E!", err)
		return nil, false
	}

	df := new(dataframe)
	// TODO handle sparse times
	df.stride = -1
	df.bounds = bounds{start: start, stop: stop}

	for {
		// Recv the next response
		var rep storage.ReadResponse
		if err := stream.RecvMsg(&rep); err != nil {
			if err == io.EOF {
				return df, true
			}
			//TODO add proper error handling
			log.Println("E!", err)
			return nil, false
		}

		for _, frame := range rep.Frames {
			if s := frame.GetSeries(); s != nil {
				//TODO get actual tag values back from storage response
				df.rows = append(df.rows, Tags{"__name__": s.Name})
			} else if p := frame.GetIntegerPoints(); p != nil {
				panic("ints not supported")
			} else if p := frame.GetFloatPoints(); p != nil {
				if df.stride == -1 {
					df.stride = len(p.Timestamps)
					df.cols = make([]Time, df.stride)
					for i, c := range p.Timestamps {
						df.cols[i] = Time(c)
					}
					df.bounds = bounds{start: df.cols[0], stop: df.cols[len(df.cols)-1]}
				}
				if len(p.Values) != df.stride {
					panic("non dense data found")
				}
				df.data = append(df.data, p.Values...)
			}
		}
	}
}

func (sr *storageReader) Close() {
	sr.conn.Close()
}
