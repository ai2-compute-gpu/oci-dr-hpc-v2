package recommender

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// TestResult represents a single test result from the reporter
type TestResult struct {
	Status            string      `json:"status"`
	GPUCount          int         `json:"gpu_count,omitempty"`
	Message           string      `json:"message,omitempty"`
	EnabledGPUIndexes []string    `json:"enabled_gpu_indexes,omitempty"`
	NumRDMANics       int         `json:"num_rdma_nics,omitempty"`
	FailedCount       int         `json:"failed_count,omitempty"`
	FailedInterfaces  string      `json:"failed_interfaces,omitempty"`
	InterfaceCount    int         `json:"interface_count,omitempty"`
	InvalidGIDIndexes []int       `json:"invalid_gid_indexes,omitempty"`
	Interfaces        interface{} `json:"interfaces,omitempty"`
	MaxUncorrectable  int         `json:"max_uncorrectable,omitempty"`
	MaxCorrectable    int         `json:"max_correctable,omitempty"`
	ModuleLoaded      bool        `json:"module_loaded,omitempty"`
	NVLinks           interface{} `json:"nvlinks,omitempty"`
	Eth0Present       bool        `json:"eth0_present,omitempty"`
	TimestampUTC      string      `json:"timestamp_utc"`
}

// HostResults represents test results for a host
type HostResults struct {
	GPUCountCheck      []TestResult `json:"gpu_count_check,omitempty"`
	GPUModeCheck       []TestResult `json:"gpu_mode_check,omitempty"`
	PCIeErrorCheck     []TestResult `json:"pcie_error_check,omitempty"`
	RDMANicsCount      []TestResult `json:"rdma_nics_count,omitempty"`
	RxDiscardsCheck    []TestResult `json:"rx_discards_check,omitempty"`
	GIDIndexCheck      []TestResult `json:"gid_index_check,omitempty"`
	LinkCheck          []TestResult `json:"link_check,omitempty"`
	EthLinkCheck       []TestResult `json:"eth_link_check,omitempty"`
	AuthCheck          []TestResult `json:"auth_check,omitempty"`
	SRAMErrorCheck     []TestResult `json:"sram_error_check,omitempty"`
	GPUDriverCheck     []TestResult `json:"gpu_driver_check,omitempty"`
	PeerMemModuleCheck []TestResult `json:"peermem_module_check,omitempty"`
	NVLinkSpeedCheck   []TestResult `json:"nvlink_speed_check,omitempty"`
	Eth0PresenceCheck  []TestResult `json:"eth0_presence_check,omitempty"`
	GPUClkCheck        []TestResult `json:"gpu_clk_check,omitempty"`
	HCAErrorCheck      []TestResult `json:"hca_error_check,omitempty"`
}

// ReportOutput represents the single report format
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

// Recommendation represents a single recommendation
type Recommendation struct {
	Type       string   `json:"type"` // "critical", "warning", "info"
	TestName   string   `json:"test_name"`
	FaultCode  string   `json:"fault_code,omitempty"`
	Issue      string   `json:"issue"`
	Suggestion string   `json:"suggestion"`
	Commands   []string `json:"commands,omitempty"`
	References []string `json:"references,omitempty"`
}

// RecommendationReport represents the final recommendations
type RecommendationReport struct {
	Summary         string           `json:"summary"`
	TotalIssues     int              `json:"total_issues"`
	CriticalIssues  int              `json:"critical_issues"`
	WarningIssues   int              `json:"warning_issues"`
	InfoIssues      int              `json:"info_issues"`
	Recommendations []Recommendation `json:"recommendations"`
	GeneratedAt     string           `json:"generated_at"`
}

// AnalyzeResults analyzes test results and provides recommendations
func AnalyzeResults(resultsFile, outputFormat string) error {
	logger.Info(fmt.Sprintf("Analyzing results file: %s", resultsFile))

	// Read the results file
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		return fmt.Errorf("failed to read results file: %w", err)
	}

	// Parse the results
	hostResults, err := parseResults(data)
	if err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	// Generate recommendations
	recommendations := generateRecommendations(hostResults)

	// Format and display recommendations based on output format
	if err := outputRecommendations(recommendations, outputFormat); err != nil {
		return fmt.Errorf("failed to output recommendations: %w", err)
	}

	return nil
}

// parseResults parses the JSON results file and returns the latest test results
func parseResults(data []byte) (HostResults, error) {
	var hostResults HostResults

	// Try to parse as AppendedReport first
	var appendedReport AppendedReport
	if err := json.Unmarshal(data, &appendedReport); err == nil && len(appendedReport.TestRuns) > 0 {
		// Use the latest test run
		latestRun := appendedReport.TestRuns[len(appendedReport.TestRuns)-1]
		hostResults = latestRun.TestResults
		logger.Info(fmt.Sprintf("Found %d test runs, analyzing latest run: %s", len(appendedReport.TestRuns), latestRun.RunID))
	} else {
		// Try to parse as single ReportOutput
		var singleReport ReportOutput
		if err := json.Unmarshal(data, &singleReport); err != nil {
			return hostResults, fmt.Errorf("failed to parse as either appended or single report format: %w", err)
		}
		hostResults = singleReport.Localhost
		logger.Info("Analyzing single report format")
	}

	return hostResults, nil
}

// generateRecommendations analyzes test results and generates recommendations using config
func generateRecommendations(results HostResults) RecommendationReport {
	// Load recommendation configuration
	config, err := LoadRecommendationConfig()
	if err != nil {
		logger.Errorf("Failed to load recommendation config: %v", err)
		return generateFallbackRecommendations(results)
	}

	var recommendations []Recommendation
	var criticalCount, warningCount, infoCount int

	// Process all test types using config
	testMappings := []struct {
		testName string
		results  []TestResult
	}{
		{"gpu_count_check", results.GPUCountCheck},
		{"gpu_mode_check", results.GPUModeCheck},
		{"pcie_error_check", results.PCIeErrorCheck},
		{"rdma_nics_count", results.RDMANicsCount},
		{"rx_discards_check", results.RxDiscardsCheck},
		{"gid_index_check", results.GIDIndexCheck},
		{"link_check", results.LinkCheck},
		{"eth_link_check", results.EthLinkCheck},
		{"auth_check", results.AuthCheck},
		{"sram_error_check", results.SRAMErrorCheck},
		{"gpu_driver_check", results.GPUDriverCheck},
		{"peermem_module_check", results.PeerMemModuleCheck},
		{"nvlink_speed_check", results.NVLinkSpeedCheck},
		{"eth0_presence_check", results.Eth0PresenceCheck},
		{"gpu_clk_check", results.GPUClkCheck},
		{"hca_error_check", results.HCAErrorCheck},
	}

	for _, mapping := range testMappings {
		for _, testResult := range mapping.results {
			if rec := config.GetRecommendation(mapping.testName, testResult.Status, testResult); rec != nil {
				recommendations = append(recommendations, *rec)

				// Count by type
				switch rec.Type {
				case "critical":
					criticalCount++
				case "warning":
					warningCount++
				case "info":
					infoCount++
				}
			}
		}
	}

	// Generate summary using config
	totalIssues := criticalCount + warningCount
	summary := config.GetSummary(totalIssues, criticalCount, warningCount)

	return RecommendationReport{
		Summary:         summary,
		TotalIssues:     totalIssues,
		CriticalIssues:  criticalCount,
		WarningIssues:   warningCount,
		InfoIssues:      infoCount,
		Recommendations: recommendations,
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
	}
}

// generateFallbackRecommendations provides basic recommendations when config loading fails
func generateFallbackRecommendations(results HostResults) RecommendationReport {
	logger.Info("Using fallback recommendations due to config load failure")

	var recommendations []Recommendation
	var criticalCount, warningCount, infoCount int

	// Basic GPU recommendations
	for _, gpu := range results.GPUCountCheck {
		if gpu.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "gpu_count_check",
				Issue:      fmt.Sprintf("GPU count mismatch detected (found: %d)", gpu.GPUCount),
				Suggestion: "Verify GPU hardware installation and driver status",
				Commands:   []string{"nvidia-smi", "lspci | grep -i nvidia"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic GPU Mode recommendations
	for _, gpuMode := range results.GPUModeCheck {
		if gpuMode.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "gpu_mode_check",
				Issue:      fmt.Sprintf("GPU MIG mode configuration violation detected on GPUs: %v", gpuMode.EnabledGPUIndexes),
				Suggestion: "Disable MIG mode on affected GPUs or verify that MIG configuration meets workload requirements",
				Commands:   []string{"nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader", "sudo nvidia-smi -mig 0"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic PCIe recommendations
	for _, pcie := range results.PCIeErrorCheck {
		if pcie.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "pcie_error_check",
				Issue:      "PCIe errors detected in system logs",
				Suggestion: "Check PCIe bus health and reseat hardware if necessary",
				Commands:   []string{"dmesg | grep -i pcie", "lspci -tv"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic RDMA recommendations
	for _, rdma := range results.RDMANicsCount {
		if rdma.Status == "FAIL" {
			rec := Recommendation{
				Type:       "warning",
				TestName:   "rdma_nics_count",
				Issue:      fmt.Sprintf("RDMA NIC count mismatch (found: %d)", rdma.NumRDMANics),
				Suggestion: "Verify RDMA hardware installation and driver configuration",
				Commands:   []string{"ibstat", "ibv_devices"},
			}
			recommendations = append(recommendations, rec)
			warningCount++
		}
	}

	// Basic RX Discards recommendations
	for _, rxDiscard := range results.RxDiscardsCheck {
		if rxDiscard.Status == "FAIL" {
			commands := []string{}
			for _, failedInterface := range strings.Split(rxDiscard.FailedInterfaces, ",") {
				commands = append(commands, fmt.Sprintf("sudo ethtool -S %s | grep rx_prio.*_discards", failedInterface))
			}
			rec := Recommendation{
				Type:       "warning",
				TestName:   "rx_discards_check",
				Issue:      fmt.Sprintf("RX discards exceeded the specified threshold for %s", rxDiscard.FailedInterfaces),
				Suggestion: "Verify Rx discard for the failed interface",
				Commands:   commands,
			}
			recommendations = append(recommendations, rec)
			warningCount++
		}
	}

	// Basic GID Index recommendations
	for _, gidIndex := range results.GIDIndexCheck {
		if gidIndex.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "gid_index_check",
				Issue:      fmt.Sprintf("GID index check failed (invalid indexes: %v)", gidIndex.InvalidGIDIndexes),
				Suggestion: "Verify RDMA GID configuration and check for interface issues",
				Commands:   []string{"show_gids", "ibstat", "rdma link show"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic Link Check recommendations
	for _, linkCheck := range results.LinkCheck {
		if linkCheck.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "link_check",
				Issue:      "RDMA link check failed - link parameters do not meet expected values",
				Suggestion: "Check RDMA link health, verify cable connections, and inspect link parameters",
				Commands:   []string{"ibstat", "rdma link show", "sudo mlxlink -d mlx5_0 --show_module"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic Ethernet Link Check recommendations
	for _, ethLinkCheck := range results.EthLinkCheck {
		if ethLinkCheck.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "eth_link_check",
				Issue:      "Ethernet link check failed - link parameters do not meet expected values",
				Suggestion: "Check Ethernet link health, verify cable connections, and inspect link parameters for 100GbE RoCE interfaces",
				Commands:   []string{"sudo ibdev2netdev", "ip link show", "sudo mst status -v"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic SRAM recommendations
	for _, sram := range results.SRAMErrorCheck {
		if sram.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "sram_error_check",
				Issue:      fmt.Sprintf("SRAM uncorrectable errors detected (max: %d)", sram.MaxUncorrectable),
				Suggestion: "Check GPU memory health and consider replacing affected hardware",
				Commands: []string{
					"sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable",
					"sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable",
				},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		} else if sram.Status == "WARN" {
			rec := Recommendation{
				Type:       "warning",
				TestName:   "sram_error_check",
				Issue:      fmt.Sprintf("SRAM correctable errors exceed threshold (max: %d)", sram.MaxCorrectable),
				Suggestion: "Monitor GPU memory health and consider maintenance scheduling",
				Commands: []string{
					"sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable",
					"sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable",
				},
			}
			recommendations = append(recommendations, rec)
			warningCount++
		}
	}

	// Basic GPU Driver Check recommendations
	for _, driverCheck := range results.GPUDriverCheck {
		if driverCheck.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "gpu_driver_check",
				Issue:      "GPU driver version validation failed",
				Suggestion: "Update to a supported GPU driver version or investigate driver installation issues",
				Commands:   []string{"nvidia-smi --query-gpu=driver_version --format=csv,noheader", "sudo apt update && sudo apt install nvidia-driver-535"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		} else if driverCheck.Status == "WARN" {
			rec := Recommendation{
				Type:       "warning",
				TestName:   "gpu_driver_check",
				Issue:      "GPU driver version is unsupported but not blacklisted",
				Suggestion: "Consider updating to a known supported driver version for optimal performance",
				Commands:   []string{"nvidia-smi --query-gpu=driver_version --format=csv,noheader", "nvidia-smi -q"},
			}
			recommendations = append(recommendations, rec)
			warningCount++
		}
	}

	// Basic PeerMem Module recommendations
	for _, peerMem := range results.PeerMemModuleCheck {
		if peerMem.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "peermem_module_check",
				Issue:      "NVIDIA Peer Memory module (nvidia_peermem) is not loaded",
				Suggestion: "Load the nvidia_peermem kernel module to enable GPU peer memory access",
				Commands:   []string{"sudo modprobe nvidia_peermem", "lsmod | grep nvidia_peermem"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic NVLink Speed Check recommendations
	for _, nvlinkCheck := range results.NVLinkSpeedCheck {
		if nvlinkCheck.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "nvlink_speed_check",
				FaultCode:  "HPCGPU-0009-0001",
				Issue:      "NVLink speed or count check failed - GPU interconnect links do not meet expected performance requirements",
				Suggestion: "Check NVLink health, verify GPU interconnect topology, inspect link parameters, and ensure proper GPU seating and cable connections",
				Commands: []string{
					"nvidia-smi nvlink -s",
					"nvidia-smi topo -m",
					"nvidia-smi topo -p2p r",
					"dmesg | grep -i nvlink",
				},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic Eth0 Presence Check recommendations
	for _, eth0Check := range results.Eth0PresenceCheck {
		if eth0Check.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "eth0_presence_check",
				FaultCode:  "HPCGPU-0010-0001",
				Issue:      "eth0 network interface is missing or not detected",
				Suggestion: "Investigate network interface configuration, check if eth0 is properly configured or renamed, and verify network drivers are loaded",
				Commands: []string{
					"ip addr show",
					"ip link show",
					"dmesg | grep -i 'eth0\\|network'",
					"lspci | grep -i ethernet",
				},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	// Basic HCA Error Check recommendations
	for _, hcaCheck := range results.HCAErrorCheck {
		if hcaCheck.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "hca_error_check",
				FaultCode:  "HPCGPU-0011-0001",
				Issue:      "Fatal MLX5 errors were detected in the system logs",
				Suggestion: "Clear dmesg and reboot the node. If the problem persists, return the node to OCI",
				Commands:   []string{"dmesg -T | grep -i mlx5 | grep -i fatal"},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		}
	}

	totalIssues := criticalCount + warningCount
	summary := fmt.Sprintf("Found %d issue(s) requiring attention: %d critical, %d warning (fallback mode)",
		totalIssues, criticalCount, warningCount)

	if totalIssues == 0 {
		summary = "All diagnostic tests passed! Your HPC environment appears healthy. (fallback mode)"
	}

	return RecommendationReport{
		Summary:         summary,
		TotalIssues:     totalIssues,
		CriticalIssues:  criticalCount,
		WarningIssues:   warningCount,
		InfoIssues:      infoCount,
		Recommendations: recommendations,
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
	}
}

// outputRecommendations outputs recommendations in the specified format
func outputRecommendations(report RecommendationReport, outputFormat string) error {
	var output string
	var err error

	switch outputFormat {
	case "json":
		output, err = formatRecommendationsJSON(report)
	case "table":
		output, err = formatRecommendationsTable(report)
	case "friendly":
		output, err = formatRecommendationsFriendly(report)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// formatRecommendationsJSON formats recommendations as JSON
func formatRecommendationsJSON(report RecommendationReport) (string, error) {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData) + "\n", nil
}

// formatRecommendationsTable formats recommendations as a table
func formatRecommendationsTable(report RecommendationReport) (string, error) {
	var output strings.Builder

	output.WriteString("┌─────────────────────────────────────────────────────────────────┐\n")
	output.WriteString("│                    HPC DIAGNOSTIC RECOMMENDATIONS               │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// Summary section
	output.WriteString("│ SUMMARY                                                         │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Total Issues: %d", report.TotalIssues)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Critical: %d", report.CriticalIssues)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Warning: %d", report.WarningIssues)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Info: %d", report.InfoIssues)))
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// Recommendations section
	output.WriteString("│ RECOMMENDATIONS                                                 │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	if len(report.Recommendations) == 0 {
		output.WriteString("│ No recommendations needed. System appears healthy!             │\n")
	} else {
		for i, rec := range report.Recommendations {
			// Truncate long text to fit in table
			issue := rec.Issue
			if len(issue) > 53 {
				issue = issue[:50] + "..."
			}
			suggestion := rec.Suggestion
			if len(suggestion) > 48 {
				suggestion = suggestion[:45] + "..."
			}

			// Calculate remaining space for test name after number, type, and brackets
			typeStr := strings.ToUpper(rec.Type)
			prefixLen := len(fmt.Sprintf(" %d. [%s] ", i+1, typeStr))
			testNameSpace := 64 - prefixLen
			if testNameSpace < 0 {
				testNameSpace = 10 // minimum space
			}
			output.WriteString(fmt.Sprintf("│ %d. [%s] %-*s │\n", i+1, typeStr, testNameSpace, rec.TestName))
			if rec.FaultCode != "" {
				output.WriteString(fmt.Sprintf("│    Fault Code: %-48s │\n", rec.FaultCode))
			}
			output.WriteString(fmt.Sprintf("│    Issue: %-53s │\n", issue))
			output.WriteString(fmt.Sprintf("│    Suggestion: %-48s │\n", suggestion))

			if len(rec.Commands) > 0 && len(rec.Commands[0]) <= 51 {
				output.WriteString(fmt.Sprintf("│    Command: %-51s │\n", rec.Commands[0]))
			}

			if i < len(report.Recommendations)-1 {
				output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
			}
		}
	}

	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString(fmt.Sprintf("│ Generated at: %-49s │\n", report.GeneratedAt))
	output.WriteString("└─────────────────────────────────────────────────────────────────┘\n")

	return output.String(), nil
}

// formatRecommendationsFriendly formats recommendations in a user-friendly format
func formatRecommendationsFriendly(report RecommendationReport) (string, error) {
	var output strings.Builder

	output.WriteString("\n" + strings.Repeat("=", 70) + "\n")
	output.WriteString("🔍 HPC DIAGNOSTIC RECOMMENDATIONS\n")
	output.WriteString(strings.Repeat("=", 70) + "\n")

	output.WriteString(fmt.Sprintf("\n📊 SUMMARY: %s\n", report.Summary))
	output.WriteString(fmt.Sprintf("   • Total Issues: %d\n", report.TotalIssues))
	output.WriteString(fmt.Sprintf("   • Critical: %d\n", report.CriticalIssues))
	output.WriteString(fmt.Sprintf("   • Warning: %d\n", report.WarningIssues))
	output.WriteString(fmt.Sprintf("   • Info: %d\n", report.InfoIssues))

	if len(report.Recommendations) == 0 {
		output.WriteString("\n✅ No recommendations needed. System appears healthy!\n")
		return output.String(), nil
	}

	output.WriteString("\n" + strings.Repeat("-", 70) + "\n")
	output.WriteString("📋 DETAILED RECOMMENDATIONS\n")
	output.WriteString(strings.Repeat("-", 70) + "\n")

	for i, rec := range report.Recommendations {
		var icon string
		switch rec.Type {
		case "critical":
			icon = "🚨"
		case "warning":
			icon = "⚠️"
		case "info":
			icon = "ℹ️"
		default:
			icon = "•"
		}

		output.WriteString(fmt.Sprintf("\n%s %d. %s [%s]\n", icon, i+1, strings.ToUpper(rec.Type), rec.TestName))
		if rec.FaultCode != "" {
			output.WriteString(fmt.Sprintf("   Fault Code: %s\n", rec.FaultCode))
		}
		output.WriteString(fmt.Sprintf("   Issue: %s\n", rec.Issue))
		output.WriteString(fmt.Sprintf("   Suggestion: %s\n", rec.Suggestion))

		if len(rec.Commands) > 0 {
			output.WriteString("   Commands to run:\n")
			for _, cmd := range rec.Commands {
				output.WriteString(fmt.Sprintf("     $ %s\n", cmd))
			}
		}

		if len(rec.References) > 0 {
			output.WriteString("   References:\n")
			for _, ref := range rec.References {
				output.WriteString(fmt.Sprintf("     - %s\n", ref))
			}
		}
	}

	output.WriteString("\n" + strings.Repeat("=", 70) + "\n")
	output.WriteString(fmt.Sprintf("Generated at: %s\n", report.GeneratedAt))
	output.WriteString(strings.Repeat("=", 70) + "\n")

	return output.String(), nil
}

// PrintHello prints a simple hello message (keeping for backward compatibility)
func PrintHello() {
	fmt.Println("Hello, World from recommender!")
}
