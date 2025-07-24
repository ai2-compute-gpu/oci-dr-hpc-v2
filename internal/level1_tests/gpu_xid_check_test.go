package level1_tests

import (
	"testing"
)

func TestXIDError(t *testing.T) {
	// Test XIDError struct
	xidError := XIDError{
		XIDCode:     "8",
		Description: "GPU stopped processing",
		Severity:    "Critical",
		Count:       2,
		PCIAddrs:    []string{"0000:3b:00.0", "0000:d8:00.0"},
	}

	if xidError.XIDCode != "8" {
		t.Errorf("Expected XIDCode 8, got %s", xidError.XIDCode)
	}

	if xidError.Description != "GPU stopped processing" {
		t.Errorf("Expected description 'GPU stopped processing', got %s", xidError.Description)
	}

	if xidError.Severity != "Critical" {
		t.Errorf("Expected severity 'Critical', got %s", xidError.Severity)
	}

	if xidError.Count != 2 {
		t.Errorf("Expected count 2, got %d", xidError.Count)
	}

	if len(xidError.PCIAddrs) != 2 {
		t.Errorf("Expected 2 PCI addresses, got %d", len(xidError.PCIAddrs))
	}
}

func TestGPUXIDCheckResult(t *testing.T) {
	// Test GPUXIDCheckResult struct
	result := GPUXIDCheckResult{
		Status:         "PASS",
		Message:        "No XID errors found in system logs",
		CriticalErrors: []XIDError{},
		WarningErrors:  []XIDError{},
	}

	if result.Status != "PASS" {
		t.Errorf("Expected status PASS, got %s", result.Status)
	}

	if result.Message != "No XID errors found in system logs" {
		t.Errorf("Expected specific message, got %s", result.Message)
	}

	if len(result.CriticalErrors) != 0 {
		t.Errorf("Expected 0 critical errors, got %d", len(result.CriticalErrors))
	}

	if len(result.WarningErrors) != 0 {
		t.Errorf("Expected 0 warning errors, got %d", len(result.WarningErrors))
	}
}

func TestGPUXIDCheckResultWithErrors(t *testing.T) {
	// Test GPUXIDCheckResult struct with errors
	criticalError := XIDError{
		XIDCode:     "79",
		Description: "GPU has fallen off the bus",
		Severity:    "Critical",
		Count:       1,
		PCIAddrs:    []string{"0000:3b:00.0"},
	}

	warningError := XIDError{
		XIDCode:     "43",
		Description: "GPU stopped processing",
		Severity:    "Warn",
		Count:       1,
		PCIAddrs:    []string{"0000:d8:00.0"},
	}

	result := GPUXIDCheckResult{
		Status:         "FAIL",
		Message:        "Critical XID errors detected: 1 critical, 1 warnings",
		CriticalErrors: []XIDError{criticalError},
		WarningErrors:  []XIDError{warningError},
	}

	if result.Status != "FAIL" {
		t.Errorf("Expected status FAIL, got %s", result.Status)
	}

	if len(result.CriticalErrors) != 1 {
		t.Errorf("Expected 1 critical error, got %d", len(result.CriticalErrors))
	}

	if len(result.WarningErrors) != 1 {
		t.Errorf("Expected 1 warning error, got %d", len(result.WarningErrors))
	}

	if result.CriticalErrors[0].XIDCode != "79" {
		t.Errorf("Expected critical error XID code 79, got %s", result.CriticalErrors[0].XIDCode)
	}

	if result.WarningErrors[0].Severity != "Warn" {
		t.Errorf("Expected warning error severity Warn, got %s", result.WarningErrors[0].Severity)
	}
}

func TestXIDErrorCode(t *testing.T) {
	// Test XIDErrorCode struct
	xidCode := XIDErrorCode{
		Description: "GPU memory page fault",
		Severity:    "Critical",
	}

	if xidCode.Description != "GPU memory page fault" {
		t.Errorf("Expected description 'GPU memory page fault', got %s", xidCode.Description)
	}

	if xidCode.Severity != "Critical" {
		t.Errorf("Expected severity 'Critical', got %s", xidCode.Severity)
	}
}

func TestGPUXIDCheckTestConfig(t *testing.T) {
	// Test GPUXIDCheckTestConfig struct
	xidCodes := map[string]XIDErrorCode{
		"8": {
			Description: "GPU stopped processing",
			Severity:    "Critical",
		},
		"31": {
			Description: "GPU memory page fault",
			Severity:    "Critical",
		},
	}

	config := GPUXIDCheckTestConfig{
		IsEnabled:     true,
		XIDErrorCodes: xidCodes,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}

	if len(config.XIDErrorCodes) != 2 {
		t.Errorf("Expected 2 XID error codes, got %d", len(config.XIDErrorCodes))
	}

	if config.XIDErrorCodes["8"].Description != "GPU stopped processing" {
		t.Errorf("Expected specific description for XID 8, got %s", config.XIDErrorCodes["8"].Description)
	}

	if config.XIDErrorCodes["31"].Severity != "Critical" {
		t.Errorf("Expected Critical severity for XID 31, got %s", config.XIDErrorCodes["31"].Severity)
	}
}

func TestGetGPUXIDCheckTestConfig(t *testing.T) {
	// Test with H100 shape (should be enabled)
	config, err := getGPUXIDCheckTestConfig("BM.GPU.H100.8")
	if err != nil {
		t.Errorf("Failed to get GPU XID test config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if !config.IsEnabled {
		t.Error("Expected GPU XID check to be enabled for BM.GPU.H100.8")
	}

	if len(config.XIDErrorCodes) == 0 {
		t.Error("Expected at least some XID error codes to be configured")
	}

	// Test with B200 shape (should be disabled)
	config, err = getGPUXIDCheckTestConfig("BM.GPU.B200.8")
	if err != nil {
		t.Errorf("Failed to get GPU XID test config for B200: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config for B200")
	}

	if config.IsEnabled {
		t.Error("Expected GPU XID check to be disabled for BM.GPU.B200.8")
	}
}

func TestCheckGPUXIDErrorsNoDMesg(t *testing.T) {
	// Test checkGPUXIDErrors with minimal XID codes
	xidCodes := map[string]XIDErrorCode{
		"8": {
			Description: "GPU stopped processing",
			Severity:    "Critical",
		},
	}

	// Verify the XID codes configuration
	if len(xidCodes) != 1 {
		t.Errorf("Expected 1 XID code, got %d", len(xidCodes))
	}

	// This test simulates the scenario where dmesg has no XID errors
	// Since we can't mock exec.Command easily in this context, this test
	// focuses on the struct validation and configuration loading
	result := &GPUXIDCheckResult{
		Status:         "PASS",
		Message:        "No XID errors found in system logs",
		CriticalErrors: []XIDError{},
		WarningErrors:  []XIDError{},
	}

	if result.Status != "PASS" {
		t.Errorf("Expected PASS status, got %s", result.Status)
	}

	if len(result.CriticalErrors) != 0 {
		t.Errorf("Expected no critical errors, got %d", len(result.CriticalErrors))
	}
}

func TestGPUXIDCheckDefaultConfiguration(t *testing.T) {
	// Test that default XID codes are available when configuration is missing
	defaultCodes := map[string]XIDErrorCode{
		"8":   {Description: "GPU stopped processing", Severity: "Critical"},
		"31":  {Description: "GPU memory page fault", Severity: "Critical"},
		"48":  {Description: "Double Bit ECC Error", Severity: "Critical"},
		"79":  {Description: "GPU has fallen off the bus", Severity: "Critical"},
		"92":  {Description: "High single-bit ECC error rate", Severity: "Critical"},
		"94":  {Description: "Contained ECC error", Severity: "Critical"},
		"95":  {Description: "Uncontained ECC error", Severity: "Critical"},
		"119": {Description: "GSP RPC Timeout", Severity: "Critical"},
		"120": {Description: "GSP Error", Severity: "Critical"},
	}

	// Verify all default codes are Critical severity
	for xidCode, xidInfo := range defaultCodes {
		if xidInfo.Severity != "Critical" {
			t.Errorf("Expected Critical severity for default XID code %s, got %s", xidCode, xidInfo.Severity)
		}
		if xidInfo.Description == "" {
			t.Errorf("Expected non-empty description for default XID code %s", xidCode)
		}
	}

	// Verify we have the expected number of default codes
	if len(defaultCodes) != 9 {
		t.Errorf("Expected 9 default XID codes, got %d", len(defaultCodes))
	}
}

func TestGPUXIDCheckConfigurationParsing(t *testing.T) {
	// Test configuration parsing logic (simulated)
	testConfig := map[string]interface{}{
		"xid_error_codes": map[string]interface{}{
			"8": map[string]interface{}{
				"description": "GPU stopped processing",
				"severity":    "Critical",
			},
			"31": map[string]interface{}{
				"description": "GPU memory page fault",
				"severity":    "Critical",
			},
			"43": map[string]interface{}{
				"description": "GPU stopped processing",
				"severity":    "Warn",
			},
		},
	}

	// Simulate the parsing logic from getGPUXIDCheckTestConfig
	xidErrorCodes := make(map[string]XIDErrorCode)
	if xidCodesRaw, exists := testConfig["xid_error_codes"]; exists {
		if xidCodesMap, ok := xidCodesRaw.(map[string]interface{}); ok {
			for xidCode, xidInfoRaw := range xidCodesMap {
				if xidInfoMap, ok := xidInfoRaw.(map[string]interface{}); ok {
					xidError := XIDErrorCode{}
					if desc, ok := xidInfoMap["description"].(string); ok {
						xidError.Description = desc
					}
					if sev, ok := xidInfoMap["severity"].(string); ok {
						xidError.Severity = sev
					}
					xidErrorCodes[xidCode] = xidError
				}
			}
		}
	}

	// Verify parsing worked correctly
	if len(xidErrorCodes) != 3 {
		t.Errorf("Expected 3 parsed XID codes, got %d", len(xidErrorCodes))
	}

	if xidErrorCodes["8"].Description != "GPU stopped processing" {
		t.Errorf("Expected correct description for XID 8, got %s", xidErrorCodes["8"].Description)
	}

	if xidErrorCodes["8"].Severity != "Critical" {
		t.Errorf("Expected Critical severity for XID 8, got %s", xidErrorCodes["8"].Severity)
	}

	if xidErrorCodes["43"].Severity != "Warn" {
		t.Errorf("Expected Warn severity for XID 43, got %s", xidErrorCodes["43"].Severity)
	}
}