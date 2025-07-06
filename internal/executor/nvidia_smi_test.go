package executor

import (
	"os"
	"os/exec"
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
