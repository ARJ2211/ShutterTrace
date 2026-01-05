package denoise

import (
	"fmt"
	"math"
)

/*
This function is responsible to build a normalized
1D Gaussian kernal.
The radius is chose as ceil(3*sigma) which captures
99% of gaussian mass.
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

	denom := 2 * float64(sigma) * float64(sigma)

	var sum float32
	for i := 0; i < size; i++ {
		x := float64(i - radius)
		v := math.Exp(-(x * x) / denom)
		k[i] = float32(v)
		sum += k[i]
	}

	for i := range k {
		k[i] /= sum
	}

	return k, radius, nil
}

/*
This function is responsible for clamping the values of
v between lo and hi
*/
func clampInt(v int, lo int, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

/*
This function applies a Gaussian blir to a grayscale image that
is stored as a flat slice. Src length must we w*h and the output
is a slice of the same length. Border handling uses clamping
*/
func GaussianBlurGray(
	src []float32,
	w, h int,
	sigma float32) ([]float32, error) {

	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid dims: %dx%d", w, h)
	}
	if len(src) != w*h {
		return nil, fmt.Errorf("src length mismatch: got %d want %d", len(src), w*h)
	}

	k, radius, err := GaussianKernel1D(sigma)
	if err != nil {
		return nil, err
	}

	// make a horizontal blur
	tmp := make([]float32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum float32
			for dx := -radius; dx <= radius; dx++ {
				nx := clampInt(x+dx, 0, w-1)
				sum += src[y*w+nx] * k[dx+radius]
			}
			tmp[y*w+x] = sum
		}
	}

	// make the vertical blur
	out := make([]float32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum float32
			for dy := -radius; dy <= radius; dy++ {
				ny := clampInt(y+dy, 0, h-1)
				sum += tmp[ny*w+x] * k[dy+radius]
			}
			out[y*w+x] = sum
		}
	}

	return out, nil
}

// ResidualGray computes residual = src - smooth.
func ResidualGray(src, smooth []float32) ([]float32, error) {
	if len(src) != len(smooth) {
		return nil, fmt.Errorf("length mismatch: src=%d smooth=%d", len(src), len(smooth))
	}
	out := make([]float32, len(src))
	for i := 0; i < len(src); i++ {
		out[i] = src[i] - smooth[i]
	}
	return out, nil
}
