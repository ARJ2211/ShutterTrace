package imageio

import "testing"

func TestCropAtGray(t *testing.T) {
	// 4x4: 0..15
	w, h := 4, 4
	src := make([]float32, 16)
	for i := range src {
		src[i] = float32(i)
	}

	// crop 2x2 from (1,1):
	// 5 6
	// 9 10
	got, err := CropAtGray(src, w, h, 1, 1, 2, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := []float32{5, 6, 9, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d got %v want %v", i, got[i], want[i])
		}
	}
}
