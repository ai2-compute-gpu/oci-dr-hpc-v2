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
	Status       string `json:"status"`
	GPUCount     int    `json:"gpu_count,omitempty"`
	NumRDMANics  int    `json:"num_rdma_nics,omitempty"`
	TimestampUTC string `json:"timestamp_utc"`
}

// HostResults represents test results for a host
type HostResults struct {
	GPUCountCheck  []TestResult `json:"gpu_count_check,omitempty"`
	PCIeErrorCheck []TestResult `json:"pcie_error_check,omitempty"`
	RDMANicsCount  []TestResult `json:"rdma_nics_count,omitempty"`
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
func AnalyzeResults(resultsFile string) error {
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

	// Display recommendations
	displayRecommendations(recommendations)

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

// generateRecommendations analyzes test results and generates recommendations
func generateRecommendations(results HostResults) RecommendationReport {
	var recommendations []Recommendation
	var criticalCount, warningCount, infoCount int

	// Analyze GPU test results
	for _, gpu := range results.GPUCountCheck {
		if gpu.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "gpu_count_check",
				Issue:      fmt.Sprintf("GPU count mismatch detected. Expected count not met (found: %d)", gpu.GPUCount),
				Suggestion: "Verify GPU hardware installation and driver status",
				Commands: []string{
					"nvidia-smi",
					"lspci | grep -i nvidia",
					"dmesg | grep -i nvidia",
					"sudo nvidia-smi -pm 1",
				},
				References: []string{
					"https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/",
					"https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm",
				},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		} else if gpu.Status == "PASS" {
			rec := Recommendation{
				Type:       "info",
				TestName:   "gpu_count_check",
				Issue:      fmt.Sprintf("GPU count check passed (%d GPUs detected)", gpu.GPUCount),
				Suggestion: "GPU hardware is properly detected and configured",
				Commands: []string{
					"nvidia-smi -q",
					"nvidia-smi topo -m",
				},
			}
			recommendations = append(recommendations, rec)
			infoCount++
		}
	}

	// Analyze PCIe test results
	for _, pcie := range results.PCIeErrorCheck {
		if pcie.Status == "FAIL" {
			rec := Recommendation{
				Type:       "critical",
				TestName:   "pcie_error_check",
				Issue:      "PCIe errors detected in system logs",
				Suggestion: "Check PCIe bus health and reseat hardware if necessary",
				Commands: []string{
					"dmesg | grep -i pcie",
					"dmesg | grep -i 'corrected error'",
					"dmesg | grep -i 'uncorrectable error'",
					"lspci -tv",
					"sudo pcieport-error-inject (advanced debugging)",
				},
				References: []string{
					"https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm",
					"https://www.kernel.org/doc/Documentation/PCI/pci-error-recovery.txt",
				},
			}
			recommendations = append(recommendations, rec)
			criticalCount++
		} else if pcie.Status == "PASS" {
			rec := Recommendation{
				Type:       "info",
				TestName:   "pcie_error_check",
				Issue:      "PCIe error check passed",
				Suggestion: "PCIe bus appears healthy with no errors detected",
				Commands: []string{
					"lspci -tv",
					"dmesg | tail -50",
				},
			}
			recommendations = append(recommendations, rec)
			infoCount++
		}
	}

	// Analyze RDMA test results
	for _, rdma := range results.RDMANicsCount {
		if rdma.Status == "FAIL" {
			rec := Recommendation{
				Type:       "warning",
				TestName:   "rdma_nics_count",
				Issue:      fmt.Sprintf("RDMA NIC count mismatch (found: %d)", rdma.NumRDMANics),
				Suggestion: "Verify RDMA hardware installation and driver configuration",
				Commands: []string{
					"ibstat",
					"ibv_devices",
					"lspci | grep -i mellanox",
					"rdma link show",
					"systemctl status openibd",
				},
				References: []string{
					"https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/configuringrdma.htm",
					"https://docs.mellanox.com/display/MLNXOFEDv461000/",
				},
			}
			recommendations = append(recommendations, rec)
			warningCount++
		} else if rdma.Status == "PASS" {
			rec := Recommendation{
				Type:       "info",
				TestName:   "rdma_nics_count",
				Issue:      fmt.Sprintf("RDMA NIC count check passed (%d NICs detected)", rdma.NumRDMANics),
				Suggestion: "RDMA hardware is properly detected and configured",
				Commands: []string{
					"ibstat",
					"ibv_devinfo",
					"rdma link show",
				},
			}
			recommendations = append(recommendations, rec)
			infoCount++
		}
	}

	// Generate summary
	totalIssues := criticalCount + warningCount
	var summary string
	if totalIssues == 0 {
		summary = "🎉 All diagnostic tests passed! Your HPC environment appears healthy."
	} else {
		summary = fmt.Sprintf("⚠️ Found %d issue(s) requiring attention: %d critical, %d warning",
			totalIssues, criticalCount, warningCount)
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

// displayRecommendations displays the recommendations in a user-friendly format
func displayRecommendations(report RecommendationReport) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🔍 HPC DIAGNOSTIC RECOMMENDATIONS")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("\n📊 SUMMARY: %s\n", report.Summary)
	fmt.Printf("   • Total Issues: %d\n", report.TotalIssues)
	fmt.Printf("   • Critical: %d\n", report.CriticalIssues)
	fmt.Printf("   • Warning: %d\n", report.WarningIssues)
	fmt.Printf("   • Info: %d\n", report.InfoIssues)

	if len(report.Recommendations) == 0 {
		fmt.Println("\n✅ No recommendations needed. System appears healthy!")
		return
	}

	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("📋 DETAILED RECOMMENDATIONS")
	fmt.Println(strings.Repeat("-", 70))

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

		fmt.Printf("\n%s %d. %s [%s]\n", icon, i+1, strings.ToUpper(rec.Type), rec.TestName)
		fmt.Printf("   Issue: %s\n", rec.Issue)
		fmt.Printf("   Suggestion: %s\n", rec.Suggestion)

		if len(rec.Commands) > 0 {
			fmt.Printf("   Commands to run:\n")
			for _, cmd := range rec.Commands {
				fmt.Printf("     $ %s\n", cmd)
			}
		}

		if len(rec.References) > 0 {
			fmt.Printf("   References:\n")
			for _, ref := range rec.References {
				fmt.Printf("     - %s\n", ref)
			}
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 70))
	fmt.Printf("\nGenerated at: %s", report.GeneratedAt)
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
}

// PrintHello prints a simple hello message (keeping for backward compatibility)
func PrintHello() {
	fmt.Println("Hello, World from recommender!")
}
