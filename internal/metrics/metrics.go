package metrics

import (
	"fmt"
	"math"
)

/*
This function is required to compute the pearsons
corellation between 2 equal length vectors. Returns
a float64 value of the similarity between 2 vectors
*/
func PearsonCorr(v1, v2 []float32) (float64, error) {
	if len(v1) != len(v2) {
		return 0, fmt.Errorf(
			"non-equal lengths vector detected %d, %d", len(v1), len(v2))
	}
	var size int = len(v1)
	if len(v1) == 0 || len(v2) == 0 {
		return 0, fmt.Errorf("0 lengths vectors detected")
	}

	var sumV1 float64 = 0
	var sumV2 float64 = 0
	for i := range size {
		sumV1 += float64(v1[i])
		sumV2 += float64(v2[i])
	}
	meanV1 := sumV1 / float64(size)
	meanV2 := sumV2 / float64(size)

	var num float64 = 0
	var denV1, denV2 float64 = 0.0, 0.0
	for i := range size {
		dv1 := float64(v1[i]) - meanV1
		dv2 := float64(v2[i]) - meanV2

		num += dv1 * dv2
		denV1 += math.Pow(dv1, 2)
		denV2 += math.Pow(dv2, 2)
	}

	den := math.Sqrt(denV1 * denV2)
	if den == 0.0 {
		return 0, fmt.Errorf("0 varience in input")
	}
	corr := num / den
	return corr, nil
}

func NCCZeroMean(a, b []float32) (float64, error) {
	return PearsonCorr(a, b)
}
