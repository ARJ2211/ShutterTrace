package denoise

import (
	"math"
	"testing"
)

func meanF32(x []float32) float64 {
	var s float64
	for _, v := range x {
		s += float64(v)
	}
	return s / float64(len(x))
}

func l2F32(x []float32) float64 {
	var ss float64
	for _, v := range x {
		f := float64(v)
		ss += f * f
	}
	return math.Sqrt(ss)
}

func TestZeroMean(t *testing.T) {
	x := []float32{1, 2, 3, 4}
	if err := ZeroMean(x); err != nil {
		t.Fatalf("ZeroMean err: %v", err)
	}
	m := meanF32(x)
	if math.Abs(m) > 1e-6 {
		t.Fatalf("mean not ~0: %v", m)
	}
}

func TestRemoveRowColMean(t *testing.T) {
	// 3x3 matrix:
	// 1 2 3
	// 4 5 6
	// 7 8 9
	w, h := 3, 3
	x := []float32{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}
	if err := RemoveRowColMean(x, w, h); err != nil {
		t.Fatalf("RemoveRowColMean err: %v", err)
	}

	// Check each row mean ~ 0
	for y := 0; y < h; y++ {
		row := x[y*w : (y+1)*w]
		m := meanF32(row)
		if math.Abs(m) > 1e-5 {
			t.Fatalf("row %d mean not ~0: %v", y, m)
		}
	}

	// Check each col mean ~ 0
	for x0 := 0; x0 < w; x0++ {
		var s float64
		for y := 0; y < h; y++ {
			s += float64(x[y*w+x0])
		}
		m := s / float64(h)
		if math.Abs(m) > 1e-5 {
			t.Fatalf("col %d mean not ~0: %v", x0, m)
		}
	}
}

func TestNormalizeL2(t *testing.T) {
	x := []float32{3, 4}
	if err := NormalizeL2(x); err != nil {
		t.Fatalf("NormalizeL2 err: %v", err)
	}
	n := l2F32(x)
	if math.Abs(n-1.0) > 1e-6 {
		t.Fatalf("norm not ~1: %v", n)
	}
}
