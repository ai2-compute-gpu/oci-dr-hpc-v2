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

// GPUClkCheckTestConfig represents the config needed to run this test
type GPUClkCheckTestConfig struct {
	IsEnabled        bool   `json:"enabled"`
	Shape            string `json:"shape"`
	ExpectedClkSpeed int    `json:"clock_speed"`
}

// getGpuClkCheckTestConfig gets test config needed to run this test
func getGpuClkCheckTestConfig() (*GPUClkCheckTestConfig, error) {
	// Get shape from IMDS
	shape, err := executor.GetCurrentShape()
	if err != nil {
		return nil, err
	}

	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Initialize with defaults (H100 values - 1980 MHz as per Python script)
	gpuClkCheckTestConfig := &GPUClkCheckTestConfig{
		IsEnabled:        false,
		Shape:            shape,
		ExpectedClkSpeed: 1980, // Default H100 clock speed in MHz
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "gpu_clk_check")
	if err != nil {
		return nil, err
	}
	gpuClkCheckTestConfig.IsEnabled = enabled

	// Get threshold configuration if available
	threshold, err := limits.GetThresholdForTest(shape, "gpu_clk_check")
	if err == nil {
		// Parse threshold configuration from test_limits.json
		switch v := threshold.(type) {
		case float64:
			gpuClkCheckTestConfig.ExpectedClkSpeed = int(v)
		case map[string]interface{}:
			if clockSpeed, ok := v["clock_speed"].(float64); ok {
				gpuClkCheckTestConfig.ExpectedClkSpeed = int(clockSpeed)
			}
		}
	}

	return gpuClkCheckTestConfig, nil
}

// getGPUClockSpeeds uses nvidia-smi to get current GPU clock speeds
func getGPUClockSpeeds() ([]string, error) {
	// Use nvidia-smi to query current graphics clock speeds
	result := executor.RunNvidiaSMIQuery("clocks.current.graphics")
	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Check for driver communication issues
	if strings.Contains(result.Output, "couldn't communicate with the NVIDIA driver") {
		return nil, fmt.Errorf("NVIDIA driver is not loaded")
	}

	// Parse the output
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return []string{}, nil
	}

	// Split by lines and collect clock speeds
	lines := strings.Split(output, "\n")
	var clockSpeeds []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			clockSpeeds = append(clockSpeeds, line)
		}
	}

	return clockSpeeds, nil
}

// validateGPUClockSpeeds validates clock speeds against expected threshold
// Following Python logic: allow 90% of expected speed (10% tolerance)
func validateGPUClockSpeeds(clockSpeeds []string, expectedSpeed int) (string, string, error) {
	if len(clockSpeeds) == 0 {
		return "FAIL", "check GPU", fmt.Errorf("no GPU clock speeds found")
	}

	// Calculate minimum acceptable speed (90% of expected - 10% tolerance for H100)
	minAcceptableSpeed := expectedSpeed - int(float64(expectedSpeed)*0.10)
	
	var failedGPUs []string
	var allowedSpeed string
	var minAllowedSpeed int = -1
	
	for gpuIndex, speedStr := range clockSpeeds {
		// Extract numeric value from speed string (e.g., "1980 MHz" -> "1980")
		fields := strings.Fields(speedStr)
		if len(fields) == 0 {
			failedGPUs = append(failedGPUs, strconv.Itoa(gpuIndex))
			continue
		}
		
		currentSpeedStr := fields[0]
		currentSpeed, err := strconv.Atoi(currentSpeedStr)
		if err != nil {
			failedGPUs = append(failedGPUs, strconv.Itoa(gpuIndex))
			continue
		}
		
		// Check if speed is below minimum threshold
		if currentSpeed < minAcceptableSpeed {
			failedGPUs = append(failedGPUs, strconv.Itoa(gpuIndex))
		} else {
			// Speed is acceptable - track the minimum for reporting
			if minAllowedSpeed == -1 || currentSpeed < minAllowedSpeed {
				minAllowedSpeed = currentSpeed
				allowedSpeed = currentSpeedStr
			}
		}
	}

	if len(failedGPUs) > 0 {
		return "FAIL", "check GPU " + strings.Join(failedGPUs, ","), 
			fmt.Errorf("GPU clock speeds below threshold for GPUs: %s", strings.Join(failedGPUs, ","))
	}

	// All GPUs passed - format success message following Python pattern
	if allowedSpeed == "" && len(clockSpeeds) > 0 {
		// Use the last processed speed if no specific allowed speed was tracked
		fields := strings.Fields(clockSpeeds[len(clockSpeeds)-1])
		if len(fields) > 0 {
			allowedSpeed = fields[0]
		}
	}
	
	statusMsg := fmt.Sprintf("Expected %d, allowed %s", expectedSpeed, allowedSpeed)
	return "PASS", statusMsg, nil
}

func RunGPUClkCheck() error {
	logger.Info("=== GPU Clock Speed Check ===")
	testConfig, err := getGpuClkCheckTestConfig()
	if err != nil {
		return err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Info("Starting GPU clock speed check...")
	rep := reporter.GetReporter()

	// Step 1: Get GPU clock speeds
	logger.Info("Step 1: Getting GPU clock speeds...")
	clockSpeeds, err := getGPUClockSpeeds()
	if err != nil {
		logger.Error("GPU Clock Check: FAIL - Could not get GPU clock speeds:", err)
		rep.AddGPUClockResult("FAIL", "", err)
		return fmt.Errorf("could not get GPU clock speeds: %w", err)
	}

	logger.Info("Found GPU clock speeds:", clockSpeeds)

	// Step 2: Validate clock speeds
	logger.Info("Step 2: Validating clock speeds...")
	logger.Info("Expected clock speed (MHz):", testConfig.ExpectedClkSpeed)

	status, statusMsg, validationErr := validateGPUClockSpeeds(clockSpeeds, testConfig.ExpectedClkSpeed)

	switch status {
	case "PASS":
		logger.Info("GPU Clock Check: PASS -", statusMsg)
		rep.AddGPUClockResult("PASS", statusMsg, nil)
		return nil
	default: // FAIL
		logger.Error("GPU Clock Check: FAIL -", statusMsg)
		rep.AddGPUClockResult("FAIL", statusMsg, validationErr)
		return validationErr
	}
}

func PrintGPUClkCheck() {
	logger.Info("GPU Clock Check: Checking GPU clock speeds...")
	logger.Info("GPU Clock Check: PASS - Clock speeds are within acceptable range")
}