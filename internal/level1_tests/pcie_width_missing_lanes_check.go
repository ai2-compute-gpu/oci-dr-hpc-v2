package level1_tests

import (
	"fmt"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
)

// PCIeWidthResult represents PCIe width analysis results
type PCIeWidthResult struct {
	WidthCounts map[string]int `json:"width_counts"`
	SpeedCounts map[string]int `json:"speed_counts"`
	StateErrors []string       `json:"state_errors"`
	Success     bool           `json:"success"`
	ErrorMsg    string         `json:"error_message,omitempty"`
}

// PCIeWidthMissingLanesTestConfig holds configuration for this test
type PCIeWidthMissingLanesTestConfig struct {
	IsEnabled            bool               `json:"enabled"`
	ExpectedGPUWidths    map[string]int     `json:"expected_gpu_widths"`
	ExpectedRDMAWidths   map[string]int     `json:"expected_rdma_widths"`
	ExpectedGPUSpeeds    map[string]int     `json:"expected_gpu_speeds"`
	ExpectedRDMASpeeds   map[string]int     `json:"expected_rdma_speeds"`
	ExpectedLinkState    string             `json:"expected_link_state"`
}

// getpcieWidthMissingLanesTestConfig loads test configuration
func getPcieWidthMissingLanesTestConfig(shape string) (*PCIeWidthMissingLanesTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Initialize config with defaults
	config := &PCIeWidthMissingLanesTestConfig{
		IsEnabled:           false,
		ExpectedGPUWidths:   make(map[string]int),
		ExpectedRDMAWidths:  make(map[string]int),
		ExpectedGPUSpeeds:   make(map[string]int),
		ExpectedRDMASpeeds:  make(map[string]int),
		ExpectedLinkState:   "Ok",
	}

	// Check if test is enabled
	enabled, err := limits.IsTestEnabled(shape, "pcie_width_missing_lanes_check")
	if err != nil {
		return nil, err
	}
	config.IsEnabled = enabled

	// Get expected GPU widths
	gpuWidthsData, err := limits.GetThresholdForTest(shape, "pcie_width_missing_lanes_check")
	if err != nil {
		return nil, err
	}

	// Parse the threshold data
	if thresholdMap, ok := gpuWidthsData.(map[string]interface{}); ok {
		// Parse GPU widths
		if gpuWidths, exists := thresholdMap["gpu_widths"]; exists {
			if gpuMap, ok := gpuWidths.(map[string]interface{}); ok {
				for width, count := range gpuMap {
					if countFloat, ok := count.(float64); ok {
						config.ExpectedGPUWidths[width] = int(countFloat)
					}
				}
			}
		}

		// Parse RDMA widths
		if rdmaWidths, exists := thresholdMap["rdma_widths"]; exists {
			if rdmaMap, ok := rdmaWidths.(map[string]interface{}); ok {
				for width, count := range rdmaMap {
					if countFloat, ok := count.(float64); ok {
						config.ExpectedRDMAWidths[width] = int(countFloat)
					}
				}
			}
		}

		// Parse GPU speeds
		if gpuSpeeds, exists := thresholdMap["gpu_speeds"]; exists {
			if gpuSpeedMap, ok := gpuSpeeds.(map[string]interface{}); ok {
				for speed, count := range gpuSpeedMap {
					if countFloat, ok := count.(float64); ok {
						config.ExpectedGPUSpeeds[speed] = int(countFloat)
					}
				}
			}
		}

		// Parse RDMA speeds
		if rdmaSpeeds, exists := thresholdMap["rdma_speeds"]; exists {
			if rdmaSpeedMap, ok := rdmaSpeeds.(map[string]interface{}); ok {
				for speed, count := range rdmaSpeedMap {
					if countFloat, ok := count.(float64); ok {
						config.ExpectedRDMASpeeds[speed] = int(countFloat)
					}
				}
			}
		}

		// Parse expected link state
		if linkState, exists := thresholdMap["expected_link_state"]; exists {
			if stateStr, ok := linkState.(string); ok {
				config.ExpectedLinkState = stateStr
			}
		}
	}

	return config, nil
}

// PCIeParseResult holds parsed PCIe information
type PCIeParseResult struct {
	WidthCounts map[string]int
	SpeedCounts map[string]int
	StateErrors []string
}

// FilterLspciForDevice filters lspci output to extract LnkSta lines for specific device type
func FilterLspciForDevice(lspciOutput string, deviceType string) string {
	lines := strings.Split(lspciOutput, "\n")
	var filteredLines []string
	var deviceFound bool
	
	devicePattern := strings.ToLower(deviceType)
	
	for _, line := range lines {
		// Check if this line is a PCI device header (starts with bus:device.function)
		if regexp.MustCompile(`^[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]`).MatchString(line) {
			// Reset device found flag
			deviceFound = false
			
			// Check if this device matches our target type
			lowerLine := strings.ToLower(line)
			switch devicePattern {
			case "nvidia", "gpu", "nvswitch":
				deviceFound = strings.Contains(lowerLine, "nvidia")
			case "mellanox", "rdma":
				deviceFound = strings.Contains(lowerLine, "mellanox")
			}
		}
		
		// If we found a matching device and this line contains LnkSta, include it
		if deviceFound && strings.Contains(line, "LnkSta:") {
			filteredLines = append(filteredLines, line)
		}
	}
	
	return strings.Join(filteredLines, "\n")
}

// AggregateLinkStatistics aggregates LnkSta lines by counting identical entries
func AggregateLinkStatistics(linkStatusLines string) string {
	if strings.TrimSpace(linkStatusLines) == "" {
		return ""
	}
	
	lines := strings.Split(strings.TrimSpace(linkStatusLines), "\n")
	countMap := make(map[string]int)
	
	// Count occurrences of each unique LnkSta line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			countMap[line]++
		}
	}
	
	// Convert back to the format expected by parseLspciWidthOutput
	var result []string
	for line, count := range countMap {
		result = append(result, fmt.Sprintf("%d\t%s", count, line))
	}
	
	return strings.Join(result, "\n")
}

// parseLspciWidthOutput parses lspci output to extract PCIe width, speed, and state information
func parseLspciWidthOutput(output string, expectedLinkState string) PCIeParseResult {
	result := PCIeParseResult{
		WidthCounts: make(map[string]int),
		SpeedCounts: make(map[string]int),
		StateErrors: []string{},
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Enhanced regex to capture count, speed with (ok), width with (ok)
	// Example: "4       LnkSta: Speed 16GT/s (ok), Width x16 (ok)"
	re := regexp.MustCompile(`^\s*(\d+)\s+LnkSta:\s*Speed\s+([^\s]+)\s*\(([^)]+)\),\s*Width\s+x(\d+)\s*\(([^)]+)\)`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		matches := re.FindStringSubmatch(line)
		if len(matches) == 6 {
			count, err := strconv.Atoi(matches[1])
			if err != nil {
				logger.Debugf("Failed to parse count from line: %s", line)
				continue
			}
			
			speed := strings.TrimSpace(matches[2])
			speedState := strings.TrimSpace(matches[3])
			width := matches[4]
			widthState := strings.TrimSpace(matches[5])
			
			// Count widths only if width state is ok
			widthKey := fmt.Sprintf("Width x%s", width)
			if widthState == "ok" {
				result.WidthCounts[widthKey] += count
			} else {
				result.StateErrors = append(result.StateErrors, 
					fmt.Sprintf("%d devices have width state '%s' instead of 'ok'", count, widthState))
			}
			
			// Count speeds only if speed state is ok
			speedKey := fmt.Sprintf("Speed %s", speed)
			if speedState == "ok" {
				result.SpeedCounts[speedKey] += count
			} else {
				result.StateErrors = append(result.StateErrors, 
					fmt.Sprintf("%d devices have speed state '%s' instead of 'ok'", count, speedState))
			}
			
			logger.Debugf("Parsed PCIe: %s (%s) = %d, %s (%s) = %d", widthKey, widthState, count, speedKey, speedState, count)
		}
	}
	
	return result
}

// checkGPUNVSwitchPCIeWidth checks PCIe width, speed, and state for GPU and NVSwitch interfaces
func checkGPUNVSwitchPCIeWidth(expectedLinkState string) PCIeWidthResult {
	logger.Info("Checking GPU/NVSwitch PCIe width, speed, and state...")
	
	// Execute the lspci command with verbose output
	result, err := executor.RunLspci("-vvv")
	if err != nil {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("Failed to execute lspci command: %v", err),
		}
	}
	
	if result.ExitCode != 0 {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("lspci command failed with exit code %d: %s", result.ExitCode, result.Output),
		}
	}
	
	// Filter lspci output for NVIDIA devices and extract LnkSta lines
	filteredOutput := FilterLspciForDevice(result.Output, "nvidia")
	if strings.TrimSpace(filteredOutput) == "" {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: "No NVIDIA PCIe devices found",
		}
	}
	
	// Aggregate identical LnkSta lines with counts
	aggregatedOutput := AggregateLinkStatistics(filteredOutput)
	
	parseResult := parseLspciWidthOutput(aggregatedOutput, expectedLinkState)
	
	return PCIeWidthResult{
		WidthCounts: parseResult.WidthCounts,
		SpeedCounts: parseResult.SpeedCounts,
		StateErrors: parseResult.StateErrors,
		Success:     true,
	}
}

// checkRDMAPCIeWidth checks PCIe width, speed, and state for RDMA interfaces
func checkRDMAPCIeWidth(expectedLinkState string) PCIeWidthResult {
	logger.Info("Checking RDMA PCIe width, speed, and state...")
	
	// Execute the lspci command with verbose output
	result, err := executor.RunLspci("-vvv")
	if err != nil {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("Failed to execute lspci command: %v", err),
		}
	}
	
	if result.ExitCode != 0 {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("lspci command failed with exit code %d: %s", result.ExitCode, result.Output),
		}
	}
	
	// Filter lspci output for Mellanox devices and extract LnkSta lines
	filteredOutput := FilterLspciForDevice(result.Output, "mellanox")
	if strings.TrimSpace(filteredOutput) == "" {
		return PCIeWidthResult{
			Success:  false,
			ErrorMsg: "No Mellanox PCIe devices found",
		}
	}
	
	// Aggregate identical LnkSta lines with counts
	aggregatedOutput := AggregateLinkStatistics(filteredOutput)
	
	parseResult := parseLspciWidthOutput(aggregatedOutput, expectedLinkState)
	
	return PCIeWidthResult{
		WidthCounts: parseResult.WidthCounts,
		SpeedCounts: parseResult.SpeedCounts,
		StateErrors: parseResult.StateErrors,
		Success:     true,
	}
}

// validateWidthCounts compares actual vs expected width counts
func validateWidthCounts(actual, expected map[string]int, deviceType string) (bool, string) {
	for width, expectedCount := range expected {
		actualCount, exists := actual[width]
		if !exists {
			actualCount = 0
		}
		
		if actualCount != expectedCount {
			return false, fmt.Sprintf("%s PCIe width mismatch: expected %dx %s, got %dx %s", 
				deviceType, expectedCount, width, actualCount, width)
		}
	}
	
	return true, ""
}

// validateSpeedCounts compares actual vs expected speed counts
func validateSpeedCounts(actual, expected map[string]int, deviceType string) (bool, string) {
	for speed, expectedCount := range expected {
		actualCount, exists := actual[speed]
		if !exists {
			actualCount = 0
		}
		
		if actualCount != expectedCount {
			return false, fmt.Sprintf("%s PCIe speed mismatch: expected %dx %s, got %dx %s", 
				deviceType, expectedCount, speed, actualCount, speed)
		}
	}
	
	return true, ""
}

// RunPCIeWidthMissingLanesCheck performs the PCIe width missing lanes diagnostic
func RunPCIeWidthMissingLanesCheck() error {
	logger.Info("=== PCIe Width Missing Lanes Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("PCIe Width Missing Lanes Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddPCIeWidthResult("FAIL", nil, nil, nil, nil, nil, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Load test configuration
	config, err := getPcieWidthMissingLanesTestConfig(shape)
	if err != nil {
		logger.Error("PCIe Width Missing Lanes Check: FAIL - Could not load test configuration:", err)
		rep.AddPCIeWidthResult("FAIL", nil, nil, nil, nil, nil, err)
		return fmt.Errorf("failed to load test configuration: %w", err)
	}

	if !config.IsEnabled {
		errorMsg := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Error(errorMsg)
		rep.AddPCIeWidthResult("SKIP", nil, nil, nil, nil, nil, fmt.Errorf(errorMsg))
		return fmt.Errorf(errorMsg)
	}

	// Step 3: Check GPU/NVSwitch PCIe width, speed, and state
	logger.Info("Step 2: Checking GPU/NVSwitch PCIe width, speed, and state...")
	gpuResult := checkGPUNVSwitchPCIeWidth(config.ExpectedLinkState)
	
	var errorMessages []string
	var allGPUWidths map[string]int
	var allRDMAWidths map[string]int
	var allGPUSpeeds map[string]int
	var allRDMASpeeds map[string]int
	var allStateErrors []string

	if !gpuResult.Success {
		errorMessages = append(errorMessages, fmt.Sprintf("GPU/NVSwitch: %s", gpuResult.ErrorMsg))
		allGPUWidths = make(map[string]int)
		allGPUSpeeds = make(map[string]int)
	} else {
		allGPUWidths = gpuResult.WidthCounts
		allGPUSpeeds = gpuResult.SpeedCounts
		allStateErrors = append(allStateErrors, gpuResult.StateErrors...)
		
		// Validate GPU widths
		if len(config.ExpectedGPUWidths) > 0 {
			gpuValid, gpuError := validateWidthCounts(allGPUWidths, config.ExpectedGPUWidths, "GPU/NVSwitch")
			if !gpuValid {
				errorMessages = append(errorMessages, gpuError)
			}
		}
		
		// Validate GPU speeds
		if len(config.ExpectedGPUSpeeds) > 0 {
			gpuSpeedValid, gpuSpeedError := validateSpeedCounts(gpuResult.SpeedCounts, config.ExpectedGPUSpeeds, "GPU/NVSwitch")
			if !gpuSpeedValid {
				errorMessages = append(errorMessages, gpuSpeedError)
			}
		}
		
		// Check GPU state errors
		if len(gpuResult.StateErrors) > 0 {
			for _, stateError := range gpuResult.StateErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("GPU/NVSwitch state error: %s", stateError))
			}
		}
	}

	// Step 4: Check RDMA PCIe width, speed, and state
	logger.Info("Step 3: Checking RDMA PCIe width, speed, and state...")
	rdmaResult := checkRDMAPCIeWidth(config.ExpectedLinkState)
	
	if !rdmaResult.Success {
		errorMessages = append(errorMessages, fmt.Sprintf("RDMA: %s", rdmaResult.ErrorMsg))
		allRDMAWidths = make(map[string]int)
		allRDMASpeeds = make(map[string]int)
	} else {
		allRDMAWidths = rdmaResult.WidthCounts
		allRDMASpeeds = rdmaResult.SpeedCounts
		allStateErrors = append(allStateErrors, rdmaResult.StateErrors...)
		
		// Validate RDMA widths
		if len(config.ExpectedRDMAWidths) > 0 {
			rdmaValid, rdmaError := validateWidthCounts(allRDMAWidths, config.ExpectedRDMAWidths, "RDMA")
			if !rdmaValid {
				errorMessages = append(errorMessages, rdmaError)
			}
		}
		
		// Validate RDMA speeds
		if len(config.ExpectedRDMASpeeds) > 0 {
			rdmaSpeedValid, rdmaSpeedError := validateSpeedCounts(rdmaResult.SpeedCounts, config.ExpectedRDMASpeeds, "RDMA")
			if !rdmaSpeedValid {
				errorMessages = append(errorMessages, rdmaSpeedError)
			}
		}
		
		// Check RDMA state errors
		if len(rdmaResult.StateErrors) > 0 {
			for _, stateError := range rdmaResult.StateErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("RDMA state error: %s", stateError))
			}
		}
	}

	// Step 5: Generate final result
	if len(errorMessages) == 0 {
		logger.Info("PCIe Width Missing Lanes Check: PASS - All PCIe interfaces operating at expected width and speed")
		rep.AddPCIeWidthResult("PASS", allGPUWidths, allRDMAWidths, allGPUSpeeds, allRDMASpeeds, allStateErrors, nil)
		return nil
	} else {
		// Combine error messages
		combinedError := strings.Join(errorMessages, "; ") + ". Please reboot the host and if the issue persists, send the node to OCI."
		logger.Error("PCIe Width Missing Lanes Check: FAIL -", combinedError)
		
		err := fmt.Errorf(combinedError)
		rep.AddPCIeWidthResult("FAIL", allGPUWidths, allRDMAWidths, allGPUSpeeds, allRDMASpeeds, allStateErrors, err)
		return err
	}
}