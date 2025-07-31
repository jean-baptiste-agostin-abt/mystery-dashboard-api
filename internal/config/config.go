package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	// Server configuration
	Port         int    `mapstructure:"PORT"`
	Environment  string `mapstructure:"ENVIRONMENT"`
	ServiceName  string `mapstructure:"SERVICE_NAME"`
	ReadTimeout  int    `mapstructure:"READ_TIMEOUT"`
	WriteTimeout int    `mapstructure:"WRITE_TIMEOUT"`
	IdleTimeout  int    `mapstructure:"IDLE_TIMEOUT"`

	// CORS configuration
	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	// Database configuration
	DatabaseDSN string `mapstructure:"DATABASE_DSN"`

	// JWT configuration
	JWTSecret     string `mapstructure:"JWT_SECRET"`
	JWTExpiration int    `mapstructure:"JWT_EXPIRATION"`

	// Logging configuration
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// OpenTelemetry configuration
	JaegerEndpoint string `mapstructure:"JAEGER_ENDPOINT"`

	// AWS configuration
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	S3Bucket           string `mapstructure:"S3_BUCKET"`

	// Multi-tenant configuration
	DefaultTenantID string `mapstructure:"DEFAULT_TENANT_ID"`
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/mysteryfactory")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; ignore error if desired
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVICE_NAME", "mysteryfactory-api")
	viper.SetDefault("READ_TIMEOUT", 30)
	viper.SetDefault("WRITE_TIMEOUT", 30)
	viper.SetDefault("IDLE_TIMEOUT", 120)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("JWT_EXPIRATION", 3600) // 1 hour in seconds
	viper.SetDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	viper.SetDefault("AWS_REGION", "us-east-1")
	viper.SetDefault("DEFAULT_TENANT_ID", "default")
}

// validate checks that required configuration values are present
func validate(config *Config) error {
	required := map[string]string{
		"DATABASE_DSN": config.DatabaseDSN,
		"JWT_SECRET":   config.JWTSecret,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	isValidEnv := false
	for _, env := range validEnvs {
		if config.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return fmt.Errorf("invalid environment: %s (must be one of: %s)",
			config.Environment, strings.Join(validEnvs, ", "))
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if config.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			config.LogLevel, strings.Join(validLogLevels, ", "))
	}

	return nil
}
