package config

import (
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Supabase SupabaseConfig `mapstructure:"supabase"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port           string        `mapstructure:"port" validate:"required"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" validate:"required"`
	RequestTimeout time.Duration `mapstructure:"request_timeout" validate:"required"`
	BearerTokens   []string      `mapstructure:"bearer_tokens"` // Valid bearer tokens for API authentication
}

// SupabaseConfig holds Supabase connection configuration
type SupabaseConfig struct {
	URL    string `mapstructure:"url" validate:"required,url"`
	APIKey string `mapstructure:"api_key" validate:"required"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string        `mapstructure:"host" validate:"required"`
	Port     string        `mapstructure:"port" validate:"required"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl" validate:"required"`
}

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	URL string `mapstructure:"url" validate:"required"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
}
