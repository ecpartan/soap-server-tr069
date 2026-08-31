package fileserver

type Config struct {
	UploadPath   string
	MaxFileSize  int64
	AllowedTypes []string
}

func LoadConfig() *Config {
	return &Config{
		UploadPath:   "./uploads",
		MaxFileSize:  10 << 20, // 10 MB
		AllowedTypes: []string{"image/jpeg", "image/png", "application/pdf"},
	}
}
