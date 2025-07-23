package level1_tests

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// FabricManagerCheckResult represents the result of fabric manager service check
type FabricManagerCheckResult struct {
	Status      string `json:"status"`
	IsRunning   bool   `json:"is_running"`
	ServiceInfo string `json:"service_info,omitempty"`
	Message     string `json:"message"`
}

// FabricManagerCheckTestConfig represents the test configuration for fabric manager check
type FabricManagerCheckTestConfig struct {
	IsEnabled bool `json:"enabled"`
}

// getFabricManagerCheckTestConfig gets test config needed to run this test
func getFabricManagerCheckTestConfig(shape string) (*FabricManagerCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults from test_limits.json
	fabricManagerTestConfig := &FabricManagerCheckTestConfig{
		IsEnabled: false,
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "fabricmanager_check")
	if err != nil {
		logger.Info("Fabric manager check test not found for shape", shape, ", defaulting to disabled")
		return fabricManagerTestConfig, nil
	}
	fabricManagerTestConfig.IsEnabled = enabled

	// If test is disabled, return early
	if !enabled {
		logger.Info("Fabric manager check test disabled for shape", shape)
		return fabricManagerTestConfig, nil
	}

	logger.Info("Successfully loaded fabricmanager_check configuration for shape", shape)
	return fabricManagerTestConfig, nil
}

// checkFabricManagerService checks if nvidia-fabricmanager service is running
func checkFabricManagerService() *FabricManagerCheckResult {
	result := &FabricManagerCheckResult{
		Status:    "FAIL",
		IsRunning: false,
	}

	// Use systemctl to check service status
	cmd := exec.Command("systemctl", "status", "nvidia-fabricmanager")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		result.ServiceInfo = fmt.Sprintf("Failed to check service status: %s", err.Error())
		result.Message = "nvidia-fabricmanager service check failed"
		return result
	}

	// Check if service is active (running)
	outputLower := strings.ToLower(outputStr)
	if strings.Contains(outputLower, "active (running)") {
		result.Status = "PASS"
		result.IsRunning = true
		result.ServiceInfo = "nvidia-fabricmanager service is active and running"
		result.Message = "nvidia-fabricmanager service is properly running"
	} else {
		// Extract status information for debugging
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "active:") {
				result.ServiceInfo = strings.TrimSpace(line)
				break
			}
		}
		if result.ServiceInfo == "" {
			result.ServiceInfo = "nvidia-fabricmanager service is not active"
		}
		result.Message = "nvidia-fabricmanager service is not running"
	}

	return result
}

// RunFabricManagerCheck performs the fabric manager service check
func RunFabricManagerCheck() error {
	logger.Info("=== Fabric Manager Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Fabric Manager Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddFabricManagerResult("FAIL", &FabricManagerCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get shape from IMDS: %v", err),
		}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	fabricManagerTestConfig, err := getFabricManagerCheckTestConfig(shape)
	if err != nil {
		logger.Error("Fabric Manager Check: FAIL - Could not get test configuration:", err)
		rep.AddFabricManagerResult("FAIL", &FabricManagerCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get test configuration: %v", err),
		}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !fabricManagerTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Check nvidia-fabricmanager service
	logger.Info("Step 3: Checking nvidia-fabricmanager service...")
	result := checkFabricManagerService()

	// Step 4: Report results
	logger.Info("Step 4: Reporting results...")
	if result.Status == "PASS" {
		logger.Info("Fabric Manager Check: PASS -", result.Message)
		rep.AddFabricManagerResult("PASS", result, nil)
		return nil
	} else {
		logger.Error("Fabric Manager Check: FAIL -", result.Message)
		err = fmt.Errorf(result.Message)
		rep.AddFabricManagerResult("FAIL", result, err)
		return err
	}
}