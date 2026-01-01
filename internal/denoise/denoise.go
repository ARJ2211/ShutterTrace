package denoise

import "errors"

// TODO next:
// GaussianBlurGray(src []float32, w, h int, sigma float32) []float32
// ResidualGray(src, smooth []float32) []float32

var ErrNotImplemented = errors.New("not implemented")
