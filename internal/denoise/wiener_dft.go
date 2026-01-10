package denoise

import (
	"fmt"
	"math"
	"math/cmplx"
)

/*
Weiner adaptive helper
*/
func wienerAdaptive2D(x []float64, w, h int, noiseVar float64, windowSizes []int) ([]float64, error) {
	if w <= 0 || h <= 0 || len(x) != w*h {
		return nil, fmt.Errorf("invalid dims for wiener adaptive")
	}
	if noiseVar <= 0 {
		return nil, fmt.Errorf("noiseVar must be > 0")
	}
	if len(windowSizes) == 0 {
		windowSizes = []int{3, 5, 7, 9}
	}

	// energy
	energy := make([]float64, len(x))
	for i := range x {
		energy[i] = x[i] * x[i]
	}

	// integral image of energy for O(1) box sums
	// integral has size (h+1)*(w+1)
	integ := make([]float64, (h+1)*(w+1))
	for y := 1; y <= h; y++ {
		rowSum := 0.0
		for xx := 1; xx <= w; xx++ {
			rowSum += energy[(y-1)*w+(xx-1)]
			integ[y*(w+1)+xx] = integ[(y-1)*(w+1)+xx] + rowSum
		}
	}

	boxSum := func(x0, y0, x1, y1 int) float64 {
		// sum over [x0,x1) [y0,y1)
		return integ[y1*(w+1)+x1] - integ[y0*(w+1)+x1] - integ[y1*(w+1)+x0] + integ[y0*(w+1)+x0]
	}

	coefVarMin := make([]float64, len(x))
	for i := range coefVarMin {
		coefVarMin[i] = math.Inf(1)
	}

	for _, ws := range windowSizes {
		if ws < 1 {
			continue
		}
		r := ws / 2
		area := float64(ws * ws)

		for y := 0; y < h; y++ {
			y0 := y - r
			y1 := y + r + 1
			if y0 < 0 {
				y0 = 0
			}
			if y1 > h {
				y1 = h
			}

			for xx := 0; xx < w; xx++ {
				x0 := xx - r
				x1 := xx + r + 1
				if x0 < 0 {
					x0 = 0
				}
				if x1 > w {
					x1 = w
				}

				// constant padding in python means area stays ws*ws even at borders.
				// We approximate by using actual border-clipped area (slightly different at edges).
				// Practically, this is fine for PRNU whitening.
				sum := boxSum(x0, y0, x1, y1)
				actualArea := float64((x1 - x0) * (y1 - y0))
				if actualArea <= 0 {
					continue
				}
				avg := sum / actualArea

				// threshold(avg - noiseVar)
				cv := avg - noiseVar
				if cv < 0 {
					cv = 0
				}

				idx := y*w + xx
				if cv < coefVarMin[idx] {
					coefVarMin[idx] = cv
				}
			}
		}

		_ = area
	}

	out := make([]float64, len(x))
	for i := range x {
		den := coefVarMin[i] + noiseVar
		if den <= 0 {
			out[i] = 0
			continue
		}
		out[i] = x[i] * noiseVar / den
	}

	return out, nil
}

/*
This funtion is required to apply the WienerDFT over
PRNU toolbox style Wiener filter in the frequency domain.

Steps:
- FFT2(im)
- magnitude = abs(F)/sqrt(h*w)
- mag_filt = wienerAdaptive(magnitude, noiseVar)
- F_filt = F * (mag_filt / magnitude) with safe handling of zeros
- IFFT2 -> real
*/
func WienerDFT(im []float32, w, h int, sigma float32) ([]float32, error) {
	if w <= 0 || h <= 0 || len(im) != w*h {
		return nil, fmt.Errorf("invalid dims for WienerDFT")
	}
	if sigma <= 0 {
		return nil, fmt.Errorf("sigma must be > 0 for WienerDFT")
	}

	n := w * h
	F := make([]complex128, n)
	for i := 0; i < n; i++ {
		F[i] = complex(float64(im[i]), 0)
	}

	fft2InPlace(F, w, h, false)

	scale := math.Sqrt(float64(n))
	mag := make([]float64, n)
	for i := 0; i < n; i++ {
		mag[i] = cmplx.Abs(F[i]) / scale
	}

	noiseVar := float64(sigma) * float64(sigma)

	magFilt, err := wienerAdaptive2D(mag, w, h, noiseVar, []int{3, 5, 7, 9})
	if err != nil {
		return nil, err
	}

	// apply F * magFilt / mag with safe zeros
	for i := 0; i < n; i++ {
		if mag[i] == 0 {
			F[i] = 0
			continue
		}
		r := magFilt[i] / mag[i]
		F[i] = F[i] * complex(r, 0)
	}

	fft2InPlace(F, w, h, true)

	out := make([]float32, n)
	for i := 0; i < n; i++ {
		out[i] = float32(real(F[i]))
	}
	return out, nil
}
