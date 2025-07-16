package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test helper functions

func createTestReporter() *Reporter {
	return &Reporter{
		results: make(map[string]TestResult),
	}
}

func createTempFile(t *testing.T, filename string) string {
	tempDir := t.TempDir()
	return filepath.Join(tempDir, filename)
}

func assertResultCount(t *testing.T, reporter *Reporter, expected int) {
	if count := reporter.GetResultsCount(); count != expected {
		t.Errorf("Expected %d results, got %d", expected, count)
	}
}

func assertResultExists(t *testing.T, reporter *Reporter, testName string, expectedStatus string) {
	result, exists := reporter.results[testName]
	if !exists {
		t.Errorf("Result '%s' not found", testName)
		return
	}
	if result.Status != expectedStatus {
		t.Errorf("Expected status '%s' for %s, got '%s'", expectedStatus, testName, result.Status)
	}
}

func assertJSONValid(t *testing.T, data []byte) ReportOutput {
	var report ReportOutput
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	return report
}

// Core functionality tests

func TestReporter_Initialize(t *testing.T) {
	reporter := createTestReporter()
	outputFile := createTempFile(t, "test_report.json")

	err := reporter.Initialize(outputFile)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	if reporter.outputFile != outputFile {
		t.Errorf("Expected output file %s, got %s", outputFile, reporter.outputFile)
	}
	if !reporter.initialized {
		t.Error("Reporter should be initialized")
	}
}

func TestReporter_SingletonPattern(t *testing.T) {
	reporter1 := GetReporter()
	reporter2 := GetReporter()

	if reporter1 != reporter2 {
		t.Error("GetReporter should return the same instance")
	}
	if reporter1.hostname != "localhost" {
		t.Errorf("Expected default hostname 'localhost', got %s", reporter1.hostname)
	}
}

func TestReporter_BasicOperations(t *testing.T) {
	reporter := createTestReporter()

	// Test adding results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	assertResultCount(t, reporter, 2)

	// Test getting results
	assertResultExists(t, reporter, "gpu_count_check", "PASS")
	assertResultExists(t, reporter, "pcie_error_check", "PASS")

	// Test clear
	reporter.Clear()
	assertResultCount(t, reporter, 0)
}

func TestReporter_ResultCounting(t *testing.T) {
	reporter := createTestReporter()

	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("FAIL", fmt.Errorf("error"))
	reporter.AddRDMAResult("PASS", 16, nil)

	passedTests := reporter.GetPassedTests()
	failedTests := reporter.GetFailedTests()

	if len(passedTests) != 2 {
		t.Errorf("Expected 2 passed tests, got %d", len(passedTests))
	}
	if len(failedTests) != 1 {
		t.Errorf("Expected 1 failed test, got %d", len(failedTests))
	}
}

// Test result types - using table-driven tests for extensibility

func TestReporter_AllResultTypes(t *testing.T) {
	tests := []struct {
		name       string
		addFunc    func(*Reporter)
		resultKey  string
		wantStatus string
	}{
		{
			name: "GPU Result",
			addFunc: func(r *Reporter) {
				r.AddGPUResult("PASS", 8, nil)
			},
			resultKey:  "gpu_count_check",
			wantStatus: "PASS",
		},
		{
			name: "GPU Mode Result",
			addFunc: func(r *Reporter) {
				r.AddGPUModeResult("PASS", "MIG disabled", []string{}, nil)
			},
			resultKey:  "gpu_mode_check",
			wantStatus: "PASS",
		},
		{
			name: "PCIe Result",
			addFunc: func(r *Reporter) {
				r.AddPCIeResult("PASS", nil)
			},
			resultKey:  "pcie_error_check",
			wantStatus: "PASS",
		},
		{
			name: "RDMA Result",
			addFunc: func(r *Reporter) {
				r.AddRDMAResult("PASS", 16, nil)
			},
			resultKey:  "rdma_nic_count",
			wantStatus: "PASS",
		},
		{
			name: "Network Result",
			addFunc: func(r *Reporter) {
				r.AddRXDiscardsCheckResult("PASS", 16, []string{}, nil)
			},
			resultKey:  "rx_discards_check",
			wantStatus: "PASS",
		},
		{
			name: "GID Index Result",
			addFunc: func(r *Reporter) {
				r.AddGIDIndexResult("PASS", []int{}, nil)
			},
			resultKey:  "gid_index_check",
			wantStatus: "PASS",
		},
		{
			name: "Link Result",
			addFunc: func(r *Reporter) {
				r.AddLinkResult("PASS", []map[string]interface{}{{"device": "rdma0"}}, nil)
			},
			resultKey:  "link_check",
			wantStatus: "PASS",
		},
		{
			name: "Ethernet Link Result",
			addFunc: func(r *Reporter) {
				r.AddEthLinkResult("PASS", []map[string]interface{}{{"device": "eth0"}}, nil)
			},
			resultKey:  "eth_link_check",
			wantStatus: "PASS",
		},
		{
			name: "SRAM Error Result",
			addFunc: func(r *Reporter) {
				r.AddSRAMErrorResult("PASS", 0, 25, nil)
			},
			resultKey:  "sram_error_check",
			wantStatus: "PASS",
		},
		{
			name: "GPU Driver Result",
			addFunc: func(r *Reporter) {
				r.AddGPUDriverResult("PASS", "550.54.15", nil)
			},
			resultKey:  "gpu_driver_check",
			wantStatus: "PASS",
		},
		{
			name: "PeerMem Result",
			addFunc: func(r *Reporter) {
				r.AddPeerMemResult("PASS", true, nil)
			},
			resultKey:  "peermem_module_check",
			wantStatus: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := createTestReporter()
			tt.addFunc(reporter)
			assertResultExists(t, reporter, tt.resultKey, tt.wantStatus)
		})
	}
}

// Test specific result details

func TestReporter_ResultDetails(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Reporter)
		checkFunc func(*testing.T, TestResult)
		resultKey string
	}{
		{
			name: "SRAM Error Details",
			setupFunc: func(r *Reporter) {
				r.AddSRAMErrorResult("PASS", 5, 100, nil)
			},
			resultKey: "sram_error_check",
			checkFunc: func(t *testing.T, result TestResult) {
				if maxUncorr, ok := result.Details["max_uncorrectable"]; !ok || maxUncorr != 5 {
					t.Errorf("Expected max_uncorrectable 5, got %v", maxUncorr)
				}
				if maxCorr, ok := result.Details["max_correctable"]; !ok || maxCorr != 100 {
					t.Errorf("Expected max_correctable 100, got %v", maxCorr)
				}
			},
		},
		{
			name: "Network Details with Failures",
			setupFunc: func(r *Reporter) {
				r.AddRXDiscardsCheckResult("FAIL", 16, []string{"rdma2", "rdma3"}, fmt.Errorf("failed"))
			},
			resultKey: "rx_discards_check",
			checkFunc: func(t *testing.T, result TestResult) {
				if count, ok := result.Details["interface_count"]; !ok || count != 16 {
					t.Errorf("Expected interface_count 16, got %v", count)
				}
				if failed, ok := result.Details["failed_count"]; !ok || failed != 2 {
					t.Errorf("Expected failed_count 2, got %v", failed)
				}
			},
		},
		{
			name: "GPU Mode Details",
			setupFunc: func(r *Reporter) {
				r.AddGPUModeResult("FAIL", "MIG enabled", []string{"0", "1"}, fmt.Errorf("mig enabled"))
			},
			resultKey: "gpu_mode_check",
			checkFunc: func(t *testing.T, result TestResult) {
				if msg, ok := result.Details["message"]; !ok || msg != "MIG enabled" {
					t.Errorf("Expected message 'MIG enabled', got %v", msg)
				}
				if indexes, ok := result.Details["enabled_gpu_indexes"]; !ok {
					t.Error("Expected enabled_gpu_indexes to be present")
				} else if idxSlice, ok := indexes.([]string); !ok || len(idxSlice) != 2 {
					t.Errorf("Expected 2 GPU indexes, got %v", indexes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := createTestReporter()
			tt.setupFunc(reporter)

			result, exists := reporter.results[tt.resultKey]
			if !exists {
				t.Fatalf("Result '%s' not found", tt.resultKey)
			}

			tt.checkFunc(t, result)
		})
	}
}

// Report generation tests

func TestReporter_GenerateReport(t *testing.T) {
	reporter := createTestReporter()

	// Add sample results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMErrorResult("PASS", 1, 75, nil)
	reporter.AddRXDiscardsCheckResult("FAIL", 16, []string{"rdma2"}, fmt.Errorf("error"))

	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Validate structure
	if len(report.Localhost.GPUCountCheck) != 1 {
		t.Error("Expected 1 GPU result")
	}
	if len(report.Localhost.SRAMErrorCheck) != 1 {
		t.Error("Expected 1 SRAM result")
	}
	if len(report.Localhost.RXDiscardsCheck) != 1 {
		t.Error("Expected 1 network result")
	}

	// Validate timestamps
	gpuResult := report.Localhost.GPUCountCheck[0]
	if _, err := time.Parse(time.RFC3339, gpuResult.TimestampUTC); err != nil {
		t.Errorf("Invalid timestamp format: %s", gpuResult.TimestampUTC)
	}
}

func TestReporter_WriteReport(t *testing.T) {
	reporter := createTestReporter()
	outputFile := createTempFile(t, "test_report.json")
	reporter.outputFile = outputFile

	// Add test data
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMErrorResult("FAIL", 15, 200, fmt.Errorf("threshold exceeded"))

	// Write report
	err := reporter.WriteReport()
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Validate file
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	report := assertJSONValid(t, data)

	// Validate content
	if len(report.Localhost.GPUCountCheck) != 1 {
		t.Error("Expected 1 GPU result in file")
	}
	if len(report.Localhost.SRAMErrorCheck) != 1 {
		t.Error("Expected 1 SRAM result in file")
	}

	sramResult := report.Localhost.SRAMErrorCheck[0]
	if sramResult.Status != "FAIL" {
		t.Errorf("Expected SRAM status FAIL, got %s", sramResult.Status)
	}
}

// JSON serialization tests

func TestReporter_JSONSerialization(t *testing.T) {
	reporter := createTestReporter()

	// Add comprehensive test data
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMErrorResult("PASS", 2, 150, nil)
	reporter.AddLinkResult("PASS", []map[string]interface{}{
		{"device": "rdma0", "link_speed": "PASS"},
	}, nil)

	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var testReport ReportOutput
	if err := json.Unmarshal(jsonData, &testReport); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Validate structure preservation
	if len(testReport.Localhost.GPUCountCheck) != 1 {
		t.Error("GPU results not preserved in JSON")
	}
	if len(testReport.Localhost.SRAMErrorCheck) != 1 {
		t.Error("SRAM results not preserved in JSON")
	}
	if len(testReport.Localhost.LinkCheck) != 1 {
		t.Error("Link results not preserved in JSON")
	}
}

// Error handling tests

func TestReporter_ErrorHandling(t *testing.T) {
	reporter := createTestReporter()

	// Test with errors
	gpuErr := fmt.Errorf("GPU count mismatch")
	sramErr := fmt.Errorf("SRAM errors exceed threshold")

	reporter.AddGPUResult("FAIL", 6, gpuErr)
	reporter.AddSRAMErrorResult("FAIL", 20, 300, sramErr)

	// Verify errors are stored
	if result, exists := reporter.results["gpu_count_check"]; exists {
		if result.Error != gpuErr.Error() {
			t.Errorf("Expected GPU error %s, got %s", gpuErr.Error(), result.Error)
		}
	}

	if result, exists := reporter.results["sram_error_check"]; exists {
		if result.Error != sramErr.Error() {
			t.Errorf("Expected SRAM error %s, got %s", sramErr.Error(), result.Error)
		}
	}
}

// Edge cases and validation

func TestReporter_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "Zero Values",
			test: func(t *testing.T) {
				reporter := createTestReporter()
				reporter.AddSRAMErrorResult("PASS", 0, 0, nil)
				reporter.AddRXDiscardsCheckResult("PASS", 0, []string{}, nil)
				assertResultCount(t, reporter, 2)
			},
		},
		{
			name: "Large Values",
			test: func(t *testing.T) {
				reporter := createTestReporter()
				reporter.AddSRAMErrorResult("FAIL", 999, 10000, fmt.Errorf("excessive"))
				reporter.AddRXDiscardsCheckResult("PASS", 128, []string{}, nil)
				assertResultCount(t, reporter, 2)
			},
		},
		{
			name: "Empty Collections",
			test: func(t *testing.T) {
				reporter := createTestReporter()
				reporter.AddGIDIndexResult("PASS", []int{}, nil)
				reporter.AddLinkResult("PASS", []map[string]interface{}{}, nil)
				assertResultCount(t, reporter, 2)
			},
		},
		{
			name: "Configuration Methods",
			test: func(t *testing.T) {
				reporter := createTestReporter()
				reporter.SetHostname("test-host")
				reporter.SetAppendMode(false)

				if reporter.hostname != "test-host" {
					t.Error("Hostname not set correctly")
				}
				if reporter.GetAppendMode() != false {
					t.Error("Append mode not set correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Benchmark test for performance validation

func BenchmarkReporter_AddResults(b *testing.B) {
	reporter := createTestReporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reporter.AddGPUResult("PASS", 8, nil)
		reporter.Clear()
	}
}

func BenchmarkReporter_GenerateReport(b *testing.B) {
	reporter := createTestReporter()
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMErrorResult("PASS", 1, 50, nil)
	reporter.AddRDMAResult("PASS", 16, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := reporter.GenerateReport()
		if err != nil {
			b.Fatal(err)
		}
	}
}
