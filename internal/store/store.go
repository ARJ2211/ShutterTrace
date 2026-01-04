package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MetaFile        = "meta.json"
	FingerprintFile = "fingerprint.bin"
)

func WriteMeta(cameraDir string, meta Meta) error {
	path := filepath.Join(cameraDir, MetaFile)
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func ReadMeta(cameraDir string) (Meta, error) {
	path := filepath.Join(cameraDir, MetaFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, err
	}
	var m Meta
	if err := json.Unmarshal(b, &m); err != nil {
		return Meta{}, err
	}
	if m.CameraID == "" {
		return Meta{}, fmt.Errorf("invalid meta.json: missing camera_id")
	}
	return m, nil
}

/*
This function is required to write the fingerprint
as a raw little-endian float32 value to fingerprint.bin
*/
func WriteFingerprint(cameraDir string, fp []float32) error {
	if len(cameraDir) == 0 {
		return fmt.Errorf("camera directory not provided")
	}
	if len(fp) == 0 {
		return fmt.Errorf("no fingerprint provided")
	}

	path := filepath.Join(cameraDir, FingerprintFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writeErr := binary.Write(f, binary.LittleEndian, fp)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

/*
This function is required to read the fingerprint.bin
file from within the camera directory ASSUMING that
the finerprint.bin exists else return with some error.
*/
func ReadFingerprint(cameraDir string) ([]float32, error) {
	path := filepath.Join(cameraDir, FingerprintFile)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stats.Size()
	if size <= 0 {
		return nil, fmt.Errorf("fingerprint file is empty")
	}
	if size%4 != 0 {
		return nil, fmt.Errorf(
			"incorrect bytes returned, not a multiple of 4")
	}

	out := make([]float32, size/4)
	readErr := binary.Read(f, binary.LittleEndian, out)
	if readErr != nil {
		return nil, readErr
	}

	return out, nil
}
