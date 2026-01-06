package imageio

import "fmt"

/*
Responsible for cropping a linear image at perticular x0, y0 value
and returning the cropped regions.
*/
func CropAtGray(src []float32, w, h int, x0, y0, cw, ch int) ([]float32, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid dims: %dx%d", w, h)
	}
	if len(src) != w*h {
		return nil, fmt.Errorf("src length mismatch: got %d want %d", len(src), w*h)
	}
	if cw <= 0 || ch <= 0 {
		return nil, fmt.Errorf("invalid crop size: %dx%d", cw, ch)
	}
	if x0 < 0 || y0 < 0 || x0+cw > w || y0+ch > h {
		return nil, fmt.Errorf("crop out of bounds: origin=(%d,%d) size=%dx%d on %dx%d", x0, y0, cw, ch, w, h)
	}

	out := make([]float32, cw*ch)
	for y := 0; y < ch; y++ {
		srcStart := (y0+y)*w + x0
		copy(out[y*cw:(y+1)*cw], src[srcStart:srcStart+cw])
	}
	return out, nil
}
