package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReporter_Initialize(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report.json")

	// Create a new reporter instance
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test initialization
	err := reporter.Initialize(outputFile)
	if err != nil {
		t.Fatalf("Failed to initialize reporter: %v", err)
	}

	if reporter.outputFile != outputFile {
		t.Errorf("Expected output file %s, got %s", outputFile, reporter.outputFile)
	}

	if !reporter.initialized {
		t.Error("Reporter should be marked as initialized")
	}
}

func TestReporter_AddResults(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test adding GPU result
	reporter.AddGPUResult("PASS", 8, nil)

	// Test adding PCIe result
	reporter.AddPCIeResult("PASS", nil)

	// Test adding RDMA result
	reporter.AddRDMAResult("PASS", 16, nil)

	// Check if all results were added
	if len(reporter.results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(reporter.results))
	}

	// Check GPU result
	if result, exists := reporter.results["gpu_count_check"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected GPU status PASS, got %s", result.Status)
		}
		if gpuCount, ok := result.Details["gpu_count"]; ok {
			if gpuCount != 8 {
				t.Errorf("Expected GPU count 8, got %v", gpuCount)
			}
		} else {
			t.Error("GPU count not found in details")
		}
	} else {
		t.Error("GPU result not found")
	}

	// Check PCIe result
	if result, exists := reporter.results["pcie_error_check"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected PCIe status PASS, got %s", result.Status)
		}
	} else {
		t.Error("PCIe result not found")
	}

	// Check RDMA result
	if result, exists := reporter.results["rdma_nic_count"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected RDMA status PASS, got %s", result.Status)
		}
		if rdmaCount, ok := result.Details["rdma_nic_count"]; ok {
			if rdmaCount != 16 {
				t.Errorf("Expected RDMA count 16, got %v", rdmaCount)
			}
		} else {
			t.Error("RDMA count not found in details")
		}
	} else {
		t.Error("RDMA result not found")
	}
}

func TestReporter_GenerateReport(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add test results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check GPU result
	if len(report.Localhost.GPUCountCheck) != 1 {
		t.Errorf("Expected 1 GPU result, got %d", len(report.Localhost.GPUCountCheck))
	} else {
		if report.Localhost.GPUCountCheck[0].Status != "PASS" {
			t.Errorf("Expected GPU status PASS, got %s", report.Localhost.GPUCountCheck[0].Status)
		}
		if report.Localhost.GPUCountCheck[0].GPUCount != 8 {
			t.Errorf("Expected GPU count 8, got %d", report.Localhost.GPUCountCheck[0].GPUCount)
		}
	}

	// Check PCIe result
	if len(report.Localhost.PCIeErrorCheck) != 1 {
		t.Errorf("Expected 1 PCIe result, got %d", len(report.Localhost.PCIeErrorCheck))
	} else {
		if report.Localhost.PCIeErrorCheck[0].Status != "PASS" {
			t.Errorf("Expected PCIe status PASS, got %s", report.Localhost.PCIeErrorCheck[0].Status)
		}
	}

	// Check RDMA result
	if len(report.Localhost.RDMANicsCount) != 1 {
		t.Errorf("Expected 1 RDMA result, got %d", len(report.Localhost.RDMANicsCount))
	} else {
		if report.Localhost.RDMANicsCount[0].Status != "PASS" {
			t.Errorf("Expected RDMA status PASS, got %s", report.Localhost.RDMANicsCount[0].Status)
		}
		if report.Localhost.RDMANicsCount[0].NumRDMANics != 16 {
			t.Errorf("Expected RDMA count 16, got %d", report.Localhost.RDMANicsCount[0].NumRDMANics)
		}
	}
}

func TestReporter_WriteReport(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report.json")

	reporter := &Reporter{
		results:    make(map[string]TestResult),
		outputFile: outputFile,
	}

	// Add test results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)

	// Write report
	err := reporter.WriteReport()
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	// Read and validate the JSON
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var report ReportOutput
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Failed to unmarshal report JSON: %v", err)
	}

	// Validate the structure
	if len(report.Localhost.GPUCountCheck) != 1 {
		t.Errorf("Expected 1 GPU result in file, got %d", len(report.Localhost.GPUCountCheck))
	}
	if len(report.Localhost.PCIeErrorCheck) != 1 {
		t.Errorf("Expected 1 PCIe result in file, got %d", len(report.Localhost.PCIeErrorCheck))
	}
	if len(report.Localhost.RDMANicsCount) != 1 {
		t.Errorf("Expected 1 RDMA result in file, got %d", len(report.Localhost.RDMANicsCount))
	}
}

func TestReporter_FailedResults(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add mixed results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("FAIL", fmt.Errorf("PCIe error found"))
	reporter.AddRDMAResult("FAIL", 14, fmt.Errorf("RDMA count mismatch"))

	// Test failed tests
	failedTests := reporter.GetFailedTests()
	if len(failedTests) != 2 {
		t.Errorf("Expected 2 failed tests, got %d", len(failedTests))
	}

	// Test passed tests
	passedTests := reporter.GetPassedTests()
	if len(passedTests) != 1 {
		t.Errorf("Expected 1 passed test, got %d", len(passedTests))
	}

	// Check if the failed tests are correct
	expectedFailedTests := []string{"pcie_error_check", "rdma_nic_count"}
	for _, expected := range expectedFailedTests {
		found := false
		for _, actual := range failedTests {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected failed test %s not found in results", expected)
		}
	}
}

func TestReporter_Clear(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add some results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)

	// Verify results were added
	if len(reporter.results) != 2 {
		t.Errorf("Expected 2 results before clear, got %d", len(reporter.results))
	}

	// Clear results
	reporter.Clear()

	// Verify results were cleared
	if len(reporter.results) != 0 {
		t.Errorf("Expected 0 results after clear, got %d", len(reporter.results))
	}
}

func TestReporter_GetReporter(t *testing.T) {
	// Test singleton pattern
	reporter1 := GetReporter()
	reporter2 := GetReporter()

	if reporter1 != reporter2 {
		t.Error("GetReporter should return the same instance")
	}

	// Test that it's properly initialized
	if reporter1.results == nil {
		t.Error("Reporter results map should be initialized")
	}

	if reporter1.hostname != "localhost" {
		t.Errorf("Expected default hostname 'localhost', got %s", reporter1.hostname)
	}
}

func TestReporter_SetHostname(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	testHostname := "test-host"
	reporter.SetHostname(testHostname)

	if reporter.hostname != testHostname {
		t.Errorf("Expected hostname %s, got %s", testHostname, reporter.hostname)
	}
}

func TestReporter_ResultsCount(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Initially should be 0
	if count := reporter.GetResultsCount(); count != 0 {
		t.Errorf("Expected 0 results initially, got %d", count)
	}

	// Add results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)

	// Should be 2 now
	if count := reporter.GetResultsCount(); count != 2 {
		t.Errorf("Expected 2 results after adding, got %d", count)
	}
}

func TestReporter_WithErrors(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add results with errors
	gpuErr := fmt.Errorf("GPU count mismatch")
	reporter.AddGPUResult("FAIL", 6, gpuErr)

	pcieErr := fmt.Errorf("PCIe error detected")
	reporter.AddPCIeResult("FAIL", pcieErr)

	rdmaErr := fmt.Errorf("RDMA NIC count mismatch")
	reporter.AddRDMAResult("FAIL", 14, rdmaErr)

	// Check that errors are stored
	if result, exists := reporter.results["gpu_count_check"]; exists {
		if result.Error != gpuErr.Error() {
			t.Errorf("Expected GPU error %s, got %s", gpuErr.Error(), result.Error)
		}
	}

	if result, exists := reporter.results["pcie_error_check"]; exists {
		if result.Error != pcieErr.Error() {
			t.Errorf("Expected PCIe error %s, got %s", pcieErr.Error(), result.Error)
		}
	}

	if result, exists := reporter.results["rdma_nic_count"]; exists {
		if result.Error != rdmaErr.Error() {
			t.Errorf("Expected RDMA error %s, got %s", rdmaErr.Error(), result.Error)
		}
	}
}

func TestReporter_Timestamp(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	before := time.Now()
	reporter.AddGPUResult("PASS", 8, nil)
	after := time.Now()

	if result, exists := reporter.results["gpu_count_check"]; exists {
		if result.Timestamp.Before(before) || result.Timestamp.After(after) {
			t.Error("Timestamp should be set to current time when result is added")
		}
	} else {
		t.Error("GPU result not found")
	}
}

func TestReporter_JSONMarshal(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add test results
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal report to JSON: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var testReport ReportOutput
	if err := json.Unmarshal(jsonData, &testReport); err != nil {
		t.Fatalf("Generated JSON is not valid: %v", err)
	}

	// Verify the JSON structure contains expected fields
	if !json.Valid(jsonData) {
		t.Error("Generated JSON is not valid")
	}

	// Verify specific fields exist
	if testReport.Localhost.GPUCountCheck[0].Status != "PASS" {
		t.Error("GPU status field not properly marshaled")
	}
	if testReport.Localhost.GPUCountCheck[0].GPUCount != 8 {
		t.Error("GPU gpu_count field not properly marshaled")
	}
	if testReport.Localhost.GPUCountCheck[0].TimestampUTC == "" {
		t.Error("GPU timestamp_utc field not properly marshaled")
	}
	if testReport.Localhost.PCIeErrorCheck[0].Status != "PASS" {
		t.Error("PCIe status field not properly marshaled")
	}
	if testReport.Localhost.PCIeErrorCheck[0].TimestampUTC == "" {
		t.Error("PCIe timestamp_utc field not properly marshaled")
	}
	if testReport.Localhost.RDMANicsCount[0].NumRDMANics != 16 {
		t.Error("RDMA num_rdma_nics field not properly marshaled")
	}
	if testReport.Localhost.RDMANicsCount[0].Status != "PASS" {
		t.Error("RDMA status field not properly marshaled")
	}
	if testReport.Localhost.RDMANicsCount[0].TimestampUTC == "" {
		t.Error("RDMA timestamp_utc field not properly marshaled")
	}

	// Verify timestamp format (should be RFC3339 format)
	if _, err := time.Parse(time.RFC3339, testReport.Localhost.GPUCountCheck[0].TimestampUTC); err != nil {
		t.Errorf("GPU timestamp_utc is not in valid RFC3339 format: %s", testReport.Localhost.GPUCountCheck[0].TimestampUTC)
	}
	if _, err := time.Parse(time.RFC3339, testReport.Localhost.PCIeErrorCheck[0].TimestampUTC); err != nil {
		t.Errorf("PCIe timestamp_utc is not in valid RFC3339 format: %s", testReport.Localhost.PCIeErrorCheck[0].TimestampUTC)
	}
	if _, err := time.Parse(time.RFC3339, testReport.Localhost.RDMANicsCount[0].TimestampUTC); err != nil {
		t.Errorf("RDMA timestamp_utc is not in valid RFC3339 format: %s", testReport.Localhost.RDMANicsCount[0].TimestampUTC)
	}

	t.Logf("Generated JSON structure matches expected format:\n%s", string(jsonData))
}
