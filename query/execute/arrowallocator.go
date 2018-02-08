package execute

import "github.com/influxdata/arrow/memory"

type arrowAllocator struct {
	*Resource
	alloc memory.Allocator
}

func (a *arrowAllocator) Allocate(size int) []byte {
	a.Resource.Reserve(size, 1)
	return a.alloc.Allocate(size)
}

func (a *arrowAllocator) Reallocate(size int, b []byte) []byte {
	diff := size - len(b)
	if diff < 0 {
		a.Resource.Release(-diff, 1)
	} else {
		a.Resource.Reserve(diff, 1)
	}
	return a.alloc.Reallocate(size, b)
}

func (a *arrowAllocator) Free(b []byte) {
	a.Resource.Release(len(b), 1)
	a.alloc.Free(b)
}
