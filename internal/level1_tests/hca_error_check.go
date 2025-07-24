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

// HcaErrorCheckTestConfig represents the config needed to run this test
type HcaErrorCheckTestConfig struct {
	IsEnabled bool   `json:"enabled"`
	Shape     string `json:"shape"`
}

// Gets test config needed to run this test
func getHcaErrorCheckTestConfig() (*HcaErrorCheckTestConfig, error) {
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
	hcaErrorCheckTestConfig := &HcaErrorCheckTestConfig{
		IsEnabled: false,
		Shape:     shape,
	}

	enabled, err := limits.IsTestEnabled(shape, "hca_error_check")
	if err != nil {
		return nil, err
	}
	hcaErrorCheckTestConfig.IsEnabled = enabled
	return hcaErrorCheckTestConfig, nil
}

func parseDmesgForMLX5FatalErrors(output string) []string {
	lines := strings.Split(output, "\n")
	var mlx5FatalLines []string

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "mlx5") && strings.Contains(lowerLine, "fatal") {
			mlx5FatalLines = append(mlx5FatalLines, line)
		}
	}

	return mlx5FatalLines
}

func RunHCAErrorCheck() error {
	logger.Info("=== HCA Error Check ===")
	testConfig, err := getHcaErrorCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting HCA error check...")
	logger.Info("This will take about 1 minute to complete.")
	rep := reporter.GetReporter()

	// Run the dmesg command to get MLX5 fatal error messages
	// dmesg -T shows timestamped kernel messages, filtered for mlx5 and Fatal errors
	logger.Info("Checking for MLX5 fatal errors...")
	result, err := executor.RunDmesg("-T")
	if err != nil {
		logger.Error("Failed to run dmesg command:", err)
		logger.Info("HCA Error Check: FAIL - Could not run dmesg command")
		rep.AddHCAResult("FAIL", fmt.Errorf("could not run dmesg command: %v", err))
		return fmt.Errorf("could not run dmesg command: %v", err)
	}

	// Parse the dmesg output for MLX5 fatal errors
	mlx5FatalLines := parseDmesgForMLX5FatalErrors(result.Output)

	// For HCA check: if we get ANY output, that means fatal errors were found
	// This is opposite logic from PCIe check
	if len(mlx5FatalLines) > 0 {
		// Found fatal errors - check fails
		logger.Error("Found MLX5 fatal errors:")
		for _, line := range mlx5FatalLines {
			logger.Error(line)
		}
		logger.Info("HCA Error Check: FAIL - MLX5 fatal errors found")
		err = fmt.Errorf("found MLX5 fatal errors: %d errors detected", len(mlx5FatalLines))
		rep.AddHCAResult("FAIL", err)
		return err
	}

	// No fatal errors found - check passes
	logger.Info("No MLX5 fatal errors found")
	logger.Info("HCA Error Check: PASS")
	rep.AddHCAResult("PASS", nil)
	return nil
}