package execute

//go:generate go run ../../_tools/tmpl/main.go -i -data=collistallocator.tmpldata collistallocator.gen.go.tmpl

const (
	boolSize    = 1
	int64Size   = 8
	uint64Size  = 8
	float64Size = 8
	stringSize  = 16
	timeSize    = 8
)

// ColListAllocator tracks the amount of memory being consumed by a query.
// The allocator provides methods similar to make and append, to allocate large slices of data.
// The allocator also provides a Free method to account for when memory will be freed.
type ColListAllocator struct {
	alloc Allocator
}

func NewColListAllocator(a Allocator) *ColListAllocator {
	return &ColListAllocator{a}
}

// free informs the allocator that memory has been freed.
func (a *ColListAllocator) free(b []byte) {
	a.alloc.Free(b)
}

func (a *ColListAllocator) reallocate(size int, b []byte) []byte {
	if b == nil {
		return a.alloc.Allocate(size)
	}
	return a.alloc.Reallocate(size, b)
}
