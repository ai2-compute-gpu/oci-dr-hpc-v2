package level1_tests

import (
	"testing"
)

// TestSRAMErrorCounts tests the SRAMErrorCounts struct
func TestSRAMErrorCounts(t *testing.T) {
	sramError := SRAMErrorCounts{
		GPUIndex:      0,
		Uncorrectable: 5,
		Correctable:   100,
		ParityErrors:  2,
		SECDEDErrors:  3,
	}

	if sramError.GPUIndex != 0 {
		t.Errorf("Expected GPUIndex 0, got %d", sramError.GPUIndex)
	}
	if sramError.Uncorrectable != 5 {
		t.Errorf("Expected Uncorrectable 5, got %d", sramError.Uncorrectable)
	}
	if sramError.Correctable != 100 {
		t.Errorf("Expected Correctable 100, got %d", sramError.Correctable)
	}
	if sramError.ParityErrors != 2 {
		t.Errorf("Expected ParityErrors 2, got %d", sramError.ParityErrors)
	}
	if sramError.SECDEDErrors != 3 {
		t.Errorf("Expected SECDEDErrors 3, got %d", sramError.SECDEDErrors)
	}
}

// TestSRAMErrorSummary tests the SRAMErrorSummary struct
func TestSRAMErrorSummary(t *testing.T) {
	summary := SRAMErrorSummary{
		TotalGPUs:             8,
		GPUsWithUncorrectable: 1,
		GPUsWithCorrectable:   2,
		MaxUncorrectable:      10,
		MaxCorrectable:        500,
	}

	if summary.TotalGPUs != 8 {
		t.Errorf("Expected TotalGPUs 8, got %d", summary.TotalGPUs)
	}
	if summary.GPUsWithUncorrectable != 1 {
		t.Errorf("Expected GPUsWithUncorrectable 1, got %d", summary.GPUsWithUncorrectable)
	}
	if summary.GPUsWithCorrectable != 2 {
		t.Errorf("Expected GPUsWithCorrectable 2, got %d", summary.GPUsWithCorrectable)
	}
	if summary.MaxUncorrectable != 10 {
		t.Errorf("Expected MaxUncorrectable 10, got %d", summary.MaxUncorrectable)
	}
	if summary.MaxCorrectable != 500 {
		t.Errorf("Expected MaxCorrectable 500, got %d", summary.MaxCorrectable)
	}
}

// TestSRAMCheckTestConfig tests the SRAMCheckTestConfig struct
func TestSRAMCheckTestConfig(t *testing.T) {
	config := SRAMCheckTestConfig{
		IsEnabled:              true,
		UncorrectableThreshold: 5,
		CorrectableThreshold:   1000,
	}

	if !config.IsEnabled {
		t.Errorf("Expected IsEnabled true, got %v", config.IsEnabled)
	}
	if config.UncorrectableThreshold != 5 {
		t.Errorf("Expected UncorrectableThreshold 5, got %d", config.UncorrectableThreshold)
	}
	if config.CorrectableThreshold != 1000 {
		t.Errorf("Expected CorrectableThreshold 1000, got %d", config.CorrectableThreshold)
	}
}

// TestExtractErrorCounts tests the extractErrorCounts function
func TestExtractErrorCounts(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []GPUErrorDetails
	}{
		{
			name: "single GPU with parity and SEC-DED",
			input: []string{
				"  Parity                : 2",
				"  SEC-DED               : 3",
			},
			expected: []GPUErrorDetails{
				{Total: 5, Parity: 2, SECDED: 3},
			},
		},
		{
			name: "multiple GPUs",
			input: []string{
				"  Parity                : 1",
				"  SEC-DED               : 2",
				"  Parity                : 0",
				"  SEC-DED               : 1",
			},
			expected: []GPUErrorDetails{
				{Total: 3, Parity: 1, SECDED: 2},
				{Total: 1, Parity: 0, SECDED: 1},
			},
		},
		{
			name: "correctable errors only",
			input: []string{
				"Single Bit            : 25",
				"Single Bit            : 50",
			},
			expected: []GPUErrorDetails{
				{Total: 25, Parity: 0, SECDED: 0},
				{Total: 50, Parity: 0, SECDED: 0},
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []GPUErrorDetails{},
		},
		{
			name: "whitespace only",
			input: []string{
				"   ",
				"",
				"  ",
			},
			expected: []GPUErrorDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractErrorCounts(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing result at index %d", i)
					continue
				}

				actual := result[i]
				if actual.Total != expected.Total {
					t.Errorf("Result %d: Expected Total %d, got %d", i, expected.Total, actual.Total)
				}
				if actual.Parity != expected.Parity {
					t.Errorf("Result %d: Expected Parity %d, got %d", i, expected.Parity, actual.Parity)
				}
				if actual.SECDED != expected.SECDED {
					t.Errorf("Result %d: Expected SECDED %d, got %d", i, expected.SECDED, actual.SECDED)
				}
			}
		})
	}
}

// TestParseSRAMResults tests the parseSRAMResults function
func TestParseSRAMResults(t *testing.T) {
	tests := []struct {
		name                string
		uncorrectableOutput string
		correctableOutput   string
		expectedLen         int
		expectError         bool
	}{
		{
			name: "valid SRAM results",
			uncorrectableOutput: `  Parity                : 2
  SEC-DED               : 3`,
			correctableOutput: `Single Bit            : 25`,
			expectedLen:       1,
			expectError:       false,
		},
		{
			name: "multiple GPUs",
			uncorrectableOutput: `  Parity                : 1
  SEC-DED               : 2
  Parity                : 0
  SEC-DED               : 1`,
			correctableOutput: `Single Bit            : 25
Single Bit            : 50`,
			expectedLen: 2,
			expectError: false,
		},
		{
			name:                "empty outputs",
			uncorrectableOutput: "",
			correctableOutput:   "",
			expectedLen:         0,
			expectError:         true,
		},
		{
			name: "only uncorrectable errors",
			uncorrectableOutput: `  Parity                : 1
  SEC-DED               : 2`,
			correctableOutput: "",
			expectedLen:       1,
			expectError:       false,
		},
		{
			name:                "only correctable errors",
			uncorrectableOutput: "",
			correctableOutput:   `Single Bit            : 25`,
			expectedLen:         1,
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSRAMResults(tt.uncorrectableOutput, tt.correctableOutput)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d results, got %d", tt.expectedLen, len(result))
			}

			// Verify structure for successful cases
			if !tt.expectError && len(result) > 0 {
				for i, sramResult := range result {
					if sramResult.GPUIndex != i {
						t.Errorf("GPU %d: Expected GPUIndex %d, got %d", i, i, sramResult.GPUIndex)
					}
					t.Logf("GPU %d: Uncorrectable=%d, Correctable=%d, Parity=%d, SECDED=%d",
						i, sramResult.Uncorrectable, sramResult.Correctable,
						sramResult.ParityErrors, sramResult.SECDEDErrors)
				}
			}
		})
	}
}

// TestCheckSRAMThresholds tests the checkSRAMThresholds function
func TestCheckSRAMThresholds(t *testing.T) {
	config := &SRAMCheckTestConfig{
		UncorrectableThreshold: 5,
		CorrectableThreshold:   100,
	}

	tests := []struct {
		name            string
		results         []SRAMErrorCounts
		expectedStatus  string
		expectedSummary SRAMErrorSummary
	}{
		{
			name: "all within thresholds",
			results: []SRAMErrorCounts{
				{GPUIndex: 0, Uncorrectable: 2, Correctable: 50},
				{GPUIndex: 1, Uncorrectable: 1, Correctable: 25},
			},
			expectedStatus: "PASS",
			expectedSummary: SRAMErrorSummary{
				TotalGPUs:             2,
				GPUsWithUncorrectable: 0,
				GPUsWithCorrectable:   0,
				MaxUncorrectable:      2,
				MaxCorrectable:        50,
			},
		},
		{
			name: "correctable threshold exceeded",
			results: []SRAMErrorCounts{
				{GPUIndex: 0, Uncorrectable: 2, Correctable: 150},
				{GPUIndex: 1, Uncorrectable: 1, Correctable: 25},
			},
			expectedStatus: "WARN",
			expectedSummary: SRAMErrorSummary{
				TotalGPUs:             2,
				GPUsWithUncorrectable: 0,
				GPUsWithCorrectable:   1,
				MaxUncorrectable:      2,
				MaxCorrectable:        150,
			},
		},
		{
			name: "uncorrectable threshold exceeded",
			results: []SRAMErrorCounts{
				{GPUIndex: 0, Uncorrectable: 10, Correctable: 50},
				{GPUIndex: 1, Uncorrectable: 1, Correctable: 25},
			},
			expectedStatus: "FAIL",
			expectedSummary: SRAMErrorSummary{
				TotalGPUs:             2,
				GPUsWithUncorrectable: 1,
				GPUsWithCorrectable:   0,
				MaxUncorrectable:      10,
				MaxCorrectable:        50,
			},
		},
		{
			name: "both thresholds exceeded",
			results: []SRAMErrorCounts{
				{GPUIndex: 0, Uncorrectable: 10, Correctable: 150},
				{GPUIndex: 1, Uncorrectable: 1, Correctable: 25},
			},
			expectedStatus: "FAIL",
			expectedSummary: SRAMErrorSummary{
				TotalGPUs:             2,
				GPUsWithUncorrectable: 1,
				GPUsWithCorrectable:   1,
				MaxUncorrectable:      10,
				MaxCorrectable:        150,
			},
		},
		{
			name:            "empty results",
			results:         []SRAMErrorCounts{},
			expectedStatus:  "FAIL",
			expectedSummary: SRAMErrorSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, summary := checkSRAMThresholds(tt.results, config)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			if summary.TotalGPUs != tt.expectedSummary.TotalGPUs {
				t.Errorf("Expected TotalGPUs %d, got %d", tt.expectedSummary.TotalGPUs, summary.TotalGPUs)
			}
			if summary.GPUsWithUncorrectable != tt.expectedSummary.GPUsWithUncorrectable {
				t.Errorf("Expected GPUsWithUncorrectable %d, got %d", tt.expectedSummary.GPUsWithUncorrectable, summary.GPUsWithUncorrectable)
			}
			if summary.GPUsWithCorrectable != tt.expectedSummary.GPUsWithCorrectable {
				t.Errorf("Expected GPUsWithCorrectable %d, got %d", tt.expectedSummary.GPUsWithCorrectable, summary.GPUsWithCorrectable)
			}
			if summary.MaxUncorrectable != tt.expectedSummary.MaxUncorrectable {
				t.Errorf("Expected MaxUncorrectable %d, got %d", tt.expectedSummary.MaxUncorrectable, summary.MaxUncorrectable)
			}
			if summary.MaxCorrectable != tt.expectedSummary.MaxCorrectable {
				t.Errorf("Expected MaxCorrectable %d, got %d", tt.expectedSummary.MaxCorrectable, summary.MaxCorrectable)
			}
		})
	}
}

// TestGPUErrorDetails tests the GPUErrorDetails struct
func TestGPUErrorDetails(t *testing.T) {
	details := GPUErrorDetails{
		Total:  10,
		Parity: 3,
		SECDED: 7,
	}

	if details.Total != 10 {
		t.Errorf("Expected Total 10, got %d", details.Total)
	}
	if details.Parity != 3 {
		t.Errorf("Expected Parity 3, got %d", details.Parity)
	}
	if details.SECDED != 7 {
		t.Errorf("Expected SECDED 7, got %d", details.SECDED)
	}
}

// TestSRAMCheckResult tests the SRAMCheckResult struct
func TestSRAMCheckResult(t *testing.T) {
	result := SRAMCheckResult{
		Status: "PASS",
		GPUResults: []SRAMErrorCounts{
			{GPUIndex: 0, Uncorrectable: 1, Correctable: 10},
		},
		Summary: SRAMErrorSummary{
			TotalGPUs:        1,
			MaxUncorrectable: 1,
			MaxCorrectable:   10,
		},
	}

	if result.Status != "PASS" {
		t.Errorf("Expected Status 'PASS', got '%s'", result.Status)
	}
	if len(result.GPUResults) != 1 {
		t.Errorf("Expected 1 GPU result, got %d", len(result.GPUResults))
	}
	if result.Summary.TotalGPUs != 1 {
		t.Errorf("Expected TotalGPUs 1, got %d", result.Summary.TotalGPUs)
	}
}

// BenchmarkExtractErrorCounts benchmarks the extractErrorCounts function
func BenchmarkExtractErrorCounts(b *testing.B) {
	input := []string{
		"  Parity                : 2",
		"  SEC-DED               : 3",
		"  Parity                : 1",
		"  SEC-DED               : 4",
		"Single Bit            : 25",
		"Single Bit            : 50",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractErrorCounts(input)
	}
}

// BenchmarkParseSRAMResults benchmarks the parseSRAMResults function
func BenchmarkParseSRAMResults(b *testing.B) {
	uncorrectableOutput := `  Parity                : 2
  SEC-DED               : 3
  Parity                : 1
  SEC-DED               : 4`
	correctableOutput := `Single Bit            : 25
Single Bit            : 50`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseSRAMResults(uncorrectableOutput, correctableOutput)
	}
}

// BenchmarkCheckSRAMThresholds benchmarks the checkSRAMThresholds function
func BenchmarkCheckSRAMThresholds(b *testing.B) {
	config := &SRAMCheckTestConfig{
		UncorrectableThreshold: 5,
		CorrectableThreshold:   100,
	}
	results := []SRAMErrorCounts{
		{GPUIndex: 0, Uncorrectable: 2, Correctable: 50},
		{GPUIndex: 1, Uncorrectable: 1, Correctable: 25},
		{GPUIndex: 2, Uncorrectable: 3, Correctable: 75},
		{GPUIndex: 3, Uncorrectable: 0, Correctable: 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkSRAMThresholds(results, config)
	}
}
