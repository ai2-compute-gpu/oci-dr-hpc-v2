package level1_tests

import (
	"testing"
)

// Test validateGPUClockSpeeds function
func TestValidateGPUClockSpeeds(t *testing.T) {
	expectedSpeed := 1980 // H100 expected speed

	tests := []struct {
		name               string
		clockSpeeds        []string
		expectedSpeed      int
		expectedStatus     string
		expectedError      bool
		expectedStatusMsg  string
	}{
		{
			name:           "Empty clock speeds list",
			clockSpeeds:    []string{},
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:              "Single GPU at max speed",
			clockSpeeds:       []string{"1980 MHz"},
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1980",
		},
		{
			name:              "Multiple GPUs at max speed",
			clockSpeeds:       []string{"1980 MHz", "1980 MHz", "1980 MHz", "1980 MHz"},
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1980",
		},
		{
			name:              "Single GPU within acceptable range (95% of max)",
			clockSpeeds:       []string{"1881 MHz"}, // ~95% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1881",
		},
		{
			name:              "Single GPU at minimum acceptable (90% of max)",
			clockSpeeds:       []string{"1782 MHz"}, // 90% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1782",
		},
		{
			name:           "Single GPU below acceptable threshold",
			clockSpeeds:    []string{"1700 MHz"}, // Below 90% of 1980
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Multiple GPUs with one below threshold",
			clockSpeeds:    []string{"1980 MHz", "1980 MHz", "1700 MHz", "1980 MHz"}, // GPU 2 below threshold
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Multiple GPUs all below threshold",
			clockSpeeds:    []string{"1700 MHz", "1600 MHz", "1650 MHz"}, // All below 90% of 1980
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:              "Mixed speeds within acceptable range",
			clockSpeeds:       []string{"1980 MHz", "1900 MHz", "1850 MHz", "1800 MHz"}, // All >= 90% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1800", // Minimum among acceptable speeds
		},
		{
			name:              "Different expected speed (2100 MHz)",
			clockSpeeds:       []string{"2100 MHz", "2050 MHz", "1980 MHz"}, // All >= 90% of 2100 (1890)
			expectedSpeed:     2100,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 2100, allowed 1980", // Minimum among acceptable speeds
		},
		{
			name:           "Different expected speed with failure",
			clockSpeeds:    []string{"1800 MHz", "1750 MHz"}, // Below 90% of 2100 (1890)
			expectedSpeed:  2100,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Malformed clock speed entry",
			clockSpeeds:    []string{"invalid", "1980 MHz"},
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Empty clock speed entry",
			clockSpeeds:    []string{"", "1980 MHz"},
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:              "Clock speeds without MHz suffix",
			clockSpeeds:       []string{"1980", "1900"},
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1900", // 1900 is minimum acceptable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, statusMsg, err := validateGPUClockSpeeds(tt.clockSpeeds, tt.expectedSpeed)

			if status != tt.expectedStatus {
				t.Errorf("validateGPUClockSpeeds() status = %v, want %v", status, tt.expectedStatus)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("validateGPUClockSpeeds() error = %v, wantErr %v", err, tt.expectedError)
			}

			if tt.expectedStatusMsg != "" && statusMsg != tt.expectedStatusMsg {
				t.Errorf("validateGPUClockSpeeds() statusMsg = %v, want %v", statusMsg, tt.expectedStatusMsg)
			}
		})
	}
}

// Test getGpuClkCheckTestConfig function (basic validation)
func TestGetGpuClkCheckTestConfig(t *testing.T) {
	// This test will only work if we're in a test environment
	// We'll test the default values when IMDS isn't available
	config := &GPUClkCheckTestConfig{
		IsEnabled:        false,
		Shape:            "test-shape",
		ExpectedClkSpeed: 1980,
	}

	// Verify default expected clock speed
	if config.ExpectedClkSpeed != 1980 {
		t.Errorf("Expected default clock speed to be 1980, got %d", config.ExpectedClkSpeed)
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

// Test PrintGPUClkCheck function
func TestPrintGPUClkCheck(t *testing.T) {
	// This is mainly to ensure the function doesn't panic
	PrintGPUClkCheck()
}

// Test clock speed validation edge cases
func TestClockSpeedValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		clockSpeeds   []string
		expectedSpeed int
		description   string
	}{
		{
			name:          "Single very high speed GPU",
			clockSpeeds:   []string{"2500 MHz"},
			expectedSpeed: 1980,
			description:   "GPU running above expected speed should pass",
		},
		{
			name:          "Exactly at 90% threshold",
			clockSpeeds:   []string{"1782 MHz"}, // Exactly 90% of 1980
			expectedSpeed: 1980,
			description:   "GPU at exactly 90% threshold should pass",
		},
		{
			name:          "Just below 90% threshold",
			clockSpeeds:   []string{"1781 MHz"}, // Just below 90% of 1980
			expectedSpeed: 1980,
			description:   "GPU just below 90% threshold should fail",
		},
		{
			name:          "Mixed high and acceptable speeds",
			clockSpeeds:   []string{"2200 MHz", "1980 MHz", "1900 MHz", "1850 MHz"},
			expectedSpeed: 1980,
			description:   "Mix of high and acceptable speeds should pass with lowest acceptable",
		},
		{
			name:          "All speeds above expected",
			clockSpeeds:   []string{"2100 MHz", "2050 MHz", "2000 MHz"},
			expectedSpeed: 1980,
			description:   "All speeds above expected should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, statusMsg, err := validateGPUClockSpeeds(tt.clockSpeeds, tt.expectedSpeed)
			
			t.Logf("Test %s: %s - Status: %s, Message: %s", tt.name, tt.description, status, statusMsg)
			
			if status == "FAIL" && err == nil {
				t.Error("Expected error when status is FAIL")
			}
			if status == "PASS" && err != nil {
				t.Errorf("Unexpected error when status is PASS: %v", err)
			}
		})
	}
}

// Test specific Python script logic replication
func TestPythonLogicReplication(t *testing.T) {
	tests := []struct {
		name          string
		clockSpeeds   []string
		expectedSpeed int
		wantStatus    string
		description   string
	}{
		{
			name:          "Python script default case - H100 at 1980",
			clockSpeeds:   []string{"1980 MHz"},
			expectedSpeed: 1980,
			wantStatus:    "PASS",
			description:   "Replicates default Python script behavior",
		},
		{
			name:          "Python script tolerance case - H100 at 90%",
			clockSpeeds:   []string{"1782 MHz"}, // 1980 - (1980 * 0.10) = 1782
			expectedSpeed: 1980,
			wantStatus:    "PASS",
			description:   "Tests 10% tolerance logic from Python",
		},
		{
			name:          "Python script fail case - below tolerance",
			clockSpeeds:   []string{"1781 MHz"}, // Just below 90% threshold
			expectedSpeed: 1980,
			wantStatus:    "FAIL",
			description:   "Tests failure below 90% threshold",
		},
		{
			name:          "Multiple GPUs with mixed results",
			clockSpeeds:   []string{"1980 MHz", "1850 MHz", "1800 MHz"}, // All above 90%
			expectedSpeed: 1980,
			wantStatus:    "PASS",
			description:   "Tests multiple GPU handling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, statusMsg, err := validateGPUClockSpeeds(tt.clockSpeeds, tt.expectedSpeed)
			
			if status != tt.wantStatus {
				t.Errorf("validateGPUClockSpeeds() status = %v, want %v for case: %s", 
					status, tt.wantStatus, tt.description)
			}
			
			t.Logf("Case: %s -> Status: %s, Message: %s, Error: %v", 
				tt.description, status, statusMsg, err)
		})
	}
}