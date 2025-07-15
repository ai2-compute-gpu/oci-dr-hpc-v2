package level1_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/shapes"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// EthLinkCheckResult represents the result of Ethernet link parsing
type EthLinkCheckResult struct {
	Device                      string `json:"device"`
	EthLinkSpeed               string `json:"eth_link_speed"`
	EthLinkState               string `json:"eth_link_state"`
	PhysicalState              string `json:"physical_state"`
	EthLinkWidth               string `json:"eth_link_width"`
	EthLinkStatus              string `json:"eth_link_status"`
	EffectivePhysicalErrors    string `json:"effective_physical_errors"`
	EffectivePhysicalBER       string `json:"effective_physical_ber"`
	RawPhysicalErrorsPerLane   string `json:"raw_physical_errors_per_lane"`
	RawPhysicalBER             string `json:"raw_physical_ber"`
}

// EthLinkCheckTestConfig represents the test configuration for Ethernet link check
type EthLinkCheckTestConfig struct {
	IsEnabled                           bool    `json:"enabled"`
	ExpectedSpeed                       string  `json:"speed"`
	ExpectedWidth                       string  `json:"width"`
	EffectivePhysicalErrorsThreshold    int     `json:"effective_physical_errors"`
	RawPhysicalErrorsPerLaneThreshold   int     `json:"raw_physical_errors_per_lane"`
	EffectivePhysicalBERThreshold       float64 `json:"effective_physical_ber"`
	RawPhysicalBERThreshold             float64 `json:"raw_physical_ber"`
}

// getEthLinkCheckTestConfig gets test config needed to run this test
func getEthLinkCheckTestConfig(shape string) (*EthLinkCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults from test_limits.json
	ethLinkCheckTestConfig := &EthLinkCheckTestConfig{
		IsEnabled: false,
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "eth_link_check")
	if err != nil {
		logger.Info("Ethernet link check test not found for shape", shape, ", defaulting to disabled")
		return ethLinkCheckTestConfig, nil
	}
	ethLinkCheckTestConfig.IsEnabled = enabled

	// If test is disabled, return early
	if !enabled {
		logger.Info("Ethernet link check test disabled for shape", shape)
		return ethLinkCheckTestConfig, nil
	}

	// Get threshold configuration
	threshold, err := limits.GetThresholdForTest(shape, "eth_link_check")
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold configuration for eth_link_check on shape %s: %w", shape, err)
	}

	// Parse threshold configuration from test_limits.json
	switch v := threshold.(type) {
	case map[string]interface{}:
		// Update speed if specified
		if speed, ok := v["speed"].(string); ok {
			ethLinkCheckTestConfig.ExpectedSpeed = speed
			logger.Info("Using configured speed:", speed, "for shape", shape)
		}
		
		// Update width if specified
		if width, ok := v["width"].(string); ok {
			ethLinkCheckTestConfig.ExpectedWidth = width
			logger.Info("Using configured width:", width, "for shape", shape)
		}
		
		// Update effective physical errors threshold if specified
		if effErrors, ok := v["effective_physical_errors"].(float64); ok {
			ethLinkCheckTestConfig.EffectivePhysicalErrorsThreshold = int(effErrors)
			logger.Info("Using configured effective physical errors threshold:", int(effErrors), "for shape", shape)
		}
		
		// Update raw physical errors per lane threshold if specified
		if rawErrors, ok := v["raw_physical_errors_per_lane"].(float64); ok {
			ethLinkCheckTestConfig.RawPhysicalErrorsPerLaneThreshold = int(rawErrors)
			logger.Info("Using configured raw physical errors per lane threshold:", int(rawErrors), "for shape", shape)
		}
		
		// Update effective physical BER threshold if specified
		if effBER, ok := v["effective_physical_ber"].(float64); ok {
			ethLinkCheckTestConfig.EffectivePhysicalBERThreshold = effBER
			logger.Info("Using configured effective physical BER threshold:", effBER, "for shape", shape)
		}
		
		// Update raw physical BER threshold if specified
		if rawBER, ok := v["raw_physical_ber"].(float64); ok {
			ethLinkCheckTestConfig.RawPhysicalBERThreshold = rawBER
			logger.Info("Using configured raw physical BER threshold:", rawBER, "for shape", shape)
		}
		
		logger.Info("Successfully loaded eth_link_check configuration for shape", shape)
	default:
		return nil, fmt.Errorf("unexpected threshold format for eth_link_check on shape %s", shape)
	}

	return ethLinkCheckTestConfig, nil
}

// parseEthLinkResults parses the output from mlxlink command and validates Ethernet link parameters
func parseEthLinkResults(interfaceName string, mlxlinkOutput string, expectedSpeed string, expectedWidth string,
	rawPhysicalErrorsPerLaneThreshold int, effectivePhysicalErrorsThreshold int,
	effectivePhysicalBERThreshold float64, rawPhysicalBERThreshold float64) (*EthLinkCheckResult, error) {

	result := &EthLinkCheckResult{
		Device: interfaceName,
	}

	// Helper functions
	isFloat := func(val string) bool {
		_, err := strconv.ParseFloat(val, 64)
		return err == nil
	}

	isInt := func(val string) bool {
		_, err := strconv.Atoi(val)
		return err == nil
	}

	parseRawPhysicalErrorsPerLane := func(val interface{}) []int {
		switch v := val.(type) {
		case []interface{}:
			var errors []int
			for _, item := range v {
				if itemStr, ok := item.(string); ok && itemStr != "undefined" {
					if itemInt, err := strconv.Atoi(itemStr); err == nil {
						errors = append(errors, itemInt)
					}
				} else if itemFloat, ok := item.(float64); ok {
					errors = append(errors, int(itemFloat))
				}
			}
			return errors
		case string:
			if v != "undefined" {
				if itemInt, err := strconv.Atoi(v); err == nil {
					return []int{itemInt}
				}
			}
		}
		return []int{}
	}

	// If error, try to extract JSON
	if strings.HasPrefix(mlxlinkOutput, "Error:") {
		index := strings.Index(mlxlinkOutput, "{")
		if index != -1 {
			mlxlinkOutput = mlxlinkOutput[index:]
		} else {
			result.EthLinkSpeed = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.EthLinkState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.PhysicalState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.EthLinkWidth = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.EthLinkStatus = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.EffectivePhysicalErrors = "PASS"
			result.EffectivePhysicalBER = "FAIL - Unable to get data"
			result.RawPhysicalErrorsPerLane = "PASS"
			result.RawPhysicalBER = "FAIL - Unable to get data"
			return result, nil
		}
	}

	if strings.TrimSpace(mlxlinkOutput) == "" {
		result.EthLinkSpeed = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.EthLinkState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.PhysicalState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.EthLinkWidth = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.EthLinkStatus = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.EffectivePhysicalErrors = "PASS"
		result.EffectivePhysicalBER = "FAIL - Unable to get data"
		result.RawPhysicalErrorsPerLane = "PASS"
		result.RawPhysicalBER = "FAIL - Unable to get data"
		return result, nil
	}

	// Parse JSON output
	var mlxData map[string]interface{}
	if err := json.Unmarshal([]byte(mlxlinkOutput), &mlxData); err != nil {
		result.EthLinkSpeed = "FAIL - Unable to parse mlxlink output"
		result.EthLinkState = "FAIL - Unable to parse mlxlink output"
		result.PhysicalState = "FAIL - Unable to parse mlxlink output"
		result.EthLinkWidth = "FAIL - Unable to parse mlxlink output"
		result.EthLinkStatus = "FAIL - Unable to parse mlxlink output"
		result.EffectivePhysicalErrors = "PASS"
		result.EffectivePhysicalBER = "FAIL - Unable to parse mlxlink output"
		result.RawPhysicalErrorsPerLane = "PASS"
		result.RawPhysicalBER = "FAIL - Unable to parse mlxlink output"
		return result, nil
	}

	// Extract result.output from JSON
	resultData, ok := mlxData["result"].(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("unable to find result in mlxlink output")
	}

	output, ok := resultData["output"].(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("unable to find output in mlxlink result")
	}

	// Expected values for Ethernet interfaces
	expectedState := "Active"
	expectedPhysStates := []string{"LinkUp", "ETH_AN_FSM_ENABLE"}

	// Extract fields
	var speed, state, physState, width, statusOpcode, recommendation string
	var effectivePhysicalErrors, effectivePhysicalBER, rawPhysicalBER string
	var rawPhysicalErrorsPerLane []int

	if opInfo, ok := output["Operational Info"].(map[string]interface{}); ok {
		if s, ok := opInfo["Speed"].(string); ok {
			speed = s
		}
		if s, ok := opInfo["State"].(string); ok {
			state = s
		}
		if s, ok := opInfo["Physical state"].(string); ok {
			physState = s
		}
		if s, ok := opInfo["Width"].(string); ok {
			width = s
		}
	}

	if troubleInfo, ok := output["Troubleshooting Info"].(map[string]interface{}); ok {
		if s, ok := troubleInfo["Status Opcode"].(string); ok {
			statusOpcode = s
		}
		if s, ok := troubleInfo["Recommendation"].(string); ok {
			recommendation = s
		}
	}

	if physCounters, ok := output["Physical Counters and BER Info"].(map[string]interface{}); ok {
		if s, ok := physCounters["Effective Physical Errors"].(string); ok {
			effectivePhysicalErrors = s
		}
		if s, ok := physCounters["Effective Physical BER"].(string); ok {
			effectivePhysicalBER = s
		}
		if s, ok := physCounters["Raw Physical BER"].(string); ok {
			rawPhysicalBER = s
		}
		rawPhysicalErrorsPerLane = parseRawPhysicalErrorsPerLane(physCounters["Raw Physical Errors Per Lane"])
	}

	// Set initial FAIL results
	result.EthLinkSpeed = fmt.Sprintf("FAIL - %s, expected %s", speed, expectedSpeed)
	result.EthLinkState = fmt.Sprintf("FAIL - %s, expected %s", state, expectedState)
	result.PhysicalState = fmt.Sprintf("FAIL - %s, expected %v", physState, expectedPhysStates)
	result.EthLinkWidth = fmt.Sprintf("FAIL - %s, expected %s", width, expectedWidth)
	result.EthLinkStatus = fmt.Sprintf("FAIL - %s", recommendation)
	result.EffectivePhysicalErrors = "PASS"
	result.EffectivePhysicalBER = fmt.Sprintf("FAIL - %s", effectivePhysicalBER)
	result.RawPhysicalErrorsPerLane = "PASS"
	result.RawPhysicalBER = fmt.Sprintf("FAIL - %s", rawPhysicalBER)

	// Set PASS if matches
	if strings.Contains(speed, expectedSpeed) {
		result.EthLinkSpeed = "PASS"
	}
	if state == expectedState {
		result.EthLinkState = "PASS"
	}
	for _, expectedPhysState := range expectedPhysStates {
		if physState == expectedPhysState {
			result.PhysicalState = "PASS"
			break
		}
	}
	if width == expectedWidth {
		result.EthLinkWidth = "PASS"
	}
	if statusOpcode == "0" {
		result.EthLinkStatus = "PASS"
	}
	if isFloat(effectivePhysicalBER) {
		if berFloat, err := strconv.ParseFloat(effectivePhysicalBER, 64); err == nil && berFloat < effectivePhysicalBERThreshold {
			result.EffectivePhysicalBER = "PASS"
		}
	}
	if isFloat(rawPhysicalBER) {
		if berFloat, err := strconv.ParseFloat(rawPhysicalBER, 64); err == nil && berFloat < rawPhysicalBERThreshold {
			result.RawPhysicalBER = "PASS"
		}
	}
	if isInt(effectivePhysicalErrors) {
		if errInt, err := strconv.Atoi(effectivePhysicalErrors); err == nil && errInt > effectivePhysicalErrorsThreshold {
			result.EffectivePhysicalErrors = fmt.Sprintf("FAIL - %s", effectivePhysicalErrors)
		}
	}

	// Check raw physical errors per lane
	for _, laneError := range rawPhysicalErrorsPerLane {
		if laneError > rawPhysicalErrorsPerLaneThreshold {
			var errorsStr []string
			for _, err := range rawPhysicalErrorsPerLane {
				errorsStr = append(errorsStr, strconv.Itoa(err))
			}
			result.RawPhysicalErrorsPerLane = fmt.Sprintf("WARN - %s", strings.Join(errorsStr, " "))
			break
		}
	}

	return result, nil
}

// RunEthLinkCheck performs the Ethernet link check
func RunEthLinkCheck() error {
	logger.Info("=== Ethernet Link Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	ethLinkCheckTestConfig, err := getEthLinkCheckTestConfig(shape)
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not get test configuration:", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !ethLinkCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Get device mapping using ibdev2netdev
	logger.Info("Step 3: Getting device mapping...")
	deviceMap, err := executor.GetIbdevToNetdevMap()
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not get device mapping:", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to get device mapping: %w", err)
	}

	// Step 4: Get PCI mapping using mst status
	logger.Info("Step 4: Getting PCI mapping...")
	mstResult, err := executor.RunMstStatus()
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not get MST status:", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to get MST status: %w", err)
	}

	// Parse MST output to build device to PCI mapping
	deviceToPCIMap := make(map[string]string)
	lines := strings.Split(mstResult.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Look for lines containing device names from deviceMap
		for device := range deviceMap {
			if strings.Contains(line, device) && len(strings.Fields(line)) >= 3 {
				pci := strings.Fields(line)[2]
				deviceToPCIMap[device] = pci
				break
			}
		}
	}

	// Step 5: Get VCN NICs (non-RDMA Ethernet interfaces) from shapes.json
	logger.Info("Step 5: Getting VCN NICs from shapes configuration...")
	shapeManager, err := shapes.NewShapeManager("internal/shapes/shapes.json")
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not load shapes configuration:", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to load shapes configuration: %w", err)
	}

	vcnNics, err := shapeManager.GetVCNNics(shape)
	if err != nil {
		logger.Error("Ethernet Link Check: FAIL - Could not get VCN NICs for shape", shape, ":", err)
		rep.AddEthLinkResult("FAIL", []EthLinkCheckResult{}, err)
		return fmt.Errorf("failed to get VCN NICs for shape %s: %w", shape, err)
	}

	logger.Info("Found", len(vcnNics), "VCN NICs for shape", shape)

	// Step 6: Map VCN device names to actual interfaces
	var interfacesToCheck []string
	deviceToInterfaceMap := make(map[string]string)

	for _, vcnNic := range vcnNics {
		logger.Info("Looking for VCN device:", vcnNic.DeviceName, "PCI:", vcnNic.PCI)
		
		// Find the corresponding interface for this VCN device
		if interfaceName, exists := deviceMap[vcnNic.DeviceName]; exists {
			interfacesToCheck = append(interfacesToCheck, interfaceName)
			deviceToInterfaceMap[vcnNic.DeviceName] = interfaceName
			logger.Info("Mapped VCN device", vcnNic.DeviceName, "to interface", interfaceName)
		} else {
			logger.Info("VCN device", vcnNic.DeviceName, "not found in device mapping")
		}
	}

	if len(interfacesToCheck) == 0 {
		errorStatement := "No Ethernet interfaces found for checking"
		logger.Info("Ethernet Link Check: INFO -", errorStatement)
		// This is not an error condition, just no interfaces to check
		rep.AddEthLinkResult("SKIP", []EthLinkCheckResult{}, nil)
		return nil
	}

	logger.Info("Step 7: Found", len(interfacesToCheck), "VCN Ethernet interfaces to check:", interfacesToCheck)

	// Step 7: Check all VCN interfaces  
	var allResults []EthLinkCheckResult

	for _, interfaceName := range interfacesToCheck {
		var mlxlinkOutput string
		var deviceName string

		// Find the VCN device for this interface
		deviceFound := false
		for device, netInterface := range deviceToInterfaceMap {
			if interfaceName == netInterface {
				deviceName = device
				// Run mlxlink for this VCN device
				result, err := executor.RunMlxlink(device)
				if err != nil {
					logger.Error("Failed to run mlxlink for VCN device", device, "(interface", interfaceName, "):", err)
					mlxlinkOutput = ""
				} else {
					mlxlinkOutput = result.Output
				}
				deviceFound = true
				break
			}
		}

		if !deviceFound {
			logger.Error("No VCN device found for interface", interfaceName)
			mlxlinkOutput = ""
		} else {
			logger.Info("Running mlxlink check for VCN device", deviceName, "interface", interfaceName)
		}

		// Parse results
		ethLinkResult, err := parseEthLinkResults(
			interfaceName,
			mlxlinkOutput,
			ethLinkCheckTestConfig.ExpectedSpeed,
			ethLinkCheckTestConfig.ExpectedWidth,
			ethLinkCheckTestConfig.RawPhysicalErrorsPerLaneThreshold,
			ethLinkCheckTestConfig.EffectivePhysicalErrorsThreshold,
			ethLinkCheckTestConfig.EffectivePhysicalBERThreshold,
			ethLinkCheckTestConfig.RawPhysicalBERThreshold,
		)
		if err != nil {
			logger.Errorf("Failed to parse Ethernet link results for %s: %v", interfaceName, err)
			continue
		}

		allResults = append(allResults, *ethLinkResult)
	}

	// Step 8: Report results
	logger.Info("Step 8: Reporting results...")
	if len(allResults) == 0 {
		logger.Error("Ethernet Link Check: FAIL - No Ethernet link results obtained")
		err = fmt.Errorf("no Ethernet link results obtained")
		rep.AddEthLinkResult("FAIL", allResults, err)
		return err
	}

	// Check if all links passed
	allPassed := true
	for _, result := range allResults {
		if !strings.HasPrefix(result.EthLinkSpeed, "PASS") ||
			!strings.HasPrefix(result.EthLinkState, "PASS") ||
			!strings.HasPrefix(result.PhysicalState, "PASS") ||
			!strings.HasPrefix(result.EthLinkWidth, "PASS") ||
			!strings.HasPrefix(result.EthLinkStatus, "PASS") ||
			!strings.HasPrefix(result.EffectivePhysicalBER, "PASS") ||
			!strings.HasPrefix(result.RawPhysicalBER, "PASS") ||
			strings.HasPrefix(result.EffectivePhysicalErrors, "FAIL") ||
			strings.HasPrefix(result.RawPhysicalErrorsPerLane, "WARN") {
			allPassed = false
			break
		}
	}

	if allPassed {
		logger.Info("Ethernet Link Check: PASS - All Ethernet links are healthy")
		rep.AddEthLinkResult("PASS", allResults, nil)
		return nil
	} else {
		logger.Error("Ethernet Link Check: FAIL - Some Ethernet links have issues")
		err = fmt.Errorf("some Ethernet links have issues")
		rep.AddEthLinkResult("FAIL", allResults, err)
		return err
	}
}