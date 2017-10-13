package functions_test

import (
	"math/rand"

	"github.com/gonum/stat/distuv"
)

var NormalData []float64

func init() {
	dist := distuv.Normal{
		Mu:     10,
		Sigma:  3,
		Source: rand.New(rand.NewSource(42)),
	}
	NormalData = make([]float64, 1e6)
	for i := range NormalData {
		NormalData[i] = dist.Rand()
	}
}
