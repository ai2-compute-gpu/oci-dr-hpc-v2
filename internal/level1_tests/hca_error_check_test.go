package level1_tests

import (
	"testing"
)

func TestHcaErrorCheckTestConfig(t *testing.T) {
	config := &HcaErrorCheckTestConfig{
		IsEnabled: true,
		Shape:     "BM.GPU.H100.8",
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
	if config.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected Shape BM.GPU.H100.8, got %s", config.Shape)
	}
}

func TestParseDmesgForMLX5FatalErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "No errors",
			input:    "kernel: normal log message\nkernel: another message",
			expected: 0,
		},
		{
			name:     "MLX5 fatal error",
			input:    "kernel: mlx5_core 0000:17:00.0: Fatal error detected\nkernel: normal message",
			expected: 1,
		},
		{
			name:     "Multiple MLX5 fatal errors",
			input:    "kernel: mlx5_core 0000:17:00.0: Fatal error\nkernel: mlx5_core 0000:18:00.0: FATAL issue",
			expected: 2,
		},
		{
			name:     "Case insensitive matching",
			input:    "kernel: MLX5_CORE 0000:17:00.0: FATAL ERROR DETECTED",
			expected: 1,
		},
		{
			name:     "MLX5 without fatal",
			input:    "kernel: mlx5_core 0000:17:00.0: info message",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDmesgForMLX5FatalErrors(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected %d errors, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestParseDmesgForMLX5FatalErrors_EmptyInput(t *testing.T) {
	result := parseDmesgForMLX5FatalErrors("")
	if len(result) != 0 {
		t.Errorf("Expected 0 errors for empty input, got %d", len(result))
	}
}

func BenchmarkParseDmesgForMLX5FatalErrors(b *testing.B) {
	input := "kernel: mlx5_core 0000:17:00.0: Fatal error detected\nkernel: normal message\nkernel: mlx5_core 0000:18:00.0: another fatal error"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDmesgForMLX5FatalErrors(input)
	}
}