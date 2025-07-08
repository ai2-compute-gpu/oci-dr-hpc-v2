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

// PcieErrorCheckTestConfig represents the config needed to run this test
type PcieErrorCheckTestConfig struct {
	IsEnabled bool   `json:"enabled"`
	Shape     string `json:"shape"`
}

// Gets test config needed to run this test
func getPcieErrorCheckTestConfig() (*PcieErrorCheckTestConfig, error) {
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
	pcieErrorCheckTestConfig := &PcieErrorCheckTestConfig{
		IsEnabled: false,
		Shape:     shape,
	}

	enabled, err := limits.IsTestEnabled(shape, "pcie_error_check")
	if err != nil {
		return nil, err
	}
	pcieErrorCheckTestConfig.IsEnabled = enabled
	return pcieErrorCheckTestConfig, nil
}

func RunPCIeErrorCheck() error {
	logger.Info("=== PCIe Error Check ===")
	testConfig, err := getPcieErrorCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting PCIe health check...")
	logger.Info("This will take about 1 minute to complete.")
	rep := reporter.GetReporter()

	// Run the dmesg command to get system messages
	// dmesg shows kernel ring buffer messages including hardware errors
	logger.Info("Getting system messages...")
	result, err := executor.RunDmesg()
	if err != nil {
		logger.Error("Failed to run dmesg command:", err)
		logger.Info("PCIe Error Check: FAIL - Could not run dmesg command")
		rep.AddPCIeResult("FAIL", fmt.Errorf("could not run dmesg command: %v", err))
		return fmt.Errorf("could not run dmesg command: %v", err)
	}

	// Check if dmesg output is empty
	outputStr := result.Output
	if len(strings.TrimSpace(outputStr)) == 0 {
		logger.Error("No system messages found")
		logger.Info("PCIe Error Check: FAIL - No system messages found")
		err = fmt.Errorf("no system messages found")
		rep.AddPCIeResult("FAIL", err)
		return err
	}

	// Start with PASS status - we'll change to FAIL if we find errors
	logger.Info("Checking for PCIe errors...")

	// Look through each line of the dmesg output
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		// Skip lines that contain "capabilities" - these are not error messages
		if strings.Contains(line, "capabilities") {
			continue
		}

		// Look for lines that contain both "pcieport" and "error" (case insensitive)
		// pcieport = PCIe port driver messages
		// error = indicates an actual error occurred
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "pcieport") && strings.Contains(lowerLine, "error") {
			logger.Error(fmt.Sprintf("Found PCIe error: %s", line))
			logger.Info("PCIe Error Check: FAIL - PCIe errors found")
			err = fmt.Errorf("found PCIe error: %s", line)
			rep.AddPCIeResult("FAIL", err)
			return err
		}
	}

	logger.Info("PCIe Error Check: PASS - No PCIe errors found")
	rep.AddPCIeResult("PASS", nil)
	return nil
}
