package store

import (
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
