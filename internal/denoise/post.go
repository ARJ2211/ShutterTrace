package denoise

import (
	"fmt"
	"math"
)

/*
This function is required for NCC where
we do an element wise substraction with
the mean
*/
func ZeroMean(x []float32) error {
	if len(x) == 0 {
		return fmt.Errorf("empty slice revieved")
	}

	var sum float64
	for _, val := range x {
		sum += float64(val)
	}
	mean := float32(sum / float64(len(x)))

	for i := range x {
		x[i] -= mean
	}
	return nil
}

/*
This function is responsible to remove the row and
column wise mean in place so that no artifacts can inflate
the correlation according to:
Goljan, Miroslav, and Jessica Fridrich. "Camera identification from cropped and scaled images." Security, Forensics, Steganography, and Watermarking of Multimedia Contents X. Vol. 6819. SPIE, 2008.
*/
func RemoveRowColMean(x []float32, w, h int) error {
	if w <= 0 || h <= 0 || len(x) != w*h {
		return fmt.Errorf("invalid dimentions recieved")
	}

	// Remove row wise mean
	for y := range h {
		base := y * w

		var sum float64
		for x0 := range w {
			sum += float64(x[base+x0])
		}
		rowMean := float32(sum / float64(w))

		for x0 := range w {
			x[base+x0] -= rowMean
		}
	}

	// Remove column wise mean
	for x0 := 0; x0 < w; x0++ {
		var sum float64
		for y := 0; y < h; y++ {
			sum += float64(x[y*w+x0])
		}
		colMean := float32(sum / float64(h))

		for y := 0; y < h; y++ {
			x[y*w+x0] -= colMean
		}
	}

	return nil
}

/*
This function scales x in place so that L2(x) == 1
L2 normalizer function (Get unit vector)
*/
func NormalizeL2(x []float32) error {
	if len(x) == 0 {
		return fmt.Errorf("empty slice revieved")
	}

	var ss float64
	for _, val := range x {
		f := float64(val)
		ss += math.Pow(f, 2)
	}

	if ss == 0 {
		return fmt.Errorf("zero score not allowed")
	}

	for i := range x {
		x[i] *= float32(1.0 / math.Sqrt(ss))
	}
	return nil
}
