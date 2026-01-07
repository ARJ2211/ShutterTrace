package metrics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ARJ2211/ShutterTrace/internal/denoise"
	"github.com/ARJ2211/ShutterTrace/internal/imageio"
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

/*
This test uses real images if dataset exists.
Skipped automatically if not found.

We do PCE on residual (PRNU-ish signal),
not on raw image pixels.
*/
func TestPCESameImage(t *testing.T) {
	imgPath := findAnyImage(t)

	// load image
	pix, w, h, err := imageio.LoadGray(imgPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// compute residual
	blur, err := denoise.GaussianBlurGray(pix, w, h, 1.0)
	if err != nil {
		t.Fatalf("blur failed: %v", err)
	}

	res, err := denoise.ResidualGray(pix, blur)
	if err != nil {
		t.Fatalf("residual failed: %v", err)
	}

	// postprocess (whitening-ish)
	if err := denoise.ZeroMean(res); err != nil {
		t.Fatalf("ZeroMean failed: %v", err)
	}
	if err := denoise.RemoveRowColMean(res, w, h); err != nil {
		t.Fatalf("RemoveRowColMean failed: %v", err)
	}
	if err := denoise.NormalizeL2(res); err != nil {
		t.Fatalf("NormalizeL2 failed: %v", err)
	}

	// correlation map and PCE
	corr, err := NCCMapFFT(res, res, w, h)
	if err != nil {
		t.Fatalf("corr failed: %v", err)
	}

	stats, err := ComputePCE(corr, w, h, 11)
	if err != nil {
		t.Fatalf("pce failed: %v", err)
	}

	// Dont set an aggressive threshold here.
	// Just assert it is clearly > 1, meaning peak stands out.
	if stats.PCE <= 1.2 {
		t.Fatalf("expected PCE > 1.2 for same residual, got %f", stats.PCE)
	}
}
