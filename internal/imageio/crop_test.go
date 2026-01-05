package imageio

import "testing"

func TestCropCenterGray(t *testing.T) {
	// 4x4 image with values 0..15
	w, h := 4, 4
	src := make([]float32, 16)
	for i := 0; i < 16; i++ {
		src[i] = float32(i)
	}

	// crop 2x2 center should be:
	// [5 6
	//  9 10]
	got, err := CropCenterGray(src, w, h, 2, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := []float32{5, 6, 9, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d: got %v want %v", i, got[i], want[i])
		}
	}
}
