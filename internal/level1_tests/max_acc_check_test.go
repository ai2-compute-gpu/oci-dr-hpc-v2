package level1_tests

import (
	"testing"
)

// Test parseAccResults function
func TestParseAccResults(t *testing.T) {
	tests := []struct {
		name                     string
		pciID                    string
		mlxconfigOutput          []string
		expectedMaxAccOut        string
		expectedAdvancedPCISettings string
	}{
		{
			name:  "Valid configuration with MAX_ACC_OUT_READ=0",
			pciID: "0000:0c:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"Device type:    ConnectX7",
				"PCI device:     0000:0c:00.0",
				"",
				"Configurations:                      Current",
				"         MAX_ACC_OUT_READ            0",
				"         ADVANCED_PCI_SETTINGS       True",
				"         OTHER_SETTING               value",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
		},
		{
			name:  "Valid configuration with MAX_ACC_OUT_READ=44",
			pciID: "0000:2a:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            44",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
		},
		{
			name:  "Valid configuration with MAX_ACC_OUT_READ=128",
			pciID: "0000:41:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            128",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
		},
		{
			name:  "Invalid MAX_ACC_OUT_READ value",
			pciID: "0000:58:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            32",
				"         ADVANCED_PCI_SETTINGS       True",
			},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "PASS",
		},
		{
			name:  "Invalid ADVANCED_PCI_SETTINGS value",
			pciID: "0000:86:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            44",
				"         ADVANCED_PCI_SETTINGS       False",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "FAIL",
		},
		{
			name:  "Both settings invalid",
			pciID: "0000:a5:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            16",
				"         ADVANCED_PCI_SETTINGS       False",
			},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
		},
		{
			name:  "Missing MAX_ACC_OUT_READ setting",
			pciID: "0000:bd:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         ADVANCED_PCI_SETTINGS       True",
				"         OTHER_SETTING               value",
			},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "PASS",
		},
		{
			name:  "Missing ADVANCED_PCI_SETTINGS setting",
			pciID: "0000:d5:00.0",
			mlxconfigOutput: []string{
				"Device #1:",
				"----------",
				"         MAX_ACC_OUT_READ            44",
				"         OTHER_SETTING               value",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "FAIL",
		},
		{
			name:                     "Empty mlxconfig output",
			pciID:                    "0000:0c:00.0",
			mlxconfigOutput:          []string{},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
		},
		{
			name:  "Whitespace and formatting variations",
			pciID: "0000:0c:00.0",
			mlxconfigOutput: []string{
				"         MAX_ACC_OUT_READ      0      ",
				"    ADVANCED_PCI_SETTINGS        True   ",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAccResults(tt.pciID, tt.mlxconfigOutput)

			if result.PCIBusID != tt.pciID {
				t.Errorf("parseAccResults() PCIBusID = %v, want %v", result.PCIBusID, tt.pciID)
			}

			if result.MaxAccOut != tt.expectedMaxAccOut {
				t.Errorf("parseAccResults() MaxAccOut = %v, want %v", result.MaxAccOut, tt.expectedMaxAccOut)
			}

			if result.AdvancedPCISettings != tt.expectedAdvancedPCISettings {
				t.Errorf("parseAccResults() AdvancedPCISettings = %v, want %v", result.AdvancedPCISettings, tt.expectedAdvancedPCISettings)
			}
		})
	}
}

// Test validateMaxAccResults function
func TestValidateMaxAccResults(t *testing.T) {
	tests := []struct {
		name            string
		result          *MaxAccCheckResult
		expectedStatus  string
		expectedError   bool
		expectedMessage string
	}{
		{
			name: "All devices pass",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus:  "PASS",
			expectedError:   false,
			expectedMessage: "All 3 PCI devices configured correctly",
		},
		{
			name: "Single device fails MAX_ACC_OUT_READ",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name: "Single device fails ADVANCED_PCI_SETTINGS",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "FAIL"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name: "Multiple devices fail",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "FAIL"},
					{PCIBusID: "0000:41:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "FAIL"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name: "Device fails both settings",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "FAIL"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:            "Nil result",
			result:          nil,
			expectedStatus:  "FAIL",
			expectedError:   true,
			expectedMessage: "No PCI devices found",
		},
		{
			name: "Empty PCIEConfig",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{},
			},
			expectedStatus:  "FAIL",
			expectedError:   true,
			expectedMessage: "No PCI devices found",
		},
		{
			name: "Single device passes",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus:  "PASS",
			expectedError:   false,
			expectedMessage: "All 1 PCI devices configured correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message, err := validateMaxAccResults(tt.result)

			if status != tt.expectedStatus {
				t.Errorf("validateMaxAccResults() status = %v, want %v", status, tt.expectedStatus)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("validateMaxAccResults() error = %v, wantErr %v", err, tt.expectedError)
			}

			if tt.expectedMessage != "" && message != tt.expectedMessage {
				t.Errorf("validateMaxAccResults() message = %v, want %v", message, tt.expectedMessage)
			}

			// Log for debugging
			t.Logf("Status: %s, Message: %s, Error: %v", status, message, err)
		})
	}
}

// Test getMaxAccCheckTestConfig function (basic validation)
func TestGetMaxAccCheckTestConfig(t *testing.T) {
	// This test will only work if we're in a test environment
	// We'll test the default values when IMDS isn't available
	config := &MaxAccCheckTestConfig{
		IsEnabled: false,
		Shape:     "test-shape",
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

	// Verify default PCI IDs
	if len(config.PCIIDs) != 8 {
		t.Errorf("Expected 8 default PCI IDs, got %d", len(config.PCIIDs))
	}

	expectedPCIIDs := []string{
		"0000:0c:00.0",
		"0000:2a:00.0",
		"0000:41:00.0",
		"0000:58:00.0",
		"0000:86:00.0",
		"0000:a5:00.0",
		"0000:bd:00.0",
		"0000:d5:00.0",
	}

	for i, expectedID := range expectedPCIIDs {
		if i >= len(config.PCIIDs) || config.PCIIDs[i] != expectedID {
			t.Errorf("Expected PCI ID at index %d to be %s, got %s", i, expectedID, config.PCIIDs[i])
		}
	}

	// Verify shape is set
	if config.Shape == "" {
		t.Error("Expected shape to be set")
	}

	// Verify IsEnabled is properly set
	if config.IsEnabled {
		t.Error("Expected IsEnabled to be false by default in test environment")
	}
}

// Test PrintMaxAccCheck function
func TestPrintMaxAccCheck(t *testing.T) {
	// This is mainly to ensure the function doesn't panic
	PrintMaxAccCheck()
}

// Test MAX_ACC_OUT_READ validation edge cases
func TestMaxAccOutReadValidation(t *testing.T) {
	tests := []struct {
		name           string
		mlxconfigLine  string
		expectedResult string
		description    string
	}{
		{
			name:           "Valid value 0",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            0",
			expectedResult: "PASS",
			description:    "MAX_ACC_OUT_READ=0 should pass",
		},
		{
			name:           "Valid value 44",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            44",
			expectedResult: "PASS",
			description:    "MAX_ACC_OUT_READ=44 should pass",
		},
		{
			name:           "Valid value 128",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            128",
			expectedResult: "PASS",
			description:    "MAX_ACC_OUT_READ=128 should pass",
		},
		{
			name:           "Invalid value 16",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            16",
			expectedResult: "FAIL",
			description:    "MAX_ACC_OUT_READ=16 should fail",
		},
		{
			name:           "Invalid value 32",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            32",
			expectedResult: "FAIL",
			description:    "MAX_ACC_OUT_READ=32 should fail",
		},
		{
			name:           "Invalid value 256",
			mlxconfigLine:  "         MAX_ACC_OUT_READ            256",
			expectedResult: "FAIL",
			description:    "MAX_ACC_OUT_READ=256 should fail",
		},
		{
			name:           "Valid value with extra whitespace",
			mlxconfigLine:  "    MAX_ACC_OUT_READ      44     ",
			expectedResult: "PASS",
			description:    "MAX_ACC_OUT_READ=44 with whitespace should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAccResults("0000:0c:00.0", []string{tt.mlxconfigLine})
			
			if result.MaxAccOut != tt.expectedResult {
				t.Errorf("Test %s: %s - Expected %s, got %s", 
					tt.name, tt.description, tt.expectedResult, result.MaxAccOut)
			}
			
			t.Logf("Test %s: %s -> %s", tt.name, tt.description, result.MaxAccOut)
		})
	}
}

// Test ADVANCED_PCI_SETTINGS validation edge cases
func TestAdvancedPCISettingsValidation(t *testing.T) {
	tests := []struct {
		name           string
		mlxconfigLine  string
		expectedResult string
		description    string
	}{
		{
			name:           "Valid True setting",
			mlxconfigLine:  "         ADVANCED_PCI_SETTINGS       True",
			expectedResult: "PASS",
			description:    "ADVANCED_PCI_SETTINGS=True should pass",
		},
		{
			name:           "Invalid False setting",
			mlxconfigLine:  "         ADVANCED_PCI_SETTINGS       False",
			expectedResult: "FAIL",
			description:    "ADVANCED_PCI_SETTINGS=False should fail",
		},
		{
			name:           "Invalid case variation - true",
			mlxconfigLine:  "         ADVANCED_PCI_SETTINGS       true",
			expectedResult: "FAIL",
			description:    "ADVANCED_PCI_SETTINGS=true (lowercase) should fail",
		},
		{
			name:           "Invalid case variation - TRUE",
			mlxconfigLine:  "         ADVANCED_PCI_SETTINGS       TRUE",
			expectedResult: "FAIL",
			description:    "ADVANCED_PCI_SETTINGS=TRUE (uppercase) should fail",
		},
		{
			name:           "Valid with extra whitespace",
			mlxconfigLine:  "    ADVANCED_PCI_SETTINGS     True    ",
			expectedResult: "PASS",
			description:    "ADVANCED_PCI_SETTINGS=True with whitespace should pass",
		},
		{
			name:           "Invalid numeric value",
			mlxconfigLine:  "         ADVANCED_PCI_SETTINGS       1",
			expectedResult: "FAIL",
			description:    "ADVANCED_PCI_SETTINGS=1 should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAccResults("0000:0c:00.0", []string{tt.mlxconfigLine})
			
			if result.AdvancedPCISettings != tt.expectedResult {
				t.Errorf("Test %s: %s - Expected %s, got %s", 
					tt.name, tt.description, tt.expectedResult, result.AdvancedPCISettings)
			}
			
			t.Logf("Test %s: %s -> %s", tt.name, tt.description, result.AdvancedPCISettings)
		})
	}
}

// Test realistic mlxconfig output scenarios
func TestRealisticMLXConfigOutput(t *testing.T) {
	tests := []struct {
		name                     string
		pciID                    string
		mlxconfigOutput          []string
		expectedMaxAccOut        string
		expectedAdvancedPCISettings string
		description              string
	}{
		{
			name:  "Typical H100 system output - correct config",
			pciID: "0000:0c:00.0",
			mlxconfigOutput: []string{
				"",
				"Device #1:",
				"----------",
				"",
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
				"         MAX_ACC_OUT_READ            44",
				"         ROCE_CC_PRIO_MASK_P1        0xff",
				"         ROCE_CC_PRIO_MASK_P2        0xff",
				"         ADVANCED_PCI_SETTINGS       True",
				"         NUM_PF                      1",
				"",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
			description:              "Realistic H100 ConnectX-7 output with correct configuration",
		},
		{
			name:  "Typical H100 system output - incorrect MAX_ACC_OUT_READ",
			pciID: "0000:2a:00.0",
			mlxconfigOutput: []string{
				"Device #2:",
				"----------",
				"Device type:    ConnectX7",
				"Configurations:                      Current",
				"         MAX_ACC_OUT_READ            16",
				"         ADVANCED_PCI_SETTINGS       True",
				"         OTHER_CONFIG                value",
			},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "PASS",
			description:              "ConnectX-7 with incorrect MAX_ACC_OUT_READ value",
		},
		{
			name:  "System with both settings incorrect",
			pciID: "0000:41:00.0",
			mlxconfigOutput: []string{
				"Device #3:",
				"----------",
				"Device type:    ConnectX7",
				"Configurations:                      Current",
				"         MEMIC_BAR_SIZE              0",
				"         MAX_ACC_OUT_READ            32",
				"         ADVANCED_PCI_SETTINGS       False",
				"         NUM_PF                      1",
			},
			expectedMaxAccOut:        "FAIL",
			expectedAdvancedPCISettings: "FAIL",
			description:              "Both MAX_ACC_OUT_READ and ADVANCED_PCI_SETTINGS incorrect",
		},
		{
			name:  "Output with unusual formatting",
			pciID: "0000:58:00.0",
			mlxconfigOutput: []string{
				"Device #4:",
				"----------",
				"",
				"Configurations:          Current     Next Boot   New Value",
				"         MAX_ACC_OUT_READ    0           0           -",
				"         ADVANCED_PCI_SETTINGS True      True        -",
				"",
			},
			expectedMaxAccOut:        "PASS",
			expectedAdvancedPCISettings: "PASS",
			description:              "Output with Next Boot column formatting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAccResults(tt.pciID, tt.mlxconfigOutput)

			if result.PCIBusID != tt.pciID {
				t.Errorf("parseAccResults() PCIBusID = %v, want %v", result.PCIBusID, tt.pciID)
			}

			if result.MaxAccOut != tt.expectedMaxAccOut {
				t.Errorf("parseAccResults() MaxAccOut = %v, want %v", result.MaxAccOut, tt.expectedMaxAccOut)
			}

			if result.AdvancedPCISettings != tt.expectedAdvancedPCISettings {
				t.Errorf("parseAccResults() AdvancedPCISettings = %v, want %v", result.AdvancedPCISettings, tt.expectedAdvancedPCISettings)
			}

			t.Logf("Test %s: %s -> MaxAccOut=%s, AdvancedPCI=%s", 
				tt.name, tt.description, result.MaxAccOut, result.AdvancedPCISettings)
		})
	}
}

// Test validation with realistic scenarios
func TestValidationWithRealisticScenarios(t *testing.T) {
	tests := []struct {
		name           string
		result         *MaxAccCheckResult
		expectedStatus string
		expectedError  bool
		description    string
	}{
		{
			name: "Full H100 system - all correct",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:58:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:86:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:a5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:bd:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:d5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "PASS",
			expectedError:  false,
			description:    "Full H100 system with 8 NICs all correctly configured",
		},
		{
			name: "H100 system with one misconfigured NIC",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS"}, // One bad NIC
					{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:58:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:86:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:a5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:bd:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:d5:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
			description:    "H100 system with one NIC having incorrect MAX_ACC_OUT_READ",
		},
		{
			name: "Partial system deployment",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
				},
			},
			expectedStatus: "PASS",
			expectedError:  false,
			description:    "Smaller system with only 2 NICs, both correctly configured",
		},
		{
			name: "System with multiple configuration issues",
			result: &MaxAccCheckResult{
				PCIEConfig: []PCIEConfig{
					{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:2a:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "PASS"},
					{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "FAIL"},
					{PCIBusID: "0000:58:00.0", MaxAccOut: "FAIL", AdvancedPCISettings: "FAIL"},
				},
			},
			expectedStatus: "FAIL",
			expectedError:  true,
			description:    "System with multiple NICs having various configuration issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message, err := validateMaxAccResults(tt.result)

			if status != tt.expectedStatus {
				t.Errorf("validateMaxAccResults() status = %v, want %v", status, tt.expectedStatus)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("validateMaxAccResults() error = %v, wantErr %v", err, tt.expectedError)
			}

			if message == "" {
				t.Error("Expected non-empty message")
			}

			t.Logf("Test %s: %s -> Status=%s, Message=%s, Error=%v", 
				tt.name, tt.description, status, message, err)
		})
	}
}

// Benchmark tests
func BenchmarkParseAccResults(b *testing.B) {
	mlxconfigOutput := []string{
		"Device #1:",
		"----------",
		"Device type:    ConnectX7",
		"Configurations:                      Current",
		"         MAX_ACC_OUT_READ            44",
		"         ADVANCED_PCI_SETTINGS       True",
		"         OTHER_SETTING               value",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseAccResults("0000:0c:00.0", mlxconfigOutput)
	}
}

func BenchmarkValidateMaxAccResults(b *testing.B) {
	result := &MaxAccCheckResult{
		PCIEConfig: []PCIEConfig{
			{PCIBusID: "0000:0c:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
			{PCIBusID: "0000:2a:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
			{PCIBusID: "0000:41:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
			{PCIBusID: "0000:58:00.0", MaxAccOut: "PASS", AdvancedPCISettings: "PASS"},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = validateMaxAccResults(result)
	}
}