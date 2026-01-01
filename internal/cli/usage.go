package cli

import "fmt"

func PrintUsage() {
	fmt.Println("ShutterTrace: PRNU camera fingerprint experiments")
	fmt.Println()
	fmt.Println("usage:")
	fmt.Println("  prnu enroll --camera <id> --in <folder> [--out <db>]")
	fmt.Println("  prnu verify --camera <id> --db <db> --img <path>")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println("  prnu enroll --camera Fuji_1 --in ./data/Fuji_1/enroll --out ./db")
	fmt.Println("  prnu verify --camera Fuji_1 --db ./db --img ./data/Fuji_1/test/001.jpg")
	fmt.Println()
}

func PrintEnrollUsage() {
	fmt.Println("usage:")
	fmt.Println("  prnu enroll --camera <id> --in <folder> [--out <db>]")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --in       folder containing enrollment images (required)")
	fmt.Println("  --out      output db folder (default ./db)")
	fmt.Println()
}

func PrintVerifyUsage() {
	fmt.Println("usage:")
	fmt.Println("  prnu verify --camera <id> --db <db> --img <path>")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --camera   camera id/name (required)")
	fmt.Println("  --db       db folder (default ./db)")
	fmt.Println("  --img      test image path (required)")
	fmt.Println()
}
