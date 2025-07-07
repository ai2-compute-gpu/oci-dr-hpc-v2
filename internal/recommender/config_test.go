package recommender

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRecommendationConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "recommendations.json")
	
	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	testConfig := RecommendationConfig{
		Recommendations: map[string]TestRecommendations{
			"gpu_count_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU count mismatch detected (found: {gpu_count})",
					Suggestion: "Verify GPU hardware installation",
					Commands:   []string{"nvidia-smi"},
					References: []string{"https://docs.nvidia.com/"},
				},
				Pass: &RecommendationTemplate{
					Type:       "info",
					Issue:      "GPU count check passed ({gpu_count} GPUs detected)",
					Suggestion: "GPU hardware is properly configured",
					Commands:   []string{"nvidia-smi -q"},
				},
			},
		},
		SummaryTemplates: map[string]string{
			"no_issues":  "All tests passed!",
			"has_issues": "Found {total_issues} issue(s)",
		},
	}

	// Write test config to file
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create current directory config so it takes priority over user config
	currentDirConfig := "./recommendations.json"
	if err := os.WriteFile(currentDirConfig, configData, 0644); err != nil {
		t.Fatalf("Failed to write current dir config file: %v", err)
	}
	defer os.Remove(currentDirConfig)

	// Test loading config
	config, err := LoadRecommendationConfig()
	if err != nil {
		t.Fatalf("LoadRecommendationConfig failed: %v", err)
	}

	// Verify loaded config
	if config == nil {
		t.Fatal("Config is nil")
	}

	if len(config.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(config.Recommendations))
	}

	gpuConfig, exists := config.Recommendations["gpu_count_check"]
	if !exists {
		t.Error("gpu_count_check recommendation not found")
	}

	if gpuConfig.Fail == nil {
		t.Error("gpu_count_check fail template is nil")
	}

	if gpuConfig.Fail.Type != "critical" {
		t.Errorf("Expected fail type 'critical', got '%s'", gpuConfig.Fail.Type)
	}
}

func TestGetRecommendation(t *testing.T) {
	config := &RecommendationConfig{
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
		},
	}

	// Test FAIL recommendation
	testResult := TestResult{
		Status:   "FAIL",
		GPUCount: 4,
		TimestampUTC: time.Now().UTC().Format(time.RFC3339),
	}

	rec := config.GetRecommendation("gpu_count_check", "FAIL", testResult)
	if rec == nil {
		t.Fatal("GetRecommendation returned nil for FAIL status")
	}

	if rec.Type != "critical" {
		t.Errorf("Expected type 'critical', got '%s'", rec.Type)
	}

	if rec.TestName != "gpu_count_check" {
		t.Errorf("Expected test name 'gpu_count_check', got '%s'", rec.TestName)
	}

	expectedIssue := "GPU count mismatch detected (found: 4)"
	if rec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, rec.Issue)
	}

	// Test PASS recommendation
	rec = config.GetRecommendation("gpu_count_check", "PASS", testResult)
	if rec == nil {
		t.Fatal("GetRecommendation returned nil for PASS status")
	}

	if rec.Type != "info" {
		t.Errorf("Expected type 'info', got '%s'", rec.Type)
	}

	expectedIssue = "GPU count check passed (4 GPUs detected)"
	if rec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, rec.Issue)
	}

	// Test unknown test
	rec = config.GetRecommendation("unknown_test", "FAIL", testResult)
	if rec != nil {
		t.Error("Expected nil for unknown test, got recommendation")
	}

	// Test unknown status
	rec = config.GetRecommendation("gpu_count_check", "UNKNOWN", testResult)
	if rec != nil {
		t.Error("Expected nil for unknown status, got recommendation")
	}
}

func TestGetSummary(t *testing.T) {
	config := &RecommendationConfig{
		SummaryTemplates: map[string]string{
			"no_issues":  "🎉 All tests passed!",
			"has_issues": "⚠️ Found {total_issues} issue(s): {critical_count} critical, {warning_count} warning",
		},
	}

	// Test no issues
	summary := config.GetSummary(0, 0, 0)
	expected := "🎉 All tests passed!"
	if summary != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary)
	}

	// Test with issues
	summary = config.GetSummary(3, 2, 1)
	expected = "⚠️ Found 3 issue(s): 2 critical, 1 warning"
	if summary != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary)
	}
}

func TestApplyVariableSubstitution(t *testing.T) {
	testResult := TestResult{
		GPUCount:     8,
		NumRDMANics:  2,
		TimestampUTC: time.Now().UTC().Format(time.RFC3339),
	}

	testCases := []struct {
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
			template: "GPU count: {gpu_count}, RDMA count: {num_rdma_nics}",
			expected: "GPU count: 8, RDMA count: 2",
		},
		{
			template: "No variables here",
			expected: "No variables here",
		},
	}

	for _, tc := range testCases {
		result := applyVariableSubstitution(tc.template, testResult)
		if result != tc.expected {
			t.Errorf("For template '%s', expected '%s', got '%s'", tc.template, tc.expected, result)
		}
	}
}

func TestConfigBasedRecommendations(t *testing.T) {
	// Create a temporary config file  
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "recommendations.json")
	
	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	testConfig := RecommendationConfig{
		Recommendations: map[string]TestRecommendations{
			"gpu_count_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU test failed with {gpu_count} GPUs",
					Suggestion: "Check GPU installation",
					Commands:   []string{"nvidia-smi"},
				},
			},
		},
		SummaryTemplates: map[string]string{
			"has_issues": "Test found {total_issues} problem(s)",
		},
	}

	// Write test config to file
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create current directory config so it takes priority over user config
	currentDirConfig := "./recommendations.json"
	if err := os.WriteFile(currentDirConfig, configData, 0644); err != nil {
		t.Fatalf("Failed to write current dir config file: %v", err)
	}
	defer os.Remove(currentDirConfig)

	// Test with failing GPU test
	results := HostResults{
		GPUCountCheck: []TestResult{
			{
				Status:   "FAIL",
				GPUCount: 4,
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	report := generateRecommendations(results)
	
	if len(report.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(report.Recommendations))
	}

	if report.CriticalIssues != 1 {
		t.Errorf("Expected 1 critical issue, got %d", report.CriticalIssues)
	}

	if report.TotalIssues != 1 {
		t.Errorf("Expected 1 total issue, got %d", report.TotalIssues)
	}

	rec := report.Recommendations[0]
	expectedIssue := "GPU test failed with 4 GPUs"
	if rec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, rec.Issue)
	}

	expectedSummary := "Test found 1 problem(s)"
	if report.Summary != expectedSummary {
		t.Errorf("Expected summary '%s', got '%s'", expectedSummary, report.Summary)
	}
}

func TestFallbackRecommendations(t *testing.T) {
	// Test fallback when config loading fails (no config file)
	tempDir := t.TempDir()
	
	// Save original working directory and HOME
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Change to temp directory first 
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Create an invalid JSON file to trigger fallback
	invalidConfig := `{"invalid": json}`  // Missing quotes - invalid JSON
	configPath := "recommendations.json" 
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	// Test with failing tests
	results := HostResults{
		GPUCountCheck: []TestResult{
			{
				Status:   "FAIL",
				GPUCount: 2,
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
		PCIeErrorCheck: []TestResult{
			{
				Status:       "FAIL",
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	report := generateRecommendations(results)
	
	// Should generate fallback recommendations
	if len(report.Recommendations) != 2 {
		t.Errorf("Expected 2 fallback recommendations, got %d", len(report.Recommendations))
	}

	if report.CriticalIssues != 2 {
		t.Errorf("Expected 2 critical issues in fallback, got %d", report.CriticalIssues)
	}

	// Should indicate fallback mode in summary
	if !strings.Contains(report.Summary, "fallback mode") {
		t.Errorf("Expected fallback mode indicator in summary, got: %s", report.Summary)
	}
}