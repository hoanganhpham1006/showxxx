package misc

import (
	"math"
	"math/rand"
)

func init() {
	_ = rand.Int
}

// Normal distribution:
//		https://www.mathsisfun.com/data/standard-normal-distribution.html
// 		https://en.wikipedia.org/wiki/Normal_distribution
//      This func return value almost certainly within 3*StandardDeviation of mean
// inputs:
//		mu: mean
//		sigma: standard deviation, so sigma**2 is variance
func RandNormDist(mu float64, sigma float64) float64 {
	return rand.NormFloat64()*sigma + mu
}

type MathFunction func(x float64) float64

// Normal distribution probability density function
func normalPDF(mu float64, sigma float64) MathFunction {
	return func(x float64) float64 {
		return math.Pow(math.E, -(x-mu)*(x-mu)/2/sigma/sigma) /
			sigma / math.Sqrt(2*math.Pi)
	}
}

func CalcMeanAndDeviation(a []float64) (float64, float64) {
	sum := float64(0)
	sumVariance := float64(0)
	for _, e := range a {
		sum += e
	}
	mean := sum / float64(len(a))
	for _, e := range a {
		sumVariance += (e - mean) * (e - mean)
	}
	variance := sumVariance / float64(len(a))
	return mean, math.Sqrt(variance)
}
