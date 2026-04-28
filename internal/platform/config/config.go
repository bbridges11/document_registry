package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type Profile string

const (
	ProfileLocal      Profile = "local"
	ProfileCloudLocal Profile = "cloud-local"
	ProfileCloudDev   Profile = "cloud-dev"
	ProfileCloudProd  Profile = "cloud-prod"
)

type ServerDriver string

const (
	ServerDriverHTTP ServerDriver = "http"
	ServerDriverGRPC ServerDriver = "grpc"
)

type DatabaseDriver string

const (
	DatabaseDriverPostgres DatabaseDriver = "postgres"
)

type Config struct {
	App      AppConfig `envPrefix:"APP"`
	Logging  LoggingConfig
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
}

type AppConfig struct {
	Name    string  `env:"NAME" envDefault:"document-registry"`
	Version string  `env:"VERSION" envDefault:"0.1.0"`
	Profile Profile `env:"PROFILE" envDefault:"local"`
}

type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type ServerConfig struct {
	Driver ServerDriver `env:"SERVER_DRIVER" envDefault:"http"`
	HTTP   HTTPConfig
	GRPC   GRPCConfig
}

type HTTPConfig struct {
	Host string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" envDefault:"8080"`
}

type GRPCConfig struct {
	Host string `env:"GRPC_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"GRPC_PORT" envDefault:"9090"`
}

type DatabaseConfig struct {
	Driver   DatabaseDriver `env:"DB_DRIVER" envDefault:"postgres"`
	Postgres PostgresConfig `envPrefix:"POSTGRES_"`
}

type PostgresConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"6160"`
	Database string `env:"DATABASE" envDefault:"document_registry"`
	User     string `env:"USER" envDefault:"postgres"`
	Password string `env:"PASSWORD" envDefault:"postgres"` // Only for local profile
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`  // Only for local profile
	MaxConns int    `env:"MAX_CONNS" envDefault:"25"`
}

// Validate validates postgres config based on profile
func (c PostgresConfig) Validate(profile Profile) error {
	if profile == ProfileLocal {
		// Local requires password
		if c.Password == "" {
			return fmt.Errorf("POSTGRES_PASSWORD is required for local profile")
		}
	} else {
		// Cloud profiles must NOT have password (ignore default value)
		if c.Password != "" && c.Password != "postgres" {
			return fmt.Errorf("POSTGRES_PASSWORD must not be set for cloud profiles (use IAM auth)")
		}
	}
	return nil
}

// IsCloudProfile returns true if profile requires IAM auth
func (c PostgresConfig) IsCloudProfile(profile Profile) bool {
	return profile != ProfileLocal
}

// DSN generates connection string for local profile only
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

type AWSConfig struct {
	Region string    `env:"AWS_REGION" envDefault:"us-east-1"`
	S3     S3Config  `envPrefix:"S3_"`
	SNS    SNSConfig `envPrefix:"SNS_"`
}

type S3Config struct {
	Bucket   string `env:"BUCKET" envDefault:"document-registry"`
	Endpoint string `env:"ENDPOINT"`
}

// SNSConfig holds AWS SNS configuration
type SNSConfig struct {
	// TopicARN is the ARN of the SNS topic to publish notifications to
	TopicARN string `env:"TOPIC_ARN"`

	// Enabled controls whether SNS is enabled or in bypass mode
	// When false, notifications are logged but not sent
	Enabled bool `env:"ENABLED" envDefault:"false"`

	// Endpoint is the SNS endpoint URL (for LocalStack)
	// Leave empty for AWS SNS
	Endpoint string `env:"ENDPOINT"`
}

func Load() (cfg *Config, err error) {
	cfg = &Config{}

	// Parse all config with DOCUMENT_REGISTRY_ prefix
	if err = env.ParseWithOptions(cfg, env.Options{
		Prefix: "DOCUMENT_REGISTRY_",
	}); err != nil {
		return nil, err
	}

	// Fallback to standard AWS_REGION if prefixed version not set
	// This supports standard AWS SDK environment variable
	if cfg.AWS.Region == "" {
		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg.AWS.Region = region
		}
	}

	// Apply profile-specific defaults
	cfg.applyProfile()

	// Validate database config based on profile
	if err = cfg.Database.Postgres.Validate(cfg.App.Profile); err != nil {
		return nil, fmt.Errorf("postgres config validation failed: %w", err)
	}

	// Validate AWS region for cloud profiles
	if cfg.Database.Postgres.IsCloudProfile(cfg.App.Profile) {
		if cfg.AWS.Region == "" {
			return nil, fmt.Errorf("AWS_REGION or DOCUMENT_REGISTRY_AWS_REGION is required for cloud profiles")
		}
	}

	return cfg, nil
}

func (c *Config) applyProfile() {
	switch c.App.Profile {
	case ProfileLocal:
		c.applyLocalProfile()
	case ProfileCloudLocal:
		c.applyCloudLocalProfile()
	case ProfileCloudDev:
		c.applyCloudDevProfile()
	case ProfileCloudProd:
		c.applyCloudProdProfile()
	}
}

func (c *Config) applyLocalProfile() {
	if c.Logging.Level == "" {
		c.Logging.Level = "debug"
	}
}

func (c *Config) applyCloudLocalProfile() {
	if c.Logging.Level == "" {
		c.Logging.Level = "debug"
	}
	// Use local endpoints for AWS services
	if c.AWS.S3.Endpoint == "" {
		c.AWS.S3.Endpoint = "http://localhost:4566"
	}
	if c.AWS.SNS.Endpoint == "" {
		c.AWS.SNS.Endpoint = "http://localhost:4566"
	}
}

func (c *Config) applyCloudDevProfile() {
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
}

func (c *Config) applyCloudProdProfile() {
	if c.Logging.Level == "" {
		c.Logging.Level = "warn"
	}
}
