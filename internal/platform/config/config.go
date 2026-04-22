package config

import (
	"fmt"

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
	App      AppConfig
	Logging  LoggingConfig
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	OpenFGA  OpenFGAConfig
	User     UserServiceConfig
}

type AppConfig struct {
	Name    string  `env:"APP_NAME" envDefault:"document-registry"`
	Version string  `env:"APP_VERSION" envDefault:"0.1.0"`
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
	Postgres PostgresConfig
	Dynamo   DynamoConfig
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	Database string `env:"POSTGRES_DATABASE" envDefault:"document_registry"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	SSLMode  string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
	MaxConns int    `env:"POSTGRES_MAX_CONNS" envDefault:"25"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

type DynamoConfig struct {
	Region   string `env:"DYNAMO_REGION" envDefault:"us-east-1"`
	Endpoint string `env:"DYNAMO_ENDPOINT"`
}

type AWSConfig struct {
	Region string `env:"AWS_REGION" envDefault:"us-east-1"`
	S3     S3Config
	SNS    SNSConfig
}

type S3Config struct {
	Bucket   string `env:"S3_BUCKET" envDefault:"document-registry"`
	Endpoint string `env:"S3_ENDPOINT"`
}

// SNSConfig holds AWS SNS configuration
type SNSConfig struct {
	// TopicARN is the ARN of the SNS topic to publish notifications to
	TopicARN string `env:"SNS_TOPIC_ARN"`

	// Enabled controls whether SNS is enabled or in bypass mode
	// When false, notifications are logged but not sent
	Enabled bool `env:"SNS_ENABLED" envDefault:"false"`

	// Endpoint is the SNS endpoint URL (for LocalStack)
	// Leave empty for AWS SNS
	Endpoint string `env:"SNS_ENDPOINT"`
}

type OpenFGAConfig struct {
	Enabled    bool   `env:"OPENFGA_ENABLED" envDefault:"true"`
	Host       string `env:"OPENFGA_HOST" envDefault:"localhost"`
	Port       int    `env:"OPENFGA_PORT" envDefault:"8080"`
	StoreID    string `env:"OPENFGA_STORE_ID" envDefault:""`
	ModelID    string `env:"OPENFGA_MODEL_ID" envDefault:""`
	AutoCreate bool   `env:"OPENFGA_AUTO_CREATE" envDefault:"true"`
}

func (c OpenFGAConfig) URL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

type UserServiceConfig struct {
	BaseURL string `env:"USER_SERVICE_URL" envDefault:"http://localhost:8081"`
	Timeout int    `env:"USER_SERVICE_TIMEOUT" envDefault:"10"`
}

func Load() (cfg *Config, err error) {
	cfg = &Config{}
	if err = env.Parse(cfg); err != nil {
		return nil, err
	}

	// Apply profile-specific defaults
	cfg.applyProfile()

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
	if c.Database.Dynamo.Endpoint == "" {
		c.Database.Dynamo.Endpoint = "http://localhost:8000"
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
