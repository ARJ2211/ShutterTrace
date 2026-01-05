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
	"github.com/ARJ2211/ShutterTrace/internal/store"

	"github.com/yarlson/pin"
)

func main() {
	// ---- BASIC ARG CHECK ----
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(2)
	}

	switch os.Args[1] {

	// ============================================================
	// ======================== ENROLL =============================
	// ============================================================
	case "enroll":
		enrollCmd := flag.NewFlagSet("enroll", flag.ExitOnError)

		// ---- CLI FLAGS ----
		camera := enrollCmd.String("camera", "", "camera id/name (required)")
		inDir := enrollCmd.String("in", "", "input folder of enrollment images (required)")
		outDir := enrollCmd.String("out", "./db", "output db folder (default ./db)")
		sigma := enrollCmd.Float64("sigma", 1.0, "sigma value for gaussian kernel")

		enrollCmd.Parse(os.Args[2:])

		// ---- REQUIRED FLAGS CHECK ----
		if *camera == "" || *inDir == "" {
			fmt.Println("missing required flags: --camera and --in")
			fmt.Println()
			cli.PrintEnrollUsage()
			os.Exit(2)
		}

		// ---- CREATE OUTPUT DB DIRECTORY ----
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fmt.Println("failed to create output dir:", err)
			os.Exit(1)
		}

		// ---- CREATE CAMERA SPECIFIC DIRECTORY ----
		camDir := filepath.Join(*outDir, *camera)
		if err := os.MkdirAll(camDir, 0o755); err != nil {
			fmt.Println("failed to create camera dir:", err)
			os.Exit(1)
		}

		// ---- LIST ALL ENROLLMENT IMAGES ----
		imageList, err := imageio.ListImages(*inDir)
		if err != nil {
			fmt.Println("failed to read input directory:", err)
			os.Exit(1)
		}
		if len(imageList) == 0 {
			fmt.Printf("no images found in directory %s\n", *inDir)
			os.Exit(1)
		}

		// ---- SPINNER START ----
		start := time.Now()
		p := pin.New("Starting enrollment...")
		cancel := p.Start(context.Background())
		defer cancel()

		p.UpdateMessage("Loading first image (lock width/height)...")

		// ---- LOAD FIRST IMAGE TO FIX WIDTH / HEIGHT ----
		_, w, h, err := imageio.LoadGray(imageList[0])
		if err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to load first image:", err)
			os.Exit(1)
		}

		// ---- PROCESS ALL IMAGES AND COLLECT RESIDUALS ----
		p.UpdateMessage(fmt.Sprintf("Computing residuals (0/%d)...", len(imageList)))

		residuals := make([][]float32, 0, len(imageList))
		total := len(imageList)

		for idx, path := range imageList {
			// progress update
			p.UpdateMessage(fmt.Sprintf("Processing %d/%d: %s", idx+1, total, filepath.Base(path)))

			// load grayscale image
			img, wN, hN, err := imageio.LoadGray(path)
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error loading image:", path, err)
				os.Exit(1)
			}

			// enforce identical resolution across all images
			if w != wN || h != hN {
				p.Stop("Enroll failed")
				fmt.Printf("dimension mismatch for %s: got %dx%d expected %dx%d\n", path, wN, hN, w, h)
				os.Exit(1)
			}

			// apply gaussian denoising
			blur, err := denoise.GaussianBlurGray(img, wN, hN, float32(*sigma))
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error blurring image:", path, err)
				os.Exit(1)
			}

			// compute noise residual
			res, err := denoise.ResidualGray(img, blur)
			if err != nil {
				p.Stop("Enroll failed")
				fmt.Println("error computing residual:", path, err)
				os.Exit(1)
			}

			residuals = append(residuals, res)
		}

		// ---- ESTIMATE CAMERA FINGERPRINT ----
		p.UpdateMessage("Estimating fingerprint (averaging residuals)...")

		fp, err := fingerprint.Estimate(residuals)
		if err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to estimate fingerprint:", err)
			os.Exit(1)
		}

		// ---- WRITE META INFORMATION ----
		p.UpdateMessage("Writing meta.json and fingerprint.bin...")
		var version int = 1
		existingMeta, err := store.ReadMeta(camDir)
		if err != nil {
			p.UpdateMessage("No existing meta found for device")
		} else {
			p.UpdateMessage("Existing camera ID found, bumping version")
			version = existingMeta.Version + 1
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

		// ---- WRITE FINGERPRINT TO DISK ----
		if err := store.WriteFingerprint(camDir, fp); err != nil {
			p.Stop("Enroll failed")
			fmt.Println("failed to write fingerprint:", err)
			os.Exit(1)
		}

		// ---- SPINNER STOP ----
		elapsed := time.Since(start).Round(time.Millisecond)
		p.Stop(fmt.Sprintf("Enroll complete (%d images, %dx%d, sigma=%.2f) in %s", total, w, h, *sigma, elapsed))

		// ---- ENROLL SUMMARY ----
		fmt.Println("enroll ok")
		fmt.Println("camera:", *camera)
		fmt.Println("input:", *inDir)
		fmt.Println("images used:", len(imageList))
		fmt.Println("resolution:", w, "x", h)
		fmt.Println("sigma:", *sigma)
		fmt.Println("db:", camDir)

	// ============================================================
	// ======================== VERIFY ==============================
	// ============================================================
	case "verify":
		verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)

		// ---- CLI FLAGS ----
		camera := verifyCmd.String("camera", "", "camera id/name (required)")
		dbDir := verifyCmd.String("db", "./db", "db folder (default ./db)")
		img := verifyCmd.String("img", "", "test image path (required)")

		verifyCmd.Parse(os.Args[2:])

		// ---- REQUIRED FLAGS CHECK ----
		if *camera == "" || *img == "" {
			fmt.Println("missing required flags: --camera and --img")
			fmt.Println()
			cli.PrintVerifyUsage()
			os.Exit(2)
		}

		// ---- LOAD META (FINGERPRINT + PARAMS) ----
		camDir := filepath.Join(*dbDir, *camera)
		meta, err := store.ReadMeta(camDir)
		if err != nil {
			fmt.Println("failed to read camera meta:", err)
			os.Exit(1)
		}

		// ---- TODO: ADD VERIFY PIPELINE ----
		fmt.Println("verify stub ok")
		fmt.Println("camera:", meta.CameraID)
		fmt.Println("img:", *img)
		fmt.Println("db:", camDir)
		fmt.Println("note: scoring not implemented yet")

	// ============================================================
	// ======================== HELP ================================
	// ============================================================
	case "help", "-h", "--help":
		cli.PrintUsage()

	default:
		fmt.Println("unknown command:", os.Args[1])
		fmt.Println()
		cli.PrintUsage()
		os.Exit(2)
	}
}
