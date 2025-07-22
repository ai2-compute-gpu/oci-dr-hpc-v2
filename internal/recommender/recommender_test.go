package recommender

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test helper functions

func createTestReport() RecommendationReport {
	return RecommendationReport{
		Summary:        "Test summary with 2 critical issues",
		TotalIssues:    2,
		CriticalIssues: 2,
		WarningIssues:  0,
		InfoIssues:     0,
		Recommendations: []Recommendation{
			{
				Type:       "critical",
				TestName:   "gpu_count_check",
				Issue:      "GPU count mismatch detected",
				Suggestion: "Verify GPU hardware installation",
				Commands:   []string{"nvidia-smi"},
			},
			{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      "GPU MIG mode violation detected",
				Suggestion: "Disable MIG mode on affected GPUs",
				Commands:   []string{"sudo nvidia-smi -mig 0"},
			},
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func createTestConfig() RecommendationConfig {
	return RecommendationConfig{
		Recommendations: map[string]TestRecommendations{
			"gpu_count_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU count mismatch detected (found: {gpu_count})",
					Suggestion: "Verify GPU hardware installation",
					Commands:   []string{"nvidia-smi"},
				},
				Pass: &RecommendationTemplate{
					Type:       "info",
					Issue:      "GPU count check passed ({gpu_count} GPUs detected)",
					Suggestion: "GPU hardware is properly configured",
				},
			},
			"gpu_mode_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU MIG mode violation on GPUs: {enabled_gpu_indexes}",
					Suggestion: "Disable MIG mode on affected GPUs",
					Commands:   []string{"sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0"},
				},
			},
			"peermem_module_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "NVIDIA Peer Memory module is not loaded",
					Suggestion: "Load the nvidia_peermem kernel module",
					Commands:   []string{"sudo modprobe nvidia_peermem"},
				},
			},
			"nvlink_speed_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					FaultCode:  "HPCGPU-0009-0001",
					Issue:      "NVLink speed or count check failed",
					Suggestion: "Check NVLink health and verify GPU interconnect topology",
					Commands:   []string{"nvidia-smi nvlink -s", "nvidia-smi topo -m"},
				},
			},
		},
		SummaryTemplates: map[string]string{
			"no_issues":  "All tests passed!",
			"has_issues": "Found {total_issues} issue(s)",
		},
	}
}

func createTempConfigFile(t *testing.T, config RecommendationConfig) string {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "recommendations.json")

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	return configFile
}

func assertValidJSON(t *testing.T, jsonStr string) RecommendationReport {
	var report RecommendationReport
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	return report
}

func assertContainsStrings(t *testing.T, text string, expected []string) {
	for _, str := range expected {
		if !strings.Contains(text, str) {
			t.Errorf("Output missing expected string: %s", str)
		}
	}
}

// Core functionality tests

func TestRecommendationConfig_LoadAndGetRecommendation(t *testing.T) {
	config := createTestConfig()
	tempFile := createTempConfigFile(t, config)

	// Create current directory config for priority testing
	currentDirConfig := "./recommendations.json"
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(currentDirConfig, configData, 0644)
	defer os.Remove(currentDirConfig)

	loadedConfig, err := LoadRecommendationConfig()
	if err != nil {
		t.Fatalf("LoadRecommendationConfig failed: %v", err)
	}

	if len(loadedConfig.Recommendations) != 4 {
		t.Errorf("Expected 4 recommendations, got %d", len(loadedConfig.Recommendations))
	}

	_ = tempFile // Keep for cleanup
}

func TestRecommendationConfig_GetRecommendation(t *testing.T) {
	config := createTestConfig()

	tests := []struct {
		name         string
		testName     string
		status       string
		testResult   TestResult
		expectedType string
		expectNil    bool
	}{
		{
			name:     "GPU Count FAIL",
			testName: "gpu_count_check",
			status:   "FAIL",
			testResult: TestResult{
				Status:   "FAIL",
				GPUCount: 4,
			},
			expectedType: "critical",
		},
		{
			name:     "GPU Count PASS",
			testName: "gpu_count_check",
			status:   "PASS",
			testResult: TestResult{
				Status:   "PASS",
				GPUCount: 8,
			},
			expectedType: "info",
		},
		{
			name:     "GPU Mode with Indexes",
			testName: "gpu_mode_check",
			status:   "FAIL",
			testResult: TestResult{
				Status:            "FAIL",
				EnabledGPUIndexes: []string{"0", "1"},
			},
			expectedType: "critical",
		},
		{
			name:     "NVLink Speed Check FAIL",
			testName: "nvlink_speed_check",
			status:   "FAIL",
			testResult: TestResult{
				Status: "FAIL",
			},
			expectedType: "critical",
		},
		{
			name:      "Unknown Test",
			testName:  "unknown_test",
			status:    "FAIL",
			expectNil: true,
		},
		{
			name:      "Unknown Status",
			testName:  "gpu_count_check",
			status:    "UNKNOWN",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := config.GetRecommendation(tt.testName, tt.status, tt.testResult)

			if tt.expectNil {
				if rec != nil {
					t.Error("Expected nil recommendation")
				}
				return
			}

			if rec == nil {
				t.Fatal("Expected recommendation but got nil")
			}

			if rec.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, rec.Type)
			}

			if rec.TestName != tt.testName {
				t.Errorf("Expected test name %s, got %s", tt.testName, rec.TestName)
			}

			// Check fault code for nvlink_speed_check
			if tt.testName == "nvlink_speed_check" && rec.FaultCode != "HPCGPU-0009-0001" {
				t.Errorf("Expected fault code HPCGPU-0009-0001, got %s", rec.FaultCode)
			}
		})
	}
}

func TestVariableSubstitution(t *testing.T) {
	testResult := TestResult{
		GPUCount:          8,
		NumRDMANics:       2,
		EnabledGPUIndexes: []string{"0", "1", "3"},
		MaxUncorrectable:  5,
		MaxCorrectable:    100,
	}

	tests := []struct {
		template string
		expected string
	}{
		{"Found {gpu_count} GPUs", "Found 8 GPUs"},
		{"RDMA: {num_rdma_nics}", "RDMA: 2"},
		{"MIG on: {enabled_gpu_indexes}", "MIG on: 0,1,3"},
		{"Errors: {max_uncorrectable}/{max_correctable}", "Errors: 5/100"},
		{"No variables", "No variables"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result := applyVariableSubstitution(tt.template, testResult)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCommandSubstitutions(t *testing.T) {
	testResult := TestResult{
		EnabledGPUIndexes: []string{"0", "2"},
	}

	commands := []string{
		"nvidia-smi",
		"sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0",
	}

	expectedCommands := []string{
		"nvidia-smi",
		"sudo nvidia-smi -i 0,2 -mig 0",
	}

	result := applyCommandSubstitutions(commands, testResult)

	if len(result) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(result))
	}

	for i, expected := range expectedCommands {
		if i < len(result) && result[i] != expected {
			t.Errorf("Command %d: expected %s, got %s", i, expected, result[i])
		}
	}
}

func TestSummaryGeneration(t *testing.T) {
	config := &RecommendationConfig{
		SummaryTemplates: map[string]string{
			"no_issues":  "🎉 All tests passed!",
			"has_issues": "Found {total_issues} issue(s): {critical_count} critical, {warning_count} warning",
		},
	}

	tests := []struct {
		name          string
		totalIssues   int
		criticalCount int
		warningCount  int
		expected      string
	}{
		{
			name:     "No Issues",
			expected: "🎉 All tests passed!",
		},
		{
			name:          "With Issues",
			totalIssues:   3,
			criticalCount: 2,
			warningCount:  1,
			expected:      "Found 3 issue(s): 2 critical, 1 warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := config.GetSummary(tt.totalIssues, tt.criticalCount, tt.warningCount)
			if summary != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, summary)
			}
		})
	}
}

// Output format tests

func TestFormatRecommendations(t *testing.T) {
	report := createTestReport()

	tests := []struct {
		name    string
		format  string
		checker func(*testing.T, string)
	}{
		{
			name:   "JSON Format",
			format: "json",
			checker: func(t *testing.T, output string) {
				parsed := assertValidJSON(t, output)
				if parsed.TotalIssues != 2 {
					t.Errorf("Expected 2 total issues, got %d", parsed.TotalIssues)
				}
				if len(parsed.Recommendations) != 2 {
					t.Errorf("Expected 2 recommendations, got %d", len(parsed.Recommendations))
				}
			},
		},
		{
			name:   "Table Format",
			format: "table",
			checker: func(t *testing.T, output string) {
				expectedStrings := []string{
					"HPC DIAGNOSTIC RECOMMENDATIONS",
					"SUMMARY",
					"Total Issues: 2",
					"Critical: 2",
					"gpu_count_check",
					"gpu_mode_check",
				}
				assertContainsStrings(t, output, expectedStrings)
			},
		},
		{
			name:   "Friendly Format",
			format: "friendly",
			checker: func(t *testing.T, output string) {
				expectedStrings := []string{
					"🔍 HPC DIAGNOSTIC RECOMMENDATIONS",
					"📊 SUMMARY:",
					"Total Issues: 2",
					"Critical: 2",
					"🚨 1. CRITICAL [gpu_count_check]",
					"🚨 2. CRITICAL [gpu_mode_check]",
				}
				assertContainsStrings(t, output, expectedStrings)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			switch tt.format {
			case "json":
				output, err = formatRecommendationsJSON(report)
			case "table":
				output, err = formatRecommendationsTable(report)
			case "friendly":
				output, err = formatRecommendationsFriendly(report)
			}

			if err != nil {
				t.Fatalf("Format function failed: %v", err)
			}

			tt.checker(t, output)
		})
	}
}

func TestFormatRecommendations_NoIssues(t *testing.T) {
	report := RecommendationReport{
		Summary:         "All tests passed!",
		TotalIssues:     0,
		Recommendations: []Recommendation{},
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	output, err := formatRecommendationsFriendly(report)
	if err != nil {
		t.Fatalf("formatRecommendationsFriendly failed: %v", err)
	}

	expectedStrings := []string{
		"Total Issues: 0",
		"✅ No recommendations needed. System appears healthy!",
	}
	assertContainsStrings(t, output, expectedStrings)

	// Should not contain detailed recommendations section
	if strings.Contains(output, "📋 DETAILED RECOMMENDATIONS") {
		t.Error("Should not contain detailed recommendations section when no issues")
	}
}

func TestOutputRecommendations_InvalidFormat(t *testing.T) {
	report := createTestReport()

	err := outputRecommendations(report, "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("Expected 'unsupported output format' error, got: %v", err)
	}
}

// Test result structure and serialization

func TestTestResultSerialization(t *testing.T) {
	testResults := []TestResult{
		{
			Status:   "FAIL",
			GPUCount: 4,
		},
		{
			Status:            "FAIL",
			EnabledGPUIndexes: []string{"0", "1"},
		},
		{
			Status:       "FAIL",
			ModuleLoaded: false,
		},
		{
			Status:           "FAIL",
			MaxUncorrectable: 10,
			MaxCorrectable:   200,
		},
		{
			Status:  "FAIL",
			NVLinks: map[string]interface{}{"speed": 26, "count": 18},
		},
	}

	for i, original := range testResults {
		t.Run(fmt.Sprintf("TestResult_%d", i), func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Unmarshal from JSON
			var unmarshaled TestResult
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Verify key fields
			if unmarshaled.Status != original.Status {
				t.Errorf("Status mismatch: expected %s, got %s", original.Status, unmarshaled.Status)
			}

			if original.GPUCount > 0 && unmarshaled.GPUCount != original.GPUCount {
				t.Errorf("GPUCount mismatch: expected %d, got %d", original.GPUCount, unmarshaled.GPUCount)
			}

			if len(original.EnabledGPUIndexes) > 0 {
				if len(unmarshaled.EnabledGPUIndexes) != len(original.EnabledGPUIndexes) {
					t.Errorf("EnabledGPUIndexes length mismatch")
				}
			}
		})
	}
}

// Generation and integration tests

func TestGenerateRecommendations_WithConfig(t *testing.T) {
	// Create temp config
	config := createTestConfig()
	createTempConfigFile(t, config)

	// Create current directory config for testing
	currentDirConfig := "./recommendations.json"
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(currentDirConfig, configData, 0644)
	defer os.Remove(currentDirConfig)

	results := HostResults{
		GPUCountCheck: []TestResult{
			{Status: "FAIL", GPUCount: 4},
		},
		GPUModeCheck: []TestResult{
			{Status: "FAIL", EnabledGPUIndexes: []string{"0", "1"}},
		},
		PeerMemModuleCheck: []TestResult{
			{Status: "FAIL", ModuleLoaded: false},
		},
		NVLinkSpeedCheck: []TestResult{
			{Status: "FAIL"},
		},
	}

	report := generateRecommendations(results)

	if report.TotalIssues < 4 {
		t.Errorf("Expected at least 4 issues, got %d", report.TotalIssues)
	}

	if report.CriticalIssues < 4 {
		t.Errorf("Expected at least 4 critical issues, got %d", report.CriticalIssues)
	}
}

func TestGenerateRecommendations_FallbackMode(t *testing.T) {
	// Ensure no config file exists
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tempDir)
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", "")

	results := HostResults{
		GPUCountCheck: []TestResult{
			{Status: "FAIL", GPUCount: 2},
		},
		PeerMemModuleCheck: []TestResult{
			{Status: "FAIL", ModuleLoaded: false},
		},
		NVLinkSpeedCheck: []TestResult{
			{Status: "FAIL"},
		},
		Eth0PresenceCheck: []TestResult{
			{Status: "FAIL", Eth0Present: false},
		},
	}

	report := generateRecommendations(results)

	if len(report.Recommendations) < 4 {
		t.Errorf("Expected at least 4 fallback recommendations, got %d", len(report.Recommendations))
	}

	if !strings.Contains(report.Summary, "fallback mode") {
		t.Error("Expected fallback mode indicator in summary")
	}
}

// Specific test type validations

func TestSpecificTestTypes(t *testing.T) {
	tests := []struct {
		name       string
		hostResult HostResults
		expectType string
		expectTest string
	}{
		{
			name: "GPU Mode Check",
			hostResult: HostResults{
				GPUModeCheck: []TestResult{
					{Status: "FAIL", EnabledGPUIndexes: []string{"0", "1"}},
				},
			},
			expectType: "critical",
			expectTest: "gpu_mode_check",
		},
		{
			name: "PeerMem Module Check",
			hostResult: HostResults{
				PeerMemModuleCheck: []TestResult{
					{Status: "FAIL", ModuleLoaded: false},
				},
			},
			expectType: "critical",
			expectTest: "peermem_module_check",
		},
		{
			name: "SRAM Error Check",
			hostResult: HostResults{
				SRAMErrorCheck: []TestResult{
					{Status: "FAIL", MaxUncorrectable: 15},
				},
			},
			expectType: "critical",
			expectTest: "sram_error_check",
		},
		{
			name: "NVLink Speed Check",
			hostResult: HostResults{
				NVLinkSpeedCheck: []TestResult{
					{Status: "FAIL"},
				},
			},
			expectType: "critical",
			expectTest: "nvlink_speed_check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := generateRecommendations(tt.hostResult)

			if len(report.Recommendations) == 0 {
				t.Fatal("Expected at least one recommendation")
			}

			// Find the specific recommendation
			var found *Recommendation
			for i := range report.Recommendations {
				if report.Recommendations[i].TestName == tt.expectTest {
					found = &report.Recommendations[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("Expected %s recommendation not found", tt.expectTest)
			}

			if found.Type != tt.expectType {
				t.Errorf("Expected type %s, got %s", tt.expectType, found.Type)
			}

			// Check fault code for nvlink_speed_check
			if tt.expectTest == "nvlink_speed_check" && found.FaultCode != "HPCGPU-0009-0001" {
				t.Errorf("Expected fault code HPCGPU-0009-0001, got %s", found.FaultCode)
			}
		})
	}
}
