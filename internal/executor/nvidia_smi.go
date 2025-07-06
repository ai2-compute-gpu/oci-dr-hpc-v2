package executor

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// NvidiaSMIResult represents the result of nvidia-smi execution
type NvidiaSMIResult struct {
	Available bool
	Output    string
	Error     string
}

// GPUInfo represents information about a single GPU
type GPUInfo struct {
	PCI   string `json:"pci"`
	Model string `json:"model"`
	ID    int    `json:"id"`
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

// GetGPUInfo queries nvidia-smi for comprehensive GPU information
func GetGPUInfo() ([]GPUInfo, error) {
	logger.Info("Querying GPU information from nvidia-smi...")

	// Query for PCI bus info, GPU name, and index
	result := RunNvidiaSMIQuery("pci.bus_id,name,index")

	if !result.Available {
		return nil, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	if result.Output == "" {
		logger.Info("No GPUs detected by nvidia-smi")
		return []GPUInfo{}, nil
	}

	gpus, err := parseGPUInfo(result.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GPU information: %w", err)
	}

	logger.Infof("Successfully detected %d GPUs", len(gpus))
	return gpus, nil
}

// parseGPUInfo parses the nvidia-smi output into GPUInfo structs
func parseGPUInfo(output string) ([]GPUInfo, error) {
	var gpus []GPUInfo

	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split CSV line: pci.bus_id, name, index
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			logger.Errorf("Invalid GPU info line: %s", line)
			continue
		}

		// Clean up the parts
		pci := formatPCIAddress(strings.TrimSpace(parts[0]))
		model := strings.TrimSpace(parts[1])
		indexStr := strings.TrimSpace(parts[2])

		// Parse GPU index
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			logger.Errorf("Failed to parse GPU index '%s' on line %d: %v", indexStr, i+1, err)
			// Use line number as fallback
			index = i
		}

		gpu := GPUInfo{
			PCI:   pci,
			Model: model,
			ID:    index,
		}

		gpus = append(gpus, gpu)
		logger.Debugf("Parsed GPU %d: PCI=%s, Model=%s", index, pci, model)
	}

	return gpus, nil
}

// GetGPUCount queries nvidia-smi for the number of GPUs
func GetGPUCount() (int, error) {
	logger.Info("Querying GPU count from nvidia-smi...")

	result := RunNvidiaSMIQuery("count")

	if !result.Available {
		return 0, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// nvidia-smi returns just the count number
	countStr := strings.TrimSpace(result.Output)
	if countStr == "" {
		return 0, nil
	}

	// Handle both single count and multi-line output
	// On some systems, nvidia-smi returns one line per GPU with count
	lines := strings.Split(countStr, "\n")

	if len(lines) == 1 {
		// Single line format (e.g., "8")
		count, err := strconv.Atoi(strings.TrimSpace(lines[0]))
		if err != nil {
			return 0, fmt.Errorf("failed to parse GPU count '%s': %w", countStr, err)
		}
		logger.Infof("GPU count: %d", count)
		return count, nil
	}

	// Multi-line format - count non-empty lines
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			count++
		}
	}

	logger.Infof("GPU count: %d (from %d lines)", count, len(lines))
	return count, nil
}

// IsNvidiaSMIAvailable checks if nvidia-smi is available and working
func IsNvidiaSMIAvailable() bool {
	result := CheckNvidiaSMI()
	return result.Available
}

// formatPCIAddress converts nvidia-smi PCI format to standard PCI format
// Input: "00000000:65:00.0" (nvidia-smi format)
// Output: "0000:65:00.0" (standard [domain]:[bus]:[device].[function] format)
func formatPCIAddress(nvidiaPCI string) string {
	// Pattern to match nvidia-smi PCI format: 8-digit-domain:2-digit-bus:2-digit-device.1-digit-function
	pciPattern := regexp.MustCompile(`^([0-9a-fA-F]{8}):([0-9a-fA-F]{2}):([0-9a-fA-F]{2})\.([0-9a-fA-F])$`)

	matches := pciPattern.FindStringSubmatch(strings.TrimSpace(nvidiaPCI))
	if len(matches) != 5 {
		// If pattern doesn't match, return the original (lowercased)
		logger.Debugf("PCI address '%s' doesn't match expected format, returning as-is", nvidiaPCI)
		return strings.ToLower(nvidiaPCI)
	}

	// Extract components
	domain := matches[1]
	bus := matches[2]
	device := matches[3]
	function := matches[4]

	// Format domain to 4 digits (take last 4 digits)
	if len(domain) > 4 {
		domain = domain[len(domain)-4:]
	}

	// Format as standard PCI address: domain:bus:device.function
	formatted := fmt.Sprintf("%s:%s:%s.%s",
		strings.ToLower(domain),
		strings.ToLower(bus),
		strings.ToLower(device),
		strings.ToLower(function))

	logger.Debugf("Formatted PCI address: '%s' -> '%s'", nvidiaPCI, formatted)
	return formatted
}
