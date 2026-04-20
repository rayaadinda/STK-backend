package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            string
	CORSAllowOrigin string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	TimeZone        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "stk-backend"),
			Env:             getEnv("APP_ENV", "development"),
			Port:            getEnv("APP_PORT", "8080"),
			CORSAllowOrigin: getEnv("APP_CORS_ALLOW_ORIGIN", "http://localhost:3000"),
			ReadTimeout:     time.Duration(getEnvAsInt("APP_READ_TIMEOUT_SECONDS", 15)) * time.Second,
			WriteTimeout:    time.Duration(getEnvAsInt("APP_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
			ShutdownTimeout: time.Duration(getEnvAsInt("APP_SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "stk_backend"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			TimeZone:        getEnv("DB_TIMEZONE", "UTC"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 50),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	missing := make([]string, 0)

	requiredFields := map[string]string{
		"DB_HOST": c.Database.Host,
		"DB_PORT": c.Database.Port,
		"DB_USER": c.Database.User,
		"DB_NAME": c.Database.Name,
	}

	for key, value := range requiredFields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if c.Database.MaxIdleConns < 0 || c.Database.MaxOpenConns < 1 {
		return errors.New("invalid database connection pool configuration")
	}

	if c.App.ReadTimeout <= 0 || c.App.WriteTimeout <= 0 || c.App.ShutdownTimeout <= 0 {
		return errors.New("application timeout values must be greater than zero")
	}

	return nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host,
		d.User,
		d.Password,
		d.Name,
		d.Port,
		d.SSLMode,
		d.TimeZone,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}

	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}

	return valueInt
}
