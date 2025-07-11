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

// GIDIndexResult represents the result of GID index parsing
type GIDIndexResult struct {
	Device   string `json:"interface"`
	Port     string `json:"port"`
	GIDIndex int    `json:"gid_index"`
	GIDValue string `json:"gid_value"`
}

// GIDIndexCheckTestConfig represents the test configuration for GID index check
type GIDIndexCheckTestConfig struct {
	IsEnabled          bool  `json:"enabled"`
	ExpectedGIDIndexes []int `json:"expected_gid_indexes"`
}

// getGIDIndexCheckTestConfig gets test config needed to run this test
func getGIDIndexCheckTestConfig(shape string) (*GIDIndexCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, err
	}

	// Result
	gidIndexCheckTestConfig := &GIDIndexCheckTestConfig{
		IsEnabled:          false,
		ExpectedGIDIndexes: []int{0, 1, 2, 3}, // Default expected values
	}

	enabled, err := limits.IsTestEnabled(shape, "gid_index_check")
	if err != nil {
		return nil, err
	}
	gidIndexCheckTestConfig.IsEnabled = enabled

	threshold, err := limits.GetThresholdForTest(shape, "gid_index_check")
	if err != nil {
		return nil, err
	}

	// Handle different threshold formats
	switch v := threshold.(type) {
	case []interface{}:
		// Convert []interface{} to []int
		var indexes []int
		for _, item := range v {
			if val, ok := item.(float64); ok {
				indexes = append(indexes, int(val))
			} else if val, ok := item.(int); ok {
				indexes = append(indexes, val)
			}
		}
		if len(indexes) > 0 {
			gidIndexCheckTestConfig.ExpectedGIDIndexes = indexes
		}
	case []int:
		gidIndexCheckTestConfig.ExpectedGIDIndexes = v
	default:
		// Keep default values if threshold format is unexpected
		logger.Info("Unexpected threshold format for gid_index_check, using default [0,1,2,3]")
	}

	return gidIndexCheckTestConfig, nil
}

// parseGIDIndexResults parses the output from show_gids command
func parseGIDIndexResults(output string) ([]GIDIndexResult, error) {
	var results []GIDIndexResult

	if len(output) == 0 {
		return results, fmt.Errorf("no GID index output to parse")
	}

	lines := strings.Split(output, "\n")
	for index, line := range lines {
		if index == 0 || index == 1 || index == len(lines)-1 || index == len(lines)-2 {
			// Excluding headers (0,1) and footer lines
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse GID index
		fields := strings.Fields(line)
		gidIndex, err := strconv.Atoi(fields[2])
		if err != nil {
			logger.Info("Invalid GID index in line:", line, "Error:", err)
			continue
		}

		result := GIDIndexResult{
			Device:   fields[0],
			Port:     fields[1],
			GIDIndex: gidIndex,
			GIDValue: fields[3],
		}

		results = append(results, result)
	}

	return results, nil
}

// checkGIDIndexes validates that all GID indexes are within expected values
func checkGIDIndexes(results []GIDIndexResult, expectedIndexes []int) (bool, []int, error) {
	if len(results) == 0 {
		return false, nil, fmt.Errorf("no GID index results to validate")
	}

	// Convert expected indexes to a map for faster lookup
	expectedMap := make(map[int]bool)
	for _, idx := range expectedIndexes {
		expectedMap[idx] = true
	}

	var invalidIndexes []int
	allValid := true

	for _, result := range results {
		if !expectedMap[result.GIDIndex] {
			allValid = false
			// Add to invalid list if not already present
			found := false
			for _, invalid := range invalidIndexes {
				if invalid == result.GIDIndex {
					found = true
					break
				}
			}
			if !found {
				invalidIndexes = append(invalidIndexes, result.GIDIndex)
			}
		}
	}

	return allValid, invalidIndexes, nil
}

// RunGIDIndexCheck performs the GID index check
func RunGIDIndexCheck() error {
	logger.Info("=== GID Index Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("GID Index Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddGIDIndexResult("FAIL", []int{}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	gidIndexCheckTestConfig, err := getGIDIndexCheckTestConfig(shape)
	if err != nil {
		logger.Error("GID Index Check: FAIL - Could not get test configuration:", err)
		rep.AddGIDIndexResult("FAIL", []int{}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !gidIndexCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Get expected GID indexes from configuration
	logger.Info("Step 2: Getting expected GID indexes from configuration...")
	expectedIndexes := gidIndexCheckTestConfig.ExpectedGIDIndexes
	logger.Info("Expected GID indexes:", expectedIndexes)

	// Step 4: Execute show_gids command and parse output
	// https://enterprise-support.nvidia.com/s/article/understanding-show-gids-script#jive_content_id_References
	logger.Info("Step 3: Getting GID index information from show_gids...")
	gidOutput, err := executor.RunShowGids()

	if err != nil {
		logger.Error("GID Index Check: FAIL - Could not get GID index output:", err)
		rep.AddGIDIndexResult("FAIL", []int{}, err)
		return fmt.Errorf("failed to get GID index output: %w", err)
	}

	// Step 5: Parse the GID index results
	logger.Info("Step 4: Parsing GID index results...")
	gidResults, err := parseGIDIndexResults(gidOutput.Output)
	if err != nil {
		logger.Error("GID Index Check: FAIL - Could not parse GID index results:", err)
		rep.AddGIDIndexResult("FAIL", []int{}, err)
		return fmt.Errorf("failed to parse GID index results: %w", err)
	}
	logger.Info("Found ", len(gidResults), " GID entries")

	// Step 6: Validate GID indexes against expected values
	logger.Info("Step 5: Validating GID indexes...")
	allValid, invalidIndexes, err := checkGIDIndexes(gidResults, expectedIndexes)
	if err != nil {
		logger.Error("GID Index Check: FAIL - Could not validate GID indexes:", err)
		rep.AddGIDIndexResult("FAIL", invalidIndexes, err)
		return fmt.Errorf("failed to validate GID indexes: %w", err)
	}

	// Step 7: Report results
	if allValid {
		logger.Info("GID Index Check: PASS - All GID indexes are within expected values:", expectedIndexes)
		rep.AddGIDIndexResult("PASS", []int{}, nil)
		return nil
	} else {
		logger.Error("GID Index Check: FAIL - Found invalid GID indexes:", invalidIndexes)
		logger.Error("Expected GID indexes:", expectedIndexes)
		err = fmt.Errorf("invalid GID indexes found: %v, expected: %v", invalidIndexes, expectedIndexes)
		rep.AddGIDIndexResult("FAIL", invalidIndexes, err)
		return err
	}
}

// PrintGIDIndexCheck prints a placeholder message for GID index check
func PrintGIDIndexCheck() {
	// This function is a placeholder for GID index check logic.
	logger.Info("GID Index Check: Checking system GID indexes...")

	// Example implementation (to be replaced with actual logic):
	expectedIndexes := []int{0, 1, 2, 3}
	logger.Info("GID Index Check: PASS - Expected GID indexes:", expectedIndexes)
}
