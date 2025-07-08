package level1_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
	"github.com/oracle/oci-dr-hpc-v2/internal/utils"
	"math"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
)

// RXDiscardsResult represents the result of RX discards check for a single interface
type RXDiscardsResult struct {
	RXDiscards RXDiscardsDevice `json:"rx_discards"`
}

// RXDiscardsDevice represents the device-specific RX discards information
type RXDiscardsDevice struct {
	Device string `json:"device"`
	Status string `json:"status"`
}

// RXDiscardsConfig represents the configuration for RX discards checking
type RXDiscardsConfig struct {
	Interfaces []string `json:"interfaces"`
}

// RxDiscardTestConfig represents the config needed to run this test
type RxDiscardTestConfig struct {
	Threshold float64 `json:"threshold"`
	IsEnabled bool    `json:"enabled"`
	Shape     string  `json:"shape"`
}

// TODO: bhrajan read configs from file
// getRXDiscardsConfig returns the configuration for RX discards checking
func getRXDiscardsConfig() *RXDiscardsConfig {
	return &RXDiscardsConfig{
		Interfaces: []string{
			"rdma0", "rdma1",
			"rdma2", "rdma3",
			"rdma4", "rdma5",
			"rdma6", "rdma7",
			"rdma8", "rdma9",
			"rdma10", "rdma11",
			"rdma12", "rdma13",
			"rdma14", "rdma15",
		},
	}
}

// parseRXDiscardsResults parses RX discards results for a single interface
func parseRXDiscardsResults(interfaceName string, results []string, threshold float64) RXDiscardsResult {
	// Create default result structure with PASS status
	result := RXDiscardsResult{
		RXDiscards: RXDiscardsDevice{
			Device: interfaceName,
			Status: "PASS",
		},
	}

	// Check if ethtool command returned any results
	if len(results) == 0 {
		// No results means interface doesn't exist or ethtool failed
		result.RXDiscards.Status = "FAIL"
		return result
	}

	// Process each line of ethtool output
	for _, line := range results {
		if len(line) > 0 {
			// Parse ethtool output format: "stat_name: value"
			// Remove spaces and split on colon to extract the numeric value
			cleanLine := strings.ReplaceAll(line, " ", "")
			parts := strings.Split(cleanLine, ":")

			if len(parts) >= 2 {
				discardsStr := parts[1]

				// Validate that the discard count is a valid integer
				if utils.IsInt(discardsStr) {
					discards, _ := strconv.Atoi(discardsStr)
					// Check if discard count exceeds threshold
					if float64(discards) > threshold {
						result.RXDiscards.Status = "FAIL"
						break // Exit early on first failure
					}
				} else {
					// Invalid discard value indicates parsing error or interface issue
					result.RXDiscards.Status = "FAIL"
					break // Exit early on parsing failure
				}
			}
		}
	}

	return result
}

// Gets test config needed to run this test
func getRxDiscardTestConfig() (*RxDiscardTestConfig, error) {
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

	// Result
	rxDiscardTestConfig := &RxDiscardTestConfig{
		Threshold: math.MinInt,
		IsEnabled: false,
		Shape:     shape,
	}

	enabled, err := limits.IsTestEnabled(shape, "rx_discards_check")
	if err != nil {
		return nil, err
	}
	rxDiscardTestConfig.IsEnabled = enabled

	if !enabled {
		return rxDiscardTestConfig, nil
	}
	defaultThreshold, err := limits.GetThresholdForTest(shape, "rx_discards_check")
	logger.Info("Fetched test config for this test ", rxDiscardTestConfig)
	if threshold, ok := defaultThreshold.(float64); ok {
		rxDiscardTestConfig.Threshold = threshold
	}
	return rxDiscardTestConfig, nil
}

// runRXDiscardsCheck executes RX discards health check across all relevant network interfaces
func runRXDiscardsCheck() ([]RXDiscardsResult, error) {
	// Process each interface and collect results
	var results []RXDiscardsResult

	config := getRXDiscardsConfig()
	interfacesList := config.Interfaces

	testConfig, err := getRxDiscardTestConfig()
	if err != nil {
		return nil, err
	}

	if !testConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", testConfig.Shape)
		logger.Info(errorStatement)
		return nil, errors.New(errorStatement)
	}

	// Threshold for considering RX discards problematic
	// Values above this indicate potential network issues
	threshold := testConfig.Threshold
	logger.Info("Running RX discards check with threshold:", threshold)

	for _, interfaceName := range interfacesList {
		logger.Debugf("Checking interface: %s", interfaceName)

		// Execute ethtool command to get RX discards statistics
		result, err := executor.RunEthtoolStats(interfaceName, "rx_prio.*_discards")
		if err != nil {
			logger.Debugf("ethtool failed for interface %s: %v", interfaceName, err)
			// Create failed result for this interface
			failResult := RXDiscardsResult{
				RXDiscards: RXDiscardsDevice{
					Device: interfaceName,
					Status: "FAIL",
				},
			}
			results = append(results, failResult)
			continue
		}

		// Parse the results and determine pass/fail status
		output := strings.TrimSpace(result.Output)
		var lines []string
		if output != "" {
			lines = strings.Split(output, "\n")
		}

		parsedResult := parseRXDiscardsResults(interfaceName, lines, threshold)
		results = append(results, parsedResult)
	}

	return results, nil
}

// RunRXDiscardsCheck is the main entry point for RX discards health check
func RunRXDiscardsCheck() error {
	logger.Info("=== RX Discards Health Check ===")
	rep := reporter.GetReporter()

	logger.Info("Health check is in progress...")

	// Run the RX discards check
	results, err := runRXDiscardsCheck()
	if err != nil {
		logger.Error("RX Discards Check: FAIL - Error during check:", err)
		rep.AddNetworkRxDiscardsResult("FAIL", 0, err)
		return fmt.Errorf("failed to run RX discards check: %w", err)
	}

	// Convert results to JSON for logging and reporting
	jsonResults, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logger.Error("RX Discards Check: FAIL - Failed to marshal results:", err)
		rep.AddNetworkRxDiscardsResult("FAIL", 0, err)
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	logger.Info("RX Discards Check results:")
	logger.Info(string(jsonResults))

	// Check if any interface failed
	failedCount := 0
	passedCount := 0
	for _, result := range results {
		if result.RXDiscards.Status == "FAIL" {
			failedCount++
			logger.Errorf("Interface %s: FAIL", result.RXDiscards.Device)
		} else {
			passedCount++
			logger.Infof("Interface %s: PASS", result.RXDiscards.Device)
		}
	}

	// Report overall result
	if failedCount > 0 {
		err := fmt.Errorf("RX discards check failed for %d out of %d interfaces", failedCount, len(results))
		logger.Error("RX Discards Check: FAIL -", err)
		rep.AddNetworkRxDiscardsResult("FAIL", failedCount, err)
		return err
	}

	logger.Info("RX Discards Check: PASS - All interfaces passed")
	rep.AddNetworkRxDiscardsResult("PASS", len(results), nil)
	return nil
}
