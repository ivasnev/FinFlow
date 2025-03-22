package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	} `yaml:"server"`

	Database struct {
		URL string `yaml:"url" env:"DATABASE_URL" env-default:"postgres://postgres:postgres@localhost:5432/finflow?sslmode=disable"`
	} `yaml:"database"`

	Dev struct {
		Enabled bool   `yaml:"enabled" env:"DEV_MODE" env-default:"false"`
		Secret  string `yaml:"secret" env:"DEV_SECRET" env-default:"dev_secret_key"`
	} `yaml:"dev"`

	Migrations struct {
		Path string `yaml:"path" env:"MIGRATIONS_PATH" env-default:"migrations"`
	} `yaml:"migrations"`
}

func Load() *Config {
	cfg := &Config{}

	// Загружаем конфигурацию из YAML файла, если он существует
	if err := loadFromYAML(cfg); err != nil {
		fmt.Printf("Warning: failed to load config from YAML: %v\n", err)
	}

	// Загружаем конфигурацию из переменных окружения
	if err := loadFromEnv(cfg); err != nil {
		fmt.Printf("Warning: failed to load config from environment: %v\n", err)
	}

	return cfg
}

func loadFromYAML(cfg *Config) error {
	// Ищем конфигурационный файл в текущей директории и родительских директориях
	configPaths := []string{
		"config.yaml",
		"config.yml",
		"config/config.yaml",
		"config/config.yml",
	}

	var configData []byte
	var err error

	for _, path := range configPaths {
		configData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("no config file found: %w", err)
	}

	if err := yaml.Unmarshal(configData, cfg); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

func loadFromEnv(cfg *Config) error {
	// Здесь можно добавить логику загрузки из переменных окружения
	// Например, используя библиотеку envconfig или подобную
	// Для простоты оставим базовую реализацию

	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.Database.URL = dbURL
	}

	if devMode := os.Getenv("DEV_MODE"); devMode != "" {
		cfg.Dev.Enabled = devMode == "true"
	}

	if devSecret := os.Getenv("DEV_SECRET"); devSecret != "" {
		cfg.Dev.Secret = devSecret
	}

	return nil
}
