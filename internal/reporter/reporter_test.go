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

func TestReporter_AddNetworkRxDiscardsResult(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test adding network result with PASS status
	reporter.AddNetworkRxDiscardsResult("PASS", 16, nil)

	// Check if result was added
	if len(reporter.results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(reporter.results))
	}

	// Check network result
	if result, exists := reporter.results["network_rx_discards"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected network status PASS, got %s", result.Status)
		}
		if interfaceCount, ok := result.Details["interface_count"]; ok {
			if interfaceCount != 16 {
				t.Errorf("Expected interface count 16, got %v", interfaceCount)
			}
		} else {
			t.Error("Interface count not found in details")
		}
		// For PASS status, failed_count should not be present
		if _, ok := result.Details["failed_count"]; ok {
			t.Error("Failed count should not be present for PASS status")
		}
	} else {
		t.Error("Network result not found")
	}
}

func TestReporter_AddNetworkRxDiscardsResultWithFailure(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test adding network result with FAIL status
	networkErr := fmt.Errorf("2 interfaces failed RX discards check")
	reporter.AddNetworkRxDiscardsResult("FAIL", 2, networkErr)

	// Check network result
	if result, exists := reporter.results["network_rx_discards"]; exists {
		if result.Status != "FAIL" {
			t.Errorf("Expected network status FAIL, got %s", result.Status)
		}
		if interfaceCount, ok := result.Details["interface_count"]; ok {
			if interfaceCount != 2 {
				t.Errorf("Expected interface count 2, got %v", interfaceCount)
			}
		} else {
			t.Error("Interface count not found in details")
		}
		if result.Error != networkErr.Error() {
			t.Errorf("Expected network error %s, got %s", networkErr.Error(), result.Error)
		}
	} else {
		t.Error("Network result not found")
	}
}

func TestReporter_GenerateReportWithNetwork(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add test results including network
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddNetworkRxDiscardsResult("PASS", 16, nil)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check Network result
	if len(report.Localhost.NetworkRXDiscards) != 1 {
		t.Errorf("Expected 1 Network result, got %d", len(report.Localhost.NetworkRXDiscards))
	} else {
		if report.Localhost.NetworkRXDiscards[0].Status != "PASS" {
			t.Errorf("Expected Network status PASS, got %s", report.Localhost.NetworkRXDiscards[0].Status)
		}
		if report.Localhost.NetworkRXDiscards[0].InterfaceCount != 16 {
			t.Errorf("Expected Network interface count 16, got %d", report.Localhost.NetworkRXDiscards[0].InterfaceCount)
		}
		if report.Localhost.NetworkRXDiscards[0].FailedCount != 0 {
			t.Errorf("Expected Network failed count 0, got %d", report.Localhost.NetworkRXDiscards[0].FailedCount)
		}
	}
}

func TestReporter_GenerateReportWithNetworkFailure(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add network result with failure
	networkErr := fmt.Errorf("3 interfaces failed RX discards check")
	reporter.AddNetworkRxDiscardsResult("FAIL", 3, networkErr)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check Network result with failure
	if len(report.Localhost.NetworkRXDiscards) != 1 {
		t.Errorf("Expected 1 Network result, got %d", len(report.Localhost.NetworkRXDiscards))
	} else {
		if report.Localhost.NetworkRXDiscards[0].Status != "FAIL" {
			t.Errorf("Expected Network status FAIL, got %s", report.Localhost.NetworkRXDiscards[0].Status)
		}
		if report.Localhost.NetworkRXDiscards[0].InterfaceCount != 3 {
			t.Errorf("Expected Network interface count 3, got %d", report.Localhost.NetworkRXDiscards[0].InterfaceCount)
		}
		if report.Localhost.NetworkRXDiscards[0].FailedCount != 3 {
			t.Errorf("Expected Network failed count 3, got %d", report.Localhost.NetworkRXDiscards[0].FailedCount)
		}
	}
}

func TestReporter_WriteReportWithNetwork(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report_network.json")

	reporter := &Reporter{
		results:    make(map[string]TestResult),
		outputFile: outputFile,
	}

	// Add test results including network
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddNetworkRxDiscardsResult("PASS", 16, nil)

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

	// Validate the network structure in file
	if len(report.Localhost.NetworkRXDiscards) != 1 {
		t.Errorf("Expected 1 Network result in file, got %d", len(report.Localhost.NetworkRXDiscards))
	}

	// Verify network result content
	networkResult := report.Localhost.NetworkRXDiscards[0]
	if networkResult.Status != "PASS" {
		t.Errorf("Expected Network status PASS in file, got %s", networkResult.Status)
	}
	if networkResult.InterfaceCount != 16 {
		t.Errorf("Expected Network interface count 16 in file, got %d", networkResult.InterfaceCount)
	}
}

func TestReporter_AllResultTypesWithNetwork(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add all types of results including network
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("FAIL", fmt.Errorf("PCIe error found"))
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddNetworkRxDiscardsResult("FAIL", 2, fmt.Errorf("2 interfaces failed"))

	// Test counts
	if len(reporter.results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(reporter.results))
	}

	// Test failed tests (should include PCIe and Network)
	failedTests := reporter.GetFailedTests()
	if len(failedTests) != 2 {
		t.Errorf("Expected 2 failed tests, got %d", len(failedTests))
	}

	// Test passed tests (should include GPU and RDMA)
	passedTests := reporter.GetPassedTests()
	if len(passedTests) != 2 {
		t.Errorf("Expected 2 passed tests, got %d", len(passedTests))
	}

	// Check if network test is in failed tests
	networkTestFound := false
	for _, testName := range failedTests {
		if testName == "network_rx_discards" {
			networkTestFound = true
			break
		}
	}
	if !networkTestFound {
		t.Error("Network test should be in failed tests list")
	}
}

func TestReporter_NetworkResultEdgeCases(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test with zero interface count
	reporter.AddNetworkRxDiscardsResult("FAIL", 0, fmt.Errorf("no interfaces found"))

	if result, exists := reporter.results["network_rx_discards"]; exists {
		if interfaceCount, ok := result.Details["interface_count"]; ok {
			if interfaceCount != 0 {
				t.Errorf("Expected interface count 0, got %v", interfaceCount)
			}
		}
	}

	// Clear and test with large interface count
	reporter.Clear()
	reporter.AddNetworkRxDiscardsResult("PASS", 32, nil)

	if result, exists := reporter.results["network_rx_discards"]; exists {
		if interfaceCount, ok := result.Details["interface_count"]; ok {
			if interfaceCount != 32 {
				t.Errorf("Expected interface count 32, got %v", interfaceCount)
			}
		}
	}
}

func TestReporter_NetworkJSONMarshal(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add network result
	reporter.AddNetworkRxDiscardsResult("PASS", 16, nil)

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

	// Verify network-specific fields
	if len(testReport.Localhost.NetworkRXDiscards) != 1 {
		t.Error("Network RX discards field not properly marshaled")
	}

	networkResult := testReport.Localhost.NetworkRXDiscards[0]
	if networkResult.Status != "PASS" {
		t.Error("Network status field not properly marshaled")
	}
	if networkResult.InterfaceCount != 16 {
		t.Error("Network interface_count field not properly marshaled")
	}
	if networkResult.FailedCount != 0 {
		t.Error("Network failed_count field should be 0 for PASS status")
	}

	t.Logf("Generated JSON with network results:\n%s", string(jsonData))
}

func TestReporter_NetworkJSONMarshalWithFailure(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add network result with failure
	reporter.AddNetworkRxDiscardsResult("FAIL", 3, fmt.Errorf("interfaces failed"))

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

	// Verify network failure fields
	networkResult := testReport.Localhost.NetworkRXDiscards[0]
	if networkResult.Status != "FAIL" {
		t.Error("Network status field should be FAIL")
	}
	if networkResult.InterfaceCount != 3 {
		t.Error("Network interface_count field not properly marshaled for failure case")
	}
	if networkResult.FailedCount != 3 {
		t.Error("Network failed_count field not properly marshaled for failure case")
	}

	t.Logf("Generated JSON with network failure:\n%s", string(jsonData))
}
