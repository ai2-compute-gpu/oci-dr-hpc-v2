package level1_tests

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Test structures (self-contained)
type testGPUClkCheckTestConfig struct {
	IsEnabled        bool `json:"enabled"`
	Shape            string `json:"shape"`
	ExpectedClkSpeed int  `json:"clock_speed"`
}

type testGPUClockResult struct {
	GPUID       string  `json:"gpu_id"`
	ClockSpeed  int     `json:"clock_speed"`
	IsValid     bool    `json:"is_valid"`
	RawSpeedStr string  `json:"raw_speed_str"`
}

// Test helper functions

func createTestGPUClockOutput(gpuCount int, clockSpeed int, unit string) string {
	var output strings.Builder
	
	for i := 0; i < gpuCount; i++ {
		if unit == "" {
			output.WriteString(fmt.Sprintf("%d\n", clockSpeed))
		} else {
			output.WriteString(fmt.Sprintf("%d %s\n", clockSpeed, unit))
		}
	}
	
	return strings.TrimSpace(output.String())
}

func createMixedGPUClockOutput() string {
	return `1980 MHz
1950 MHz  
1800 MHz
1700 MHz`
}

func createInvalidGPUClockOutput() string {
	return `invalid MHz
1980 MHz
not_a_number MHz
1950`
}

// Core functionality tests

func TestGPUClkCheckTestConfig(t *testing.T) {
	config := &testGPUClkCheckTestConfig{
		IsEnabled:        true,
		Shape:            "BM.GPU.H100.8",
		ExpectedClkSpeed: 1980,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected Shape 'BM.GPU.H100.8', got '%s'", config.Shape)
	}
	if config.ExpectedClkSpeed != 1980 {
		t.Errorf("Expected ExpectedClkSpeed 1980, got %d", config.ExpectedClkSpeed)
	}
}

func TestGPUClockResult(t *testing.T) {
	result := &testGPUClockResult{
		GPUID:       "0",
		ClockSpeed:  1980,
		IsValid:     true,
		RawSpeedStr: "1980 MHz",
	}

	if result.GPUID != "0" {
		t.Errorf("Expected GPUID '0', got '%s'", result.GPUID)
	}
	if result.ClockSpeed != 1980 {
		t.Errorf("Expected ClockSpeed 1980, got %d", result.ClockSpeed)
	}
	if !result.IsValid {
		t.Error("Expected IsValid to be true")
	}
	if result.RawSpeedStr != "1980 MHz" {
		t.Errorf("Expected RawSpeedStr '1980 MHz', got '%s'", result.RawSpeedStr)
	}
}

// Self-contained parsing function for testing
func testParseGPUClockSpeeds(output string) ([]testGPUClockResult, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty nvidia-smi clock output")
	}

	lines := strings.Split(output, "\n")
	var results []testGPUClockResult

	for gpuIndex, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result := testGPUClockResult{
			GPUID:       strconv.Itoa(gpuIndex),
			RawSpeedStr: line,
			IsValid:     false,
			ClockSpeed:  0,
		}

		// Extract numeric value from speed string (e.g., "1980 MHz" -> "1980")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			results = append(results, result)
			continue
		}

		currentSpeedStr := fields[0]
		currentSpeed, err := strconv.Atoi(currentSpeedStr)
		if err != nil {
			results = append(results, result)
			continue
		}

		result.ClockSpeed = currentSpeed
		result.IsValid = true
		results = append(results, result)
	}

	return results, nil
}

// Self-contained validation function for testing
func testValidateGPUClockResults(results []testGPUClockResult, expectedSpeed int) (bool, []string, string) {
	if len(results) == 0 {
		return false, nil, "no GPU clock speeds found"
	}

	// Calculate minimum acceptable speed (90% of expected - 10% tolerance)
	minAcceptableSpeed := expectedSpeed - int(float64(expectedSpeed)*0.10)
	
	var failedGPUs []string
	var minAllowedSpeed int = -1
	
	for _, result := range results {
		if !result.IsValid {
			failedGPUs = append(failedGPUs, result.GPUID)
			continue
		}
		
		// Check if speed is below minimum threshold
		if result.ClockSpeed < minAcceptableSpeed {
			failedGPUs = append(failedGPUs, result.GPUID)
		} else {
			// Speed is acceptable - track the minimum for reporting
			if minAllowedSpeed == -1 || result.ClockSpeed < minAllowedSpeed {
				minAllowedSpeed = result.ClockSpeed
			}
		}
	}

	if len(failedGPUs) > 0 {
		return false, failedGPUs, fmt.Sprintf("check GPU %s", strings.Join(failedGPUs, ","))
	}

	// All GPUs passed - format success message
	statusMsg := fmt.Sprintf("Expected %d, allowed %d", expectedSpeed, minAllowedSpeed)
	return true, nil, statusMsg
}

// Parsing tests

func TestParseGPUClockSpeeds(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		expectedGPUs int
		expectError bool
	}{
		{
			name:         "valid H100 output with MHz",
			output:       createTestGPUClockOutput(8, 1980, "MHz"),
			expectedGPUs: 8,
			expectError:  false,
		},
		{
			name:         "valid output without unit",
			output:       createTestGPUClockOutput(4, 1950, ""),
			expectedGPUs: 4,
			expectError:  false,
		},
		{
			name:         "single GPU output",
			output:       createTestGPUClockOutput(1, 2100, "MHz"),
			expectedGPUs: 1,
			expectError:  false,
		},
		{
			name:         "empty output",
			output:       "",
			expectedGPUs: 0,
			expectError:  true,
		},
		{
			name:         "mixed speed output",
			output:       createMixedGPUClockOutput(),
			expectedGPUs: 4,
			expectError:  false,
		},
		{
			name:         "invalid entries mixed with valid",
			output:       createInvalidGPUClockOutput(),
			expectedGPUs: 4,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := testParseGPUClockSpeeds(tt.output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(results) != tt.expectedGPUs {
					t.Errorf("Expected %d GPUs, got %d", tt.expectedGPUs, len(results))
				}

				// Validate parsed content
				for i, result := range results {
					t.Logf("GPU %d: Speed=%d, Valid=%v, Raw='%s'", 
						i, result.ClockSpeed, result.IsValid, result.RawSpeedStr)
					
					if result.GPUID == "" {
						t.Errorf("GPU %d has empty GPUID", i)
					}
					if result.RawSpeedStr == "" {
						t.Errorf("GPU %d has empty RawSpeedStr", i)
					}
				}
			}
		})
	}
}

func TestParseGPUClockSpeedsDetails(t *testing.T) {
	// Test with specific H100 output
	output := `1980 MHz
1950 MHz
1800 MHz`

	results, err := testParseGPUClockSpeeds(output)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 GPUs, got %d", len(results))
	}

	expectedSpeeds := []int{1980, 1950, 1800}
	for i, expected := range expectedSpeeds {
		if !results[i].IsValid {
			t.Errorf("GPU %d: Expected IsValid true, got false", i)
		}
		if results[i].ClockSpeed != expected {
			t.Errorf("GPU %d: Expected speed %d, got %d", i, expected, results[i].ClockSpeed)
		}
		if results[i].GPUID != strconv.Itoa(i) {
			t.Errorf("GPU %d: Expected GPUID '%d', got '%s'", i, i, results[i].GPUID)
		}
	}
}

// Validation tests

func TestValidateGPUClockResults(t *testing.T) {
	tests := []struct {
		name          string
		results       []testGPUClockResult
		expectedSpeed int
		shouldPass    bool
		expectedFails int
		description   string
	}{
		{
			name: "all GPUs pass at max speed",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
				{GPUID: "1", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
			},
			expectedSpeed: 1980,
			shouldPass:    true,
			expectedFails: 0,
			description:   "All GPUs at expected speed",
		},
		{
			name: "all GPUs pass within tolerance",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
				{GPUID: "1", ClockSpeed: 1850, IsValid: true, RawSpeedStr: "1850 MHz"},
				{GPUID: "2", ClockSpeed: 1800, IsValid: true, RawSpeedStr: "1800 MHz"},
			},
			expectedSpeed: 1980,
			shouldPass:    true,
			expectedFails: 0,
			description:   "All GPUs within 90% tolerance",
		},
		{
			name: "some GPUs fail below threshold",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
				{GPUID: "1", ClockSpeed: 1700, IsValid: true, RawSpeedStr: "1700 MHz"}, // Below 90%
			},
			expectedSpeed: 1980,
			shouldPass:    false,
			expectedFails: 1,
			description:   "One GPU below 90% threshold",
		},
		{
			name: "all GPUs fail",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1600, IsValid: true, RawSpeedStr: "1600 MHz"},
				{GPUID: "1", ClockSpeed: 1650, IsValid: true, RawSpeedStr: "1650 MHz"},
			},
			expectedSpeed: 1980,
			shouldPass:    false,
			expectedFails: 2,
			description:   "All GPUs below threshold",
		},
		{
			name: "invalid GPU results",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 0, IsValid: false, RawSpeedStr: "invalid"},
				{GPUID: "1", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
			},
			expectedSpeed: 1980,
			shouldPass:    false,
			expectedFails: 1,
			description:   "One invalid GPU result",
		},
		{
			name:          "empty results",
			results:       []testGPUClockResult{},
			expectedSpeed: 1980,
			shouldPass:    false,
			expectedFails: 0,
			description:   "No GPU results",
		},
		{
			name: "exactly at 90% threshold",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1782, IsValid: true, RawSpeedStr: "1782 MHz"}, // Exactly 90%
			},
			expectedSpeed: 1980,
			shouldPass:    true,
			expectedFails: 0,
			description:   "GPU at exactly 90% threshold",
		},
		{
			name: "just below 90% threshold",
			results: []testGPUClockResult{
				{GPUID: "0", ClockSpeed: 1781, IsValid: true, RawSpeedStr: "1781 MHz"}, // Just below 90%
			},
			expectedSpeed: 1980,
			shouldPass:    false,
			expectedFails: 1,
			description:   "GPU just below 90% threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, failedGPUs, statusMsg := testValidateGPUClockResults(tt.results, tt.expectedSpeed)

			if isValid != tt.shouldPass {
				t.Errorf("Expected validation result %v, got %v", tt.shouldPass, isValid)
			}

			if len(failedGPUs) != tt.expectedFails {
				t.Errorf("Expected %d failed GPUs, got %d", tt.expectedFails, len(failedGPUs))
			}

			if statusMsg == "" {
				t.Error("Expected non-empty status message")
			}

			t.Logf("Test %s: %s -> Valid=%v, FailedGPUs=%v, Status='%s'", 
				tt.name, tt.description, isValid, failedGPUs, statusMsg)

			// Validate failed GPU IDs exist in input
			if tt.expectedFails > 0 {
				for _, failedGPU := range failedGPUs {
					found := false
					for _, result := range tt.results {
						if result.GPUID == failedGPU {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Failed GPU '%s' not found in input results", failedGPU)
					}
				}
			}
		})
	}
}

// Edge cases and error handling

func TestParseGPUClockSpeedsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		expectError bool
		description string
	}{
		{
			name:        "whitespace only",
			output:      "   \n  \t  \n   ",
			expectError: true,
			description: "Only whitespace should be treated as empty",
		},
		{
			name:        "mixed valid and invalid lines",
			output:      "1980 MHz\ninvalid\n1950 MHz\n\n1800",
			expectError: false,
			description: "Should handle mixed valid/invalid lines gracefully",
		},
		{
			name:        "extremely high speeds",
			output:      "9999 MHz\n8888 MHz",
			expectError: false,
			description: "Should handle very high clock speeds",
		},
		{
			name:        "zero speeds",
			output:      "0 MHz\n0",
			expectError: false,
			description: "Should handle zero clock speeds",
		},
		{
			name:        "negative speeds (invalid)",
			output:      "-100 MHz\n1980 MHz",
			expectError: false,
			description: "Should handle negative numbers gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := testParseGPUClockSpeeds(tt.output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && results != nil {
				t.Logf("Test %s: %s -> Parsed %d results", tt.name, tt.description, len(results))
				for i, result := range results {
					t.Logf("  GPU %d: Speed=%d, Valid=%v", i, result.ClockSpeed, result.IsValid)
				}
			}
		})
	}
}

func TestGPUClockSpeedThresholds(t *testing.T) {
	output := `2000 MHz
1900 MHz
1800 MHz
1700 MHz`

	tests := []struct {
		name          string
		expectedSpeed int
		expectedPass  bool
		description   string
	}{
		{
			name:          "threshold_1980",
			expectedSpeed: 1980,
			expectedPass:  false, // 1700 is below 90% of 1980 (1782)
			description:   "1980 MHz threshold - should fail due to 1700 MHz",
		},
		{
			name:          "threshold_2000",
			expectedSpeed: 2000,
			expectedPass:  false, // 1700 is below 90% of 2000 (1800)
			description:   "2000 MHz threshold - should fail",
		},
		{
			name:          "threshold_1800",
			expectedSpeed: 1800,
			expectedPass:  true, // All speeds above 90% of 1800 (1620)
			description:   "1800 MHz threshold - should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := testParseGPUClockSpeeds(output)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			isValid, failedGPUs, statusMsg := testValidateGPUClockResults(results, tt.expectedSpeed)

			if isValid != tt.expectedPass {
				t.Errorf("Test %s: Expected %v, got %v", tt.description, tt.expectedPass, isValid)
			}

			t.Logf("Test %s: %s -> Valid=%v, FailedGPUs=%v, Status='%s'",
				tt.name, tt.description, isValid, failedGPUs, statusMsg)
		})
	}
}

// Performance tests

func TestLargeGPUClockOutput(t *testing.T) {
	// Test with 8 GPUs at 1980 MHz
	output := createTestGPUClockOutput(8, 1980, "MHz")

	results, err := testParseGPUClockSpeeds(output)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 8 {
		t.Errorf("Expected 8 GPUs, got %d", len(results))
	}

	// Validate each GPU result
	for i, result := range results {
		if !result.IsValid {
			t.Errorf("GPU %d: Expected IsValid true, got false", i)
		}
		if result.ClockSpeed != 1980 {
			t.Errorf("GPU %d: Expected speed 1980, got %d", i, result.ClockSpeed)
		}
		if result.GPUID != strconv.Itoa(i) {
			t.Errorf("GPU %d: Expected GPUID '%d', got '%s'", i, i, result.GPUID)
		}
	}

	// Test validation with all GPUs
	isValid, failedGPUs, statusMsg := testValidateGPUClockResults(results, 1980)
	if !isValid {
		t.Errorf("Expected validation to pass, got failed GPUs: %v", failedGPUs)
	}
	if len(failedGPUs) != 0 {
		t.Errorf("Expected 0 failed GPUs, got %d", len(failedGPUs))
	}
	if statusMsg == "" {
		t.Error("Expected non-empty status message")
	}

	t.Logf("Large output test: Valid=%v, Status='%s'", isValid, statusMsg)
}

func BenchmarkParseGPUClockSpeeds(b *testing.B) {
	output := createTestGPUClockOutput(8, 1980, "MHz")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testParseGPUClockSpeeds(output)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateGPUClockResults(b *testing.B) {
	results := []testGPUClockResult{
		{GPUID: "0", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
		{GPUID: "1", ClockSpeed: 1950, IsValid: true, RawSpeedStr: "1950 MHz"},
		{GPUID: "2", ClockSpeed: 1900, IsValid: true, RawSpeedStr: "1900 MHz"},
		{GPUID: "3", ClockSpeed: 1850, IsValid: true, RawSpeedStr: "1850 MHz"},
		{GPUID: "4", ClockSpeed: 1800, IsValid: true, RawSpeedStr: "1800 MHz"},
		{GPUID: "5", ClockSpeed: 1980, IsValid: true, RawSpeedStr: "1980 MHz"},
		{GPUID: "6", ClockSpeed: 1950, IsValid: true, RawSpeedStr: "1950 MHz"},
		{GPUID: "7", ClockSpeed: 1900, IsValid: true, RawSpeedStr: "1900 MHz"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = testValidateGPUClockResults(results, 1980)
	}
}

// Integration tests with actual validation function

func TestIntegrationWithValidateGPUClockSpeeds(t *testing.T) {
	tests := []struct {
		name           string
		clockSpeeds    []string
		expectedSpeed  int
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "integration test - all pass",
			clockSpeeds:    []string{"1980 MHz", "1950 MHz", "1900 MHz"},
			expectedSpeed:  1980,
			expectedStatus: "PASS",
			expectError:    false,
		},
		{
			name:           "integration test - some fail",
			clockSpeeds:    []string{"1980 MHz", "1700 MHz"},
			expectedSpeed:  1980,
			expectedStatus: "FAIL",
			expectError:    true,
		},
		{
			name:           "integration test - empty input",
			clockSpeeds:    []string{},
			expectedSpeed:  1980,
			expectedStatus: "FAIL",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, err := validateGPUClockSpeeds(tt.clockSpeeds, tt.expectedSpeed)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error %v, got error: %v", tt.expectError, err)
			}

			t.Logf("Integration test: Status=%s, Error=%v", status, err)
		})
	}
}