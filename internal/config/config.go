package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OMDB     OMDBConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpiryHours int
}

type OMDBConfig struct {
	APIKey  string
	BaseURL string
}

type CacheConfig struct {
	SearchTTL time.Duration
	MovieTTL  time.Duration
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// RedisAddr returns host:port for Redis.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Not fatal — env vars may be set directly (e.g. Docker)
		fmt.Printf("Warning: .env file not found, relying on environment variables\n")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:    getStringOrDefault("SERVER_PORT", "8080"),
			GinMode: getStringOrDefault("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getStringOrDefault("DB_HOST", "localhost"),
			Port:     getStringOrDefault("DB_PORT", "5432"),
			User:     getStringOrDefault("DB_USER", "postgres"),
			Password: getStringOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getStringOrDefault("DB_NAME", "movie_recommend"),
			SSLMode:  getStringOrDefault("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getStringOrDefault("REDIS_HOST", "localhost"),
			Port:     getStringOrDefault("REDIS_PORT", "6379"),
			Password: getStringOrDefault("REDIS_PASSWORD", ""),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:     getStringOrDefault("JWT_SECRET", "default-secret-change-me"),
			ExpiryHours: getIntOrDefault("JWT_EXPIRY_HOURS", 24),
		},
		OMDB: OMDBConfig{
			APIKey:  viper.GetString("OMDB_API_KEY"),
			BaseURL: getStringOrDefault("OMDB_BASE_URL", "http://www.omdbapi.com"),
		},
		Cache: CacheConfig{
			SearchTTL: time.Duration(getIntOrDefault("CACHE_SEARCH_TTL", 86400)) * time.Second,
			MovieTTL:  time.Duration(getIntOrDefault("CACHE_MOVIE_TTL", 604800)) * time.Second,
		},
	}

	return cfg, nil
}

func getStringOrDefault(key, defaultVal string) string {
	val := viper.GetString(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getIntOrDefault(key string, defaultVal int) int {
	val := viper.GetInt(key)
	if val == 0 {
		return defaultVal
	}
	return val
}
