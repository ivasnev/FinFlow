package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// LogLevel представляет уровень логирования
type LogLevel string

// Константы для уровней логирования
const (
	LogLevelSilent LogLevel = "silent"
	LogLevelError  LogLevel = "error"
	LogLevelWarn   LogLevel = "warn"
	LogLevelInfo   LogLevel = "info" // По умолчанию
)

type Config struct {
	Server struct {
		Port int `yaml:"port" env:"SERVER_PORT" env-default:"8083"`
	} `yaml:"server"`

	Postgres struct {
		Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
		Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
		User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DBName   string `yaml:"dbname" env:"POSTGRES_DB" env-default:"ff_id"`
	} `yaml:"postgres"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
		Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
		Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	} `yaml:"redis"`

	AuthClient struct {
		Host           string `yaml:"host" env:"AUTH_CLIENT_HOST" env-default:"localhost"`
		Port           int    `yaml:"port" env:"AUTH_CLIENT_PORT" env-default:"8084"`
		UpdateInterval int    `yaml:"update_interval" env:"UPDATE_INTERVAL" env-default:"60"`
	} `yaml:"auth"`

	FileService struct {
		BaseURL   string `yaml:"base_url" env:"FILE_SERVICE_BASE_URL" env-default:"http://localhost:8082"`
		ServiceID int    `yaml:"service_id" env:"FILE_SERVICE_ID" env-default:"2"`
	} `yaml:"file_service"`

	TVM struct {
		BaseURL       string `yaml:"base_url" env:"TVM_BASE_URL" env-default:"http://localhost:8081"`
		ServiceID     int    `yaml:"service_id" env:"TVM_SERVICE_ID" env-default:"2"`
		ServiceSecret string `yaml:"service_secret" env:"TVM_SERVICE_SECRET" env-default:"secret"`
	} `yaml:"tvm"`

	Migrations struct {
		Path string `yaml:"path" env:"MIGRATIONS_PATH" env-default:"migrations"`
	} `yaml:"migrations"`

	Logger struct {
		Level LogLevel `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	} `yaml:"logger"`
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

	cfg.Redis.Host = getEnv("REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)

	cfg.AuthClient.Host = getEnv("AUTH_CLIENT_HOST", cfg.AuthClient.Host)
	cfg.AuthClient.Port = getEnvAsInt("AUTH_CLIENT_PORT", cfg.AuthClient.Port)
	cfg.AuthClient.UpdateInterval = getEnvAsInt("UPDATE_INTERVAL", cfg.AuthClient.UpdateInterval)

	cfg.FileService.BaseURL = getEnv("FILE_SERVICE_BASE_URL", cfg.FileService.BaseURL)
	cfg.FileService.ServiceID = getEnvAsInt("FILE_SERVICE_ID", cfg.FileService.ServiceID)

	cfg.TVM.BaseURL = getEnv("TVM_BASE_URL", cfg.TVM.BaseURL)
	cfg.TVM.ServiceID = getEnvAsInt("TVM_SERVICE_ID", cfg.TVM.ServiceID)
	cfg.TVM.ServiceSecret = getEnv("TVM_SERVICE_SECRET", cfg.TVM.ServiceSecret)

	cfg.Migrations.Path = getEnv("MIGRATIONS_PATH", cfg.Migrations.Path)

	// Устанавливаем уровень логирования
	if logLevel := LogLevel(getEnv("LOG_LEVEL", string(cfg.Logger.Level))); logLevel != "" {
		cfg.Logger.Level = logLevel
	}
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
