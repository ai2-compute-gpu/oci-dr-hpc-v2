package level1_tests

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Test structures (self-contained)
type testNVLinkSpeedCheckTestConfig struct {
	IsEnabled     bool    `json:"enabled"`
	ExpectedSpeed float64 `json:"expected_speed"`
	ExpectedCount int     `json:"expected_count"`
}

type testNVLinkResult struct {
	GPUID     string           `json:"gpu_id"`
	LinkCount int              `json:"link_count"`
	Links     []testNVLinkInfo `json:"links"`
}

type testNVLinkInfo struct {
	LinkID   int     `json:"link_id"`
	Speed    float64 `json:"speed"`
	IsActive bool    `json:"is_active"`
}

// Test helper functions

func createTestNVLinkOutput(gpuCount int, linkCount int, speed float64) string {
	var output strings.Builder

	for i := 0; i < gpuCount; i++ {
		output.WriteString(fmt.Sprintf("GPU %d: NVIDIA H100 80GB HBM3 (UUID: GPU-12345678-1234-1234-1234-123456789012)\n", i))
		for j := 0; j < linkCount; j++ {
			output.WriteString(fmt.Sprintf("\tLink %d: %.1f GB/s\n", j, speed))
		}
		if i < gpuCount-1 {
			output.WriteString("\n")
		}
	}

	return output.String()
}

func createMixedNVLinkOutput() string {
	return `GPU 0: NVIDIA H100 80GB HBM3 (UUID: GPU-12345678-1234-1234-1234-123456789012)
	Link 0: 50.0 GB/s
	Link 1: 25.0 GB/s
	Link 2: 50.0 GB/s
	Link 3: N/A

GPU 1: NVIDIA H100 80GB HBM3 (UUID: GPU-87654321-4321-4321-4321-210987654321)
	Link 0: 50.0 GB/s
	Link 1: 50.0 GB/s
	Link 2: 50.0 GB/s
	Link 3: 50.0 GB/s`
}

// Core functionality tests

func TestNVLinkSpeedCheckTestConfig(t *testing.T) {
	config := &testNVLinkSpeedCheckTestConfig{
		IsEnabled:     true,
		ExpectedSpeed: 26.0,
		ExpectedCount: 18,
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.ExpectedSpeed != 26.0 {
		t.Errorf("Expected ExpectedSpeed 26.0, got %.1f", config.ExpectedSpeed)
	}
	if config.ExpectedCount != 18 {
		t.Errorf("Expected ExpectedCount 18, got %d", config.ExpectedCount)
	}
}

func TestNVLinkResult(t *testing.T) {
	result := &testNVLinkResult{
		GPUID:     "0",
		LinkCount: 2,
		Links: []testNVLinkInfo{
			{LinkID: 0, Speed: 50.0, IsActive: true},
			{LinkID: 1, Speed: 25.0, IsActive: true},
		},
	}

	if result.GPUID != "0" {
		t.Errorf("Expected GPUID '0', got '%s'", result.GPUID)
	}
	if result.LinkCount != 2 {
		t.Errorf("Expected LinkCount 2, got %d", result.LinkCount)
	}
	if len(result.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(result.Links))
	}
}

func TestNVLinkInfo(t *testing.T) {
	link := testNVLinkInfo{
		LinkID:   0,
		Speed:    50.0,
		IsActive: true,
	}

	if link.LinkID != 0 {
		t.Errorf("Expected LinkID 0, got %d", link.LinkID)
	}
	if link.Speed != 50.0 {
		t.Errorf("Expected Speed 50.0, got %.1f", link.Speed)
	}
	if !link.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

// Self-contained parsing function for testing
func testParseNVLinkOutput(output string, expectedSpeed float64) (map[string]*testNVLinkResult, error) {
	if output == "" {
		return nil, fmt.Errorf("empty nvidia-smi nvlink output")
	}

	results := make(map[string]*testNVLinkResult)
	lines := strings.Split(output, "\n")

	var currentGPU *testNVLinkResult
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
			currentGPU = &testNVLinkResult{
				GPUID:     gpuID,
				LinkCount: 0,
				Links:     []testNVLinkInfo{},
			}
			results[gpuID] = currentGPU
			continue
		}

		// Check for Link line
		if linkMatch := linkPattern.FindStringSubmatch(line); linkMatch != nil && currentGPU != nil {
			linkIDStr := linkMatch[1]
			speedStr := linkMatch[2]

			linkID, err := strconv.Atoi(linkIDStr)
			if err != nil {
				continue
			}

			speed, err := strconv.ParseFloat(speedStr, 64)
			if err != nil {
				continue
			}

			isActive := !strings.Contains(strings.ToLower(line), "inactive")
			isGoodSpeed := speed >= expectedSpeed

			linkInfo := testNVLinkInfo{
				LinkID:   linkID,
				Speed:    speed,
				IsActive: isActive,
			}

			currentGPU.Links = append(currentGPU.Links, linkInfo)

			// Count only active links that meet speed requirements
			if isActive && isGoodSpeed {
				currentGPU.LinkCount++
			}
			continue
		}

		// Check for unexpected output
		if !strings.Contains(line, "GPU") && !strings.Contains(line, "Link") && line != "" {
			return nil, fmt.Errorf("unexpected entry in nvidia-smi nvlink -s output: %s", line)
		}
	}

	return results, nil
}

// Self-contained validation function for testing
func testValidateNVLinkResults(results map[string]*testNVLinkResult, expectedCount int) (bool, []string) {
	var failedGPUs []string

	for gpuID, result := range results {
		if result.LinkCount != expectedCount {
			failedGPUs = append(failedGPUs, gpuID)
		}
	}

	return len(failedGPUs) == 0, failedGPUs
}

// Parsing tests

func TestParseNVLinkOutput(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		expectedSpeed float64
		expectedGPUs  int
		expectError   bool
	}{
		{
			name:          "valid H100 output",
			output:        createTestNVLinkOutput(2, 18, 50.0),
			expectedSpeed: 26.0,
			expectedGPUs:  2,
			expectError:   false,
		},
		{
			name:          "single GPU output",
			output:        createTestNVLinkOutput(1, 4, 25.781),
			expectedSpeed: 25.0,
			expectedGPUs:  1,
			expectError:   false,
		},
		{
			name:          "empty output",
			output:        "",
			expectedSpeed: 26.0,
			expectedGPUs:  0,
			expectError:   true,
		},
		{
			name:          "mixed speed output",
			output:        createMixedNVLinkOutput(),
			expectedSpeed: 26.0,
			expectedGPUs:  2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := testParseNVLinkOutput(tt.output, tt.expectedSpeed)

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
				for gpuID, result := range results {
					t.Logf("GPU %s: LinkCount=%d, Links=%d", gpuID, result.LinkCount, len(result.Links))
					if result.GPUID == "" {
						t.Errorf("GPU %s has empty GPUID", gpuID)
					}
					if result.Links == nil {
						t.Errorf("GPU %s has nil Links", gpuID)
					}
				}
			}
		})
	}
}

func TestParseNVLinkOutputDetails(t *testing.T) {
	// Test with specific H100 output
	output := `GPU 0: NVIDIA H100 80GB HBM3 (UUID: GPU-12345678-1234-1234-1234-123456789012)
	Link 0: 50.0 GB/s
	Link 1: 50.0 GB/s
	Link 2: 25.0 GB/s`

	results, err := testParseNVLinkOutput(output, 26.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 GPU, got %d", len(results))
	}

	gpu0, exists := results["0"]
	if !exists {
		t.Fatal("GPU 0 not found in results")
	}

	if gpu0.GPUID != "0" {
		t.Errorf("Expected GPUID '0', got '%s'", gpu0.GPUID)
	}

	if len(gpu0.Links) != 3 {
		t.Errorf("Expected 3 links, got %d", len(gpu0.Links))
	}

	// Check link count (only links >= 26.0 GB/s should count)
	expectedCount := 2 // Link 0 and 1 are >= 26.0, Link 2 is 25.0
	if gpu0.LinkCount != expectedCount {
		t.Errorf("Expected LinkCount %d, got %d", expectedCount, gpu0.LinkCount)
	}

	// Validate individual links
	for i, link := range gpu0.Links {
		if link.LinkID != i {
			t.Errorf("Link %d: Expected LinkID %d, got %d", i, i, link.LinkID)
		}
		if !link.IsActive {
			t.Errorf("Link %d: Expected IsActive true, got false", i)
		}
	}
}

// Validation tests

func TestValidateNVLinkResults(t *testing.T) {
	tests := []struct {
		name          string
		results       map[string]*testNVLinkResult
		expectedCount int
		shouldPass    bool
		expectedFails int
	}{
		{
			name: "all GPUs pass",
			results: map[string]*testNVLinkResult{
				"0": {GPUID: "0", LinkCount: 18, Links: []testNVLinkInfo{}},
				"1": {GPUID: "1", LinkCount: 18, Links: []testNVLinkInfo{}},
			},
			expectedCount: 18,
			shouldPass:    true,
			expectedFails: 0,
		},
		{
			name: "some GPUs fail",
			results: map[string]*testNVLinkResult{
				"0": {GPUID: "0", LinkCount: 18, Links: []testNVLinkInfo{}},
				"1": {GPUID: "1", LinkCount: 16, Links: []testNVLinkInfo{}},
			},
			expectedCount: 18,
			shouldPass:    false,
			expectedFails: 1,
		},
		{
			name: "all GPUs fail",
			results: map[string]*testNVLinkResult{
				"0": {GPUID: "0", LinkCount: 10, Links: []testNVLinkInfo{}},
				"1": {GPUID: "1", LinkCount: 12, Links: []testNVLinkInfo{}},
			},
			expectedCount: 18,
			shouldPass:    false,
			expectedFails: 2,
		},
		{
			name:          "empty results",
			results:       map[string]*testNVLinkResult{},
			expectedCount: 18,
			shouldPass:    true,
			expectedFails: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, failedGPUs := testValidateNVLinkResults(tt.results, tt.expectedCount)

			if isValid != tt.shouldPass {
				t.Errorf("Expected validation result %v, got %v", tt.shouldPass, isValid)
			}

			if len(failedGPUs) != tt.expectedFails {
				t.Errorf("Expected %d failed GPUs, got %d", tt.expectedFails, len(failedGPUs))
			}

			// Validate failed GPU IDs
			if tt.expectedFails > 0 {
				for _, failedGPU := range failedGPUs {
					if _, exists := tt.results[failedGPU]; !exists {
						t.Errorf("Failed GPU '%s' not found in input results", failedGPU)
					}
				}
			}
		})
	}
}

// Edge cases and error handling

func TestParseNVLinkOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		expectError bool
	}{
		{
			name:        "invalid format",
			output:      "invalid output format",
			expectError: true,
		},
		{
			name:        "invalid link ID",
			output:      "GPU 0: NVIDIA H100\n\tLink abc: 50.0 GB/s",
			expectError: false, // Should skip invalid lines
		},
		{
			name:        "invalid speed",
			output:      "GPU 0: NVIDIA H100\n\tLink 0: abc GB/s",
			expectError: false, // Should skip invalid lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testParseNVLinkOutput(tt.output, 26.0)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNVLinkSpeedThresholds(t *testing.T) {
	output := `GPU 0: NVIDIA H100 80GB HBM3 (UUID: GPU-12345678-1234-1234-1234-123456789012)
	Link 0: 30.0 GB/s
	Link 1: 25.0 GB/s
	Link 2: 26.0 GB/s
	Link 3: 50.0 GB/s`

	tests := []struct {
		name          string
		expectedSpeed float64
		expectedCount int
		description   string
	}{
		{
			name:          "threshold_26",
			expectedSpeed: 26.0,
			expectedCount: 3, // Links 0, 2, 3 meet threshold
			description:   "26.0 GB/s threshold",
		},
		{
			name:          "threshold_30",
			expectedSpeed: 30.0,
			expectedCount: 2, // Links 0, 3 meet threshold
			description:   "30.0 GB/s threshold",
		},
		{
			name:          "threshold_50",
			expectedSpeed: 50.0,
			expectedCount: 1, // Only Link 3 meets threshold
			description:   "50.0 GB/s threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := testParseNVLinkOutput(output, tt.expectedSpeed)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			gpu0, exists := results["0"]
			if !exists {
				t.Fatal("GPU 0 not found")
			}

			if gpu0.LinkCount != tt.expectedCount {
				t.Errorf("Test %s: Expected %d links meeting threshold, got %d",
					tt.description, tt.expectedCount, gpu0.LinkCount)
			}

			t.Logf("Test %s: %d links meet %.1f GB/s threshold",
				tt.description, gpu0.LinkCount, tt.expectedSpeed)
		})
	}
}

// Performance tests

func TestLargeNVLinkOutput(t *testing.T) {
	// Test with 8 GPUs, 18 links each
	output := createTestNVLinkOutput(8, 18, 50.0)

	results, err := testParseNVLinkOutput(output, 26.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 8 {
		t.Errorf("Expected 8 GPUs, got %d", len(results))
	}

	// Validate each GPU
	for i := 0; i < 8; i++ {
		gpuID := fmt.Sprintf("%d", i)
		gpu, exists := results[gpuID]
		if !exists {
			t.Errorf("GPU %s not found", gpuID)
			continue
		}

		if gpu.LinkCount != 18 {
			t.Errorf("GPU %s: Expected 18 links, got %d", gpuID, gpu.LinkCount)
		}

		if len(gpu.Links) != 18 {
			t.Errorf("GPU %s: Expected 18 link entries, got %d", gpuID, len(gpu.Links))
		}
	}
}

func BenchmarkParseNVLinkOutput(b *testing.B) {
	output := createTestNVLinkOutput(8, 18, 50.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testParseNVLinkOutput(output, 26.0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
