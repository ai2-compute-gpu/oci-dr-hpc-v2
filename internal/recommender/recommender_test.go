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
		Summary:        "Test summary with 1 critical issue",
		TotalIssues:    1,
		CriticalIssues: 1,
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
	if unmarshaled.TotalIssues != 1 {
		t.Errorf("Expected TotalIssues=1, got %d", unmarshaled.TotalIssues)
	}
	if unmarshaled.CriticalIssues != 1 {
		t.Errorf("Expected CriticalIssues=1, got %d", unmarshaled.CriticalIssues)
	}
	if len(unmarshaled.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(unmarshaled.Recommendations))
	}
	if unmarshaled.Recommendations[0].Type != "critical" {
		t.Errorf("Expected type 'critical', got '%s'", unmarshaled.Recommendations[0].Type)
	}
}

func TestFormatRecommendationsTable(t *testing.T) {
	// Create a test recommendation report
	testReport := RecommendationReport{
		Summary:        "Test summary with mixed issues",
		TotalIssues:    2,
		CriticalIssues: 1,
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
		"Total Issues: 2",
		"Critical: 1",
		"Warning: 1",
		"RECOMMENDATIONS",
		"gpu_count_check",
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
		Summary:        "🎉 All diagnostic tests passed! Your HPC environment appears healthy.",
		TotalIssues:    0,
		CriticalIssues: 0,
		WarningIssues:  0,
		InfoIssues:     1,
		Recommendations: []Recommendation{
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
		"Total Issues: 0",
		"Critical: 0",
		"Warning: 0",
		"Info: 1",
		"📋 DETAILED RECOMMENDATIONS",
		"ℹ️ 1. INFO [gpu_count_check]",
		"Commands to run:",
		"$ nvidia-smi -q",
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
		TotalIssues:    1,
		CriticalIssues: 1,
		WarningIssues:  0,
		InfoIssues:     0,
		Recommendations: []Recommendation{
			{
				Type:       "critical",
				TestName:   "test_check",
				Issue:      "Test issue",
				Suggestion: "Test suggestion",
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
		Summary:        "Test summary with 2 issues",
		TotalIssues:    2,
		CriticalIssues: 1,
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

	// Verify first recommendation
	if len(unmarshaled.Recommendations) > 0 {
		origRec := original.Recommendations[0]
		unmarshaledRec := unmarshaled.Recommendations[0]
		
		if unmarshaledRec.Type != origRec.Type {
			t.Errorf("Recommendation type mismatch: expected %s, got %s", origRec.Type, unmarshaledRec.Type)
		}
		if len(unmarshaledRec.Commands) != len(origRec.Commands) {
			t.Errorf("Commands count mismatch: expected %d, got %d", len(origRec.Commands), len(unmarshaledRec.Commands))
		}
		if len(unmarshaledRec.References) != len(origRec.References) {
			t.Errorf("References count mismatch: expected %d, got %d", len(origRec.References), len(unmarshaledRec.References))
		}
	}
}