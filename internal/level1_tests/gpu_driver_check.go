package level1_tests

import (
	"errors"
	"fmt"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
)

// GPUDriverCheckTestConfig represents the config needed to run this test
type GPUDriverCheckTestConfig struct {
	IsEnabled           bool     `json:"enabled"`
	Shape               string   `json:"shape"`
	BlacklistedVersions []string `json:"blacklisted_versions"`
	SupportedVersions   []string `json:"supported_versions"`
}

// getGpuDriverCheckTestConfig gets test config needed to run this test
func getGpuDriverCheckTestConfig() (*GPUDriverCheckTestConfig, error) {
	// Get shape from IMDS
	shape, err := executor.GetCurrentShape()
	if err != nil {
		return nil, err
	}

	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Initialize with defaults (H100 values)
	gpuDriverCheckTestConfig := &GPUDriverCheckTestConfig{
		IsEnabled:           false,
		Shape:               shape,
		BlacklistedVersions: []string{"470.57.02"},
		SupportedVersions:   []string{"450.119.03", "450.142.0", "470.103.01", "470.129.06", "470.141.03", "510.47.03", "535.104.12", "550.90.12"},
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "gpu_driver_check")
	if err != nil {
		return nil, err
	}
	gpuDriverCheckTestConfig.IsEnabled = enabled

	// Get threshold configuration if available
	threshold, err := limits.GetThresholdForTest(shape, "gpu_driver_check")
	if err == nil {
		// Parse threshold configuration from test_limits.json
		switch v := threshold.(type) {
		case map[string]interface{}:
			// Update blacklisted versions if specified
			if blacklisted, ok := v["blacklisted_versions"].([]interface{}); ok {
				var blacklistedVersions []string
				for _, version := range blacklisted {
					if versionStr, ok := version.(string); ok {
						blacklistedVersions = append(blacklistedVersions, versionStr)
					}
				}
				if len(blacklistedVersions) > 0 {
					gpuDriverCheckTestConfig.BlacklistedVersions = blacklistedVersions
				}
			}

			// Update supported versions if specified
			if supported, ok := v["supported_versions"].([]interface{}); ok {
				var supportedVersions []string
				for _, version := range supported {
					if versionStr, ok := version.(string); ok {
						supportedVersions = append(supportedVersions, versionStr)
					}
				}
				if len(supportedVersions) > 0 {
					gpuDriverCheckTestConfig.SupportedVersions = supportedVersions
				}
			}
		}
	}

	return gpuDriverCheckTestConfig, nil
}

// getGPUDriverVersions uses nvidia-smi to get GPU driver versions
func getGPUDriverVersions() ([]string, error) {
	// Use nvidia-smi to query driver versions
	result := executor.RunNvidiaSMIQuery("driver_version")
	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Parse the output
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return []string{}, nil
	}

	// Split by lines and filter out empty lines
	lines := strings.Split(output, "\n")
	var versions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			versions = append(versions, line)
		}
	}

	return versions, nil
}

// validateDriverVersions validates driver versions against blacklist and supported list
func validateDriverVersions(versions []string, blacklisted []string, supported []string) (string, error) {
	if len(versions) == 0 {
		return "FAIL", fmt.Errorf("no GPU driver versions found")
	}

	// Check if all versions are the same
	currentVersion := versions[0]
	for _, version := range versions {
		if version != currentVersion {
			return "FAIL", fmt.Errorf("driver versions are mismatched")
		}
	}

	// Check if version is blacklisted
	for _, blacklistedVersion := range blacklisted {
		if currentVersion == blacklistedVersion {
			return "FAIL", fmt.Errorf("driver version %s is blacklisted", currentVersion)
		}
	}

	// Check if version is supported
	for _, supportedVersion := range supported {
		if currentVersion == supportedVersion {
			return "PASS", nil
		}
	}

	// Version is not blacklisted but also not in supported list
	return "WARN", fmt.Errorf("driver version %s is unsupported but not blacklisted", currentVersion)
}

func RunGPUDriverCheck() error {
	logger.Info("=== GPU Driver Check ===")
	testConfig, err := getGpuDriverCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting GPU driver version check...")
	rep := reporter.GetReporter()

	// Step 1: Get GPU driver versions
	logger.Info("Step 1: Getting GPU driver versions...")
	versions, err := getGPUDriverVersions()
	if err != nil {
		logger.Error("GPU Driver Check: FAIL - Could not get GPU driver versions:", err)
		rep.AddGPUDriverResult("FAIL", "", err)
		return fmt.Errorf("could not get GPU driver versions: %w", err)
	}

	logger.Info("Found GPU driver versions:", versions)

	// Step 2: Validate driver versions
	logger.Info("Step 2: Validating driver versions...")
	logger.Info("Blacklisted versions:", testConfig.BlacklistedVersions)
	logger.Info("Supported versions:", testConfig.SupportedVersions)

	status, validationErr := validateDriverVersions(versions, testConfig.BlacklistedVersions, testConfig.SupportedVersions)

	var driverVersion string
	if len(versions) > 0 {
		driverVersion = versions[0]
	}

	switch status {
	case "PASS":
		logger.Info("GPU Driver Check: PASS - Driver version", driverVersion, "is supported")
		rep.AddGPUDriverResult("PASS", driverVersion, nil)
		return nil
	case "WARN":
		logger.Info("GPU Driver Check: WARN - Driver version", driverVersion, "is unsupported but not blacklisted")
		rep.AddGPUDriverResult("WARN", driverVersion, validationErr)
		return validationErr
	default: // FAIL
		logger.Error("GPU Driver Check: FAIL - Driver version", driverVersion, "validation failed:", validationErr)
		rep.AddGPUDriverResult("FAIL", driverVersion, validationErr)
		return validationErr
	}
}

func PrintGPUDriverCheck() {
	logger.Info("GPU Driver Check: Checking GPU driver version compatibility...")
	logger.Info("GPU Driver Check: PASS - Driver version is supported")
}