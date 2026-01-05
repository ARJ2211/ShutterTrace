package store

type Meta struct {
	Version    int     `json:"version"`
	CameraID   string  `json:"camera_id"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	ColorMode  string  `json:"color_mode"`
	DenoiseAlg string  `json:"denoise_alg"`
	Notes      string  `json:"notes"`
	Sigma      float32 `json:"sigma"`
}
