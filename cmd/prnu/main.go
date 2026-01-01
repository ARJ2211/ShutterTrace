package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ARJ2211/ShutterTrace/internal/cli"
	"github.com/ARJ2211/ShutterTrace/internal/store"
)

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

		meta := store.Meta{
			Version:    1,
			CameraID:   *camera,
			Width:      0,
			Height:     0,
			ColorMode:  "grayscale",
			DenoiseAlg: "gaussian",
			Notes:      "fingerprint not yet computed (skeleton)",
		}
		if err := store.WriteMeta(camDir, meta); err != nil {
			fmt.Println("failed to write meta:", err)
			os.Exit(1)
		}

		fmt.Println("enroll ok")
		fmt.Println("camera:", *camera)
		fmt.Println("input:", *inDir)
		fmt.Println("db:", camDir)

	case "verify":
		verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
		camera := verifyCmd.String("camera", "", "camera id/name (required)")
		dbDir := verifyCmd.String("db", "./db", "db folder (default ./db)")
		img := verifyCmd.String("img", "", "test image path (required)")
		verifyCmd.Parse(os.Args[2:])

		if *camera == "" || *img == "" {
			fmt.Println("missing required flags: --camera and --img")
			fmt.Println()
			cli.PrintVerifyUsage()
			os.Exit(2)
		}

		camDir := filepath.Join(*dbDir, *camera)
		meta, err := store.ReadMeta(camDir)
		if err != nil {
			fmt.Println("failed to read camera meta:", err)
			os.Exit(1)
		}

		fmt.Println("verify stub ok")
		fmt.Println("camera:", meta.CameraID)
		fmt.Println("img:", *img)
		fmt.Println("db:", camDir)
		fmt.Println("note: scoring not implemented yet")

	case "help", "-h", "--help":
		cli.PrintUsage()

	default:
		fmt.Println("unknown command:", os.Args[1])
		fmt.Println()
		cli.PrintUsage()
		os.Exit(2)
	}
}
