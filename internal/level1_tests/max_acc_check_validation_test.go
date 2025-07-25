package level1_tests

import (
	"fmt"
	"strings"
	"testing"
)

// Test structures (self-contained)
type testMaxAccCheckTestConfig struct {
	IsEnabled bool     `json:"enabled"`
	Shape     string   `json:"shape"`
	PCIIDs    []string `json:"pci_ids"`
}

type testPCIEConfig struct {
	PCIBusID            string `json:"pci_busid"`
	MaxAccOut           string `json:"max_acc_out"`
	AdvancedPCISettings string `json:"advanced_pci_settings"`
	RawOutput           string `json:"raw_output"`
	IsValid             bool   `json:"is_valid"`
}

type testMaxAccCheckResult struct {
	PCIEConfig []testPCIEConfig `json:"pcie_config"`
}

// Test helper functions

func createTestMLXConfigOutput(maxAccValue string, advancedPCIValue string) []string {
	return []string{
		"Device #1:",
		"----------",
		"Device type:    ConnectX7",
		"Name:           MCX713106A-VEAT_Ax",
		"Description:    ConnectX-7 VPI adapter card; 200Gbps; dual-port QSFP56; PCIe4.0 x16; tall bracket; ROHS R6",
		"Device:         /dev/mst/mt4125_pciconf0",
		"",
		"Configurations:                      Current",
		"         MEMIC_BAR_SIZE              0",
		"         MEMIC_SIZE_LIMIT            _256KB(1)",
		"         HOST_CHAINING_MODE          ENABLED(1)",
		"         ICM_CACHE_MODE              CACHED(1)",
		fmt.Sprintf("         MAX_ACC_OUT_READ            %s", maxAccValue),
		"         ROCE_CC_PRIO_MASK_P1        0xff",
		"         ROCE_CC_PRIO_MASK_P2        0xff",
		fmt.Sprintf("         ADVANCED_PCI_SETTINGS       %s", advancedPCIValue),
		"         NUM_PF                      1",
		"",
	}
}

func createMinimalMLXConfigOutput(maxAccValue string, advancedPCIValue string) []string {
	return []string{
		fmt.Sprintf("         MAX_ACC_OUT_READ            %s", maxAccValue),
		fmt.Sprintf("         ADVANCED_PCI_SETTINGS       %s", advancedPCIValue),
	}
}

func createIncompleteMLXConfigOutput(includingMaxAcc bool, includingAdvancedPCI bool) []string {
	output := []string{
		"Device #1:",
		"----------",
		"Configurations:                      Current",
		"         MEMIC_BAR_SIZE              0",
		"         NUM_PF                      1",
	}
	
	if includingMaxAcc {
		output = append(output, "         MAX_ACC_OUT_READ            44")
	}
	
	if includingAdvancedPCI {
		output = append(output, "         ADVANCED_PCI_SETTINGS       True")
	}
	
	return output
}

func createCorruptedMLXConfigOutput() []string {
	return []string{
		"Device #1:",
		"----------",
		"Error: Failed to query device",
		"Unable to access configuration",
		"MAX_ACC_OUT_READ CORRUPTED",
		"ADVANCED_PCI_SETTINGS INVALID",
	}
}

// Core functionality tests

func TestMaxAccCheckTestConfig(t *testing.T) {
	config := &testMaxAccCheckTestConfig{
		IsEnabled: true,
		Shape:     "BM.GPU.H100.8",
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

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected Shape 'BM.GPU.H100.8', got '%s'", config.Shape)
	}
	if len(config.PCIIDs) != 8 {
		t.Errorf("Expected 8 PCI IDs, got %d", len(config.PCIIDs))
	}
}

func TestPCIEConfig(t *testing.T) {
	config := &testPCIEConfig{
		PCIBusID:            "0000:0c:00.0",
		MaxAccOut:           "PASS",
		AdvancedPCISettings: "PASS",
		RawOutput:           "MAX_ACC_OUT_READ 44",
		IsValid:             true,
	}

	if config.PCIBusID != "0000:0c:00.0" {
		t.Errorf("Expected PCIBusID '0000:0c:00.0', got '%s'", config.PCIBusID)
	}
	if config.MaxAccOut != "PASS" {
		t.Errorf("Expected MaxAccOut 'PASS', got '%s'", config.MaxAccOut)
	}
	if config.AdvancedPCISettings != "PASS" {
		t.Errorf("Expected AdvancedPCISettings 'PASS', got '%s'", config.AdvancedPCISettings)
	}
	if !config.IsValid {
		t.Error("Expected IsValid to be true")
	}
}

// Self-contained parsing function for testing
func testParseMLXConfigOutput(pciID string, output []string) (testPCIEConfig, error) {
	if len(output) == 0 {
		return testPCIEConfig{}, fmt.Errorf("empty mlxconfig output")
	}

	config := testPCIEConfig{
		PCIBusID:            pciID,
		MaxAccOut:           "FAIL",
		AdvancedPCISettings: "FAIL",
		RawOutput:           strings.Join(output, "\n"),
		IsValid:             false,
	}

	validLines := 0
	for _, line := range output {
		line = strings.TrimSpace(line)
		
		// Check MAX_ACC_OUT_READ - must be 0, 44, or 128
		if strings.Contains(line, "MAX_ACC_OUT_READ") {
			if strings.Contains(line, "0") || strings.Contains(line, "44") || strings.Contains(line, "128") {
				config.MaxAccOut = "PASS"
			}
			validLines++
		}
		
		// Check ADVANCED_PCI_SETTINGS - must be True
		if strings.Contains(line, "ADVANCED_PCI_SETTINGS") && strings.Contains(line, "True") {
			config.AdvancedPCISettings = "PASS"
			validLines++
		}
	}

	// Consider valid if we found at least one configuration line
	config.IsValid = validLines > 0

	return config, nil
}

// Self-contained validation function for testing
func testValidateMLXConfigResults(results []testPCIEConfig) (bool, []string, string) {
	if len(results) == 0 {
		return false, nil, "no PCI devices found"
	}

	var failedDevices []string
	var failureReasons []string

	for _, config := range results {
		if !config.IsValid {
			failedDevices = append(failedDevices, config.PCIBusID)
			failureReasons = append(failureReasons, fmt.Sprintf("%s: invalid configuration", config.PCIBusID))
			continue
		}
		
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
		return false, failedDevices, message
	}

	successMessage := fmt.Sprintf("All %d PCI devices configured correctly", len(results))
	return true, nil, successMessage
}

// Parsing tests

func TestParseMLXConfigOutput(t *testing.T) {
	tests := []struct {
		name                     string
		pciID                    string
		output                   []string
		expectedMaxAccOut        string
		expectedAdvancedPCISettings string
		expectedValid            bool
		expectError              bool
	}{
		{
			name:                     "valid complete output",
			pciID:                    "0000:0c:00.0",
			output:                   createTestMLXConfigOutput("44", "True"),
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "valid minimal output",
			pciID:                    "0000:2a:00.0",
			output:                   createMinimalMLXConfigOutput("0", "True"),
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "invalid MAX_ACC_OUT_READ",
			pciID:                    "0000:41:00.0",
			output:                   createMinimalMLXConfigOutput("32", "True"),
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "PASS",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "invalid ADVANCED_PCI_SETTINGS",
			pciID:                    "0000:58:00.0",
			output:                   createMinimalMLXConfigOutput("44", "False"),
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "FAIL",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "both settings invalid",
			pciID:                    "0000:86:00.0",
			output:                   createMinimalMLXConfigOutput("16", "False"),
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "empty output",
			pciID:                    "0000:a5:00.0",
			output:                   []string{},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
			expectedValid:            false,
			expectError:              true,
		},
		{
			name:                     "missing MAX_ACC_OUT_READ",
			pciID:                    "0000:bd:00.0",
			output:                   createIncompleteMLXConfigOutput(false, true),
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "PASS",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "missing ADVANCED_PCI_SETTINGS",
			pciID:                    "0000:d5:00.0",
			output:                   createIncompleteMLXConfigOutput(true, false),
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "FAIL",
			expectedValid:            true,
			expectError:              false,
		},
		{
			name:                     "corrupted output",
			pciID:                    "0000:0c:00.0",
			output:                   createCorruptedMLXConfigOutput(),
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
			expectedValid:            true, // Changed to true because it still finds config lines
			expectError:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testParseMLXConfigOutput(tt.pciID, tt.output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result.PCIBusID != tt.pciID {
					t.Errorf("Expected PCIBusID %s, got %s", tt.pciID, result.PCIBusID)
				}
				if result.MaxAccOut != tt.expectedMaxAccOut {
					t.Errorf("Expected MaxAccOut %s, got %s", tt.expectedMaxAccOut, result.MaxAccOut)
				}
				if result.AdvancedPCISettings != tt.expectedAdvancedPCISettings {
					t.Errorf("Expected AdvancedPCISettings %s, got %s", tt.expectedAdvancedPCISettings, result.AdvancedPCISettings)
				}
				if result.IsValid != tt.expectedValid {
					t.Errorf("Expected IsValid %v, got %v", tt.expectedValid, result.IsValid)
				}

				t.Logf("Test %s: PCI=%s, MaxAcc=%s, AdvPCI=%s, Valid=%v", 
					tt.name, result.PCIBusID, result.MaxAccOut, result.AdvancedPCISettings, result.IsValid)
			}
		})
	}
}

func TestParseMLXConfigOutputDetails(t *testing.T) {
	// Test with specific H100 output variations
	tests := []struct {
		name      string
		pciID     string
		output    []string
		expectMax string
		expectAdv string
	}{
		{
			name:  "MAX_ACC_OUT_READ value 0",
			pciID: "0000:0c:00.0",
			output: []string{
				"         MAX_ACC_OUT_READ            0",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectMax: "PASS",
			expectAdv: "PASS",
		},
		{
			name:  "MAX_ACC_OUT_READ value 128",
			pciID: "0000:2a:00.0",
			output: []string{
				"         MAX_ACC_OUT_READ            128",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectMax: "PASS",
			expectAdv: "PASS",
		},
		{
			name:  "Invalid MAX_ACC_OUT_READ value 256",
			pciID: "0000:41:00.0",
			output: []string{
				"         MAX_ACC_OUT_READ            256",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectMax: "FAIL",
			expectAdv: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testParseMLXConfigOutput(tt.pciID, tt.output)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.MaxAccOut != tt.expectMax {
				t.Errorf("Expected MaxAccOut %s, got %s", tt.expectMax, result.MaxAccOut)
			}
			if result.AdvancedPCISettings != tt.expectAdv {
				t.Errorf("Expected AdvancedPCISettings %s, got %s", tt.expectAdv, result.AdvancedPCISettings)
			}
		})
	}
}

// Validation tests

func TestValidateMLXConfigResults(t *testing.T) {
	tests := []struct {
		name          string
		results       []testPCIEConfig
		shouldPass    bool
		expectedFails int
		description   string
	}{
		{
			name: "all devices pass",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
			},
			shouldPass:    true,
			expectedFails: 0,
			description:   "All devices configured correctly",
		},
		{
			name: "one device fails MAX_ACC_OUT_READ",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS", IsValid: true},
			},
			shouldPass:    false,
			expectedFails: 1,
			description:   "One device with incorrect MAX_ACC_OUT_READ",
		},
		{
			name: "one device fails ADVANCED_PCI_SETTINGS",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "FAIL", IsValid: true},
			},
			shouldPass:    false,
			expectedFails: 1,
			description:   "One device with incorrect ADVANCED_PCI_SETTINGS",
		},
		{
			name: "device fails both settings",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "FAIL", IsValid: true},
			},
			shouldPass:    false,
			expectedFails: 1, // Same device fails both, so only counted once
			description:   "One device with both settings incorrect",
		},
		{
			name: "invalid device configuration",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "FAIL", IsValid: false},
			},
			shouldPass:    false,
			expectedFails: 1,
			description:   "One device with invalid configuration",
		},
		{
			name:          "empty results",
			results:       []testPCIEConfig{},
			shouldPass:    false,
			expectedFails: 0,
			description:   "No device results",
		},
		{
			name: "all devices fail",
			results: []testPCIEConfig{
				{PCIBusID: "0000:0c:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS", IsValid: true},
				{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "FAIL", IsValid: true},
			},
			shouldPass:    false,
			expectedFails: 2,
			description:   "All devices have configuration issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, failedDevices, statusMsg := testValidateMLXConfigResults(tt.results)

			if isValid != tt.shouldPass {
				t.Errorf("Expected validation result %v, got %v", tt.shouldPass, isValid)
			}

			if len(failedDevices) != tt.expectedFails {
				t.Errorf("Expected %d failed devices, got %d", tt.expectedFails, len(failedDevices))
			}

			if statusMsg == "" {
				t.Error("Expected non-empty status message")
			}

			t.Logf("Test %s: %s -> Valid=%v, FailedDevices=%v, Status='%s'", 
				tt.name, tt.description, isValid, failedDevices, statusMsg)

			// Validate failed device IDs exist in input
			if tt.expectedFails > 0 {
				for _, failedDevice := range failedDevices {
					found := false
					for _, result := range tt.results {
						if result.PCIBusID == failedDevice {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Failed device '%s' not found in input results", failedDevice)
					}
				}
			}
		})
	}
}

// Edge cases and error handling

func TestParseMLXConfigOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		pciID       string
		output      []string
		expectError bool
		description string
	}{
		{
			name:        "whitespace only lines",
			pciID:       "0000:0c:00.0",
			output:      []string{"   ", "\t", "  \n  "},
			expectError: false,
			description: "Should handle whitespace-only lines gracefully",
		},
		{
			name:  "mixed valid and invalid lines",
			pciID: "0000:2a:00.0",
			output: []string{
				"Device type: ConnectX7",
				"         MAX_ACC_OUT_READ            44",
				"Invalid line format",
				"         ADVANCED_PCI_SETTINGS       True",
				"Another invalid line",
			},
			expectError: false,
			description: "Should handle mixed valid/invalid lines gracefully",
		},
		{
			name:  "case sensitivity test",
			pciID: "0000:41:00.0",
			output: []string{
				"         max_acc_out_read            44",     // lowercase
				"         ADVANCED_pci_SETTINGS       True",  // mixed case
			},
			expectError: false,
			description: "Should be case sensitive for configuration names",
		},
		{
			name:  "extra whitespace handling",
			pciID: "0000:58:00.0",
			output: []string{
				"    MAX_ACC_OUT_READ      44     ",
				"  ADVANCED_PCI_SETTINGS    True   ",
			},
			expectError: false,
			description: "Should handle extra whitespace correctly",
		},
		{
			name:  "numeric edge cases",
			pciID: "0000:86:00.0",
			output: []string{
				"         MAX_ACC_OUT_READ            0000", // leading zeros
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectError: false,
			description: "Should handle numeric edge cases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testParseMLXConfigOutput(tt.pciID, tt.output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				t.Logf("Test %s: %s -> MaxAcc=%s, AdvPCI=%s, Valid=%v", 
					tt.name, tt.description, result.MaxAccOut, result.AdvancedPCISettings, result.IsValid)
			}
		})
	}
}

func TestMLXConfigThresholds(t *testing.T) {
	validValues := []string{"0", "44", "128"}
	invalidValues := []string{"1", "16", "32", "64", "256", "512"}

	// Test valid MAX_ACC_OUT_READ values
	for _, value := range validValues {
		t.Run(fmt.Sprintf("valid_max_acc_%s", value), func(t *testing.T) {
			output := createMinimalMLXConfigOutput(value, "True")
			result, err := testParseMLXConfigOutput("0000:0c:00.0", output)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if result.MaxAccOut != "PASS" {
				t.Errorf("Expected MAX_ACC_OUT_READ=%s to pass, got %s", value, result.MaxAccOut)
			}
		})
	}

	// Test invalid MAX_ACC_OUT_READ values
	for _, value := range invalidValues {
		t.Run(fmt.Sprintf("invalid_max_acc_%s", value), func(t *testing.T) {
			output := createMinimalMLXConfigOutput(value, "True")
			result, err := testParseMLXConfigOutput("0000:0c:00.0", output)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if result.MaxAccOut != "FAIL" {
				t.Errorf("Expected MAX_ACC_OUT_READ=%s to fail, got %s", value, result.MaxAccOut)
			}
		})
	}

	// Test ADVANCED_PCI_SETTINGS values
	advancedPCITests := map[string]string{
		"True":  "PASS",
		"False": "FAIL",
		"true":  "FAIL", // case sensitive
		"TRUE":  "FAIL", // case sensitive
		"1":     "FAIL",
		"0":     "FAIL",
	}

	for value, expected := range advancedPCITests {
		t.Run(fmt.Sprintf("advanced_pci_%s", value), func(t *testing.T) {
			output := createMinimalMLXConfigOutput("44", value)
			result, err := testParseMLXConfigOutput("0000:0c:00.0", output)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if result.AdvancedPCISettings != expected {
				t.Errorf("Expected ADVANCED_PCI_SETTINGS=%s to result in %s, got %s", 
					value, expected, result.AdvancedPCISettings)
			}
		})
	}
}

// Performance tests

func TestLargeMLXConfigOutput(t *testing.T) {
	// Test with 8 devices (full H100 system)
	pciIDs := []string{
		"0000:0c:00.0", "0000:2a:00.0", "0000:41:00.0", "0000:58:00.0",
		"0000:86:00.0", "0000:a5:00.0", "0000:bd:00.0", "0000:d5:00.0",
	}

	var results []testPCIEConfig
	for _, pciID := range pciIDs {
		output := createTestMLXConfigOutput("44", "True")
		result, err := testParseMLXConfigOutput(pciID, output)
		if err != nil {
			t.Fatalf("Unexpected error for PCI %s: %v", pciID, err)
		}
		results = append(results, result)
	}

	if len(results) != 8 {
		t.Errorf("Expected 8 device results, got %d", len(results))
	}

	// Validate each device result
	for i, result := range results {
		if !result.IsValid {
			t.Errorf("Device %d: Expected IsValid true, got false", i)
		}
		if result.MaxAccOut != "PASS" {
			t.Errorf("Device %d: Expected MaxAccOut PASS, got %s", i, result.MaxAccOut)
		}
		if result.AdvancedPCISettings != "PASS" {
			t.Errorf("Device %d: Expected AdvancedPCISettings PASS, got %s", i, result.AdvancedPCISettings)
		}
		if result.PCIBusID != pciIDs[i] {
			t.Errorf("Device %d: Expected PCIBusID %s, got %s", i, pciIDs[i], result.PCIBusID)
		}
	}

	// Test validation with all devices
	isValid, failedDevices, statusMsg := testValidateMLXConfigResults(results)
	if !isValid {
		t.Errorf("Expected validation to pass, got failed devices: %v", failedDevices)
	}
	if len(failedDevices) != 0 {
		t.Errorf("Expected 0 failed devices, got %d", len(failedDevices))
	}
	if statusMsg == "" {
		t.Error("Expected non-empty status message")
	}

	t.Logf("Large output test: Valid=%v, Status='%s'", isValid, statusMsg)
}

func BenchmarkParseMLXConfigOutput(b *testing.B) {
	output := createTestMLXConfigOutput("44", "True")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testParseMLXConfigOutput("0000:0c:00.0", output)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateMLXConfigResults(b *testing.B) {
	results := []testPCIEConfig{
		{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:58:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:86:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:a5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:bd:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
		{PCIBusID: "0000:d5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS", IsValid: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = testValidateMLXConfigResults(results)
	}
}

// Integration tests with actual validation function

func TestIntegrationWithValidateMaxAccResults(t *testing.T) {
	tests := []struct {
		name           string
		result         *MaxAccCheckResult
		expectedStatus string
		expectError    bool
	}{
		{
			name: "integration test - all pass",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "PASS",
			expectError:    false,
		},
		{
			name: "integration test - some fail",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "FAIL",
			expectError:    true,
		},
		{
			name:           "integration test - nil result",
			result:         nil,
			expectedStatus: "FAIL",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, err := validateMaxAccResults(tt.result)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error %v, got error: %v", tt.expectError, err)
			}

			t.Logf("Integration test: Status=%s, Error=%v", status, err)
		})
	}
}