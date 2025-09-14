package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server struct {
		Port int `env:"SERVER_PORT" envDefault:"8080"`
	}
	Postgres struct {
		Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
		Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
		User     string `env:"POSTGRES_USER" envDefault:"postgres"`
		Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
		DBName   string `env:"POSTGRES_DB" envDefault:"ff_files"`
	}
	MinIO struct {
		Endpoint        string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
		AccessKeyID     string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
		SecretAccessKey string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
		BucketName      string `env:"MINIO_BUCKET" envDefault:"ff-files"`
		UseSSL          bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	}
	Redis struct {
		Host     string `env:"REDIS_HOST" envDefault:"localhost"`
		Port     int    `env:"REDIS_PORT" envDefault:"6379"`
		Password string `env:"REDIS_PASSWORD" envDefault:""`
	}
	TVM struct {
		BaseURL   string `env:"TVM_BASE_URL" envDefault:"http://localhost:8081"`
		ServiceID int    `env:"TVM_SERVICE_ID" envDefault:"1"`
	}
}

func Load() *Config {
	cfg := &Config{}

	// Загружаем конфигурацию из переменных окружения
	cfg.Server.Port, _ = strconv.Atoi(getEnvOrDefault("POSTGRES_HOST", "8082"))

	cfg.Postgres.Host = getEnvOrDefault("POSTGRES_HOST", "localhost")
	cfg.Postgres.Port, _ = strconv.Atoi(getEnvOrDefault("POSTGRES_PORT", "5432"))
	cfg.Postgres.User = getEnvOrDefault("POSTGRES_USER", "postgres")
	cfg.Postgres.Password = getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	cfg.Postgres.DBName = getEnvOrDefault("POSTGRES_DB", "ff_files")

	cfg.MinIO.Endpoint = getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000")
	cfg.MinIO.AccessKeyID = getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin")
	cfg.MinIO.SecretAccessKey = getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin")
	cfg.MinIO.BucketName = getEnvOrDefault("MINIO_BUCKET", "ff-files")
	cfg.MinIO.UseSSL = getEnvOrDefault("MINIO_USE_SSL", "false") == "true"

	cfg.Redis.Host = getEnvOrDefault("REDIS_HOST", "localhost")
	cfg.Redis.Port, _ = strconv.Atoi(getEnvOrDefault("REDIS_PORT", "6379"))
	cfg.Redis.Password = getEnvOrDefault("REDIS_PASSWORD", "")

	cfg.TVM.BaseURL = getEnvOrDefault("TVM_BASE_URL", "http://localhost:8081")
	cfg.TVM.ServiceID, _ = strconv.Atoi(getEnvOrDefault("TVM_SERVICE_ID", "1"))

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
