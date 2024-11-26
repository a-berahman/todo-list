package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port    string    `mapstructure:"SERVER_PORT"`
	DBURL   string    `mapstructure:"DATABASE_URL"`
	AWSConf AWSConfig `mapstructure:",squash"`
}

type AWSConfig struct {
	Endpoint string    `mapstructure:"AWS_ENDPOINT"`
	SQSConf  SQSConfig `mapstructure:",squash"`
	S3Conf   S3Config  `mapstructure:",squash"`
}

type SQSConfig struct {
	QueueURL   string `mapstructure:"AWS_SQS_QUEUE_URL"`
	Region     string `mapstructure:"AWS_SQS_REGION"`
	DisableSSL bool   `mapstructure:"AWS_SQS_DISABLE_SSL"`
}

type S3Config struct {
	Bucket         string `mapstructure:"AWS_S3_BUCKET"`
	Region         string `mapstructure:"AWS_S3_REGION"`
	DisableSSL     bool   `mapstructure:"AWS_S3_DISABLE_SSL"`
	ForcePathStyle bool   `mapstructure:"AWS_S3_FORCE_PATH_STYLE"`
}

// NewConfig initializes and returns a Config struct
func NewConfig() (*Config, error) {
	viper.Reset()

	setDefaults()
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("CLIENT_TIMEOUT", 5)
	viper.SetDefault("CRON_INTERVAL", 5)
	viper.SetDefault("PROVIDER_ENDPOINT", "https://default-endpoint.com")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SSL_MODE", "disable")
	viper.SetDefault("BINARY_PARAMS", "yes")
	viper.SetDefault("MAX_OPEN_CONNS", 25)
	viper.SetDefault("MAX_IDLE_CONNS", 10)
	viper.SetDefault("MIN_OPEN_CONNS", 5)
	viper.SetDefault("MAX_CONN_IDLE_TIME", 30*time.Minute)
	viper.SetDefault("MAX_CONN_LIFE_TIME", 10*time.Minute)
	viper.SetDefault("BREAKER_INTERVAL", 60*time.Second)
	viper.SetDefault("BREAKER_TIMEOUT", 10*time.Second)
	viper.SetDefault("BREAKER_FAILURES_THRESHOLD", 3)
	viper.SetDefault("BACKOFF_MAX_INTERVAL", 5*time.Second)
	viper.SetDefault("BACKOFF_MAX_ELAPSED_TIME", 25*time.Second)
}

func (c *Config) Validate() error {
	if c.Port == "" {
		slog.Error("SERVER_PORT is not set")
		return ErrMissingConfig("SERVER_PORT")
	}

	return nil
}

func ErrMissingConfig(key string) error {
	return fmt.Errorf("missing required configuration: %s", key)
}
