// =============================================================================
// internal/config/config.go - Configuration management
package config

import (
	"os"
	"path/filepath"

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
	ShapesFile   string        `mapstructure:"shapes_file"`
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

// GetShapesFilePath returns the path to the shapes.json file
// It checks environment variables first, then config, then falls back to internal path
func GetShapesFilePath() string {
	// First check for environment variable override
	envPath := viper.GetString("shapes_file")
	if envPath != "" {
		// Check if the env-specified file exists
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		// If env var is set but file doesn't exist, still return it (let caller handle error)
		if viper.IsSet("shapes_file") {
			return envPath
		}
	}

	// Try to get from config file
	config, err := LoadConfig()
	if err == nil && config.ShapesFile != "" {
		// Check if the configured file exists
		if _, err := os.Stat(config.ShapesFile); err == nil {
			return config.ShapesFile
		}
	}

	// Fall back to internal path for development
	internalPath := filepath.Join("internal", "shapes", "shapes.json")
	if _, err := os.Stat(internalPath); err == nil {
		return internalPath
	}

	// If neither exists, return the configured path (or default)
	if config != nil && config.ShapesFile != "" {
		return config.ShapesFile
	}

	// Final fallback to default location
	return "/etc/oci-dr-hpc-shapes.json"
}
