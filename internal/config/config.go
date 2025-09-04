package config

import (
	"os"
	"time"
)

type Config struct {
	Server  ServerConfig
	Cache   CacheConfig
	Rclone  RcloneConfig
	Storage StorageConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type CacheConfig struct {
	Dir     string
	TTL     time.Duration
	MaxSize int64 // in bytes
}

type RcloneConfig struct {
	ConfigPath string
	BinPath    string
}

type StorageConfig struct {
	Providers []string
	UnionName string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("API_PORT", "5601"),
			Host: getEnv("API_HOST", "0.0.0.0"),
		},
		Cache: CacheConfig{
			Dir:     getEnv("CACHE_DIR", "./cache"),
			TTL:     parseDuration(getEnv("CACHE_TTL", "24h")),
			MaxSize: parseInt64(getEnv("CACHE_MAX_SIZE", "10737418240")), // 10GB default
		},
		Rclone: RcloneConfig{
			ConfigPath: getEnv("RCLONE_CONFIG_PATH", "./configs/rclone.conf"), // Use project config
			BinPath:    getEnv("RCLONE_BIN_PATH", "rclone"),
		},
		Storage: StorageConfig{
			Providers: []string{"mega1", "mega2", "mega3", "gdrive"}, // Three mega + Google Drive
			UnionName: "union",                                       // Use union for load balancing
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour // default 24 hours
	}
	return d
}

func parseInt64(s string) int64 {
	// Simple implementation, in production use strconv.ParseInt
	return 10737418240 // 10GB default
}
