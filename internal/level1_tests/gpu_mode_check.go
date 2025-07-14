package level1_tests

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// GPUModeResult represents the result of GPU mode check
type GPUModeResult struct {
	Status            string   `json:"status"`
	Message           string   `json:"message"`
	EnabledGPUIndexes []string `json:"enabled_gpu_indexes,omitempty"`
}

// GPUModeInfo represents information about a single GPU's MIG mode
type GPUModeInfo struct {
	Index string `json:"index"`
	Mode  string `json:"mode"`
}

type GpuModeCheckTestConfig struct {
	IsEnabled    bool     `json:"enabled"`
	AllowedModes []string `json:"allowed_modes"`
}

// Gets test config needed to run this test
func getGpuModeCheckTestConfig(shape string) (*GpuModeCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Result
	gpuModeCheckTestConfig := &GpuModeCheckTestConfig{
		IsEnabled:    false,
		AllowedModes: []string{"Disabled", "N/A"}, // Default allowed modes
	}

	enabled, err := limits.IsTestEnabled(shape, "gpu_mode_check")
	if err != nil {
		return nil, err
	}
	gpuModeCheckTestConfig.IsEnabled = enabled

	// Get threshold configuration for allowed modes
	threshold, err := limits.GetThresholdForTest(shape, "gpu_mode_check")
	if err != nil {
		return nil, err
	}

	if thresholdMap, ok := threshold.(map[string]interface{}); ok {
		if allowedModesInterface, exists := thresholdMap["allowed_modes"]; exists {
			if allowedModesList, ok := allowedModesInterface.([]interface{}); ok {
				allowedModes := make([]string, 0, len(allowedModesList))
				for _, mode := range allowedModesList {
					if modeStr, ok := mode.(string); ok {
						allowedModes = append(allowedModes, modeStr)
					}
				}
				if len(allowedModes) > 0 {
					gpuModeCheckTestConfig.AllowedModes = allowedModes
				}
			}
		}
	}

	return gpuModeCheckTestConfig, nil
}

// getGPUModeInfo uses nvidia-smi to get GPU MIG mode information
func getGPUModeInfo() ([]GPUModeInfo, error) {
	// Use nvidia-smi to query GPU index and MIG mode
	result := executor.RunNvidiaSMIQuery("index,mig.mode.current")
	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Parse the output
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return nil, fmt.Errorf("no GPU mode information returned from nvidia-smi")
	}

	return parseGPUModeInfo(output)
}

// parseGPUModeInfo parses the nvidia-smi output into GPUModeInfo structs
func parseGPUModeInfo(output string) ([]GPUModeInfo, error) {
	var gpuModes []GPUModeInfo

	lines := strings.Split(strings.TrimSpace(output), "\n")
	filteredLines := make([]string, 0)

	// Filter out empty lines
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	for i, line := range filteredLines {
		// Split CSV line: index, mig.mode.current
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			logger.Errorf("Invalid GPU mode info line: %s", line)
			continue
		}

		index := strings.TrimSpace(parts[0])
		mode := strings.TrimSpace(parts[1])

		// Validate that index is numeric
		if _, err := strconv.Atoi(index); err != nil {
			logger.Errorf("Invalid GPU index '%s' on line %d: %v", index, i+1, err)
			return nil, fmt.Errorf("invalid GPU index format: %s", index)
		}

		// Validate that mode is one of the expected values
		if !strings.Contains(mode, "Enabled") && !strings.Contains(mode, "Disabled") && !strings.Contains(mode, "N/A") {
			logger.Errorf("Invalid GPU mode '%s' on line %d", mode, i+1)
			return nil, fmt.Errorf("invalid GPU mode format: %s", mode)
		}

		gpuMode := GPUModeInfo{
			Index: index,
			Mode:  mode,
		}

		gpuModes = append(gpuModes, gpuMode)
		logger.Debugf("Parsed GPU %s: Mode=%s", index, mode)
	}

	return gpuModes, nil
}

// checkGPUModeResults analyzes GPU mode information and returns result
func checkGPUModeResults(gpuModes []GPUModeInfo, allowedModes []string) *GPUModeResult {
	result := &GPUModeResult{
		Status:            "PASS",
		Message:           "PASS",
		EnabledGPUIndexes: []string{},
	}

	enabledGPUs := make([]string, 0)

	// Convert allowed modes to a map for easier lookup (case-insensitive)
	allowedModesMap := make(map[string]bool)
	for _, mode := range allowedModes {
		allowedModesMap[strings.ToUpper(mode)] = true
	}

	for _, gpu := range gpuModes {
		modeUpper := strings.ToUpper(gpu.Mode)

		// Check if the mode is in the allowed list
		if !allowedModesMap[modeUpper] {
			// If mode contains "Enabled", it's definitely not allowed
			if strings.Contains(modeUpper, "ENABLED") {
				enabledGPUs = append(enabledGPUs, gpu.Index)
			} else {
				// For other unexpected modes, also consider them as failures
				enabledGPUs = append(enabledGPUs, gpu.Index)
			}
		}
	}

	if len(enabledGPUs) > 0 {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("FAIL - Invalid GPU modes detected on GPUs %s", strings.Join(enabledGPUs, ","))
		result.EnabledGPUIndexes = enabledGPUs
	}

	return result
}

// RunGPUModeCheck performs the GPU MIG mode check
func RunGPUModeCheck() error {
	logger.Info("=== GPU Mode Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("GPU Mode Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddGPUModeResult("FAIL", fmt.Sprintf("Could not get shape from IMDS: %v", err), []string{}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	gpuModeCheckTestConfig, err := getGpuModeCheckTestConfig(shape)
	if err != nil {
		logger.Error("GPU Mode Check: FAIL - Could not get test configuration:", err)
		rep.AddGPUModeResult("FAIL", fmt.Sprintf("Could not get test configuration: %v", err), []string{}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !gpuModeCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Infof("Allowed GPU modes for shape %s: %v", shape, gpuModeCheckTestConfig.AllowedModes)

	// Step 3: Get GPU mode information from nvidia-smi
	logger.Info("Step 2: Getting GPU mode information from nvidia-smi...")
	gpuModes, err := getGPUModeInfo()
	if err != nil {
		logger.Error("GPU Mode Check: FAIL - Could not get GPU mode information:", err)
		rep.AddGPUModeResult("FAIL", fmt.Sprintf("Could not get GPU mode information: %v", err), []string{}, err)
		return fmt.Errorf("failed to get GPU mode information: %w", err)
	}

	if len(gpuModes) == 0 {
		logger.Error("GPU Mode Check: FAIL - No GPU information returned")
		err = fmt.Errorf("no GPU information returned from nvidia-smi")
		rep.AddGPUModeResult("FAIL", "No GPU information returned", []string{}, err)
		return err
	}

	logger.Infof("Found %d GPUs to check", len(gpuModes))

	// Log GPU mode information for debugging
	for _, gpu := range gpuModes {
		logger.Debugf("GPU %s: Mode=%s", gpu.Index, gpu.Mode)
	}

	// Step 4: Check GPU mode results against allowed modes
	logger.Info("Step 3: Checking GPU MIG mode status against allowed modes...")
	result := checkGPUModeResults(gpuModes, gpuModeCheckTestConfig.AllowedModes)

	// Step 5: Report results
	logger.Info("Step 4: Reporting results...")
	if result.Status == "PASS" {
		logger.Info("GPU Mode Check: PASS - All GPUs have acceptable modes")
		rep.AddGPUModeResult("PASS", result.Message, []string{}, err)
		return nil
	} else {
		logger.Error("GPU Mode Check:", result.Message)
		rep.AddGPUModeResult("FAIL", result.Message, result.EnabledGPUIndexes, nil)
		return fmt.Errorf("GPU mode check failed: %s", result.Message)
	}
}

func PrintGPUModeCheck() {
	logger.Info("GPU Mode Check: PASS - All GPUs have MIG mode disabled")
}
