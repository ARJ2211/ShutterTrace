package imageio

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, 0, 0, fmt.Errorf(
			"invalid image bounds detected: %dx%d", w, h)
	}

	pix := make([]float32, w*h)

	for y := range h {
		for x := range w {
			pixel := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			r, g, bl, _ := pixel.RGBA()

			rf := float32(r) / 65535.0
			gf := float32(g) / 65535.0
			bf := float32(bl) / 65535.0

			gray := float32(0.299)*rf + float32(0.587)*gf + float32(0.114)*bf
			pix[y*w+x] = gray
		}
	}

	return pix, w, h, nil
}

/*
This function is responsible for loading all
the images in a directory and returns the
sorted image paths. WE IGNORE SUB-DIRS
*/
func ListImages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			out = append(out, filepath.Join(dir, name))
		}
	}

	sort.Strings(out)
	return out, nil
}
