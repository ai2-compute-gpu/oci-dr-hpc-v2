package level1_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock lspci output for testing
const (
	mockNvidiaLspciOutput = `      4           LnkSta:     Speed 16GT/s (ok), Width x2 (ok)
      8           LnkSta:     Speed 32GT/s (ok), Width x16 (ok)`

	mockMellanoxLspciOutput = `      2           LnkSta:     Speed 16GT/s (ok), Width x8 (ok)
     16           LnkSta:     Speed 32GT/s (ok), Width x16 (ok)`

	mockInvalidNvidiaLspciOutput = `      2           LnkSta:     Speed 16GT/s (ok), Width x2 (ok)
      6           LnkSta:     Speed 32GT/s (ok), Width x16 (ok)`

	mockInvalidMellanoxLspciOutput = `      1           LnkSta:     Speed 16GT/s (ok), Width x8 (ok)
     15           LnkSta:     Speed 32GT/s (ok), Width x16 (ok)`

	mockEmptyLspciOutput = ``

	mockMixedFormatLspciOutput = `      4           LnkSta:     Speed 16GT/s (ok), Width x2 (ok)
      8           LnkSta2:    Speed 32GT/s (ok), Width x16 (ok)
invalid line without width info
      2           LnkSta:     Speed 16GT/s (ok), Width x8 (ok)`

	mockErrorStateLspciOutput = `      4           LnkSta:     Speed 16GT/s (error), Width x2 (ok)
      8           LnkSta:     Speed 16GT/s (ok), Width x16 (degraded)`

	mockSpeedVariationsOutput = `      4           LnkSta:     Speed 8GT/s (ok), Width x2 (ok)
      8           LnkSta:     Speed 16GT/s (ok), Width x16 (ok)`
)

// createTempTestLimitsFile creates a temporary test_limits.json file for testing
func createTempTestLimitsFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	limitsFile := filepath.Join(tempDir, "test_limits.json")

	err := os.WriteFile(limitsFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp test limits file: %v", err)
	}

	return limitsFile
}

func TestParseLspciWidthOutput(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedLinkState string
		expectedWidths    map[string]int
		expectedSpeeds    map[string]int
		expectedStateErrs int
		description       string
	}{
		{
			name:              "Valid NVIDIA output",
			input:             mockNvidiaLspciOutput,
			expectedLinkState: "ok",
			expectedWidths: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			expectedSpeeds: map[string]int{
				"Speed 16GT/s": 4,
				"Speed 32GT/s": 8,
			},
			expectedStateErrs: 0,
			description:       "Should parse valid NVIDIA lspci output correctly",
		},
		{
			name:              "Valid Mellanox output",
			input:             mockMellanoxLspciOutput,
			expectedLinkState: "ok",
			expectedWidths: map[string]int{
				"Width x8":  2,
				"Width x16": 16,
			},
			expectedSpeeds: map[string]int{
				"Speed 16GT/s": 2,
				"Speed 32GT/s": 16,
			},
			expectedStateErrs: 0,
			description:       "Should parse valid Mellanox lspci output correctly",
		},
		{
			name:              "Empty output",
			input:             mockEmptyLspciOutput,
			expectedLinkState: "ok",
			expectedWidths:    map[string]int{},
			expectedSpeeds:    map[string]int{},
			expectedStateErrs: 0,
			description:       "Should handle empty output gracefully",
		},
		{
			name:              "Error state output",
			input:             mockErrorStateLspciOutput,
			expectedLinkState: "ok",
			expectedWidths: map[string]int{
				"Width x2": 4, // Only ok width states count
			},
			expectedSpeeds: map[string]int{
				"Speed 16GT/s": 8, // Only ok speed states count
			},
			expectedStateErrs: 2, // Speed error and width error
			description:       "Should detect state errors and exclude bad states from counts",
		},
		{
			name:              "Mixed format output",
			input:             mockMixedFormatLspciOutput,
			expectedLinkState: "ok",
			expectedWidths: map[string]int{
				"Width x2": 4,
				"Width x8": 2,
			},
			expectedSpeeds: map[string]int{
				"Speed 16GT/s": 6,
			},
			expectedStateErrs: 0,
			description:       "Should handle mixed format output and ignore invalid lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLspciWidthOutput(tt.input, tt.expectedLinkState)

			// Check width counts
			if len(result.WidthCounts) != len(tt.expectedWidths) {
				t.Errorf("%s: expected %d width entries, got %d", 
					tt.description, len(tt.expectedWidths), len(result.WidthCounts))
			}

			for width, expectedCount := range tt.expectedWidths {
				if actualCount, exists := result.WidthCounts[width]; !exists {
					if expectedCount != 0 {
						t.Errorf("%s: expected width %s not found in result", tt.description, width)
					}
				} else if actualCount != expectedCount {
					t.Errorf("%s: width %s expected count %d, got %d", 
						tt.description, width, expectedCount, actualCount)
				}
			}

			// Check speed counts
			for speed, expectedCount := range tt.expectedSpeeds {
				if actualCount, exists := result.SpeedCounts[speed]; !exists {
					t.Errorf("%s: expected speed %s not found in result", tt.description, speed)
				} else if actualCount != expectedCount {
					t.Errorf("%s: speed %s expected count %d, got %d", 
						tt.description, speed, expectedCount, actualCount)
				}
			}

			// Check state errors
			if len(result.StateErrors) != tt.expectedStateErrs {
				t.Errorf("%s: expected %d state errors, got %d", 
					tt.description, tt.expectedStateErrs, len(result.StateErrors))
			}
		})
	}
}

func TestValidateWidthCounts(t *testing.T) {
	tests := []struct {
		name         string
		actual       map[string]int
		expected     map[string]int
		deviceType   string
		shouldPass   bool
		expectedMsg  string
		description  string
	}{
		{
			name: "Valid GPU widths",
			actual: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			expected: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			deviceType:  "GPU/NVSwitch",
			shouldPass:  true,
			description: "Should pass with matching GPU width counts",
		},
		{
			name: "Invalid GPU width count",
			actual: map[string]int{
				"Width x2":  2,
				"Width x16": 6,
			},
			expected: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			deviceType:   "GPU/NVSwitch",
			shouldPass:   false,
			expectedMsg:  "GPU/NVSwitch PCIe width mismatch:",
			description:  "Should fail with mismatched GPU width counts",
		},
		{
			name: "Missing width entry",
			actual: map[string]int{
				"Width x16": 8,
			},
			expected: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			deviceType:   "GPU/NVSwitch",
			shouldPass:   false,
			expectedMsg:  "GPU/NVSwitch PCIe width mismatch: expected 4x Width x2, got 0x Width x2",
			description:  "Should fail when expected width entry is missing",
		},
		{
			name: "Valid RDMA widths",
			actual: map[string]int{
				"Width x8":  2,
				"Width x16": 16,
			},
			expected: map[string]int{
				"Width x8":  2,
				"Width x16": 16,
			},
			deviceType:  "RDMA",
			shouldPass:  true,
			description: "Should pass with matching RDMA width counts",
		},
		{
			name:         "Empty expected counts",
			actual:       map[string]int{"Width x16": 8},
			expected:     map[string]int{},
			deviceType:   "GPU/NVSwitch",
			shouldPass:   true,
			description:  "Should pass when no expected counts are specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success, errorMsg := validateWidthCounts(tt.actual, tt.expected, tt.deviceType)

			if success != tt.shouldPass {
				t.Errorf("%s: expected success=%v, got success=%v", tt.description, tt.shouldPass, success)
			}

			if !tt.shouldPass && !strings.Contains(errorMsg, tt.expectedMsg) {
				t.Errorf("%s: expected error message to contain '%s', got '%s'", 
					tt.description, tt.expectedMsg, errorMsg)
			}

			if tt.shouldPass && errorMsg != "" {
				t.Errorf("%s: expected no error message for passing test, got '%s'", 
					tt.description, errorMsg)
			}
		})
	}
}

func TestPCIeWidthResult(t *testing.T) {
	tests := []struct {
		name         string
		widthCounts  map[string]int
		success      bool
		errorMsg     string
		description  string
	}{
		{
			name: "Successful result",
			widthCounts: map[string]int{
				"Width x2":  4,
				"Width x16": 8,
			},
			success:     true,
			errorMsg:    "",
			description: "Should create successful PCIeWidthResult",
		},
		{
			name:        "Failed result",
			widthCounts: nil,
			success:     false,
			errorMsg:    "Test error message",
			description: "Should create failed PCIeWidthResult",
		},
		{
			name:        "Empty width counts",
			widthCounts: map[string]int{},
			success:     true,
			errorMsg:    "",
			description: "Should handle empty width counts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PCIeWidthResult{
				WidthCounts: tt.widthCounts,
				Success:     tt.success,
				ErrorMsg:    tt.errorMsg,
			}

			if result.Success != tt.success {
				t.Errorf("%s: expected Success=%v, got Success=%v", 
					tt.description, tt.success, result.Success)
			}

			if result.ErrorMsg != tt.errorMsg {
				t.Errorf("%s: expected ErrorMsg='%s', got ErrorMsg='%s'", 
					tt.description, tt.errorMsg, result.ErrorMsg)
			}

			if tt.widthCounts == nil && result.WidthCounts != nil {
				t.Errorf("%s: expected nil WidthCounts, got non-nil", tt.description)
			}

			if tt.widthCounts != nil {
				if len(result.WidthCounts) != len(tt.widthCounts) {
					t.Errorf("%s: expected %d width entries, got %d", 
						tt.description, len(tt.widthCounts), len(result.WidthCounts))
				}

				for width, expectedCount := range tt.widthCounts {
					if actualCount, exists := result.WidthCounts[width]; !exists {
						t.Errorf("%s: expected width %s not found", tt.description, width)
					} else if actualCount != expectedCount {
						t.Errorf("%s: width %s expected count %d, got %d", 
							tt.description, width, expectedCount, actualCount)
					}
				}
			}
		})
	}
}

func TestPCIeWidthMissingLanesTestConfig(t *testing.T) {
	config := &PCIeWidthMissingLanesTestConfig{
		IsEnabled: true,
		ExpectedGPUWidths: map[string]int{
			"Width x2":  4,
			"Width x16": 8,
		},
		ExpectedRDMAWidths: map[string]int{
			"Width x8":  2,
			"Width x16": 16,
		},
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}

	expectedGPUWidths := map[string]int{
		"Width x2":  4,
		"Width x16": 8,
	}

	if len(config.ExpectedGPUWidths) != len(expectedGPUWidths) {
		t.Errorf("Expected %d GPU width entries, got %d", 
			len(expectedGPUWidths), len(config.ExpectedGPUWidths))
	}

	for width, expectedCount := range expectedGPUWidths {
		if actualCount, exists := config.ExpectedGPUWidths[width]; !exists {
			t.Errorf("Expected GPU width %s not found", width)
		} else if actualCount != expectedCount {
			t.Errorf("GPU width %s expected count %d, got %d", width, expectedCount, actualCount)
		}
	}

	expectedRDMAWidths := map[string]int{
		"Width x8":  2,
		"Width x16": 16,
	}

	if len(config.ExpectedRDMAWidths) != len(expectedRDMAWidths) {
		t.Errorf("Expected %d RDMA width entries, got %d", 
			len(expectedRDMAWidths), len(config.ExpectedRDMAWidths))
	}

	for width, expectedCount := range expectedRDMAWidths {
		if actualCount, exists := config.ExpectedRDMAWidths[width]; !exists {
			t.Errorf("Expected RDMA width %s not found", width)
		} else if actualCount != expectedCount {
			t.Errorf("RDMA width %s expected count %d, got %d", width, expectedCount, actualCount)
		}
	}
}

// Test edge cases for width parsing
func TestParseLspciWidthOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCounts map[string]int
		description    string
	}{
		{
			name:           "Single valid line",
			input:          "      4           LnkSta:     Speed 16GT/s (ok), Width x2 (ok)",
			expectedCounts: map[string]int{"Width x2": 4},
			description:    "Should parse single valid line",
		},
		{
			name: "Invalid count format",
			input: `      abc         LnkSta:     Speed 16GT/s (ok), Width x2 (ok)
      4           LnkSta:     Speed 32GT/s (ok), Width x16 (ok)`,
			expectedCounts: map[string]int{"Width x16": 4},
			description:    "Should skip lines with invalid count format",
		},
		{
			name:           "Line without LnkSta",
			input:          "      4           SomeOther:  Speed 16GT/s (ok), Width x2 (ok)",
			expectedCounts: map[string]int{},
			description:    "Should skip lines without LnkSta",
		},
		{
			name:           "Line without Width",
			input:          "      4           LnkSta:     Speed 16GT/s (ok), Speed x2 (ok)",
			expectedCounts: map[string]int{},
			description:    "Should skip lines without Width",
		},
		{
			name:           "Malformed Width",
			input:          "      4           LnkSta:     Speed 16GT/s (ok), Width 2 (ok)",
			expectedCounts: map[string]int{},
			description:    "Should skip lines with malformed Width (missing x)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLspciWidthOutput(tt.input, "ok")

			if len(result.WidthCounts) != len(tt.expectedCounts) {
				t.Errorf("%s: expected %d width entries, got %d", 
					tt.description, len(tt.expectedCounts), len(result.WidthCounts))
			}

			for width, expectedCount := range tt.expectedCounts {
				if actualCount, exists := result.WidthCounts[width]; !exists {
					if expectedCount > 0 {
						t.Errorf("%s: expected width %s not found in result", tt.description, width)
					}
				} else if actualCount != expectedCount {
					t.Errorf("%s: width %s expected count %d, got %d", 
						tt.description, width, expectedCount, actualCount)
				}
			}
		})
	}
}

// Benchmark the PCIe width parsing logic
func BenchmarkParseLspciWidthOutput(b *testing.B) {
	input := mockNvidiaLspciOutput + "\n" + mockMellanoxLspciOutput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseLspciWidthOutput(input, "ok")
	}
}

// Benchmark width validation logic
func BenchmarkValidateWidthCounts(b *testing.B) {
	actual := map[string]int{
		"Width x2":  4,
		"Width x8":  2,
		"Width x16": 24,
	}
	expected := map[string]int{
		"Width x2":  4,
		"Width x8":  2,
		"Width x16": 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateWidthCounts(actual, expected, "GPU/NVSwitch")
	}
}

// Test the new filtering functions
const mockLspciFullOutput = `00:00.0 Host bridge: Intel Corporation Device 7d01 (rev 11)
	Subsystem: Intel Corporation Device 7270
	Control: I/O- Mem+ BusMaster+ SpecCycle- MemWINV- VGASnoop- ParErr- Stepping- SERR- FastB2B- DisINTx-

17:00.0 3D controller: NVIDIA Corporation GH200 120GB [Grace Hopper "Superchip"] (rev a1)
	Subsystem: NVIDIA Corporation Device 1791
	Control: I/O+ Mem+ BusMaster+ SpecCycle- MemWINV- VGASnoop- ParErr- Stepping- SERR- FastB2B- DisINTx+
	Status: Cap+ 66MHz- UDF- FastB2B- ParErr- DEVSEL=fast >TAbort- <TAbort- <MAbort- >SERR- <PERR- INTx-
	Latency: 0
	Interrupt: pin A routed to IRQ 16
	Region 0: Memory at c6000000 (32-bit, non-prefetchable) [size=16M]
	Region 1: Memory at 1800000000 (64-bit, prefetchable) [size=32G]
	Region 3: Memory at 1900000000 (64-bit, prefetchable) [size=32M]
	Capabilities: [60] Power Management version 3
	Capabilities: [68] MSI: Enable+ Count=1/1 Maskable- 64bit+
	Capabilities: [78] Express (v2) Endpoint, MSI 00
		DevCap:	MaxPayload 256 bytes, PhantFunc 0, Latency L0s unlimited, L1 <64us
		DevCtl:	CorrErr+ NonFatalErr+ FatalErr+ UnsupReq+
		DevSta:	CorrErr- NonFatalErr- FatalErr- UnsupReq- AuxPwr- TransPend-
		LnkCap:	Port #0, Speed 16GT/s, Width x16, ASPM L0s L1, Exit Latency L0s <1us, L1 <4us
		LnkCtl:	ASPM Disabled; RCB 64 bytes Disabled- CommClk+
		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)

ca:00.0 Infiniband controller: Mellanox Technologies MT2910 Family [ConnectX-7]
	Subsystem: Mellanox Technologies Device 0051
	Control: I/O- Mem+ BusMaster+ SpecCycle- MemWINV- VGASnoop- ParErr- Stepping- SERR- FastB2B- DisINTx+
	Status: Cap+ 66MHz- UDF- FastB2B- ParErr- DEVSEL=fast >TAbort- <TAbort- <MAbort- >SERR- <PERR- INTx-
	Latency: 0, Cache Line Size: 64 bytes
	Interrupt: pin A routed to IRQ 16
	Region 0: Memory at c4000000 (64-bit, non-prefetchable) [size=32M]
	Capabilities: [60] Express (v2) Endpoint, MSI 00
		DevCap:	MaxPayload 512 bytes, PhantFunc 0, Latency L0s <64ns, L1 <1us
		DevCtl:	CorrErr+ NonFatalErr+ FatalErr+ UnsupReq+
		DevSta:	CorrErr- NonFatalErr- FatalErr- UnsupReq- AuxPwr- TransPend-
		LnkCap:	Port #0, Speed 32GT/s, Width x16, ASPM not supported
		LnkCtl:	ASPM Disabled; RCB 128 bytes Disabled- CommClk+
		LnkSta:	Speed 32GT/s (ok), Width x16 (ok)`

func TestFilterLspciForDevice(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		deviceType     string
		expectedOutput string
		description    string
	}{
		{
			name:       "Filter NVIDIA devices",
			input:      mockLspciFullOutput,
			deviceType: "nvidia",
			expectedOutput: "		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)",
			description: "Should extract LnkSta line for NVIDIA device",
		},
		{
			name:       "Filter Mellanox devices",
			input:      mockLspciFullOutput,
			deviceType: "mellanox",
			expectedOutput: "		LnkSta:	Speed 32GT/s (ok), Width x16 (ok)",
			description: "Should extract LnkSta line for Mellanox device",
		},
		{
			name:        "No matching devices",
			input:       mockLspciFullOutput,
			deviceType:  "amd",
			expectedOutput: "",
			description: "Should return empty string when no matching devices found",
		},
		{
			name:        "Empty input",
			input:       "",
			deviceType:  "nvidia",
			expectedOutput: "",
			description: "Should handle empty input gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterLspciForDevice(tt.input, tt.deviceType)
			
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expectedOutput) {
				t.Errorf("%s: expected output '%s', got '%s'", 
					tt.description, strings.TrimSpace(tt.expectedOutput), strings.TrimSpace(result))
			}
		})
	}
}

func TestAggregateLinkStatistics(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedLines  int
		description    string
	}{
		{
			name: "Single unique line",
			input: "		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)",
			expectedLines: 1,
			description: "Should return single line with count 1",
		},
		{
			name: "Multiple identical lines", 
			input: `		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)
		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)
		LnkSta:	Speed 16GT/s (ok), Width x16 (ok)`,
			expectedLines: 1,
			description: "Should aggregate identical lines",
		},
		{
			name: "Multiple different lines",
			input: `		LnkSta:	Speed 16GT/s (ok), Width x2 (ok)
		LnkSta:	Speed 32GT/s (ok), Width x16 (ok)`,
			expectedLines: 2,
			description: "Should keep different lines separate",
		},
		{
			name: "Empty input",
			input: "",
			expectedLines: 0,
			description: "Should handle empty input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateLinkStatistics(tt.input)
			
			if tt.expectedLines == 0 {
				if result != "" {
					t.Errorf("%s: expected empty result, got '%s'", tt.description, result)
				}
				return
			}
			
			lines := strings.Split(strings.TrimSpace(result), "\n")
			if len(lines) != tt.expectedLines {
				t.Errorf("%s: expected %d lines, got %d", tt.description, tt.expectedLines, len(lines))
			}
			
			// Verify each line starts with a count and tab
			for _, line := range lines {
				if line != "" {
					parts := strings.SplitN(line, "\t", 2)
					if len(parts) != 2 {
						t.Errorf("%s: line should have count + tab + content format: '%s'", tt.description, line)
					}
				}
			}
		})
	}
}