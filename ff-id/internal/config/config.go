package config

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
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

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     int
	RefreshTTL    int
}

func LoadConfig() (*Config, error) {
	// TODO: Implement configuration loading using viper
	return &Config{
		Server: ServerConfig{
			Port: ":8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "ff_id",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			AccessSecret:  "your-access-secret",
			RefreshSecret: "your-refresh-secret",
			AccessTTL:     15,    // 15 minutes
			RefreshTTL:    10080, // 7 days
		},
	}, nil
} 