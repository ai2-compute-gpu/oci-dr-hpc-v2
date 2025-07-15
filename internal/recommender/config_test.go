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
			"gpu_mode_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU MIG mode configuration violation detected on GPUs: {enabled_gpu_indexes}",
					Suggestion: "Disable MIG mode on affected GPUs",
					Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader", "sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0"},
					References: []string{"https://docs.nvidia.com/datacenter/tesla/mig-user-guide/"},
				},
				Pass: &RecommendationTemplate{
					Type:       "info",
					Issue:      "GPU MIG mode check passed - all GPUs have acceptable mode configuration",
					Suggestion: "GPU MIG mode configuration is compliant",
					Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader"},
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
	config, err := LoadRecommendationConfig([]string{currentDirConfig})
	if err != nil {
		t.Fatalf("LoadRecommendationConfig failed: %v", err)
	}

	// Verify loaded config
	if config == nil {
		t.Fatal("Config is nil")
	}

	if len(config.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d", len(config.Recommendations))
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

	// Test GPU mode check config
	gpuModeConfig, exists := config.Recommendations["gpu_mode_check"]
	if !exists {
		t.Error("gpu_mode_check recommendation not found")
	}

	if gpuModeConfig.Fail == nil {
		t.Error("gpu_mode_check fail template is nil")
	}

	if gpuModeConfig.Fail.Type != "critical" {
		t.Errorf("Expected gpu_mode_check fail type 'critical', got '%s'", gpuModeConfig.Fail.Type)
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
			"gpu_mode_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU MIG mode violation on GPUs: {enabled_gpu_indexes}",
					Suggestion: "Disable MIG mode on affected GPUs",
					Commands:   []string{"sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0"},
				},
				Pass: &RecommendationTemplate{
					Type:       "info",
					Issue:      "GPU MIG mode check passed",
					Suggestion: "GPU MIG configuration is compliant",
				},
			},
		},
	}

	// Test GPU count FAIL recommendation
	testResult := TestResult{
		Status:       "FAIL",
		GPUCount:     4,
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

	// Test GPU mode FAIL recommendation
	gpuModeResult := TestResult{
		Status:            "FAIL",
		EnabledGPUIndexes: []string{"0", "1"},
		TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
	}

	rec = config.GetRecommendation("gpu_mode_check", "FAIL", gpuModeResult)
	if rec == nil {
		t.Fatal("GetRecommendation returned nil for gpu_mode_check FAIL status")
	}

	if rec.Type != "critical" {
		t.Errorf("Expected type 'critical', got '%s'", rec.Type)
	}

	expectedIssue = "GPU MIG mode violation on GPUs: 0,1"
	if rec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, rec.Issue)
	}

	expectedCommand := "sudo nvidia-smi -i 0,1 -mig 0"
	if len(rec.Commands) == 0 || rec.Commands[0] != expectedCommand {
		t.Errorf("Expected command '%s', got '%v'", expectedCommand, rec.Commands)
	}

	// Test GPU mode PASS recommendation
	gpuModeResultPass := TestResult{
		Status:       "PASS",
		TimestampUTC: time.Now().UTC().Format(time.RFC3339),
	}

	rec = config.GetRecommendation("gpu_mode_check", "PASS", gpuModeResultPass)
	if rec == nil {
		t.Fatal("GetRecommendation returned nil for gpu_mode_check PASS status")
	}

	if rec.Type != "info" {
		t.Errorf("Expected type 'info', got '%s'", rec.Type)
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
		GPUCount:          8,
		NumRDMANics:       2,
		EnabledGPUIndexes: []string{"0", "1", "3"},
		TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
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
			template: "MIG enabled on GPUs: {enabled_gpu_indexes}",
			expected: "MIG enabled on GPUs: 0,1,3",
		},
		{
			template: "No variables here",
			expected: "No variables here",
		},
		{
			template: "Command: sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0",
			expected: "Command: sudo nvidia-smi -i 0,1,3 -mig 0",
		},
	}

	for _, tc := range testCases {
		result := applyVariableSubstitution(tc.template, testResult)
		if result != tc.expected {
			t.Errorf("For template '%s', expected '%s', got '%s'", tc.template, tc.expected, result)
		}
	}
}

func TestApplyCommandSubstitutions(t *testing.T) {
	testResult := TestResult{
		GPUCount:          4,
		EnabledGPUIndexes: []string{"0", "2"},
		TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
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
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected, result[i])
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
			"gpu_mode_check": {
				Fail: &RecommendationTemplate{
					Type:       "critical",
					Issue:      "GPU MIG mode violation on {enabled_gpu_indexes}",
					Suggestion: "Disable MIG mode",
					Commands:   []string{"sudo nvidia-smi -i {enabled_gpu_indexes} -mig 0"},
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

	// Test with failing GPU tests
	results := HostResults{
		GPUCountCheck: []TestResult{
			{
				Status:       "FAIL",
				GPUCount:     4,
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
		GPUModeCheck: []TestResult{
			{
				Status:            "FAIL",
				EnabledGPUIndexes: []string{"0", "1"},
				TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	report := generateRecommendations(results)

	if len(report.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d", len(report.Recommendations))
	}

	if report.CriticalIssues != 2 {
		t.Errorf("Expected 2 critical issues, got %d", report.CriticalIssues)
	}

	if report.TotalIssues != 2 {
		t.Errorf("Expected 2 total issues, got %d", report.TotalIssues)
	}

	// Find GPU count check recommendation
	var gpuCountRec *Recommendation
	var gpuModeRec *Recommendation
	for i := range report.Recommendations {
		if report.Recommendations[i].TestName == "gpu_count_check" {
			gpuCountRec = &report.Recommendations[i]
		}
		if report.Recommendations[i].TestName == "gpu_mode_check" {
			gpuModeRec = &report.Recommendations[i]
		}
	}

	if gpuCountRec == nil {
		t.Fatal("gpu_count_check recommendation not found")
	}

	expectedIssue := "GPU test failed with 4 GPUs"
	if gpuCountRec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, gpuCountRec.Issue)
	}

	if gpuModeRec == nil {
		t.Fatal("gpu_mode_check recommendation not found")
	}

	expectedModeIssue := "GPU MIG mode violation on 0,1"
	if gpuModeRec.Issue != expectedModeIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedModeIssue, gpuModeRec.Issue)
	}

	expectedSummary := "Test found 2 problem(s)"
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
	invalidConfig := `{"invalid": json}` // Missing quotes - invalid JSON
	configPath := "recommendations.json"
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	// Test with failing tests
	results := HostResults{
		GPUCountCheck: []TestResult{
			{
				Status:       "FAIL",
				GPUCount:     2,
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
		GPUModeCheck: []TestResult{
			{
				Status:            "FAIL",
				EnabledGPUIndexes: []string{"0"},
				TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
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
	if len(report.Recommendations) != 3 {
		t.Errorf("Expected 3 fallback recommendations, got %d", len(report.Recommendations))
	}

	if report.CriticalIssues != 3 {
		t.Errorf("Expected 3 critical issues in fallback, got %d", report.CriticalIssues)
	}

	// Should indicate fallback mode in summary
	if !strings.Contains(report.Summary, "fallback mode") {
		t.Errorf("Expected fallback mode indicator in summary, got: %s", report.Summary)
	}

	// Verify GPU mode check fallback recommendation exists
	var gpuModeRec *Recommendation
	for i := range report.Recommendations {
		if report.Recommendations[i].TestName == "gpu_mode_check" {
			gpuModeRec = &report.Recommendations[i]
			break
		}
	}

	if gpuModeRec == nil {
		t.Fatal("gpu_mode_check fallback recommendation not found")
	}

	if gpuModeRec.Type != "critical" {
		t.Errorf("Expected gpu_mode_check fallback type 'critical', got '%s'", gpuModeRec.Type)
	}

	expectedIssue := "GPU MIG mode configuration violation detected on GPUs: [0]"
	if gpuModeRec.Issue != expectedIssue {
		t.Errorf("Expected issue '%s', got '%s'", expectedIssue, gpuModeRec.Issue)
	}
}
