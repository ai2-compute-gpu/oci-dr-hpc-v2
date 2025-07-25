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

// MaxAccCheckTestConfig represents the config needed to run this test
type MaxAccCheckTestConfig struct {
	IsEnabled bool     `json:"enabled"`
	Shape     string   `json:"shape"`
	PCIIDs    []string `json:"pci_ids"`
}

// PCIEConfig represents the PCIe configuration for a single device
type PCIEConfig struct {
	PCIBusID           string `json:"pci_busid"`
	MaxAccOut          string `json:"max_acc_out"`
	AdvancedPCISettings string `json:"advanced_pci_settings"`
}

// MaxAccCheckResult represents the result from max_acc_check
type MaxAccCheckResult struct {
	PCIEConfig []PCIEConfig `json:"pcie_config"`
}

// getMaxAccCheckTestConfig gets test config needed to run this test
func getMaxAccCheckTestConfig() (*MaxAccCheckTestConfig, error) {
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

	// Initialize with defaults (H100 PCI IDs as per Python script)
	maxAccCheckTestConfig := &MaxAccCheckTestConfig{
		IsEnabled: false,
		Shape:     shape,
		PCIIDs: []string{
			"0000:0c:00.0",
			"0000:2a:00.0",
			"0000:41:00.0",
			"0000:58:00.0",
			"0000:86:00.0",
			"0000:a5:00.0",
			"0000:bd:00.0",
			"0000:d5:00.0",
		},
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "max_acc_check")
	if err != nil {
		return nil, err
	}
	maxAccCheckTestConfig.IsEnabled = enabled

	// Get PCI IDs configuration if available
	threshold, err := limits.GetThresholdForTest(shape, "max_acc_check")
	if err == nil {
		switch v := threshold.(type) {
		case map[string]interface{}:
			if pciIDs, ok := v["pci_ids"].([]interface{}); ok {
				var pciIDStrings []string
				for _, pciID := range pciIDs {
					if pciIDStr, ok := pciID.(string); ok {
						pciIDStrings = append(pciIDStrings, pciIDStr)
					}
				}
				if len(pciIDStrings) > 0 {
					maxAccCheckTestConfig.PCIIDs = pciIDStrings
				}
			}
		}
	}

	return maxAccCheckTestConfig, nil
}

// runMLXConfig runs mlxconfig query for a specific PCI device
func runMLXConfig(pciID string) ([]string, error) {
	mlxconfigBin := "/usr/bin/mlxconfig"
	cmd := exec.Command("sudo", mlxconfigBin, "-d", pciID, "query")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("mlxconfig command failed for %s: %w, output: %s", pciID, err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

// parseAccResults parses mlxconfig output for a specific PCI device
func parseAccResults(pciID string, results []string) PCIEConfig {
	config := PCIEConfig{
		PCIBusID:            pciID,
		MaxAccOut:           "FAIL",
		AdvancedPCISettings: "FAIL",
	}

	for _, line := range results {
		line = strings.TrimSpace(line)
		
		// Check MAX_ACC_OUT_READ - must be 0, 44, or 128
		if strings.Contains(line, "MAX_ACC_OUT_READ") {
			if strings.Contains(line, "0") || strings.Contains(line, "44") || strings.Contains(line, "128") {
				config.MaxAccOut = "PASS"
			}
		}
		
		// Check ADVANCED_PCI_SETTINGS - must be True
		if strings.Contains(line, "ADVANCED_PCI_SETTINGS") && strings.Contains(line, "True") {
			config.AdvancedPCISettings = "PASS"
		}
	}

	return config
}

// runMaxAccCheck performs the max_acc_check validation for all configured PCI devices
func runMaxAccCheck(config *MaxAccCheckTestConfig) (*MaxAccCheckResult, error) {
	var pcieConfigs []PCIEConfig

	for _, pciID := range config.PCIIDs {
		logger.Info(fmt.Sprintf("Checking PCI device: %s", pciID))
		
		output, err := runMLXConfig(pciID)
		if err != nil {
			logger.Errorf("Failed to query PCI device %s: %v", pciID, err)
			// Add failed config for this device
			pcieConfigs = append(pcieConfigs, PCIEConfig{
				PCIBusID:            pciID,
				MaxAccOut:           "FAIL",
				AdvancedPCISettings: "FAIL",
			})
			continue
		}

		pcieConfig := parseAccResults(pciID, output)
		pcieConfigs = append(pcieConfigs, pcieConfig)
	}

	return &MaxAccCheckResult{
		PCIEConfig: pcieConfigs,
	}, nil
}

// validateMaxAccResults validates the max_acc_check results
func validateMaxAccResults(result *MaxAccCheckResult) (string, string, error) {
	if result == nil || len(result.PCIEConfig) == 0 {
		return "FAIL", "No PCI devices found", fmt.Errorf("no PCI devices found")
	}

	var failedDevices []string
	var failureReasons []string

	for _, config := range result.PCIEConfig {
		deviceFailed := false
		
		if config.MaxAccOut == "FAIL" {
			failedDevices = append(failedDevices, config.PCIBusID)
			failureReasons = append(failureReasons, fmt.Sprintf("%s: MAX_ACC_OUT_READ invalid", config.PCIBusID))
			deviceFailed = true
		}
		
		if config.AdvancedPCISettings == "FAIL" {
			if !deviceFailed {
				failedDevices = append(failedDevices, config.PCIBusID)
			}
			failureReasons = append(failureReasons, fmt.Sprintf("%s: ADVANCED_PCI_SETTINGS not True", config.PCIBusID))
		}
	}

	if len(failedDevices) > 0 {
		message := fmt.Sprintf("Failed devices: %s", strings.Join(failureReasons, ", "))
		return "FAIL", message, fmt.Errorf("max_acc_check failed for devices: %s", strings.Join(failedDevices, ", "))
	}

	successMessage := fmt.Sprintf("All %d PCI devices configured correctly", len(result.PCIEConfig))
	return "PASS", successMessage, nil
}

// RunMaxAccCheck performs the max_acc_check test
func RunMaxAccCheck() error {
	logger.Info("=== MAX_ACC_OUT_READ Configuration Check ===")
	testConfig, err := getMaxAccCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting MAX_ACC_OUT_READ configuration check...")
	rep := reporter.GetReporter()

	// Step 1: Check mlxconfig availability
	logger.Info("Step 1: Checking mlxconfig availability...")
	if _, err := exec.LookPath("/usr/bin/mlxconfig"); err != nil {
		logger.Error("MAX_ACC Check: FAIL - mlxconfig not found")
		rep.AddMaxAccResult("FAIL", nil, fmt.Errorf("mlxconfig not found: %w", err))
		return fmt.Errorf("mlxconfig not found: %w", err)
	}

	// Step 2: Run max_acc_check for all PCI devices
	logger.Info("Step 2: Checking PCI device configurations...")
	logger.Info(fmt.Sprintf("Checking %d PCI devices: %v", len(testConfig.PCIIDs), testConfig.PCIIDs))

	result, err := runMaxAccCheck(testConfig)
	if err != nil {
		logger.Error("MAX_ACC Check: FAIL - Could not check device configurations:", err)
		rep.AddMaxAccResult("FAIL", nil, err)
		return fmt.Errorf("could not check device configurations: %w", err)
	}

	// Step 3: Validate results
	logger.Info("Step 3: Validating device configurations...")
	status, statusMsg, validationErr := validateMaxAccResults(result)

	switch status {
	case "PASS":
		logger.Info("MAX_ACC Check: PASS -", statusMsg)
		rep.AddMaxAccResult("PASS", result, nil)
		return nil
	default: // FAIL
		logger.Error("MAX_ACC Check: FAIL -", statusMsg)
		rep.AddMaxAccResult("FAIL", result, validationErr)
		return validationErr
	}
}

// PrintMaxAccCheck prints information about the max_acc_check test
func PrintMaxAccCheck() {
	logger.Info("MAX_ACC Check: Checking MAX_ACC_OUT_READ and ADVANCED_PCI_SETTINGS configuration...")
	logger.Info("MAX_ACC Check: PASS - All PCI devices configured correctly")
}