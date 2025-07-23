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

// GPUModeTestResult represents GPU mode test results
type GPUModeTestResult struct {
	Status            string   `json:"status"`
	Message           string   `json:"message,omitempty"`
	EnabledGPUIndexes []string `json:"enabled_gpu_indexes,omitempty"`
	TimestampUTC      string   `json:"timestamp_utc"`
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
type RXDiscardsCheckTestResult struct {
	InterfaceCount   int    `json:"interface_count"`
	FailedCount      int    `json:"failed_count,omitempty"`
	FailedInterfaces string `json:"failed_interfaces,omitempty"`
	Status           string `json:"status"`
	TimestampUTC     string `json:"timestamp_utc"`
}

// GIDIndexTestResult represents GID index test results
type GIDIndexTestResult struct {
	Status         string `json:"status"`
	InvalidIndexes []int  `json:"invalid_indexes,omitempty"`
	TimestampUTC   string `json:"timestamp_utc"`
}

// LinkTestResult represents link check test results
type LinkTestResult struct {
	Status       string      `json:"status"`
	Links        interface{} `json:"links,omitempty"`
	TimestampUTC string      `json:"timestamp_utc"`
}

// EthLinkTestResult represents Ethernet link check test results
type EthLinkTestResult struct {
	Status       string      `json:"status"`
	EthLinks     interface{} `json:"eth_links,omitempty"`
	TimestampUTC string      `json:"timestamp_utc"`
}

// AuthCheckTestResult represents authentication check test results
type AuthCheckTestResult struct {
	Status       string      `json:"status"`
	Interfaces   interface{} `json:"interfaces,omitempty"`
	TimestampUTC string      `json:"timestamp_utc"`
}

// SRAMErrorTestResult represents SRAM error test results
type SRAMErrorTestResult struct {
	Status           string `json:"status"`
	MaxUncorrectable int    `json:"max_uncorrectable,omitempty"`
	MaxCorrectable   int    `json:"max_correctable,omitempty"`
	TimestampUTC     string `json:"timestamp_utc"`
}

// GPUDriverTestResult represents GPU driver test results
type GPUDriverTestResult struct {
	Status        string `json:"status"`
	DriverVersion string `json:"driver_version,omitempty"`
	TimestampUTC  string `json:"timestamp_utc"`
}

// PeerMemTestResult represents peermem module test results
type PeerMemTestResult struct {
	Status       string `json:"status"`
	ModuleLoaded bool   `json:"module_loaded"`
	TimestampUTC string `json:"timestamp_utc"`
}

// NVLinkTestResult represents NVLink test results
type NVLinkTestResult struct {
	Status       string      `json:"status"`
	NVLinks      interface{} `json:"nvlinks,omitempty"`
	TimestampUTC string      `json:"timestamp_utc"`
}

// GPUClockTestResult represents GPU clock speed test results
type GPUClockTestResult struct {
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
	TimestampUTC string `json:"timestamp_utc"`
}

type Eth0PresenceTestResult struct {
	Status       string `json:"status"`
	Eth0Present  bool   `json:"eth0_present"`
	TimestampUTC string `json:"timestamp_utc"`
}

// CDFPCableCheckTestResult represents CDFP cable check test results
type CDFPCableCheckTestResult struct {
	Status       string      `json:"status"`
	CDFPResult   interface{} `json:"cdfp_result,omitempty"`
	TimestampUTC string      `json:"timestamp_utc"`
}

// HostResults represents test results for a host
type HostResults struct {
	GPUCountCheck      []GPUTestResult             `json:"gpu_count_check,omitempty"`
	GPUModeCheck       []GPUModeTestResult         `json:"gpu_mode_check,omitempty"`
	PCIeErrorCheck     []PCIeTestResult            `json:"pcie_error_check,omitempty"`
	RDMANicsCount      []RDMATestResult            `json:"rdma_nics_count,omitempty"`
	RXDiscardsCheck    []RXDiscardsCheckTestResult `json:"rx_discards_check,omitempty"`
	GIDIndexCheck      []GIDIndexTestResult        `json:"gid_index_check,omitempty"`
	LinkCheck          []LinkTestResult            `json:"link_check,omitempty"`
	EthLinkCheck       []EthLinkTestResult         `json:"eth_link_check,omitempty"`
	AuthCheck          []AuthCheckTestResult       `json:"auth_check,omitempty"`
	SRAMErrorCheck     []SRAMErrorTestResult       `json:"sram_error_check,omitempty"`
	GPUDriverCheck     []GPUDriverTestResult       `json:"gpu_driver_check,omitempty"`
	GPUClockCheck      []GPUClockTestResult        `json:"gpu_clk_check,omitempty"`
	PeerMemModuleCheck []PeerMemTestResult         `json:"peermem_module_check,omitempty"`
	NVLinkSpeedCheck   []NVLinkTestResult          `json:"nvlink_speed_check,omitempty"`
	Eth0PresenceCheck  []Eth0PresenceTestResult    `json:"eth0_presence_check,omitempty"`
	CDFPCableCheck     []CDFPCableCheckTestResult  `json:"cdfp_cable_check,omitempty"`
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

// AddGPUModeResult adds GPU mode test results
func (r *Reporter) AddGPUModeResult(status string, message string, enabledGPUIndexes []string, err error) {
	details := map[string]interface{}{
		"message":             message,
		"enabled_gpu_indexes": enabledGPUIndexes,
	}
	r.AddResult("gpu_mode_check", status, details, err)
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

// AddRXDiscardsCheckResult adds network discards test results
func (r *Reporter) AddRXDiscardsCheckResult(status string, interfaceCount int, failedInterfaces []string, err error) {
	details := map[string]interface{}{
		"interface_count": interfaceCount,
	}

	// Add failed count if status is FAIL and it's available from the error
	if status == "FAIL" {
		// For network tests, interfaceCount might represent failed interfaces
		// when status is FAIL, otherwise it represents total interfaces checked
		details["failed_count"] = len(failedInterfaces)
		details["failed_interfaces"] = strings.Join(failedInterfaces, ",")
	}

	r.AddResult("rx_discards_check", status, details, err)
}

// AddGIDIndexResult adds GID index test results
func (r *Reporter) AddGIDIndexResult(status string, invalidIndexes []int, err error) {
	details := map[string]interface{}{
		"invalid_indexes": invalidIndexes,
	}
	r.AddResult("gid_index_check", status, details, err)
}

// AddLinkResult adds link check test results
func (r *Reporter) AddLinkResult(status string, links interface{}, err error) {
	details := map[string]interface{}{
		"links": links,
	}
	r.AddResult("link_check", status, details, err)
}

// AddEthLinkResult adds Ethernet link check test results
func (r *Reporter) AddEthLinkResult(status string, ethLinks interface{}, err error) {
	details := map[string]interface{}{
		"eth_links": ethLinks,
	}
	r.AddResult("eth_link_check", status, details, err)
}

// AddAuthCheckResult adds authentication check test results
func (r *Reporter) AddAuthCheckResult(status string, interfaces interface{}, err error) {
	details := map[string]interface{}{
		"interfaces": interfaces,
	}
	r.AddResult("auth_check", status, details, err)
}

// AddSRAMErrorResult adds SRAM error test results
func (r *Reporter) AddSRAMErrorResult(status string, maxUncorrectable int, maxCorrectable int, err error) {
	details := map[string]interface{}{
		"max_uncorrectable": maxUncorrectable,
		"max_correctable":   maxCorrectable,
	}
	r.AddResult("sram_error_check", status, details, err)
}

// AddGPUDriverResult adds GPU driver test results
func (r *Reporter) AddGPUDriverResult(status string, driverVersion string, err error) {
	details := map[string]interface{}{
		"driver_version": driverVersion,
	}
	r.AddResult("gpu_driver_check", status, details, err)
}

// AddGPUClockResult adds GPU clock speed test results
func (r *Reporter) AddGPUClockResult(status string, message string, err error) {
	details := map[string]interface{}{
		"message": message,
	}
	r.AddResult("gpu_clk_check", status, details, err)
}

// AddPeerMemResult adds peermem module test results
func (r *Reporter) AddPeerMemResult(status string, moduleLoaded bool, err error) {
	details := map[string]interface{}{
		"module_loaded": moduleLoaded,
	}
	r.AddResult("peermem_module_check", status, details, err)
}

// AddNVLinkResult adds NVLink test results
func (r *Reporter) AddNVLinkResult(status string, nvlinks interface{}, err error) {
	details := map[string]interface{}{}
	if status != "PASS" {
		details = map[string]interface{}{
			"nvlinks": nvlinks,
		}
	}

	r.AddResult("nvlink_speed_check", status, details, err)
}

// AddEth0PresenceResult adds eth0 presence test results
func (r *Reporter) AddEth0PresenceResult(status string, eth0Present bool, err error) {
	details := map[string]interface{}{
		"eth0_present": eth0Present,
	}
	r.AddResult("eth0_presence_check", status, details, err)
}

// AddCDFPCableCheckResult adds CDFP cable check test results
func (r *Reporter) AddCDFPCableCheckResult(status string, cdfpResult interface{}, err error) {
	details := map[string]interface{}{}
	if cdfpResult != nil {
		details = map[string]interface{}{
			"cdfp_result": cdfpResult,
		}
	}
	r.AddResult("cdfp_cable_check", status, details, err)
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

	// Process GPU Mode Check results
	if result, exists := r.results["gpu_mode_check"]; exists {
		message := ""
		if msgVal, ok := result.Details["message"]; ok {
			if msg, ok := msgVal.(string); ok {
				message = msg
			}
		}

		var enabledGPUIndexes []string
		if indexesVal, ok := result.Details["enabled_gpu_indexes"]; ok {
			if indexes, ok := indexesVal.([]string); ok {
				enabledGPUIndexes = indexes
			}
		}

		gpuModeResult := GPUModeTestResult{
			Status:            result.Status,
			Message:           message,
			EnabledGPUIndexes: enabledGPUIndexes,
			TimestampUTC:      result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.GPUModeCheck = []GPUModeTestResult{gpuModeResult}
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
	if result, exists := r.results["rx_discards_check"]; exists {
		interfaceCount := 0
		failedCount := 0
		failedInterfaces := ""

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

		if failedInterfacesInResult, ok := result.Details["failed_interfaces"]; ok {
			if failedInterfacesInResultStr, ok := failedInterfacesInResult.(string); ok {
				failedInterfaces = failedInterfacesInResultStr
			}
		}

		networkResult := RXDiscardsCheckTestResult{
			InterfaceCount:   interfaceCount,
			FailedCount:      failedCount,
			Status:           result.Status,
			FailedInterfaces: failedInterfaces,
			TimestampUTC:     result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.RXDiscardsCheck = []RXDiscardsCheckTestResult{networkResult}
	}

	// Process GID Index results
	if result, exists := r.results["gid_index_check"]; exists {
		var invalidIndexes []int
		if indexesVal, ok := result.Details["invalid_indexes"]; ok {
			if indexes, ok := indexesVal.([]int); ok {
				invalidIndexes = indexes
			}
		}
		gidResult := GIDIndexTestResult{
			Status:         result.Status,
			InvalidIndexes: invalidIndexes,
			TimestampUTC:   result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.GIDIndexCheck = []GIDIndexTestResult{gidResult}
	}

	// Process Link Check results
	if result, exists := r.results["link_check"]; exists {
		var links interface{}
		if linksVal, ok := result.Details["links"]; ok {
			links = linksVal
		}

		linkResult := LinkTestResult{
			Status:       result.Status,
			Links:        links,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.LinkCheck = []LinkTestResult{linkResult}
	}

	// Process Ethernet link check results
	if result, exists := r.results["eth_link_check"]; exists {
		var ethLinks interface{}
		if ethLinksVal, ok := result.Details["eth_links"]; ok {
			ethLinks = ethLinksVal
		}

		ethLinkResult := EthLinkTestResult{
			Status:       result.Status,
			EthLinks:     ethLinks,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.EthLinkCheck = []EthLinkTestResult{ethLinkResult}
	}

	// Process Auth check results
	if result, exists := r.results["auth_check"]; exists {
		var interfaces interface{}
		if interfacesVal, ok := result.Details["interfaces"]; ok {
			interfaces = interfacesVal
		}

		authResult := AuthCheckTestResult{
			Status:       result.Status,
			Interfaces:   interfaces,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.AuthCheck = []AuthCheckTestResult{authResult}
	}

	// Process SRAM Error Check results
	if result, exists := r.results["sram_error_check"]; exists {
		maxUncorrectable := 0
		if uncorrectableVal, ok := result.Details["max_uncorrectable"]; ok {
			if uncorrectable, ok := uncorrectableVal.(int); ok {
				maxUncorrectable = uncorrectable
			}
		}

		maxCorrectable := 0
		if correctableVal, ok := result.Details["max_correctable"]; ok {
			if correctable, ok := correctableVal.(int); ok {
				maxCorrectable = correctable
			}
		}

		sramResult := SRAMErrorTestResult{
			Status:           result.Status,
			MaxUncorrectable: maxUncorrectable,
			MaxCorrectable:   maxCorrectable,
			TimestampUTC:     result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.SRAMErrorCheck = []SRAMErrorTestResult{sramResult}
	}

	// Process GPU Driver Check results
	if result, exists := r.results["gpu_driver_check"]; exists {
		driverVersion := ""
		if versionVal, ok := result.Details["driver_version"]; ok {
			if version, ok := versionVal.(string); ok {
				driverVersion = version
			}
		}
		gpuDriverResult := GPUDriverTestResult{
			Status:        result.Status,
			DriverVersion: driverVersion,
			TimestampUTC:  result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.GPUDriverCheck = []GPUDriverTestResult{gpuDriverResult}
	}

	// Process GPU Clock Check results
	if result, exists := r.results["gpu_clk_check"]; exists {
		message := ""
		if messageVal, ok := result.Details["message"]; ok {
			if msg, ok := messageVal.(string); ok {
				message = msg
			}
		}
		gpuClockResult := GPUClockTestResult{
			Status:       result.Status,
			Message:      message,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.GPUClockCheck = []GPUClockTestResult{gpuClockResult}
	}

	// Process PeerMem Module results
	if result, exists := r.results["peermem_module_check"]; exists {
		moduleLoaded := false
		if loadedVal, ok := result.Details["module_loaded"]; ok {
			if loaded, ok := loadedVal.(bool); ok {
				moduleLoaded = loaded
			}
		}
		peerMemResult := PeerMemTestResult{
			Status:       result.Status,
			ModuleLoaded: moduleLoaded,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.PeerMemModuleCheck = []PeerMemTestResult{peerMemResult}
	}

	// Process NVLink Speed Check results
	if result, exists := r.results["nvlink_speed_check"]; exists {
		var nvlinks interface{}
		if nvlinksVal, ok := result.Details["nvlinks"]; ok {
			nvlinks = nvlinksVal
		}

		nvlinkResult := NVLinkTestResult{
			Status:       result.Status,
			NVLinks:      nvlinks,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.NVLinkSpeedCheck = []NVLinkTestResult{nvlinkResult}
	}

	// Process Eth0 Presence Check results
	if result, exists := r.results["eth0_presence_check"]; exists {
		eth0Present := false
		if presentVal, ok := result.Details["eth0_present"]; ok {
			if present, ok := presentVal.(bool); ok {
				eth0Present = present
			}
		}
		eth0Result := Eth0PresenceTestResult{
			Status:       result.Status,
			Eth0Present:  eth0Present,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.Eth0PresenceCheck = []Eth0PresenceTestResult{eth0Result}
	}

	// Process CDFP Cable Check results
	if result, exists := r.results["cdfp_cable_check"]; exists {
		var cdfpResult interface{}
		if cdfpVal, ok := result.Details["cdfp_result"]; ok {
			cdfpResult = cdfpVal
		}

		cdfpCableResult := CDFPCableCheckTestResult{
			Status:       result.Status,
			CDFPResult:   cdfpResult,
			TimestampUTC: result.Timestamp.UTC().Format(time.RFC3339),
		}
		report.Localhost.CDFPCableCheck = []CDFPCableCheckTestResult{cdfpCableResult}
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
	output.WriteString("│ TEST NAME              │ STATUS  │ DETAILS                      │\n")
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

	// GPU Mode Tests
	if len(report.Localhost.GPUModeCheck) > 0 {
		for _, gpuMode := range report.Localhost.GPUModeCheck {
			status := gpuMode.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "MIG Mode Disabled"
			if len(gpuMode.EnabledGPUIndexes) > 0 {
				details = fmt.Sprintf("MIG Enabled: %v", gpuMode.EnabledGPUIndexes)
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s         │\n",
				"GPU Mode Check", statusSymbol, statusSymbol, details))
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
	if len(report.Localhost.RXDiscardsCheck) > 0 {
		for _, network := range report.Localhost.RXDiscardsCheck {
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

	// GID Index Tests
	if len(report.Localhost.GIDIndexCheck) > 0 {
		for _, gid := range report.Localhost.GIDIndexCheck {
			status := gid.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "All indexes valid"
			if len(gid.InvalidIndexes) > 0 {
				details = fmt.Sprintf("invalid Index: %v", gid.InvalidIndexes)
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s         │\n",
				"GID Index Check", statusSymbol, statusSymbol, details))
		}
	}

	// Link Check Tests
	if len(report.Localhost.LinkCheck) > 0 {
		for _, link := range report.Localhost.LinkCheck {
			status := link.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "RDMA Links Checked"
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s        │\n",
				"RDMA Link Check", statusSymbol, statusSymbol, details))
		}
	}

	// Ethernet Link Check Tests
	if len(report.Localhost.EthLinkCheck) > 0 {
		for _, ethLink := range report.Localhost.EthLinkCheck {
			status := ethLink.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "Ethernet Links Checked"
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s    │\n",
				"Ethernet Link Check", statusSymbol, statusSymbol, details))
		}
	}

	// Auth Check Tests
	if len(report.Localhost.AuthCheck) > 0 {
		for _, auth := range report.Localhost.AuthCheck {
			status := auth.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "RDMA Auth Checked"
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s         │\n",
				"Authentication Check", statusSymbol, statusSymbol, details))
		}
	}

	// SRAM Tests
	if len(report.Localhost.SRAMErrorCheck) > 0 {
		for _, sram := range report.Localhost.SRAMErrorCheck {
			status := sram.Status
			statusSymbol := "✅"
			if status == "FAIL" || status == "WARN" {
				statusSymbol = "❌"
			}
			details := fmt.Sprintf("Uncorr: %d, Corr: %d", sram.MaxUncorrectable, sram.MaxCorrectable)
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s        │\n",
				"SRAM Error Check", statusSymbol, statusSymbol, details))
		}
	}

	// GPU Driver Check Tests
	if len(report.Localhost.GPUDriverCheck) > 0 {
		for _, driver := range report.Localhost.GPUDriverCheck {
			status := driver.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			} else if status == "WARN" {
				statusSymbol = "⚠️"
			}
			details := fmt.Sprintf("Version: %s", driver.DriverVersion)
			if len(details) > 25 {
				details = details[:22] + "..."
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s        │\n",
				"GPU Driver Check", statusSymbol, statusSymbol, details))
		}
	}

	// GPU Clock Check Tests
	if len(report.Localhost.GPUClockCheck) > 0 {
		for _, clock := range report.Localhost.GPUClockCheck {
			status := clock.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "Clock Speeds OK"
			if status == "FAIL" {
				details = "Clock Speed Issues"
			} else if clock.Message != "" {
				details = clock.Message
				if len(details) > 25 {
					details = details[:22] + "..."
				}
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s        │\n",
				"GPU Clock Check", statusSymbol, statusSymbol, details))
		}
	}

	// PeerMem Module Tests
	if len(report.Localhost.PeerMemModuleCheck) > 0 {
		for _, peerMem := range report.Localhost.PeerMemModuleCheck {
			status := peerMem.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "Module Loaded"
			if !peerMem.ModuleLoaded {
				details = "Module Not Loaded"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s             │\n",
				"PeerMem Module Check", statusSymbol, statusSymbol, details))
		}
	}

	// NVLink Tests
	if len(report.Localhost.NVLinkSpeedCheck) > 0 {
		for _, nvlink := range report.Localhost.NVLinkSpeedCheck {
			status := nvlink.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "NVLink Speed/Count OK"
			if status == "FAIL" {
				details = "NVLink Issues Found"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s     │\n",
				"NVLink Speed Check", statusSymbol, statusSymbol, details))
		}
	}

	// Eth0 Presence Tests
	if len(report.Localhost.Eth0PresenceCheck) > 0 {
		for _, eth0 := range report.Localhost.Eth0PresenceCheck {
			status := eth0.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			}
			details := "eth0 Interface Present"
			if !eth0.Eth0Present {
				details = "eth0 Interface Missing"
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s    │\n",
				"Eth0 Presence Check", statusSymbol, statusSymbol, details))
		}
	}

	// CDFP Cable Check Tests
	if len(report.Localhost.CDFPCableCheck) > 0 {
		for _, cdfp := range report.Localhost.CDFPCableCheck {
			status := cdfp.Status
			statusSymbol := "✅"
			if status == "FAIL" {
				statusSymbol = "❌"
			} else if status == "SKIP" {
				statusSymbol = "⏭️"
			}
			details := "CDFP Cables OK"
			if status == "FAIL" {
				details = "CDFP Cable Issues"
			} else if status == "SKIP" {
				details = "CDFP Check Skipped"
			}
			// Calculate padding to align with 67-character table width
			contentLength := 1 + 22 + 3 + 6 + 3 + 1 + len(statusSymbol) + 1 + len(details)
			padding := 67 - contentLength - 1 // -1 for final │
			if padding < 0 {
				padding = 0
			}
			output.WriteString(fmt.Sprintf("│ %-22s │ %-6s │ %s %s%s│\n",
				"CDFP Cable Check", statusSymbol, statusSymbol, details, strings.Repeat(" ", padding)))
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

	// GPU Mode Tests
	if len(report.Localhost.GPUModeCheck) > 0 {
		output.WriteString("🖥️  GPU Mode Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, gpuMode := range report.Localhost.GPUModeCheck {
			totalTests++
			if gpuMode.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ GPU Mode: MIG disabled on all GPUs (PASSED)\n")
			} else {
				failedTests++
				if len(gpuMode.EnabledGPUIndexes) > 0 {
					output.WriteString(fmt.Sprintf("   ❌ GPU Mode: MIG enabled on GPUs %v (FAILED)\n", gpuMode.EnabledGPUIndexes))
				} else {
					output.WriteString("   ❌ GPU Mode: Check failed (FAILED)\n")
				}
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
	if len(report.Localhost.RXDiscardsCheck) > 0 {
		output.WriteString("🌐 Network RX Discards Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, network := range report.Localhost.RXDiscardsCheck {
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

	// GID Index Tests
	if len(report.Localhost.GIDIndexCheck) > 0 {
		output.WriteString("🔗 GID Index Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, gid := range report.Localhost.GIDIndexCheck {
			totalTests++
			if gid.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ GID Indexes: All valid (PASSED)\n")
			} else {
				failedTests++
				if len(gid.InvalidIndexes) > 0 {
					output.WriteString(fmt.Sprintf("   ❌ GID Indexes: Invalid indexes found %v (FAILED)\n", gid.InvalidIndexes))
				} else {
					output.WriteString("   ❌ GID Indexes: Check failed (FAILED)\n")
				}
			}
		}
		output.WriteString("\n")
	}

	// Link Check Tests
	if len(report.Localhost.LinkCheck) > 0 {
		output.WriteString("🌐 RDMA Link Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, link := range report.Localhost.LinkCheck {
			totalTests++
			if link.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ RDMA Links: All links healthy (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ RDMA Links: Link issues detected (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// Ethernet Link Check Tests
	if len(report.Localhost.EthLinkCheck) > 0 {
		output.WriteString("🌐 Ethernet Link Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, ethLink := range report.Localhost.EthLinkCheck {
			totalTests++
			if ethLink.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ Ethernet Links: All links healthy (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ Ethernet Links: Link issues detected (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// Auth Check Tests
	if len(report.Localhost.AuthCheck) > 0 {
		output.WriteString("🔐 Authentication Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, auth := range report.Localhost.AuthCheck {
			totalTests++
			if auth.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ Authentication: All RDMA interfaces authenticated (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ Authentication: RDMA interface authentication issues (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// SRAM Tests
	if len(report.Localhost.SRAMErrorCheck) > 0 {
		output.WriteString("💾 SRAM Error Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, sram := range report.Localhost.SRAMErrorCheck {
			totalTests++
			if sram.Status == "PASS" {
				passedTests++
				output.WriteString(fmt.Sprintf("   ✅ SRAM Errors: Uncorrectable: %d, Correctable: %d (PASSED)\n",
					sram.MaxUncorrectable, sram.MaxCorrectable))
			} else {
				failedTests++
				if sram.Status == "WARN" {
					output.WriteString(fmt.Sprintf("   ⚠️  SRAM Errors: Uncorrectable: %d, Correctable: %d (WARNING)\n",
						sram.MaxUncorrectable, sram.MaxCorrectable))
				} else {
					output.WriteString(fmt.Sprintf("   ❌ SRAM Errors: Uncorrectable: %d, Correctable: %d (FAILED)\n",
						sram.MaxUncorrectable, sram.MaxCorrectable))
				}
			}
		}
		output.WriteString("\n")
	}

	// GPU Driver Check Tests
	if len(report.Localhost.GPUDriverCheck) > 0 {
		output.WriteString("🎮 GPU Driver Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, driver := range report.Localhost.GPUDriverCheck {
			totalTests++
			if driver.Status == "PASS" {
				passedTests++
				output.WriteString(fmt.Sprintf("   ✅ GPU Driver: Version %s (PASSED)\n", driver.DriverVersion))
			} else if driver.Status == "WARN" {
				// Count warnings as passed but note them
				passedTests++
				output.WriteString(fmt.Sprintf("   ⚠️ GPU Driver: Version %s (WARNING - unsupported)\n", driver.DriverVersion))
			} else {
				failedTests++
				output.WriteString(fmt.Sprintf("   ❌ GPU Driver: Version %s (FAILED)\n", driver.DriverVersion))
			}
		}
		output.WriteString("\n")
	}

	// GPU Clock Check Tests
	if len(report.Localhost.GPUClockCheck) > 0 {
		output.WriteString("⏱️ GPU Clock Speed Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, clock := range report.Localhost.GPUClockCheck {
			totalTests++
			if clock.Status == "PASS" {
				passedTests++
				if clock.Message != "" {
					output.WriteString(fmt.Sprintf("   ✅ GPU Clock Speeds: %s (PASSED)\n", clock.Message))
				} else {
					output.WriteString("   ✅ GPU Clock Speeds: All GPUs running at acceptable speeds (PASSED)\n")
				}
			} else {
				failedTests++
				if clock.Message != "" {
					output.WriteString(fmt.Sprintf("   ❌ GPU Clock Speeds: %s (FAILED)\n", clock.Message))
				} else {
					output.WriteString("   ❌ GPU Clock Speeds: Some GPUs below acceptable speed threshold (FAILED)\n")
				}
			}
		}
		output.WriteString("\n")
	}

	// PeerMem Module Tests
	if len(report.Localhost.PeerMemModuleCheck) > 0 {
		output.WriteString("🔧 PeerMem Module Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, peerMem := range report.Localhost.PeerMemModuleCheck {
			totalTests++
			if peerMem.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ PeerMem Module: nvidia_peermem loaded (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ PeerMem Module: nvidia_peermem not loaded (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// NVLink Tests
	if len(report.Localhost.NVLinkSpeedCheck) > 0 {
		output.WriteString("🔗 NVLink Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, nvlink := range report.Localhost.NVLinkSpeedCheck {
			totalTests++
			if nvlink.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ NVLink: All links meet speed and count requirements (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ NVLink: Speed or count issues detected (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// Eth0 Presence Tests
	if len(report.Localhost.Eth0PresenceCheck) > 0 {
		output.WriteString("🌐 Eth0 Interface Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, eth0 := range report.Localhost.Eth0PresenceCheck {
			totalTests++
			if eth0.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ Eth0 Interface: eth0 is present (PASSED)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ Eth0 Interface: eth0 is missing (FAILED)\n")
			}
		}
		output.WriteString("\n")
	}

	// CDFP Cable Check Tests
	if len(report.Localhost.CDFPCableCheck) > 0 {
		output.WriteString("🔌 CDFP Cable Health Check\n")
		output.WriteString("   " + strings.Repeat("-", 30) + "\n")
		for _, cdfp := range report.Localhost.CDFPCableCheck {
			totalTests++
			if cdfp.Status == "PASS" {
				passedTests++
				output.WriteString("   ✅ CDFP Cables: All GPU-to-module mappings correct (PASSED)\n")
			} else if cdfp.Status == "SKIP" {
				// Count skipped tests as neither passed nor failed
				totalTests-- // Adjust total count as SKIP doesn't count
				output.WriteString("   ⏭️ CDFP Cables: Check skipped (not applicable for this shape)\n")
			} else {
				failedTests++
				output.WriteString("   ❌ CDFP Cables: GPU-to-module mapping issues detected (FAILED)\n")
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
