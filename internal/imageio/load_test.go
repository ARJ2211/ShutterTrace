package imageio

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func findAnyImage(t *testing.T) string {
	t.Helper()

	root := filepath.Join("..", "..", "datasets")
	_, err := os.Stat(root)
	if err != nil {
		t.Skip("datasets/ folder not found, skipping imageio tests " + err.Error())
		return ""
	}

	var found string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG":
			found = path
		}
		return nil
	})

	if found == "" {
		t.Skip("no images found under datasets/, skipping")
	}
	return found
}

func TestLoadGrayBasic(t *testing.T) {
	imgPath := findAnyImage(t)

	pix, w, h, err := LoadGray(imgPath)
	if err != nil {
		t.Fatalf("LoadGray error: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Fatalf("invalid dims: %dx%d", w, h)
	}
	if len(pix) != w*h {
		t.Fatalf("pix length mismatch: got %d want %d", len(pix), w*h)
	}

	minV := float32(1.0)
	maxV := float32(0.0)
	sum := float64(0)

	for _, v := range pix {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
		sum += float64(v)
		if v < 0 || v > 1 || math.IsNaN(float64(v)) {
			t.Fatalf("value out of range [0,1] or NaN: %v", v)
		}
	}

	mean := sum / float64(len(pix))
	t.Logf("loaded %s (%dx%d) min=%.4f max=%.4f mean=%.4f", imgPath, w, h, minV, maxV, mean)
}
