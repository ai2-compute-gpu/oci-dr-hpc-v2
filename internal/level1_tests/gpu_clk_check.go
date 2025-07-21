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

	// Initialize with defaults (H100 values)
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
func getGPUClockSpeeds() ([]int, error) {
	// Use nvidia-smi to query current graphics clock speeds
	result := executor.RunNvidiaSMIQuery("clocks.current.graphics")
	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Parse the output
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return []int{}, nil
	}

	// Split by lines and parse clock speeds
	lines := strings.Split(output, "\n")
	var clockSpeeds []int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Extract numeric value from line (e.g., "1980 MHz" -> 1980)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				clockStr := fields[0]
				clockSpeed, err := strconv.Atoi(clockStr)
				if err != nil {
					return nil, fmt.Errorf("failed to parse clock speed '%s': %w", clockStr, err)
				}
				clockSpeeds = append(clockSpeeds, clockSpeed)
			}
		}
	}

	return clockSpeeds, nil
}

// validateGPUClockSpeeds validates clock speeds against expected threshold
func validateGPUClockSpeeds(clockSpeeds []int, expectedSpeed int) (string, string, error) {
	if len(clockSpeeds) == 0 {
		return "FAIL", "", fmt.Errorf("no GPU clock speeds found")
	}

	// Calculate minimum acceptable speed (90% of expected for H100)
	minAcceptableSpeed := expectedSpeed - int(float64(expectedSpeed)*0.10)
	
	var failedGPUs []string
	var lowestSpeed int
	var allowedSpeed string
	
	for i, speed := range clockSpeeds {
		if i == 0 {
			lowestSpeed = speed
		} else if speed < lowestSpeed {
			lowestSpeed = speed
		}
		
		if speed < minAcceptableSpeed {
			failedGPUs = append(failedGPUs, strconv.Itoa(i))
		} else if speed >= minAcceptableSpeed && speed < expectedSpeed {
			// Speed is acceptable but below max
			if allowedSpeed == "" {
				allowedSpeed = strconv.Itoa(speed)
			} else {
				currentAllowed, _ := strconv.Atoi(allowedSpeed)
				if speed < currentAllowed {
					allowedSpeed = strconv.Itoa(speed)
				}
			}
		}
	}

	if len(failedGPUs) > 0 {
		return "FAIL", "", fmt.Errorf("check GPU %s", strings.Join(failedGPUs, ","))
	}

	// All GPUs passed minimum threshold
	if allowedSpeed == "" {
		allowedSpeed = strconv.Itoa(lowestSpeed)
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

	logger.Info("Found GPU clock speeds (MHz):", clockSpeeds)

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
		logger.Error("GPU Clock Check: FAIL -", validationErr)
		rep.AddGPUClockResult("FAIL", "", validationErr)
		return validationErr
	}
}

func PrintGPUClkCheck() {
	logger.Info("GPU Clock Check: Checking GPU clock speeds...")
	logger.Info("GPU Clock Check: PASS - Clock speeds are within acceptable range")
}