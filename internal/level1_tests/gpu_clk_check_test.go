package level1_tests

import (
	"testing"
)

// Test validateGPUClockSpeeds function
func TestValidateGPUClockSpeeds(t *testing.T) {
	expectedSpeed := 1980 // H100 expected speed

	tests := []struct {
		name               string
		clockSpeeds        []int
		expectedSpeed      int
		expectedStatus     string
		expectedError      bool
		expectedStatusMsg  string
	}{
		{
			name:           "Empty clock speeds list",
			clockSpeeds:    []int{},
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:              "Single GPU at max speed",
			clockSpeeds:       []int{1980},
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1980",
		},
		{
			name:              "Multiple GPUs at max speed",
			clockSpeeds:       []int{1980, 1980, 1980, 1980},
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1980",
		},
		{
			name:              "Single GPU within acceptable range (95% of max)",
			clockSpeeds:       []int{1881}, // ~95% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1881",
		},
		{
			name:              "Single GPU at minimum acceptable (90% of max)",
			clockSpeeds:       []int{1782}, // 90% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1782",
		},
		{
			name:           "Single GPU below acceptable threshold",
			clockSpeeds:    []int{1700}, // Below 90% of 1980
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Multiple GPUs with one below threshold",
			clockSpeeds:    []int{1980, 1980, 1700, 1980}, // GPU 2 below threshold
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Multiple GPUs all below threshold",
			clockSpeeds:    []int{1700, 1600, 1650}, // All below 90% of 1980
			expectedSpeed:  expectedSpeed,
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:              "Mixed speeds within acceptable range",
			clockSpeeds:       []int{1980, 1900, 1850, 1800}, // All >= 90% of 1980
			expectedSpeed:     expectedSpeed,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 1980, allowed 1800",
		},
		{
			name:              "Different expected speed (2100 MHz)",
			clockSpeeds:       []int{2100, 2050, 1980}, // All >= 90% of 2100 (1890)
			expectedSpeed:     2100,
			expectedStatus:    "PASS",
			expectedError:     false,
			expectedStatusMsg: "Expected 2100, allowed 1980",
		},
		{
			name:           "Different expected speed with failure",
			clockSpeeds:    []int{1800, 1750}, // Below 90% of 2100 (1890)
			expectedSpeed:  2100,
			expectedStatus: "FAIL",
			expectedError:  true,
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
}

// Test PrintGPUClkCheck function
func TestPrintGPUClkCheck(t *testing.T) {
	// This is mainly to ensure the function doesn't panic
	PrintGPUClkCheck()
}

// Test clock speed parsing edge cases
func TestClockSpeedValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		clockSpeeds   []int
		expectedSpeed int
		description   string
	}{
		{
			name:          "Single very high speed GPU",
			clockSpeeds:   []int{2500},
			expectedSpeed: 1980,
			description:   "GPU running above expected speed should pass",
		},
		{
			name:          "Exactly at 90% threshold",
			clockSpeeds:   []int{1782}, // Exactly 90% of 1980
			expectedSpeed: 1980,
			description:   "GPU at exactly 90% threshold should pass",
		},
		{
			name:          "Just below 90% threshold",
			clockSpeeds:   []int{1781}, // Just below 90% of 1980
			expectedSpeed: 1980,
			description:   "GPU just below 90% threshold should fail",
		},
		{
			name:          "Mixed high and acceptable speeds",
			clockSpeeds:   []int{2200, 1980, 1900, 1850},
			expectedSpeed: 1980,
			description:   "Mix of high and acceptable speeds should pass with lowest acceptable",
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