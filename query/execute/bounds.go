package execute

import "fmt"

type Bounds struct {
	Start Time
	Stop  Time
}

func (b Bounds) String() string {
	return fmt.Sprintf("[%v, %v]", b.Start, b.Stop)
}

func (b Bounds) Contains(t Time) bool {
	return t >= b.Start && t < b.Stop
}

func (b Bounds) Overlaps(o Bounds) bool {
	return b.Contains(o.Start) || b.Contains(o.Stop)
}

func (b Bounds) Equal(o Bounds) bool {
	return b == o
}
