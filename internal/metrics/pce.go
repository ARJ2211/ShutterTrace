package metrics

import (
	"fmt"
	"math"
)

type PCEStats struct {
	PCE    float64
	Peak   float64
	PeakX  int
	PeakY  int
	ShiftX int // wrapped shift in [-w/2, w/2]
	ShiftY int // wrapped shift in [-h/2, h/2]
}

/*
small helper for int abs
*/
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

/*
small helper for circular distance
example:
w=10, px=0, x=9 -> |9-0|=9 but circular dist is 1
*/
func circDist(a, b, n int) int {
	d := absInt(a - b)
	if n <= 0 {
		return d
	}
	if d > n-d {
		return n - d
	}
	return d
}

/*
This function is required to compute the
Peak to Correlation Energy (PCE) score
from a correlation map.

Fixes:
- exclude window is now circular (wrap around edges)
- so peak near (0,0) wont break energy estimate
*/
func ComputePCE(corr []float64, w, h int, excludeRadius int) (PCEStats, error) {
	if w <= 0 || h <= 0 {
		return PCEStats{}, fmt.Errorf("invalid dims: %dx%d", w, h)
	}
	if len(corr) != w*h {
		return PCEStats{}, fmt.Errorf("corr length mismatch: got %d want %d", len(corr), w*h)
	}
	if excludeRadius < 0 {
		excludeRadius = 0
	}

	// ---- find peak ----
	peak := corr[0]
	peakIdx := 0
	for i := 1; i < len(corr); i++ {
		if corr[i] > peak {
			peak = corr[i]
			peakIdx = i
		}
	}
	px := peakIdx % w
	py := peakIdx / w

	// ---- compute mean energy excluding circular window around peak ----
	var sum float64
	var count int

	for y := 0; y < h; y++ {
		dy := circDist(y, py, h)
		for x := 0; x < w; x++ {
			dx := circDist(x, px, w)

			// exclude a (2r+1)x(2r+1) box, but circular
			if dx <= excludeRadius && dy <= excludeRadius {
				continue
			}

			v := corr[y*w+x]
			sum += v * v
			count++
		}
	}

	if count == 0 {
		return PCEStats{}, fmt.Errorf("excludeRadius too large: excluded all pixels")
	}

	meanEnergy := sum / float64(count)
	if meanEnergy <= 0 {
		return PCEStats{}, fmt.Errorf("invalid mean energy: %v", meanEnergy)
	}

	pce := (peak * peak) / meanEnergy

	// ---- wrapped shift (circular corr) ----
	shiftX := px
	if shiftX > w/2 {
		shiftX = shiftX - w
	}
	shiftY := py
	if shiftY > h/2 {
		shiftY = shiftY - h
	}

	if math.IsNaN(pce) || math.IsInf(pce, 0) {
		return PCEStats{}, fmt.Errorf("invalid pce computed: %v", pce)
	}

	return PCEStats{
		PCE:    pce,
		Peak:   peak,
		PeakX:  px,
		PeakY:  py,
		ShiftX: shiftX,
		ShiftY: shiftY,
	}, nil
}
