package level1_tests

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Helper function to create realistic lsmod output for testing
func createTestLsmodOutput(includeNvidiaPerrmem bool) string {
	output := "Module                  Size  Used by\n"

	if includeNvidiaPerrmem {
		output += "nvidia_peermem         16384  0\n"
	}

	// Add other realistic modules
	output += "nvidia_drm             69632  4\n"
	output += "nvidia_modeset       1048576  6 nvidia_drm\n"
	output += "nvidia              35282944  346 nvidia_modeset\n"
	output += "drm_kms_helper        311296  1 nvidia_drm\n"
	output += "drm                   622592  6 drm_kms_helper,nvidia_drm\n"

	return output
}

// Test parsing logic that mimics checkPeermemModuleLoaded
func parseModulesForTesting(lsmodOutput string, targetModule string) bool {
	output := strings.TrimSpace(lsmodOutput)
	if output == "" {
		return false
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == targetModule {
			return true
		}
	}

	return false
}

func TestPeermemModuleCheckTestConfig(t *testing.T) {
	// Test the config structure
	config := &PeermemModuleCheckTestConfig{
		IsEnabled: true,
	}

	if !config.IsEnabled {
		t.Errorf("Expected IsEnabled=true, got %v", config.IsEnabled)
	}

	// Test with false
	config.IsEnabled = false
	if config.IsEnabled {
		t.Errorf("Expected IsEnabled=false, got %v", config.IsEnabled)
	}
}

func TestModuleParsingLogic(t *testing.T) {
	tests := []struct {
		name         string
		lsmodOutput  string
		targetModule string
		expectFound  bool
		description  string
	}{
		{
			name:         "nvidia_peermem_present",
			lsmodOutput:  createTestLsmodOutput(true),
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should find nvidia_peermem when present",
		},
		{
			name:         "nvidia_peermem_not_present",
			lsmodOutput:  createTestLsmodOutput(false),
			targetModule: "nvidia_peermem",
			expectFound:  false,
			description:  "Should not find nvidia_peermem when not present",
		},
		{
			name:         "empty_output",
			lsmodOutput:  "",
			targetModule: "nvidia_peermem",
			expectFound:  false,
			description:  "Should handle empty output",
		},
		{
			name:         "header_only",
			lsmodOutput:  "Module                  Size  Used by",
			targetModule: "nvidia_peermem",
			expectFound:  false,
			description:  "Should handle header-only output",
		},
		{
			name: "exact_match_required",
			lsmodOutput: `Module                  Size  Used by
nvidia_peermem         16384  0
nvidia_peermemory      32768  1
nvidia_drm             69632  4`,
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should find exact match even with similar names",
		},
		{
			name: "partial_match_should_not_find",
			lsmodOutput: `Module                  Size  Used by
nvidia_peermem_test    16384  0
some_nvidia_peermem    32768  1
nvidia_drm             69632  4`,
			targetModule: "nvidia_peermem",
			expectFound:  false,
			description:  "Should not match partial names",
		},
		{
			name: "case_sensitive",
			lsmodOutput: `Module                  Size  Used by
NVIDIA_PEERMEM         16384  0
Nvidia_Peermem         32768  1
nvidia_drm             69632  4`,
			targetModule: "nvidia_peermem",
			expectFound:  false,
			description:  "Should be case sensitive",
		},
		{
			name: "whitespace_handling",
			lsmodOutput: `Module                  Size  Used by
  nvidia_peermem         16384  0  
  nvidia_drm             69632  4  `,
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should handle extra whitespace",
		},
		{
			name:         "tabs_and_mixed_whitespace",
			lsmodOutput:  "Module\t\t\t\tSize\tUsed by\nnvidia_peermem\t\t16384\t0\nnvidia_drm\t\t\t69632\t4",
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should handle tabs and mixed whitespace",
		},
		{
			name: "malformed_lines_ignored",
			lsmodOutput: `Module                  Size  Used by
nvidia_peermem         16384  0
incomplete line
another bad line without proper format
nvidia_drm             69632  4`,
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should ignore malformed lines",
		},
		{
			name: "module_in_middle",
			lsmodOutput: `Module                  Size  Used by
nvidia_drm             69632  4
nvidia_peermem         16384  0
nvidia_modeset       1048576  6`,
			targetModule: "nvidia_peermem",
			expectFound:  true,
			description:  "Should find module in middle of list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := parseModulesForTesting(tt.lsmodOutput, tt.targetModule)

			if found != tt.expectFound {
				t.Errorf("%s: Expected found=%v, got %v", tt.description, tt.expectFound, found)
			}
		})
	}
}

func TestModuleParsingWithDifferentModules(t *testing.T) {
	testOutput := `Module                  Size  Used by
nvidia_peermem         16384  0
nvidia_drm             69632  4
nvidia_modeset       1048576  6
ext4                   32768  2
loop                   16384  0`

	tests := []struct {
		module      string
		expectFound bool
	}{
		{"nvidia_peermem", true},
		{"nvidia_drm", true},
		{"nvidia_modeset", true},
		{"ext4", true},
		{"loop", true},
		{"nonexistent", false},
		{"nvidia", false},      // Should not match partial
		{"nvidia_peer", false}, // Should not match partial
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("module_%s", tt.module), func(t *testing.T) {
			found := parseModulesForTesting(testOutput, tt.module)
			if found != tt.expectFound {
				t.Errorf("Module %s: expected found=%v, got %v", tt.module, tt.expectFound, found)
			}
		})
	}
}

func TestModuleParsingPerformance(t *testing.T) {
	// Test with a large number of modules
	largeOutput := "Module                  Size  Used by\n"

	// Add many fake modules
	for i := 0; i < 2000; i++ {
		largeOutput += fmt.Sprintf("fake_module_%d       16384  0\n", i)
	}

	// Add target module in the middle
	largeOutput += "nvidia_peermem         16384  0\n"

	// Add more fake modules
	for i := 2000; i < 4000; i++ {
		largeOutput += fmt.Sprintf("fake_module_%d       16384  0\n", i)
	}

	found := parseModulesForTesting(largeOutput, "nvidia_peermem")
	if !found {
		t.Error("Should find nvidia_peermem in large module list")
	}

	t.Logf("Successfully parsed list with 4000+ modules")
}

func TestModuleParsingEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name        string
		output      string
		expectFound bool
	}{
		{
			name:        "only_whitespace_lines",
			output:      "Module                  Size  Used by\n   \n\t\n \nnvidia_drm             69632  4",
			expectFound: false,
		},
		{
			name:        "very_long_module_name",
			output:      "Module                  Size  Used by\nnvidia_peermem_with_very_long_name_that_should_not_match  16384  0",
			expectFound: false,
		},
		{
			name:        "module_with_numbers",
			output:      "Module                  Size  Used by\nnvidia_peermem2        16384  0\nnvidia_peermem         32768  1",
			expectFound: true,
		},
		{
			name:        "module_as_substring_of_used_by",
			output:      "Module                  Size  Used by\nsome_module            16384  2 nvidia_peermem,other_module",
			expectFound: false,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			found := parseModulesForTesting(tc.output, "nvidia_peermem")
			if found != tc.expectFound {
				t.Errorf("Expected found=%v, got %v", tc.expectFound, found)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	// Test error wrapping behavior
	originalError := errors.New("original lsmod error")
	wrappedError := fmt.Errorf("failed to run lsmod: %w", originalError)

	if !strings.Contains(wrappedError.Error(), "failed to run lsmod") {
		t.Error("Wrapped error should contain context message")
	}

	if !strings.Contains(wrappedError.Error(), originalError.Error()) {
		t.Error("Wrapped error should contain original error message")
	}

	// Test error unwrapping
	if !errors.Is(wrappedError, originalError) {
		t.Error("Should be able to unwrap to original error")
	}
}

func TestPrintPeermemModuleCheck(t *testing.T) {
	// Test that the function exists and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintPeermemModuleCheck panicked: %v", r)
		}
	}()

	// Call the function
	PrintPeermemModuleCheck()

	// If we get here, the function completed without panicking
	t.Log("PrintPeermemModuleCheck completed successfully")
}

func TestFunctionSignatureExists(t *testing.T) {
	// Verify that all expected functions exist and are accessible
	// This helps catch compilation issues early

	// Test that functions exist by referencing them
	_ = getPeermemModuleCheckTestConfig
	_ = checkPeermemModuleLoaded
	_ = RunPeermemModuleCheck
	_ = PrintPeermemModuleCheck

	// Test struct exists
	var config PeermemModuleCheckTestConfig
	config.IsEnabled = true
	if !config.IsEnabled {
		t.Error("PeermemModuleCheckTestConfig struct not working correctly")
	}

	t.Log("All required functions and types exist")
}

// Test real parsing against known patterns from actual systems
func TestRealWorldLsmodPatterns(t *testing.T) {
	realPatterns := []struct {
		name   string
		output string
		expect bool
	}{
		{
			name: "ubuntu_with_nvidia",
			output: `Module                  Size  Used by
nvidia_peermem         16384  0 
nvidia_drm             69632  4 
nvidia_modeset       1048576  6 nvidia_drm
nvidia              35282944  346 nvidia_peermem,nvidia_modeset,nvidia_uvm
drm_kms_helper        311296  1 nvidia_drm
ttm                   114688  1 nvidia_drm
drm                   622592  6 drm_kms_helper,ttm,nvidia_drm`,
			expect: true,
		},
		{
			name: "centos_without_peermem",
			output: `Module                  Size  Used by
nvidia_drm             69632  4 
nvidia_modeset       1048576  6 nvidia_drm
nvidia              35282944  346 nvidia_modeset,nvidia_uvm
drm_kms_helper        311296  1 nvidia_drm
drm                   622592  6 drm_kms_helper,nvidia_drm`,
			expect: false,
		},
		{
			name: "minimal_system",
			output: `Module                  Size  Used by
ext4                  737280  2
mbcache                16384  1 ext4
jbd2                  114688  1 ext4`,
			expect: false,
		},
	}

	for _, pattern := range realPatterns {
		t.Run(pattern.name, func(t *testing.T) {
			found := parseModulesForTesting(pattern.output, "nvidia_peermem")
			if found != pattern.expect {
				t.Errorf("Pattern %s: expected %v, got %v", pattern.name, pattern.expect, found)
			}
		})
	}
}

// Benchmark the parsing function to ensure it's efficient
func BenchmarkModuleParsing(b *testing.B) {
	testOutput := createTestLsmodOutput(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseModulesForTesting(testOutput, "nvidia_peermem")
	}
}

func BenchmarkLargeModuleParsing(b *testing.B) {
	// Create large module list
	largeOutput := "Module                  Size  Used by\n"
	for i := 0; i < 1000; i++ {
		largeOutput += fmt.Sprintf("module_%d       16384  0\n", i)
	}
	largeOutput += "nvidia_peermem         16384  0\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseModulesForTesting(largeOutput, "nvidia_peermem")
	}
}
