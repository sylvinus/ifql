package execute

import "github.com/influxdata/arrow/memory"

type Allocator = memory.Allocator

type LimitedAllocator struct {
	alloc Allocator
	res   *Resource
}

func NewLimitedAllocator(a Allocator, r *Resource) *LimitedAllocator {
	return &LimitedAllocator{alloc: a, res: r}
}

func (a *LimitedAllocator) Allocate(size int) []byte {
	a.res.Reserve(size, 1)
	return a.alloc.Allocate(size)
}

func (a *LimitedAllocator) Reallocate(size int, b []byte) []byte {
	diff := size - len(b)
	a.res.Reserve(diff, 1)
	return a.alloc.Reallocate(size, b)
}

func (a *LimitedAllocator) Free(b []byte) {
	a.res.Release(cap(b), 1)
	a.alloc.Free(b)
}
