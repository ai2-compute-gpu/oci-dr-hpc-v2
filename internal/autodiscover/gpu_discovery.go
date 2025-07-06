package autodiscover

import (
	"fmt"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// DiscoverGPUs attempts to discover GPU information using nvidia-smi
func DiscoverGPUs() []GPU {
	logger.Info("Discovering GPU devices using nvidia-smi...")

	// Check if nvidia-smi is available
	if !executor.IsNvidiaSMIAvailable() {
		logger.Info("nvidia-smi not available, no GPU discovery possible")
		return []GPU{}
	}

	// Get GPU information from nvidia-smi
	gpuInfos, err := executor.GetGPUInfo()
	if err != nil {
		logger.Errorf("Failed to get GPU information from nvidia-smi: %v", err)
		return []GPU{}
	}

	// Convert executor.GPUInfo to autodiscover.GPU
	gpus := make([]GPU, len(gpuInfos))
	for i, info := range gpuInfos {
		gpus[i] = GPU{
			PCI:   info.PCI,
			Model: info.Model,
			ID:    fmt.Sprintf("%d", info.ID),
		}
	}

	logger.Infof("Successfully discovered %d GPU devices", len(gpus))
	for _, gpu := range gpus {
		logger.Debugf("Discovered GPU %s: %s at %s", gpu.ID, gpu.Model, gpu.PCI)
	}

	return gpus
}

// DiscoverGPUsWithFallback attempts to discover GPUs with fallback to undefined values when discovery fails
func DiscoverGPUsWithFallback() []GPU {
	// Try to discover real GPUs first
	gpus := DiscoverGPUs()

	// If no GPUs found but nvidia-smi is available, this might be a system without GPUs
	if len(gpus) == 0 && executor.IsNvidiaSMIAvailable() {
		logger.Info("No GPUs detected by nvidia-smi")
		return []GPU{}
	}

	// If nvidia-smi is not available, return undefined values to indicate discovery failed
	if len(gpus) == 0 {
		logger.Info("nvidia-smi not available, GPU discovery failed - using undefined values")
		return []GPU{
			{PCI: "undefined", Model: "undefined", ID: "undefined"},
		}
	}

	return gpus
}
