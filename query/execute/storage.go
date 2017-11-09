package execute

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/yarpc"
	"github.com/pkg/errors"
)

type StorageReader interface {
	Read(rs ReadSpec, start, stop Time) (BlockIterator, error)
	Close()
}

type ReadSpec struct {
	RAMLimit   uint64
	Database   string
	Predicate  expression.Expression
	Limit      int64
	Descending bool

	AggregateType string

	// OrderByTime indicates that series reads should produce all
	// series for a time before producing any series for a larger time.
	// By default this is false meaning all values of time are produced for a given series,
	// before any values are produced from the next series.
	OrderByTime bool
	// MergeAll indicates that all series should be merged into a single group
	MergeAll bool
	// GroupKeys is the list of dimensions along which to group
	GroupKeys []string
	// GroupIgnore is the list of dimensions along which to not group
	GroupIgnore []string
	// GroupKeep is the list of tags to keep but not group by.
	GroupKeep []string
}

func NewStorageReader(hosts []string) (StorageReader, error) {
	if len(hosts) == 0 {
		return nil, errors.New("must provide at least one storage host")
	}
	conn, err := connect(hosts[0])
	if err != nil {
		return nil, err
	}
	return &storageReader{
		conn: conn,
		c:    storage.NewStorageClient(conn),
	}, nil
}

type storageReader struct {
	conn *yarpc.ClientConn
	c    storage.StorageClient
}

func connect(addr string) (*yarpc.ClientConn, error) {
	return yarpc.Dial(addr)
}

func determineAggregateType(agg string) (storage.Aggregate_AggregateType, error) {
	if agg == "" {
		return storage.AggregateTypeNone, nil
	}

	if t, ok := storage.Aggregate_AggregateType_value[strings.ToUpper(agg)]; ok {
		return storage.Aggregate_AggregateType(t), nil
	}
	return 0, fmt.Errorf("unknown aggregate type %q", agg)
}

func (sr *storageReader) Read(readSpec ReadSpec, start, stop Time) (BlockIterator, error) {
	bi := &storageBlockIterator{
		bounds: Bounds{
			Start: start,
			Stop:  stop,
		},
		readSpec: readSpec,
		c:        sr.c,
	}
	return bi, nil
}

func (sr *storageReader) Close() {
	sr.conn.Close()
}

type storageBlockIterator struct {
	bounds Bounds
	data   *readState

	readSpec ReadSpec
	c        storage.StorageClient
}

func (bi *storageBlockIterator) Do(f func(Block)) error {
	// Setup read request
	var req storage.ReadRequest
	if bi.readSpec.Predicate.Root != nil {
		predicate, err := ExpressionToStoragePredicate(bi.readSpec.Predicate.Root)
		if err != nil {
			return err
		}
		req.Predicate = predicate
	}

	req.Database = bi.readSpec.Database
	req.Descending = bi.readSpec.Descending
	req.TimestampRange.Start = int64(bi.bounds.Start)
	req.TimestampRange.End = int64(bi.bounds.Stop)
	req.Grouping = bi.readSpec.GroupKeys
	if agg, err := determineAggregateType(bi.readSpec.AggregateType); err != nil {
		return err
	} else if agg != storage.AggregateTypeNone {
		req.Aggregate = &storage.Aggregate{Type: agg}
	}

	stream, err := bi.c.Read(context.Background(), &req)
	if err != nil {
		return err
	}
	data := &readState{
		stream: stream,
	}

	for data.more() {
		if p := data.peek(); frameType(p) != seriesType {
			//This means the consumer didn't read all the data off the block
			return errors.New("internal error: short read")
		}
		frame := data.next()
		s := frame.GetSeries()
		typ := convertDataType(s.DataType)
		tags, keptTags := bi.determineBlockTags(s)
		block := newStorageBlock(bi.bounds, tags, keptTags, data, &bi.readSpec, typ)
		f(block)
		// Wait until the block has been read.
		block.wait()
	}
	return nil
}

func convertDataType(t storage.ReadResponse_DataType) DataType {
	switch t {
	case storage.DataTypeFloat:
		return TFloat
	case storage.DataTypeInteger:
		return TInt
	case storage.DataTypeUnsigned:
		return TUInt
	case storage.DataTypeBoolean:
		return TBool
	case storage.DataTypeString:
		return TString
	default:
		return TInvalid
	}
}

func (bi *storageBlockIterator) determineBlockTags(s *storage.ReadResponse_SeriesFrame) (tags, keptTags Tags) {
	if len(bi.readSpec.GroupKeys) > 0 {
		tags = make(Tags, len(bi.readSpec.GroupKeys))
		for _, key := range bi.readSpec.GroupKeys {
			for _, tag := range s.Tags {
				if string(tag.Key) == key {
					tags[key] = string(tag.Value)
					break
				}
			}
		}
		if len(bi.readSpec.GroupKeep) > 0 {
			keptTags = make(Tags, len(bi.readSpec.GroupKeep))
			for _, key := range bi.readSpec.GroupKeep {
				for _, tag := range s.Tags {
					if string(tag.Key) == key {
						keptTags[key] = string(tag.Value)
						break
					}
				}
			}
		}
	} else if len(bi.readSpec.GroupIgnore) > 0 {
		tags = make(Tags, len(s.Tags)-len(bi.readSpec.GroupIgnore))
		keptTags = make(Tags, len(bi.readSpec.GroupKeep))
	TAGS:
		for _, t := range s.Tags {
			k := string(t.Key)
			for _, key := range bi.readSpec.GroupKeep {
				if k == key {
					keptTags[key] = string(t.Value)
				}
			}
			for _, key := range bi.readSpec.GroupIgnore {
				if k == key {
					continue TAGS
				}
			}
			tags[k] = string(t.Value)
		}
	} else if !bi.readSpec.MergeAll {
		tags = make(Tags, len(s.Tags))
		for _, t := range s.Tags {
			tags[string(t.Key)] = string(t.Value)
		}
	} else {
		keptTags = make(Tags, len(bi.readSpec.GroupKeep))
		for _, t := range s.Tags {
			k := string(t.Key)
			for _, key := range bi.readSpec.GroupKeep {
				if k == key {
					keptTags[key] = string(t.Value)
				}
			}
		}
	}
	return
}

type storageBlock struct {
	bounds   Bounds
	tags     Tags
	keptTags Tags
	colMeta  []ColMeta

	readSpec *ReadSpec

	done chan struct{}

	data *readState
}

func newStorageBlock(bounds Bounds, tags, keptTags Tags, data *readState, readSpec *ReadSpec, typ DataType) *storageBlock {
	colMeta := make([]ColMeta, 2, 2+len(tags)+len(keptTags))
	colMeta[0] = TimeCol
	colMeta[1] = ColMeta{
		Label: ValueColLabel,
		Type:  typ,
	}

	for _, k := range tags.Keys() {
		colMeta = append(colMeta, ColMeta{
			Label:    k,
			Type:     TString,
			IsTag:    true,
			IsCommon: true,
		})
	}
	for _, k := range keptTags.Keys() {
		colMeta = append(colMeta, ColMeta{
			Label:    k,
			Type:     TString,
			IsTag:    true,
			IsCommon: false,
		})
	}
	return &storageBlock{
		bounds:   bounds,
		tags:     tags,
		keptTags: keptTags,
		colMeta:  colMeta,
		readSpec: readSpec,
		data:     data,
		done:     make(chan struct{}),
	}
}

func (b *storageBlock) wait() {
	<-b.done
}

// onetime satisfies the OneTimeBlock interface since this block may only be read once.
func (b *storageBlock) onetime() {}

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
		tags:     b.tags,
		data:     b.data,
		colMeta:  b.colMeta,
		col:      c,
		done:     b.done,
		readSpec: b.readSpec,
		keptTags: b.keptTags,
	}
}

func (b *storageBlock) Times() ValueIterator {
	return b.Col(0)
}
func (b *storageBlock) Values() ValueIterator {
	return b.Col(1)
}

type storageBlockValueIterator struct {
	tags Tags
	data *readState
	done chan<- struct{}

	readSpec *ReadSpec

	// colMeta always has at least two columns, where the first is a TimeCol
	// and the second is any Value column.
	colMeta []ColMeta
	col     int

	// colBufs are the buffers for the given columns.
	colBufs [2]interface{}
	// keptTags is a set of non common tags.
	keptTags Tags

	// resuable buffer for the time column
	timeBuf []Time

	// resuable buffers for the different types of values
	boolBuf   []bool
	intBuf    []int64
	uintBuf   []uint64
	floatBuf  []float64
	stringBuf []string
}

func (b *storageBlockValueIterator) Cols() []ColMeta {
	return b.colMeta
}
func (b *storageBlockValueIterator) DoBool(f func([]bool, RowReader)) {
	checkColType(b.colMeta[b.col], TBool)
	for b.advance() {
		f(b.colBufs[b.col].([]bool), b)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoInt(f func([]int64, RowReader)) {
	checkColType(b.colMeta[b.col], TInt)
	for b.advance() {
		f(b.colBufs[b.col].([]int64), b)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoUInt(f func([]uint64, RowReader)) {
	checkColType(b.colMeta[b.col], TUInt)
	for b.advance() {
		f(b.colBufs[b.col].([]uint64), b)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(b.colMeta[b.col], TFloat)
	for b.advance() {
		f(b.colBufs[b.col].([]float64), b)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoString(f func([]string, RowReader)) {
	defer close(b.done)

	meta := b.colMeta[b.col]
	checkColType(meta, TString)
	if meta.IsTag {
		// Handle creating a strs slice that can be ranged according to actual data received.
		var strs []string
		var value string
		if meta.IsCommon {
			value = b.tags[meta.Label]
		} else {
			value = b.keptTags[meta.Label]
		}
		for b.advance() {
			l := len(b.timeBuf)
			if cap(strs) < l {
				strs = make([]string, l)
				for i := range strs {
					strs[i] = value
				}
			} else if len(strs) < l {
				new := strs[len(strs)-1 : l]
				for i := range new {
					new[i] = value
				}
				strs = strs[0:l]
			} else {
				strs = strs[0:l]
			}
			f(strs, b)
		}
		return
	}
	// Do ordinary range over column data.
	for b.advance() {
		f(b.colBufs[b.col].([]string), b)
	}
}

func (b *storageBlockValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(b.colMeta[b.col], TTime)
	for b.advance() {
		f(b.colBufs[b.col].([]Time), b)
	}
	close(b.done)
}

func (b *storageBlockValueIterator) AtBool(i, j int) bool {
	checkColType(b.colMeta[j], TBool)
	return b.colBufs[j].([]bool)[i]
}
func (b *storageBlockValueIterator) AtInt(i, j int) int64 {
	checkColType(b.colMeta[j], TInt)
	return b.colBufs[j].([]int64)[i]
}
func (b *storageBlockValueIterator) AtUInt(i, j int) uint64 {
	checkColType(b.colMeta[j], TUInt)
	return b.colBufs[j].([]uint64)[i]
}
func (b *storageBlockValueIterator) AtFloat(i, j int) float64 {
	checkColType(b.colMeta[j], TFloat)
	return b.colBufs[j].([]float64)[i]
}
func (b *storageBlockValueIterator) AtString(i, j int) string {
	meta := b.colMeta[j]
	checkColType(meta, TString)
	if meta.IsTag {
		if meta.IsCommon {
			return b.tags[meta.Label]
		}
		return b.keptTags[meta.Label]
	}
	return b.colBufs[j].([]string)[i]
}
func (b *storageBlockValueIterator) AtTime(i, j int) Time {
	checkColType(b.colMeta[j], TTime)
	return b.colBufs[j].([]Time)[i]
}

func (b *storageBlockValueIterator) advance() bool {
	for b.data.more() {
		//reset buffers
		b.timeBuf = b.timeBuf[0:0]
		b.boolBuf = b.boolBuf[0:0]
		b.intBuf = b.intBuf[0:0]
		b.uintBuf = b.uintBuf[0:0]
		b.stringBuf = b.stringBuf[0:0]
		b.floatBuf = b.floatBuf[0:0]

		switch p := b.data.peek(); frameType(p) {
		case seriesType:
			s := p.GetSeries()
			for _, t := range s.Tags {
				key := string(t.Key)
				if v, ok := b.tags[key]; ok && v != string(t.Value) {
					return false
				}
			}
			// Populate keptTags with new series values
			b.keptTags = make(Tags, len(b.readSpec.GroupKeep))
			for _, t := range s.Tags {
				k := string(t.Key)
				for _, key := range b.readSpec.GroupKeep {
					if k == key {
						b.keptTags[key] = string(t.Value)
					}
				}
			}
			// Advance to next frame
			b.data.next()
		case boolPointsType:
			if b.colMeta[1].Type != TBool {
				// TODO: Add error handling
				// Type changed,
				return false
			}
			// read next frame
			frame := b.data.next()
			p := frame.GetBooleanPoints()
			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.boolBuf) {
				b.boolBuf = make([]bool, l)
			} else {
				b.boolBuf = b.boolBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.boolBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.boolBuf
			return true
		case intPointsType:
			if b.colMeta[1].Type != TInt {
				// TODO: Add error handling
				// Type changed,
				return false
			}
			// read next frame
			frame := b.data.next()
			p := frame.GetIntegerPoints()
			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.uintBuf) {
				b.intBuf = make([]int64, l)
			} else {
				b.intBuf = b.intBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.intBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.intBuf
			return true
		case uintPointsType:
			if b.colMeta[1].Type != TUInt {
				// TODO: Add error handling
				// Type changed,
				return false
			}
			// read next frame
			frame := b.data.next()
			p := frame.GetUnsignedPoints()
			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.intBuf) {
				b.uintBuf = make([]uint64, l)
			} else {
				b.uintBuf = b.uintBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.uintBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.uintBuf
			return true
		case floatPointsType:
			if b.colMeta[1].Type != TFloat {
				// TODO: Add error handling
				// Type changed,
				return false
			}
			// read next frame
			frame := b.data.next()
			p := frame.GetFloatPoints()

			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.floatBuf) {
				b.floatBuf = make([]float64, l)
			} else {
				b.floatBuf = b.floatBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.floatBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.floatBuf
			return true
		case stringPointsType:
			if b.colMeta[1].Type != TString {
				// TODO: Add error handling
				// Type changed,
				return false
			}
			// read next frame
			frame := b.data.next()
			p := frame.GetStringPoints()

			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.stringBuf) {
				b.stringBuf = make([]string, l)
			} else {
				b.stringBuf = b.stringBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.stringBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.stringBuf
			return true
		}
	}
	return false
}

type readState struct {
	stream storage.Storage_ReadClient
	rep    storage.ReadResponse
}

type responseType int

const (
	seriesType responseType = iota
	boolPointsType
	intPointsType
	uintPointsType
	floatPointsType
	stringPointsType
)

func (s *readState) peek() storage.ReadResponse_Frame {
	return s.rep.Frames[0]
}

func frameType(frame storage.ReadResponse_Frame) responseType {
	switch frame.Data.(type) {
	case *storage.ReadResponse_Frame_Series:
		return seriesType
	case *storage.ReadResponse_Frame_BooleanPoints:
		return boolPointsType
	case *storage.ReadResponse_Frame_IntegerPoints:
		return intPointsType
	case *storage.ReadResponse_Frame_UnsignedPoints:
		return uintPointsType
	case *storage.ReadResponse_Frame_FloatPoints:
		return floatPointsType
	case *storage.ReadResponse_Frame_StringPoints:
		return stringPointsType
	default:
		panic(fmt.Errorf("unknown read response frame type: %T", frame.Data))
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
