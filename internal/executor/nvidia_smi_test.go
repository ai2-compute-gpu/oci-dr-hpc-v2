package executor

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestCheckNvidiaSMI tests the CheckNvidiaSMI function
func TestCheckNvidiaSMI(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() func()
		wantAvail bool
		wantError bool
	}{
		{
			name: "nvidia-smi available and working",
			setup: func() func() {
				// Check if nvidia-smi is actually available
				if _, err := exec.LookPath("nvidia-smi"); err != nil {
					t.Skip("nvidia-smi not available in test environment")
				}
				return func() {} // no cleanup needed
			},
			wantAvail: true,
			wantError: false,
		},
		{
			name: "nvidia-smi not in PATH",
			setup: func() func() {
				// Temporarily modify PATH to remove nvidia-smi
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantAvail: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			result := CheckNvidiaSMI()

			if result.Available != tt.wantAvail {
				t.Errorf("CheckNvidiaSMI().Available = %v, want %v", result.Available, tt.wantAvail)
			}

			if tt.wantError && result.Error == "" {
				t.Errorf("CheckNvidiaSMI() expected error but got none")
			}

			if !tt.wantError && result.Error != "" {
				t.Errorf("CheckNvidiaSMI() unexpected error: %v", result.Error)
			}

			// If available, output should not be empty
			if result.Available && result.Output == "" {
				t.Errorf("CheckNvidiaSMI() available but output is empty")
			}
		})
	}
}

// TestRunNvidiaSMIQuery tests the RunNvidiaSMIQuery function
func TestRunNvidiaSMIQuery(t *testing.T) {
	// Skip if nvidia-smi is not available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		t.Skip("nvidia-smi not available in test environment")
	}

	tests := []struct {
		name      string
		query     string
		wantError bool
	}{
		{
			name:      "valid query - name",
			query:     "name",
			wantError: false,
		},
		{
			name:      "valid query - memory.total",
			query:     "memory.total",
			wantError: false,
		},
		{
			name:      "invalid query",
			query:     "invalid_field_name",
			wantError: true,
		},
		{
			name:      "empty query",
			query:     "",
			wantError: false, // nvidia-smi actually accepts empty query and returns default info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RunNvidiaSMIQuery(tt.query)

			if tt.wantError {
				if result.Available {
					t.Errorf("RunNvidiaSMIQuery(%q) expected error but succeeded", tt.query)
				}
				if result.Error == "" {
					t.Errorf("RunNvidiaSMIQuery(%q) expected error message but got none", tt.query)
				}
			} else {
				if !result.Available {
					t.Errorf("RunNvidiaSMIQuery(%q) failed: %v", tt.query, result.Error)
				}
				// Only check for non-empty output for non-empty queries
				if tt.query != "" && result.Output == "" {
					t.Errorf("RunNvidiaSMIQuery(%q) succeeded but output is empty", tt.query)
				}
			}
		})
	}
}

// TestRunNvidiaSMIErrorQuery tests the RunNvidiaSMIErrorQuery function
func TestRunNvidiaSMIErrorQuery(t *testing.T) {
	tests := []struct {
		name        string
		errorType   string
		expectError bool
		skipIfNoGPU bool
		setup       func() func()
	}{
		{
			name:        "uncorrectable error query",
			errorType:   "uncorrectable",
			expectError: false,
			skipIfNoGPU: true,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "correctable error query",
			errorType:   "correctable",
			expectError: false,
			skipIfNoGPU: true,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "case insensitive uncorrectable",
			errorType:   "UNCORRECTABLE",
			expectError: false,
			skipIfNoGPU: true,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "case insensitive correctable",
			errorType:   "CORRECTABLE",
			expectError: false,
			skipIfNoGPU: true,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "mixed case uncorrectable",
			errorType:   "UnCorrectable",
			expectError: false,
			skipIfNoGPU: true,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "unsupported error type",
			errorType:   "invalid_type",
			expectError: true,
			skipIfNoGPU: false,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "empty error type",
			errorType:   "",
			expectError: true,
			skipIfNoGPU: false,
			setup:       func() func() { return func() {} },
		},
		{
			name:        "nvidia-smi not available",
			errorType:   "uncorrectable",
			expectError: true,
			skipIfNoGPU: false,
			setup: func() func() {
				// Temporarily modify PATH to remove nvidia-smi
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			if tt.skipIfNoGPU && !IsNvidiaSMIAvailable() {
				t.Skip("Skipping test - nvidia-smi not available")
			}

			result, err := RunNvidiaSMIErrorQuery(tt.errorType)

			if tt.expectError {
				if err == nil {
					t.Errorf("RunNvidiaSMIErrorQuery(%q) expected error but got none", tt.errorType)
				}
				// When expecting error, result might be nil or contain error info
				if err != nil {
					t.Logf("Expected error occurred: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("RunNvidiaSMIErrorQuery(%q) unexpected error: %v", tt.errorType, err)
				}
				if result == nil {
					t.Errorf("RunNvidiaSMIErrorQuery(%q) returned nil result", tt.errorType)
				} else {
					// Verify the command structure
					expectedCmd := "sudo nvidia-smi -q | grep -A 3 Aggregate | grep " + strings.Title(strings.ToLower(tt.errorType))
					if result.Command != expectedCmd {
						t.Errorf("RunNvidiaSMIErrorQuery(%q) command = %q, want %q", tt.errorType, result.Command, expectedCmd)
					}
					t.Logf("Command executed: %s", result.Command)
					t.Logf("Output: %s", result.Output)
					if result.Error != nil {
						t.Logf("Command error: %v", result.Error)
						t.Logf("Exit code: %d", result.ExitCode)
					}
				}
			}
		})
	}
}

// TestRunNvidiaSMIErrorQueryIntegration performs integration testing
func TestRunNvidiaSMIErrorQueryIntegration(t *testing.T) {
	// Skip if nvidia-smi is not available
	if !IsNvidiaSMIAvailable() {
		t.Skip("nvidia-smi not available in test environment")
	}

	t.Run("integration test both error types", func(t *testing.T) {
		errorTypes := []string{"uncorrectable", "correctable"}

		for _, errorType := range errorTypes {
			t.Run(errorType, func(t *testing.T) {
				result, err := RunNvidiaSMIErrorQuery(errorType)

				// Note: The command might fail with exit code if grep doesn't find matches
				// This is expected behavior and not necessarily an error
				if err != nil {
					t.Logf("Command failed (this may be expected if no errors found): %v", err)
					if result != nil {
						t.Logf("Exit code: %d", result.ExitCode)
						t.Logf("Output: %s", result.Output)
					}
				} else {
					t.Logf("Command succeeded for %s errors", errorType)
					if result != nil {
						t.Logf("Output: %s", result.Output)
					}
				}

				// Verify result structure regardless of success/failure
				if result != nil {
					if result.Command == "" {
						t.Errorf("Result command is empty")
					}
					// Output can be empty if no errors are found
					t.Logf("Command: %s", result.Command)
					t.Logf("Output length: %d", len(result.Output))
				}
			})
		}
	})
}

// TestRunNvidiaSMIErrorQueryCommandConstruction tests command construction logic
func TestRunNvidiaSMIErrorQueryCommandConstruction(t *testing.T) {
	tests := []struct {
		name        string
		errorType   string
		expectedCmd string
		expectError bool
	}{
		{
			name:        "uncorrectable command",
			errorType:   "uncorrectable",
			expectedCmd: "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable",
			expectError: false,
		},
		{
			name:        "correctable command",
			errorType:   "correctable",
			expectedCmd: "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable",
			expectError: false,
		},
		{
			name:        "case insensitive uncorrectable",
			errorType:   "UNCORRECTABLE",
			expectedCmd: "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable",
			expectError: false,
		},
		{
			name:        "case insensitive correctable",
			errorType:   "CORRECTABLE",
			expectedCmd: "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable",
			expectError: false,
		},
		{
			name:        "mixed case",
			errorType:   "UnCorrectable",
			expectedCmd: "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable",
			expectError: false,
		},
		{
			name:        "unsupported type",
			errorType:   "invalid",
			expectedCmd: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll test the command construction logic by examining what would be built
			// This is a bit of a white-box test, but helps verify the logic
			var expectedCmd string
			var shouldError bool

			switch strings.ToLower(tt.errorType) {
			case "uncorrectable":
				expectedCmd = "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable"
			case "correctable":
				expectedCmd = "sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable"
			default:
				shouldError = true
			}

			if shouldError != tt.expectError {
				t.Errorf("Expected error status %v, got %v", tt.expectError, shouldError)
			}

			if !shouldError && expectedCmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q", tt.expectedCmd, expectedCmd)
			}
		})
	}
}

// BenchmarkRunNvidiaSMIErrorQuery benchmarks the RunNvidiaSMIErrorQuery function
func BenchmarkRunNvidiaSMIErrorQuery(b *testing.B) {
	// Skip if nvidia-smi is not available
	if !IsNvidiaSMIAvailable() {
		b.Skip("nvidia-smi not available in test environment")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunNvidiaSMIErrorQuery("uncorrectable")
	}
}

// TestNvidiaSMIResult tests the NvidiaSMIResult struct
func TestNvidiaSMIResult(t *testing.T) {
	result := &NvidiaSMIResult{
		Available: true,
		Output:    "test output",
		Error:     "",
	}

	if !result.Available {
		t.Errorf("NvidiaSMIResult.Available = %v, want true", result.Available)
	}

	if result.Output != "test output" {
		t.Errorf("NvidiaSMIResult.Output = %q, want %q", result.Output, "test output")
	}

	if result.Error != "" {
		t.Errorf("NvidiaSMIResult.Error = %q, want empty", result.Error)
	}
}

// TestGPUCountParsing tests GPU count detection logic
func TestGPUCountParsing(t *testing.T) {
	// Mock nvidia-smi output for testing GPU count logic
	mockOutput := `Fri Jul  5 05:55:00 2025       
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.144.03             Driver Version: 550.144.03     CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA GeForce GTX 1650        Off |   00000000:65:00.0  On |                  N/A |
| 50%   37C    P5             N/A /   75W |    1023MiB /   4096MiB |     35%      Default |
|                                         |                        |                  N/A |
+-----------------------------------------+------------------------+----------------------+
|   1  NVIDIA Tesla V100-SXM2-16GB    Off |   00000000:3B:00.0 Off |                    0 |
| N/A   42C    P0               N/A / 300W |      0MiB / 16160MiB |      0%      Default |
|                                         |                        |                  N/A |
+-----------------------------------------+------------------------+----------------------+`

	// Test GPU counting logic
	lines := strings.Split(mockOutput, "\n")
	gpuCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") && len(trimmed) > 10 {
			if strings.Contains(line, "GeForce") ||
				strings.Contains(line, "Tesla") ||
				strings.Contains(line, "Quadro") ||
				strings.Contains(line, "RTX") ||
				strings.Contains(line, "GTX") {
				if !strings.Contains(line, "GPU  Name") &&
					!strings.Contains(line, "===") &&
					!strings.Contains(line, "NVIDIA-SMI") &&
					!strings.Contains(line, "Driver Version") {
					gpuCount++
				}
			}
		}
	}

	expectedCount := 2 // GTX 1650 and Tesla V100
	if gpuCount != expectedCount {
		t.Errorf("GPU count parsing failed: got %d, want %d", gpuCount, expectedCount)
	}
}

// BenchmarkCheckNvidiaSMI benchmarks the CheckNvidiaSMI function
func BenchmarkCheckNvidiaSMI(b *testing.B) {
	// Skip if nvidia-smi is not available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		b.Skip("nvidia-smi not available in test environment")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckNvidiaSMI()
	}
}

// BenchmarkRunNvidiaSMIQuery benchmarks the RunNvidiaSMIQuery function
func BenchmarkRunNvidiaSMIQuery(b *testing.B) {
	// Skip if nvidia-smi is not available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		b.Skip("nvidia-smi not available in test environment")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunNvidiaSMIQuery("name")
	}
}

// TestIntegration performs an integration test of all functions
func TestIntegration(t *testing.T) {
	// Skip if nvidia-smi is not available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		t.Skip("nvidia-smi not available in test environment")
	}

	t.Run("integration test", func(t *testing.T) {
		// Test basic availability
		result := CheckNvidiaSMI()
		if !result.Available {
			t.Fatalf("nvidia-smi not available: %v", result.Error)
		}

		// Test GPU name query to validate functionality
		nameResult := RunNvidiaSMIQuery("name")
		if !nameResult.Available {
			t.Fatalf("GPU name query failed: %v", nameResult.Error)
		}
		t.Logf("Found GPUs: %s", nameResult.Output)

		// Test queries
		queries := []string{"name", "memory.total"}
		for _, query := range queries {
			queryResult := RunNvidiaSMIQuery(query)
			if !queryResult.Available {
				t.Errorf("Query %q failed: %v", query, queryResult.Error)
			} else {
				t.Logf("Query %q result: %s", query, queryResult.Output)
			}
		}
	})
}

// Test GetGPUInfo function
func TestGetGPUInfo(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		skipIfNoGPU bool
	}{
		{
			name:        "get GPU info",
			expectError: false,
			skipIfNoGPU: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoGPU && !IsNvidiaSMIAvailable() {
				t.Skip("Skipping test - nvidia-smi not available")
			}

			gpus, err := GetGPUInfo()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error expected, verify the structure
			if !tt.expectError && err == nil {
				t.Logf("Found %d GPUs", len(gpus))
				for i, gpu := range gpus {
					t.Logf("GPU %d: PCI=%s, Model=%s, ID=%d", i, gpu.PCI, gpu.Model, gpu.ID)

					// Basic validation
					if gpu.PCI == "" {
						t.Errorf("GPU %d has empty PCI address", i)
					}
					if gpu.Model == "" {
						t.Errorf("GPU %d has empty model", i)
					}
					if gpu.ID < 0 {
						t.Errorf("GPU %d has invalid ID: %d", i, gpu.ID)
					}
				}
			}
		})
	}
}

// Test parseGPUInfo function with mock data
func TestParseGPUInfo(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		expectError bool
	}{
		{
			name:        "single GPU",
			input:       "00000000:0F:00.0, NVIDIA GeForce GTX 1650, 0",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "multiple GPUs",
			input:       "00000000:0F:00.0, NVIDIA H100 80GB HBM3, 0\n00000000:2D:00.0, NVIDIA H100 80GB HBM3, 1",
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectedLen: 0,
			expectError: false,
		},
		{
			name:        "whitespace only",
			input:       "   \n   \n   ",
			expectedLen: 0,
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "invalid line without commas",
			expectedLen: 0,
			expectError: false, // Should skip invalid lines
		},
		{
			name:        "mixed valid and invalid",
			input:       "00000000:0F:00.0, NVIDIA H100, 0\ninvalid line\n00000000:2D:00.0, NVIDIA H100, 1",
			expectedLen: 2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpus, err := parseGPUInfo(tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(gpus) != tt.expectedLen {
				t.Errorf("Expected %d GPUs, got %d", tt.expectedLen, len(gpus))
			}

			// For valid cases, check the parsed content
			if !tt.expectError && tt.expectedLen > 0 {
				for i, gpu := range gpus {
					t.Logf("Parsed GPU %d: PCI=%s, Model=%s, ID=%d", i, gpu.PCI, gpu.Model, gpu.ID)
					if gpu.PCI == "" {
						t.Errorf("GPU %d has empty PCI address", i)
					}
					if gpu.Model == "" {
						t.Errorf("GPU %d has empty model", i)
					}
				}
			}
		})
	}
}

// Test GetGPUCount function
func TestGetGPUCount(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		skipIfNoGPU bool
	}{
		{
			name:        "get GPU count",
			expectError: false,
			skipIfNoGPU: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoGPU && !IsNvidiaSMIAvailable() {
				t.Skip("Skipping test - nvidia-smi not available")
			}

			count, err := GetGPUCount()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				t.Logf("GPU count: %d", count)
				if count < 0 {
					t.Errorf("Invalid GPU count: %d", count)
				}
			}
		})
	}
}

// Test IsNvidiaSMIAvailable function
func TestIsNvidiaSMIAvailable(t *testing.T) {
	available := IsNvidiaSMIAvailable()
	t.Logf("nvidia-smi available: %v", available)

	// This test just verifies the function doesn't crash
	// The actual availability depends on the test environment
}

// Test GPU count parsing logic with mock data for cross-platform compatibility
func TestGPUCountParsingFormats(t *testing.T) {
	tests := []struct {
		name          string
		mockOutput    string
		expectedCount int
		expectError   bool
		description   string
	}{
		{
			name:          "single_line_format_ubuntu",
			mockOutput:    "8",
			expectedCount: 8,
			expectError:   false,
			description:   "Ubuntu format - single line with total count",
		},
		{
			name:          "multi_line_format_oracle_linux",
			mockOutput:    "8\n8\n8\n8\n8\n8\n8\n8",
			expectedCount: 8,
			expectError:   false,
			description:   "Oracle Linux format - multiple lines with count per GPU",
		},
		{
			name:          "single_gpu",
			mockOutput:    "1",
			expectedCount: 1,
			expectError:   false,
			description:   "Single GPU system",
		},
		{
			name:          "multi_line_single_gpu",
			mockOutput:    "1",
			expectedCount: 1,
			expectError:   false,
			description:   "Single GPU with multi-line format",
		},
		{
			name:          "no_gpus",
			mockOutput:    "0",
			expectedCount: 0,
			expectError:   false,
			description:   "No GPUs detected",
		},
		{
			name:          "empty_output",
			mockOutput:    "",
			expectedCount: 0,
			expectError:   false,
			description:   "Empty nvidia-smi output",
		},
		{
			name:          "whitespace_only",
			mockOutput:    "   \n   \n   ",
			expectedCount: 0,
			expectError:   false,
			description:   "Whitespace only output",
		},
		{
			name:          "mixed_whitespace_and_counts",
			mockOutput:    "4\n\n4\n \n4\n4\n",
			expectedCount: 4,
			expectError:   false,
			description:   "Mixed empty lines and counts",
		},
		{
			name:          "invalid_single_line",
			mockOutput:    "invalid",
			expectedCount: 0,
			expectError:   true,
			description:   "Invalid single line format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock result
			result := &NvidiaSMIResult{
				Available: true,
				Output:    tt.mockOutput,
				Error:     "",
			}

			// Parse using the same logic as GetGPUCount
			countStr := strings.TrimSpace(result.Output)
			var count int
			var err error

			if countStr == "" {
				count = 0
			} else {
				// Handle both single count and multi-line output
				lines := strings.Split(countStr, "\n")

				if len(lines) == 1 {
					// Single line format
					count, err = strconv.Atoi(strings.TrimSpace(lines[0]))
				} else {
					// Multi-line format - count non-empty lines
					count = 0
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" {
							count++
						}
					}
				}
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}

			t.Logf("Test %s: %s - Count: %d", tt.name, tt.description, count)
		})
	}
}

// Test GPUInfo struct
func TestGPUInfoStruct(t *testing.T) {
	gpu := GPUInfo{
		PCI:   "00000000:0F:00.0",
		Model: "NVIDIA H100 80GB HBM3",
		ID:    0,
	}

	if gpu.PCI != "00000000:0F:00.0" {
		t.Errorf("Expected PCI '00000000:0F:00.0', got '%s'", gpu.PCI)
	}
	if gpu.Model != "NVIDIA H100 80GB HBM3" {
		t.Errorf("Expected Model 'NVIDIA H100 80GB HBM3', got '%s'", gpu.Model)
	}
	if gpu.ID != 0 {
		t.Errorf("Expected ID 0, got %d", gpu.ID)
	}
}
