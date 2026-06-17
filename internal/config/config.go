package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MySQLUser      string
	MySQLPwd       string
	MySQLHost      string
	MySQLDatabase  string
	JWTSecret      string
	FrontendOrigin string
	AIAPIKey       string
	AIModel        string
	Port           string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		MySQLUser:      getEnv("MYSQL_USER", "hobbyuser"),
		MySQLPwd:       getEnv("MYSQL_PWD", "hobbypass"),
		MySQLHost:      getEnv("MYSQL_HOST", "127.0.0.1:3306"),
		MySQLDatabase:  getEnv("MYSQL_DATABASE", "hobby_relay"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		AIAPIKey:       getEnv("AI_API_KEY", ""),
		AIModel:        getEnv("AI_MODEL", "claude-sonnet-4-6"),
		Port:           getEnv("PORT", "8080"),
	}
}

func (c *Config) DSN() string {
	host := c.MySQLHost
	// Cloud SQL unix socket
	if len(host) > 5 && host[:5] == "unix(" {
		return fmt.Sprintf("%s:%s@%s/%s?parseTime=true&charset=utf8mb4",
			c.MySQLUser, c.MySQLPwd, host, c.MySQLDatabase)
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4",
		c.MySQLUser, c.MySQLPwd, host, c.MySQLDatabase)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
