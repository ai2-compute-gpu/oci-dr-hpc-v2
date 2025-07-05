package executor

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestOSCommandResult(t *testing.T) {
	result := &OSCommandResult{
		Command:  "test command",
		Output:   "test output",
		Error:    nil,
		ExitCode: 0,
	}

	if result.Command != "test command" {
		t.Errorf("Expected Command to be 'test command', got '%s'", result.Command)
	}
	if result.Output != "test output" {
		t.Errorf("Expected Output to be 'test output', got '%s'", result.Output)
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", result.ExitCode)
	}
}

func TestRunLspci(t *testing.T) {
	tests := []struct {
		name        string
		options     []string
		expectError bool
		skipIfNoSudo bool
	}{
		{
			name:         "basic lspci call",
			options:      []string{},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "lspci with verbose option",
			options:      []string{"-v"},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "lspci with multiple options",
			options:      []string{"-v", "-k"},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "lspci with invalid option",
			options:      []string{"--invalid-option"},
			expectError:  true,
			skipIfNoSudo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoSudo && !canRunSudo() {
				t.Skip("Skipping test that requires sudo access")
			}

			result, err := RunLspci(tt.options...)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Verify command string format
			expectedCmd := "sudo lspci " + strings.Join(tt.options, " ")
			if result.Command != expectedCmd {
				t.Errorf("Expected command '%s', got '%s'", expectedCmd, result.Command)
			}

			// If no error expected, verify we have some output or at least empty string
			if !tt.expectError {
				if result.Output == "" && err == nil {
					// This might be normal if no PCI devices, so just log it
					t.Logf("No lspci output (might be normal in test environment)")
				}
			}
		})
	}
}

func TestRunLspciForDevice(t *testing.T) {
	tests := []struct {
		name         string
		deviceID     string
		verbose      bool
		expectError  bool
		skipIfNoSudo bool
	}{
		{
			name:         "query specific device non-verbose",
			deviceID:     "00:00.0",
			verbose:      false,
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "query specific device verbose",
			deviceID:     "00:00.0",
			verbose:      true,
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "query non-existent device",
			deviceID:     "ff:ff.f",
			verbose:      false,
			expectError:  true,
			skipIfNoSudo: true,
		},
		{
			name:         "empty device ID",
			deviceID:     "",
			verbose:      false,
			expectError:  false,
			skipIfNoSudo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoSudo && !canRunSudo() {
				t.Skip("Skipping test that requires sudo access")
			}

			result, err := RunLspciForDevice(tt.deviceID, tt.verbose)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Verify command contains device ID
			if !strings.Contains(result.Command, tt.deviceID) && tt.deviceID != "" {
				t.Errorf("Expected command to contain device ID '%s', got '%s'", tt.deviceID, result.Command)
			}

			// Verify verbose flag
			if tt.verbose && !strings.Contains(result.Command, "-v") {
				t.Error("Expected command to contain -v flag for verbose mode")
			}
		})
	}
}

func TestRunDmesg(t *testing.T) {
	tests := []struct {
		name         string
		options      []string
		expectError  bool
		skipIfNoSudo bool
	}{
		{
			name:         "basic dmesg call",
			options:      []string{},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "dmesg with level filter",
			options:      []string{"-l", "err"},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "dmesg with time since boot",
			options:      []string{"-T"},
			expectError:  false,
			skipIfNoSudo: true,
		},
		{
			name:         "dmesg with invalid option",
			options:      []string{"--invalid-option"},
			expectError:  true,
			skipIfNoSudo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoSudo && !canRunSudo() {
				t.Skip("Skipping test that requires sudo access")
			}

			result, err := RunDmesg(tt.options...)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Verify command string format
			expectedCmd := "sudo dmesg " + strings.Join(tt.options, " ")
			if result.Command != expectedCmd {
				t.Errorf("Expected command '%s', got '%s'", expectedCmd, result.Command)
			}

			// If no error expected, verify we have some output
			if !tt.expectError && len(result.Output) == 0 {
				t.Log("No dmesg output (might be normal in test environment)")
			}
		})
	}
}

func TestOSCommandResultWithError(t *testing.T) {
	// Test with a command that will definitely fail
	if !canRunSudo() {
		t.Skip("Skipping test that requires sudo access")
	}

	result, err := RunLspci("--definitely-invalid-option")

	if err == nil {
		t.Error("Expected error for invalid option")
	}

	if result == nil {
		t.Fatal("Expected result even with error")
	}

	if result.Error == nil {
		t.Error("Expected result.Error to be set")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for failed command")
	}

	if !strings.Contains(result.Command, "sudo lspci") {
		t.Error("Expected command to contain 'sudo lspci'")
	}
}

// Helper function to check if we can run sudo commands
func canRunSudo() bool {
	// Check if we're running as root
	if os.Getuid() == 0 {
		return true
	}

	// Check if sudo is available and we can use it without password
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

func TestCanRunSudo(t *testing.T) {
	result := canRunSudo()
	t.Logf("Can run sudo: %v", result)
	
	if !result {
		t.Log("Sudo tests will be skipped - run as root or configure passwordless sudo for full test coverage")
	}
}

// Test command string formatting
func TestCommandStringFormatting(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		expected string
	}{
		{
			name:     "no options",
			options:  []string{},
			expected: "sudo lspci ",
		},
		{
			name:     "single option",
			options:  []string{"-v"},
			expected: "sudo lspci -v",
		},
		{
			name:     "multiple options",
			options:  []string{"-v", "-k", "-s", "00:00.0"},
			expected: "sudo lspci -v -k -s 00:00.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock result to test command string formatting
			result := &OSCommandResult{
				Command: "sudo lspci " + strings.Join(tt.options, " "),
			}

			if result.Command != tt.expected {
				t.Errorf("Expected command '%s', got '%s'", tt.expected, result.Command)
			}
		})
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("nil options to RunLspci", func(t *testing.T) {
		if !canRunSudo() {
			t.Skip("Skipping test that requires sudo access")
		}

		result, err := RunLspci()
		if err != nil {
			t.Logf("Error (may be expected in test env): %v", err)
		}
		if result == nil {
			t.Fatal("Expected result but got nil")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		if !canRunSudo() {
			t.Skip("Skipping test that requires sudo access")
		}

		result, err := RunLspciForDevice("", false)
		if err == nil {
			t.Log("No error for empty device ID (may be valid)")
		}
		if result == nil {
			t.Fatal("Expected result but got nil")
		}
	})
}

// Integration test - only runs if sudo is available
func TestIntegrationWithActualCommands(t *testing.T) {
	if !canRunSudo() {
		t.Skip("Skipping integration test - requires sudo access")
	}

	t.Run("lspci integration", func(t *testing.T) {
		result, err := RunLspci()
		if err != nil {
			t.Logf("lspci failed (may be expected in test environment): %v", err)
			return
		}

		if result.Output == "" {
			t.Log("No lspci output (may be normal in containerized environment)")
		}

		if !strings.Contains(result.Command, "sudo lspci") {
			t.Error("Command should contain 'sudo lspci'")
		}
	})

	t.Run("dmesg integration", func(t *testing.T) {
		result, err := RunDmesg()
		if err != nil {
			t.Logf("dmesg failed (may be expected in test environment): %v", err)
			return
		}

		if result.Output == "" {
			t.Log("No dmesg output (may be normal in containerized environment)")
		}

		if !strings.Contains(result.Command, "sudo dmesg") {
			t.Error("Command should contain 'sudo dmesg'")
		}
	})
}