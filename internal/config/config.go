// =============================================================================
// internal/config/config.go - Configuration management
package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Verbose      bool   `mapstructure:"verbose"`
	OutputFormat string `mapstructure:"output"`
	TestLevel    string `mapstructure:"level"`
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
