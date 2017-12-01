package execute

import (
	"fmt"
	"sync/atomic"
)

const (
	boolSize    = 1
	int64Size   = 8
	uint64Size  = 8
	float64Size = 8
	stringSize  = 16
	timeSize    = 8
)

type Allocator struct {
	Limit          int64
	bytesAllocated int64
	maxAllocated   int64
}

func (a *Allocator) count(n, size int) (c int64) {
	c = atomic.AddInt64(&a.bytesAllocated, int64(n*size))
	for max := atomic.LoadInt64(&a.maxAllocated); c > max; max = atomic.LoadInt64(&a.maxAllocated) {
		if atomic.CompareAndSwapInt64(&a.maxAllocated, max, c) {
			return
		}
	}
	return
}
func (a *Allocator) Free(n, size int) {
	a.count(-n, size)
}

func (a *Allocator) Max() int64 {
	return atomic.LoadInt64(&a.maxAllocated)
}

func (a *Allocator) account(n, size int) {
	if want := a.count(n, size); want > a.Limit {
		allocated := a.count(-n, size)
		panic(AllocError{
			Limit:     a.Limit,
			Allocated: allocated,
			Wanted:    want - allocated,
		})
	}
}
func (a *Allocator) Bools(l, c int) []bool {
	a.account(c, boolSize)
	return make([]bool, l, c)
}

func (a *Allocator) AppendBools(slice []bool, vs ...bool) []bool {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, boolSize)
	return s
}

func (a *Allocator) Ints(l, c int) []int64 {
	a.account(c, int64Size)
	return make([]int64, l, c)
}

func (a *Allocator) AppendInts(slice []int64, vs ...int64) []int64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, int64Size)
	return s
}
func (a *Allocator) UInts(l, c int) []uint64 {
	a.account(c, uint64Size)
	return make([]uint64, l, c)
}

func (a *Allocator) AppendUInts(slice []uint64, vs ...uint64) []uint64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, uint64Size)
	return s
}

func (a *Allocator) Floats(l, c int) []float64 {
	a.account(c, float64Size)
	return make([]float64, l, c)
}

func (a *Allocator) AppendFloats(slice []float64, vs ...float64) []float64 {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, float64Size)
	return s
}

func (a *Allocator) Strings(l, c int) []string {
	a.account(c, stringSize)
	return make([]string, l, c)
}

func (a *Allocator) AppendStrings(slice []string, vs ...string) []string {
	//TODO(nathanielc): Account for actual size of strings
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, stringSize)
	return s
}

func (a *Allocator) Times(l, c int) []Time {
	a.account(c, timeSize)
	return make([]Time, l, c)
}

func (a *Allocator) AppendTimes(slice []Time, vs ...Time) []Time {
	if cap(slice)-len(slice) > len(vs) {
		return append(slice, vs...)
	}
	s := append(slice, vs...)
	diff := cap(s) - cap(slice)
	a.account(diff, timeSize)
	return s
}

type AllocError struct {
	Limit     int64
	Allocated int64
	Wanted    int64
}

func (a AllocError) Error() string {
	return fmt.Sprintf("allocation limit reached: limit %d, allocated: %d, wanted: %d", a.Limit, a.Allocated, a.Wanted)
}
