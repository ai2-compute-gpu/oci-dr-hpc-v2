package level1_tests

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

// MockNvidiaSMIQuerier interface for dependency injection in tests
type MockNvidiaSMIQuerier interface {
	RunNvidiaSMIQuery(query string) *executor.NvidiaSMIResult
}

// mockQuerier implements MockNvidiaSMIQuerier for testing
type mockQuerier struct {
	result *executor.NvidiaSMIResult
}

func (m *mockQuerier) RunNvidiaSMIQuery(query string) *executor.NvidiaSMIResult {
	return m.result
}

// getGPUModeInfoWithQuerier is a testable version that accepts a querier
func getGPUModeInfoWithQuerier(querier MockNvidiaSMIQuerier) ([]GPUModeInfo, error) {
	// Use the querier to get GPU index and MIG mode
	result := querier.RunNvidiaSMIQuery("index,mig.mode.current")
	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Parse the output
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return nil, fmt.Errorf("no GPU mode information returned from nvidia-smi")
	}

	return parseGPUModeInfo(output)
}

// TestParseGPUModeInfo tests the parseGPUModeInfo function
func TestParseGPUModeInfo(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      []GPUModeInfo
		expectError   bool
		errorContains string
	}{
		{
			name:  "Valid single GPU with Disabled mode",
			input: "0, Disabled",
			expected: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
			},
			expectError: false,
		},
		{
			name:  "Valid multiple GPUs with different modes",
			input: "0, Disabled\n1, Enabled\n2, N/A",
			expected: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "Enabled"},
				{Index: "2", Mode: "N/A"},
			},
			expectError: false,
		},
		{
			name:  "Valid GPUs with whitespace",
			input: "  0  ,  Disabled  \n  1  ,  Enabled  ",
			expected: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "Enabled"},
			},
			expectError: false,
		},
		{
			name:          "Invalid GPU index (non-numeric)",
			input:         "abc, Disabled",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid GPU index format",
		},
		{
			name:          "Invalid mode format",
			input:         "0, InvalidMode",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid GPU mode format",
		},
		{
			name:  "Empty lines should be filtered",
			input: "0, Disabled\n\n1, Enabled\n\n",
			expected: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "Enabled"},
			},
			expectError: false,
		},
		{
			name:        "Insufficient fields in line",
			input:       "0",
			expected:    nil,
			expectError: false, // Invalid lines are skipped, not errored
		},
		{
			name:        "Empty input",
			input:       "",
			expected:    []GPUModeInfo{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGPUModeInfo(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %s", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

// TestCheckGPUModeResults tests the checkGPUModeResults function
func TestCheckGPUModeResults(t *testing.T) {
	tests := []struct {
		name         string
		gpuModes     []GPUModeInfo
		allowedModes []string
		expected     *GPUModeResult
	}{
		{
			name: "All GPUs have allowed modes",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "N/A"},
			},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "PASS",
				Message:           "PASS",
				EnabledGPUIndexes: []string{},
			},
		},
		{
			name: "Some GPUs have disallowed modes",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "Enabled"},
				{Index: "2", Mode: "N/A"},
			},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "FAIL",
				Message:           "FAIL - Invalid GPU modes detected on GPUs 1",
				EnabledGPUIndexes: []string{"1"},
			},
		},
		{
			name: "Multiple GPUs with disallowed modes",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "Enabled"},
				{Index: "1", Mode: "Enabled"},
				{Index: "2", Mode: "Disabled"},
			},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "FAIL",
				Message:           "FAIL - Invalid GPU modes detected on GPUs 0,1",
				EnabledGPUIndexes: []string{"0", "1"},
			},
		},
		{
			name: "Case insensitive mode checking",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "disabled"},
				{Index: "1", Mode: "ENABLED"},
			},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "FAIL",
				Message:           "FAIL - Invalid GPU modes detected on GPUs 1",
				EnabledGPUIndexes: []string{"1"},
			},
		},
		{
			name: "Unknown mode is treated as invalid",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "Unknown"},
				{Index: "1", Mode: "Disabled"},
			},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "FAIL",
				Message:           "FAIL - Invalid GPU modes detected on GPUs 0",
				EnabledGPUIndexes: []string{"0"},
			},
		},
		{
			name:         "Empty GPU list",
			gpuModes:     []GPUModeInfo{},
			allowedModes: []string{"Disabled", "N/A"},
			expected: &GPUModeResult{
				Status:            "PASS",
				Message:           "PASS",
				EnabledGPUIndexes: []string{},
			},
		},
		{
			name: "All modes allowed including Enabled",
			gpuModes: []GPUModeInfo{
				{Index: "0", Mode: "Enabled"},
				{Index: "1", Mode: "Disabled"},
			},
			allowedModes: []string{"Disabled", "N/A", "Enabled"},
			expected: &GPUModeResult{
				Status:            "PASS",
				Message:           "PASS",
				EnabledGPUIndexes: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkGPUModeResults(tt.gpuModes, tt.allowedModes)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

// TestGetGPUModeInfo tests the getGPUModeInfo function with mocked nvidia-smi
func TestGetGPUModeInfo(t *testing.T) {
	tests := []struct {
		name          string
		mockOutput    string
		mockAvailable bool
		mockError     string
		expected      []GPUModeInfo
		expectError   bool
		errorContains string
	}{
		{
			name:          "Successful GPU mode query",
			mockOutput:    "0, Disabled\n1, Enabled",
			mockAvailable: true,
			mockError:     "",
			expected: []GPUModeInfo{
				{Index: "0", Mode: "Disabled"},
				{Index: "1", Mode: "Enabled"},
			},
			expectError: false,
		},
		{
			name:          "nvidia-smi not available",
			mockOutput:    "",
			mockAvailable: false,
			mockError:     "nvidia-smi not found",
			expected:      nil,
			expectError:   true,
			errorContains: "nvidia-smi not available",
		},
		{
			name:          "Empty output from nvidia-smi",
			mockOutput:    "",
			mockAvailable: true,
			mockError:     "",
			expected:      nil,
			expectError:   true,
			errorContains: "no GPU mode information returned",
		},
		{
			name:          "Invalid output format",
			mockOutput:    "invalid,format,data",
			mockAvailable: true,
			mockError:     "",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid GPU index format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock querier
			mockQuerier := &mockQuerier{
				result: &executor.NvidiaSMIResult{
					Available: tt.mockAvailable,
					Output:    tt.mockOutput,
					Error:     tt.mockError,
				},
			}

			result, err := getGPUModeInfoWithQuerier(mockQuerier)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %s", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

// TestPrintGPUModeCheck tests the PrintGPUModeCheck function
func TestPrintGPUModeCheck(t *testing.T) {
	// This is a simple test to ensure the function doesn't panic
	// In a real implementation, you might want to capture the output
	t.Run("PrintGPUModeCheck does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintGPUModeCheck panicked: %v", r)
			}
		}()
		PrintGPUModeCheck()
	})
}

// BenchmarkParseGPUModeInfo benchmarks the parseGPUModeInfo function
func BenchmarkParseGPUModeInfo(b *testing.B) {
	input := "0, Disabled\n1, Enabled\n2, N/A\n3, Disabled\n4, Enabled\n5, N/A\n6, Disabled\n7, Enabled"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseGPUModeInfo(input)
		if err != nil {
			b.Fatalf("Unexpected error in benchmark: %v", err)
		}
	}
}

// BenchmarkCheckGPUModeResults benchmarks the checkGPUModeResults function
func BenchmarkCheckGPUModeResults(b *testing.B) {
	gpuModes := []GPUModeInfo{
		{Index: "0", Mode: "Disabled"},
		{Index: "1", Mode: "Enabled"},
		{Index: "2", Mode: "N/A"},
		{Index: "3", Mode: "Disabled"},
		{Index: "4", Mode: "Enabled"},
		{Index: "5", Mode: "N/A"},
		{Index: "6", Mode: "Disabled"},
		{Index: "7", Mode: "Enabled"},
	}
	allowedModes := []string{"Disabled", "N/A"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := checkGPUModeResults(gpuModes, allowedModes)
		if result == nil {
			b.Fatal("Unexpected nil result in benchmark")
		}
	}
}

// TestIntegrationGetGPUModeInfo tests the actual getGPUModeInfo function
// This test will only run if nvidia-smi is available and should be marked as integration test
func TestIntegrationGetGPUModeInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Integration test with real nvidia-smi", func(t *testing.T) {
		// This test will attempt to call the real getGPUModeInfo function
		// It may fail if nvidia-smi is not available, which is expected
		result, err := getGPUModeInfo()

		// We don't assert specific values since this depends on the actual system
		// We just check that the function doesn't panic and returns consistent data
		if err != nil {
			// It's okay if nvidia-smi is not available in test environment
			t.Logf("nvidia-smi not available in test environment: %v", err)
			return
		}

		// If no error, result should not be nil
		if result == nil {
			t.Error("Expected non-nil result when no error occurred")
		}

		// Basic validation of returned data structure
		for i, gpu := range result {
			if gpu.Index == "" {
				t.Errorf("GPU %d has empty index", i)
			}
			if gpu.Mode == "" {
				t.Errorf("GPU %d has empty mode", i)
			}
		}
	})
}

// TestGPUModeResultStruct tests the GPUModeResult struct
func TestGPUModeResultStruct(t *testing.T) {
	t.Run("GPUModeResult struct creation", func(t *testing.T) {
		result := &GPUModeResult{
			Status:            "FAIL",
			Message:           "Test message",
			EnabledGPUIndexes: []string{"0", "1"},
		}

		if result.Status != "FAIL" {
			t.Errorf("Expected Status to be 'FAIL', got '%s'", result.Status)
		}
		if result.Message != "Test message" {
			t.Errorf("Expected Message to be 'Test message', got '%s'", result.Message)
		}
		if len(result.EnabledGPUIndexes) != 2 {
			t.Errorf("Expected 2 enabled GPU indexes, got %d", len(result.EnabledGPUIndexes))
		}
	})
}

// TestGPUModeInfoStruct tests the GPUModeInfo struct
func TestGPUModeInfoStruct(t *testing.T) {
	t.Run("GPUModeInfo struct creation", func(t *testing.T) {
		info := &GPUModeInfo{
			Index: "0",
			Mode:  "Disabled",
		}

		if info.Index != "0" {
			t.Errorf("Expected Index to be '0', got '%s'", info.Index)
		}
		if info.Mode != "Disabled" {
			t.Errorf("Expected Mode to be 'Disabled', got '%s'", info.Mode)
		}
	})
}

// Helper function to create test GPU mode info slices
func createTestGPUModes(configs map[string]string) []GPUModeInfo {
	var modes []GPUModeInfo
	for index, mode := range configs {
		modes = append(modes, GPUModeInfo{
			Index: index,
			Mode:  mode,
		})
	}
	return modes
}

// TestCreateTestGPUModes tests the helper function
func TestCreateTestGPUModes(t *testing.T) {
	t.Run("Helper function creates correct GPU modes", func(t *testing.T) {
		configs := map[string]string{
			"0": "Disabled",
			"1": "Enabled",
		}

		result := createTestGPUModes(configs)

		if len(result) != 2 {
			t.Errorf("Expected 2 GPU modes, got %d", len(result))
		}

		// Note: map iteration order is not guaranteed, so we check if both entries exist
		foundModes := make(map[string]string)
		for _, mode := range result {
			foundModes[mode.Index] = mode.Mode
		}

		if foundModes["0"] != "Disabled" {
			t.Errorf("Expected GPU 0 to have mode 'Disabled', got '%s'", foundModes["0"])
		}
		if foundModes["1"] != "Enabled" {
			t.Errorf("Expected GPU 1 to have mode 'Enabled', got '%s'", foundModes["1"])
		}
	})
}
