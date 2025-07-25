package level1_tests

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// XIDError represents a single XID error with its details
type XIDError struct {
	XIDCode     string   `json:"xid_code"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Count       int      `json:"count"`
	PCIAddrs    []string `json:"pci_addresses"`
}

// GPUXIDCheckResult represents the result of GPU XID error check
type GPUXIDCheckResult struct {
	Status         string     `json:"status"`
	Message        string     `json:"message"`
	CriticalErrors []XIDError `json:"critical_errors,omitempty"`
	WarningErrors  []XIDError `json:"warning_errors,omitempty"`
}

// XIDErrorCode represents the structure of XID error codes in test_limits.json
type XIDErrorCode struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// GPUXIDCheckTestConfig represents the test configuration for GPU XID check
type GPUXIDCheckTestConfig struct {
	IsEnabled     bool                     `json:"enabled"`
	XIDErrorCodes map[string]XIDErrorCode `json:"xid_error_codes"`
}

// getGPUXIDCheckTestConfig gets test config needed to run this test
func getGPUXIDCheckTestConfig(shape string) (*GPUXIDCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults from test_limits.json
	gpuXIDTestConfig := &GPUXIDCheckTestConfig{
		IsEnabled:     false,
		XIDErrorCodes: make(map[string]XIDErrorCode),
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "gpu_xid_check")
	if err != nil {
		logger.Info("GPU XID check test not found for shape", shape, ", defaulting to disabled")
		return gpuXIDTestConfig, nil
	}
	gpuXIDTestConfig.IsEnabled = enabled

	// If test is disabled, return early
	if !enabled {
		logger.Info("GPU XID check test disabled for shape", shape)
		return gpuXIDTestConfig, nil
	}

	// Get XID error codes configuration from threshold
	xidCodesRaw, err := limits.GetThresholdForTest(shape, "gpu_xid_check")
	if err != nil {
		// Use a subset of critical XID codes as fallback
		logger.Info("XID error codes not found in configuration, using default critical codes")
		gpuXIDTestConfig.XIDErrorCodes = map[string]XIDErrorCode{
			"8":   {Description: "GPU stopped processing", Severity: "Critical"},
			"31":  {Description: "GPU memory page fault", Severity: "Critical"},
			"48":  {Description: "Double Bit ECC Error", Severity: "Critical"},
			"79":  {Description: "GPU has fallen off the bus", Severity: "Critical"},
			"92":  {Description: "High single-bit ECC error rate", Severity: "Critical"},
			"94":  {Description: "Contained ECC error", Severity: "Critical"},
			"95":  {Description: "Uncontained ECC error", Severity: "Critical"},
			"119": {Description: "GSP RPC Timeout", Severity: "Critical"},
			"120": {Description: "GSP Error", Severity: "Critical"},
		}
	} else {
		// Parse XID error codes from configuration - it should be under "xid_error_codes" key
		if thresholdMap, ok := xidCodesRaw.(map[string]interface{}); ok {
			if xidCodesRaw, exists := thresholdMap["xid_error_codes"]; exists {
				if xidCodesMap, ok := xidCodesRaw.(map[string]interface{}); ok {
					for xidCode, xidInfoRaw := range xidCodesMap {
						if xidInfoMap, ok := xidInfoRaw.(map[string]interface{}); ok {
							xidError := XIDErrorCode{}
							if desc, ok := xidInfoMap["description"].(string); ok {
								xidError.Description = desc
							}
							if sev, ok := xidInfoMap["severity"].(string); ok {
								xidError.Severity = sev
							}
							gpuXIDTestConfig.XIDErrorCodes[xidCode] = xidError
						}
					}
				}
			}
		}
	}

	logger.Info("Successfully loaded gpu_xid_check configuration for shape", shape)
	return gpuXIDTestConfig, nil
}

// checkGPUXIDErrors checks for XID errors in dmesg output
func checkGPUXIDErrors(xidErrorCodes map[string]XIDErrorCode) *GPUXIDCheckResult {
	result := &GPUXIDCheckResult{
		Status:         "PASS",
		Message:        "No XID errors found in system logs",
		CriticalErrors: []XIDError{},
		WarningErrors:  []XIDError{},
	}

	// Get dmesg output
	cmd := exec.Command("sudo", "dmesg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Failed to get dmesg output: %v", err)
		return result
	}

	dmesgOutput := string(output)

	// Check if any XID errors exist
	if !strings.Contains(dmesgOutput, "NVRM: Xid") {
		// No XID errors found
		return result
	}

	// Parse XID errors
	criticalCount := 0
	warningCount := 0

	for xidCode, xidInfo := range xidErrorCodes {
		// Pattern to match XID errors: NVRM: Xid (PCI:xxxx:xx:xx.x): <code>,
		pattern := fmt.Sprintf(`NVRM: Xid \(PCI:([^:]+): %s,`, xidCode)
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.FindAllStringSubmatch(dmesgOutput, -1)
		if len(matches) > 0 {
			// Extract PCI addresses
			pciAddrs := make([]string, 0, len(matches))
			for _, match := range matches {
				if len(match) > 1 {
					pciAddrs = append(pciAddrs, match[1])
				}
			}

			// Remove duplicates
			pciAddrMap := make(map[string]bool)
			for _, addr := range pciAddrs {
				pciAddrMap[addr] = true
			}
			uniquePCIAddrs := make([]string, 0, len(pciAddrMap))
			for addr := range pciAddrMap {
				uniquePCIAddrs = append(uniquePCIAddrs, addr)
			}

			xidError := XIDError{
				XIDCode:     xidCode,
				Description: xidInfo.Description,
				Severity:    xidInfo.Severity,
				Count:       len(matches),
				PCIAddrs:    uniquePCIAddrs,
			}

			if xidInfo.Severity == "Critical" {
				result.CriticalErrors = append(result.CriticalErrors, xidError)
				criticalCount++
			} else {
				result.WarningErrors = append(result.WarningErrors, xidError)
				warningCount++
			}
		}
	}

	// Set result status and message based on findings
	if criticalCount > 0 {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("Critical XID errors detected: %d critical, %d warnings", criticalCount, warningCount)
	} else if warningCount > 0 {
		result.Status = "WARN"
		result.Message = fmt.Sprintf("Warning XID errors detected: %d warnings", warningCount)
	} else {
		// XID messages found but no recognized error codes
		result.Status = "WARN"
		result.Message = "XID messages found but no recognized error codes"
	}

	return result
}

// RunGPUXIDCheck performs the GPU XID error check
func RunGPUXIDCheck() error {
	logger.Info("=== GPU XID Error Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("GPU XID Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddGPUXIDResult("FAIL", &GPUXIDCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get shape from IMDS: %v", err),
		}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	gpuXIDTestConfig, err := getGPUXIDCheckTestConfig(shape)
	if err != nil {
		logger.Error("GPU XID Check: FAIL - Could not get test configuration:", err)
		rep.AddGPUXIDResult("FAIL", &GPUXIDCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get test configuration: %v", err),
		}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !gpuXIDTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Check for GPU XID errors
	logger.Info("Step 3: Checking for GPU XID errors in system logs...")
	result := checkGPUXIDErrors(gpuXIDTestConfig.XIDErrorCodes)

	// Step 4: Report results
	logger.Info("Step 4: Reporting results...")
	if result.Status == "PASS" {
		logger.Info("GPU XID Check: PASS -", result.Message)
		rep.AddGPUXIDResult("PASS", result, nil)
		return nil
	} else if result.Status == "WARN" {
		logger.Info("GPU XID Check: WARN -", result.Message)
		rep.AddGPUXIDResult("WARN", result, nil)
		return nil // Warnings are not fatal
	} else {
		logger.Error("GPU XID Check: FAIL -", result.Message)
		err = fmt.Errorf("%s", result.Message)
		rep.AddGPUXIDResult("FAIL", result, err)
		return err
	}
}