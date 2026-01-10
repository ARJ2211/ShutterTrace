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

/*
StdDev computes sample stddev
*/
func StdDev(x []float32) float32 {
	n := len(x)
	if n <= 1 {
		return 0
	}

	var sum float64
	for _, v := range x {
		sum += float64(v)
	}
	mean := sum / float64(n)

	var ss float64
	for _, v := range x {
		d := float64(v) - mean
		ss += d * d
	}

	return float32(math.Sqrt(ss / float64(n-1)))
}

func zeroMean2DSubsample(x []float32, w, h int, yStart, yStep, xStart, xStep int) error {
	if w <= 0 || h <= 0 || len(x) != w*h {
		return fmt.Errorf("invalid dims")
	}

	// collect indices for this subgrid
	ys := make([]int, 0, (h+1)/2)
	for y := yStart; y < h; y += yStep {
		ys = append(ys, y)
	}
	xs := make([]int, 0, (w+1)/2)
	for xx := xStart; xx < w; xx += xStep {
		xs = append(xs, xx)
	}

	if len(ys) == 0 || len(xs) == 0 {
		return nil
	}

	// global mean over subgrid
	var sum float64
	var count int
	for _, y := range ys {
		base := y * w
		for _, xx := range xs {
			sum += float64(x[base+xx])
			count++
		}
	}
	if count == 0 {
		return nil
	}
	globalMean := float32(sum / float64(count))

	for _, y := range ys {
		base := y * w
		for _, xx := range xs {
			x[base+xx] -= globalMean
		}
	}

	// row means over subgrid
	for _, y := range ys {
		base := y * w
		var rsum float64
		for _, xx := range xs {
			rsum += float64(x[base+xx])
		}
		rowMean := float32(rsum / float64(len(xs)))
		for _, xx := range xs {
			x[base+xx] -= rowMean
		}
	}

	// col means over subgrid
	for _, xx := range xs {
		var csum float64
		for _, y := range ys {
			csum += float64(x[y*w+xx])
		}
		colMean := float32(csum / float64(len(ys)))
		for _, y := range ys {
			x[y*w+xx] -= colMean
		}
	}

	return nil
}

func ZeroMeanTotal(x []float32, w, h int) error {
	if w <= 0 || h <= 0 || len(x) != w*h {
		return fmt.Errorf("invalid dims for ZeroMeanTotal")
	}

	if err := zeroMean2DSubsample(x, w, h, 0, 2, 0, 2); err != nil {
		return err
	}
	if err := zeroMean2DSubsample(x, w, h, 1, 2, 0, 2); err != nil {
		return err
	}
	if err := zeroMean2DSubsample(x, w, h, 0, 2, 1, 2); err != nil {
		return err
	}
	if err := zeroMean2DSubsample(x, w, h, 1, 2, 1, 2); err != nil {
		return err
	}

	return nil
}
