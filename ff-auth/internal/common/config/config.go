package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
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
		DBName   string `yaml:"dbname" env:"POSTGRES_DB" env-default:"ff_auth"`
	} `yaml:"postgres"`

	Auth struct {
		JWTSecret            string `yaml:"jwt_secret" env:"JWT_SECRET" env-default:"default_jwt_secret"`
		AccessTokenDuration  int    `yaml:"access_token_duration" env:"ACCESS_TOKEN_DURATION" env-default:"15"`      // в минутах
		RefreshTokenDuration int    `yaml:"refresh_token_duration" env:"REFRESH_TOKEN_DURATION" env-default:"10080"` // в минутах (по умолчанию 7 дней)
		PasswordMinLength    int    `yaml:"password_min_length" env:"PASSWORD_MIN_LENGTH" env-default:"8"`
		PasswordHashCost     int    `yaml:"password_hash_cost" env:"PASSWORD_HASH_COST" env-default:"10"`
	} `yaml:"auth"`

	IDClient struct {
		BaseURL string `yaml:"base_url" env:"ID_BASE_URL" env-default:"http://localhost:8083"`
		TVMID   int    `yaml:"tvm_id" env:"ID_TVM_ID" env-default:"4"`
	} `yaml:"id_client"`

	TVM struct {
		BaseURL       string `yaml:"base_url" env:"TVM_BASE_URL" env-default:"http://localhost:8081"`
		ServiceID     int    `yaml:"service_id" env:"TVM_SERVICE_ID" env-default:"2"`
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

	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Auth.AccessTokenDuration = getEnvAsInt("ACCESS_TOKEN_DURATION", cfg.Auth.AccessTokenDuration)
	cfg.Auth.RefreshTokenDuration = getEnvAsInt("REFRESH_TOKEN_DURATION", cfg.Auth.RefreshTokenDuration)
	cfg.Auth.PasswordMinLength = getEnvAsInt("PASSWORD_MIN_LENGTH", cfg.Auth.PasswordMinLength)
	cfg.Auth.PasswordHashCost = getEnvAsInt("PASSWORD_HASH_COST", cfg.Auth.PasswordHashCost)

	cfg.IDClient.BaseURL = getEnv("ID_BASE_URL", cfg.IDClient.BaseURL)
	cfg.IDClient.TVMID = getEnvAsInt("ID_TVM_ID", cfg.IDClient.TVMID)

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
