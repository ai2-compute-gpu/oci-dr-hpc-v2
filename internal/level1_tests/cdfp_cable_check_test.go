package level1_tests

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

// MockNvidiaSMIQuerier interface for dependency injection in tests
type MockCDFPNvidiaSMIQuerier interface {
	RunNvidiaSMIQuery(query string) *executor.NvidiaSMIResult
}

// mockCDFPQuerier implements MockCDFPNvidiaSMIQuerier for testing
type mockCDFPQuerier struct {
	pciResult    *executor.NvidiaSMIResult
	moduleResult *executor.NvidiaSMIResult
}

func (m *mockCDFPQuerier) RunNvidiaSMIQuery(query string) *executor.NvidiaSMIResult {
	switch query {
	case "pci.bus_id":
		return m.pciResult
	case "module_id":
		return m.moduleResult
	default:
		return &executor.NvidiaSMIResult{
			Available: false,
			Error:     "unknown query",
		}
	}
}

// parseGPUInfoWithQuerier is a testable version that accepts a querier
func parseGPUInfoWithQuerier(querier MockCDFPNvidiaSMIQuerier) ([]string, []string, error) {
	// Get PCI bus IDs
	pciResult := querier.RunNvidiaSMIQuery("pci.bus_id")
	if !pciResult.Available || pciResult.Error != "" {
		return nil, nil, fmt.Errorf("failed to get GPU PCI addresses: %s", pciResult.Error)
	}

	// Get module IDs
	moduleResult := querier.RunNvidiaSMIQuery("module_id")
	if !moduleResult.Available || moduleResult.Error != "" {
		return nil, nil, fmt.Errorf("failed to get GPU module IDs: %s", moduleResult.Error)
	}

	// Parse PCI addresses
	var pciAddresses []string
	for _, line := range strings.Split(pciResult.Output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			pciAddresses = append(pciAddresses, normalizePCIAddress(line))
		}
	}

	// Parse module IDs
	var moduleIDs []string
	for _, line := range strings.Split(moduleResult.Output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			moduleIDs = append(moduleIDs, line)
		}
	}

	if len(pciAddresses) != len(moduleIDs) {
		return nil, nil, fmt.Errorf("mismatch between PCI address count (%d) and module ID count (%d)", 
			len(pciAddresses), len(moduleIDs))
	}

	return pciAddresses, moduleIDs, nil
}

// TestNormalizePCIAddress tests the normalizePCIAddress function
func TestNormalizePCIAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard PCI address",
			input:    "00000000:0F:00.0",
			expected: "0000:0f:00.0",
		},
		{
			name:     "PCI address with 000000 prefix",
			input:    "000000000F:00.0",
			expected: "00000f:00.0",
		},
		{
			name:     "Already lowercase",
			input:    "00000000:2d:00.0",
			expected: "0000:2d:00.0",
		},
		{
			name:     "Mixed case",
			input:    "00000000:A8:00.0",
			expected: "0000:a8:00.0",
		},
		{
			name:     "Short format with 000000 prefix",
			input:    "0000008D:00.0",
			expected: "008d:00.0",
		},
		{
			name:     "No normalization needed",
			input:    "0001:2d:00.0",
			expected: "0001:2d:00.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePCIAddress(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePCIAddress(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseGPUInfoWithQuerier tests GPU information parsing with mock querier
func TestParseGPUInfoWithQuerier(t *testing.T) {
	tests := []struct {
		name             string
		pciResult        *executor.NvidiaSMIResult
		moduleResult     *executor.NvidiaSMIResult
		expectedPCIs      []string
		expectedModuleIDs []string
		expectError      bool
		errorContains    string
	}{
		{
			name: "Successful parsing",
			pciResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "00000000:0F:00.0\n00000000:2D:00.0\n00000000:44:00.0\n",
				Error:     "",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "0\n1\n2\n",
				Error:     "",
			},
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			expectedModuleIDs: []string{"0", "1", "2"},
			expectError:     false,
		},
		{
			name: "PCI query failed",
			pciResult: &executor.NvidiaSMIResult{
				Available: false,
				Output:    "",
				Error:     "nvidia-smi not found",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "0\n1\n2\n",
				Error:     "",
			},
			expectError:   true,
			errorContains: "failed to get GPU PCI addresses",
		},
		{
			name: "Module query failed",
			pciResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "00000000:0F:00.0\n00000000:2D:00.0\n",
				Error:     "",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: false,
				Output:    "",
				Error:     "module query failed",
			},
			expectError:   true,
			errorContains: "failed to get GPU module IDs",
		},
		{
			name: "Mismatch in count",
			pciResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "00000000:0F:00.0\n00000000:2D:00.0\n",
				Error:     "",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "0\n1\n2\n",
				Error:     "",
			},
			expectError:   true,
			errorContains: "mismatch between PCI address count",
		},
		{
			name: "Empty outputs",
			pciResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "",
				Error:     "",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "",
				Error:     "",
			},
			expectedPCIs:    nil,
			expectedModuleIDs: nil,
			expectError:     false,
		},
		{
			name: "With 000000 prefix normalization",
			pciResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "000000000F:00.0\n0000002D:00.0\n",
				Error:     "",
			},
			moduleResult: &executor.NvidiaSMIResult{
				Available: true,
				Output:    "0\n1\n",
				Error:     "",
			},
			expectedPCIs:    []string{"00000f:00.0", "002d:00.0"},
			expectedModuleIDs: []string{"0", "1"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			querier := &mockCDFPQuerier{
				pciResult:    tt.pciResult,
				moduleResult: tt.moduleResult,
			}

			actualPCIs, actualModuleIDs, err := parseGPUInfoWithQuerier(querier)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
				}
				if !reflect.DeepEqual(actualPCIs, tt.expectedPCIs) {
					t.Errorf("PCI addresses mismatch. Expected: %v, Got: %v", tt.expectedPCIs, actualPCIs)
				}
				if !reflect.DeepEqual(actualModuleIDs, tt.expectedModuleIDs) {
					t.Errorf("GPU module IDs mismatch. Expected: %v, Got: %v", tt.expectedModuleIDs, actualModuleIDs)
				}
			}
		})
	}
}

// TestValidateCDFPCables tests the CDFP cable validation logic
func TestValidateCDFPCables(t *testing.T) {
	tests := []struct {
		name            string
		expectedPCIs    []string
		expectedModuleIDs []string
		actualPCIs      []string
		actualModuleIDs   []string
		expectedStatus  string
		expectFailures  bool
		failureContains []string
	}{
		{
			name:            "Perfect match",
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			expectedModuleIDs: []string{"2", "4", "3"},
			actualPCIs:      []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			actualModuleIDs:   []string{"2", "4", "3"},
			expectedStatus:  "PASS",
			expectFailures:  false,
		},
		{
			name:            "Missing PCI address",
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			expectedModuleIDs: []string{"2", "4", "3"},
			actualPCIs:      []string{"0000:0f:00.0", "0000:2d:00.0"},
			actualModuleIDs:   []string{"2", "4"},
			expectedStatus:  "FAIL",
			expectFailures:  true,
			failureContains: []string{"Expected GPU with PCI Address 0000:44:00.0 not found"},
		},
		{
			name:            "Module ID mismatch",
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0"},
			expectedModuleIDs: []string{"2", "4"},
			actualPCIs:      []string{"0000:0f:00.0", "0000:2d:00.0"},
			actualModuleIDs:   []string{"2", "3"},
			expectedStatus:  "FAIL",
			expectFailures:  true,
			failureContains: []string{"Mismatch for PCI 0000:2d:00.0: Expected GPU module ID 4, found 3"},
		},
		{
			name:            "Multiple failures",
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			expectedModuleIDs: []string{"2", "4", "3"},
			actualPCIs:      []string{"0000:0f:00.0", "0000:2d:00.0"},
			actualModuleIDs:   []string{"2", "5"},
			expectedStatus:  "FAIL",
			expectFailures:  true,
			failureContains: []string{
				"Mismatch for PCI 0000:2d:00.0: Expected GPU module ID 4, found 5",
				"Expected GPU with PCI Address 0000:44:00.0 not found",
			},
		},
		{
			name:            "Empty configurations",
			expectedPCIs:    []string{},
			expectedModuleIDs: []string{},
			actualPCIs:      []string{},
			actualModuleIDs:   []string{},
			expectedStatus:  "PASS",
			expectFailures:  false,
		},
		{
			name:            "Actual has extra GPUs",
			expectedPCIs:    []string{"0000:0f:00.0", "0000:2d:00.0"},
			expectedModuleIDs: []string{"2", "4"},
			actualPCIs:      []string{"0000:0f:00.0", "0000:2d:00.0", "0000:44:00.0"},
			actualModuleIDs:   []string{"2", "4", "3"},
			expectedStatus:  "PASS",
			expectFailures:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCDFPCables(tt.expectedPCIs, tt.expectedModuleIDs, tt.actualPCIs, tt.actualModuleIDs)

			// Check status
			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}

			// Check failures
			if tt.expectFailures {
				if len(result.Failures) == 0 {
					t.Errorf("Expected failures but got none")
				} else {
					for _, expectedFailure := range tt.failureContains {
						found := false
						for _, actualFailure := range result.Failures {
							if strings.Contains(actualFailure, expectedFailure) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected failure containing '%s' not found in: %v", expectedFailure, result.Failures)
						}
					}
				}
			} else {
				if len(result.Failures) > 0 {
					t.Errorf("Expected no failures but got: %v", result.Failures)
				}
			}

			// Check mappings are populated
			if result.ExpectedMapping == nil {
				t.Errorf("Expected mapping should not be nil")
			}
			if result.ActualMapping == nil {
				t.Errorf("Actual mapping should not be nil")
			}

			// Verify expected mapping
			expectedMappingCount := len(tt.expectedPCIs)
			if len(result.ExpectedMapping) != expectedMappingCount {
				t.Errorf("Expected mapping count mismatch. Expected: %d, Got: %d", 
					expectedMappingCount, len(result.ExpectedMapping))
			}

			// Verify actual mapping
			actualMappingCount := len(tt.actualPCIs)
			if len(result.ActualMapping) != actualMappingCount {
				t.Errorf("Actual mapping count mismatch. Expected: %d, Got: %d", 
					actualMappingCount, len(result.ActualMapping))
			}
		})
	}
}

// TestCDFPCableCheckTestConfig tests the test configuration parsing
func TestCDFPCableCheckTestConfig(t *testing.T) {
	// Note: This would require mocking the test_limits.LoadTestLimits() function
	// For now, we'll test the structure
	config := &CDFPCableCheckTestConfig{
		IsEnabled:       true,
		ExpectedPCIIDs:  []string{"00000000:0f:00.0", "00000000:2d:00.0"},
		ExpectedModuleIDs: []string{"0", "1"},
	}

	if !config.IsEnabled {
		t.Errorf("Expected config to be enabled")
	}
	
	if len(config.ExpectedPCIIDs) != 2 {
		t.Errorf("Expected 2 PCI IDs, got %d", len(config.ExpectedPCIIDs))
	}
	
	if len(config.ExpectedModuleIDs) != 2 {
		t.Errorf("Expected 2 GPU module IDs, got %d", len(config.ExpectedModuleIDs))
	}
}

// TestCDFPCableCheckConfigWithShapes tests the configuration loading with shapes.json
func TestCDFPCableCheckConfigWithShapes(t *testing.T) {
	// Test getting configuration for H100 shape - this should work with shapes.json
	config, err := getCDFPCableCheckTestConfig("BM.GPU.H100.8")
	if err != nil {
		t.Errorf("Failed to get CDFP config for H100: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// Test should be enabled for H100 (as per test_limits.json)
	if !config.IsEnabled {
		t.Error("Expected CDFP test to be enabled for H100")
	}

	// Should have GPU PCI IDs and indices from shapes.json
	if len(config.ExpectedPCIIDs) == 0 {
		t.Error("Expected PCI IDs to be loaded from shapes.json")
	}

	if len(config.ExpectedModuleIDs) == 0 {
		t.Error("Expected GPU module IDs to be loaded from shapes.json")
	}

	// For H100, we should have 8 GPUs
	if len(config.ExpectedPCIIDs) != 8 {
		t.Errorf("Expected 8 GPU PCI IDs for H100, got %d", len(config.ExpectedPCIIDs))
	}

	if len(config.ExpectedModuleIDs) != 8 {
		t.Errorf("Expected 8 GPU module IDs for H100, got %d", len(config.ExpectedModuleIDs))
	}

	// Verify module IDs are the expected values for H100: 2, 4, 3, 1, 7, 5, 8, 6
	expectedModuleIDs := []string{"2", "4", "3", "1", "7", "5", "8", "6"}
	for i, expected := range expectedModuleIDs {
		if i < len(config.ExpectedModuleIDs) && config.ExpectedModuleIDs[i] != expected {
			t.Errorf("Expected GPU module ID %s at position %d, got %s", expected, i, config.ExpectedModuleIDs[i])
		}
	}

	// Verify PCI addresses have the correct normalized format
	for i, pci := range config.ExpectedPCIIDs {
		if !strings.HasPrefix(pci, "0000:") {
			t.Errorf("Expected PCI address at index %d to have normalized format, got %s", i, pci)
		}
	}

	t.Logf("Successfully loaded %d PCI addresses and %d module IDs from shapes.json for H100", 
		len(config.ExpectedPCIIDs), len(config.ExpectedModuleIDs))
}

// TestCDFPCableCheckConfigWithNonGPUShape tests config for a shape without GPUs
func TestCDFPCableCheckConfigWithNonGPUShape(t *testing.T) {
	// Test with a shape that doesn't have GPUs or doesn't exist in shapes.json
	config, err := getCDFPCableCheckTestConfig("BM.Standard.E5.48")
	if err != nil {
		t.Errorf("Unexpected error for non-GPU shape: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// Test should be disabled for non-GPU shapes (not in test_limits.json for CDFP)
	if config.IsEnabled {
		t.Error("Expected CDFP test to be disabled for non-GPU shape")
	}

	// Should have empty arrays
	if len(config.ExpectedPCIIDs) != 0 {
		t.Errorf("Expected empty PCI IDs for non-GPU shape, got %d", len(config.ExpectedPCIIDs))
	}

	if len(config.ExpectedModuleIDs) != 0 {
		t.Errorf("Expected empty module IDs for non-GPU shape, got %d", len(config.ExpectedModuleIDs))
	}
}

// TestCDFPCableCheckConfigWithB200 tests the configuration loading with B200 shape
func TestCDFPCableCheckConfigWithB200(t *testing.T) {
	// Test getting configuration for B200 shape - this should work with shapes.json
	config, err := getCDFPCableCheckTestConfig("BM.GPU.B200.8")
	if err != nil {
		t.Errorf("Failed to get CDFP config for B200: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// Test should be disabled for B200 (as per test_limits.json)
	if config.IsEnabled {
		t.Error("Expected CDFP test to be disabled for B200")
	}

	// Since test is disabled, arrays should be empty
	if len(config.ExpectedPCIIDs) != 0 {
		t.Errorf("Expected empty PCI IDs for disabled test, got %d", len(config.ExpectedPCIIDs))
	}

	if len(config.ExpectedModuleIDs) != 0 {
		t.Errorf("Expected empty module IDs for disabled test, got %d", len(config.ExpectedModuleIDs))
	}

	t.Logf("B200 CDFP test correctly disabled with empty configuration")
}

// TestCDFPCableCheckResult tests the result structure
func TestCDFPCableCheckResult(t *testing.T) {
	result := &CDFPCableCheckResult{
		Status:             "PASS",
		ExpectedMapping:    map[string]string{"0000:0f:00.0": "0"},
		ActualMapping:      map[string]string{"0000:0f:00.0": "0"},
		Failures:           []string{},
		Message:            "All CDFP cables correctly connected",
	}

	if result.Status != "PASS" {
		t.Errorf("Expected status PASS, got %s", result.Status)
	}

	if len(result.Failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(result.Failures))
	}

	if len(result.ExpectedMapping) != 1 {
		t.Errorf("Expected mapping should have 1 entry")
	}

	if result.ExpectedMapping["0000:0f:00.0"] != "0" {
		t.Errorf("Expected mapping value mismatch")
	}
}