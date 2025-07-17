package level1_tests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

type PeermemModuleCheckTestConfig struct {
	IsEnabled bool `json:"enabled"`
}

// Gets test config needed to run this test
func getPeermemModuleCheckTestConfig(shape string) (*PeermemModuleCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Result
	peermemModuleCheckTestConfig := &PeermemModuleCheckTestConfig{
		IsEnabled: false,
	}

	enabled, err := limits.IsTestEnabled(shape, "peermem_module_check")
	if err != nil {
		return nil, err
	}
	peermemModuleCheckTestConfig.IsEnabled = enabled

	return peermemModuleCheckTestConfig, nil
}

// checkPeermemModuleLoaded checks if the nvidia_peermem module is loaded
func checkPeermemModuleLoaded() (bool, error) {
	// Run lsmod command to get loaded modules
	result, err := executor.RunLsmod()
	if err != nil {
		return false, fmt.Errorf("failed to run lsmod: %w", err)
	}

	// Check if nvidia_peermem module is in the output
	moduleName := "nvidia_peermem"
	output := strings.TrimSpace(result.Output)

	if output == "" {
		return false, nil
	}

	// Split output into lines and check each line
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Get the first field (module name)
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == moduleName {
			return true, nil
		}
	}

	return false, nil
}

func PrintPeermemModuleCheck() {
	// This function is a placeholder for peermem module check logic.
	// It should be implemented to check if the nvidia_peermem module is loaded
	// and print the result.

	// Example implementation (to be replaced with actual logic):
	moduleLoaded := true // Placeholder value, replace with actual module check logic
	if moduleLoaded {
		logger.Info("Peermem Module Check: PASS - nvidia_peermem module is loaded")
	} else {
		logger.Info("Peermem Module Check: FAIL - nvidia_peermem module is not loaded")
	}
}

func RunPeermemModuleCheck() error {
	logger.Info("=== Peermem Module Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Peermem Module Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddPeerMemResult("FAIL", false, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	peermemModuleCheckTestConfig, err := getPeermemModuleCheckTestConfig(shape)
	if err != nil {
		logger.Error("Peermem Module Check: FAIL - Could not get test configuration:", err)
		rep.AddPeerMemResult("FAIL", false, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !peermemModuleCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Error(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Check if nvidia_peermem module is loaded
	logger.Info("Step 2: Checking if nvidia_peermem module is loaded...")
	moduleLoaded, err := checkPeermemModuleLoaded()
	if err != nil {
		logger.Error("Peermem Module Check: FAIL - Could not check module status:", err)
		rep.AddPeerMemResult("FAIL", false, err)
		return fmt.Errorf("failed to check module status: %w", err)
	}

	// Step 4: Report results
	if moduleLoaded {
		logger.Info("Peermem Module Check: PASS - nvidia_peermem module is loaded")
		rep.AddPeerMemResult("PASS", true, nil)
		return nil
	} else {
		logger.Error("Peermem Module Check: FAIL - nvidia_peermem module is not loaded")
		err = fmt.Errorf("nvidia_peermem module is not loaded")
		rep.AddPeerMemResult("FAIL", false, err)
		return err
	}
}
