package cli

import "fmt"

func PrintUsage() {
	fmt.Println("ShutterTrace: PRNU camera fingerprint experiments")
	fmt.Println()
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace enroll --camera <id> --in <folder> [--out <db>] [--sigma <float>]")
	fmt.Println("  ./ShutterTrace verify --camera <id> [--db <db>] --img <path> [--metric pearson|pce|both]")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println("  ./ShutterTrace enroll --camera iphone_14 --in ./datasets/iphone_14 --out ./db --sigma 1.0")
	fmt.Println("  ./ShutterTrace verify --camera iphone_14 --db ./db --img ./datasets/test/IMG_9112.JPG --metric pearson")
	fmt.Println("  ./ShutterTrace verify --camera iphone_14 --db ./db --img ./datasets/test/IMG_9112.JPG --metric pce")
	fmt.Println("  ./ShutterTrace verify --camera iphone_14 --db ./db --img ./datasets/test/IMG_9112.JPG --metric both")
	fmt.Println()
}

func PrintEnrollUsage() {
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace enroll --camera <id> --in <folder> [--out <db>] [--sigma <float>]")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --in       folder containing enrollment images (required)")
	fmt.Println("  --out      output db folder (default ./db)")
	fmt.Println("  --sigma    gaussian blur sigma (default 1.0)")
	fmt.Println()
}

func PrintVerifyUsage() {
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace verify --camera <id> [--db <db>] --img <path> [--metric pearson|pce|both]")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --db       db folder (default ./db)")
	fmt.Println("  --img      test image path (required)")
	fmt.Println("  --metric   scoring method: pearson, pce, or both (default pearson)")
	fmt.Println()
	fmt.Println("notes:")
	fmt.Println("  - pearson is faster")
	fmt.Println("  - pce is slower but more forensic style")
	fmt.Println()
}
