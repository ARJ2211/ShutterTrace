package cli

import "fmt"

func PrintUsage() {
	fmt.Println("ShutterTrace: PRNU camera fingerprint experiments")
	fmt.Println()
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace enroll --camera <id> --in <folder> [--out <db>] --sigma 1.0")
	fmt.Println("  ./ShutterTrace verify --camera <id> --db <db> --img <path>")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println("  ./ShutterTrace enroll --camera Fuji_1 --in ./data/Fuji_1/enroll --out ./db")
	fmt.Println("  ./ShutterTrace verify --camera Fuji_1 --db ./db --img ./data/Fuji_1/test/001.jpg")
	fmt.Println()
}

func PrintEnrollUsage() {
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace enroll --camera <id> --in <folder> [--out <db>]")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --in       folder containing enrollment images (required)")
	fmt.Println("  --out      output db folder (default ./db)")
	fmt.Println("  --sigma    sigma value for gaussian kernal")
	fmt.Println()
}

func PrintVerifyUsage() {
	fmt.Println("usage:")
	fmt.Println("  ./ShutterTrace verify --camera <id> --db <db> --img <path>")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --db       db folder (default ./db)")
	fmt.Println("  --img      test image path (required)")
	fmt.Println()
}
