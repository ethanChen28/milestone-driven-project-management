package app

import "os"

type Config struct {
	ListenAddr string
	MySQLDSN   string
	RedisAddr  string
	DefaultLng string
}

func LoadConfig() Config {
	return Config{
		ListenAddr: getenv("APP_LISTEN_ADDR", ":8080"),
		MySQLDSN:   getenv("MYSQL_DSN", "goal:goal@tcp(mysql:3306)/goal_manager?parseTime=true"),
		RedisAddr:  getenv("REDIS_ADDR", "redis:6379"),
		DefaultLng: getenv("APP_DEFAULT_LOCALE", "zh-CN"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
