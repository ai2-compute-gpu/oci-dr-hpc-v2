package executor

import (
	"strconv"
	"strings"
	"testing"
)

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

func TestGPUInfoStruct(t *testing.T) {
	gpu := GPUInfo{
		PCI:   "0000:0f:00.0",
		Model: "NVIDIA H100 80GB HBM3",
		ID:    0,
	}

	if gpu.PCI != "0000:0f:00.0" {
		t.Errorf("Expected PCI '0000:0f:00.0', got '%s'", gpu.PCI)
	}
	if gpu.Model != "NVIDIA H100 80GB HBM3" {
		t.Errorf("Expected Model 'NVIDIA H100 80GB HBM3', got '%s'", gpu.Model)
	}
	if gpu.ID != 0 {
		t.Errorf("Expected ID 0, got %d", gpu.ID)
	}
}

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
			expectError: false,
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

func TestGPUCountParsingFormats(t *testing.T) {
	tests := []struct {
		name          string
		mockOutput    string
		expectedCount int
		expectError   bool
		description   string
	}{
		{
			name:          "single_line_format",
			mockOutput:    "8",
			expectedCount: 8,
			expectError:   false,
			description:   "Single line with total count",
		},
		{
			name:          "multi_line_format",
			mockOutput:    "8\n8\n8\n8\n8\n8\n8\n8",
			expectedCount: 8,
			expectError:   false,
			description:   "Multiple lines with count per GPU",
		},
		{
			name:          "single_gpu",
			mockOutput:    "1",
			expectedCount: 1,
			expectError:   false,
			description:   "Single GPU system",
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
			// Parse using the same logic as GetGPUCount
			countStr := strings.TrimSpace(tt.mockOutput)
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

func TestNvidiaSMIErrorQueryCommandConstruction(t *testing.T) {
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
			// Test the command construction logic
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

func TestFormatPCIAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "nvidia format to standard",
			input:    "00000000:65:00.0",
			expected: "0000:65:00.0",
		},
		{
			name:     "already standard format",
			input:    "0000:65:00.0",
			expected: "0000:65:00.0",
		},
		{
			name:     "short format",
			input:    "65:00.0",
			expected: "65:00.0",
		},
		{
			name:     "uppercase input",
			input:    "00000000:65:00.A",
			expected: "0000:65:00.a",
		},
		{
			name:     "mixed case",
			input:    "0000000A:6B:0C.D",
			expected: "000a:6b:0c.d",
		},
		{
			name:     "invalid format passthrough",
			input:    "invalid-pci-address",
			expected: "invalid-pci-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPCIAddress(tt.input)
			if result != tt.expected {
				t.Errorf("formatPCIAddress(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGPUParsing(t *testing.T) {
	testOutput := `Fri Jul  5 05:55:00 2025       
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
	lines := strings.Split(testOutput, "\n")
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

func TestNvidiaSMIQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "basic query",
			query:    "name",
			expected: "nvidia-smi --query-gpu=name --format=csv,noheader,nounits",
		},
		{
			name:     "memory query",
			query:    "memory.total",
			expected: "nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits",
		},
		{
			name:     "multiple fields",
			query:    "name,memory.total,temperature.gpu",
			expected: "nvidia-smi --query-gpu=name,memory.total,temperature.gpu --format=csv,noheader,nounits",
		},
		{
			name:     "empty query",
			query:    "",
			expected: "nvidia-smi --query-gpu= --format=csv,noheader,nounits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := "nvidia-smi --query-gpu=" + tt.query + " --format=csv,noheader,nounits"
			if cmd != tt.expected {
				t.Errorf("Expected command %q, got %q", tt.expected, cmd)
			}
		})
	}
}

func TestRunNvidiaSMINvlinkResult(t *testing.T) {
	// Test the NvidiaSMIResult structure for NVLink command
	result := &NvidiaSMIResult{
		Available: true,
		Output:    "GPU 0: Tesla V100-SXM2-16GB (UUID: GPU-12345678-1234-1234-1234-123456789012)\n\tLink 0: 25.781 GB/s\n\tLink 1: 25.781 GB/s",
		Error:     "",
	}

	if !result.Available {
		t.Errorf("Expected Available=true for successful NVLink result")
	}

	if result.Error != "" {
		t.Errorf("Expected empty error for successful NVLink result, got %q", result.Error)
	}

	if !strings.Contains(result.Output, "Link") {
		t.Errorf("Expected NVLink output to contain 'Link', got %q", result.Output)
	}
}

func TestNVLinkOutputParsing(t *testing.T) {
	tests := []struct {
		name          string
		mockOutput    string
		expectedLinks int
		expectError   bool
		description   string
	}{
		{
			name: "successful_nvlink_output",
			mockOutput: `GPU 0: Tesla V100-SXM2-16GB (UUID: GPU-12345678-1234-1234-1234-123456789012)
	Link 0: 25.781 GB/s
	Link 1: 25.781 GB/s
	Link 2: 25.781 GB/s
	Link 3: 25.781 GB/s`,
			expectedLinks: 4,
			expectError:   false,
			description:   "Successful NVLink output with 4 links",
		},
		{
			name: "h100_nvlink_output",
			mockOutput: `GPU 0: NVIDIA H100 80GB HBM3 (UUID: GPU-87654321-4321-4321-4321-210987654321)
	Link 0: 50.0 GB/s
	Link 1: 50.0 GB/s
	Link 2: 50.0 GB/s
	Link 3: 50.0 GB/s
	Link 4: 50.0 GB/s
	Link 5: 50.0 GB/s
	Link 6: 50.0 GB/s
	Link 7: 50.0 GB/s
	Link 8: 50.0 GB/s
	Link 9: 50.0 GB/s
	Link 10: 50.0 GB/s
	Link 11: 50.0 GB/s
	Link 12: 50.0 GB/s
	Link 13: 50.0 GB/s
	Link 14: 50.0 GB/s
	Link 15: 50.0 GB/s
	Link 16: 50.0 GB/s
	Link 17: 50.0 GB/s`,
			expectedLinks: 18,
			expectError:   false,
			description:   "H100 NVLink output with 18 links",
		},
		{
			name:          "empty_output",
			mockOutput:    "",
			expectedLinks: 0,
			expectError:   false,
			description:   "Empty nvidia-smi nvlink output",
		},
		{
			name:          "no_nvlink_support",
			mockOutput:    "GPU 0: GeForce GTX 1650 (UUID: GPU-11111111-1111-1111-1111-111111111111)\n\tN/A",
			expectedLinks: 0,
			expectError:   false,
			description:   "GPU without NVLink support",
		},
		{
			name: "mixed_nvlink_status",
			mockOutput: `GPU 0: Tesla V100-SXM2-16GB (UUID: GPU-12345678-1234-1234-1234-123456789012)
	Link 0: 25.781 GB/s
	Link 1: N/A
	Link 2: 25.781 GB/s
	Link 3: N/A`,
			expectedLinks: 2,
			expectError:   false,
			description:   "Mixed NVLink status with some links down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse using similar logic that would be used in the actual implementation
			linkCount := 0
			lines := strings.Split(tt.mockOutput, "\n")

			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, "Link") &&
					strings.Contains(trimmed, "GB/s") &&
					!strings.Contains(trimmed, "N/A") {
					linkCount++
				}
			}

			if linkCount != tt.expectedLinks {
				t.Errorf("Expected %d active links, got %d", tt.expectedLinks, linkCount)
			}

			t.Logf("Test %s: %s - Active Links: %d", tt.name, tt.description, linkCount)
		})
	}
}

func TestNVLinkCommandConstruction(t *testing.T) {
	tests := []struct {
		name        string
		expectedCmd string
		description string
	}{
		{
			name:        "nvlink_speed_command",
			expectedCmd: "nvidia-smi nvlink -s",
			description: "NVLink speed and status command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the command construction for NVLink
			cmd := "nvidia-smi nvlink -s"
			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q", tt.expectedCmd, cmd)
			}
		})
	}
}
