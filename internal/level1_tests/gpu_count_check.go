package level1_tests

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// GPU represents a GPU configuration in shapes.json
type GPU struct {
	PCI      string `json:"pci"`
	Model    string `json:"model"`
	ID       int    `json:"id"`
	ModuleID int    `json:"module_id"`
}

// ShapeHardware represents the hardware configuration for a shape
type ShapeHardware struct {
	Shape    string        `json:"shape"`
	GPU      *[]GPU        `json:"-"` // Handle manually
	VCNNics  []interface{} `json:"vcn-nics"`
	RDMANics []interface{} `json:"rdma-nics"`
}

// UnmarshalJSON handles the custom GPU field that can be false or an array
func (sh *ShapeHardware) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to handle the dynamic GPU field
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	// Handle shape field
	if shapeData, exists := rawData["shape"]; exists {
		if err := json.Unmarshal(shapeData, &sh.Shape); err != nil {
			return err
		}
	}

	// Handle vcn-nics field
	if vncData, exists := rawData["vcn-nics"]; exists {
		if err := json.Unmarshal(vncData, &sh.VCNNics); err != nil {
			return err
		}
	}

	// Handle rdma-nics field
	if rdmaData, exists := rawData["rdma-nics"]; exists {
		if err := json.Unmarshal(rdmaData, &sh.RDMANics); err != nil {
			return err
		}
	}

	// Handle GPU field specially
	if gpuData, exists := rawData["gpu"]; exists {
		// Try to unmarshal as boolean first
		var gpuBool bool
		if err := json.Unmarshal(gpuData, &gpuBool); err == nil {
			// It's a boolean
			if !gpuBool {
				sh.GPU = nil // false means no GPUs
			}
			return nil
		}

		// Try to unmarshal as GPU array
		var gpuArray []GPU
		if err := json.Unmarshal(gpuData, &gpuArray); err == nil {
			sh.GPU = &gpuArray
			return nil
		}

		// If neither worked, return error
		return fmt.Errorf("invalid GPU field format in shape %s", sh.Shape)
	}

	return nil
}

// GetGPUCount returns the number of GPUs for this shape
func (sh *ShapeHardware) GetGPUCount() int {
	if sh.GPU == nil {
		return 0
	}
	return len(*sh.GPU)
}

// ShapesConfig represents the structure of shapes.json
type ShapesConfig struct {
	Version      string          `json:"version"`
	RDMANetwork  []interface{}   `json:"rdma-network"`
	RDMASettings []interface{}   `json:"rdma-settings"`
	HPCShapes    []ShapeHardware `json:"hpc-shapes"`
}

func PrintGPUCountCheck() {
	// This function is a placeholder for GPU count check logic.
	// It should be implemented to check the number of GPUs available
	// and print the result.

	// Example implementation (to be replaced with actual logic):
	gpuCount := 4 // Placeholder value, replace with actual GPU count retrieval logic
	logger.Info("GPU Count Check: PASS - Number of GPUs detected:", gpuCount)
}

// getExpectedGPUCount reads shapes.json and returns the expected GPU count for the given shape
func getExpectedGPUCount(shapeName string) (int, error) {
	shapesFilePath := "internal/shapes/shapes.json"

	// Read the shapes.json file
	data, err := os.ReadFile(shapesFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read shapes.json: %w", err)
	}

	// Parse the JSON
	var shapesConfig ShapesConfig
	if err := json.Unmarshal(data, &shapesConfig); err != nil {
		return 0, fmt.Errorf("failed to parse shapes.json: %w", err)
	}

	// Find the shape in hpc-shapes
	for _, shapeHW := range shapesConfig.HPCShapes {
		if shapeHW.Shape == shapeName {
			return shapeHW.GetGPUCount(), nil
		}
	}

	return 0, fmt.Errorf("shape %s not found in shapes.json", shapeName)
}

// getActualGPUCount uses nvidia-smi to get the actual number of GPUs
func getActualGPUCount() (int, error) {
	// Use nvidia-smi to query GPU names and count them
	result := executor.RunNvidiaSMIQuery("name")
	if !result.Available {
		return 0, fmt.Errorf("nvidia-smi not available: %s", result.Error)
	}

	// Count the number of lines in the output (each line is a GPU)
	output := strings.TrimSpace(result.Output)
	if output == "" {
		return 0, nil
	}

	lines := strings.Split(output, "\n")
	return len(lines), nil
}

func RunGPUCountCheck() error {
	logger.Info("=== GPU Count Check ===")

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("GPU Count Check: FAIL - Could not get shape from IMDS:", err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Get expected GPU count from shapes.json
	logger.Info("Step 2: Getting expected GPU count from shapes.json...")
	expectedGPUCount, err := getExpectedGPUCount(shape)
	if err != nil {
		logger.Error("GPU Count Check: FAIL - Could not get expected GPU count:", err)
		return fmt.Errorf("failed to get expected GPU count: %w", err)
	}
	logger.Info("Expected GPU count for shape", shape+":", expectedGPUCount)

	// Step 3: Get actual GPU count from nvidia-smi
	logger.Info("Step 3: Getting actual GPU count from nvidia-smi...")
	actualGPUCount, err := getActualGPUCount()
	if err != nil {
		logger.Error("GPU Count Check: FAIL - Could not get actual GPU count:", err)
		return fmt.Errorf("failed to get actual GPU count: %w", err)
	}
	logger.Info("Actual GPU count from nvidia-smi:", actualGPUCount)

	// Step 4: Compare expected vs actual
	logger.Info("Step 4: Comparing expected vs actual GPU counts...")
	if expectedGPUCount == actualGPUCount {
		logger.Info("GPU Count Check: PASS - Expected:", expectedGPUCount, "Actual:", actualGPUCount)
		return nil
	} else {
		logger.Error("GPU Count Check: FAIL - Expected:", expectedGPUCount, "Actual:", actualGPUCount)
		return fmt.Errorf("GPU count mismatch: expected %d, actual %d", expectedGPUCount, actualGPUCount)
	}
}
