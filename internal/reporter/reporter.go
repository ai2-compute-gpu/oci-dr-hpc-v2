package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	DeviceCount string `json:"device_count"`
}

// PCIeTestResult represents PCIe test results
type PCIeTestResult struct {
	Status string `json:"status"`
}

// RDMATestResult represents RDMA test results
type RDMATestResult struct {
	NumRDMANics int    `json:"num_rdma_nics"`
	Status      string `json:"status"`
}

// HostResults represents test results for a host
type HostResults struct {
	GPU          []GPUTestResult  `json:"gpu,omitempty"`
	PCIeError    []PCIeTestResult `json:"pcie_error,omitempty"`
	RDMANicCount []RDMATestResult `json:"rdma_nic_count,omitempty"`
}

// ReportOutput represents the final JSON output structure
type ReportOutput struct {
	Localhost HostResults `json:"localhost"`
}

// Reporter handles collecting and formatting test results
type Reporter struct {
	mutex       sync.RWMutex
	results     map[string]TestResult
	outputFile  string
	hostname    string
	initialized bool
}

// Global reporter instance
var globalReporter *Reporter
var once sync.Once

// GetReporter returns the global reporter instance
func GetReporter() *Reporter {
	once.Do(func() {
		globalReporter = &Reporter{
			results:  make(map[string]TestResult),
			hostname: "localhost", // Default hostname
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

// GenerateReport generates the final JSON report
func (r *Reporter) GenerateReport() (*ReportOutput, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	report := &ReportOutput{
		Localhost: HostResults{},
	}

	// Process GPU results
	if result, exists := r.results["gpu_count_check"]; exists {
		gpuResult := GPUTestResult{
			DeviceCount: result.Status,
		}
		report.Localhost.GPU = []GPUTestResult{gpuResult}
	}

	// Process PCIe results
	if result, exists := r.results["pcie_error_check"]; exists {
		pcieResult := PCIeTestResult{
			Status: result.Status,
		}
		report.Localhost.PCIeError = []PCIeTestResult{pcieResult}
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
			NumRDMANics: rdmaCount,
			Status:      result.Status,
		}
		report.Localhost.RDMANicCount = []RDMATestResult{rdmaResult}
	}

	return report, nil
}

// WriteReport writes the report to the configured output
func (r *Reporter) WriteReport() error {
	report, err := r.GenerateReport()
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	// Write to file if configured
	if r.outputFile != "" {
		if err := os.WriteFile(r.outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write report to file %s: %w", r.outputFile, err)
		}
		logger.Infof("Report written to file: %s", r.outputFile)
	} else {
		// Write to console if no file specified
		fmt.Println(string(jsonData))
	}

	return nil
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
