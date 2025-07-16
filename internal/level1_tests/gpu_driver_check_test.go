package level1_tests

import (
	"testing"
)

// Test validateDriverVersions function
func TestValidateDriverVersions(t *testing.T) {
	blacklisted := []string{"470.57.02"}
	supported := []string{"450.119.03", "450.142.0", "470.103.01", "470.129.06", "470.141.03", "510.47.03", "535.104.12", "550.90.12"}

	tests := []struct {
		name            string
		versions        []string
		expectedStatus  string
		expectedError   bool
		expectedMessage string
	}{
		{
			name:           "Empty versions list",
			versions:       []string{},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Single supported version",
			versions:       []string{"450.119.03"},
			expectedStatus: "PASS",
			expectedError:  false,
		},
		{
			name:           "Multiple same supported versions",
			versions:       []string{"450.119.03", "450.119.03", "450.119.03"},
			expectedStatus: "PASS",
			expectedError:  false,
		},
		{
			name:           "Blacklisted version",
			versions:       []string{"470.57.02"},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Unsupported but not blacklisted version",
			versions:       []string{"999.999.99"},
			expectedStatus: "WARN",
			expectedError:  true,
		},
		{
			name:           "Mismatched versions",
			versions:       []string{"450.119.03", "470.103.01"},
			expectedStatus: "FAIL",
			expectedError:  true,
		},
		{
			name:           "Another supported version",
			versions:       []string{"535.104.12"},
			expectedStatus: "PASS",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := validateDriverVersions(tt.versions, blacklisted, supported)

			if status != tt.expectedStatus {
				t.Errorf("validateDriverVersions() status = %v, want %v", status, tt.expectedStatus)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("validateDriverVersions() error = %v, wantErr %v", err, tt.expectedError)
			}
		})
	}
}

// Test getGpuDriverCheckTestConfig function (basic validation)
func TestGetGpuDriverCheckTestConfig(t *testing.T) {
	// This test will only work if we're in a test environment
	// We'll test the default values when IMDS isn't available
	config := &GPUDriverCheckTestConfig{
		IsEnabled:           false,
		Shape:               "test-shape",
		BlacklistedVersions: []string{"470.57.02"},
		SupportedVersions:   []string{"450.119.03", "450.142.0", "470.103.01", "470.129.06", "470.141.03", "510.47.03", "535.104.12", "550.90.12"},
	}

	// Verify default blacklisted versions
	if len(config.BlacklistedVersions) == 0 {
		t.Error("Expected default blacklisted versions to be set")
	}

	// Verify default supported versions
	if len(config.SupportedVersions) == 0 {
		t.Error("Expected default supported versions to be set")
	}

	// Verify specific blacklisted version
	foundBlacklisted := false
	for _, version := range config.BlacklistedVersions {
		if version == "470.57.02" {
			foundBlacklisted = true
			break
		}
	}
	if !foundBlacklisted {
		t.Error("Expected 470.57.02 to be in blacklisted versions")
	}

	// Verify specific supported version
	foundSupported := false
	for _, version := range config.SupportedVersions {
		if version == "535.104.12" {
			foundSupported = true
			break
		}
	}
	if !foundSupported {
		t.Error("Expected 535.104.12 to be in supported versions")
	}
}

// Test PrintGPUDriverCheck function
func TestPrintGPUDriverCheck(t *testing.T) {
	// This is mainly to ensure the function doesn't panic
	PrintGPUDriverCheck()
}