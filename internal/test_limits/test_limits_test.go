package test_limits

import (
	"path/filepath"
	"testing"
)

func TestLoadTestLimits(t *testing.T) {
	// Test loading from default location
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	if limits == nil {
		t.Fatal("Expected test limits but got nil")
	}

	if len(limits.TestLimits) == 0 {
		t.Error("Expected non-empty test limits")
	}
}

func TestLoadTestLimitsFromFile(t *testing.T) {
	// Test loading from specific file
	packageDir, err := getPackageDir()
	if err != nil {
		t.Fatalf("Failed to get package directory: %v", err)
	}

	filePath := filepath.Join(packageDir, "test_limits.json")
	limits, err := LoadTestLimitsFromFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load test limits from file: %v", err)
	}

	if limits == nil {
		t.Fatal("Expected test limits but got nil")
	}

	// Test with non-existent file
	_, err = LoadTestLimitsFromFile("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetTestConfig(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	// Test valid configuration
	config, err := limits.GetTestConfig("BM.GPU.H100.8", "rx_discards_check")
	if err != nil {
		t.Errorf("Failed to get test config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected test config but got nil")
	}

	if !config.Enabled {
		t.Error("Expected GID index check to be enabled")
	}

	if config.TestCategory != "LEVEL_1" {
		t.Errorf("Expected test category 'LEVEL_1', got '%s'", config.TestCategory)
	}

	// Test invalid shape
	_, err = limits.GetTestConfig("INVALID.SHAPE", "rx_discards_check")
	if err == nil {
		t.Error("Expected error for invalid shape")
	}

	// Test invalid test type
	_, err = limits.GetTestConfig("BM.GPU.H100.8", "invalid_test")
	if err == nil {
		t.Error("Expected error for invalid test type")
	}
}

func TestIsTestEnabled(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	// Test enabled test
	enabled, err := limits.IsTestEnabled("BM.GPU.H100.8", "rx_discards_check")
	if err != nil {
		t.Errorf("Failed to check if test is enabled: %v", err)
	}
	if !enabled {
		t.Error("Expected rx_discards_check to be enabled for H100")
	}

	// Test disabled test
	enabled, err = limits.IsTestEnabled("BM.GPU.B200.8", "rx_discards_check")
	if err != nil {
		t.Errorf("Failed to check if test is enabled: %v", err)
	}
	if enabled {
		t.Error("Expected rx_discards_check to be disabled for B200")
	}

	// Test invalid shape
	_, err = limits.IsTestEnabled("INVALID.SHAPE", "rx_discards_check")
	if err == nil {
		t.Error("Expected error for invalid shape")
	}
}

func TestGetThresholdForTest(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	// Test RX discards threshold (number)
	threshold, err := limits.GetThresholdForTest("BM.GPU.H100.8", "rx_discards_check")
	if err != nil {
		t.Errorf("Failed to get RX discards threshold: %v", err)
	}
	if thresholdNum, ok := threshold.(float64); ok {
		if thresholdNum != 100 {
			t.Errorf("Expected RX discards threshold 100, got %v", thresholdNum)
		}
	} else {
		t.Error("Expected threshold to be a number")
	}

	// Test SRAM error threshold (object)
	threshold, err = limits.GetThresholdForTest("BM.GPU.H100.8", "sram_error_check")
	if err != nil {
		t.Errorf("Failed to get SRAM error threshold: %v", err)
	}
	if thresholdObj, ok := threshold.(map[string]interface{}); ok {
		if uncorrectable, exists := thresholdObj["uncorrectable"]; exists {
			if uncorrectable.(float64) != 5 {
				t.Errorf("Expected uncorrectable threshold 5, got %v", uncorrectable)
			}
		} else {
			t.Error("Expected uncorrectable field in SRAM threshold")
		}
	} else {
		t.Error("Expected threshold to be an object")
	}

	// Test GID index threshold (array)
	threshold, err = limits.GetThresholdForTest("BM.GPU.H100.8", "gid_index_check")
	if err != nil {
		t.Errorf("Failed to get GID index threshold: %v", err)
	}
	if threshold == nil {
		t.Error("Expected non-nil threshold")
	}

	// Test GPU mode threshold (object with allowed_modes array)
	threshold, err = limits.GetThresholdForTest("BM.GPU.H100.8", "gpu_mode_check")
	if err != nil {
		t.Errorf("Failed to get GPU mode threshold: %v", err)
	}
	if threshold == nil {
		t.Error("Expected non-nil threshold")
	}

	// Verify it's an object with allowed_modes
	if thresholdObj, ok := threshold.(map[string]interface{}); ok {
		if allowedModes, exists := thresholdObj["allowed_modes"]; exists {
			if allowedModesArray, ok := allowedModes.([]interface{}); ok {
				if len(allowedModesArray) != 3 {
					t.Errorf("Expected 3 allowed modes, got %d", len(allowedModesArray))
				}
				// Verify the expected modes are present
				expectedModes := map[string]bool{"N/A": false, "DISABLED": false, "ENABLED": false}
				for _, mode := range allowedModesArray {
					if modeStr, ok := mode.(string); ok {
						if _, exists := expectedModes[modeStr]; exists {
							expectedModes[modeStr] = true
						} else {
							t.Errorf("Unexpected allowed mode: %s", modeStr)
						}
					} else {
						t.Error("Expected mode to be a string")
					}
				}
				// Check all expected modes were found
				for mode, found := range expectedModes {
					if !found {
						t.Errorf("Expected mode %s not found in allowed_modes", mode)
					}
				}
			} else {
				t.Error("Expected allowed_modes to be an array")
			}
		} else {
			t.Error("Expected allowed_modes field in GPU mode threshold")
		}
	} else {
		t.Error("Expected threshold to be an object")
	}

	// Test nvlink speed threshold (object)
	threshold, err = limits.GetThresholdForTest("BM.GPU.H100.8", "nvlink_speed_check")
	if err != nil {
		t.Errorf("Failed to get nvlink speed threshold: %v", err)
	}
	if thresholdObj, ok := threshold.(map[string]interface{}); ok {
		if speed, exists := thresholdObj["speed"]; exists {
			if speed.(float64) != 26 {
				t.Errorf("Expected nvlink speed threshold 26, got %v", speed)
			}
		} else {
			t.Error("Expected speed field in nvlink threshold")
		}
		if count, exists := thresholdObj["count"]; exists {
			if count.(float64) != 18 {
				t.Errorf("Expected nvlink count threshold 18, got %v", count)
			}
		} else {
			t.Error("Expected count field in nvlink threshold")
		}
	} else {
		t.Error("Expected threshold to be an object")
	}

	// Test disabled test
	_, err = limits.GetThresholdForTest("BM.GPU.B200.8", "rx_discards_check")
	if err == nil {
		t.Error("Expected error for disabled test")
	}

	// Test invalid test type
	_, err = limits.GetThresholdForTest("BM.GPU.H100.8", "invalid_test")
	if err == nil {
		t.Error("Expected error for invalid test type")
	}
}

func TestGetAvailableShapes(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	shapes := limits.GetAvailableShapes()
	if len(shapes) != 3 {
		t.Errorf("Expected 3 shapes, got %d", len(shapes))
	}

	expectedShapes := map[string]bool{
		"BM.GPU.H100.8":  false,
		"BM.GPU.B200.8":  false,
		"BM.GPU.GB200.4": false,
	}

	for _, shape := range shapes {
		if _, exists := expectedShapes[shape]; !exists {
			t.Errorf("Unexpected shape: %s", shape)
		} else {
			expectedShapes[shape] = true
		}
	}

	// Verify all expected shapes were found
	for shape, found := range expectedShapes {
		if !found {
			t.Errorf("Expected shape %s not found", shape)
		}
	}
}

func TestGetEnabledTests(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	// Test H100 shape (all tests enabled)
	enabledTests, err := limits.GetEnabledTests("BM.GPU.H100.8")
	if err != nil {
		t.Errorf("Failed to get enabled tests: %v", err)
	}
	if len(enabledTests) != 12 {
		t.Errorf("Expected 12 enabled tests for H100, got %d", len(enabledTests))
	}

	expectedTests := map[string]bool{
		"gid_index_check":      false,
		"gpu_mode_check":       false,
		"rx_discards_check":    false,
		"sram_error_check":     false,
		"gpu_count_check":      false,
		"rdma_nic_count":       false,
		"pcie_error_check":     false,
		"link_check":           false,
		"eth_link_check":       false,
		"gpu_driver_check":     false,
		"peermem_module_check": false,
		"nvlink_speed_check":   false,
	}

	for _, test := range enabledTests {
		if _, exists := expectedTests[test]; !exists {
			t.Errorf("Unexpected enabled test: %s", test)
		} else {
			expectedTests[test] = true
		}
	}

	// Test B200 shape (all tests disabled)
	enabledTests, err = limits.GetEnabledTests("BM.GPU.B200.8")
	if err != nil {
		t.Errorf("Failed to get enabled tests: %v", err)
	}
	if len(enabledTests) != 0 {
		t.Errorf("Expected 0 enabled tests for B200, got %d", len(enabledTests))
	}

	// Test invalid shape
	_, err = limits.GetEnabledTests("INVALID.SHAPE")
	if err == nil {
		t.Error("Expected error for invalid shape")
	}
}

func TestPackageHelperFunctions(t *testing.T) {
	// Test getPackageDir
	dir, err := getPackageDir()
	if err != nil {
		t.Errorf("Failed to get package directory: %v", err)
	}
	if dir == "" {
		t.Error("Expected non-empty package directory")
	}

	// Test getDefaultConfigPath
	path, err := getDefaultConfigPath()
	if err != nil {
		t.Errorf("Failed to get default config path: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty config path")
	}
	// With the new multi-path resolution, we might get a relative path
	// if ./test_limits.json exists (which it does in the test environment)
	if path == "" {
		t.Error("Expected non-empty config path")
	}
}

func TestJSONStructureParsing(t *testing.T) {
	limits, err := LoadTestLimits()
	if err != nil {
		t.Fatalf("Failed to load test limits: %v", err)
	}

	// Verify the JSON structure was parsed correctly
	if limits.TestLimits == nil {
		t.Fatal("Expected test limits map but got nil")
	}

	// Check H100 configuration exists
	h100Config, exists := limits.TestLimits["BM.GPU.H100.8"]
	if !exists {
		t.Fatal("Expected BM.GPU.H100.8 configuration")
	}

	// Check test configurations
	gidConfig, exists := h100Config["gid_index_check"]
	if !exists {
		t.Error("Expected gid_index_check configuration")
	} else {
		if !gidConfig.Enabled {
			t.Error("Expected gid_index_check to be enabled")
		}
		if gidConfig.TestCategory != "LEVEL_1" {
			t.Errorf("Expected test category LEVEL_1, got %s", gidConfig.TestCategory)
		}
	}

	rxConfig, exists := h100Config["rx_discards_check"]
	if !exists {
		t.Error("Expected rx_discards_check configuration")
	} else {
		if !rxConfig.Enabled {
			t.Error("Expected rx_discards_check to be enabled")
		}
	}

	sramConfig, exists := h100Config["sram_error_check"]
	if !exists {
		t.Error("Expected sram_error_check configuration")
	} else {
		if !sramConfig.Enabled {
			t.Error("Expected sram_error_check to be enabled")
		}
	}

	// Check GPU mode configuration
	gpuModeConfig, exists := h100Config["gpu_mode_check"]
	if !exists {
		t.Error("Expected gpu_mode_check configuration")
	} else {
		if !gpuModeConfig.Enabled {
			t.Error("Expected gpu_mode_check to be enabled")
		}
		if gpuModeConfig.TestCategory != "LEVEL_1" {
			t.Errorf("Expected test category LEVEL_1, got %s", gpuModeConfig.TestCategory)
		}
		// Verify threshold structure
		if gpuModeConfig.Threshold == nil {
			t.Error("Expected gpu_mode_check to have threshold configuration")
		}
	}

	// Check nvlink speed configuration
	nvlinkConfig, exists := h100Config["nvlink_speed_check"]
	if !exists {
		t.Error("Expected nvlink_speed_check configuration")
	} else {
		if !nvlinkConfig.Enabled {
			t.Error("Expected nvlink_speed_check to be enabled")
		}
		if nvlinkConfig.TestCategory != "LEVEL_1" {
			t.Errorf("Expected test category LEVEL_1, got %s", nvlinkConfig.TestCategory)
		}
		// Verify threshold structure
		if nvlinkConfig.Threshold == nil {
			t.Error("Expected nvlink_speed_check to have threshold configuration")
		}
	}
}
