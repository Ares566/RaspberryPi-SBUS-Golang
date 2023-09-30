package utils

import "golang.org/x/exp/constraints"

// Abs returns the absolute value of x.
func Abs[T constraints.Integer](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func Mapping(x float64, inMin float64, inMax float64, outMin float64, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
