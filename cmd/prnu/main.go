package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

/*
Helper function to stop process exec
*/
func stopProcessExec(p Stoppable, failMsg, msg string) {
	if p != nil {
		p.Stop(failMsg)
	}
	fmt.Println(msg)
	os.Exit(1)
}

/*
Helper function to return smallest integer
*/
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			p.Stop("Enroll failed")
			fmt.Println("failed to load first image:", err)
			os.Exit(1)
		}

		residuals := make([][]float32, 0, len(imageList))

		for i, path := range imageList {
			p.UpdateMessage(fmt.Sprintf("Processing %d/%d: %s", i+1, len(imageList), filepath.Base(path)))

			img, wN, hN, err := imageio.LoadGray(path)
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error loading image:", path, err)
				os.Exit(1)
			}

			if w != wN || h != hN {
				p.Stop("Enroll failed")
				fmt.Printf("dimension mismatch for %s: got %dx%d expected %dx%d\n", path, wN, hN, w, h)
				os.Exit(1)
			}

			blur, err := denoise.GaussianBlurGray(img, wN, hN, float32(*sigma))
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error blurring image:", path, err)
				os.Exit(1)
			}

			res, err := denoise.ResidualGray(img, blur)
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error computing residual:", path, err)
				os.Exit(1)
			}

			// Post processing
			if err := denoise.ZeroMean(res); err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess ZeroMean failed: %v", err))
			}
			if err := denoise.RemoveRowColMean(res, wN, hN); err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess RemoveRowColMean failed: %v", err))
			}
			if err := denoise.NormalizeL2(res); err != nil {
				stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess NormalizeL2 failed: %v", err))
			}

			residuals = append(residuals, res)
		}

		p.UpdateMessage("Estimating fingerprint")

		fp, err := fingerprint.Estimate(residuals)
		if err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to estimate fingerprint:", err)
			os.Exit(1)
		}
		// Post processing
		if err := denoise.ZeroMean(fp); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess ZeroMean failed: %v", err))
		}
		if err := denoise.RemoveRowColMean(fp, w, h); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess RemoveRowColMean failed: %v", err))
		}
		if err := denoise.NormalizeL2(fp); err != nil {
			stopProcessExec(p, "Enroll failed", fmt.Sprintf("postprocess NormalizeL2 failed: %v", err))
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
			Notes:      "fingerprint computed by averaging gaussian residuals",
			Sigma:      float32(*sigma),
		}

		if err := store.WriteMeta(camDir, meta); err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to write meta:", err)
			os.Exit(1)
		}

		if err := store.WriteFingerprint(camDir, fp); err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to write fingerprint:", err)
			os.Exit(1)
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		p.Stop(fmt.Sprintf("Enroll complete (%d images, %dx%d, sigma=%.2f) in %s",
			len(imageList), w, h, *sigma, elapsed))

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

		verifyCmd.Parse(os.Args[2:])

		if *camera == "" || *imgPath == "" {
			fmt.Println("missing required flags: --camera and --img")
			fmt.Println()
			cli.PrintVerifyUsage()
			os.Exit(2)
		}

		start := time.Now()
		p := pin.New("Starting verification...")
		cancel := p.Start(context.Background())
		defer cancel()

		p.UpdateMessage("Loading camera metadata")

		camDir := filepath.Join(*dbDir, *camera)
		meta, err := store.ReadMeta(camDir)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to read camera meta:", err)
			os.Exit(1)
		}

		p.UpdateMessage("Loading stored fingerprint")

		fp, err := store.ReadFingerprint(camDir)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to read fingerprint:", err)
			os.Exit(1)
		}

		p.UpdateMessage("Loading test image")

		imgPix, w, h, err := imageio.LoadGray(*imgPath)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to load test image:", err)
			os.Exit(1)
		}

		enW, enH := meta.Width, meta.Height

		cw := minInt(enW, w)
		ch := minInt(enH, h)

		maxTile := 2048
		if cw > maxTile {
			cw = maxTile
		}
		if ch > maxTile {
			ch = maxTile
		}

		mode := "exact"
		if w != enW || h != enH {
			mode = "center-crop"
		}

		tileCap := 2048
		if cw > tileCap {
			cw = tileCap
		}
		if ch > tileCap {
			ch = tileCap
		}

		if mode == "exact" && (cw != enW || ch != enH) {
			mode = "exact-tile"
		}
		if mode == "center-crop" && (cw != minInt(enW, w) || ch != minInt(enH, h)) {
			mode = "center-crop-tile"
		}

		p.UpdateMessage(fmt.Sprintf(
			"Preparing match region (%s, %dx%d)",
			mode, cw, ch,
		))

		fpCrop, err := imageio.CropCenterGray(fp, enW, enH, cw, ch)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to crop fingerprint:", err)
			os.Exit(1)
		}

		// Post processing
		if err := denoise.ZeroMean(fpCrop); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute zero mean")
		}
		if err := denoise.RemoveRowColMean(fpCrop, cw, ch); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute remove row col mean")
		}
		if err := denoise.NormalizeL2(fpCrop); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute l2 normalization")
		}

		imgCrop, err := imageio.CropCenterGray(imgPix, w, h, cw, ch)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to crop test image:", err)
			os.Exit(1)
		}

		p.UpdateMessage("Computing residual")

		blur, err := denoise.GaussianBlurGray(imgCrop, cw, ch, meta.Sigma)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to blur test image:", err)
			os.Exit(1)
		}

		res, err := denoise.ResidualGray(imgCrop, blur)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to compute residual:", err)
			os.Exit(1)
		}

		// Post processing
		if err := denoise.ZeroMean(res); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute zero mean")
		}
		if err := denoise.RemoveRowColMean(res, cw, ch); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute remove row col mean")
		}
		if err := denoise.NormalizeL2(res); err != nil {
			stopProcessExec(p, "Verify failed", "could not compute l2 normalization")
		}

		p.UpdateMessage("Computing similarity score")

		score, err := metrics.PearsonCorr(fpCrop, res)
		if err != nil {
			p.Stop("Verify failed")
			fmt.Println("failed to compute score:", err)
			os.Exit(1)
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		p.Stop(fmt.Sprintf("Verify complete in %s", elapsed))

		fmt.Println("verify ok")
		fmt.Println("camera:", meta.CameraID)
		fmt.Println("db:", camDir)
		fmt.Println("enrolled resolution:", fmt.Sprintf("%dx%d", enW, enH))
		fmt.Println("test resolution:", fmt.Sprintf("%dx%d", w, h))
		fmt.Println("match mode:", mode)
		fmt.Println("region used:", fmt.Sprintf("%dx%d", cw, ch))
		fmt.Println("score (pearson):", score)

	case "help", "-h", "--help":
		cli.PrintUsage()

	default:
		fmt.Println("unknown command:", os.Args[1])
		fmt.Println()
		cli.PrintUsage()
		os.Exit(2)
	}
}
