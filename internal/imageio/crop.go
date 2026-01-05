package imageio

import "fmt"

/*
This function is required to find the optimal centre crop
of an image when there is a size missmatch between the two
during the verification process
*/
func CropCenterGray(src []float32, w, h, cw, ch int) ([]float32, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid dims recieved")
	}
	if len(src) != w*h {
		return nil, fmt.Errorf("src length missmatch against width and height")
	}
	if cw <= 0 || ch <= 0 || cw > w || ch > h {
		return nil, fmt.Errorf("invalid crop")
	}

	x0 := (w - cw) / 2
	y0 := (h - ch) / 2

	out := make([]float32, cw*ch)
	for y := 0; y < ch; y++ {
		srcStart := (y0+y)*w + x0
		copy(out[y*cw:(y+1)*cw], src[srcStart:srcStart+cw])
	}
	return out, nil
}
