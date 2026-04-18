package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/mydb")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("JWT_EXPIRATION_HOURS", "24")
	t.Setenv("PORT", "8080")
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/mydb")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("JWT_EXPIRATION_HOURS", "48")
	t.Setenv("PORT", "9090")

	cfg := Load()

	assert.Equal(t, "postgres://user:pass@db:5432/mydb", cfg.DatabaseURL)
	assert.Equal(t, "supersecret", cfg.JWTSecret)
	assert.Equal(t, 48*time.Hour, cfg.JWTExpiration)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	validEnv(t)
	require.NoError(t, os.Unsetenv("DATABASE_URL"))
	assert.PanicsWithValue(t, "invalid config: DATABASE_URL is required", func() { Load() })
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	validEnv(t)
	require.NoError(t, os.Unsetenv("JWT_SECRET"))
	assert.PanicsWithValue(t, "invalid config: JWT_SECRET is required", func() { Load() })
}

func TestLoad_InvalidJWTExpiration(t *testing.T) {
	validEnv(t)
	t.Setenv("JWT_EXPIRATION_HOURS", "not-a-number")
	assert.PanicsWithValue(t, "invalid config: JWT_EXPIRATION_HOURS is required and must be > 0", func() { Load() })
}

func TestLoad_ZeroJWTExpiration(t *testing.T) {
	validEnv(t)
	t.Setenv("JWT_EXPIRATION_HOURS", "0")
	assert.PanicsWithValue(t, "invalid config: JWT_EXPIRATION_HOURS is required and must be > 0", func() { Load() })
}

func TestLoad_MissingPort(t *testing.T) {
	validEnv(t)
	require.NoError(t, os.Unsetenv("PORT"))
	assert.PanicsWithValue(t, "invalid config: PORT is required", func() { Load() })
}
