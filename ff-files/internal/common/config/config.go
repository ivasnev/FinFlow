package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
)

type MinIO struct {
	Endpoint           string `yaml:"endpoint" env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	InternalEndpoint   string `yaml:"internal_endpoint" env:"MINIO_INTERNAL_ENDPOINT" env-default:"localhost:9000"`
	AccessKeyID        string `yaml:"access_key" env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`
	SecretAccessKey    string `yaml:"secret_key" env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
	BucketName         string `yaml:"bucket" env:"MINIO_BUCKET" env-default:"ff-files"`
	UseSSL             bool   `yaml:"use_ssl" env:"MINIO_USE_SSL" env-default:"false"`
	RootUser           string `yaml:"root_user" env:"MINIO_ROOT_USER" env-default:"root"`
	RootPassword       string `yaml:"root_password" env:"MINIO_ROOT_PASSWORD" env-default:"minio_password"`
	FileTimeExpiration int    `yaml:"file_time_expiration" env:"MINIO_FILE_TIME_EXPIRATION" env-default:"300"`
}

type Config struct {
	Server struct {
		Port int `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	} `yaml:"server"`

	Postgres struct {
		Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
		Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
		User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DBName   string `yaml:"dbname" env:"POSTGRES_DB" env-default:"ff_files"`
	} `yaml:"postgres"`

	MinIO `yaml:"minio"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
		Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
		Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	} `yaml:"redis"`

	TVM struct {
		BaseURL       string `yaml:"base_url" env:"TVM_BASE_URL" env-default:"http://localhost:8081"`
		ServiceID     int    `yaml:"service_id" env:"TVM_SERVICE_ID" env-default:"1"`
		ServiceSecret string `yaml:"service_secret" env:"TVM_SERVICE_SECRET" env-default:"secret"`
	} `yaml:"tvm"`

	Migrations struct {
		Path string `yaml:"path" env:"MIGRATIONS_PATH" env-default:"migrations"`
	} `yaml:"migrations"`
}

func Load() *Config {
	cfg := &Config{}

	if err := loadFromYAML(cfg); err != nil {
		fmt.Printf("⚠️ Warning: failed to load config from YAML: %v\n", err)
	}

	loadFromEnv(cfg)

	return cfg
}

func loadFromYAML(cfg *Config) error {
	// Получаем текущую рабочую директорию
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Строим путь к файлу конфигурации относительно рабочей директории
	configPath := filepath.Join(wd, "config", "config.yaml")

	// Открываем файл
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no config file found at %s", configPath)
}

func loadFromEnv(cfg *Config) {
	cfg.Server.Port = getEnvAsInt("SERVER_PORT", cfg.Server.Port)

	cfg.Postgres.Host = getEnv("POSTGRES_HOST", cfg.Postgres.Host)
	cfg.Postgres.Port = getEnvAsInt("POSTGRES_PORT", cfg.Postgres.Port)
	cfg.Postgres.User = getEnv("POSTGRES_USER", cfg.Postgres.User)
	cfg.Postgres.Password = getEnv("POSTGRES_PASSWORD", cfg.Postgres.Password)
	cfg.Postgres.DBName = getEnv("POSTGRES_DB", cfg.Postgres.DBName)

	cfg.MinIO.Endpoint = getEnv("MINIO_ENDPOINT", cfg.MinIO.Endpoint)
	cfg.MinIO.InternalEndpoint = getEnv("MINIO_INTERNAL_ENDPOINT", cfg.MinIO.InternalEndpoint)
	cfg.MinIO.AccessKeyID = getEnv("MINIO_ACCESS_KEY", cfg.MinIO.AccessKeyID)
	cfg.MinIO.SecretAccessKey = getEnv("MINIO_SECRET_KEY", cfg.MinIO.SecretAccessKey)
	cfg.MinIO.BucketName = getEnv("MINIO_BUCKET", cfg.MinIO.BucketName)
	cfg.MinIO.UseSSL = getEnvAsBool("MINIO_USE_SSL", cfg.MinIO.UseSSL)
	cfg.MinIO.RootUser = getEnv("MINIO_ROOT_USER", cfg.MinIO.RootUser)
	cfg.MinIO.RootPassword = getEnv("MINIO_ROOT_PASSWORD", cfg.MinIO.RootPassword)
	cfg.MinIO.FileTimeExpiration = getEnvAsInt("FILE_TIME_EXPIRATION", cfg.MinIO.FileTimeExpiration)

	cfg.Redis.Host = getEnv("REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)

	cfg.TVM.BaseURL = getEnv("TVM_BASE_URL", cfg.TVM.BaseURL)
	cfg.TVM.ServiceID = getEnvAsInt("TVM_SERVICE_ID", cfg.TVM.ServiceID)
	cfg.TVM.ServiceSecret = getEnv("TVM_SERVICE_SECRET", cfg.TVM.ServiceSecret)

	cfg.Migrations.Path = getEnv("MIGRATIONS_PATH", cfg.Migrations.Path)
}

// getEnv получает строковую переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения в формате int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool получает переменную окружения в формате bool или возвращает значение по умолчанию
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
