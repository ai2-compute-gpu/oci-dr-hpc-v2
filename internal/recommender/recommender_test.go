package recommender

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestFormatRecommendationsJSON(t *testing.T) {
	// Create a test recommendation report
	testReport := RecommendationReport{
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
				Commands:   []string{"nvidia-smi", "lspci | grep -i nvidia"},
				References: []string{"https://docs.nvidia.com/datacenter/tesla/"},
			},
			{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      "GPU MIG mode configuration violation detected on GPUs: 0,1",
				Suggestion: "Disable MIG mode on affected GPUs",
				Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader", "sudo nvidia-smi -i 0,1 -mig 0"},
				References: []string{"https://docs.nvidia.com/datacenter/tesla/mig-user-guide/"},
			},
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Test JSON formatting
	result, err := formatRecommendationsJSON(testReport)
	if err != nil {
		t.Fatalf("formatRecommendationsJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var unmarshaled RecommendationReport
	if err := json.Unmarshal([]byte(result), &unmarshaled); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Verify key fields
	if unmarshaled.TotalIssues != 2 {
		t.Errorf("Expected TotalIssues=2, got %d", unmarshaled.TotalIssues)
	}
	if unmarshaled.CriticalIssues != 2 {
		t.Errorf("Expected CriticalIssues=2, got %d", unmarshaled.CriticalIssues)
	}
	if len(unmarshaled.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d", len(unmarshaled.Recommendations))
	}
	if unmarshaled.Recommendations[0].Type != "critical" {
		t.Errorf("Expected type 'critical', got '%s'", unmarshaled.Recommendations[0].Type)
	}
	if unmarshaled.Recommendations[1].TestName != "gpu_mode_check" {
		t.Errorf("Expected second test name 'gpu_mode_check', got '%s'", unmarshaled.Recommendations[1].TestName)
	}
}

func TestFormatRecommendationsTable(t *testing.T) {
	// Create a test recommendation report
	testReport := RecommendationReport{
		Summary:        "Test summary with mixed issues",
		TotalIssues:    3,
		CriticalIssues: 2,
		WarningIssues:  1,
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
			{
				Type:       "warning",
				TestName:   "rdma_nics_count",
				Issue:      "RDMA NIC count mismatch",
				Suggestion: "Check RDMA configuration",
				Commands:   []string{"ibstat"},
			},
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Test table formatting
	result, err := formatRecommendationsTable(testReport)
	if err != nil {
		t.Fatalf("formatRecommendationsTable failed: %v", err)
	}

	// Verify table contains expected content
	expectedStrings := []string{
		"HPC DIAGNOSTIC RECOMMENDATIONS",
		"SUMMARY",
		"Total Issues: 3",
		"Critical: 2",
		"Warning: 1",
		"RECOMMENDATIONS",
		"gpu_count_check",
		"gpu_mode_check",
		"rdma_nics_count",
		"Generated at:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Table output missing expected string: %s", expected)
		}
	}
}

func TestFormatRecommendationsFriendly(t *testing.T) {
	// Create a test recommendation report
	testReport := RecommendationReport{
		Summary:        "Found 1 issue requiring attention",
		TotalIssues:    1,
		CriticalIssues: 1,
		WarningIssues:  0,
		InfoIssues:     1,
		Recommendations: []Recommendation{
			{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      "GPU MIG mode configuration violation detected on GPUs: 0,1",
				Suggestion: "Disable MIG mode on affected GPUs or verify configuration",
				Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader", "sudo nvidia-smi -i 0,1 -mig 0"},
				References: []string{"https://docs.nvidia.com/datacenter/tesla/mig-user-guide/"},
			},
			{
				Type:       "info",
				TestName:   "gpu_count_check",
				Issue:      "GPU count check passed (8 GPUs detected)",
				Suggestion: "GPU hardware is properly detected and configured",
				Commands:   []string{"nvidia-smi -q", "nvidia-smi topo -m"},
			},
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Test friendly formatting
	result, err := formatRecommendationsFriendly(testReport)
	if err != nil {
		t.Fatalf("formatRecommendationsFriendly failed: %v", err)
	}

	// Verify friendly format contains expected content
	expectedStrings := []string{
		"🔍 HPC DIAGNOSTIC RECOMMENDATIONS",
		"📊 SUMMARY:",
		"Total Issues: 1",
		"Critical: 1",
		"Warning: 0",
		"Info: 1",
		"📋 DETAILED RECOMMENDATIONS",
		"🚨 1. CRITICAL [gpu_mode_check]",
		"ℹ️ 2. INFO [gpu_count_check]",
		"Commands to run:",
		"$ nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader",
		"$ sudo nvidia-smi -i 0,1 -mig 0",
		"References:",
		"- https://docs.nvidia.com/datacenter/tesla/mig-user-guide/",
		"Generated at:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Friendly output missing expected string: %s", expected)
		}
	}
}

func TestFormatRecommendationsFriendlyNoIssues(t *testing.T) {
	// Create a test recommendation report with no recommendations
	testReport := RecommendationReport{
		Summary:         "🎉 All diagnostic tests passed! Your HPC environment appears healthy.",
		TotalIssues:     0,
		CriticalIssues:  0,
		WarningIssues:   0,
		InfoIssues:      0,
		Recommendations: []Recommendation{}, // No recommendations
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	// Test friendly formatting
	result, err := formatRecommendationsFriendly(testReport)
	if err != nil {
		t.Fatalf("formatRecommendationsFriendly failed: %v", err)
	}

	// Verify it shows no recommendations message
	expectedStrings := []string{
		"🔍 HPC DIAGNOSTIC RECOMMENDATIONS",
		"📊 SUMMARY:",
		"Total Issues: 0",
		"✅ No recommendations needed. System appears healthy!",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Friendly output missing expected string: %s", expected)
		}
	}

	// Should not contain detailed recommendations section when no recommendations
	if strings.Contains(result, "📋 DETAILED RECOMMENDATIONS") {
		t.Errorf("Friendly output should not contain detailed recommendations section when no recommendations exist")
	}
}

func TestOutputRecommendations(t *testing.T) {
	testReport := RecommendationReport{
		Summary:        "Test summary",
		TotalIssues:    2,
		CriticalIssues: 2,
		WarningIssues:  0,
		InfoIssues:     0,
		Recommendations: []Recommendation{
			{
				Type:       "critical",
				TestName:   "gpu_count_check",
				Issue:      "GPU count test issue",
				Suggestion: "GPU count test suggestion",
			},
			{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      "GPU mode test issue",
				Suggestion: "GPU mode test suggestion",
			},
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Test unsupported format
	err := outputRecommendations(testReport, "unsupported")
	if err == nil {
		t.Errorf("Expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("Expected 'unsupported output format' error, got: %v", err)
	}

	// Test supported formats (these should not error)
	formats := []string{"json", "table", "friendly"}
	for _, format := range formats {
		err := outputRecommendations(testReport, format)
		if err != nil {
			t.Errorf("outputRecommendations failed for format '%s': %v", format, err)
		}
	}
}

func TestRecommendationReportStructure(t *testing.T) {
	// Test that RecommendationReport can be marshaled and unmarshaled correctly
	original := RecommendationReport{
		Summary:        "Test summary with 3 issues",
		TotalIssues:    3,
		CriticalIssues: 2,
		WarningIssues:  1,
		InfoIssues:     0,
		Recommendations: []Recommendation{
			{
				Type:       "critical",
				TestName:   "gpu_count_check",
				Issue:      "GPU count mismatch detected",
				Suggestion: "Verify GPU hardware installation",
				Commands:   []string{"nvidia-smi", "lspci | grep -i nvidia"},
				References: []string{"https://docs.nvidia.com/datacenter/tesla/"},
			},
			{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      "GPU MIG mode configuration violation detected on GPUs: 0,1",
				Suggestion: "Disable MIG mode on affected GPUs",
				Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader", "sudo nvidia-smi -i 0,1 -mig 0"},
				References: []string{"https://docs.nvidia.com/datacenter/tesla/mig-user-guide/"},
			},
			{
				Type:       "warning",
				TestName:   "rdma_nics_count",
				Issue:      "RDMA NIC count mismatch",
				Suggestion: "Check RDMA hardware configuration",
				Commands:   []string{"ibstat", "ibv_devices"},
				References: []string{"https://docs.oracle.com/rdma"},
			},
		},
		GeneratedAt: "2023-12-01T10:00:00Z",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal RecommendationReport: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled RecommendationReport
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal RecommendationReport: %v", err)
	}

	// Verify all fields
	if unmarshaled.Summary != original.Summary {
		t.Errorf("Summary mismatch: expected %s, got %s", original.Summary, unmarshaled.Summary)
	}
	if unmarshaled.TotalIssues != original.TotalIssues {
		t.Errorf("TotalIssues mismatch: expected %d, got %d", original.TotalIssues, unmarshaled.TotalIssues)
	}
	if len(unmarshaled.Recommendations) != len(original.Recommendations) {
		t.Errorf("Recommendations count mismatch: expected %d, got %d", len(original.Recommendations), len(unmarshaled.Recommendations))
	}

	// Verify first recommendation (GPU count check)
	if len(unmarshaled.Recommendations) > 0 {
		origRec := original.Recommendations[0]
		unmarshaledRec := unmarshaled.Recommendations[0]

		if unmarshaledRec.Type != origRec.Type {
			t.Errorf("Recommendation type mismatch: expected %s, got %s", origRec.Type, unmarshaledRec.Type)
		}
		if unmarshaledRec.TestName != origRec.TestName {
			t.Errorf("Test name mismatch: expected %s, got %s", origRec.TestName, unmarshaledRec.TestName)
		}
		if len(unmarshaledRec.Commands) != len(origRec.Commands) {
			t.Errorf("Commands count mismatch: expected %d, got %d", len(origRec.Commands), len(unmarshaledRec.Commands))
		}
		if len(unmarshaledRec.References) != len(origRec.References) {
			t.Errorf("References count mismatch: expected %d, got %d", len(origRec.References), len(unmarshaledRec.References))
		}
	}

	// Verify second recommendation (GPU mode check)
	if len(unmarshaled.Recommendations) > 1 {
		unmarshaledRec := unmarshaled.Recommendations[1]

		if unmarshaledRec.TestName != "gpu_mode_check" {
			t.Errorf("Expected second recommendation to be gpu_mode_check, got %s", unmarshaledRec.TestName)
		}
		if unmarshaledRec.Type != "critical" {
			t.Errorf("Expected gpu_mode_check type to be critical, got %s", unmarshaledRec.Type)
		}

		expectedIssue := "GPU MIG mode configuration violation detected on GPUs: 0,1"
		if unmarshaledRec.Issue != expectedIssue {
			t.Errorf("GPU mode check issue mismatch: expected %s, got %s", expectedIssue, unmarshaledRec.Issue)
		}

		if len(unmarshaledRec.Commands) != 2 {
			t.Errorf("Expected 2 commands for gpu_mode_check, got %d", len(unmarshaledRec.Commands))
		}
	}
}

func TestGenerateRecommendationsWithGPUModeCheck(t *testing.T) {
	// Test generateRecommendations function with GPU mode check results
	results := HostResults{
		GPUCountCheck: []TestResult{
			{
				Status:       "PASS",
				GPUCount:     8,
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
		GPUModeCheck: []TestResult{
			{
				Status:            "FAIL",
				Message:           "FAIL - MIG Mode enabled on GPUs 0,1",
				EnabledGPUIndexes: []string{"0", "1"},
				TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
			},
		},
		PCIeErrorCheck: []TestResult{
			{
				Status:       "PASS",
				TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	report := generateRecommendations(results)

	// Should have recommendations for GPU mode check failure and GPU count check pass
	// Note: PASS recommendations may or may not be included depending on config
	if len(report.Recommendations) < 1 {
		t.Errorf("Expected at least 1 recommendation, got %d", len(report.Recommendations))
	}

	// Should have 1 critical issue (GPU mode check failure)
	if report.CriticalIssues < 1 {
		t.Errorf("Expected at least 1 critical issue, got %d", report.CriticalIssues)
	}

	// Find GPU mode check recommendation
	var gpuModeRec *Recommendation
	for i := range report.Recommendations {
		if report.Recommendations[i].TestName == "gpu_mode_check" {
			gpuModeRec = &report.Recommendations[i]
			break
		}
	}

	// In fallback mode, should still have a GPU mode check recommendation
	if gpuModeRec == nil {
		t.Fatal("Expected gpu_mode_check recommendation not found")
	}

	if gpuModeRec.Type != "critical" {
		t.Errorf("Expected gpu_mode_check type 'critical', got '%s'", gpuModeRec.Type)
	}

	// Verify the issue contains information about enabled GPUs
	if !strings.Contains(gpuModeRec.Issue, "0") || !strings.Contains(gpuModeRec.Issue, "1") {
		t.Errorf("Expected issue to mention enabled GPUs 0 and 1, got: %s", gpuModeRec.Issue)
	}
}

func TestTestResultStructWithGPUModeFields(t *testing.T) {
	// Test that TestResult struct properly handles GPU mode check fields
	testResult := TestResult{
		Status:            "FAIL",
		GPUCount:          8,
		Message:           "FAIL - MIG Mode enabled on GPUs 0,1",
		EnabledGPUIndexes: []string{"0", "1"},
		NumRDMANics:       2,
		TimestampUTC:      time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(testResult)
	if err != nil {
		t.Fatalf("Failed to marshal TestResult: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled TestResult
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestResult: %v", err)
	}

	// Verify GPU mode specific fields
	if unmarshaled.Message != testResult.Message {
		t.Errorf("Message mismatch: expected %s, got %s", testResult.Message, unmarshaled.Message)
	}

	if len(unmarshaled.EnabledGPUIndexes) != len(testResult.EnabledGPUIndexes) {
		t.Errorf("EnabledGPUIndexes length mismatch: expected %d, got %d",
			len(testResult.EnabledGPUIndexes), len(unmarshaled.EnabledGPUIndexes))
	}

	for i, index := range testResult.EnabledGPUIndexes {
		if i < len(unmarshaled.EnabledGPUIndexes) && unmarshaled.EnabledGPUIndexes[i] != index {
			t.Errorf("EnabledGPUIndexes[%d] mismatch: expected %s, got %s",
				i, index, unmarshaled.EnabledGPUIndexes[i])
		}
	}

	// Verify other fields are preserved
	if unmarshaled.Status != testResult.Status {
		t.Errorf("Status mismatch: expected %s, got %s", testResult.Status, unmarshaled.Status)
	}
	if unmarshaled.GPUCount != testResult.GPUCount {
		t.Errorf("GPUCount mismatch: expected %d, got %d", testResult.GPUCount, unmarshaled.GPUCount)
	}
}
