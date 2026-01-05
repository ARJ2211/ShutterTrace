package metrics

import (
	"math"
	"testing"
)

func TestPearson_PerfectPositive(t *testing.T) {
	a := []float32{1, 2, 3, 4, 5}
	b := []float32{2, 4, 6, 8, 10}
	r, err := PearsonCorr(a, b)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(r-1.0) > 1e-9 {
		t.Fatalf("want ~1 got %v", r)
	}
}

func TestPearson_PerfectNegative(t *testing.T) {
	a := []float32{1, 2, 3, 4}
	b := []float32{4, 3, 2, 1}
	r, err := PearsonCorr(a, b)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(r+1.0) > 1e-9 {
		t.Fatalf("want ~-1 got %v", r)
	}
}

func TestPearson_LengthMismatch(t *testing.T) {
	_, err := PearsonCorr([]float32{1, 2}, []float32{1})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPearson_ZeroVariance(t *testing.T) {
	_, err := PearsonCorr([]float32{1, 1, 1}, []float32{2, 3, 4})
	if err == nil {
		t.Fatalf("expected error")
	}
}
