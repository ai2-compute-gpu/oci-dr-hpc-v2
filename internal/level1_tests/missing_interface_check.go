package level1_tests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
)

// MissingInterfaceCheckTestConfig represents the config needed to run this test
type MissingInterfaceCheckTestConfig struct {
	IsEnabled bool   `json:"enabled"`
	Shape     string `json:"shape"`
	Threshold int    `json:"threshold"`
}

// Gets test config needed to run this test
func getMissingInterfaceCheckTestConfig() (*MissingInterfaceCheckTestConfig, error) {
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

	// Result
	missingInterfaceCheckTestConfig := &MissingInterfaceCheckTestConfig{
		IsEnabled: false,
		Shape:     shape,
		Threshold: 0,
	}

	enabled, err := limits.IsTestEnabled(shape, "missing_interface_check")
	if err != nil {
		return nil, err
	}
	missingInterfaceCheckTestConfig.IsEnabled = enabled

	// Get threshold value
	if enabled {
		threshold, err := limits.GetThresholdForTest(shape, "missing_interface_check")
		if err != nil {
			return nil, err
		}

		if thresholdFloat, ok := threshold.(float64); ok {
			missingInterfaceCheckTestConfig.Threshold = int(thresholdFloat)
		}
	}

	return missingInterfaceCheckTestConfig, nil
}

func parseLspciForMissingInterfaces(output string) (int, error) {
	lines := strings.Split(output, "\n")
	missingCount := 0

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "rev ff") {
			missingCount++
		}
	}

	return missingCount, nil
}

func RunMissingInterfaceCheck() error {
	logger.Info("=== Missing Interface Check ===")
	testConfig, err := getMissingInterfaceCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting missing interface check...")
	logger.Info("This will take about 1 minute to complete.")
	rep := reporter.GetReporter()

	// Run the lspci command to check for missing interfaces
	logger.Info("Checking for missing PCIe interfaces...")
	result, err := executor.RunLspci()
	if err != nil {
		logger.Error("Failed to run lspci command:", err)
		logger.Info("Missing Interface Check: FAIL - Could not run lspci command")
		rep.AddMissingInterfaceResult("FAIL", 0, fmt.Errorf("could not run lspci command: %v", err))
		return fmt.Errorf("could not run lspci command: %v", err)
	}

	// Parse the lspci output for missing interfaces
	missingCount, err := parseLspciForMissingInterfaces(result.Output)
	if err != nil {
		logger.Error("Failed to parse lspci output:", err)
		logger.Info("Missing Interface Check: FAIL - Could not parse lspci output")
		rep.AddMissingInterfaceResult("FAIL", 0, fmt.Errorf("could not parse lspci output: %v", err))
		return fmt.Errorf("could not parse lspci output: %v", err)
	}

	// Check if missing interfaces exceed threshold
	if missingCount > testConfig.Threshold {
		logger.Error(fmt.Sprintf("Found %d missing interface(s), exceeds threshold of %d", missingCount, testConfig.Threshold))
		logger.Info("Missing Interface Check: FAIL - Missing PCIe interfaces detected")
		err = fmt.Errorf("found %d missing PCIe interfaces, exceeds threshold of %d", missingCount, testConfig.Threshold)
		rep.AddMissingInterfaceResult("FAIL", missingCount, err)
		return err
	}

	// No missing interfaces found - check passes
	logger.Info("No missing interfaces found")
	logger.Info("Missing Interface Check: PASS")
	rep.AddMissingInterfaceResult("PASS", 0, nil)
	return nil
}
