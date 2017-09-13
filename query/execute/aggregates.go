package execute

import (
	"github.com/influxdata/ifql/query/plan"
)

type AggregateTransformation interface {
	Do(Block, BlockBuilder)
}

type sumAT struct {
	spec *plan.SumProcedureSpec
}

func (sumAT) Do(b Block, builder BlockBuilder) {
	values := b.Values()
	sum := 0.0
	for vs, ok := values.NextValues(); ok; vs, ok = values.NextValues() {
		for _, v := range vs {
			sum += v
		}
	}

	builder.SetTags(b.Tags())
	builder.SetBounds(b.Bounds())
	builder.AddCol(b.Bounds().Stop)
	builder.AddRow(b.Tags())
	builder.Set(0, 0, sum)
}

//type mergeProc struct {
//	spec *plan.MergeProcedureSpec
//}
//
//func (mergeProc) Do(src DataFrame) (FrameIterator, bool) {
//	dfBuilder := newDataFrameBuilder()
//	dfBuilder.SetBounds(src.Bounds())
//	blockBuilders := make(map[SeriesKey]BlockBuilder)
//
//	blocks := src.Blocks()
//	for b, ok := blocks.NextBlock(); ok; b, ok = blocks.NextBlock() {
//		cells := b.Cells()
//		for c, ok := cells.NextCell(); ok; c, ok = cells.NextCell() {
//			key := p.seriesKey(b.Tags())
//			builder := blockBuilders[key]
//			if builder != nil {
//				builder = newRowListBlockBuilder()
//				builder.SetBounds(b.Bounds())
//				builder.SetTags(b.Tags)
//
//				blockBuilders[key] = builder
//				dfBuilder.AddBlock(builder)
//			}
//
//			builder.AddCell(c)
//		}
//	}
//
//	return newFrameIterator(dfBuilder.DataFrame()), true
//}
//
//type SeriesKey string
//
//func (p mergeProc) seriesKey(tags Tags) SeriesKey {
//	var buf bytes.Buffer
//	for i, k := range p.Keys {
//		if i != 0 {
//			buf.WriteRune(',')
//		}
//		buf.WriteString(k)
//		buf.WriteRune('=')
//		buf.WriteString(tags[k])
//	}
//	return SeriesKey(buf.Bytes())
//}
//
//func (p mergeProc) Spec() plan.ProcedureSpec {
//	return p.spec
//}
