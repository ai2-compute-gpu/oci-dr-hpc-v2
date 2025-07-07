package level1_tests

import (
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-dr-hpc-v2/internal/Utils"
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
	Interfaces               []string `json:"interfaces"`
	RXDiscardsCheckThreshold int      `json:"rx_discards_check_threshold"`
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
		// Threshold for considering RX discards problematic
		// Values above this indicate potential network issues
		RXDiscardsCheckThreshold: 100,
	}
}

// parseRXDiscardsResults parses RX discards results for a single interface
func parseRXDiscardsResults(interfaceName string, results []string, threshold int) RXDiscardsResult {
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
				if Utils.IsInt(discardsStr) {
					discards, _ := strconv.Atoi(discardsStr)
					// Check if discard count exceeds threshold
					if discards > threshold {
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

// runRXDiscardsCheck executes RX discards health check across all relevant network interfaces
func runRXDiscardsCheck() ([]RXDiscardsResult, error) {
	config := getRXDiscardsConfig()
	interfacesList := config.Interfaces
	threshold := config.RXDiscardsCheckThreshold

	logger.Info("Running RX discards check with threshold:", threshold)

	// Process each interface and collect results
	var results []RXDiscardsResult

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
