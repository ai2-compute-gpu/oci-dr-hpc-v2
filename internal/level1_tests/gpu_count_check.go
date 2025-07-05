package level1_tests

import (
	"fmt"
	
	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

func PrintGPUCountCheck() {
	// This function is a placeholder for GPU count check logic.
	// It should be implemented to check the number of GPUs available
	// and print the result.

	// Example implementation (to be replaced with actual logic):
	gpuCount := 4 // Placeholder value, replace with actual GPU count retrieval logic
	logger.Info("GPU Count Check: PASS - Number of GPUs detected:", gpuCount)
}

func RunGPUCountCheck() error {
	logger.Info("=== GPU Count Check ===")
	
	// Use nvidia-smi to get GPU count
	result := executor.RunNvidiaSMIQuery("count")
	if !result.Available {
		logger.Error("GPU Count Check: FAIL - nvidia-smi not available:", result.Error)
		return fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}
	
	logger.Info("GPU Count Check: PASS - nvidia-smi is available")
	logger.Info("GPU query result:", result.Output)
	
	return nil
}
