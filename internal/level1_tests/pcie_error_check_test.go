package level1_tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

// Test data for various dmesg outputs
const (
	cleanDmesgOutput = `[    0.000000] Linux version 5.4.0-150-generic (buildd@lgw01-amd64-038)
[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-150-generic
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Hygon HygonGenuine
[    0.000000]   Centaur CentaurHauls
[    0.000000]   zhaoxin   Shanghai  
[    0.000000] x86/fpu: Supporting XSAVE feature 0x001: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x002: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x004: 'AVX registers'
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] BIOS-provided physical RAM map:`

	pciErrorDmesgOutput = `[    0.000000] Linux version 5.4.0-150-generic (buildd@lgw01-amd64-038)
[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-150-generic
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Hygon HygonGenuine
[    0.000000]   Centaur CentaurHauls
[    0.000000]   zhaoxin   Shanghai  
[    0.000000] x86/fpu: Supporting XSAVE feature 0x001: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x002: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x004: 'AVX registers'
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] BIOS-provided physical RAM map:
[  123.456789] pcieport 0000:00:1c.0: AER: Corrected error received: 0000:00:1c.0
[  123.456790] pcieport 0000:00:1c.0: PCIe Bus Error: severity=Corrected, type=Data Link Layer, (Receiver ID)
[  123.456791] pcieport 0000:00:1c.0: device [8086:1c16] error status/mask=00000040/00002000
[  123.456792] pcieport 0000:00:1c.0: [6] BadTLP`

	pciCapabilitiesDmesgOutput = `[    0.000000] Linux version 5.4.0-150-generic (buildd@lgw01-amd64-038)
[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-150-generic
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Hygon HygonGenuine
[    0.000000]   Centaur CentaurHauls
[    0.000000]   zhaoxin   Shanghai  
[    0.000000] x86/fpu: Supporting XSAVE feature 0x001: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x002: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x004: 'AVX registers'
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] BIOS-provided physical RAM map:
[  123.456789] pcieport 0000:00:1c.0: capabilities [40] Power Management version 3
[  123.456790] pcieport 0000:00:1c.0: capabilities [50] MSI: Enable- Count=1/1 Maskable- 64bit-
[  123.456791] pcieport 0000:00:1c.0: capabilities [70] Express (v2) Root Port (Slot+)`

	mixedDmesgOutput = `[    0.000000] Linux version 5.4.0-150-generic (buildd@lgw01-amd64-038)
[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-150-generic
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Hygon HygonGenuine
[    0.000000]   Centaur CentaurHauls
[    0.000000]   zhaoxin   Shanghai  
[    0.000000] x86/fpu: Supporting XSAVE feature 0x001: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x002: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x004: 'AVX registers'
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] BIOS-provided physical RAM map:
[  123.456789] pcieport 0000:00:1c.0: capabilities [40] Power Management version 3
[  123.456790] pcieport 0000:00:1c.0: capabilities [50] MSI: Enable- Count=1/1 Maskable- 64bit-
[  123.456791] pcieport 0000:00:1c.0: capabilities [70] Express (v2) Root Port (Slot+)
[  123.456792] pcieport 0000:00:1c.0: AER: Corrected error received: 0000:00:1c.0
[  123.456793] pcieport 0000:00:1c.0: PCIe Bus Error: severity=Corrected, type=Data Link Layer, (Receiver ID)`

	caseInsensitiveDmesgOutput = `[    0.000000] Linux version 5.4.0-150-generic (buildd@lgw01-amd64-038)
[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-150-generic
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Hygon HygonGenuine
[    0.000000]   Centaur CentaurHauls
[    0.000000]   zhaoxin   Shanghai  
[    0.000000] x86/fpu: Supporting XSAVE feature 0x001: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x002: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x004: 'AVX registers'
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] BIOS-provided physical RAM map:
[  123.456789] pcieport 0000:00:1c.0: AER: Corrected ERROR received: 0000:00:1c.0
[  123.456790] pcieport 0000:00:1c.0: PCIe Bus ERROR: severity=Corrected, type=Data Link Layer, (Receiver ID)`
)

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

// Test the PcieErrorCheckTestConfig struct
func TestPcieErrorCheckTestConfig(t *testing.T) {
	config := &PcieErrorCheckTestConfig{
		IsEnabled: true,
		Shape:     "BM.GPU.H100.8",
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}

	if config.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape 'BM.GPU.H100.8', got '%s'", config.Shape)
	}

	// Test default values
	defaultConfig := &PcieErrorCheckTestConfig{}
	if defaultConfig.IsEnabled {
		t.Error("Expected default IsEnabled to be false")
	}

	if defaultConfig.Shape != "" {
		t.Errorf("Expected default shape to be empty, got '%s'", defaultConfig.Shape)
	}
}

// testPCIeErrorDetectionLogic tests the core PCIe error detection logic
// without actually calling dmesg
func testPCIeErrorDetectionLogic(t *testing.T, dmesgOutput string, expectError bool, testName string) {
	t.Run(testName, func(t *testing.T) {
		// Check if dmesg output is empty
		outputStr := dmesgOutput
		if len(strings.TrimSpace(outputStr)) == 0 {
			if expectError {
				return // Expected to fail due to empty output
			}
			t.Error("Empty dmesg output should cause failure")
			return
		}

		// Look through each line of the dmesg output
		lines := strings.Split(outputStr, "\n")
		foundError := false
		for _, line := range lines {
			// Skip lines that contain "capabilities" - these are not error messages
			if strings.Contains(line, "capabilities") {
				continue
			}

			// Look for lines that contain both "pcieport" and "error" (case insensitive)
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "pcieport") && strings.Contains(lowerLine, "error") {
				foundError = true
				break
			}
		}

		if expectError && !foundError {
			t.Error("Expected to find PCIe error but didn't")
		}
		if !expectError && foundError {
			t.Error("Expected no PCIe error but found one")
		}
	})
}

func TestPCIeErrorDetectionLogic(t *testing.T) {
	// Test cases for the core logic
	testPCIeErrorDetectionLogic(t, cleanDmesgOutput, false, "clean dmesg output should pass")
	testPCIeErrorDetectionLogic(t, pciErrorDmesgOutput, true, "dmesg with PCI error should fail")
	testPCIeErrorDetectionLogic(t, pciCapabilitiesDmesgOutput, false, "dmesg with only capabilities should pass")
	testPCIeErrorDetectionLogic(t, mixedDmesgOutput, true, "dmesg with mixed capabilities and errors should fail")
	testPCIeErrorDetectionLogic(t, caseInsensitiveDmesgOutput, true, "case insensitive error detection should work")
	testPCIeErrorDetectionLogic(t, "", true, "empty output should fail")
	testPCIeErrorDetectionLogic(t, "   \n\t  \n  ", true, "whitespace only output should fail")
}

func TestPCIeErrorPatternDetection(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectError bool
	}{
		{
			name:        "pcieport with error lowercase",
			line:        "[12345.678] pcieport 0000:00:1c.0: PCIe Bus error: severity=Corrected",
			expectError: true,
		},
		{
			name:        "pcieport with ERROR uppercase",
			line:        "[12345.678] pcieport 0000:00:1c.0: PCIe Bus ERROR: severity=Corrected",
			expectError: true,
		},
		{
			name:        "pcieport with Error mixed case",
			line:        "[12345.678] pcieport 0000:00:1c.0: PCIe Bus Error: severity=Corrected",
			expectError: true,
		},
		{
			name:        "pcieport with capabilities (should be skipped)",
			line:        "[12345.678] pcieport 0000:00:1c.0: capabilities [40] Power Management version 3",
			expectError: false,
		},
		{
			name:        "pcieport without error",
			line:        "[12345.678] pcieport 0000:00:1c.0: enabling device (0000 -> 0002)",
			expectError: false,
		},
		{
			name:        "error without pcieport",
			line:        "[12345.678] kernel: some other error occurred",
			expectError: false,
		},
		{
			name:        "regular kernel message",
			line:        "[12345.678] kernel: Linux version 5.4.0-150-generic",
			expectError: false,
		},
		{
			name:        "empty line",
			line:        "",
			expectError: false,
		},
		{
			name:        "whitespace only line",
			line:        "   \t  \n  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the exact pattern matching logic
			line := tt.line

			// Skip lines that contain "capabilities" - these are not error messages
			if strings.Contains(line, "capabilities") {
				if tt.expectError {
					t.Error("Capabilities line should not be detected as error")
				}
				return
			}

			// Look for lines that contain both "pcieport" and "error" (case insensitive)
			lowerLine := strings.ToLower(line)
			foundError := strings.Contains(lowerLine, "pcieport") && strings.Contains(lowerLine, "error")

			if tt.expectError && !foundError {
				t.Errorf("Expected to find PCIe error in line: %s", line)
			}
			if !tt.expectError && foundError {
				t.Errorf("Expected no PCIe error in line: %s", line)
			}
		})
	}
}

// Integration test - only runs if sudo is available
func TestRunPCIeErrorCheckIntegration(t *testing.T) {
	if !canRunSudo() {
		t.Skip("Skipping integration test - requires sudo access")
	}

	t.Run("integration test with real dmesg", func(t *testing.T) {
		// This test calls the actual function which will use executor.RunDmesg()
		err := RunPCIeErrorCheck()

		// We can't predict if there will be PCIe errors on the test system,
		// so we just verify the function completes without panicking
		// and returns either nil (pass) or an error (fail)
		if err != nil {
			t.Logf("PCIe error check failed (may be expected): %v", err)
		} else {
			t.Log("PCIe error check passed")
		}
	})
}

// Test that verifies the function handles dmesg command failure
func TestRunPCIeErrorCheckDmesgFailure(t *testing.T) {
	// We can't easily mock executor.RunDmesg() without changing the function signature,
	// so we test this by modifying PATH to make dmesg unavailable
	// This test will be skipped if we can't manipulate sudo
	if !canRunSudo() {
		t.Skip("Skipping test that requires sudo access")
	}

	t.Run("handle dmesg command failure", func(t *testing.T) {
		// Temporarily modify PATH to break dmesg access
		originalPath := os.Getenv("PATH")
		defer func() {
			os.Setenv("PATH", originalPath)
		}()

		// Set PATH to a non-existent directory to simulate dmesg not being available
		os.Setenv("PATH", "/nonexistent")

		// This should fail because dmesg won't be found
		err := RunPCIeErrorCheck()
		if err == nil {
			t.Error("Expected error when dmesg is not available")
		}

		// Verify the error message contains expected text
		if err != nil && !strings.Contains(err.Error(), "dmesg") {
			t.Errorf("Expected error message to mention dmesg, got: %v", err)
		}
	})
}

func TestCanRunSudo(t *testing.T) {
	result := canRunSudo()
	t.Logf("Can run sudo: %v", result)

	if !result {
		t.Log("Sudo tests will be skipped - run as root or configure passwordless sudo for full test coverage")
	}
}

// Test executor.RunDmesg integration
func TestRunDmesgIntegration(t *testing.T) {
	if !canRunSudo() {
		t.Skip("Skipping integration test - requires sudo access")
	}

	t.Run("executor.RunDmesg integration", func(t *testing.T) {
		result, err := executor.RunDmesg()
		if err != nil {
			t.Logf("dmesg failed (may be expected in test environment): %v", err)
			return
		}

		if result == nil {
			t.Fatal("Expected result but got nil")
		}

		if !strings.Contains(result.Command, "sudo dmesg") {
			t.Error("Command should contain 'sudo dmesg'")
		}

		if result.Output == "" {
			t.Log("No dmesg output (may be normal in containerized environment)")
		} else {
			t.Logf("dmesg output length: %d characters", len(result.Output))
		}
	})
}

// Benchmark the PCIe error detection logic
func BenchmarkPCIeErrorDetection(b *testing.B) {
	output := mixedDmesgOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the core detection logic
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "capabilities") {
				continue
			}
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "pcieport") && strings.Contains(lowerLine, "error") {
				break
			}
		}
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("very long dmesg output", func(t *testing.T) {
		// Create a very long output to test performance
		longOutput := strings.Repeat(cleanDmesgOutput+"\n", 100)
		testPCIeErrorDetectionLogic(t, longOutput, false, "very long clean output should pass")
	})

	t.Run("output with only newlines", func(t *testing.T) {
		testPCIeErrorDetectionLogic(t, "\n\n\n\n\n", true, "output with only newlines should fail")
	})

	t.Run("single line error", func(t *testing.T) {
		singleLineError := "[12345.678] pcieport 0000:00:1c.0: PCIe Bus error: severity=Corrected"
		testPCIeErrorDetectionLogic(t, singleLineError, true, "single line with error should fail")
	})
}
