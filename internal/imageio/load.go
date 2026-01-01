package imageio

import (
	"fmt"
	"image"
	"os"

	// format specific packages to register decoders for images
	_ "image/jpeg"
	_ "image/png"
)

/*
This function is responsible for reading
a PNG/JPG image and returning a grayscale
float 32 normalized flat image back
*/
func LoadGray(path string) ([]float32, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, 0, 0, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, 0, 0, fmt.Errorf(
			"invalid image bounds detected: %dx%d", w, h)
	}

	pix := make([]float32, w*h)

	for y := range h {
		for x := range w {
			pixel := img.At(x, y)
			r, g, b, _ := pixel.RGBA()

			var rf float32 = float32(r / 65545.0)
			var gf float32 = float32(g / 65545.0)
			var bf float32 = float32(b / 65545.0)

			var gray float32 = (0.299 * rf) + (0.587 * gf) + (0.114 * bf)
			pix[y*w+x] = gray
		}
	}

	return pix, w, h, nil
}
