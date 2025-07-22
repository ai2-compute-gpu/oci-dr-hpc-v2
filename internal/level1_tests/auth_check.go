package level1_tests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/shapes"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// AuthCheckResult represents the result of authentication check parsing
type AuthCheckResult struct {
	Device     string `json:"device"`
	AuthStatus string `json:"auth_status"`
}

// AuthCheckTestConfig represents the test configuration for authentication check
type AuthCheckTestConfig struct {
	IsEnabled bool `json:"enabled"`
}

// getAuthCheckTestConfig gets test config needed to run this test
func getAuthCheckTestConfig(shape string) (*AuthCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults from test_limits.json
	authCheckTestConfig := &AuthCheckTestConfig{
		IsEnabled: false,
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "auth_check")
	if err != nil {
		logger.Info("Authentication check test not found for shape", shape, ", defaulting to disabled")
		return authCheckTestConfig, nil
	}
	authCheckTestConfig.IsEnabled = enabled

	// If test is disabled, return early
	if !enabled {
		logger.Info("Authentication check test disabled for shape", shape)
		return authCheckTestConfig, nil
	}

	logger.Info("Successfully loaded auth_check configuration for shape", shape)
	return authCheckTestConfig, nil
}

// parseAuthResults parses the output from wpa_cli command and validates authentication status
func parseAuthResults(interfaceName string, wpaCliOutput string) (*AuthCheckResult, error) {
	result := &AuthCheckResult{
		Device:     interfaceName,
		AuthStatus: "FAIL - Unable to check authentication",
	}

	// If error, check if it's a command execution error
	if strings.HasPrefix(wpaCliOutput, "Error:") {
		result.AuthStatus = "FAIL - Unable to run wpa_cli command"
		return result, nil
	}

	if strings.TrimSpace(wpaCliOutput) == "" {
		result.AuthStatus = "FAIL - Unable to run wpa_cli command"
		return result, nil
	}

	// Check for specific authenticated status in the output
	if strings.Contains(wpaCliOutput, "Supplicant PAE state=AUTHENTICATED") {
		result.AuthStatus = "PASS"
	} else {
		result.AuthStatus = "FAIL - Interface not authenticated"
	}

	return result, nil
}

// RunAuthCheck performs the authentication check for RDMA interfaces
func RunAuthCheck() error {
	logger.Info("=== Authentication Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Authentication Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddAuthCheckResult("FAIL", []AuthCheckResult{}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	authCheckTestConfig, err := getAuthCheckTestConfig(shape)
	if err != nil {
		logger.Error("Authentication Check: FAIL - Could not get test configuration:", err)
		rep.AddAuthCheckResult("FAIL", []AuthCheckResult{}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !authCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Get device mapping using ibdev2netdev
	logger.Info("Step 3: Getting device mapping...")
	deviceMap, err := executor.GetIbdevToNetdevMap()
	if err != nil {
		logger.Error("Authentication Check: FAIL - Could not get device mapping:", err)
		rep.AddAuthCheckResult("FAIL", []AuthCheckResult{}, err)
		return fmt.Errorf("failed to get device mapping: %w", err)
	}

	// Step 4: Get RDMA NICs (non-VCN interfaces) from shapes.json
	logger.Info("Step 4: Getting RDMA NICs from shapes configuration...")
	shapeManager, err := shapes.NewShapeManager("internal/shapes/shapes.json")
	if err != nil {
		logger.Error("Authentication Check: FAIL - Could not load shapes configuration:", err)
		rep.AddAuthCheckResult("FAIL", []AuthCheckResult{}, err)
		return fmt.Errorf("failed to load shapes configuration: %w", err)
	}

	rdmaNics, err := shapeManager.GetRDMANics(shape)
	if err != nil {
		logger.Error("Authentication Check: FAIL - Could not get RDMA NICs for shape", shape, ":", err)
		rep.AddAuthCheckResult("FAIL", []AuthCheckResult{}, err)
		return fmt.Errorf("failed to get RDMA NICs for shape %s: %w", shape, err)
	}

	logger.Info("Found", len(rdmaNics), "RDMA NICs for shape", shape)

	// Step 5: Map RDMA device names to actual interfaces
	var interfacesToCheck []string
	deviceToInterfaceMap := make(map[string]string)

	for _, rdmaNic := range rdmaNics {
		logger.Info("Looking for RDMA device:", rdmaNic.DeviceName, "PCI:", rdmaNic.PCI)
		
		// Find the corresponding interface for this RDMA device
		if interfaceName, exists := deviceMap[rdmaNic.DeviceName]; exists {
			interfacesToCheck = append(interfacesToCheck, interfaceName)
			deviceToInterfaceMap[rdmaNic.DeviceName] = interfaceName
			logger.Info("Mapped RDMA device", rdmaNic.DeviceName, "to interface", interfaceName)
		} else {
			logger.Info("RDMA device", rdmaNic.DeviceName, "not found in device mapping")
		}
	}

	if len(interfacesToCheck) == 0 {
		errorStatement := "No RDMA interfaces found for checking"
		logger.Info("Authentication Check: INFO -", errorStatement)
		// This is not an error condition, just no interfaces to check
		rep.AddAuthCheckResult("SKIP", []AuthCheckResult{}, nil)
		return nil
	}

	logger.Info("Step 6: Found", len(interfacesToCheck), "RDMA interfaces to check:", interfacesToCheck)

	// Step 6: Check all RDMA interfaces for authentication status
	var allResults []AuthCheckResult

	for _, interfaceName := range interfacesToCheck {
		logger.Info("Running authentication check for RDMA interface", interfaceName)

		// Run wpa_cli status command for this interface
		result, err := executor.RunWpaCliStatus(interfaceName)
		var wpaCliOutput string
		if err != nil {
			logger.Error("Failed to run wpa_cli for interface", interfaceName, ":", err)
			wpaCliOutput = ""
		} else {
			wpaCliOutput = result.Output
		}

		// Parse results
		authResult, err := parseAuthResults(interfaceName, wpaCliOutput)
		if err != nil {
			logger.Errorf("Failed to parse authentication results for %s: %v", interfaceName, err)
			continue
		}

		allResults = append(allResults, *authResult)
	}

	// Step 7: Report results
	logger.Info("Step 7: Reporting results...")
	if len(allResults) == 0 {
		logger.Error("Authentication Check: FAIL - No authentication results obtained")
		err = fmt.Errorf("no authentication results obtained")
		rep.AddAuthCheckResult("FAIL", allResults, err)
		return err
	}

	// Check if all interfaces passed authentication
	allPassed := true
	for _, result := range allResults {
		if !strings.HasPrefix(result.AuthStatus, "PASS") {
			allPassed = false
			break
		}
	}

	if allPassed {
		logger.Info("Authentication Check: PASS - All RDMA interfaces are authenticated")
		rep.AddAuthCheckResult("PASS", allResults, nil)
		return nil
	} else {
		logger.Error("Authentication Check: FAIL - Some RDMA interfaces are not authenticated")
		err = fmt.Errorf("some RDMA interfaces are not authenticated")
		rep.AddAuthCheckResult("FAIL", allResults, err)
		return err
	}
}