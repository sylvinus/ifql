package execute

const (
	boolSize    = 1
	int64Size   = 8
	uint64Size  = 8
	float64Size = 8
	stringSize  = 16
	timeSize    = 8
)

// Allocator tracks the amount of memory being consumed by a query.
// The allocator provides methods similar to make and append, to allocate large slices of data.
// The allocator also provides a Free method to account for when memory will be freed.
type Allocator struct {
	*Resource
}

// Free informs the allocator that memory has been freed.
func (a *Allocator) Free(n, size int) { a.Resource.Release(n, size) }

// Max reports the maximum amount of allocated memory at any point in the query.
func (a *Allocator) Max() int64 { return a.Resource.Max() }

// Bools makes a slice of bool values.
func (a *Allocator) Bools(l, c int) []bool {
	a.Resource.Reserve(c, boolSize)
	return make([]bool, l, c)
}

// AppendBools appends bools to a slice
func (a *Allocator) AppendBools(slice []bool, vs ...bool) []bool {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, boolSize)
	return s
}

// Ints makes a slice of int64 values.
func (a *Allocator) Ints(l, c int) []int64 {
	a.Resource.Reserve(c, int64Size)
	return make([]int64, l, c)
}

// AppendInts appends int64s to a slice
func (a *Allocator) AppendInts(slice []int64, vs ...int64) []int64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, int64Size)
	return s
}

// UInts makes a slice of uint64 values.
func (a *Allocator) UInts(l, c int) []uint64 {
	a.Resource.Reserve(c, uint64Size)
	return make([]uint64, l, c)
}

// AppendUInts appends uint64s to a slice
func (a *Allocator) AppendUInts(slice []uint64, vs ...uint64) []uint64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, uint64Size)
	return s
}

// Floats makes a slice of float64 values.
func (a *Allocator) Floats(l, c int) []float64 {
	a.Resource.Reserve(c, float64Size)
	return make([]float64, l, c)
}

// AppendFloats appends float64s to a slice
func (a *Allocator) AppendFloats(slice []float64, vs ...float64) []float64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, float64Size)
	return s
}

// Strings makes a slice of string values.
// Only the string headers are accounted for.
func (a *Allocator) Strings(l, c int) []string {
	a.Resource.Reserve(c, stringSize)
	return make([]string, l, c)
}

// AppendStrings appends strings to a slice.
// Only the string headers are accounted for.
func (a *Allocator) AppendStrings(slice []string, vs ...string) []string {
	//TODO(nathanielc): Account for actual size of strings
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, stringSize)
	return s
}

// Times makes a slice of Time values.
func (a *Allocator) Times(l, c int) []Time {
	a.Resource.Reserve(c, timeSize)
	return make([]Time, l, c)
}

// AppendTimes appends Times to a slice
func (a *Allocator) AppendTimes(slice []Time, vs ...Time) []Time {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.Resource.Reserve(diff, timeSize)
	return s
}
