package executor

import (
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
