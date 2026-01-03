package denoise

import (
	"fmt"
	"math"
)

/*
This function is responsible to build a normalized
1D Gaussian kernal.
The radius is chose as ceil(3*sigma) which captures 9
9% of gaussian mass.
*/
func GaussianKernel1D(sigma float32) ([]float32, int, error) {
	if sigma <= 0 {
		return nil, 0, fmt.Errorf("sigma must be > 0, got %v", sigma)
	}

	radius := int(math.Ceil(float64(3 * sigma)))
	if radius < 1 {
		radius = 1
	}

	size := 2*radius + 1
	k := make([]float32, size)

	for x := range size {
		X := float64(x - size/2)
		k[x] = float32(math.Exp(-(X * X))) / (2 * sigma * sigma)
	}

	gausSum := 0
	for g := range k {
		gausSum += g
	}
	for i, ki := range k {
		k[i] = ki / float32(gausSum)
	}

	return k, radius, nil
}
