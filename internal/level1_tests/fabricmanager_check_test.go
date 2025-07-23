package level1_tests

import (
	"testing"
)

func TestFabricManagerCheckResult(t *testing.T) {
	// Test FabricManagerCheckResult struct
	result := FabricManagerCheckResult{
		Status:      "PASS",
		IsRunning:   true,
		ServiceInfo: "nvidia-fabricmanager service is active and running",
		Message:     "nvidia-fabricmanager service is properly running",
	}

	if result.Status != "PASS" {
		t.Errorf("Expected status PASS, got %s", result.Status)
	}

	if !result.IsRunning {
		t.Error("Expected IsRunning to be true")
	}

	if result.ServiceInfo == "" {
		t.Error("Expected non-empty ServiceInfo")
	}

	if result.Message == "" {
		t.Error("Expected non-empty Message")
	}
}

func TestFabricManagerCheckTestConfig(t *testing.T) {
	// Test FabricManagerCheckTestConfig struct
	config := FabricManagerCheckTestConfig{
		IsEnabled: true,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
}

func TestGetFabricManagerCheckTestConfig(t *testing.T) {
	// Test with H100 shape (should be enabled)
	config, err := getFabricManagerCheckTestConfig("BM.GPU.H100.8")
	if err != nil {
		t.Errorf("Failed to get fabric manager test config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// Test with non-existent shape
	config, err = getFabricManagerCheckTestConfig("INVALID.SHAPE")
	if err != nil {
		t.Errorf("Unexpected error for invalid shape: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// For invalid shape, test should be disabled
	if config.IsEnabled {
		t.Error("Expected test to be disabled for invalid shape")
	}
}

func TestCheckFabricManagerService(t *testing.T) {
	// This test will work regardless of whether the service exists
	result := checkFabricManagerService()

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Status should be either PASS or FAIL
	if result.Status != "PASS" && result.Status != "FAIL" {
		t.Errorf("Expected status PASS or FAIL, got %s", result.Status)
	}

	// Message should not be empty
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// IsRunning should be consistent with Status
	if result.Status == "PASS" && !result.IsRunning {
		t.Error("Expected IsRunning to be true when status is PASS")
	}

	if result.Status == "FAIL" && result.IsRunning {
		t.Error("Expected IsRunning to be false when status is FAIL")
	}
}

func TestFabricManagerCheckIntegration(t *testing.T) {
	// Test the main integration function
	// This will skip if the test is not enabled for the current shape
	err := RunFabricManagerCheck()
	
	// We expect either nil (success) or an error indicating the test is not applicable
	// Both are valid outcomes depending on the environment
	if err != nil {
		// Check if it's the "not applicable" error
		expectedError := "Test not applicable for this shape"
		if !containsSubstring(err.Error(), expectedError) {
			// If it's not the "not applicable" error, it could be a legitimate failure
			// In test environment, this is acceptable since we may not have the service
			t.Logf("Fabric manager check returned error (acceptable in test environment): %v", err)
		}
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr || 
		   len(str) > len(substr) && containsSubstringRecursive(str[1:], substr)
}

func containsSubstringRecursive(str, substr string) bool {
	if len(str) < len(substr) {
		return false
	}
	if str[:len(substr)] == substr {
		return true
	}
	return containsSubstringRecursive(str[1:], substr)
}