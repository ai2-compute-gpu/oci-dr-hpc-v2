package executor

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// NvidiaSMIResult represents the result of nvidia-smi execution
type NvidiaSMIResult struct {
	Available bool
	Output    string
	Error     string
}

// CheckNvidiaSMI checks if nvidia-smi is available and functional
func CheckNvidiaSMI() *NvidiaSMIResult {
	result := &NvidiaSMIResult{
		Available: false,
		Output:    "",
		Error:     "",
	}

	logger.Info("Checking nvidia-smi availability...")

	// Check if nvidia-smi exists in PATH
	_, err := exec.LookPath("nvidia-smi")
	if err != nil {
		result.Error = "nvidia-smi not found in PATH"
		logger.Error("nvidia-smi check failed:", result.Error)
		return result
	}

	logger.Info("nvidia-smi found in PATH, executing command...")

	// Execute nvidia-smi command
	cmd := exec.Command("nvidia-smi")
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = err.Error()
		result.Output = string(output)
		logger.Error("nvidia-smi execution failed:", err)
		logger.Error("nvidia-smi output:", string(output))
		return result
	}

	result.Available = true
	result.Output = string(output)

	logger.Info("nvidia-smi executed successfully")
	logger.Debug("nvidia-smi output:", result.Output)

	return result
}

// RunNvidiaSMIQuery runs nvidia-smi with specific query parameters
func RunNvidiaSMIQuery(query string) *NvidiaSMIResult {
	result := &NvidiaSMIResult{
		Available: false,
		Output:    "",
		Error:     "",
	}

	logger.Info("Running nvidia-smi query:", query)

	// Check if nvidia-smi exists
	_, err := exec.LookPath("nvidia-smi")
	if err != nil {
		result.Error = "nvidia-smi not found in PATH"
		logger.Error("nvidia-smi not available for query:", result.Error)
		return result
	}

	// Execute nvidia-smi with query
	cmd := exec.Command("nvidia-smi", "--query-gpu="+query, "--format=csv,noheader,nounits")
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = err.Error()
		result.Output = string(output)
		logger.Error("nvidia-smi query failed:", err)
		logger.Error("Query output:", string(output))
		return result
	}

	result.Available = true
	result.Output = strings.TrimSpace(string(output))

	logger.Info("nvidia-smi query completed successfully")
	logger.Debug("Query result:", result.Output)

	return result
}

// GetGPUCount extracts GPU count from nvidia-smi output
func GetGPUCount() (int, error) {
	result := CheckNvidiaSMI()

	if !result.Available {
		logger.Error("Cannot get GPU count: nvidia-smi not available")
		return 0, errors.New("nvidia-smi not available: " + result.Error)
	}

	// Count GPU entries in output - look for lines like "|   0  NVIDIA GeForce GTX 1650"
	lines := strings.Split(result.Output, "\n")
	gpuCount := 0

	for _, line := range lines {
		// Look for GPU device lines that start with "|   X  " where X is a number
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") && len(trimmed) > 10 {
			// Check if this line contains a GPU device (has GPU number and name)
			if strings.Contains(line, "GeForce") ||
				strings.Contains(line, "Tesla") ||
				strings.Contains(line, "Quadro") ||
				strings.Contains(line, "RTX") ||
				strings.Contains(line, "GTX") {
				// Additional check: make sure it's not a header line and contains device number pattern
				if !strings.Contains(line, "GPU  Name") &&
					!strings.Contains(line, "===") &&
					!strings.Contains(line, "NVIDIA-SMI") &&
					!strings.Contains(line, "Driver Version") {
					gpuCount++
					logger.Debug("Found GPU device line:", trimmed)
				}
			}
		}
	}

	logger.Info("Detected GPU count:", gpuCount)
	return gpuCount, nil
}

// main function for standalone testing
// Comment out this function when using as a package
//func main() {
//	logger.Info("Starting nvidia-smi standalone test...")
//
//	// Test basic nvidia-smi availability
//	logger.Info("=== Testing CheckNvidiaSMI ===")
//	result := CheckNvidiaSMI()
//	if result.Available {
//		logger.Info("✅ nvidia-smi is available and working")
//		logger.Info("Output preview:", result.Output[:minInt(len(result.Output), 200)]+"...")
//	} else {
//		logger.Error("❌ nvidia-smi failed:", result.Error)
//	}
//
//	// Test GPU count detection
//	logger.Info("=== Testing GetGPUCount ===")
//	gpuCount, err := GetGPUCount()
//	if err != nil {
//		logger.Error("❌ Failed to get GPU count:", err)
//	} else {
//		logger.Info("✅ GPU count detected:", gpuCount)
//	}
//
//	// Test specific queries
//	logger.Info("=== Testing RunNvidiaSMIQuery ===")
//	queries := []string{
//		"name",
//		"memory.total",
//		"temperature.gpu",
//		"name,memory.total,temperature.gpu",
//	}
//
//	for _, query := range queries {
//		logger.Info("Testing query:", query)
//		queryResult := RunNvidiaSMIQuery(query)
//		if queryResult.Available {
//			logger.Info("✅ Query successful:", queryResult.Output)
//		} else {
//			logger.Error("❌ Query failed:", queryResult.Error)
//		}
//	}
//
//	logger.Info("nvidia-smi standalone test completed")
//}

// Helper function for minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
