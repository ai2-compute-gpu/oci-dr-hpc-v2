package level1_tests

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/shapes"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// RowRemapErrorCheckTestConfig represents the config needed to run this test
type RowRemapErrorCheckTestConfig struct {
	IsEnabled        bool   `json:"enabled"`
	Shape            string `json:"shape"`
	Threshold        int    `json:"threshold"`
	NvidiaSMIVersion int    `json:"nvidia-smi-version"`
}

// Gets test config needed to run this test
func getRowRemapErrorCheckTestConfig() (*RowRemapErrorCheckTestConfig, error) {
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
	rowRemapErrorCheckTestConfig := &RowRemapErrorCheckTestConfig{
		IsEnabled:        false,
		Shape:            shape,
		Threshold:        0,
		NvidiaSMIVersion: 550,
	}

	enabled, err := limits.IsTestEnabled(shape, "row_remap_error_check")
	if err != nil {
		return nil, err
	}
	rowRemapErrorCheckTestConfig.IsEnabled = enabled

	// Get threshold configuration if available
	threshold, err := limits.GetThresholdForTest(shape, "row_remap_error_check")
	if err == nil {
		// Parse threshold configuration from test_limits.json
		switch v := threshold.(type) {
		case map[string]interface{}:
			// Get minimum-error threshold
			if minError, ok := v["minimum-error"].(float64); ok {
				rowRemapErrorCheckTestConfig.Threshold = int(minError)
			}

			// Get minimum-nvidia-smi-version
			if minVersion, ok := v["minimum-nvidia-smi-version"].(float64); ok {
				rowRemapErrorCheckTestConfig.NvidiaSMIVersion = int(minVersion)
			}
		}
	}

	return rowRemapErrorCheckTestConfig, nil
}

func parseRemappedRowsResults(output string, expectedBusIDs []string, threshold int) (int, []string, []string, error) {
	lines := strings.Split(output, "\n")
	foundBusIDs := make(map[string]bool)
	var failedBusIDs []string
	var missingBusIDs []string

	// Create a map of expected bus IDs in lowercase for case insensitive comparison
	expectedBusIDsMap := make(map[string]string)
	for _, expectedID := range expectedBusIDs {
		expectedBusIDsMap[strings.ToLower(expectedID)] = expectedID
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Error:") {
			continue
		}

		// Parse CSV format: gpu_bus_id, remapped_rows.failure
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		busID := strings.TrimSpace(parts[0])
		failureCountStr := strings.TrimSpace(parts[1])

		// Convert to standard PCI format (shapes.json uses 0000:xx:xx.x format)
		if strings.HasPrefix(busID, "00000000:") {
			busID = busID[8:] // Remove the first 8 characters
			busID = "0000" + busID
		}

		// Normalize bus ID to lowercase for comparison
		busIDLower := strings.ToLower(busID)

		// Find the matching expected bus ID (case insensitive)
		var matchedBusID string
		if originalID, exists := expectedBusIDsMap[busIDLower]; exists {
			matchedBusID = originalID
		} else {
			matchedBusID = busID // Use the original format if no match found
		}

		foundBusIDs[matchedBusID] = true

		// Check if failure count exceeds threshold
		failureCount, err := strconv.Atoi(failureCountStr)
		if err != nil {
			// If we can't parse the failure count, treat as failure
			failedBusIDs = append(failedBusIDs, matchedBusID)
			continue
		}

		if failureCount > threshold {
			failedBusIDs = append(failedBusIDs, matchedBusID)
		}
	}

	// Check for missing GPUs
	for _, expectedID := range expectedBusIDs {
		if !foundBusIDs[expectedID] {
			missingBusIDs = append(missingBusIDs, expectedID)
		}
	}

	return len(failedBusIDs), failedBusIDs, missingBusIDs, nil
}

func RunRowRemapErrorCheck() error {
	logger.Info("=== Row Remap Error Check ===")
	testConfig, err := getRowRemapErrorCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Check nvidia-smi driver version first
	driverVersion, err := executor.GetNvidiaSMIDriverVersion()
	if err != nil {
		logger.Error("Failed to get nvidia-smi driver version:", err)
		logger.Info("Row Remap Error Check: FAIL - Could not determine nvidia-smi driver version")
		rep := reporter.GetReporter()
		rep.AddRowRemapResult("FAIL", fmt.Errorf("could not determine nvidia-smi driver version: %v", err), 0)
		return fmt.Errorf("could not determine nvidia-smi driver version: %v", err)
	}

	if driverVersion < testConfig.NvidiaSMIVersion {
		errorStatement := fmt.Sprintf("Not applicable for nvidia-smi driver : %d", driverVersion)
		logger.Info(errorStatement)
		rep := reporter.GetReporter()
		rep.AddRowRemapResult(errorStatement, nil, 0)
		return errors.New(errorStatement)
	}

	// Get expected GPU bus IDs from shapes configuration
	shapeManager, err := shapes.GetDefaultShapeManager()
	if err != nil {
		logger.Error("Failed to load shapes configuration:", err)
		logger.Info("Row Remap Error Check: FAIL - Could not load shapes configuration")
		rep := reporter.GetReporter()
		rep.AddRowRemapResult("FAIL", fmt.Errorf("could not load shapes configuration: %v", err), 0)
		return fmt.Errorf("could not load shapes configuration: %v", err)
	}

	expectedBusIDs, err := shapeManager.GetGPUPCIAddresses(testConfig.Shape)
	if err != nil {
		logger.Error("Failed to get GPU PCI addresses for shape:", testConfig.Shape, err)
		logger.Info("Row Remap Error Check: FAIL - Could not get expected GPU PCI addresses")
		rep := reporter.GetReporter()
		rep.AddRowRemapResult("FAIL", fmt.Errorf("could not get expected GPU PCI addresses: %v", err), 0)
		return fmt.Errorf("could not get expected GPU PCI addresses: %v", err)
	}

	logger.Info("Starting row remap error check...")
	logger.Info("This will take about 1 minute to complete.")
	rep := reporter.GetReporter()

	// Run the nvidia-smi remapped rows query
	logger.Info("Checking for GPU row remap errors...")
	result := executor.RunNvidiaSMIRemappedRowsQuery()
	if !result.Available {
		logger.Error("Failed to run nvidia-smi remapped rows query:", result.Error)
		logger.Info("Row Remap Error Check: FAIL - Could not run nvidia-smi remapped rows query")
		rep.AddRowRemapResult("FAIL", fmt.Errorf("could not run nvidia-smi remapped rows query: %s", result.Error), 0)
		return fmt.Errorf("could not run nvidia-smi remapped rows query: %s", result.Error)
	}

	// Parse the nvidia-smi output for row remap failures
	failureCount, failedBusIDs, missingBusIDs, err := parseRemappedRowsResults(result.Output, expectedBusIDs, testConfig.Threshold)
	if err != nil {
		logger.Error("Failed to parse remapped rows results:", err)
		logger.Info("Row Remap Error Check: FAIL - Could not parse remapped rows results")
		rep.AddRowRemapResult("FAIL", fmt.Errorf("could not parse remapped rows results: %v", err), 0)
		return fmt.Errorf("could not parse remapped rows results: %v", err)
	}

	// Check for failures
	if len(failedBusIDs) > 0 || len(missingBusIDs) > 0 {
		if len(failedBusIDs) > 0 {
			logger.Error("Found GPUs with row remap failures:")
			for _, busID := range failedBusIDs {
				logger.Error("GPU with failures:", busID)
			}
		}
		if len(missingBusIDs) > 0 {
			logger.Error("Missing expected GPUs:")
			for _, busID := range missingBusIDs {
				logger.Error("Missing GPU:", busID)
			}
		}
		logger.Info("Row Remap Error Check: FAIL - Row remap errors or missing GPUs found")
		err = fmt.Errorf("found %d GPU(s) with row remap failures, %d missing GPU(s)", len(failedBusIDs), len(missingBusIDs))
		rep.AddRowRemapResult("FAIL", err, failureCount)
		return err
	}

	// No failures found - check passes
	logger.Info("No row remap errors found")
	logger.Info("Row Remap Error Check: PASS")
	rep.AddRowRemapResult("PASS", nil, 0)
	return nil
}
