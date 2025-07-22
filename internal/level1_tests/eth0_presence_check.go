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

type Eth0PresenceCheckTestConfig struct {
	IsEnabled bool `json:"enabled"`
}

// Gets test config needed to run this test
func getEth0PresenceCheckTestConfig(shape string) (*Eth0PresenceCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Result
	eth0PresenceCheckTestConfig := &Eth0PresenceCheckTestConfig{
		IsEnabled: false,
	}

	enabled, err := limits.IsTestEnabled(shape, "eth0_presence_check")
	if err != nil {
		return nil, err
	}
	eth0PresenceCheckTestConfig.IsEnabled = enabled

	return eth0PresenceCheckTestConfig, nil
}

// checkEth0Present checks if the eth0 interface is present
func checkEth0Present() (bool, error) {
	// Run ip addr command to get network interface information
	result, err := executor.RunIPAddr()
	if err != nil {
		return false, fmt.Errorf("failed to run ip addr: %w", err)
	}

	// Check if eth0 interface is in the output
	interfaceName := "eth0"
	output := strings.TrimSpace(result.Output)

	if output == "" {
		return false, nil
	}

	// Split output into lines and check each line for eth0
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if the line contains eth0
		if strings.Contains(line, interfaceName) {
			return true, nil
		}
	}

	return false, nil
}

func PrintEth0PresenceCheck() {
	// This function is a placeholder for eth0 presence check logic.
	// It should be implemented to check if the eth0 interface is present
	// and print the result.

	// Example implementation (to be replaced with actual logic):
	eth0Present := true // Placeholder value, replace with actual eth0 check logic
	if eth0Present {
		logger.Info("Eth0 Presence Check: PASS - eth0 interface is present")
	} else {
		logger.Info("Eth0 Presence Check: FAIL - eth0 interface is not present")
	}
}

func RunEth0PresenceCheck() error {
	logger.Info("=== Eth0 Presence Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Eth0 Presence Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddEth0PresenceResult("FAIL", false, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	eth0PresenceCheckTestConfig, err := getEth0PresenceCheckTestConfig(shape)
	if err != nil {
		logger.Error("Eth0 Presence Check: FAIL - Could not get test configuration:", err)
		rep.AddEth0PresenceResult("FAIL", false, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !eth0PresenceCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Error(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Check if eth0 interface is present
	logger.Info("Step 2: Checking if eth0 interface is present...")
	eth0Present, err := checkEth0Present()
	if err != nil {
		logger.Error("Eth0 Presence Check: FAIL - Could not check eth0 status:", err)
		rep.AddEth0PresenceResult("FAIL", false, err)
		return fmt.Errorf("failed to check eth0 status: %w", err)
	}

	// Step 4: Report results
	if eth0Present {
		logger.Info("Eth0 Presence Check: PASS - eth0 interface is present")
		rep.AddEth0PresenceResult("PASS", true, nil)
		return nil
	} else {
		logger.Error("Eth0 Presence Check: FAIL - eth0 interface is not present")
		err = fmt.Errorf("eth0 interface is not present")
		rep.AddEth0PresenceResult("FAIL", false, err)
		return err
	}
}