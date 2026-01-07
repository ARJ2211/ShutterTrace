package metrics

import (
	"fmt"
	"math/cmplx"

	"gonum.org/v1/gonum/dsp/fourier"
)

/*
This function is required to compute the
normalized cross correlation map between
two same size vectors using FFT.

Input vectors must already be:
- zero mean
- row and column mean removed
- L2 normalized

This function does NOT return a single score.
It returns a full correlation surface which
shows how much the two patterns match at
different spatial shifts.

This is needed for PCE computation.
*/
func NCCMapFFT(a, b []float32, w, h int) ([]float64, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid dims: %dx%d", w, h)
	}
	if len(a) != w*h || len(b) != w*h {
		return nil, fmt.Errorf("length mismatch: a=%d b=%d want=%d", len(a), len(b), w*h)
	}

	// Convert to complex grids
	A := make([]complex128, w*h)
	B := make([]complex128, w*h)
	for i := 0; i < w*h; i++ {
		A[i] = complex(float64(a[i]), 0)
		B[i] = complex(float64(b[i]), 0)
	}

	// Forward 2D FFT on both
	fft2InPlace(A, w, h, false)
	fft2InPlace(B, w, h, false)

	// Cross spectrum: A * conj(B)
	C := make([]complex128, w*h)
	for i := 0; i < w*h; i++ {
		C[i] = A[i] * cmplx.Conj(B[i])
	}

	// Inverse 2D FFT to get correlation map
	fft2InPlace(C, w, h, true)

	out := make([]float64, w*h)
	for i := 0; i < w*h; i++ {
		out[i] = real(C[i])
	}

	return out, nil
}

/*
This function performs a 2D FFT or inverse FFT
on a flat complex array stored in row-major order.

Implementation:
- first FFT on each row
- then FFT on each column

Used internally for correlation computation.
Not meant to be called directly.
*/
func fft2InPlace(data []complex128, w, h int, inverse bool) {
	fftW := fourier.NewCmplxFFT(w)
	fftH := fourier.NewCmplxFFT(h)

	// Row FFTs
	row := make([]complex128, w)
	for y := 0; y < h; y++ {
		base := y * w
		copy(row, data[base:base+w])

		if inverse {
			fftW.Sequence(row, row)
		} else {
			fftW.Coefficients(row, row)
		}

		copy(data[base:base+w], row)
	}

	// Column FFTs
	col := make([]complex128, h)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			col[y] = data[y*w+x]
		}

		if inverse {
			fftH.Sequence(col, col)
		} else {
			fftH.Coefficients(col, col)
		}

		for y := 0; y < h; y++ {
			data[y*w+x] = col[y]
		}
	}

	// Note:
	// gonum inverse (Sequence) is not normalized in the way you expect always.
	// For PCE it is fine because PCE depends on peak-to-energy ratio,
	// scaling cancels out. So you dont need to manually divide by w*h here.
}
