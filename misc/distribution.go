package misc

import (
//	"math"
//	"math/rand"
)

// Normal distribution:
//		https://www.mathsisfun.com/data/standard-normal-distribution.html
// 		https://en.wikipedia.org/wiki/Normal_distribution
//      This func return value almost certainly within 3*StandardDeviation of mean
// inputs:
//		mu: mean
//		sigma: standard deviation, so sigma**2 is variance
func RandNormDist(mu float64, sigma float64) float64 {
	return 0
}

type DensityFunction func(x float64) float64
