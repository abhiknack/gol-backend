package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Load reads configuration from environment variables and config file
// Environment variables take precedence over config file values
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure viper to read from config.yaml
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Read config file if it exists (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is acceptable, we'll use env vars
	}

	// Bind environment variables
	v.SetEnvPrefix("") // No prefix, use exact names
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Manually bind environment variables for nested config
	bindEnvVariables(v)

	// Unmarshal config into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.request_timeout", "30s")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.ttl", "300s")

	// Database defaults
	v.SetDefault("database.url", "postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable")

	// Logging defaults
	v.SetDefault("logging.level", "info")
}

// bindEnvVariables manually binds environment variables to config keys
func bindEnvVariables(v *viper.Viper) {
	// Server
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.request_timeout", "REQUEST_TIMEOUT")
	v.BindEnv("server.bearer_tokens", "SERVER_BEARER_TOKENS")

	// Supabase
	v.BindEnv("supabase.url", "SUPABASE_URL")
	v.BindEnv("supabase.api_key", "SUPABASE_API_KEY")

	// Redis
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")
	v.BindEnv("redis.ttl", "REDIS_TTL")

	// Database
	v.BindEnv("database.url", "DATABASE_URL")

	// Logging
	v.BindEnv("logging.level", "LOG_LEVEL")
}

// validateConfig validates the configuration using struct tags
func validateConfig(cfg *Config) error {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range validationErrors {
				errMsgs = append(errMsgs, fmt.Sprintf(
					"field '%s' failed validation: %s",
					e.Namespace(),
					e.Tag(),
				))
			}
			return fmt.Errorf("validation errors: %s", strings.Join(errMsgs, "; "))
		}
		return err
	}

	// Additional custom validation
	if cfg.Supabase.URL == "" {
		return fmt.Errorf("SUPABASE_URL is required but not set")
	}
	if cfg.Supabase.APIKey == "" {
		return fmt.Errorf("SUPABASE_API_KEY is required but not set")
	}

	return nil
}
