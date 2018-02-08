package execute

import (
	"fmt"
	"sync/atomic"
)

type ResourceError struct {
	Limit     int64
	Allocated int64
	Wanted    int64
}

func (a ResourceError) Error() string {
	return fmt.Sprintf("reservation limit reached: limit %d, allocated: %d, wanted: %d", a.Limit, a.Allocated, a.Wanted)
}

type Resource struct {
	Limit       int64
	reserved    int64
	maxReserved int64
}

// Reserve attempts to decrease the resource by n × size units.
//
// Reserve will panic with a `ResourceError` if there are insufficient remaining resources.
func (r *Resource) Reserve(n, size int) {
	if want := r.count(n, size); want > r.Limit {
		allocated := r.count(-n, size)
		panic(ResourceError{
			Limit:     r.Limit,
			Allocated: allocated,
			Wanted:    want - allocated,
		})
	}
}

// Release will return n × size units, making them available for future Reserve calls.
func (r *Resource) Release(n, size int) {
	r.count(-n, size)
}

// Max reports the maximum number of units reserved.
func (r *Resource) Max() int64 { return atomic.LoadInt64(&r.maxReserved) }

func (r *Resource) count(n, size int) (c int64) {
	c = atomic.AddInt64(&r.reserved, int64(n*size))
	for max := atomic.LoadInt64(&r.maxReserved); c > max; max = atomic.LoadInt64(&r.maxReserved) {
		if atomic.CompareAndSwapInt64(&r.maxReserved, max, c) {
			return
		}
	}
	return
}
