package test_limits

import (
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"os"
	"path/filepath"
	"runtime"
)

// SRAMThreshold represents the threshold configuration for SRAM checks
type SRAMThreshold struct {
	Uncorrectable int `json:"uncorrectable"`
	Correctable   int `json:"correctable"`
}

// TestConfig represents a generic test configuration that can be extended
type TestConfig struct {
	Enabled      bool        `json:"enabled"`
	TestCategory string      `json:"test_category"`
	Threshold    interface{} `json:"threshold,omitempty"`
}

// ShapeTestConfig represents the test configuration for a specific shape
// Uses map to allow dynamic addition of new test types
type ShapeTestConfig map[string]*TestConfig

// TestLimits represents the complete test limits configuration
type TestLimits struct {
	TestLimits map[string]ShapeTestConfig `json:"test_limits"`
}

// getPackageDir returns the directory where this package resides
func getPackageDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get current file path")
	}
	return filepath.Dir(filename), nil
}

// getDefaultConfigPath returns the default path for test_limits.json
func getDefaultConfigPath() (string, error) {
	// Check multiple locations in order of priority
	paths := []string{
		"./test_limits.json",                                  // Current directory override
		"/etc/oci-dr-hpc-test-limits.json",                   // System installation path
		func() string {                                        // User config directory
			if home := os.Getenv("HOME"); home != "" {
				return filepath.Join(home, ".config", "oci-dr-hpc", "test_limits.json")
			}
			return ""
		}(),
	}
	
	// Add development path as fallback
	packageDir, err := getPackageDir()
	if err == nil {
		paths = append(paths, filepath.Join(packageDir, "test_limits.json"))
	}
	
	// Try each path in order
	for _, path := range paths {
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				logger.Infof("LoadTestLimits: using config path %s", path)
				return path, nil
			}
		}
	}
	
	// No config file found
	return "", fmt.Errorf("test_limits.json not found in any of the expected locations: %v", paths)
}

// LoadTestLimits reads and parses the test limits JSON configuration file from the package directory
func LoadTestLimits() (*TestLimits, error) {
	configPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}
	return LoadTestLimitsFromFile(configPath)
}

// LoadTestLimitsFromFile reads and parses the test limits JSON configuration file from a specific path
func LoadTestLimitsFromFile(filePath string) (*TestLimits, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("Unable to read test limits file: %s", filePath)
		return nil, fmt.Errorf("failed to read test limits file %s: %w", filePath, err)
	}

	// Parse the JSON
	var testLimits TestLimits
	if err := json.Unmarshal(data, &testLimits); err != nil {
		logger.Errorf("failed to parse test limits JSON: %s", filePath)
		return nil, fmt.Errorf("failed to parse test limits JSON: %w", err)
	}
	logger.Infof("Test configs: %+v", testLimits)

	return &testLimits, nil
}

// GetTestConfig returns the configuration for a specific test type and shape
func (tl *TestLimits) GetTestConfig(shapeName, testType string) (*TestConfig, error) {
	if shapeConfig, exists := tl.TestLimits[shapeName]; exists {
		if testConfig, exists := shapeConfig[testType]; exists {
			return testConfig, nil
		}
		return nil, fmt.Errorf("test type %s not found for shape %s", testType, shapeName)
	}
	return nil, fmt.Errorf("no test configuration found for shape: %s", shapeName)
}

// IsTestEnabled returns whether a specific test is enabled for a shape
func (tl *TestLimits) IsTestEnabled(shapeName, testType string) (bool, error) {
	testConfig, err := tl.GetTestConfig(shapeName, testType)
	if err != nil {
		return false, err
	}
	return testConfig.Enabled, nil
}

// GetThresholdForTest returns the raw threshold value for any test type (extensible)
func (tl *TestLimits) GetThresholdForTest(shapeName, testType string) (interface{}, error) {
	testConfig, err := tl.GetTestConfig(shapeName, testType)
	if err != nil {
		return nil, err
	}
	if !testConfig.Enabled {
		return nil, fmt.Errorf("test %s is disabled for shape: %s", testType, shapeName)
	}
	return testConfig.Threshold, nil
}

// GetAvailableShapes returns a list of all available shape names
func (tl *TestLimits) GetAvailableShapes() []string {
	shapes := make([]string, 0, len(tl.TestLimits))
	for shapeName := range tl.TestLimits {
		shapes = append(shapes, shapeName)
	}
	return shapes
}

// GetEnabledTests returns a list of enabled test types for a specific shape
func (tl *TestLimits) GetEnabledTests(shapeName string) ([]string, error) {
	if shapeConfig, exists := tl.TestLimits[shapeName]; exists {
		var enabledTests []string
		for testType, testConfig := range shapeConfig {
			if testConfig.Enabled {
				enabledTests = append(enabledTests, testType)
			}
		}
		return enabledTests, nil
	}
	return nil, fmt.Errorf("no test configuration found for shape: %s", shapeName)
}
