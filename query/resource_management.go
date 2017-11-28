package query

import (
	"math"
	"strconv"

	"github.com/pkg/errors"
)

// ResourceManagement defines how the query should consume avaliable resources.
type ResourceManagement struct {
	// Priority or the query.
	// Queries with a lower value will move to the front of the priority queue.
	// A zero value indicates the highest priority.
	Priority Priority `json:"priority"`
	// MaxCPUCores is the maximum number of cores this query may consume.
	// A zero value indicates unlimited.
	MaxCPUCores int64 `json:"max_cpu_cores"`
	// MaxRAMBytes is the maximum number of bytes of RAM this query may consume.
	// There is a small amount of overhead RAM being consumed by a query that will not be counted towards this limit.
	// A zero value indicates unlimited.
	MaxRAMBytes int64 `json:"max_ram_bytes"`
}

// Priority is an integer that represents the query priority.
// Any positive 32bit integer value may be used.
// Special constants are provided to represent the extreme high and low priorities.
type Priority int32

const (
	// High is the highest possible priority = 0
	High Priority = 0
	// Low is the lowest possible priority = MaxInt32
	Low Priority = math.MaxInt32
)

func (p Priority) MarshalText() ([]byte, error) {
	switch p {
	case Low:
		return []byte("low"), nil
	case High:
		return []byte("high"), nil
	default:
		return []byte(strconv.FormatInt(int64(p), 10)), nil
	}
}

func (p *Priority) UnmarshalText(txt []byte) error {
	switch s := string(txt); s {
	case "low":
		*p = Low
	case "high":
		*p = High
	default:
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid priority, must be an integer or 'low','high'")
		}
		*p = Priority(i)
	}
	return nil
}
