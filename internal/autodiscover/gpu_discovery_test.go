package autodiscover

import (
	"testing"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

func TestDiscoverGPUs(t *testing.T) {
	// This test depends on system configuration
	gpus := DiscoverGPUs()

	// Test should not crash, result depends on environment
	t.Logf("Discovered %d GPUs", len(gpus))

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

func TestDiscoverGPUsWithFallback(t *testing.T) {
	// This test should always return some result
	gpus := DiscoverGPUsWithFallback()

	t.Logf("Discovered %d GPUs (with fallback)", len(gpus))

	// Should have some GPUs either from real discovery or fallback
	if len(gpus) == 0 && executor.IsNvidiaSMIAvailable() {
		t.Log("No GPUs detected but nvidia-smi is available - this is valid for systems without GPUs")
	}

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

func TestGPUStructConversion(t *testing.T) {
	// Test that we can convert between executor.GPUInfo and autodiscover.GPU
	executorGPU := executor.GPUInfo{
		PCI:   "0000:0f:00.0",
		Model: "NVIDIA H100 80GB HBM3",
		ID:    0,
	}

	// Convert to autodiscover.GPU (simulating what DiscoverGPUs does)
	autodiscoverGPU := GPU{
		PCI:   executorGPU.PCI,
		Model: executorGPU.Model,
		ID:    executorGPU.ID,
	}

	// Verify conversion
	if autodiscoverGPU.PCI != executorGPU.PCI {
		t.Errorf("PCI mismatch: expected %s, got %s", executorGPU.PCI, autodiscoverGPU.PCI)
	}
	if autodiscoverGPU.Model != executorGPU.Model {
		t.Errorf("Model mismatch: expected %s, got %s", executorGPU.Model, autodiscoverGPU.Model)
	}
	if autodiscoverGPU.ID != executorGPU.ID {
		t.Errorf("ID mismatch: expected %d, got %d", executorGPU.ID, autodiscoverGPU.ID)
	}
}

func TestGPUStructFields(t *testing.T) {
	gpu := GPU{
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
