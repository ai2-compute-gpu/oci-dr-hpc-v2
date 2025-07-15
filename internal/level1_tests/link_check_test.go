package level1_tests

import (
	"strconv"
	"strings"
	"testing"
)

// Test parseLinkResults function
func TestParseLinkResults(t *testing.T) {
	tests := []struct {
		name           string
		interfaceName  string
		mlxlinkOutput  string
		expectedSpeed  string
		expectError    bool
		expectedStatus string
	}{
		{
			name:           "Empty mlxlink output",
			interfaceName:  "rdma0",
			mlxlinkOutput:  "",
			expectedSpeed:  "200G",
			expectError:    false,
			expectedStatus: "FAIL",
		},
		{
			name:           "Error in mlxlink output",
			interfaceName:  "rdma0",
			mlxlinkOutput:  "Error: Invalid interface rdma0",
			expectedSpeed:  "200G",
			expectError:    false,
			expectedStatus: "FAIL",
		},
		{
			name:          "Valid JSON output with passing values",
			interfaceName: "rdma0",
			mlxlinkOutput: `{
				"result": {
					"output": {
						"Operational Info": {
							"Speed": "200G",
							"State": "Active",
							"Physical state": "LinkUp",
							"Width": "4x"
						},
						"Troubleshooting Info": {
							"Status Opcode": "0",
							"Recommendation": ""
						},
						"Physical Counters and BER Info": {
							"Effective Physical Errors": "0",
							"Effective Physical BER": "1E-13",
							"Raw Physical Errors Per Lane": ["0", "0", "0", "0"],
							"Raw Physical BER": "1E-6"
						}
					}
				}
			}`,
			expectedSpeed:  "200G",
			expectError:    false,
			expectedStatus: "PASS",
		},
		{
			name:          "Invalid JSON output",
			interfaceName: "rdma0",
			mlxlinkOutput: `{invalid json`,
			expectedSpeed:  "200G",
			expectError:    false,
			expectedStatus: "FAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLinkResults(
				tt.interfaceName,
				tt.mlxlinkOutput,
				tt.expectedSpeed,
					10000, // rawPhysicalErrorsPerLaneThreshold
				0,     // effectivePhysicalErrorsThreshold
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			if result.Device != tt.interfaceName {
				t.Errorf("Expected device %s, got %s", tt.interfaceName, result.Device)
			}

			// Check if status matches expectation
			if tt.expectedStatus == "PASS" {
				if !strings.HasPrefix(result.LinkSpeed, "PASS") {
					t.Errorf("Expected LinkSpeed to start with PASS, got %s", result.LinkSpeed)
				}
			} else if tt.expectedStatus == "FAIL" {
				if !strings.HasPrefix(result.LinkSpeed, "FAIL") {
					t.Errorf("Expected LinkSpeed to start with FAIL, got %s", result.LinkSpeed)
				}
			}
		})
	}
}

// Test LinkCheckResult structure
func TestLinkCheckResult(t *testing.T) {
	result := LinkCheckResult{
		Device:                      "rdma0",
		LinkSpeed:                   "PASS",
		LinkState:                   "PASS",
		PhysicalState:               "PASS",
		LinkStatus:                  "PASS",
		EffectivePhysicalErrors:     "PASS",
		EffectivePhysicalBER:        "PASS",
		RawPhysicalErrorsPerLane:    "PASS",
		RawPhysicalBER:              "PASS",
	}

	if result.Device != "rdma0" {
		t.Error("Device field mismatch")
	}
	if result.LinkSpeed != "PASS" {
		t.Error("LinkSpeed field mismatch")
	}
	if result.LinkState != "PASS" {
		t.Error("LinkState field mismatch")
	}
	if result.PhysicalState != "PASS" {
		t.Error("PhysicalState field mismatch")
	}
	if result.LinkStatus != "PASS" {
		t.Error("LinkStatus field mismatch")
	}
	if result.EffectivePhysicalErrors != "PASS" {
		t.Error("EffectivePhysicalErrors field mismatch")
	}
	if result.EffectivePhysicalBER != "PASS" {
		t.Error("EffectivePhysicalBER field mismatch")
	}
	if result.RawPhysicalErrorsPerLane != "PASS" {
		t.Error("RawPhysicalErrorsPerLane field mismatch")
	}
	if result.RawPhysicalBER != "PASS" {
		t.Error("RawPhysicalBER field mismatch")
	}
}

// Test link speed validation
func TestLinkSpeedValidation(t *testing.T) {
	tests := []struct {
		name          string
		actualSpeed   string
		expectedSpeed string
		shouldPass    bool
	}{
		{
			name:          "Exact match",
			actualSpeed:   "200G",
			expectedSpeed: "200G",
			shouldPass:    true,
		},
		{
			name:          "Speed with additional info",
			actualSpeed:   "200G FDR",
			expectedSpeed: "200G",
			shouldPass:    true,
		},
		{
			name:          "Different speed",
			actualSpeed:   "100G",
			expectedSpeed: "200G",
			shouldPass:    false,
		},
		{
			name:          "Empty speed",
			actualSpeed:   "",
			expectedSpeed: "200G",
			shouldPass:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.Contains(tt.actualSpeed, tt.expectedSpeed)
			if result != tt.shouldPass {
				t.Errorf("Expected %v for speed validation, got %v", tt.shouldPass, result)
			}
		})
	}
}

// Test physical state validation
func TestPhysicalStateValidation(t *testing.T) {
	tests := []struct {
		name               string
		actualPhysState    string
		expectedPhysStates []string
		shouldPass         bool
	}{
		{
			name:               "LinkUp state",
			actualPhysState:    "LinkUp",
			expectedPhysStates: []string{"LinkUp", "ETH_AN_FSM_ENABLE"},
			shouldPass:         true,
		},
		{
			name:               "ETH_AN_FSM_ENABLE state",
			actualPhysState:    "ETH_AN_FSM_ENABLE",
			expectedPhysStates: []string{"LinkUp", "ETH_AN_FSM_ENABLE"},
			shouldPass:         true,
		},
		{
			name:               "Invalid state",
			actualPhysState:    "LinkDown",
			expectedPhysStates: []string{"LinkUp", "ETH_AN_FSM_ENABLE"},
			shouldPass:         false,
		},
		{
			name:               "Empty state",
			actualPhysState:    "",
			expectedPhysStates: []string{"LinkUp", "ETH_AN_FSM_ENABLE"},
			shouldPass:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, expectedState := range tt.expectedPhysStates {
				if tt.actualPhysState == expectedState {
					found = true
					break
				}
			}
			if found != tt.shouldPass {
				t.Errorf("Expected %v for physical state validation, got %v", tt.shouldPass, found)
			}
		})
	}
}

// Test BER (Bit Error Rate) validation
func TestBERValidation(t *testing.T) {
	tests := []struct {
		name       string
		berValue   string
		threshold  float64
		shouldPass bool
	}{
		{
			name:       "Good BER below threshold",
			berValue:   "1E-13",
			threshold:  1E-12,
			shouldPass: true,
		},
		{
			name:       "Bad BER above threshold",
			berValue:   "1E-11",
			threshold:  1E-12,
			shouldPass: false,
		},
		{
			name:       "Invalid BER format",
			berValue:   "invalid",
			threshold:  1E-12,
			shouldPass: false,
		},
		{
			name:       "Empty BER",
			berValue:   "",
			threshold:  1E-12,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Helper function to validate float values
			isFloat := func(val string) bool {
				_, err := strconv.ParseFloat(val, 64)
				return err == nil
			}

			result := false
			if isFloat(tt.berValue) {
				if berFloat, err := strconv.ParseFloat(tt.berValue, 64); err == nil {
					result = berFloat < tt.threshold
				}
			}

			if result != tt.shouldPass {
				t.Errorf("Expected %v for BER validation, got %v", tt.shouldPass, result)
			}
		})
	}
}


// Test configuration structure
func TestLinkCheckTestConfig(t *testing.T) {
	config := LinkCheckTestConfig{
		IsEnabled:                           true,
		ExpectedSpeed:                       "200G",
		EffectivePhysicalErrorsThreshold:    0,
		RawPhysicalErrorsPerLaneThreshold:   10000,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.ExpectedSpeed != "200G" {
		t.Error("Expected speed to be 200G")
	}
	if config.EffectivePhysicalErrorsThreshold != 0 {
		t.Error("Expected effective physical errors threshold to be 0")
	}
	if config.RawPhysicalErrorsPerLaneThreshold != 10000 {
		t.Error("Expected raw physical errors per lane threshold to be 10000")
	}
}

// Test threshold configuration handling
func TestThresholdConfigurationHandling(t *testing.T) {
	tests := []struct {
		name              string
		threshold         interface{}
		expectedSpeed     string
	}{
		{
			name: "Map configuration",
			threshold: map[string]interface{}{
				"speed": "100G",
			},
			expectedSpeed: "100G",
		},
		{
			name:          "Invalid configuration",
			threshold:     "invalid",
			expectedSpeed: "200G", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate configuration processing
			config := &LinkCheckTestConfig{
				IsEnabled:                           true,
				ExpectedSpeed:                       "200G", // Default
				EffectivePhysicalErrorsThreshold:    0,
				RawPhysicalErrorsPerLaneThreshold:   10000,
			}

			switch v := tt.threshold.(type) {
			case map[string]interface{}:
				if speed, ok := v["speed"].(string); ok {
					config.ExpectedSpeed = speed
				}
			}

			if config.ExpectedSpeed != tt.expectedSpeed {
				t.Errorf("Expected speed %s, got %s", tt.expectedSpeed, config.ExpectedSpeed)
			}
		})
	}
}

// Test PrintLinkCheck function
func TestPrintLinkCheck(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintLinkCheck panicked: %v", r)
		}
	}()

	PrintLinkCheck()
}

// Test raw physical errors per lane validation
func TestRawPhysicalErrorsPerLaneValidation(t *testing.T) {
	tests := []struct {
		name        string
		errors      []int
		threshold   int
		shouldWarn  bool
	}{
		{
			name:       "All errors below threshold",
			errors:     []int{100, 200, 150, 300},
			threshold:  10000,
			shouldWarn: false,
		},
		{
			name:       "One error above threshold",
			errors:     []int{100, 15000, 150, 300},
			threshold:  10000,
			shouldWarn: true,
		},
		{
			name:       "All errors above threshold",
			errors:     []int{15000, 20000, 25000, 30000},
			threshold:  10000,
			shouldWarn: true,
		},
		{
			name:       "Empty errors list",
			errors:     []int{},
			threshold:  10000,
			shouldWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldWarn := false
			for _, laneError := range tt.errors {
				if laneError > tt.threshold {
					shouldWarn = true
					break
				}
			}

			if shouldWarn != tt.shouldWarn {
				t.Errorf("Expected warning=%v, got warning=%v", tt.shouldWarn, shouldWarn)
			}
		})
	}
}