package execute

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type FormatOption func(*formatter)

func Formatted(b Block, opts ...FormatOption) fmt.Formatter {
	f := formatter{
		b: b,
	}
	for _, o := range opts {
		o(&f)
	}
	return f
}

func Head(m int) FormatOption {
	return func(f *formatter) { f.head = m }
}
func Squeeze() FormatOption {
	return func(f *formatter) { f.squeeze = true }
}

type formatter struct {
	b       Block
	head    int
	squeeze bool
}

func (f formatter) Format(fs fmt.State, c rune) {
	f.b = CacheOneTimeBlock(f.b)
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "%#v", f.b)
		return
	}
	// This is useful when debugging the formatter as fmt.Println will catch the panic and eat the trace.
	//defer func() {
	//	r := recover()
	//	if r != nil {
	//		panic(fmt.Sprintf("%v\n%s", r, debug.Stack()))
	//	}
	//}()
	f.format(fs, c)
}

func (f formatter) format(fs fmt.State, c rune) {
	tags := f.b.Tags()
	keys := tags.Keys()
	cols := f.b.Cols()
	fmt.Fprintf(fs, "Block: keys: %v bounds: %v\n", keys, f.b.Bounds())
	nCols := len(cols) + len(keys)

	// Determine number of rows to print
	nrows := math.MaxInt64
	if f.head > 0 {
		nrows = f.head
	}

	// Determine precision of floating point values
	prec, pOk := fs.Precision()
	if !pOk {
		prec = -1
	}

	var widths widther
	if f.squeeze {
		widths = make(columnWidth, nCols)
	} else {
		widths = new(uniformWidth)
	}

	fmtC := byte(c)
	if fmtC == 'v' {
		fmtC = 'g'
	}
	floatBuf := make([]byte, 0, 64)
	maxWidth := computeWidths(f.b, fmtC, nrows, prec, widths, floatBuf)

	width, _ := fs.Width()
	if width < maxWidth {
		width = maxWidth
	}
	if width < 2 {
		width = 2
	}
	pad := make([]byte, width)
	for i := range pad {
		pad[i] = ' '
	}
	dash := make([]byte, width)
	for i := range dash {
		dash[i] = '-'
	}
	eol := []byte{'\n'}

	ordered := newOrderedCols(cols)
	sort.Sort(ordered)

	// Print column headers
	for oj, c := range ordered.cols {
		j := ordered.Idx(oj)
		buf := []byte(c.Label)
		// Check justification
		if fs.Flag('-') {
			fs.Write(buf)
			fs.Write(pad[:widths.width(j)-len(buf)])
		} else {
			fs.Write(pad[:widths.width(j)-len(buf)])
			fs.Write(buf)
		}
		fs.Write(pad[:2])
	}
	fs.Write(eol)
	// Print header separator
	for oj := range ordered.cols {
		j := ordered.Idx(oj)
		fs.Write(dash[:widths.width(j)])
		fs.Write(pad[:2])
	}
	fs.Write(eol)

	n := nrows
	times := f.b.Times()
	times.DoTime(func(ts []Time, rr RowReader) {
		l := len(ts)
		if n < l {
			l = n
			n = 0
		} else {
			n -= l
		}
		for i := range ts[:l] {
			for oj, c := range ordered.cols {
				j := ordered.Idx(oj)
				var buf []byte
				switch c.Type {
				case TFloat:
					buf = strconv.AppendFloat(floatBuf, rr.AtFloat(i, j), fmtC, prec, 64)
				case TTime:
					buf = []byte(rr.AtTime(i, j).String())
				case TString:
					buf = []byte(rr.AtString(i, j))
				}
				// Check justification
				if fs.Flag('-') {
					fs.Write(buf)
					fs.Write(pad[:widths.width(j)-len(buf)])
				} else {
					fs.Write(pad[:widths.width(j)-len(buf)])
					fs.Write(buf)
				}
				fs.Write(pad[:2])
			}
			fs.Write(eol)
		}
	})
}

func computeWidths(b Block, fmtC byte, rows, prec int, widths widther, buf []byte) int {
	maxWidth := 0
	for j, c := range b.Cols() {
		n := rows
		values := b.Col(j)
		width := len(c.Label)
		switch c.Type {
		case TFloat:
			values.DoFloat(func(vs []float64, _ RowReader) {
				l := len(vs)
				if n < l {
					l = n
					n = 0
				} else {
					n -= l
				}
				for _, v := range vs[:l] {
					buf = strconv.AppendFloat(buf[0:0], v, fmtC, prec, 64)
					if w := len(buf); w > width {
						width = w
					}
				}
			})
		case TString:
			values.DoString(func(vs []string, _ RowReader) {
				l := len(vs)
				if n < l {
					l = n
					n = 0
				} else {
					n -= l
				}
				for _, v := range vs[:l] {
					if w := len(v); w > width {
						width = w
					}
				}
			})
		case TTime:
			width = len(fixedWidthTimeFmt)
		}
		widths.setWidth(j, width)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

type widther interface {
	width(i int) int
	setWidth(i, w int)
}

type uniformWidth int

func (u *uniformWidth) width(_ int) int { return int(*u) }
func (u *uniformWidth) setWidth(_, w int) {
	if uniformWidth(w) > *u {
		*u = uniformWidth(w)
	}
}

type columnWidth []int

func (c columnWidth) width(i int) int   { return c[i] }
func (c columnWidth) setWidth(i, w int) { c[i] = w }

// orderedCols sorts a list of columns such that:
//
// * time
// * common tags sorted by label
// * other tags sorted by label
// * value
type orderedCols struct {
	indexMap []int
	cols     []ColMeta
}

func newOrderedCols(cols []ColMeta) orderedCols {
	indexMap := make([]int, len(cols))
	for i := range indexMap {
		indexMap[i] = i
	}
	cpy := make([]ColMeta, len(cols))
	copy(cpy, cols)
	return orderedCols{
		indexMap: indexMap,
		cols:     cpy,
	}
}

func (o orderedCols) Idx(oj int) int {
	return o.indexMap[oj]
}

func (o orderedCols) Len() int { return len(o.cols) }
func (o orderedCols) Swap(i int, j int) {
	o.cols[i], o.cols[j] = o.cols[j], o.cols[i]
	o.indexMap[i], o.indexMap[j] = o.indexMap[j], o.indexMap[i]
}

func (o orderedCols) Less(i int, j int) bool {
	// Time column is always first
	if o.cols[i].Label == timeColLabel {
		return true
	}
	if o.cols[j].Label == timeColLabel {
		return false
	}

	// Value column is always last
	if o.cols[i].Label == valueColLabel {
		return false
	}
	if o.cols[j].Label == valueColLabel {
		return true
	}

	// Common tags before other tags
	if o.cols[i].IsTag && o.cols[i].IsCommon && o.cols[j].IsTag && !o.cols[j].IsCommon {
		return true
	}
	if o.cols[i].IsTag && !o.cols[i].IsCommon && o.cols[j].IsTag && o.cols[j].IsCommon {
		return false
	}

	// within a class sort be label
	return o.cols[i].Label < o.cols[j].Label
}
