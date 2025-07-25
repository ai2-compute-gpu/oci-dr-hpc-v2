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
	IsEnabled         bool     `json:"enabled"`
	ExpectedPCIIDs    []string `json:"gpu_pci_ids"`
	ExpectedModuleIDs []string `json:"gpu_module_ids"`
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
		IsEnabled:         false,
		ExpectedPCIIDs:    []string{},
		ExpectedModuleIDs: []string{},
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

	// Get GPU module IDs from shapes.json
	moduleIDs, err := shapeManager.GetGPUModuleIDs(shape)
	if err != nil {
		logger.Info("Failed to get GPU module IDs for shape", shape, ":", err, ", using empty arrays")
		return cdfpTestConfig, nil
	}

	// Ensure we have GPU configurations
	if len(pciAddresses) == 0 || len(moduleIDs) == 0 {
		logger.Info("No GPU PCI addresses or module IDs found for shape", shape)
		return cdfpTestConfig, nil
	}

	// Normalize PCI addresses to match nvidia-smi output format
	for _, pci := range pciAddresses {
		// Normalize the PCI address to match the format from nvidia-smi
		normalizedPCI := normalizePCIAddress(pci)
		cdfpTestConfig.ExpectedPCIIDs = append(cdfpTestConfig.ExpectedPCIIDs, normalizedPCI)
	}

	cdfpTestConfig.ExpectedModuleIDs = moduleIDs

	logger.Info("Successfully loaded cdfp_cable_check configuration for shape", shape, "from shapes.json")
	logger.Info("Found", len(cdfpTestConfig.ExpectedPCIIDs), "GPU PCI addresses and", len(cdfpTestConfig.ExpectedModuleIDs), "GPU module IDs")
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

// parseGPUInfo parses GPU PCI addresses and module IDs from nvidia-smi -q output
func parseGPUInfo() ([]string, []string, error) {
	// Use nvidia-smi -q to get detailed GPU information
	queryResult := executor.RunNvidiaSMIQueryDetailed()
	if !queryResult.Available || queryResult.Error != "" {
		return nil, nil, fmt.Errorf("failed to get detailed GPU information: %s", queryResult.Error)
	}

	var pciAddresses []string
	var moduleIDs []string
	
	lines := strings.Split(queryResult.Output, "\n")
	
	// First pass: collect all PCI addresses in order
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for Bus ID lines - format: "Bus Id                        : 00000000:0F:00.0"
		if strings.Contains(strings.ToLower(line), "bus id") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 4 {
				busID := strings.TrimSpace(parts[len(parts)-3]) + ":" + 
					strings.TrimSpace(parts[len(parts)-2]) + ":" + 
					strings.TrimSpace(parts[len(parts)-1])
				if busID != "::" && busID != "" {
					pciAddresses = append(pciAddresses, normalizePCIAddress(busID))
				}
			}
		}
	}
	
	// Second pass: collect all module IDs in order
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for Module ID lines - format may vary: "Module Id                     : 2"
		if strings.Contains(strings.ToLower(line), "module id") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				moduleID := strings.TrimSpace(parts[len(parts)-1])
				if moduleID != "" && moduleID != "N/A" && moduleID != "[Not Supported]" {
					moduleIDs = append(moduleIDs, moduleID)
				}
			}
		}
	}

	if len(pciAddresses) == 0 {
		return nil, nil, fmt.Errorf("no GPU PCI addresses found in nvidia-smi -q output")
	}
	
	// If no module IDs found, use sequential numbering
	if len(moduleIDs) == 0 {
		logger.Info("No module IDs found, using sequential numbering")
		for i := 0; i < len(pciAddresses); i++ {
			moduleIDs = append(moduleIDs, fmt.Sprintf("%d", i+1))
		}
	}
	
	// Ensure we have matching counts
	if len(pciAddresses) != len(moduleIDs) {
		return nil, nil, fmt.Errorf("mismatch between PCI address count (%d) and module ID count (%d)", 
			len(pciAddresses), len(moduleIDs))
	}

	logger.Info("Successfully parsed", len(pciAddresses), "GPUs from nvidia-smi -q")
	logger.Info("PCI addresses:", pciAddresses)
	logger.Info("Module IDs:", moduleIDs)
	return pciAddresses, moduleIDs, nil
}

// validateCDFPCables validates CDFP cable connections
func validateCDFPCables(expectedPCIs, expectedModuleIDs, actualPCIs, actualModuleIDs []string) *CDFPCableCheckResult {
	result := &CDFPCableCheckResult{
		Status:          "PASS",
		ExpectedMapping: make(map[string]string),
		ActualMapping:   make(map[string]string),
		Failures:        []string{},
	}

	// Create expected mapping
	for i, pci := range expectedPCIs {
		if i < len(expectedModuleIDs) {
			normalizedPCI := normalizePCIAddress(pci)
			result.ExpectedMapping[normalizedPCI] = expectedModuleIDs[i]
		}
	}

	// Create actual mapping
	for i, pci := range actualPCIs {
		if i < len(actualModuleIDs) {
			result.ActualMapping[pci] = actualModuleIDs[i]
		}
	}

	// Validate each expected PCI and module ID pair
	for expectedPCI, expectedModuleID := range result.ExpectedMapping {
		if actualModuleID, exists := result.ActualMapping[expectedPCI]; !exists {
			result.Failures = append(result.Failures, 
				fmt.Sprintf("Expected GPU with PCI Address %s not found", expectedPCI))
		} else if actualModuleID != expectedModuleID {
			result.Failures = append(result.Failures, 
				fmt.Sprintf("Mismatch for PCI %s: Expected GPU module ID %s, found %s", 
					expectedPCI, expectedModuleID, actualModuleID))
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
	if len(cdfpTestConfig.ExpectedPCIIDs) == 0 || len(cdfpTestConfig.ExpectedModuleIDs) == 0 {
		errorStatement := fmt.Sprintf("No expected PCI IDs or GPU module IDs configured for shape %s", shape)
		logger.Info("CDFP Cable Check: SKIP -", errorStatement)
		rep.AddCDFPCableCheckResult("SKIP", &CDFPCableCheckResult{
			Status:  "SKIP",
			Message: errorStatement,
		}, nil)
		return nil
	}

	if len(cdfpTestConfig.ExpectedPCIIDs) != len(cdfpTestConfig.ExpectedModuleIDs) {
		errorStatement := "Mismatch between expected PCI IDs and GPU module IDs count in configuration"
		logger.Error("CDFP Cable Check: FAIL -", errorStatement)
		err = fmt.Errorf(errorStatement)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: errorStatement,
		}, err)
		return err
	}

	logger.Info("Expected PCI IDs:", cdfpTestConfig.ExpectedPCIIDs)
	logger.Info("Expected GPU Module IDs:", cdfpTestConfig.ExpectedModuleIDs)

	// Step 4: Get actual GPU information
	logger.Info("Step 4: Getting actual GPU information...")
	actualPCIs, actualModuleIDs, err := parseGPUInfo()
	if err != nil {
		logger.Error("CDFP Cable Check: FAIL - Could not get GPU information:", err)
		rep.AddCDFPCableCheckResult("FAIL", &CDFPCableCheckResult{
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to get GPU information: %v", err),
		}, err)
		return fmt.Errorf("failed to get GPU information: %w", err)
	}

	logger.Info("Actual PCI addresses:", actualPCIs)
	logger.Info("Actual GPU module IDs:", actualModuleIDs)

	// Step 5: Validate CDFP cables
	logger.Info("Step 5: Validating CDFP cables...")
	result := validateCDFPCables(cdfpTestConfig.ExpectedPCIIDs, cdfpTestConfig.ExpectedModuleIDs, 
		actualPCIs, actualModuleIDs)

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