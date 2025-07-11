package level1_tests

import (
	"reflect"
	"strings"
	"testing"
)

// Test parseGIDIndexResults function
func TestParseGIDIndexResults(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Empty input",
			input:         []string{},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "Invalid GID index (non-numeric)",
			input: []string{
				"DEV	PORT	INDEX	GID",
				"---	----	-----	---",
				"mlx5_0   1     0     fe80:0000:0000:0000:0202:c9ff:fe00:0000",
				"mlx5_1   1     2     fe80:0000:0000:0000:0202:c9ff:fe00:0001",
				"n_gids_found=2",
				"",
			},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := parseGIDIndexResults(strings.Join(tt.input, "\n"))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

// Test checkGIDIndexes function
func TestCheckGIDIndexes(t *testing.T) {
	tests := []struct {
		name            string
		results         []GIDIndexResult
		expectedIndexes []int
		expectValid     bool
		expectError     bool
	}{
		{
			name: "All valid GID indexes",
			results: []GIDIndexResult{
				{Device: "mlx5_0", Port: "1", GIDIndex: 0, GIDValue: "fe80::1"},
				{Device: "mlx5_0", Port: "1", GIDIndex: 1, GIDValue: "fe80::2"},
				{Device: "mlx5_1", Port: "1", GIDIndex: 2, GIDValue: "fe80::3"},
			},
			expectedIndexes: []int{0, 1, 2, 3},
			expectValid:     true,
			expectError:     false,
		},
		{
			name: "Some invalid GID indexes",
			results: []GIDIndexResult{
				{Device: "mlx5_0", Port: "1", GIDIndex: 0, GIDValue: "fe80::1"},
				{Device: "mlx5_0", Port: "1", GIDIndex: 4, GIDValue: "fe80::2"},
			},
			expectedIndexes: []int{0, 1, 2, 3},
			expectValid:     false,
			expectError:     false,
		},
		{
			name:            "Empty results",
			results:         []GIDIndexResult{},
			expectedIndexes: []int{0, 1, 2, 3},
			expectValid:     false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _, err := checkGIDIndexes(tt.results, tt.expectedIndexes)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
		})
	}
}

// Test GIDIndexResult structure
func TestGIDIndexResult(t *testing.T) {
	result := GIDIndexResult{
		Device:   "mlx5_0",
		Port:     "1",
		GIDIndex: 0,
		GIDValue: "fe80:0000:0000:0000:0202:c9ff:fe00:0000",
	}

	if result.Device != "mlx5_0" {
		t.Error("Interface field mismatch")
	}
	if result.Port != "1" {
		t.Error("Port field mismatch")
	}
	if result.GIDIndex != 0 {
		t.Error("GIDIndex field mismatch")
	}
	if result.GIDValue != "fe80:0000:0000:0000:0202:c9ff:fe00:0000" {
		t.Error("GIDValue field mismatch")
	}
}

// Test show_gids output filtering
func TestShowGIDsOutputFiltering(t *testing.T) {
	tests := []struct {
		name          string
		rawOutput     string
		expectedLines []string
	}{
		{
			name: "Complete show_gids output",
			rawOutput: `DEV	PORT	GID_IDX	GID
------	----	-------	---
mlx5_0	1	0	fe80:0000:0000:0000:0202:c9ff:fe00:0000
mlx5_0	1	1	0000:0000:0000:0000:0000:ffff:c0a8:0101
-------------------------------------`,
			expectedLines: []string{
				"mlx5_0	1	0	fe80:0000:0000:0000:0202:c9ff:fe00:0000",
				"mlx5_0	1	1	0000:0000:0000:0000:0000:ffff:c0a8:0101",
			},
		},
		{
			name:          "Empty output",
			rawOutput:     "",
			expectedLines: []string{},
		},
		{
			name: "Header only",
			rawOutput: `DEV	PORT	GID_IDX	GID
------	----	-------	---
-------------------------------------`,
			expectedLines: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rawOutput == "" {
				if len(tt.expectedLines) != 0 {
					t.Errorf("Expected empty lines for empty output")
				}
				return
			}

			lines := strings.Split(tt.rawOutput, "\n")

			// Apply filtering: remove first 2 lines (header) and last line (footer)
			if len(lines) <= 3 {
				if len(tt.expectedLines) != 0 {
					t.Errorf("Expected empty lines for insufficient output")
				}
				return
			}

			filteredLines := lines[2 : len(lines)-1]

			if len(filteredLines) != len(tt.expectedLines) {
				t.Errorf("Expected %d lines, got %d", len(tt.expectedLines), len(filteredLines))
			}

			for i, line := range filteredLines {
				if i < len(tt.expectedLines) && line != tt.expectedLines[i] {
					t.Errorf("Line %d mismatch. Expected: %q, Got: %q", i, tt.expectedLines[i], line)
				}
			}
		})
	}
}

// Test threshold format handling
func TestThresholdFormatHandling(t *testing.T) {
	tests := []struct {
		name            string
		threshold       interface{}
		expectedIndexes []int
	}{
		{
			name:            "Array of interfaces (JSON format)",
			threshold:       []interface{}{float64(0), float64(1), float64(2), float64(3)},
			expectedIndexes: []int{0, 1, 2, 3},
		},
		{
			name:            "Array of integers",
			threshold:       []int{0, 1, 2, 3},
			expectedIndexes: []int{0, 1, 2, 3},
		},
		{
			name:            "Empty array",
			threshold:       []interface{}{},
			expectedIndexes: []int{0, 1, 2, 3}, // Default
		},
		{
			name:            "Unexpected format",
			threshold:       "invalid",
			expectedIndexes: []int{0, 1, 2, 3}, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate threshold processing logic
			config := &GIDIndexCheckTestConfig{
				IsEnabled:          true,
				ExpectedGIDIndexes: []int{0, 1, 2, 3}, // Default
			}

			switch v := tt.threshold.(type) {
			case []interface{}:
				var indexes []int
				for _, item := range v {
					if val, ok := item.(float64); ok {
						indexes = append(indexes, int(val))
					} else if val, ok := item.(int); ok {
						indexes = append(indexes, val)
					}
				}
				if len(indexes) > 0 {
					config.ExpectedGIDIndexes = indexes
				}
			case []int:
				config.ExpectedGIDIndexes = v
			}

			if !reflect.DeepEqual(config.ExpectedGIDIndexes, tt.expectedIndexes) {
				t.Errorf("Expected indexes %v, got %v", tt.expectedIndexes, config.ExpectedGIDIndexes)
			}
		})
	}
}

// Test boundary conditions for GID index values
func TestGIDIndexBoundaryConditions(t *testing.T) {
	tests := []struct {
		name          string
		gidIndex      int
		shouldBeValid bool
	}{
		{
			name:          "Minimum valid index",
			gidIndex:      0,
			shouldBeValid: true,
		},
		{
			name:          "Maximum valid index",
			gidIndex:      3,
			shouldBeValid: true,
		},
		{
			name:          "Just above maximum",
			gidIndex:      4,
			shouldBeValid: false,
		},
		{
			name:          "Negative index",
			gidIndex:      -1,
			shouldBeValid: false,
		},
	}

	expectedIndexes := []int{0, 1, 2, 3}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := []GIDIndexResult{
				{
					Device:   "mlx5_0",
					Port:     "1",
					GIDIndex: tt.gidIndex,
					GIDValue: "fe80::1",
				},
			}

			valid, _, err := checkGIDIndexes(results, expectedIndexes)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if valid != tt.shouldBeValid {
				t.Errorf("Expected valid=%v for GID index %d, got valid=%v", tt.shouldBeValid, tt.gidIndex, valid)
			}
		})
	}
}

// Test PrintGIDIndexCheck function
func TestPrintGIDIndexCheck(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintGIDIndexCheck panicked: %v", r)
		}
	}()

	PrintGIDIndexCheck()
}

// Test GIDIndexCheckTestConfig structure
func TestGIDIndexCheckTestConfigStructure(t *testing.T) {
	config := GIDIndexCheckTestConfig{
		IsEnabled:          true,
		ExpectedGIDIndexes: []int{0, 1, 2, 3},
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}

	expectedIndexes := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(config.ExpectedGIDIndexes, expectedIndexes) {
		t.Errorf("Expected indexes %v, got %v", expectedIndexes, config.ExpectedGIDIndexes)
	}
}
