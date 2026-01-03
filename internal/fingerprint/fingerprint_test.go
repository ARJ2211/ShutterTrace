package fingerprint

import (
	"math"
	"testing"
)

func TestEstimate_BasicAverage(t *testing.T) {
	residuals := [][]float32{
		{1, 2, 3},
		{3, 2, 1},
	}

	fp, err := Estimate(residuals)
	if err != nil {
		t.Fatalf("Estimate error: %v", err)
	}
	if len(fp) != 3 {
		t.Fatalf("len mismatch: got %d want 3", len(fp))
	}

	want := []float32{2, 2, 2}
	for i := range fp {
		if math.Abs(float64(fp[i]-want[i])) > 1e-6 {
			t.Fatalf("fp[%d]=%v want=%v", i, fp[i], want[i])
		}
	}
}

func TestEstimate_Empty(t *testing.T) {
	_, err := Estimate(nil)
	if err == nil {
		t.Fatalf("expected error for empty residuals")
	}
}

func TestEstimate_LengthMismatch(t *testing.T) {
	residuals := [][]float32{
		{1, 2, 3},
		{1, 2},
	}
	_, err := Estimate(residuals)
	if err == nil {
		t.Fatalf("expected error for length mismatch")
	}
}

func TestEstimate_ZeroLength(t *testing.T) {
	residuals := [][]float32{
		{},
		{},
	}
	_, err := Estimate(residuals)
	if err == nil {
		t.Fatalf("expected error for zero-length residuals")
	}
}
