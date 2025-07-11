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

	// Test adding GID Index result
	reporter.AddGIDIndexResult("PASS", []int{}, nil)

	// Test adding Link result
	linkResults := []map[string]interface{}{
		{"device": "rdma0", "link_speed": "PASS", "link_state": "PASS"},
	}
	reporter.AddLinkResult("PASS", linkResults, nil)

	// Test adding SRAM result
	reporter.AddSRAMResult("PASS", 0, 50, nil)

	// Check if all results were added
	if len(reporter.results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(reporter.results))
	}

	// Check SRAM result
	if result, exists := reporter.results["sram_error_check"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected SRAM status PASS, got %s", result.Status)
		}
		if maxUncorr, ok := result.Details["max_uncorrectable"]; ok {
			if maxUncorr != 0 {
				t.Errorf("Expected max_uncorrectable 0, got %v", maxUncorr)
			}
		} else {
			t.Error("max_uncorrectable not found in details")
		}
		if maxCorr, ok := result.Details["max_correctable"]; ok {
			if maxCorr != 50 {
				t.Errorf("Expected max_correctable 50, got %v", maxCorr)
			}
		} else {
			t.Error("max_correctable not found in details")
		}
	} else {
		t.Error("SRAM result not found")
	}
}

func TestReporter_AddSRAMResult(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	tests := []struct {
		name             string
		status           string
		maxUncorrectable int
		maxCorrectable   int
		expectError      bool
	}{
		{
			name:             "PASS with low errors",
			status:           "PASS",
			maxUncorrectable: 0,
			maxCorrectable:   25,
			expectError:      false,
		},
		{
			name:             "WARN with correctable errors",
			status:           "WARN",
			maxUncorrectable: 2,
			maxCorrectable:   1500,
			expectError:      true,
		},
		{
			name:             "FAIL with uncorrectable errors",
			status:           "FAIL",
			maxUncorrectable: 10,
			maxCorrectable:   200,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous results
			reporter.Clear()

			var err error
			if tt.expectError {
				err = fmt.Errorf("SRAM errors detected")
			}

			reporter.AddSRAMResult(tt.status, tt.maxUncorrectable, tt.maxCorrectable, err)

			// Verify result was added
			if len(reporter.results) != 1 {
				t.Errorf("Expected 1 result, got %d", len(reporter.results))
			}

			result, exists := reporter.results["sram_error_check"]
			if !exists {
				t.Fatal("SRAM result not found")
			}

			if result.Status != tt.status {
				t.Errorf("Expected status %s, got %s", tt.status, result.Status)
			}

			if maxUncorr, ok := result.Details["max_uncorrectable"]; ok {
				if maxUncorr != tt.maxUncorrectable {
					t.Errorf("Expected max_uncorrectable %d, got %v", tt.maxUncorrectable, maxUncorr)
				}
			}

			if maxCorr, ok := result.Details["max_correctable"]; ok {
				if maxCorr != tt.maxCorrectable {
					t.Errorf("Expected max_correctable %d, got %v", tt.maxCorrectable, maxCorr)
				}
			}

			if tt.expectError && result.Error == "" {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Error != "" {
				t.Errorf("Unexpected error: %s", result.Error)
			}
		})
	}
}

func TestReporter_GenerateReportWithSRAM(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add test results including SRAM
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddSRAMResult("PASS", 1, 75, nil)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check SRAM result
	if len(report.Localhost.SRAMErrorCheck) != 1 {
		t.Errorf("Expected 1 SRAM result, got %d", len(report.Localhost.SRAMErrorCheck))
	} else {
		sramResult := report.Localhost.SRAMErrorCheck[0]
		if sramResult.Status != "PASS" {
			t.Errorf("Expected SRAM status PASS, got %s", sramResult.Status)
		}
		if sramResult.MaxUncorrectable != 1 {
			t.Errorf("Expected MaxUncorrectable 1, got %d", sramResult.MaxUncorrectable)
		}
		if sramResult.MaxCorrectable != 75 {
			t.Errorf("Expected MaxCorrectable 75, got %d", sramResult.MaxCorrectable)
		}
		if sramResult.TimestampUTC == "" {
			t.Error("Expected SRAM TimestampUTC to be set")
		}
	}
}

func TestReporter_WriteReportWithSRAM(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report_sram.json")

	reporter := &Reporter{
		results:    make(map[string]TestResult),
		outputFile: outputFile,
	}

	// Add test results including SRAM
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMResult("FAIL", 15, 200, fmt.Errorf("uncorrectable errors exceed threshold"))

	// Write report
	err := reporter.WriteReport()
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
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

	// Validate the SRAM structure in file
	if len(report.Localhost.SRAMErrorCheck) != 1 {
		t.Errorf("Expected 1 SRAM result in file, got %d", len(report.Localhost.SRAMErrorCheck))
	}

	sramResult := report.Localhost.SRAMErrorCheck[0]
	if sramResult.Status != "FAIL" {
		t.Errorf("Expected SRAM status FAIL, got %s", sramResult.Status)
	}
	if sramResult.MaxUncorrectable != 15 {
		t.Errorf("Expected MaxUncorrectable 15, got %d", sramResult.MaxUncorrectable)
	}
	if sramResult.MaxCorrectable != 200 {
		t.Errorf("Expected MaxCorrectable 200, got %d", sramResult.MaxCorrectable)
	}
}

func TestReporter_FailedResultsWithSRAM(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add mixed results including SRAM
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("FAIL", fmt.Errorf("PCIe error found"))
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddSRAMResult("FAIL", 10, 500, fmt.Errorf("SRAM errors exceed threshold"))

	// Test failed tests (should include PCIe and SRAM)
	failedTests := reporter.GetFailedTests()
	if len(failedTests) != 2 {
		t.Errorf("Expected 2 failed tests, got %d", len(failedTests))
	}

	// Test passed tests (should include GPU and RDMA)
	passedTests := reporter.GetPassedTests()
	if len(passedTests) != 2 {
		t.Errorf("Expected 2 passed tests, got %d", len(passedTests))
	}

	// Check if SRAM test is in failed tests
	sramTestFound := false
	for _, testName := range failedTests {
		if testName == "sram_error_check" {
			sramTestFound = true
			break
		}
	}
	if !sramTestFound {
		t.Error("SRAM test should be in failed tests list")
	}
}

func TestReporter_SRAMJSONMarshal(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add SRAM result
	reporter.AddSRAMResult("PASS", 2, 150, nil)

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

	// Verify SRAM-specific fields
	if len(testReport.Localhost.SRAMErrorCheck) != 1 {
		t.Fatal("SRAM error check field not properly marshaled")
	}

	sramResult := testReport.Localhost.SRAMErrorCheck[0]
	if sramResult.Status != "PASS" {
		t.Error("SRAM status field not properly marshaled")
	}
	if sramResult.MaxUncorrectable != 2 {
		t.Error("SRAM max_uncorrectable field not properly marshaled")
	}
	if sramResult.MaxCorrectable != 150 {
		t.Error("SRAM max_correctable field not properly marshaled")
	}
	if sramResult.TimestampUTC == "" {
		t.Error("SRAM timestamp_utc field not properly marshaled")
	}

	// Verify timestamp format
	if _, err := time.Parse(time.RFC3339, sramResult.TimestampUTC); err != nil {
		t.Errorf("SRAM timestamp_utc is not in valid RFC3339 format: %s", sramResult.TimestampUTC)
	}

	t.Logf("Generated JSON with SRAM results:\n%s", string(jsonData))
}

func TestReporter_Clear(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add some results including SRAM
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMResult("PASS", 0, 25, nil)

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

	// Add results including SRAM
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddSRAMResult("PASS", 1, 50, nil)

	// Should be 2 now
	if count := reporter.GetResultsCount(); count != 2 {
		t.Errorf("Expected 2 results after adding, got %d", count)
	}
}

func TestReporter_WithErrors(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add results with errors including SRAM
	gpuErr := fmt.Errorf("GPU count mismatch")
	reporter.AddGPUResult("FAIL", 6, gpuErr)

	sramErr := fmt.Errorf("SRAM uncorrectable errors exceed threshold")
	reporter.AddSRAMResult("FAIL", 20, 300, sramErr)

	// Check that errors are stored
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

func TestReporter_Timestamp(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	before := time.Now()
	reporter.AddSRAMResult("PASS", 0, 25, nil)
	after := time.Now()

	if result, exists := reporter.results["sram_error_check"]; exists {
		if result.Timestamp.Before(before) || result.Timestamp.After(after) {
			t.Error("Timestamp should be set to current time when result is added")
		}
	} else {
		t.Error("SRAM result not found")
	}
}

func TestReporter_SRAMEdgeCases(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test with zero error counts
	reporter.AddSRAMResult("PASS", 0, 0, nil)

	if result, exists := reporter.results["sram_error_check"]; exists {
		if maxUncorr, ok := result.Details["max_uncorrectable"]; ok {
			if maxUncorr != 0 {
				t.Errorf("Expected max_uncorrectable 0, got %v", maxUncorr)
			}
		}
		if maxCorr, ok := result.Details["max_correctable"]; ok {
			if maxCorr != 0 {
				t.Errorf("Expected max_correctable 0, got %v", maxCorr)
			}
		}
	}

	// Clear and test with large error counts
	reporter.Clear()
	reporter.AddSRAMResult("FAIL", 999, 10000, fmt.Errorf("excessive errors"))

	if result, exists := reporter.results["sram_error_check"]; exists {
		if maxUncorr, ok := result.Details["max_uncorrectable"]; ok {
			if maxUncorr != 999 {
				t.Errorf("Expected max_uncorrectable 999, got %v", maxUncorr)
			}
		}
		if maxCorr, ok := result.Details["max_correctable"]; ok {
			if maxCorr != 10000 {
				t.Errorf("Expected max_correctable 10000, got %v", maxCorr)
			}
		}
	}
}

func TestReporter_NetworkResultEdgeCases(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test with zero interface count
	reporter.AddRXDiscardsCheckResult("FAIL", 0, []string{}, fmt.Errorf("no interfaces found"))

	if result, exists := reporter.results["rx_discards_check"]; exists {
		if interfaceCount, ok := result.Details["interface_count"]; ok {
			if interfaceCount != 0 {
				t.Errorf("Expected interface count 0, got %v", interfaceCount)
			}
		}
	}

	// Clear and test with large interface count
	reporter.Clear()
	reporter.AddRXDiscardsCheckResult("PASS", 32, []string{}, nil)

	if result, exists := reporter.results["rx_discards_check"]; exists {
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
	reporter.AddRXDiscardsCheckResult("PASS", 16, []string{}, nil)

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
	if len(testReport.Localhost.RXDiscardsCheck) != 1 {
		t.Error("Network RX discards field not properly marshaled")
	}

	networkResult := testReport.Localhost.RXDiscardsCheck[0]
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
	reporter.AddRXDiscardsCheckResult("FAIL", 16, []string{"rdma2", "rdma3", "rdma6"}, fmt.Errorf("interfaces failed"))

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
	networkResult := testReport.Localhost.RXDiscardsCheck[0]
	if networkResult.Status != "FAIL" {
		t.Error("Network status field should be FAIL")
	}
	if networkResult.InterfaceCount != 16 {
		t.Error("Network interface_count field not properly marshaled for failure case")
	}
	if networkResult.FailedCount != 3 {
		t.Error("Network failed_count field not properly marshaled for failure case")
	}

	t.Logf("Generated JSON with network failure:\n%s", string(jsonData))
}

func TestReporter_AddLinkResult(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test adding link result with PASS status
	linkResults := []map[string]interface{}{
		{
			"device":     "rdma0",
			"link_speed": "PASS",
			"link_state": "PASS",
		},
		{
			"device":     "rdma1",
			"link_speed": "PASS",
			"link_state": "PASS",
		},
	}
	reporter.AddLinkResult("PASS", linkResults, nil)

	// Check if result was added
	if len(reporter.results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(reporter.results))
	}

	// Check link result
	if result, exists := reporter.results["link_check"]; exists {
		if result.Status != "PASS" {
			t.Errorf("Expected link status PASS, got %s", result.Status)
		}
		if links, ok := result.Details["links"]; ok {
			if linksSlice, ok := links.([]map[string]interface{}); ok {
				if len(linksSlice) != 2 {
					t.Errorf("Expected 2 link results, got %d", len(linksSlice))
				}
			} else {
				t.Error("Links should be a slice of maps")
			}
		} else {
			t.Error("Links not found in details")
		}
	} else {
		t.Error("Link result not found")
	}
}

func TestReporter_AddLinkResultWithFailure(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Test adding link result with FAIL status
	linkResults := []map[string]interface{}{
		{
			"device":     "rdma0",
			"link_speed": "FAIL - 100G, expected 200G",
			"link_state": "FAIL - Down, expected Active",
		},
	}
	linkErr := fmt.Errorf("link issues detected")
	reporter.AddLinkResult("FAIL", linkResults, linkErr)

	// Check link result
	if result, exists := reporter.results["link_check"]; exists {
		if result.Status != "FAIL" {
			t.Errorf("Expected link status FAIL, got %s", result.Status)
		}
		if result.Error != linkErr.Error() {
			t.Errorf("Expected link error %s, got %s", linkErr.Error(), result.Error)
		}
		if links, ok := result.Details["links"]; ok {
			if linksSlice, ok := links.([]map[string]interface{}); ok {
				if len(linksSlice) != 1 {
					t.Errorf("Expected 1 link result, got %d", len(linksSlice))
				}
			} else {
				t.Error("Links should be a slice of maps")
			}
		} else {
			t.Error("Links not found in details")
		}
	} else {
		t.Error("Link result not found")
	}
}

func TestReporter_GenerateReportWithLink(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add test results including link
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)
	linkResults := []map[string]interface{}{
		{
			"device":     "rdma0",
			"link_speed": "PASS",
			"link_state": "PASS",
		},
	}
	reporter.AddLinkResult("PASS", linkResults, nil)

	// Generate report
	report, err := reporter.GenerateReport()
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check Link result
	if len(report.Localhost.LinkCheck) != 1 {
		t.Errorf("Expected 1 Link result, got %d", len(report.Localhost.LinkCheck))
	} else {
		if report.Localhost.LinkCheck[0].Status != "PASS" {
			t.Errorf("Expected Link status PASS, got %s", report.Localhost.LinkCheck[0].Status)
		}
		if report.Localhost.LinkCheck[0].Links == nil {
			t.Error("Expected Link Links to be present")
		}
	}
}

func TestReporter_WriteReportWithLink(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report_link.json")

	reporter := &Reporter{
		results:    make(map[string]TestResult),
		outputFile: outputFile,
	}

	// Add test results including link
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("PASS", nil)
	reporter.AddRDMAResult("PASS", 16, nil)
	linkResults := []map[string]interface{}{
		{
			"device":     "rdma0",
			"link_speed": "PASS",
			"link_state": "PASS",
		},
	}
	reporter.AddLinkResult("PASS", linkResults, nil)

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

	// Validate the link structure in file
	if len(report.Localhost.LinkCheck) != 1 {
		t.Errorf("Expected 1 Link result in file, got %d", len(report.Localhost.LinkCheck))
	}

	// Verify link result content
	linkResult := report.Localhost.LinkCheck[0]
	if linkResult.Status != "PASS" {
		t.Errorf("Expected Link status PASS in file, got %s", linkResult.Status)
	}
	if linkResult.Links == nil {
		t.Error("Expected Link Links to be present in file")
	}
}

func TestReporter_AllResultTypesWithLink(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add all types of results including link
	reporter.AddGPUResult("PASS", 8, nil)
	reporter.AddPCIeResult("FAIL", fmt.Errorf("PCIe error found"))
	reporter.AddRDMAResult("PASS", 16, nil)
	reporter.AddRXDiscardsCheckResult("PASS", 16, []string{}, nil)
	linkResults := []map[string]interface{}{
		{
			"device":     "rdma0",
			"link_speed": "FAIL - 100G, expected 200G",
			"link_state": "PASS",
		},
	}
	reporter.AddLinkResult("FAIL", linkResults, fmt.Errorf("link issues detected"))

	// Test counts
	if len(reporter.results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(reporter.results))
	}

	// Test failed tests (should include PCIe and Link)
	failedTests := reporter.GetFailedTests()
	if len(failedTests) != 2 {
		t.Errorf("Expected 2 failed tests, got %d", len(failedTests))
	}

	// Test passed tests (should include GPU, RDMA, and RX Discards)
	passedTests := reporter.GetPassedTests()
	if len(passedTests) != 3 {
		t.Errorf("Expected 3 passed tests, got %d", len(passedTests))
	}

	// Check if link test is in failed tests
	linkTestFound := false
	for _, testName := range failedTests {
		if testName == "link_check" {
			linkTestFound = true
			break
		}
	}
	if !linkTestFound {
		t.Error("Link test should be in failed tests list")
	}
}

func TestReporter_LinkJSONMarshal(t *testing.T) {
	reporter := &Reporter{
		results: make(map[string]TestResult),
	}

	// Add link result
	linkResults := []map[string]interface{}{
		{
			"device":                       "rdma0",
			"link_speed":                   "PASS",
			"link_state":                   "PASS",
			"physical_state":               "PASS",
			"link_width":                   "PASS",
			"link_status":                  "PASS",
			"effective_physical_errors":    "PASS",
			"effective_physical_ber":       "PASS",
			"raw_physical_errors_per_lane": "PASS",
			"raw_physical_ber":             "PASS",
		},
	}
	reporter.AddLinkResult("PASS", linkResults, nil)

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

	// Verify link-specific fields
	if len(testReport.Localhost.LinkCheck) != 1 {
		t.Error("Link check field not properly marshaled")
	}

	linkResult := testReport.Localhost.LinkCheck[0]
	if linkResult.Status != "PASS" {
		t.Error("Link status field not properly marshaled")
	}
	if linkResult.Links == nil {
		t.Error("Link links field not properly marshaled")
	}
	if linkResult.TimestampUTC == "" {
		t.Error("Link timestamp_utc field not properly marshaled")
	}

	// Verify timestamp format (should be RFC3339 format)
	if _, err := time.Parse(time.RFC3339, linkResult.TimestampUTC); err != nil {
		t.Errorf("Link timestamp_utc is not in valid RFC3339 format: %s", linkResult.TimestampUTC)
	}

	t.Logf("Generated JSON with link results:\n%s", string(jsonData))
}
