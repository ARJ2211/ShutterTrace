package metrics

import "testing"

/*
This test checks that PCE gives
a strong peak when two identical
signals are compared.
*/
func TestPCEIdentity(t *testing.T) {
	w, h := 32, 32
	n := w * h

	a := make([]float32, n)
	b := make([]float32, n)

	// simple synthetic signal
	for i := 0; i < n; i++ {
		v := float32((i % 7) - 3)
		a[i] = v
		b[i] = v
	}

	corr, err := NCCMapFFT(a, b, w, h)
	if err != nil {
		t.Fatalf("NCCMapFFT failed: %v", err)
	}

	stats, err := ComputePCE(corr, w, h, 3)
	if err != nil {
		t.Fatalf("ComputePCE failed: %v", err)
	}

	if stats.PCE <= 1 {
		t.Fatalf("expected strong PCE, got %f", stats.PCE)
	}
}
