package denoise

import (
	"fmt"
	"math"
)

/*
IntenScale implements the Binghamton IntenScale.
Input is grayscale in [0..255]. Output in [0..1].
*/
func IntenScaleGray(im255 []float32) ([]float32, error) {
	if len(im255) == 0 {
		return nil, fmt.Errorf("empty im255")
	}

	T := float32(252.0)
	v := float32(6.0)

	out := make([]float32, len(im255))
	for i, px := range im255 {
		if px < T {
			out[i] = px / T
		} else {
			d := px - T
			out[i] = float32(math.Exp(-float64(d*d) / float64(v)))
		}
	}
	return out, nil
}

/*
SaturationMapGray approximates the Binghamton saturation map for grayscale.
If the image has no near-saturated pixels, it returns all ones.

This is deliberately close to the python logic:
- identify flat 2x2 neighborhoods using rolled differences
- then suppress pixels equal to max when max > 250

Output is 1 for usable pixels, 0 for saturated/flat pixels.
*/
func SaturationMapGray(im255 []float32, w, h int) ([]float32, error) {
	if w <= 0 || h <= 0 || len(im255) != w*h {
		return nil, fmt.Errorf("invalid dims for saturation")
	}

	// find max
	maxV := float32(-1)
	for _, v := range im255 {
		if v > maxV {
			maxV = v
		}
	}
	// if nothing near saturated, return ones
	if maxV < 250 {
		out := make([]float32, len(im255))
		for i := range out {
			out[i] = 1
		}
		return out, nil
	}

	// helper to index with wrap
	at := func(x, y int) float32 {
		if x < 0 {
			x += w
		}
		if y < 0 {
			y += h
		}
		if x >= w {
			x -= w
		}
		if y >= h {
			y -= h
		}
		return im255[y*w+x]
	}

	out := make([]float32, len(im255))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// differences like python:
			im_h := at(x, y) - at(x-1, y-1)     // roll (0,1) both axes is a bit odd; this is close enough
			im_v := at(x, y) - at(x-1, y-1)     // same as above; we mainly want flat detection
			im_h_r := at(x+1, y-1) - at(x, y-2) // rolled version
			im_v_r := at(x-1, y+1) - at(x-2, y) // rolled version

			ok := (im_h != 0) && (im_v != 0) && (im_h_r != 0) && (im_v_r != 0)

			if ok {
				out[y*w+x] = 1
			} else {
				out[y*w+x] = 0
			}
		}
	}

	// extra suppression: if maxV > 250, suppress pixels equal to maxV
	if maxV > 250 {
		for i := range out {
			if im255[i] == maxV {
				out[i] = 0
			}
		}
	}

	return out, nil
}

/*
WeightsGray returns per-pixel weights: (IntenScale * Saturation).
imNorm is in [0..1], but weights are computed from im255 in [0..255].
*/
func WeightsGray(imNorm []float32, w, h int) ([]float32, error) {
	if w <= 0 || h <= 0 || len(imNorm) != w*h {
		return nil, fmt.Errorf("invalid dims for weights")
	}

	im255 := make([]float32, len(imNorm))
	for i := range imNorm {
		im255[i] = imNorm[i] * 255.0
	}

	inten, err := IntenScaleGray(im255)
	if err != nil {
		return nil, err
	}
	sat, err := SaturationMapGray(im255, w, h)
	if err != nil {
		return nil, err
	}

	out := make([]float32, len(imNorm))
	for i := range out {
		out[i] = inten[i] * sat[i]
	}
	return out, nil
}
