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

// NVLinkSpeedCheckTestConfig represents the configuration for NVLink speed check
type NVLinkSpeedCheckTestConfig struct {
	IsEnabled     bool    `json:"enabled"`
	ExpectedSpeed float64 `json:"expected_speed"`
	ExpectedCount int     `json:"expected_count"`
}

// NVLinkResult represents the result of NVLink parsing for a single GPU
type NVLinkResult struct {
	GPUID     string       `json:"gpu_id"`
	LinkCount int          `json:"link_count"`
	Links     []NVLinkInfo `json:"links"`
}

// NVLinkInfo represents information about a single NVLink
type NVLinkInfo struct {
	LinkID   int     `json:"link_id"`
	Speed    float64 `json:"speed"`
	IsActive bool    `json:"is_active"`
}

// Gets test config needed to run this test
func getNVLinkSpeedCheckTestConfig(shape string) (*NVLinkSpeedCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Result
	nvlinkSpeedCheckTestConfig := &NVLinkSpeedCheckTestConfig{
		IsEnabled:     false,
		ExpectedSpeed: 0,
		ExpectedCount: 0,
	}

	enabled, err := limits.IsTestEnabled(shape, "nvlink_speed_check")
	if err != nil {
		return nil, err
	}
	nvlinkSpeedCheckTestConfig.IsEnabled = enabled

	threshold, err := limits.GetThresholdForTest(shape, "nvlink_speed_check")
	if err != nil {
		return nil, err
	}

	if thresholdMap, ok := threshold.(map[string]interface{}); ok {
		if speed, exists := thresholdMap["speed"]; exists {
			if speedFloat, ok := speed.(float64); ok {
				nvlinkSpeedCheckTestConfig.ExpectedSpeed = speedFloat
			}
		}
		if count, exists := thresholdMap["count"]; exists {
			if countFloat, ok := count.(float64); ok {
				nvlinkSpeedCheckTestConfig.ExpectedCount = int(countFloat)
			}
		}
	}

	return nvlinkSpeedCheckTestConfig, nil
}

// parseNVLinkOutput parses the nvidia-smi nvlink -s output
func parseNVLinkOutput(output string, expectedSpeed float64) (map[string]*NVLinkResult, error) {
	if output == "" {
		return nil, fmt.Errorf("empty nvidia-smi nvlink output")
	}

	results := make(map[string]*NVLinkResult)
	lines := strings.Split(output, "\n")

	var currentGPU *NVLinkResult
	gpuPattern := regexp.MustCompile(`GPU\s+(\d+):\s+(?:NVIDIA|HGX)`)
	linkPattern := regexp.MustCompile(`Link\s+(\d+):\s+([\d.]+)\s+GB/s`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for GPU line
		if gpuMatch := gpuPattern.FindStringSubmatch(line); gpuMatch != nil {
			gpuID := gpuMatch[1]
			currentGPU = &NVLinkResult{
				GPUID:     gpuID,
				LinkCount: 0,
				Links:     []NVLinkInfo{},
			}
			results[gpuID] = currentGPU
			logger.Debugf("Found GPU %s", gpuID)
			continue
		}

		// Check for Link line
		if linkMatch := linkPattern.FindStringSubmatch(line); linkMatch != nil && currentGPU != nil {
			linkIDStr := linkMatch[1]
			speedStr := linkMatch[2]

			linkID, err := strconv.Atoi(linkIDStr)
			if err != nil {
				logger.Errorf("Failed to parse link ID '%s': %v", linkIDStr, err)
				continue
			}

			speed, err := strconv.ParseFloat(speedStr, 64)
			if err != nil {
				logger.Errorf("Failed to parse link speed '%s': %v", speedStr, err)
				continue
			}

			isActive := !strings.Contains(strings.ToLower(line), "inactive")
			isGoodSpeed := speed >= expectedSpeed

			linkInfo := NVLinkInfo{
				LinkID:   linkID,
				Speed:    speed,
				IsActive: isActive,
			}

			currentGPU.Links = append(currentGPU.Links, linkInfo)

			// Count only active links that meet speed requirements
			if isActive && isGoodSpeed {
				currentGPU.LinkCount++
			}

			logger.Debugf("GPU %s Link %d: Speed=%.3f GB/s, Active=%v, GoodSpeed=%v",
				currentGPU.GPUID, linkID, speed, isActive, isGoodSpeed)
			continue
		}

		// Check for unexpected output
		if !strings.Contains(line, "GPU") && !strings.Contains(line, "Link") && line != "" {
			return nil, fmt.Errorf("unexpected entry in nvidia-smi nvlink -s output: %s", line)
		}
	}

	return results, nil
}

// validateNVLinkResults validates the parsed NVLink results against expected counts
func validateNVLinkResults(results map[string]*NVLinkResult, expectedCount int) (bool, []string) {
	var failedGPUs []string

	for gpuID, result := range results {
		if result.LinkCount != expectedCount {
			failedGPUs = append(failedGPUs, gpuID)
			logger.Errorf("GPU %s: Expected %d good NVLinks, found %d", gpuID, expectedCount, result.LinkCount)
		} else {
			logger.Infof("GPU %s: Found %d good NVLinks (expected %d) - PASS", gpuID, result.LinkCount, expectedCount)
		}
	}

	return len(failedGPUs) == 0, failedGPUs
}

// RunNVLinkSpeedCheck performs the NVLink speed and count check
func RunNVLinkSpeedCheck() error {
	logger.Info("=== NVLink Speed Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("NVLink Speed Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddNVLinkResult("FAIL", nil, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	nvlinkConfig, err := getNVLinkSpeedCheckTestConfig(shape)
	if err != nil {
		logger.Error("NVLink Speed Check: FAIL - Could not get test configuration:", err)
		rep.AddNVLinkResult("FAIL", nil, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !nvlinkConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Error(errorStatement)
		return errors.New(errorStatement)
	}

	logger.Infof("Step 2: Expected NVLink parameters - Speed: %.1f GB/s, Count: %d per GPU",
		nvlinkConfig.ExpectedSpeed, nvlinkConfig.ExpectedCount)

	// Step 3: Run nvidia-smi nvlink -s command
	logger.Info("Step 3: Running nvidia-smi nvlink -s...")
	nvlinkResult := executor.RunNvidiaSMINvlink()
	if !nvlinkResult.Available {
		logger.Error("NVLink Speed Check: FAIL - nvidia-smi nvlink command failed:", nvlinkResult.Error)
		err = fmt.Errorf("nvidia-smi nvlink command failed: %s", nvlinkResult.Error)
		rep.AddNVLinkResult("FAIL", nil, err)
		return err
	}

	if nvlinkResult.Output == "" {
		logger.Error("NVLink Speed Check: FAIL - Empty output from nvidia-smi nvlink -s")
		err = fmt.Errorf("empty output from nvidia-smi nvlink -s")
		rep.AddNVLinkResult("FAIL", nil, err)
		return err
	}

	logger.Debug("NVLink raw output:", nvlinkResult.Output)

	// Step 4: Parse the NVLink output
	logger.Info("Step 4: Parsing NVLink output...")
	parsedResults, err := parseNVLinkOutput(nvlinkResult.Output, nvlinkConfig.ExpectedSpeed)
	if err != nil {
		logger.Error("NVLink Speed Check: FAIL - Failed to parse nvidia-smi nvlink output:", err)
		rep.AddNVLinkResult("FAIL", nil, err)
		return fmt.Errorf("failed to parse nvidia-smi nvlink output: %w", err)
	}

	if len(parsedResults) == 0 {
		logger.Error("NVLink Speed Check: FAIL - No GPUs found in nvidia-smi nvlink output")
		err = fmt.Errorf("no GPUs found in nvidia-smi nvlink output")
		rep.AddNVLinkResult("FAIL", nil, err)
		return err
	}

	logger.Infof("Found %d GPUs with NVLink information", len(parsedResults))

	// Step 5: Validate results against expected counts
	logger.Info("Step 5: Validating NVLink counts...")
	isValid, failedGPUs := validateNVLinkResults(parsedResults, nvlinkConfig.ExpectedCount)

	// Convert results for reporting
	var reportData []interface{}
	for _, result := range parsedResults {
		reportData = append(reportData, result)
	}

	if isValid {
		logger.Info("NVLink Speed Check: PASS - All GPUs have correct NVLink speed and count")
		rep.AddNVLinkResult("PASS", reportData, nil)
		return nil
	} else {
		failedGPUList := strings.Join(failedGPUs, ",")
		logger.Errorf("NVLink Speed Check: FAIL - GPUs with incorrect NVLink count: %s", failedGPUList)
		err = fmt.Errorf("NVLink check failed for GPUs: %s", failedGPUList)
		rep.AddNVLinkResult("FAIL", reportData, err)
		return err
	}
}
