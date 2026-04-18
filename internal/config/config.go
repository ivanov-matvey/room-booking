package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	JWTExpiration time.Duration
	Port          string

	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime time.Duration
	DBMaxConnIdleTime time.Duration
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiration: getDurationHours("JWT_EXPIRATION_HOURS"),
		Port:          os.Getenv("PORT"),

		DBMaxConns:        getInt32("DB_MAX_CONNS", 20),
		DBMinConns:        getInt32("DB_MIN_CONNS", 5),
		DBMaxConnLifetime: getDurationMinutes("DB_MAX_CONN_LIFETIME_MINUTES", 60),
		DBMaxConnIdleTime: getDurationMinutes("DB_MAX_CONN_IDLE_TIME_MINUTES", 10),
	}
	if err := cfg.validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %s", err))
	}
	return cfg
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.JWTExpiration == 0 {
		return fmt.Errorf("JWT_EXPIRATION_HOURS is required and must be > 0")
	}
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	return nil
}

func getDurationHours(key string) time.Duration {
	if v := os.Getenv(key); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			return time.Duration(h) * time.Hour
		}
	}
	return 0
}

func getDurationMinutes(key string, defaultVal int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m > 0 {
			return time.Duration(m) * time.Minute
		}
	}
	return time.Duration(defaultVal) * time.Minute
}

func getInt32(key string, defaultVal int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			return int32(n)
		}
	}
	return defaultVal
}
