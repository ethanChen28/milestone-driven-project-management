package app

import "os"

type Config struct {
	ListenAddr     string
	MySQLDSN       string
	RedisAddr      string
	DefaultLng     string
	StorageBackend string
	AppEnv         string
	AuthMode       string
	TokenSecret    string
	UserServiceURL string
}

func LoadConfig() Config {
	return Config{
		ListenAddr:     getenv("APP_LISTEN_ADDR", ":8080"),
		MySQLDSN:       getenv("MYSQL_DSN", "goal:goal@tcp(mysql:3306)/goal_manager?parseTime=true"),
		RedisAddr:      getenv("REDIS_ADDR", "redis:6379"),
		DefaultLng:     getenv("APP_DEFAULT_LOCALE", "zh-CN"),
		StorageBackend: getenv("STORAGE_BACKEND", "memory"),
		AppEnv:         getenv("APP_ENV", "development"),
		AuthMode:       getenv("AUTH_MODE", "dev-header"),
		TokenSecret:    getenv("AUTH_TOKEN_SECRET", "dev-secret-change-me"),
		UserServiceURL: getenv("USER_SERVICE_URL", ""),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
