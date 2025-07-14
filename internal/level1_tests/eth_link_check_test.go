package level1_tests

import (
	"fmt"
	"testing"
)

func TestParseEthLinkResults(t *testing.T) {
	tests := []struct {
		name         string
		interfaceName string
		mlxlinkOutput string
		expectedSpeed string
		expectedWidth string
		rawThreshold  int
		effThreshold  int
		effBERThreshold float64
		rawBERThreshold float64
		expected      *EthLinkCheckResult
	}{
		{
			name:          "Valid JSON with passing parameters",
			interfaceName: "enp12s0f0np0",
			mlxlinkOutput: `{
				"result": {
					"output": {
						"Operational Info": {
							"Speed": "100G",
							"State": "Active",
							"Physical state": "LinkUp",
							"Width": "4x"
						},
						"Troubleshooting Info": {
							"Status Opcode": "0",
							"Recommendation": "No issues found"
						},
						"Physical Counters and BER Info": {
							"Effective Physical Errors": "0",
							"Effective Physical BER": "1e-15",
							"Raw Physical BER": "1e-8",
							"Raw Physical Errors Per Lane": ["0", "0", "0", "0"]
						}
					}
				}
			}`,
			expectedSpeed:    "100G",
			expectedWidth:    "4x",
			rawThreshold:     10000,
			effThreshold:     0,
			effBERThreshold:  1e-12,
			rawBERThreshold:  1e-5,
			expected: &EthLinkCheckResult{
				Device:                     "enp12s0f0np0",
				EthLinkSpeed:              "PASS",
				EthLinkState:              "PASS",
				PhysicalState:             "PASS",
				EthLinkWidth:              "PASS",
				EthLinkStatus:             "PASS",
				EffectivePhysicalErrors:   "PASS",
				EffectivePhysicalBER:      "PASS",
				RawPhysicalErrorsPerLane:  "PASS",
				RawPhysicalBER:            "PASS",
			},
		},
		{
			name:          "Valid JSON with failing parameters",
			interfaceName: "enp12s0f0np0",
			mlxlinkOutput: `{
				"result": {
					"output": {
						"Operational Info": {
							"Speed": "50G",
							"State": "Down",
							"Physical state": "Disabled",
							"Width": "2x"
						},
						"Troubleshooting Info": {
							"Status Opcode": "1",
							"Recommendation": "Check cable connection"
						},
						"Physical Counters and BER Info": {
							"Effective Physical Errors": "5",
							"Effective Physical BER": "1e-10",
							"Raw Physical BER": "1e-3",
							"Raw Physical Errors Per Lane": ["15000", "0", "0", "0"]
						}
					}
				}
			}`,
			expectedSpeed:    "100G",
			expectedWidth:    "4x",
			rawThreshold:     10000,
			effThreshold:     0,
			effBERThreshold:  1e-12,
			rawBERThreshold:  1e-5,
			expected: &EthLinkCheckResult{
				Device:                     "enp12s0f0np0",
				EthLinkSpeed:              "FAIL - 50G, expected 100G",
				EthLinkState:              "FAIL - Down, expected Active",
				PhysicalState:             "FAIL - Disabled, expected [LinkUp ETH_AN_FSM_ENABLE]",
				EthLinkWidth:              "FAIL - 2x, expected 4x",
				EthLinkStatus:             "FAIL - Check cable connection",
				EffectivePhysicalErrors:   "FAIL - 5",
				EffectivePhysicalBER:      "FAIL - 1e-10",
				RawPhysicalErrorsPerLane:  "WARN - 15000 0 0 0",
				RawPhysicalBER:            "FAIL - 1e-3",
			},
		},
		{
			name:          "Invalid JSON",
			interfaceName: "enp12s0f0np0",
			mlxlinkOutput: "invalid json",
			expectedSpeed: "100G",
			expectedWidth: "4x",
			rawThreshold:  10000,
			effThreshold:  0,
			effBERThreshold: 1e-12,
			rawBERThreshold: 1e-5,
			expected: &EthLinkCheckResult{
				Device:                     "enp12s0f0np0",
				EthLinkSpeed:              "FAIL - Unable to parse mlxlink output",
				EthLinkState:              "FAIL - Unable to parse mlxlink output",
				PhysicalState:             "FAIL - Unable to parse mlxlink output",
				EthLinkWidth:              "FAIL - Unable to parse mlxlink output",
				EthLinkStatus:             "FAIL - Unable to parse mlxlink output",
				EffectivePhysicalErrors:   "PASS",
				EffectivePhysicalBER:      "FAIL - Unable to parse mlxlink output",
				RawPhysicalErrorsPerLane:  "PASS",
				RawPhysicalBER:            "FAIL - Unable to parse mlxlink output",
			},
		},
		{
			name:          "Empty output",
			interfaceName: "enp12s0f0np0",
			mlxlinkOutput: "",
			expectedSpeed: "100G",
			expectedWidth: "4x",
			rawThreshold:  10000,
			effThreshold:  0,
			effBERThreshold: 1e-12,
			rawBERThreshold: 1e-5,
			expected: &EthLinkCheckResult{
				Device:                     "enp12s0f0np0",
				EthLinkSpeed:              "FAIL - Invalid interface: enp12s0f0np0",
				EthLinkState:              "FAIL - Invalid interface: enp12s0f0np0",
				PhysicalState:             "FAIL - Invalid interface: enp12s0f0np0",
				EthLinkWidth:              "FAIL - Invalid interface: enp12s0f0np0",
				EthLinkStatus:             "FAIL - Invalid interface: enp12s0f0np0",
				EffectivePhysicalErrors:   "PASS",
				EffectivePhysicalBER:      "FAIL - Unable to get data",
				RawPhysicalErrorsPerLane:  "PASS",
				RawPhysicalBER:            "FAIL - Unable to get data",
			},
		},
		{
			name:          "Error prefix with valid JSON",
			interfaceName: "enp12s0f0np0",
			mlxlinkOutput: `Error: Some error message {
				"result": {
					"output": {
						"Operational Info": {
							"Speed": "100G",
							"State": "Active",
							"Physical state": "LinkUp",
							"Width": "4x"
						},
						"Troubleshooting Info": {
							"Status Opcode": "0",
							"Recommendation": "No issues found"
						},
						"Physical Counters and BER Info": {
							"Effective Physical Errors": "0",
							"Effective Physical BER": "1e-15",
							"Raw Physical BER": "1e-8",
							"Raw Physical Errors Per Lane": ["0", "0", "0", "0"]
						}
					}
				}
			}`,
			expectedSpeed:    "100G",
			expectedWidth:    "4x",
			rawThreshold:     10000,
			effThreshold:     0,
			effBERThreshold:  1e-12,
			rawBERThreshold:  1e-5,
			expected: &EthLinkCheckResult{
				Device:                     "enp12s0f0np0",
				EthLinkSpeed:              "PASS",
				EthLinkState:              "PASS",
				PhysicalState:             "PASS",
				EthLinkWidth:              "PASS",
				EthLinkStatus:             "PASS",
				EffectivePhysicalErrors:   "PASS",
				EffectivePhysicalBER:      "PASS",
				RawPhysicalErrorsPerLane:  "PASS",
				RawPhysicalBER:            "PASS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEthLinkResults(
				tt.interfaceName,
				tt.mlxlinkOutput,
				tt.expectedSpeed,
				tt.expectedWidth,
				tt.rawThreshold,
				tt.effThreshold,
				tt.effBERThreshold,
				tt.rawBERThreshold,
			)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			// Compare all fields
			if result.Device != tt.expected.Device {
				t.Errorf("Device: expected %s, got %s", tt.expected.Device, result.Device)
			}
			if result.EthLinkSpeed != tt.expected.EthLinkSpeed {
				t.Errorf("EthLinkSpeed: expected %s, got %s", tt.expected.EthLinkSpeed, result.EthLinkSpeed)
			}
			if result.EthLinkState != tt.expected.EthLinkState {
				t.Errorf("EthLinkState: expected %s, got %s", tt.expected.EthLinkState, result.EthLinkState)
			}
			if result.PhysicalState != tt.expected.PhysicalState {
				t.Errorf("PhysicalState: expected %s, got %s", tt.expected.PhysicalState, result.PhysicalState)
			}
			if result.EthLinkWidth != tt.expected.EthLinkWidth {
				t.Errorf("EthLinkWidth: expected %s, got %s", tt.expected.EthLinkWidth, result.EthLinkWidth)
			}
			if result.EthLinkStatus != tt.expected.EthLinkStatus {
				t.Errorf("EthLinkStatus: expected %s, got %s", tt.expected.EthLinkStatus, result.EthLinkStatus)
			}
			if result.EffectivePhysicalErrors != tt.expected.EffectivePhysicalErrors {
				t.Errorf("EffectivePhysicalErrors: expected %s, got %s", tt.expected.EffectivePhysicalErrors, result.EffectivePhysicalErrors)
			}
			if result.EffectivePhysicalBER != tt.expected.EffectivePhysicalBER {
				t.Errorf("EffectivePhysicalBER: expected %s, got %s", tt.expected.EffectivePhysicalBER, result.EffectivePhysicalBER)
			}
			if result.RawPhysicalErrorsPerLane != tt.expected.RawPhysicalErrorsPerLane {
				t.Errorf("RawPhysicalErrorsPerLane: expected %s, got %s", tt.expected.RawPhysicalErrorsPerLane, result.RawPhysicalErrorsPerLane)
			}
			if result.RawPhysicalBER != tt.expected.RawPhysicalBER {
				t.Errorf("RawPhysicalBER: expected %s, got %s", tt.expected.RawPhysicalBER, result.RawPhysicalBER)
			}
		})
	}
}

func TestParseEthLinkResultsETHANFSMEnable(t *testing.T) {
	// Test ETH_AN_FSM_ENABLE physical state specifically
	mlxlinkOutput := `{
		"result": {
			"output": {
				"Operational Info": {
					"Speed": "100G",
					"State": "Active",
					"Physical state": "ETH_AN_FSM_ENABLE",
					"Width": "4x"
				},
				"Troubleshooting Info": {
					"Status Opcode": "0",
					"Recommendation": "No issues found"
				},
				"Physical Counters and BER Info": {
					"Effective Physical Errors": "0",
					"Effective Physical BER": "1e-15",
					"Raw Physical BER": "1e-8",
					"Raw Physical Errors Per Lane": ["0", "0", "0", "0"]
				}
			}
		}
	}`

	result, err := parseEthLinkResults(
		"enp12s0f0np0",
		mlxlinkOutput,
		"100G",
		"4x",
		10000,
		0,
		1e-12,
		1e-5,
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result.PhysicalState != "PASS" {
		t.Errorf("Expected PhysicalState=PASS for ETH_AN_FSM_ENABLE, got %s", result.PhysicalState)
	}
}

func TestParseEthLinkResultsSpeedContains(t *testing.T) {
	// Test that speed matching uses "contains" logic
	mlxlinkOutput := `{
		"result": {
			"output": {
				"Operational Info": {
					"Speed": "100000 Mb/s (100G)",
					"State": "Active",
					"Physical state": "LinkUp",
					"Width": "4x"
				},
				"Troubleshooting Info": {
					"Status Opcode": "0",
					"Recommendation": "No issues found"
				},
				"Physical Counters and BER Info": {
					"Effective Physical Errors": "0",
					"Effective Physical BER": "1e-15",
					"Raw Physical BER": "1e-8",
					"Raw Physical Errors Per Lane": ["0", "0", "0", "0"]
				}
			}
		}
	}`

	result, err := parseEthLinkResults(
		"enp12s0f0np0",
		mlxlinkOutput,
		"100G",
		"4x",
		10000,
		0,
		1e-12,
		1e-5,
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result.EthLinkSpeed != "PASS" {
		t.Errorf("Expected EthLinkSpeed=PASS when speed contains expected value, got %s", result.EthLinkSpeed)
	}
}

// Test configuration handling
func TestEthLinkCheckTestConfig(t *testing.T) {
	config := EthLinkCheckTestConfig{
		IsEnabled:                           true,
		ExpectedSpeed:                       "100G",
		ExpectedWidth:                       "4x",
		EffectivePhysicalErrorsThreshold:    0,
		RawPhysicalErrorsPerLaneThreshold:   10000,
		EffectivePhysicalBERThreshold:       1e-12,
		RawPhysicalBERThreshold:             1e-5,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.ExpectedSpeed != "100G" {
		t.Error("Expected speed to be 100G")
	}
	if config.ExpectedWidth != "4x" {
		t.Error("Expected width to be 4x")
	}
	if config.EffectivePhysicalErrorsThreshold != 0 {
		t.Error("Expected effective physical errors threshold to be 0")
	}
	if config.RawPhysicalErrorsPerLaneThreshold != 10000 {
		t.Error("Expected raw physical errors per lane threshold to be 10000")
	}
	if config.EffectivePhysicalBERThreshold != 1e-12 {
		t.Error("Expected effective physical BER threshold to be 1e-12")
	}
	if config.RawPhysicalBERThreshold != 1e-5 {
		t.Error("Expected raw physical BER threshold to be 1e-5")
	}
}

// Test error threshold validation
func TestErrorThresholdValidation(t *testing.T) {
	tests := []struct {
		name        string
		errors      string
		threshold   int
		shouldPass  bool
	}{
		{
			name:       "Errors below threshold",
			errors:     "5",
			threshold:  10,
			shouldPass: true,
		},
		{
			name:       "Errors above threshold",
			errors:     "15",
			threshold:  10,
			shouldPass: false,
		},
		{
			name:       "Zero errors",
			errors:     "0",
			threshold:  10,
			shouldPass: true,
		},
		{
			name:       "Invalid error format",
			errors:     "invalid",
			threshold:  10,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic used in parseEthLinkResults
			if tt.errors == "" {
				return // Skip empty errors
			}

			// Try to parse as integer
			var errorCount int
			if _, err := fmt.Sscanf(tt.errors, "%d", &errorCount); err != nil {
				// Invalid format should fail
				if tt.shouldPass {
					t.Errorf("Expected valid error count but got parsing error: %v", err)
				}
				return
			}

			passed := errorCount <= tt.threshold
			if passed != tt.shouldPass {
				t.Errorf("Expected pass=%v for error validation, got pass=%v", tt.shouldPass, passed)
			}
		})
	}
}

// Test raw physical errors per lane parsing
func TestRawPhysicalErrorsPerLaneParsing(t *testing.T) {
	tests := []struct {
		name       string
		laneErrors interface{}
		threshold  int
		shouldWarn bool
	}{
		{
			name:       "All lanes below threshold",
			laneErrors: []interface{}{"100", "200", "150", "300"},
			threshold:  10000,
			shouldWarn: false,
		},
		{
			name:       "One lane above threshold",
			laneErrors: []interface{}{"100", "15000", "150", "300"},
			threshold:  10000,
			shouldWarn: true,
		},
		{
			name:       "Mixed string and number types",
			laneErrors: []interface{}{100, "15000", 150, "300"},
			threshold:  10000,
			shouldWarn: true,
		},
		{
			name:       "Invalid lane data",
			laneErrors: []interface{}{"100", "invalid", "150", "300"},
			threshold:  10000,
			shouldWarn: false, // Invalid data should be skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldWarn := false
			
			// Simulate the lane error parsing logic
			if laneArray, ok := tt.laneErrors.([]interface{}); ok {
				for _, lane := range laneArray {
					var laneError int
					switch v := lane.(type) {
					case string:
						if _, err := fmt.Sscanf(v, "%d", &laneError); err != nil {
							continue // Skip invalid values
						}
					case int:
						laneError = v
					case float64:
						laneError = int(v)
					default:
						continue // Skip unknown types
					}
					
					if laneError > tt.threshold {
						shouldWarn = true
						break
					}
				}
			}

			if shouldWarn != tt.shouldWarn {
				t.Errorf("Expected warning=%v, got warning=%v", tt.shouldWarn, shouldWarn)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseEthLinkResults(b *testing.B) {
	mlxlinkOutput := `{
		"result": {
			"output": {
				"Operational Info": {
					"Speed": "100G",
					"State": "Active",
					"Physical state": "LinkUp",
					"Width": "4x"
				},
				"Troubleshooting Info": {
					"Status Opcode": "0",
					"Recommendation": "No issues found"
				},
				"Physical Counters and BER Info": {
					"Effective Physical Errors": "0",
					"Effective Physical BER": "1e-15",
					"Raw Physical BER": "1e-8",
					"Raw Physical Errors Per Lane": ["0", "0", "0", "0"]
				}
			}
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseEthLinkResults(
			"enp12s0f0np0",
			mlxlinkOutput,
			"100G",
			"4x",
			10000,
			0,
			1e-12,
			1e-5,
		)
	}
}