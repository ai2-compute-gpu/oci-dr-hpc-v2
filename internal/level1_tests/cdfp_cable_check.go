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

// CDFPCableCheckResult represents the result of CDFP cable check
type CDFPCableCheckResult struct {
	Status             string            `json:"status"`
	ExpectedMapping    map[string]string `json:"expected_mapping"`
	ActualMapping      map[string]string `json:"actual_mapping"`
	Failures           []string          `json:"failures,omitempty"`
	Message            string            `json:"message"`
}

// CDFPCableCheckTestConfig represents the test configuration for CDFP cable check
type CDFPCableCheckTestConfig struct {
	IsEnabled       bool     `json:"enabled"`
	ExpectedPCIIDs  []string `json:"gpu_pci_ids"`
	ExpectedIndices []string `json:"gpu_indices"`
}

// getCDFPCableCheckTestConfig gets test config needed to run this test
func getCDFPCableCheckTestConfig(shape string) (*CDFPCableCheckTestConfig, error) {
	// Load configuration from test_limits.json to check if test is enabled
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults
	cdfpTestConfig := &CDFPCableCheckTestConfig{
		IsEnabled:       false,
		ExpectedPCIIDs:  []string{},
		ExpectedIndices: []string{},
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "cdfp_cable_check")
	if err != nil {
		logger.Info("CDFP cable check test not found for shape", shape, ", defaulting to disabled")
		return cdfpTestConfig, nil
	}
	cdfpTestConfig.IsEnabled = enabled

	// If test is disabled, return early
	if !enabled {
		logger.Info("CDFP cable check test disabled for shape", shape)
		return cdfpTestConfig, nil
	}

	// Load GPU configuration from shapes.json
	shapeManager, err := shapes.GetDefaultShapeManager()
	if err != nil {
		logger.Info("Failed to load shapes configuration:", err, ", using empty arrays")
		return cdfpTestConfig, nil
	}

	// Check if shape has GPU configurations
	hasGPUs, err := shapeManager.HasGPUs(shape)
	if err != nil {
		logger.Info("Failed to check GPU availability for shape", shape, ":", err, ", using empty arrays")
		return cdfpTestConfig, nil
	}

	if !hasGPUs {
		logger.Info("Shape", shape, "has no GPU configurations")
		return cdfpTestConfig, nil
	}

	// Get GPU PCI addresses from shapes.json
	pciAddresses, err := shapeManager.GetGPUPCIAddresses(shape)
	if err != nil {
		logger.Info("Failed to get GPU PCI addresses for shape", shape, ":", err, ", using empty arrays")
		return cdfpTestConfig, nil
	}

	// Get GPU indices from shapes.json
	indices, err := shapeManager.GetGPUIndices(shape)
	if err != nil {
		logger.Info("Failed to get GPU indices for shape", shape, ":", err, ", using empty arrays")
		return cdfpTestConfig, nil
	}

	// Ensure we have GPU configurations
	if len(pciAddresses) == 0 || len(indices) == 0 {
		logger.Info("No GPU PCI addresses or indices found for shape", shape)
		return cdfpTestConfig, nil
	}

	// Normalize PCI addresses to match nvidia-smi output format
	for _, pci := range pciAddresses {
		// Normalize the PCI address to match the format from nvidia-smi
		normalizedPCI := normalizePCIAddress(pci)
		cdfpTestConfig.ExpectedPCIIDs = append(cdfpTestConfig.ExpectedPCIIDs, normalizedPCI)
	}

	cdfpTestConfig.ExpectedIndices = indices

	logger.Info("Successfully loaded cdfp_cable_check configuration for shape", shape, "from shapes.json")
	logger.Info("Found", len(cdfpTestConfig.ExpectedPCIIDs), "GPU PCI addresses and", len(cdfpTestConfig.ExpectedIndices), "GPU indices")
	return cdfpTestConfig, nil
}

// normalizePCIAddress normalizes PCI address format
func normalizePCIAddress(pciAddr string) string {
	normalized := strings.ToLower(pciAddr)
	// Handle cases where PCI address starts with "000000"
	if strings.HasPrefix(pciAddr, "000000") {
		normalized = "00" + strings.ToLower(pciAddr[6:])
	}
	return normalized
}

// parseGPUInfo parses GPU PCI addresses and indices from nvidia-smi -q output
func parseGPUInfo() ([]string, []string, error) {
	// Use nvidia-smi -q to get detailed GPU information
	queryResult := executor.RunNvidiaSMIQueryDetailed()
	if !queryResult.Available || queryResult.Error != "" {
		return nil, nil, fmt.Errorf("failed to get GPU information: %s", queryResult.Error)
	}

	var pciAddresses []string
	var gpuIndices []string
	currentGPUIndex := 0

	// Parse the output to extract GPU Bus IDs in order
	lines := strings.Split(queryResult.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "bus id") && strings.Contains(line, ":") {
			// Extract Bus ID from lines like "    Bus Id                        : 00000000:0F:00.0"
			parts := strings.Split(line, ":")
			if len(parts) >= 4 {
				busID := strings.TrimSpace(parts[len(parts)-3]) + ":" + 
						strings.TrimSpace(parts[len(parts)-2]) + ":" + 
						strings.TrimSpace(parts[len(parts)-1])
				if busID != "::" && busID != "" {
					pciAddresses = append(pciAddresses, normalizePCIAddress(busID))
					gpuIndices = append(gpuIndices, fmt.Sprintf("%d", currentGPUIndex))
					currentGPUIndex++
				}
			}
		}
	}

	if len(pciAddresses) == 0 {
		return nil, nil, fmt.Errorf("no GPU information found in nvidia-smi -q output")
	}

	return pciAddresses, gpuIndices, nil
}

// validateCDFPCables validates CDFP cable connections
func validateCDFPCables(expectedPCIs, expectedIndices, actualPCIs, actualIndices []string) *CDFPCableCheckResult {
	result := &CDFPCableCheckResult{
		Status:          "PASS",
		ExpectedMapping: make(map[string]string),
		ActualMapping:   make(map[string]string),
		Failures:        []string{},
	}

	// Create expected mapping
	for i, pci := range expectedPCIs {
		if i < len(expectedIndices) {
			normalizedPCI := normalizePCIAddress(pci)
			result.ExpectedMapping[normalizedPCI] = expectedIndices[i]
		}
	}

	// Create actual mapping
	for i, pci := range actualPCIs {
		if i < len(actualIndices) {
			result.ActualMapping[pci] = actualIndices[i]
		}
	}

	// Validate each expected PCI and index pair
	for expectedPCI, expectedIndex := range result.ExpectedMapping {
		if actualIndex, exists := result.ActualMapping[expectedPCI]; !exists {
			result.Failures = append(result.Failures, 
				fmt.Sprintf("Expected GPU with PCI Address %s not found", expectedPCI))
		} else if actualIndex != expectedIndex {
			result.Failures = append(result.Failures, 
				fmt.Sprintf("Mismatch for PCI %s: Expected GPU index %s, found %s", 
					expectedPCI, expectedIndex, actualIndex))
		}
	}

	// Set status and message based on failures
	if len(result.Failures) > 0 {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("CDFP cable mismatch detected: %s", strings.Join(result.Failures, ", "))
	} else {
		result.Status = "PASS"
		result.Message = "All CDFP cables correctly connected"
	}

	return result
}

// RunCDFPCableCheck performs the CDFP cable check for GPUs
func RunCDFPCableCheck() error {
	logger.Info("=== CDFP Cable Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("CDFP Cable Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get shape from IMDS: %v", err),
		}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	cdfpTestConfig, err := getCDFPCableCheckTestConfig(shape)
	if err != nil {
		logger.Error("CDFP Cable Check: FAIL - Could not get test configuration:", err)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get test configuration: %v", err),
		}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !cdfpTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Validate expected configuration
	if len(cdfpTestConfig.ExpectedPCIIDs) == 0 || len(cdfpTestConfig.ExpectedIndices) == 0 {
		errorStatement := fmt.Sprintf("No expected PCI IDs or GPU indices configured for shape %s", shape)
		logger.Info("CDFP Cable Check: SKIP -", errorStatement)
		rep.AddCDFPCableCheckResult("SKIP", &CDFPCableCheckResult{
			Status:  "SKIP",
			Message: errorStatement,
		}, nil)
		return nil
	}

	if len(cdfpTestConfig.ExpectedPCIIDs) != len(cdfpTestConfig.ExpectedIndices) {
		errorStatement := "Mismatch between expected PCI IDs and GPU indices count in configuration"
		logger.Error("CDFP Cable Check: FAIL -", errorStatement)
		err = fmt.Errorf(errorStatement)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: errorStatement,
		}, err)
		return err
	}

	logger.Info("Expected PCI IDs:", cdfpTestConfig.ExpectedPCIIDs)
	logger.Info("Expected GPU Indices:", cdfpTestConfig.ExpectedIndices)

	// Step 4: Get actual GPU information
	logger.Info("Step 4: Getting actual GPU information...")
	actualPCIs, actualIndices, err := parseGPUInfo()
	if err != nil {
		logger.Error("CDFP Cable Check: FAIL - Could not get GPU information:", err)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get GPU information: %v", err),
		}, err)
		return fmt.Errorf("failed to get GPU information: %w", err)
	}

	logger.Info("Actual PCI addresses:", actualPCIs)
	logger.Info("Actual GPU indices:", actualIndices)

	// Step 5: Validate CDFP cables
	logger.Info("Step 5: Validating CDFP cables...")
	result := validateCDFPCables(cdfpTestConfig.ExpectedPCIIDs, cdfpTestConfig.ExpectedIndices, 
		actualPCIs, actualIndices)

	// Step 6: Report results
	logger.Info("Step 6: Reporting results...")
	if result.Status == "PASS" {
		logger.Info("CDFP Cable Check: PASS -", result.Message)
		rep.AddCDFPCableCheckResult("PASS", result, nil)
		return nil
	} else {
		logger.Error("CDFP Cable Check: FAIL -", result.Message)
		err = fmt.Errorf(result.Message)
		rep.AddCDFPCableCheckResult("FAIL", result, err)
		return err
	}
}