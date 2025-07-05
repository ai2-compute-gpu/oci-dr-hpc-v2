// =============================================================================
// internal/config/config.go - Configuration management
package config

import (
	"github.com/spf13/viper"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	File  string `mapstructure:"file"`
	Level string `mapstructure:"level"`
}

// Config holds the application configuration
type Config struct {
	Verbose      bool          `mapstructure:"verbose"`
	OutputFormat string        `mapstructure:"output"`
	TestLevel    string        `mapstructure:"level"`
	Logging      LoggingConfig `mapstructure:"logging"`
}

// LoadConfig loads configuration from viper
func LoadConfig() (*Config, error) {
	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	config, _ := LoadConfig()
	return config
}
