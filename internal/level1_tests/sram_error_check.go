// SRAM error checks for GPUs, ensuring that uncorrectable and correctable errors are within acceptable
// limits. The SRAM error check is critical for maintaining the reliability of GPU memory in
// high-performance computing environments.

package level1_tests

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// SRAMErrorCounts represents SRAM error counts for a single GPU
type SRAMErrorCounts struct {
	GPUIndex      int `json:"gpu_index"`
	Uncorrectable int `json:"uncorrectable"`
	Correctable   int `json:"correctable"`
	ParityErrors  int `json:"parity_errors"`
	SECDEDErrors  int `json:"sec_ded_errors"`
}

// SRAMCheckResult represents the overall SRAM check result
type SRAMCheckResult struct {
	Status     string            `json:"status"`
	GPUResults []SRAMErrorCounts `json:"gpu_results"`
	Summary    SRAMErrorSummary  `json:"summary"`
}

// SRAMErrorSummary provides aggregate statistics
type SRAMErrorSummary struct {
	TotalGPUs             int `json:"total_gpus"`
	GPUsWithUncorrectable int `json:"gpus_with_uncorrectable"`
	GPUsWithCorrectable   int `json:"gpus_with_correctable"`
	MaxUncorrectable      int `json:"max_uncorrectable"`
	MaxCorrectable        int `json:"max_correctable"`
}

// SRAMCheckTestConfig represents the test configuration for SRAM check
type SRAMCheckTestConfig struct {
	IsEnabled              bool `json:"enabled"`
	UncorrectableThreshold int  `json:"uncorrectable_threshold"`
	CorrectableThreshold   int  `json:"correctable_threshold"`
}

// getSRAMCheckTestConfig gets test config needed to run this test
func getSRAMCheckTestConfig(shape string) (*SRAMCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Default configuration
	sramErrorCheckTestConfig := &SRAMCheckTestConfig{
		IsEnabled:              false,
		UncorrectableThreshold: 5,    // Default: any uncorrectable errors are critical
		CorrectableThreshold:   1000, // Default: high correctable errors suggest degradation
	}

	enabled, err := limits.IsTestEnabled(shape, "sram_error_check")
	if err != nil {
		return nil, err
	}
	sramErrorCheckTestConfig.IsEnabled = enabled

	if !enabled {
		return sramErrorCheckTestConfig, nil
	}

	threshold, err := limits.GetThresholdForTest(shape, "sram_error_check")
	if err != nil {
		return nil, err
	}

	// Handle different threshold formats
	switch v := threshold.(type) {
	case map[string]interface{}:
		// Handle structured threshold: {"uncorrectable": 10, "correctable": 100}
		if uncorrectable, ok := v["uncorrectable"]; ok {
			if val, ok := uncorrectable.(float64); ok {
				sramErrorCheckTestConfig.UncorrectableThreshold = int(val)
			} else if val, ok := uncorrectable.(int); ok {
				sramErrorCheckTestConfig.UncorrectableThreshold = val
			}
		}
		if correctable, ok := v["correctable"]; ok {
			if val, ok := correctable.(float64); ok {
				sramErrorCheckTestConfig.CorrectableThreshold = int(val)
			} else if val, ok := correctable.(int); ok {
				sramErrorCheckTestConfig.CorrectableThreshold = val
			}
		}
	case float64:
		// Single threshold value applies to uncorrectable errors
		sramErrorCheckTestConfig.UncorrectableThreshold = int(v)
	case int:
		sramErrorCheckTestConfig.UncorrectableThreshold = v
	default:
		logger.Info("Unexpected threshold format for sram_error_check, using defaults")
	}

	return sramErrorCheckTestConfig, nil
}

// parseSRAMResults parses nvidia-smi output to extract SRAM error counts
func parseSRAMResults(uncorrectableOutput, correctableOutput string) ([]SRAMErrorCounts, error) {
	var results []SRAMErrorCounts

	if len(uncorrectableOutput) == 0 && len(correctableOutput) == 0 {
		return results, fmt.Errorf("no SRAM error output to parse")
	}

	// Parse uncorrectable errors
	uncorrectableLines := strings.Split(uncorrectableOutput, "\n")
	correctableLines := strings.Split(correctableOutput, "\n")

	// Extract error counts from nvidia-smi output
	uncorrectableCounts := extractErrorCounts(uncorrectableLines)
	correctableCounts := extractErrorCounts(correctableLines)

	// Determine the number of GPUs (use the maximum from both lists)
	numGPUs := len(uncorrectableCounts)
	if len(correctableCounts) > numGPUs {
		numGPUs = len(correctableCounts)
	}

	// Create results for each GPU
	for i := 0; i < numGPUs; i++ {
		result := SRAMErrorCounts{
			GPUIndex: i,
		}

		// Get uncorrectable errors for this GPU
		if i < len(uncorrectableCounts) {
			result.Uncorrectable = uncorrectableCounts[i].Total
			result.ParityErrors = uncorrectableCounts[i].Parity
			result.SECDEDErrors = uncorrectableCounts[i].SECDED
		}

		// Get correctable errors for this GPU
		if i < len(correctableCounts) {
			result.Correctable = correctableCounts[i].Total
		}

		results = append(results, result)
	}

	return results, nil
}

// GPUErrorDetails holds detailed error information for a GPU
type GPUErrorDetails struct {
	Total  int
	Parity int
	SECDED int
}

// extractErrorCounts extracts error counts from nvidia-smi grep output
func extractErrorCounts(lines []string) []GPUErrorDetails {
	var gpuErrors []GPUErrorDetails
	currentGPU := GPUErrorDetails{}

	// Regex patterns to match different error types
	parityRegex := regexp.MustCompile(`Parity\s*:\s*(\d+)`)
	secdedRegex := regexp.MustCompile(`SEC-DED\s*:\s*(\d+)`)
	generalRegex := regexp.MustCompile(`:\s*(\d+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for Parity errors
		if matches := parityRegex.FindStringSubmatch(line); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				currentGPU.Parity = val
			}
		} else if matches := secdedRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Check for SEC-DED errors
			if val, err := strconv.Atoi(matches[1]); err == nil {
				currentGPU.SECDED = val
				// For uncorrectable errors, total = SEC-DED + Parity
				currentGPU.Total = currentGPU.SECDED + currentGPU.Parity
				gpuErrors = append(gpuErrors, currentGPU)
				currentGPU = GPUErrorDetails{} // Reset for next GPU
			}
		} else if matches := generalRegex.FindStringSubmatch(line); len(matches) > 1 {
			// General error count (for correctable errors)
			if val, err := strconv.Atoi(matches[1]); err == nil {
				currentGPU.Total = val
				gpuErrors = append(gpuErrors, currentGPU)
				currentGPU = GPUErrorDetails{} // Reset for next GPU
			}
		}
	}

	return gpuErrors
}

// checkSRAMThresholds validates SRAM error counts against thresholds
func checkSRAMThresholds(results []SRAMErrorCounts, config *SRAMCheckTestConfig) (string, SRAMErrorSummary) {
	if len(results) == 0 {
		return "FAIL", SRAMErrorSummary{}
	}

	summary := SRAMErrorSummary{
		TotalGPUs: len(results),
	}

	status := "PASS"
	hasWarning := false

	for _, result := range results {
		// Check uncorrectable errors (critical)
		if result.Uncorrectable > config.UncorrectableThreshold {
			status = "FAIL"
			summary.GPUsWithUncorrectable++
		}

		// Check correctable errors (warning)
		if result.Correctable > config.CorrectableThreshold {
			hasWarning = true
			summary.GPUsWithCorrectable++
		}

		// Track maximums
		if result.Uncorrectable > summary.MaxUncorrectable {
			summary.MaxUncorrectable = result.Uncorrectable
		}
		if result.Correctable > summary.MaxCorrectable {
			summary.MaxCorrectable = result.Correctable
		}
	}

	// If no failures but warnings exist, set status to WARN
	if status == "PASS" && hasWarning {
		status = "WARN"
	}

	return status, summary
}

// RunSRAMCheck performs the SRAM error check
func RunSRAMCheck() error {
	logger.Info("=== SRAM Error Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("SRAM Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddSRAMErrorResult("FAIL", 0, 0, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	sramErrorCheckTestConfig, err := getSRAMCheckTestConfig(shape)
	if err != nil {
		logger.Error("SRAM Check: FAIL - Could not get test configuration:", err)
		rep.AddSRAMErrorResult("FAIL", 0, 0, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !sramErrorCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Get SRAM error information from nvidia-smi
	logger.Info("Step 2: Getting SRAM error information from nvidia-smi...")
	logger.Info("Uncorrectable threshold:", sramErrorCheckTestConfig.UncorrectableThreshold)
	logger.Info("Correctable threshold:", sramErrorCheckTestConfig.CorrectableThreshold)

	// Execute nvidia-smi commands to get error counts
	uncorrectableOutput, err := executor.RunNvidiaSMIErrorQuery("uncorrectable")
	if err != nil {
		logger.Error("SRAM Check: FAIL - Could not get uncorrectable SRAM errors:", err)
		rep.AddSRAMErrorResult("FAIL", 0, 0, err)
		return fmt.Errorf("failed to get uncorrectable SRAM errors: %w", err)
	}

	correctableOutput, err := executor.RunNvidiaSMIErrorQuery("correctable")
	if err != nil {
		logger.Error("SRAM Check: FAIL - Could not get correctable SRAM errors:", err)
		rep.AddSRAMErrorResult("FAIL", 0, 0, err)
		return fmt.Errorf("failed to get correctable SRAM errors: %w", err)
	}

	// Step 4: Parse the SRAM error results
	logger.Info("Step 3: Parsing SRAM error results...")
	sramResults, err := parseSRAMResults(uncorrectableOutput.Output, correctableOutput.Output)
	if err != nil {
		logger.Error("SRAM Check: FAIL - Could not parse SRAM error results:", err)
		rep.AddSRAMErrorResult("FAIL", 0, 0, err)
		return fmt.Errorf("failed to parse SRAM error results: %w", err)
	}
	logger.Info("Found SRAM data for", len(sramResults), "GPUs")

	// Step 5: Validate SRAM error counts against thresholds
	logger.Info("Step 4: Validating SRAM error counts...")
	status, summary := checkSRAMThresholds(sramResults, sramErrorCheckTestConfig)

	// Step 6: Report results
	if status == "PASS" {
		logger.Info("SRAM Check: PASS - All SRAM error counts within acceptable limits")
		logger.Info("Max uncorrectable errors:", summary.MaxUncorrectable)
		logger.Info("Max correctable errors:", summary.MaxCorrectable)
		rep.AddSRAMErrorResult("PASS", summary.MaxUncorrectable, summary.MaxCorrectable, nil)
		return nil
	} else if status == "WARN" {
		logger.Info("SRAM Check: FAIL - Correctable errors exceed threshold")
		logger.Info("GPUs with excessive correctable errors:", summary.GPUsWithCorrectable)
		logger.Info("Max correctable errors:", summary.MaxCorrectable)
		err = fmt.Errorf("correctable SRAM errors exceed threshold: max=%d, threshold=%d",
			summary.MaxCorrectable, sramErrorCheckTestConfig.CorrectableThreshold)
		// Sending FAIL as the threshold is exceeded
		rep.AddSRAMErrorResult("FAIL", summary.MaxUncorrectable, summary.MaxCorrectable, err)
		return err
	} else {
		logger.Error("SRAM Check: FAIL - Uncorrectable errors exceed threshold")
		logger.Error("GPUs with uncorrectable errors:", summary.GPUsWithUncorrectable)
		logger.Error("Max uncorrectable errors:", summary.MaxUncorrectable)
		err = fmt.Errorf("uncorrectable SRAM errors exceed threshold: max=%d, threshold=%d",
			summary.MaxUncorrectable, sramErrorCheckTestConfig.UncorrectableThreshold)
		rep.AddSRAMErrorResult("FAIL", summary.MaxUncorrectable, summary.MaxCorrectable, err)
		return err
	}
}

// PrintSRAMCheck prints a placeholder message for SRAM check
func PrintSRAMCheck() {
	logger.Info("SRAM Check: Checking GPU SRAM error counts...")
	logger.Info("SRAM Check: Placeholder - checking uncorrectable and correctable errors")
}
