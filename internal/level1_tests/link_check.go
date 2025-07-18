package level1_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/config"
	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/oracle/oci-dr-hpc-v2/internal/shapes"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
)

// LinkCheckResult represents the result of link parsing
type LinkCheckResult struct {
	Device                      string `json:"device"`
	LinkSpeed                   string `json:"link_speed"`
	LinkState                   string `json:"link_state"`
	PhysicalState               string `json:"physical_state"`
	LinkStatus                  string `json:"link_status"`
	EffectivePhysicalErrors     string `json:"effective_physical_errors"`
	EffectivePhysicalBER        string `json:"effective_physical_ber"`
	RawPhysicalErrorsPerLane    string `json:"raw_physical_errors_per_lane"`
	RawPhysicalBER              string `json:"raw_physical_ber"`
}

// LinkCheckTestConfig represents the test configuration for link check
type LinkCheckTestConfig struct {
	IsEnabled                           bool    `json:"enabled"`
	ExpectedSpeed                       string  `json:"speed"`
	EffectivePhysicalErrorsThreshold    int     `json:"effective_physical_errors"`
	RawPhysicalErrorsPerLaneThreshold   int     `json:"raw_physical_errors_per_lane"`
}

// getLinkCheckTestConfig gets test config needed to run this test
func getLinkCheckTestConfig(shape string) (*LinkCheckTestConfig, error) {
	// Load configuration from test_limits.json
	limits, err := test_limits.LoadTestLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load test limits: %w", err)
	}

	// Initialize with defaults
	linkCheckTestConfig := &LinkCheckTestConfig{
		IsEnabled:                           false,
		ExpectedSpeed:                       "",
		EffectivePhysicalErrorsThreshold:    -1,
		RawPhysicalErrorsPerLaneThreshold:   -1,
	}

	// Check if test is enabled for this shape
	enabled, err := limits.IsTestEnabled(shape, "link_check")
	if err != nil {
		logger.Info("Link check test not found for shape", shape, ", defaulting to disabled")
		return linkCheckTestConfig, nil
	}
	linkCheckTestConfig.IsEnabled = enabled

	// If test is disabled, return early with defaults
	if !enabled {
		logger.Info("Link check test disabled for shape", shape)
		return linkCheckTestConfig, nil
	}

	// Get threshold configuration
	threshold, err := limits.GetThresholdForTest(shape, "link_check")
	if err != nil {
		logger.Info("No threshold configuration found for link_check on shape", shape, ", using defaults")
		return linkCheckTestConfig, nil
	}

	// Parse threshold configuration from test_limits.json
	switch v := threshold.(type) {
	case map[string]interface{}:
		// Update speed if specified
		if speed, ok := v["speed"].(string); ok {
			linkCheckTestConfig.ExpectedSpeed = speed
			logger.Info("Using configured speed:", speed, "for shape", shape)
		}
		
		// Update effective physical errors threshold if specified
		if effErrors, ok := v["effective_physical_errors"].(float64); ok {
			linkCheckTestConfig.EffectivePhysicalErrorsThreshold = int(effErrors)
			logger.Info("Using configured effective physical errors threshold:", int(effErrors), "for shape", shape)
		}
		
		// Update raw physical errors per lane threshold if specified
		if rawErrors, ok := v["raw_physical_errors_per_lane"].(float64); ok {
			linkCheckTestConfig.RawPhysicalErrorsPerLaneThreshold = int(rawErrors)
			logger.Info("Using configured raw physical errors per lane threshold:", int(rawErrors), "for shape", shape)
		}
		
		logger.Info("Successfully loaded link_check configuration for shape", shape)
	default:
		logger.Info("Unexpected threshold format for link_check on shape", shape, ", using defaults")
	}

	return linkCheckTestConfig, nil
}

// parseLinkResults parses the output from mlxlink command and validates link parameters
func parseLinkResults(interfaceName string, mlxlinkOutput string, expectedSpeed string,
	rawPhysicalErrorsPerLaneThreshold int, effectivePhysicalErrorsThreshold int) (*LinkCheckResult, error) {

	result := &LinkCheckResult{
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
			result.LinkSpeed = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.LinkState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.PhysicalState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.LinkStatus = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
			result.EffectivePhysicalErrors = "PASS"
			result.EffectivePhysicalBER = "FAIL - Unable to get data"
			result.RawPhysicalErrorsPerLane = "PASS"
			result.RawPhysicalBER = "FAIL - Unable to get data"
			return result, nil
		}
	}

	if strings.TrimSpace(mlxlinkOutput) == "" {
		result.LinkSpeed = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.LinkState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.PhysicalState = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.LinkStatus = fmt.Sprintf("FAIL - Invalid interface: %s", interfaceName)
		result.EffectivePhysicalErrors = "PASS"
		result.EffectivePhysicalBER = "FAIL - Unable to get data"
		result.RawPhysicalErrorsPerLane = "PASS"
		result.RawPhysicalBER = "FAIL - Unable to get data"
		return result, nil
	}

	// Parse JSON output
	var mlxData map[string]interface{}
	if err := json.Unmarshal([]byte(mlxlinkOutput), &mlxData); err != nil {
		result.LinkSpeed = "FAIL - Unable to parse mlxlink output"
		result.LinkState = "FAIL - Unable to parse mlxlink output"
		result.PhysicalState = "FAIL - Unable to parse mlxlink output"
		result.LinkStatus = "FAIL - Unable to parse mlxlink output"
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

	// Expected values
	expectedState := "Active"
	expectedPhysStates := []string{"LinkUp", "ETH_AN_FSM_ENABLE"}

	// Extract fields
	var speed, state, physState, statusOpcode, recommendation string
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
	result.LinkSpeed = fmt.Sprintf("FAIL - %s, expected %s", speed, expectedSpeed)
	result.LinkState = fmt.Sprintf("FAIL - %s, expected %s", state, expectedState)
	result.PhysicalState = fmt.Sprintf("FAIL - %s, expected %v", physState, expectedPhysStates)
	result.LinkStatus = fmt.Sprintf("FAIL - %s", recommendation)
	result.EffectivePhysicalErrors = "PASS"
	result.EffectivePhysicalBER = fmt.Sprintf("FAIL - %s", effectivePhysicalBER)
	result.RawPhysicalErrorsPerLane = "PASS"
	result.RawPhysicalBER = fmt.Sprintf("FAIL - %s", rawPhysicalBER)

	// Set PASS if matches
	if strings.Contains(speed, expectedSpeed) {
		result.LinkSpeed = "PASS"
	}
	if state == expectedState {
		result.LinkState = "PASS"
	}
	for _, expectedPhysState := range expectedPhysStates {
		if physState == expectedPhysState {
			result.PhysicalState = "PASS"
			break
		}
	}
	if statusOpcode == "0" {
		result.LinkStatus = "PASS"
	}
	if isFloat(effectivePhysicalBER) {
		if berFloat, err := strconv.ParseFloat(effectivePhysicalBER, 64); err == nil && berFloat < 1E-12 {
			result.EffectivePhysicalBER = "PASS"
		}
	}
	if isFloat(rawPhysicalBER) {
		if berFloat, err := strconv.ParseFloat(rawPhysicalBER, 64); err == nil && berFloat < 1E-5 {
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

// RunLinkCheck performs the link check
func RunLinkCheck() error {
	logger.Info("=== Link Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("Link Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Check if the test is enabled for this shape
	linkCheckTestConfig, err := getLinkCheckTestConfig(shape)
	if err != nil {
		logger.Error("Link Check: FAIL - Could not get test configuration:", err)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return fmt.Errorf("failed to get test configuration: %w", err)
	}

	if !linkCheckTestConfig.IsEnabled {
		errorStatement := fmt.Sprintf("Test not applicable for this shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Step 3: Get expected device names from shapes configuration
	logger.Info("Step 2: Loading shape configuration...")
	shapesFilePath := config.GetShapesFilePath()
	shapeManager, err := shapes.NewShapeManager(shapesFilePath)
	if err != nil {
		logger.Error("Link Check: FAIL - Could not load shapes configuration:", err)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return fmt.Errorf("failed to load shapes configuration: %w", err)
	}

	rdmaNics, err := shapeManager.GetRDMANics(shape)
	if err != nil {
		logger.Error("Link Check: FAIL - Could not get expected RDMA NICs for shape", shape, ":", err)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return fmt.Errorf("failed to get expected RDMA NICs: %w", err)
	}

	if len(rdmaNics) == 0 {
		errorStatement := fmt.Sprintf("No RDMA devices expected for shape %s", shape)
		logger.Info(errorStatement)
		return errors.New(errorStatement)
	}

	// Extract device names from RDMA NICs
	var expectedDeviceNames []string
	for _, nic := range rdmaNics {
		if nic.DeviceName != "" {
			expectedDeviceNames = append(expectedDeviceNames, nic.DeviceName)
		}
	}

	logger.Info("Expected", len(expectedDeviceNames), "RDMA devices based on shape:", expectedDeviceNames)

	// Step 3: Get OS device mapping using ibdev2netdev
	logger.Info("Step 3: Getting OS device mapping...")
	deviceMap, err := executor.GetIbdevToNetdevMap()
	if err != nil {
		logger.Error("Link Check: FAIL - Could not get device mapping:", err)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return fmt.Errorf("failed to get device mapping: %w", err)
	}

	// Step 4: Match expected device names with OS interface names and create failure results for missing devices
	logger.Info("Step 4: Matching expected devices with OS interfaces...")
	var interfacesToCheck []string
	var allResults []LinkCheckResult
	
	for _, expectedDevice := range expectedDeviceNames {
		found := false
		for deviceName, interfaceName := range deviceMap {
			if deviceName == expectedDevice {
				interfacesToCheck = append(interfacesToCheck, interfaceName)
				logger.Info("Found device", expectedDevice, "mapped to interface", interfaceName)
				found = true
				break
			}
		}
		
		if !found {
			logger.Error("Expected device", expectedDevice, "not found on system")
			// Create a failure result for the missing device
			failureResult := LinkCheckResult{
				Device:                      expectedDevice,
				LinkSpeed:                   fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				LinkState:                   fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				PhysicalState:               fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				LinkStatus:                  fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				EffectivePhysicalErrors:     fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				EffectivePhysicalBER:        fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				RawPhysicalErrorsPerLane:    fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
				RawPhysicalBER:              fmt.Sprintf("FAIL - Device %s not found", expectedDevice),
			}
			allResults = append(allResults, failureResult)
		}
	}

	if len(interfacesToCheck) == 0 && len(allResults) == 0 {
		errorStatement := "No expected RDMA devices found on the system"
		logger.Error("Link Check: FAIL -", errorStatement)
		err = fmt.Errorf(errorStatement)
		rep.AddLinkResult("FAIL", []LinkCheckResult{}, err)
		return err
	}

	logger.Info("Step 5: Found", len(interfacesToCheck), "interfaces to check:", interfacesToCheck)

	// Step 5: Check all matched interfaces
	logger.Info("Step 5: Checking all interfaces...")

	for _, interfaceName := range interfacesToCheck {
		var mlxlinkOutput string

		// Find the device for this interface
		deviceFound := false
		for device, netInterface := range deviceMap {
			if interfaceName == netInterface {
				// Run mlxlink for this device
				result, err := executor.RunMlxlink(device)
				if err != nil {
					logger.Error("Failed to run mlxlink for device", device, "(interface", interfaceName, "):", err)
					mlxlinkOutput = ""
				} else {
					mlxlinkOutput = result.Output
				}
				deviceFound = true
				break
			}
		}

		if !deviceFound {
			logger.Error("No device found for interface", interfaceName)
			mlxlinkOutput = ""
		}

		// Parse results
		linkResult, err := parseLinkResults(
			interfaceName,
			mlxlinkOutput,
			linkCheckTestConfig.ExpectedSpeed,
			linkCheckTestConfig.RawPhysicalErrorsPerLaneThreshold,
			linkCheckTestConfig.EffectivePhysicalErrorsThreshold,
		)
		if err != nil {
			logger.Errorf("Failed to parse link results for %s: %v", interfaceName, err)
			continue
		}

		allResults = append(allResults, *linkResult)
	}

	// Step 6: Report results
	logger.Info("Step 6: Reporting results...")
	if len(allResults) == 0 {
		logger.Error("Link Check: FAIL - No link results obtained")
		err = fmt.Errorf("no link results obtained")
		rep.AddLinkResult("FAIL", allResults, err)
		return err
	}

	// Check if all links passed
	allPassed := true
	for _, result := range allResults {
		if !strings.HasPrefix(result.LinkSpeed, "PASS") ||
			!strings.HasPrefix(result.LinkState, "PASS") ||
			!strings.HasPrefix(result.PhysicalState, "PASS") ||
			!strings.HasPrefix(result.LinkStatus, "PASS") ||
			!strings.HasPrefix(result.EffectivePhysicalBER, "PASS") ||
			!strings.HasPrefix(result.RawPhysicalBER, "PASS") ||
			strings.HasPrefix(result.EffectivePhysicalErrors, "FAIL") ||
			strings.HasPrefix(result.RawPhysicalErrorsPerLane, "WARN") {
			allPassed = false
			break
		}
	}

	if allPassed {
		logger.Info("Link Check: PASS - All links are healthy")
		rep.AddLinkResult("PASS", allResults, nil)
		return nil
	} else {
		logger.Error("Link Check: FAIL - Some links have issues")
		err = fmt.Errorf("some links have issues")
		rep.AddLinkResult("FAIL", allResults, err)
		return err
	}
}

// PrintLinkCheck prints a placeholder message for link check
func PrintLinkCheck() {
	logger.Info("Link Check: Checking RDMA link states and parameters...")
	logger.Info("Link Check: PASS - All RDMA links are healthy")
}