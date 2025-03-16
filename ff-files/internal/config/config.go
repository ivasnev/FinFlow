package config

import "time"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	TVM      TVMConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type StorageConfig struct {
	BasePath           string        // Базовый путь для хранения файлов
	MaxFileSize        int64         // Максимальный размер файла в байтах
	AllowedMimeTypes   []string      // Разрешенные MIME-типы
	TempURLExpiration  time.Duration // Время жизни временных ссылок
	SoftDeleteTimeout  time.Duration // Время до полного удаления файла
}

type TVMConfig struct {
	BaseURL    string // URL сервиса TVM
	ServiceID  string // Идентификатор сервиса
	ServiceKey string // Ключ сервиса
}

func LoadConfig() (*Config, error) {
	// TODO: Implement configuration loading using viper
	return &Config{
		Server: ServerConfig{
			Port: ":8082",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "ff_files",
		},
		Storage: StorageConfig{
			BasePath:          "./storage",
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			AllowedMimeTypes: []string{
				"image/jpeg",
				"image/png",
				"image/gif",
				"application/pdf",
				"text/plain",
			},
			TempURLExpiration: 24 * time.Hour,    // Временные ссылки живут 24 часа
			SoftDeleteTimeout: 30 * 24 * time.Hour, // 30 дней до полного удаления
		},
		TVM: TVMConfig{
			BaseURL:    "http://localhost:8083",
			ServiceID:  "ff-files",
			ServiceKey: "your-service-key",
		},
	}, nil
} 