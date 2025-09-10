package config

import "time"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type TVMConfig struct {
	KeyRotationInterval time.Duration // Интервал ротации ключей
	TicketTTL          time.Duration // Время жизни тикетов
	RSAKeyBits         int           // Размер RSA ключа в битах
}

func LoadConfig() (*Config, error) {
	// TODO: Implement configuration loading using viper
	return &Config{
		Server: ServerConfig{
			Port: ":8081",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "ff_tvm",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		TVM: TVMConfig{
			KeyRotationInterval: 24 * time.Hour,    // Ротация ключей раз в сутки
			TicketTTL:          1 * time.Hour,      // Тикеты живут 1 час
			RSAKeyBits:         2048,               // 2048-битные ключи
		},
	}, nil
} 