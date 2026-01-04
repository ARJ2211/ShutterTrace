package store

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestFingerprintRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	camDir := filepath.Join(tmpDir, "Samsung_1")
	if err := os.MkdirAll(camDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	orig := []float32{0.1, -0.2, 3.14, 0.0, 1.2345}
	if err := WriteFingerprint(camDir, orig); err != nil {
		t.Fatalf("WriteFingerprint: %v", err)
	}

	got, err := ReadFingerprint(camDir)
	if err != nil {
		t.Fatalf("ReadFingerprint: %v", err)
	}

	if len(got) != len(orig) {
		t.Fatalf("len mismatch: got %d want %d", len(got), len(orig))
	}

	for i := range orig {
		if math.Abs(float64(got[i]-orig[i])) > 1e-6 {
			t.Fatalf("value mismatch at %d: got %v want %v", i, got[i], orig[i])
		}
	}
}

func TestReadFingerprint_FileSizeMultipleOf4(t *testing.T) {
	tmpDir := t.TempDir()
	camDir := filepath.Join(tmpDir, "BadCam")
	if err := os.MkdirAll(camDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write 3 bytes (invalid for float32 stream)
	badPath := filepath.Join(camDir, FingerprintFile)
	if err := os.WriteFile(badPath, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	_, err := ReadFingerprint(camDir)
	if err == nil {
		t.Fatalf("expected error for non-multiple-of-4 file size")
	}
}
