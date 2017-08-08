package execute

// DataFrame represents a bounded set of data.
type DataFrame interface {
	Bounds() Bounds

	NRows() int
	NCols() int

	//Rows() RowIterator

	ColsIndex() []Time

	//Row(i int) Iterator
	RowSlice(i int) ([]float64, Tags)

	//Col(j int) Iterator
}

type writeDataFrame interface {
	DataFrame
	Set(i, j int, value float64)
	SetRowTags(i int, tags Tags)
	SetColTime(j int, time Time)
}

type RowIterator interface {
	Next() (Iterator, Tags)
}

type ColIterator interface {
	Next() Iterator
}

type Iterator interface {
	Next() float64
}

type Tags map[string]string
type Time int64

type Bounds interface {
	Start() Time
	Stop() Time
}

type bounds struct {
	start Time
	stop  Time
}

func (b bounds) Start() Time {
	return b.start
}
func (b bounds) Stop() Time {
	return b.stop
}

func NewWriteDataFrame(r, c int, bounds Bounds) writeDataFrame {
	return &dataframe{
		bounds: bounds,
		data:   make([]float64, r*c),
		stride: c,
		rows:   make([]Tags, r),
		cols:   make([]Time, c),
	}
}

type dataframe struct {
	bounds Bounds
	data   []float64
	stride int
	rows   []Tags
	cols   []Time
}

func (d *dataframe) Bounds() Bounds {
	return d.bounds
}

func (d *dataframe) NRows() int {
	return len(d.rows)
}

func (d *dataframe) NCols() int {
	return len(d.cols)
}

func (d *dataframe) ColsIndex() []Time {
	return d.cols
}

func (d *dataframe) RowSlice(i int) ([]float64, Tags) {
	return d.data[i*d.stride : (i+1)*d.stride], d.rows[i]
}
func (d *dataframe) Set(i, j int, value float64) {
	d.data[i*d.stride+j] = value
}
func (d *dataframe) SetRowTags(i int, tags Tags) {
	d.rows[i] = tags
}
func (d *dataframe) SetColTime(j int, time Time) {
	d.cols[j] = time
}
