package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// TestResult represents a single test result
type TestResult struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// GPUTestResult represents GPU test results
type GPUTestResult struct {
	Status       string `json:"status"`
	GPUCount     int    `json:"gpu_count,omitempty"`
	TimestampUTC string `json:"timestamp_utc"`
}

// PCIeTestResult represents PCIe test results
type PCIeTestResult struct {
	Status       string `json:"status"`
	TimestampUTC string `json:"timestamp_utc"`
}

// RDMATestResult represents RDMA test results
type RDMATestResult struct {
	Status       string `json:"status"`
	NumRDMANics  int    `json:"num_rdma_nics"`
	TimestampUTC string `json:"timestamp_utc"`
}

// NetworkTestResult represents network test results
type NetworkRXDiscardsTestResult struct {
	InterfaceCount int    `json:"interface_count"`
	FailedCount    int    `json:"failed_count,omitempty"`
	Status         string `json:"status"`
	TimestampUTC   string `json:"timestamp_utc"`
}

// HostResults represents test results for a host
type HostResults struct {
	GPUCountCheck     []GPUTestResult               `json:"gpu_count_check,omitempty"`
	PCIeErrorCheck    []PCIeTestResult              `json:"pcie_error_check,omitempty"`
	RDMANicsCount     []RDMATestResult              `json:"rdma_nics_count,omitempty"`
	NetworkRXDiscards []NetworkRXDiscardsTestResult `json:"network_rx_discards,omitempty"`
}

// ReportOutput represents the final JSON output structure
type ReportOutput struct {
	Localhost HostResults `json:"localhost"`
}

// TestRun represents a single test run with timestamp
type TestRun struct {
	RunID       string      `json:"run_id"`
	Timestamp   string      `json:"timestamp"`
	TestResults HostResults `json:"test_results"`
}

// AppendedReport represents multiple test runs in a single file
type AppendedReport struct {
	TestRuns []TestRun `json:"test_runs"`
}

// Reporter handles collecting and formatting test results
type Reporter struct {
	mutex       sync.RWMutex
	results     map[string]TestResult
	outputFile  string
	hostname    string
	initialized bool
	appendMode  bool
}

// Global reporter instance
var globalReporter *Reporter
var once sync.Once

// GetReporter returns the global reporter instance
func GetReporter() *Reporter {
	once.Do(func() {
		globalReporter = &Reporter{
			results:    make(map[string]TestResult),
			hostname:   "localhost", // Default hostname
			appendMode: true,        // Default to append mode
		}
	})
	return globalReporter
}

// Initialize sets up the reporter with configuration
func (r *Reporter) Initialize(outputFile string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.outputFile = outputFile
	r.initialized = true

	// Create output directory if it doesn't exist
	if outputFile != "" {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	logger.Debugf("Reporter initialized with output file: %s", outputFile)
	return nil
}

// SetAppendMode sets whether to append to existing files or overwrite them
func (r *Reporter) SetAppendMode(append bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.appendMode = append
}

// SetHostname sets the hostname for the report
func (r *Reporter) SetHostname(hostname string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.hostname = hostname
}

// AddResult adds a test result to the reporter
func (r *Reporter) AddResult(testName string, status string, details map[string]interface{}, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	result := TestResult{
		Name:      testName,
		Status:    status,
		Details:   details,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
	}

	r.results[testName] = result
	logger.Debugf("Added test result: %s = %s", testName, status)
}

// AddGPUResult adds GPU test results
func (r *Reporter) AddGPUResult(status string, gpuCount int, err error) {
	details := map[string]interface{}{
		"gpu_count": gpuCount,
	}
	r.AddResult("gpu_count_check", status, details, err)
}

// AddPCIeResult adds PCIe test results
func (r *Reporter) AddPCIeResult(status string, err error) {
	details := map[string]interface{}{}
	r.AddResult("pcie_error_check", status, details, err)
}

// AddRDMAResult adds RDMA test results
func (r *Reporter) AddRDMAResult(status string, rdmaNicCount int, err error) {
	details := map[string]interface{}{
		"rdma_nic_count": rdmaNicCount,
	}
	r.AddResult("rdma_nic_count", status, details, err)
}

// AddNetworkRxDiscardsResult adds network discards test results
func (r *Reporter) AddNetworkRxDiscardsResult(status string, interfaceCount int, err error) {
	details := map[string]interface{}{
		"interface_count": interfaceCount,
	}

	// Add failed count if status is FAIL and it's available from the error
	if status == "FAIL" {
		// For network tests, interfaceCount might represent failed interfaces
		// when status is FAIL, otherwise it represents total interfaces checked
		if status == "FAIL" {
			details["failed_count"] = interfaceCount
		}
	}

	r.AddResult("network_rx_discards", status, details, err)
}

// GenerateReport generates the final JSON report
func (r *Reporter) GenerateReport() (*ReportOutput, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	report := &ReportOutput{
		Localhost: HostResults{},
	}

	// Process GPU results
	if result, exists := r.results["gpu_count_check"]; exists {
		gpuCount := 0
		if countVal, ok := result.Details["gpu_count"]; ok {
			if count, ok := countVal.(int); ok {
				gpuCount = count
			}
		}
		gpuResult := GPUTestResult{
			Status:       result.Status,
			GPUCount:     gpuCount,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.GPUCountCheck = []GPUTestResult{gpuResult}
	}

	// Process PCIe results
	if result, exists := r.results["pcie_error_check"]; exists {
		pcieResult := PCIeTestResult{
			Status:       result.Status,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.PCIeErrorCheck = []PCIeTestResult{pcieResult}
	}

	// Process RDMA results
	if result, exists := r.results["rdma_nic_count"]; exists {
		rdmaCount := 0
		if countVal, ok := result.Details["rdma_nic_count"]; ok {
			if count, ok := countVal.(int); ok {
				rdmaCount = count
			}
		}
		rdmaResult := RDMATestResult{
			Status:       result.Status,
			NumRDMANics:  rdmaCount,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.RDMANicsCount = []RDMATestResult{rdmaResult}
	}

	// Process Network results
	if result, exists := r.results["network_rx_discards"]; exists {
		interfaceCount := 0
		failedCount := 0

		if countVal, ok := result.Details["interface_count"]; ok {
			if count, ok := countVal.(int); ok {
				interfaceCount = count
			}
		}

		if failedVal, ok := result.Details["failed_count"]; ok {
			if count, ok := failedVal.(int); ok {
				failedCount = count
			}
		}

		networkResult := NetworkRXDiscardsTestResult{
			InterfaceCount: interfaceCount,
			FailedCount:    failedCount,
			Status:         result.Status,
			TimestampUTC:   result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.NetworkRXDiscards = []NetworkRXDiscardsTestResult{networkResult}
	}

	return report, nil
}

// WriteReport writes the report to the configured output
func (r *Reporter) WriteReport() error {
	// Use default format (json) for backward compatibility
	return r.WriteReportWithFormat("json")
}

// WriteReportWithFormat writes the report with the specified format
func (r *Reporter) WriteReportWithFormat(format string) error {
	report, err := r.GenerateReport()
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	var output string
	switch format {
	case "json":
		output, err = r.formatJSON(report)
	case "table":
		output, err = r.formatTable(report)
	case "friendly":
		output, err = r.formatFriendly(report)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Write to file if configured
	if r.outputFile != "" {
		if r.appendMode && format == "json" {
			err = r.appendToFile(report)
		} else {
			err = os.WriteFile(r.outputFile, []byte(output), 0644)
		}

		if err != nil {
			return fmt.Errorf("failed to write report to file %s: %w", r.outputFile, err)
		}
		logger.Infof("Report written to file: %s", r.outputFile)
	} else {
		// Write to console if no file specified
		fmt.Print(output)
	}

	return nil
}

// appendToFile appends the current test results to an existing file
func (r *Reporter) appendToFile(currentReport *ReportOutput) error {
	var appendedReport AppendedReport

	// Try to read existing file
	if _, err := os.Stat(r.outputFile); err == nil {
		// File exists, read it
		existingData, err := os.ReadFile(r.outputFile)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}

		// Try to parse as AppendedReport first
		if err := json.Unmarshal(existingData, &appendedReport); err != nil {
			// If that fails, try to parse as single ReportOutput (backward compatibility)
			var singleReport ReportOutput
			if err := json.Unmarshal(existingData, &singleReport); err != nil {
				return fmt.Errorf("failed to parse existing file as JSON: %w", err)
			}

			// Convert single report to appended format
			appendedReport.TestRuns = []TestRun{
				{
					RunID:       fmt.Sprintf("run_%d", time.Now().Unix()),
					Timestamp:   time.Now().UTC().Format(time.RFC3339),
					TestResults: singleReport.Localhost,
				},
			}
		}
	}

	// Add current test run
	newRun := TestRun{
		RunID:       fmt.Sprintf("run_%d", time.Now().Unix()),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		TestResults: currentReport.Localhost,
	}
	appendedReport.TestRuns = append(appendedReport.TestRuns, newRun)

	// Write back to file
	jsonData, err := json.MarshalIndent(appendedReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal appended report: %w", err)
	}

	if err := os.WriteFile(r.outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write appended report: %w", err)
	}

	logger.Infof("Test results appended to file: %s", r.outputFile)
	return nil
}

// formatJSON formats the report as JSON
func (r *Reporter) formatJSON(report *ReportOutput) (string, error) {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}
	return string(jsonData) + "\n", nil
}

// formatTable formats the report as a table
func (r *Reporter) formatTable(report *ReportOutput) (string, error) {
	var output strings.Builder

	output.WriteString("┌─────────────────────────────────────────────────────────────────┐\n")
	output.WriteString("│                    DIAGNOSTIC TEST RESULTS                      │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString("│ TEST NAME              │ STATUS │ DETAILS                       │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// GPU Tests
	if len(report.Localhost.GPUCountCheck) > 0 {
		for _, gpu := range report.Localhost.GPUCountCheck {
			status := gpu.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s GPU Count: %s           │\n",
				"GPU Count Check", statusSymbol, statusSymbol, gpu.Status))
		}
	}

	// PCIe Tests
	if len(report.Localhost.PCIeErrorCheck) > 0 {
		for _, pcie := range report.Localhost.PCIeErrorCheck {
			status := pcie.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s PCIe Status: %s         │\n",
				"PCIe Error Check", statusSymbol, statusSymbol, pcie.Status))
		}
	}

	// RDMA Tests
	if len(report.Localhost.RDMANicsCount) > 0 {
		for _, rdma := range report.Localhost.RDMANicsCount {
			status := rdma.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s RDMA NICs: %d             │\n",
				"RDMA NIC Count", statusSymbol, statusSymbol, rdma.NumRDMANics))
		}
	}

	// Network Tests
	if len(report.Localhost.NetworkRXDiscards) > 0 {
		for _, network := range report.Localhost.NetworkRXDiscards {
			status := network.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := fmt.Sprintf("Interfaces: %d", network.InterfaceCount)
			if network.FailedCount > 0 {
				details = fmt.Sprintf("Failed: %d/%d", network.FailedCount, network.InterfaceCount)
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s            │\n",
				"Network RX Discards", statusSymbol, statusSymbol, details))
		}
	}

	output.WriteString("└─────────────────────────────────────────────────────────────────┘\n")
	return output.String(), nil
}

// formatFriendly formats the report in a user-friendly format
func (r *Reporter) formatFriendly(report *ReportOutput) (string, error) {
	var output strings.Builder

	output.WriteString("🔍 HPC Diagnostic Results\n")
	output.WriteString("=" + strings.Repeat("=", 50) + "\n\n")

	totalTests := 0
	passedTests := 0
	failedTests := 0

	// GPU Tests
	if len(report.Localhost.GPUCountCheck) > 0 {
		output.WriteString("🖥️  GPU Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, gpu := range report.Localhost.GPUCountCheck {
			totalTests++
			if gpu.Status == "PASS" {
				passedTests++
				output.WriteString(fmt.Sprintf("   ✅ GPU Count: %d (PASSED)\n", gpu.GPUCount))
			} else {
				failedTests++
				output.WriteString(fmt.Sprintf("   ❌ GPU Count: %d (FAILED)\n", gpu.GPUCount))
			}
		}
		output.WriteString("\n")
	}

	// PCIe Tests
	if len(report.Localhost.PCIeErrorCheck) > 0 {
		output.WriteString("🔗 PCIe Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, pcie := range report.Localhost.PCIeErrorCheck {
			totalTests++
			if pcie.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ PCIe Bus: No errors detected (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ PCIe Bus: Errors detected (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// RDMA Tests
	if len(report.Localhost.RDMANicsCount) > 0 {
		output.WriteString("🌐 RDMA Network Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, rdma := range report.Localhost.RDMANicsCount {
			totalTests++
			if rdma.Status == "PASS" {
				passedTests++
				output.WriteString(fmt.Sprintf("   ✅ RDMA NICs: %d detected (PASSED)\n", rdma.NumRDMANics))
			} else {
				failedTests++
				output.WriteString(fmt.Sprintf("   ❌ RDMA NICs: %d detected (FAILED)\n", rdma.NumRDMANics))
			}
		}
		output.WriteString("\n")
	}

	// Network Tests
	if len(report.Localhost.NetworkRXDiscards) > 0 {
		output.WriteString("🌐 Network RX Discards Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, network := range report.Localhost.NetworkRXDiscards {
			totalTests++
			if network.Status == "PASS" {
				passedTests++
				output.WriteString(fmt.Sprintf("   ✅ Network Interfaces: %d checked, no RX discard issues (PASSED)\n", network.InterfaceCount))
			} else {
				failedTests++
				if network.FailedCount > 0 {
					output.WriteString(fmt.Sprintf("   ❌ Network Interfaces: %d failed out of %d checked (FAILED)\n", network.FailedCount, network.InterfaceCount))
				} else {
					output.WriteString(fmt.Sprintf("   ❌ Network Interfaces: RX discard check failed (FAILED)\n"))
				}
			}
		}
		output.WriteString("\n")
	}

	// Summary
	output.WriteString("📊 Summary\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	output.WriteString(fmt.Sprintf("   Total Tests: %d\n", totalTests))
	output.WriteString(fmt.Sprintf("   Passed: %d\n", passedTests))
	output.WriteString(fmt.Sprintf("   Failed: %d\n", failedTests))

	if failedTests == 0 {
		output.WriteString("\n   🎉 All tests passed! Your HPC environment is healthy.\n")
	} else {
		output.WriteString(fmt.Sprintf("\n   ⚠️  %d test(s) failed. Please review the results above.\n", failedTests))
	}

	return output.String(), nil
}

// GetResults returns all collected results
func (r *Reporter) GetResults() map[string]TestResult {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	results := make(map[string]TestResult)
	for k, v := range r.results {
		results[k] = v
	}
	return results
}

// Clear clears all collected results
func (r *Reporter) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.results = make(map[string]TestResult)
}

// GetResultsCount returns the number of collected results
func (r *Reporter) GetResultsCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.results)
}

// GetFailedTests returns a list of failed test names
func (r *Reporter) GetFailedTests() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var failedTests []string
	for _, result := range r.results {
		if result.Status == "FAIL" {
			failedTests = append(failedTests, result.Name)
		}
	}
	return failedTests
}

// GetPassedTests returns a list of passed test names
func (r *Reporter) GetPassedTests() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var passedTests []string
	for _, result := range r.results {
		if result.Status == "PASS" {
			passedTests = append(passedTests, result.Name)
		}
	}
	return passedTests
}

// PrintSummary prints a summary of test results
func (r *Reporter) PrintSummary() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	totalTests := len(r.results)
	passedTests := r.GetPassedTests()
	failedTests := r.GetFailedTests()

	fmt.Printf("\n=== Test Summary ===\n")
	fmt.Printf("Total tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", len(passedTests))
	fmt.Printf("Failed: %d\n", len(failedTests))

	if len(failedTests) > 0 {
		fmt.Printf("Failed tests: %v\n", failedTests)
	}

	if len(failedTests) == 0 {
		fmt.Printf("✅ All tests passed!\n")
	} else {
		fmt.Printf("❌ %d test(s) failed\n", len(failedTests))
	}
}

// IsInitialized returns whether the reporter has been initialized
func (r *Reporter) IsInitialized() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.initialized
}

// GetAppendMode returns whether append mode is enabled
func (r *Reporter) GetAppendMode() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.appendMode
}
