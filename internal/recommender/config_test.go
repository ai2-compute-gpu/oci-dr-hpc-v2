package recommender

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test helper functions

func createMinimalTestConfig() RecommendationConfig {
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

func createTestConfigFile(t *testing.T, config RecommendationConfig) string {
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

func createCurrentDirConfig(t *testing.T, config RecommendationConfig) {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile("./recommendations.json", configData, 0644); err != nil {
		t.Fatalf("Failed to write current dir config: %v", err)
	}

	t.Cleanup(func() {
		os.Remove("./recommendations.json")
	})
}

func createTestResult(status string, overrides map[string]interface{}) TestResult {
	result := TestResult{
		Status:       status,
		TimestampUTC: time.Now().UTC().Format(time.RFC3339),
	}

	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "gpu_count":
			result.GPUCount = value.(int)
		case "enabled_gpu_indexes":
			result.EnabledGPUIndexes = value.([]string)
		case "module_loaded":
			result.ModuleLoaded = value.(bool)
		case "max_uncorrectable":
			result.MaxUncorrectable = value.(int)
		case "max_correctable":
			result.MaxCorrectable = value.(int)
		}
	}

	return result
}

// Core functionality tests

func TestLoadRecommendationConfig(t *testing.T) {
	config := createMinimalTestConfig()
	createTestConfigFile(t, config)
	createCurrentDirConfig(t, config)

	loadedConfig, err := LoadRecommendationConfig()
	if err != nil {
		t.Fatalf("LoadRecommendationConfig failed: %v", err)
	}

	// Validate structure
	if len(loadedConfig.Recommendations) != 4 {
		t.Errorf("Expected 4 recommendations, got %d", len(loadedConfig.Recommendations))
	}

	// Validate specific test exists
	if _, exists := loadedConfig.Recommendations["gpu_count_check"]; !exists {
		t.Error("gpu_count_check recommendation not found")
	}

	// Validate templates
	gpuConfig := loadedConfig.Recommendations["gpu_count_check"]
	if gpuConfig.Fail == nil || gpuConfig.Fail.Type != "critical" {
		t.Error("gpu_count_check fail template missing or incorrect")
	}
}

func TestGetRecommendation(t *testing.T) {
	config := createMinimalTestConfig()

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
			testResult: createTestResult("FAIL", map[string]interface{}{
				"gpu_count": 4,
			}),
			expectedType: "critical",
		},
		{
			name:     "GPU Count PASS",
			testName: "gpu_count_check",
			status:   "PASS",
			testResult: createTestResult("PASS", map[string]interface{}{
				"gpu_count": 8,
			}),
			expectedType: "info",
		},
		{
			name:     "GPU Mode with Indexes",
			testName: "gpu_mode_check",
			status:   "FAIL",
			testResult: createTestResult("FAIL", map[string]interface{}{
				"enabled_gpu_indexes": []string{"0", "1"},
			}),
			expectedType: "critical",
		},
		{
			name:     "PeerMem Module",
			testName: "peermem_module_check",
			status:   "FAIL",
			testResult: createTestResult("FAIL", map[string]interface{}{
				"module_loaded": false,
			}),
			expectedType: "critical",
		},
		{
			name:         "NVLink Speed Check",
			testName:     "nvlink_speed_check",
			status:       "FAIL",
			testResult:   createTestResult("FAIL", map[string]interface{}{}),
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

func TestGetSummary(t *testing.T) {
	config := &RecommendationConfig{
		SummaryTemplates: map[string]string{
			"no_issues":  "🎉 All tests passed!",
			"has_issues": "⚠️ Found {total_issues} issue(s): {critical_count} critical, {warning_count} warning",
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
			expected:      "⚠️ Found 3 issue(s): 2 critical, 1 warning",
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

// Variable substitution tests

func TestApplyVariableSubstitution(t *testing.T) {
	testResult := TestResult{
		GPUCount:          8,
		NumRDMANics:       2,
		EnabledGPUIndexes: []string{"0", "1", "3"},
		MaxUncorrectable:  5,
		MaxCorrectable:    100,
		FailedInterfaces:  "rdma2,rdma3",
	}

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "Found {gpu_count} GPUs",
			expected: "Found 8 GPUs",
		},
		{
			template: "Detected {num_rdma_nics} RDMA NICs",
			expected: "Detected 2 RDMA NICs",
		},
		{
			template: "MIG enabled on GPUs: {enabled_gpu_indexes}",
			expected: "MIG enabled on GPUs: 0,1,3",
		},
		{
			template: "Errors: uncorr={max_uncorrectable}, corr={max_correctable}",
			expected: "Errors: uncorr=5, corr=100",
		},
		{
			template: "Failed interfaces: {failed_interfaces}",
			expected: "Failed interfaces: rdma2,rdma3",
		},
		{
			template: "No variables here",
			expected: "No variables here",
		},
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

func TestApplyCommandSubstitutions(t *testing.T) {
	testResult := TestResult{
		GPUCount:          4,
		EnabledGPUIndexes: []string{"0", "2"},
	}

	commands := []string{
		"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader",
		"sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0",
		"nvidia-smi -q -i {enabled_gpu_indexes}",
	}

	expectedCommands := []string{
		"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader",
		"sudo nvidia-smi -i 0,2 -mig 0",
		"nvidia-smi -q -i 0,2",
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

// Integration tests

func TestConfigBasedRecommendations(t *testing.T) {
	config := createMinimalTestConfig()
	createCurrentDirConfig(t, config)

	results := HostResults{
		GPUCountCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"gpu_count": 4}),
		},
		GPUModeCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"enabled_gpu_indexes": []string{"0", "1"}}),
		},
		PeerMemModuleCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"module_loaded": false}),
		},
		NVLinkSpeedCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{}),
		},
	}

	report := generateRecommendations(results)

	// Validate report structure
	if len(report.Recommendations) != 4 {
		t.Errorf("Expected 4 recommendations, got %d", len(report.Recommendations))
	}

	if report.CriticalIssues != 4 {
		t.Errorf("Expected 4 critical issues, got %d", report.CriticalIssues)
	}

	// Validate specific recommendations
	recommendationTests := []struct {
		testName     string
		expectedType string
		issueKeyword string
	}{
		{"gpu_count_check", "critical", "4"},
		{"gpu_mode_check", "critical", "0,1"},
		{"peermem_module_check", "critical", "not loaded"},
		{"nvlink_speed_check", "critical", "failed"},
	}

	for _, tt := range recommendationTests {
		t.Run(tt.testName, func(t *testing.T) {
			var found *Recommendation
			for i := range report.Recommendations {
				if report.Recommendations[i].TestName == tt.testName {
					found = &report.Recommendations[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("%s recommendation not found", tt.testName)
			}

			if found.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, found.Type)
			}

			if !strings.Contains(found.Issue, tt.issueKeyword) {
				t.Errorf("Expected issue to contain %s, got: %s", tt.issueKeyword, found.Issue)
			}

			// Check fault code for nvlink_speed_check
			if tt.testName == "nvlink_speed_check" && found.FaultCode != "HPCGPU-0009-0001" {
				t.Errorf("Expected fault code HPCGPU-0009-0001, got %s", found.FaultCode)
			}
		})
	}
}

func TestFallbackRecommendations(t *testing.T) {
	// Setup environment with no config files
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	originalHome := os.Getenv("HOME")

	defer func() {
		os.Chdir(originalDir)
		os.Setenv("HOME", originalHome)
	}()

	os.Chdir(tempDir)
	os.Setenv("HOME", tempDir)

	results := HostResults{
		GPUCountCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"gpu_count": 2}),
		},
		GPUModeCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"enabled_gpu_indexes": []string{"0"}}),
		},
		PeerMemModuleCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{"module_loaded": false}),
		},
		NVLinkSpeedCheck: []TestResult{
			createTestResult("FAIL", map[string]interface{}{}),
		},
	}

	report := generateRecommendations(results)

	// Validate fallback behavior
	if len(report.Recommendations) < 4 {
		t.Errorf("Expected at least 4 fallback recommendations, got %d", len(report.Recommendations))
	}

	if !strings.Contains(report.Summary, "fallback mode") {
		t.Error("Expected fallback mode indicator in summary")
	}

	// Validate specific fallback recommendations exist
	testNames := []string{"gpu_count_check", "gpu_mode_check", "peermem_module_check", "nvlink_speed_check"}
	for _, testName := range testNames {
		found := false
		for _, rec := range report.Recommendations {
			if rec.TestName == testName {
				found = true
				if rec.Type != "critical" {
					t.Errorf("Expected %s fallback type 'critical', got %s", testName, rec.Type)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected %s fallback recommendation not found", testName)
		}
	}
}

// Edge cases and error handling

func TestEdgeCases(t *testing.T) {
	t.Run("Empty GPU Indexes", func(t *testing.T) {
		testResult := TestResult{EnabledGPUIndexes: []string{}}
		result := applyVariableSubstitution("GPUs: {enabled_gpu_indexes}", testResult)
		expected := "GPUs: "
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("Zero Values", func(t *testing.T) {
		testResult := TestResult{
			GPUCount:         0,
			MaxUncorrectable: 0,
			MaxCorrectable:   0,
		}
		result := applyVariableSubstitution("GPU:{gpu_count} Err:{max_uncorrectable}", testResult)
		expected := "GPU:0 Err:0"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("Missing Template", func(t *testing.T) {
		config := &RecommendationConfig{
			Recommendations: map[string]TestRecommendations{
				"test_check": {
					// Only FAIL template, no PASS template
					Fail: &RecommendationTemplate{
						Type: "critical",
					},
				},
			},
		}

		// Should return nil for missing PASS template
		rec := config.GetRecommendation("test_check", "PASS", TestResult{})
		if rec != nil {
			t.Error("Expected nil for missing PASS template")
		}
	})

	t.Run("Config Without Summary Templates", func(t *testing.T) {
		config := &RecommendationConfig{
			SummaryTemplates: map[string]string{}, // Empty templates
		}

		summary := config.GetSummary(3, 2, 1)
		expected := "Found 3 issue(s) requiring attention: 2 critical, 1 warning"
		if summary != expected {
			t.Errorf("Expected default summary, got %s", summary)
		}
	})
}

// Performance and memory tests

func BenchmarkApplyVariableSubstitution(b *testing.B) {
	testResult := TestResult{
		GPUCount:          8,
		EnabledGPUIndexes: []string{"0", "1", "2", "3"},
		MaxUncorrectable:  5,
		MaxCorrectable:    100,
	}
	template := "GPU count: {gpu_count}, MIG on: {enabled_gpu_indexes}, Errors: {max_uncorrectable}/{max_correctable}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyVariableSubstitution(template, testResult)
	}
}

func BenchmarkGetRecommendation(b *testing.B) {
	config := createMinimalTestConfig()
	testResult := createTestResult("FAIL", map[string]interface{}{
		"gpu_count": 8,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetRecommendation("gpu_count_check", "FAIL", testResult)
	}
}
