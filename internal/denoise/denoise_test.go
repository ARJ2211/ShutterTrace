package denoise

import (
	"math"
	"testing"
)

func TestGaussianKernel1D_Normalized(t *testing.T) {
	k, radius, err := GaussianKernel1D(1.2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if radius < 1 {
		t.Fatalf("radius too small: %d", radius)
	}
	if len(k) != 2*radius+1 {
		t.Fatalf("kernel size mismatch: got %d want %d", len(k), 2*radius+1)
	}

	sum := float64(0)
	for _, v := range k {
		if v <= 0 {
			t.Fatalf("kernel value not positive: %v", v)
		}
		sum += float64(v)
	}
	if math.Abs(sum-1.0) > 1e-4 {
		t.Fatalf("kernel not normalized: sum=%0.8f", sum)
	}
}

func TestGaussianBlurGray_ConstantImage(t *testing.T) {
	w, h := 32, 24
	src := make([]float32, w*h)
	for i := range src {
		src[i] = 0.42
	}

	out, err := GaussianBlurGray(src, w, h, 1.0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != len(src) {
		t.Fatalf("len mismatch: %d vs %d", len(out), len(src))
	}

	// Blurring a constant image should keep it constant (within tolerance).
	for i := range out {
		if math.Abs(float64(out[i]-0.42)) > 1e-3 {
			t.Fatalf("constant not preserved at i=%d got=%v", i, out[i])
		}
	}
}

func TestResidualGray_LengthAndNonZero(t *testing.T) {
	w, h := 16, 16
	src := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i%7) / 7.0
	}

	blur, err := GaussianBlurGray(src, w, h, 1.0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	res, err := ResidualGray(src, blur)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(res) != len(src) {
		t.Fatalf("len mismatch")
	}

	// Should not be all zeros for non-constant input
	sumAbs := float64(0)
	for _, v := range res {
		sumAbs += math.Abs(float64(v))
	}
	if sumAbs == 0 {
		t.Fatalf("residual unexpectedly all zeros")
	}
	t.Logf("sumAbs=%0.6f", sumAbs)
}
