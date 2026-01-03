package fingerprint

import "fmt"

/*
This function is responsible for computing
the average fingerprint from residual noise
samples. Each residual is a flat slice of some
identical length.
*/
func Estimate(residuals [][]float32) ([]float32, error) {
	if len(residuals) == 0 {
		return nil, fmt.Errorf("invalid residuals: empty")
	}

	size := len(residuals[0])
	if size == 0 {
		return nil, fmt.Errorf("invalid residuals: zero length")
	}

	for _, residual := range residuals {
		if len(residual) != size {
			return nil, fmt.Errorf("residual length mismatch: got %d want %d", len(residual), size)
		}
	}

	fp := make([]float32, size)
	for _, residual := range residuals {
		for i := 0; i < size; i++ {
			fp[i] += residual[i]
		}
	}

	invN := float32(1.0) / float32(len(residuals))
	for i := range fp {
		fp[i] *= invN
	}

	return fp, nil
}
