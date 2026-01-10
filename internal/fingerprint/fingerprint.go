package fingerprint

import (
	"fmt"

	"github.com/ARJ2211/ShutterTrace/internal/denoise"
)

/*
This function is responsible for computing
the average fingerprint from residual noise
samples. Each residual is a flat slice of some
identical length.
*/
func EstimateWeighted(imgs [][]float32, residuals [][]float32, w, h int) ([]float32, error) {
	if len(imgs) == 0 || len(residuals) == 0 {
		return nil, fmt.Errorf("invalid inputs: empty")
	}
	if len(imgs) != len(residuals) {
		return nil, fmt.Errorf("imgs/residuals mismatch: %d vs %d", len(imgs), len(residuals))
	}

	size := w * h
	for i := range imgs {
		if len(imgs[i]) != size {
			return nil, fmt.Errorf("img[%d] length mismatch: got %d want %d", i, len(imgs[i]), size)
		}
		if len(residuals[i]) != size {
			return nil, fmt.Errorf("residual[%d] length mismatch: got %d want %d", i, len(residuals[i]), size)
		}
	}

	RPsum := make([]float32, size)
	NN := make([]float32, size)

	for i := range imgs {
		weights, err := denoise.WeightsGray(imgs[i], w, h)
		if err != nil {
			return nil, err
		}

		im := imgs[i]
		r := residuals[i]

		for p := 0; p < size; p++ {
			RPsum[p] += r[p] * im[p]
			wv := weights[p]
			NN[p] += wv * wv
		}
	}

	K := make([]float32, size)
	for p := 0; p < size; p++ {
		K[p] = RPsum[p] / (NN[p] + 1.0)
	}

	return K, nil
}
