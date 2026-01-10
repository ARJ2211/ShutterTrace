package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ARJ2211/ShutterTrace/internal/cli"
	"github.com/ARJ2211/ShutterTrace/internal/denoise"
	"github.com/ARJ2211/ShutterTrace/internal/fingerprint"
	"github.com/ARJ2211/ShutterTrace/internal/imageio"
	"github.com/ARJ2211/ShutterTrace/internal/metrics"
	"github.com/ARJ2211/ShutterTrace/internal/store"

	"github.com/yarlson/pin"
)

type Stoppable interface {
	Stop(msg ...string)
}

func stopProcessExec(p Stoppable, failMsg, msg string) {
	if p != nil {
		p.Stop(failMsg, msg)
	}
	fmt.Println(msg)
	os.Exit(1)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

type tileSpec struct {
	name string
	x0   int
	y0   int
}

func normalizeMetric(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "pearson"
	}
	switch s {
	case "pearson", "pce", "both":
		return s
	default:
		return "pearson"
	}
}

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(2)
	}

	switch os.Args[1] {

	case "enroll":
		enrollCmd := flag.NewFlagSet("enroll", flag.ExitOnError)

		camera := enrollCmd.String("camera", "", "camera id/name (required)")
		inDir := enrollCmd.String("in", "", "input folder of enrollment images (required)")
		outDir := enrollCmd.String("out", "./db", "output db folder (default ./db)")
		sigma := enrollCmd.Float64("sigma", 1.0, "sigma value for gaussian kernel")

		enrollCmd.Parse(os.Args[2:])

		if *camera == "" || *inDir == "" {
			fmt.Println("missing required flags: --camera and --in")
			fmt.Println()
			cli.PrintEnrollUsage()
			os.Exit(2)
		}

		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fmt.Println("failed to create output dir:", err)
			os.Exit(1)
		}

		camDir := filepath.Join(*outDir, *camera)
		if err := os.MkdirAll(camDir, 0o755); err != nil {
			fmt.Println("failed to create camera dir:", err)
			os.Exit(1)
		}

		imageList, err := imageio.ListImages(*inDir)
		if err != nil {
			fmt.Println("failed to read input directory:", err)
			os.Exit(1)
		}
		if len(imageList) == 0 {
			fmt.Printf("no images found in directory %s\n", *inDir)
			os.Exit(1)
		}

		start := time.Now()
		p := pin.New("Starting enrollment...")
		cancel := p.Start(context.Background())
		defer cancel()

		p.UpdateMessage("Loading first image to lock resolution")
		_, w, h, err := imageio.LoadGray(imageList[0])
		if err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("failed to load first image: %v", err))
		}

		imgs := make([][]float32, 0, len(imageList))
		residualsRaw := make([][]float32, 0, len(imageList))

		for i, path := range imageList {
			p.UpdateMessage(fmt.Sprintf("Processing %d/%d: %s", i+1, len(imageList), filepath.Base(path)))

			img, wN, hN, err := imageio.LoadGray(path)
			if err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("error loading image %s: %v", path, err))
			}
			if w != wN || h != hN {
				stopProcessExec(
					p,
					"Enroll failed",
					fmt.Sprintf("dimension mismatch for %s: got %dx%d expected %dx%d", path, wN, hN, w, h),
				)
			}

			blur, err := denoise.GaussianBlurGray(img, wN, hN, float32(*sigma))
			if err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("error blurring image %s: %v", path, err))
			}

			res, err := denoise.ResidualGray(img, blur)
			if err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("error computing residual %s: %v", path, err))
			}

			// IMPORTANT:
			// For fingerprint estimation we want the "raw" residual (before L2 norm),
			// and then we apply ZeroMeanTotal + WienerDFT on the final fingerprint.
			imgs = append(imgs, img)
			residualsRaw = append(residualsRaw, res)
		}

		p.UpdateMessage("Estimating weighted fingerprint (RPsum/(NN+1))")

		fp, err := fingerprint.EstimateWeighted(imgs, residualsRaw, w, h)
		if err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("failed to estimate fingerprint: %v", err))
		}

		// Post-processing (PRNU toolbox style)
		if err := denoise.ZeroMeanTotal(fp, w, h); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess fp ZeroMeanTotal failed: %v", err))
		}

		s := denoise.StdDev(fp)
		if s <= 0 {
			stopProcessExec(p, "Enroll failed", "postprocess fp StdDev is zero")
		}

		fp2, err := denoise.WienerDFT(fp, w, h, s)
		if err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess fp WienerDFT failed: %v", err))
		}
		fp = fp2

		if err := denoise.NormalizeL2(fp); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess fp NormalizeL2 failed: %v", err))
		}

		version := 1
		if oldMeta, err := store.ReadMeta(camDir); err == nil {
			version = oldMeta.Version + 1
		}

		meta := store.Meta{
			Version:    version,
			CameraID:   *camera,
			Width:      w,
			Height:     h,
			ColorMode:  "grayscale",
			DenoiseAlg: "gaussian",
			Notes:      "fingerprint computed by RPsum/(NN+1) + ZeroMeanTotal + WienerDFT",
			Sigma:      float32(*sigma),
		}

		if err := store.WriteMeta(camDir, meta); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("failed to write meta: %v", err))
		}
		if err := store.WriteFingerprint(camDir, fp); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("failed to write fingerprint: %v", err))
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		p.Stop(fmt.Sprintf("Enroll complete (%d images, %dx%d, sigma=%.2f) in %s", len(imageList), w, h, *sigma, elapsed))

		fmt.Println("enroll ok")
		fmt.Println("camera:", *camera)
		fmt.Println("input:", *inDir)
		fmt.Println("images used:", len(imageList))
		fmt.Println("resolution:", w, "x", h)
		fmt.Println("sigma:", *sigma)
		fmt.Println("db:", camDir)

	case "verify":
		verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)

		camera := verifyCmd.String("camera", "", "camera id/name (required)")
		dbDir := verifyCmd.String("db", "./db", "db folder (default ./db)")
		imgPath := verifyCmd.String("img", "", "test image path (required)")
		metricFlag := verifyCmd.String("metric", "pearson", "metric: pearson | pce | both (default pearson)")

		verifyCmd.Parse(os.Args[2:])

		if *camera == "" || *imgPath == "" {
			fmt.Println("missing required flags: --camera and --img")
			fmt.Println()
			cli.PrintVerifyUsage()
			os.Exit(2)
		}

		metric := normalizeMetric(*metricFlag)

		start := time.Now()
		p := pin.New("Starting verification...")
		cancel := p.Start(context.Background())
		defer cancel()

		p.UpdateMessage("Loading camera metadata")

		camDir := filepath.Join(*dbDir, *camera)
		meta, err := store.ReadMeta(camDir)
		if err != nil {
			stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to read camera meta: %v", err))
		}

		p.UpdateMessage("Loading stored fingerprint")
		fp, err := store.ReadFingerprint(camDir)
		if err != nil {
			stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to read fingerprint: %v", err))
		}

		p.UpdateMessage("Loading test image")
		imgPix, w, h, err := imageio.LoadGray(*imgPath)
		if err != nil {
			stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to load test image: %v", err))
		}

		enW, enH := meta.Width, meta.Height

		mode := "exact"
		if w != enW || h != enH {
			mode = "center-crop"
		}

		commonW := minInt(enW, w)
		commonH := minInt(enH, h)

		commonCap := 3072
		if commonW > commonCap {
			commonW = commonCap
		}
		if commonH > commonCap {
			commonH = commonCap
		}

		tileCap := 2048
		tile := min3(tileCap, commonW, commonH)

		if mode == "exact" && (tile != enW || tile != enH) {
			mode = "exact-tiles"
		} else if mode == "center-crop" {
			mode = "center-crop-tiles"
		}

		p.UpdateMessage(fmt.Sprintf("Aligning to common region (%s, %dx%d)", mode, commonW, commonH))

		fpCommon, err := imageio.CropCenterGray(fp, enW, enH, commonW, commonH)
		if err != nil {
			stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to crop fingerprint common: %v", err))
		}

		imgCommon, err := imageio.CropCenterGray(imgPix, w, h, commonW, commonH)
		if err != nil {
			stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to crop test image common: %v", err))
		}

		tiles := []tileSpec{
			{name: "center", x0: (commonW - tile) / 2, y0: (commonH - tile) / 2},
			{name: "top-left", x0: 0, y0: 0},
			{name: "top-right", x0: commonW - tile, y0: 0},
			{name: "bottom-left", x0: 0, y0: commonH - tile},
			{name: "bottom-right", x0: commonW - tile, y0: commonH - tile},
		}

		p.UpdateMessage("Scoring tiles")

		bestPearson := float64(-1e9)
		bestPCE := float64(-1e9)
		bestTileByPearson := ""
		bestTileByPCE := ""

		tileScores := make([]string, 0, len(tiles))

		for i, t := range tiles {
			p.UpdateMessage(fmt.Sprintf("Tile %d/%d: %s", i+1, len(tiles), t.name))

			fpTile, err := imageio.CropAtGray(fpCommon, commonW, commonH, t.x0, t.y0, tile, tile)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to crop fingerprint tile %s: %v", t.name, err))
			}

			imgTile, err := imageio.CropAtGray(imgCommon, commonW, commonH, t.x0, t.y0, tile, tile)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("failed to crop test tile %s: %v", t.name, err))
			}

			blur, err := denoise.GaussianBlurGray(imgTile, tile, tile, meta.Sigma)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("blur failed (%s): %v", t.name, err))
			}

			res, err := denoise.ResidualGray(imgTile, blur)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("residual failed (%s): %v", t.name, err))
			}

			// Post-process residual (PRNU toolbox style)
			if err := denoise.ZeroMeanTotal(res, tile, tile); err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess res ZeroMeanTotal failed (%s): %v", t.name, err))
			}
			rs := denoise.StdDev(res)
			if rs <= 0 {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess res StdDev zero (%s)", t.name))
			}
			res2, err := denoise.WienerDFT(res, tile, tile, rs)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess res WienerDFT failed (%s): %v", t.name, err))
			}
			res = res2
			if err := denoise.NormalizeL2(res); err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess res NormalizeL2 failed (%s): %v", t.name, err))
			}

			// fingerprint tile should already be stored post-processed and L2-normalized,
			// but cropping changes mean/border balance slightly, so re-normalize here.
			if err := denoise.ZeroMeanTotal(fpTile, tile, tile); err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess fpTile ZeroMeanTotal failed (%s): %v", t.name, err))
			}
			fs := denoise.StdDev(fpTile)
			if fs <= 0 {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess fpTile StdDev zero (%s)", t.name))
			}
			fp2, err := denoise.WienerDFT(fpTile, tile, tile, fs)
			if err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess fpTile WienerDFT failed (%s): %v", t.name, err))
			}
			fpTile = fp2
			if err := denoise.NormalizeL2(fpTile); err != nil {
				stopProcessExec(p, "Verify failed", fmt.Sprintf("postprocess fpTile NormalizeL2 failed (%s): %v", t.name, err))
			}

			pearsonVal := float64(0)
			pceVal := float64(0)

			if metric == "pearson" || metric == "both" {
				val, e := metrics.PearsonCorr(fpTile, res)
				if e != nil {
					stopProcessExec(p, "Verify failed", fmt.Sprintf("pearson failed (%s): %v", t.name, e))
				}
				pearsonVal = val
				if pearsonVal > bestPearson {
					bestPearson = pearsonVal
					bestTileByPearson = t.name
				}
			}

			if metric == "pce" || metric == "both" {
				corrMap, e := metrics.NCCMapFFT(fpTile, res, tile, tile)
				if e != nil {
					stopProcessExec(p, "Verify failed", fmt.Sprintf("fft corr failed (%s): %v", t.name, e))
				}

				stats, e := metrics.ComputePCE(corrMap, tile, tile, 11)
				if e != nil {
					stopProcessExec(p, "Verify failed", fmt.Sprintf("pce failed (%s): %v", t.name, e))
				}

				// Gate: keep it, but you can relax to 4 while debugging.
				if !metrics.IsNearZeroShift(stats, 2) {
					stats.PCE = -1
				}

				pceVal = stats.PCE
				if pceVal > bestPCE {
					bestPCE = pceVal
					bestTileByPCE = t.name
				}
			}

			switch metric {
			case "pearson":
				tileScores = append(tileScores, fmt.Sprintf("%s=%.6f", t.name, pearsonVal))
			case "pce":
				tileScores = append(tileScores, fmt.Sprintf("%s=%.3f", t.name, pceVal))
			default:
				tileScores = append(tileScores, fmt.Sprintf("%s=pearson:%.6f pce:%.3f", t.name, pearsonVal, pceVal))
			}
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		p.Stop(fmt.Sprintf("Verify complete in %s", elapsed))

		fmt.Println("verify ok")
		fmt.Println("camera:", meta.CameraID)
		fmt.Println("db:", camDir)
		fmt.Println("enrolled resolution:", fmt.Sprintf("%dx%d", enW, enH))
		fmt.Println("test resolution:", fmt.Sprintf("%dx%d", w, h))
		fmt.Println("match mode:", mode)
		fmt.Println("common region:", fmt.Sprintf("%dx%d", commonW, commonH))
		fmt.Println("tile used:", fmt.Sprintf("%dx%d", tile, tile))
		fmt.Println("metric:", metric)

		switch metric {
		case "pearson":
			fmt.Println("best tile:", bestTileByPearson)
			fmt.Println("score (pearson):", bestPearson)
		case "pce":
			fmt.Println("best tile:", bestTileByPCE)
			fmt.Println("score (pce):", bestPCE)
		default:
			fmt.Println("best tile (pce):", bestTileByPCE)
			fmt.Println("score (pce):", bestPCE)
			fmt.Println("best tile (pearson):", bestTileByPearson)
			fmt.Println("score (pearson):", bestPearson)
		}

		fmt.Println("tile scores:", strings.Join(tileScores, ", "))

	case "help", "-h", "--help":
		cli.PrintUsage()

	default:
		fmt.Println("unknown command:", os.Args[1])
		fmt.Println()
		cli.PrintUsage()
		os.Exit(2)
	}
}
