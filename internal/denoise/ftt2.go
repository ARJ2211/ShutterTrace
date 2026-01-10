package denoise

import "gonum.org/v1/gonum/dsp/fourier"

/*
This function is required to perform  a 2D FFT or inverse FFT on a flat complex grid (row-major).
We duplicate this here (instead of importing internal/metrics) to avoid package cycles.
*/
func fft2InPlace(data []complex128, w, h int, inverse bool) {
	fftW := fourier.NewCmplxFFT(w)
	fftH := fourier.NewCmplxFFT(h)

	// rows
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

	// cols
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
}
