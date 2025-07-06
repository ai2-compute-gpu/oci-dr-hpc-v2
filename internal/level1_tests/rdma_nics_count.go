package level1_tests

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
)

// RDMANic represents an RDMA NIC configuration in shapes.json
type RDMANic struct {
	PCI        string `json:"pci"`
	Interface  string `json:"interface"`
	DeviceName string `json:"device_name"`
	Model      string `json:"model"`
	GPUPCI     string `json:"gpu_pci,omitempty"`
	GPUID      string `json:"gpu_id,omitempty"`
}

// ShapeHardwareRDMA represents the hardware configuration for a shape (RDMA-focused)
type ShapeHardwareRDMA struct {
	Shape    string        `json:"shape"`
	GPU      *[]GPUInfo    `json:"-"` // Handle manually
	VCNNics  []interface{} `json:"vcn-nics"`
	RDMANics *[]RDMANic    `json:"-"` // Handle manually
}

// GPUInfo represents a GPU configuration in shapes.json for RDMA processing
type GPUInfo struct {
	PCI      string `json:"pci"`
	Model    string `json:"model"`
	ID       int    `json:"id"`
	ModuleID int    `json:"module_id"`
}

// UnmarshalJSON handles the custom GPU and RDMA NIC fields
func (sh *ShapeHardwareRDMA) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to handle the dynamic fields
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
		// Try to unmarshal as RDMA NIC array
		var rdmaArray []RDMANic
		if err := json.Unmarshal(rdmaData, &rdmaArray); err == nil {
			sh.RDMANics = &rdmaArray
		} else {
			// If unmarshal fails, set to nil (indicates no RDMA NICs)
			sh.RDMANics = nil
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
		var gpuArray []GPUInfo
		if err := json.Unmarshal(gpuData, &gpuArray); err == nil {
			sh.GPU = &gpuArray
			return nil
		}

		// If neither worked, return error
		return fmt.Errorf("invalid GPU field format in shape %s", sh.Shape)
	}

	return nil
}

// GetRDMANicCount returns the number of RDMA NICs for this shape
func (sh *ShapeHardwareRDMA) GetRDMANicCount() int {
	if sh.RDMANics == nil {
		return 0
	}
	return len(*sh.RDMANics)
}

// GetRDMANicPCIIDs returns the PCI IDs of RDMA NICs for this shape
func (sh *ShapeHardwareRDMA) GetRDMANicPCIIDs() []string {
	if sh.RDMANics == nil {
		return []string{}
	}

	var pciIDs []string
	for _, nic := range *sh.RDMANics {
		pciIDs = append(pciIDs, nic.PCI)
	}
	return pciIDs
}

// ShapesConfigRDMA represents the structure of shapes.json for RDMA processing
type ShapesConfigRDMA struct {
	Version      string              `json:"version"`
	RDMANetwork  []interface{}       `json:"rdma-network"`
	RDMASettings []interface{}       `json:"rdma-settings"`
	HPCShapes    []ShapeHardwareRDMA `json:"hpc-shapes"`
}

// getExpectedRDMANicConfig reads shapes.json and returns the expected RDMA NIC count and PCI IDs for the given shape
func getExpectedRDMANicConfig(shapeName string) (int, []string, error) {
	shapesFilePath := "internal/shapes/shapes.json"

	// Read the shapes.json file
	data, err := os.ReadFile(shapesFilePath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read shapes.json: %w", err)
	}

	// Parse the JSON
	var shapesConfig ShapesConfigRDMA
	if err := json.Unmarshal(data, &shapesConfig); err != nil {
		return 0, nil, fmt.Errorf("failed to parse shapes.json: %w", err)
	}

	// Find the shape in hpc-shapes
	for _, shapeHW := range shapesConfig.HPCShapes {
		if shapeHW.Shape == shapeName {
			count := shapeHW.GetRDMANicCount()
			pciIDs := shapeHW.GetRDMANicPCIIDs()
			return count, pciIDs, nil
		}
	}

	return 0, nil, fmt.Errorf("shape %s not found in shapes.json", shapeName)
}

// getActualRDMANicCount uses lspci to check the actual number of RDMA NICs at the specified PCI addresses
func getActualRDMANicCount(expectedPCIIDs []string) (int, error) {
	detectedNics := 0

	// Check each expected RDMA device location
	for _, pciID := range expectedPCIIDs {
		logger.Debugf("Checking PCI device: %s", pciID)

		// Use lspci to query the specific device
		result, err := executor.RunLspciForDevice(pciID, true)
		if err != nil {
			logger.Errorf("Error running lspci for device %s: %v", pciID, err)
			continue
		}

		// Look through each line of the lspci output for this device
		lines := strings.Split(result.Output, "\n")
		for _, line := range lines {
			// Look for the Mellanox manufacturer identifier
			// Mellanox Technologies makes the RDMA controllers used in H100 systems
			if strings.Contains(line, "controller: Mellanox Technologies") {
				detectedNics++
				logger.Debugf("Found RDMA NIC at %s: %s", pciID, strings.TrimSpace(line))
				break // Found the controller for this device, move to next device
			}
		}
	}

	return detectedNics, nil
}

// RDMANicsCountResult represents the result of RDMA NIC count check
type RDMANicsCountResult struct {
	NumRDMANics int    `json:"num_rdma_nics"`
	Status      string `json:"status"`
}

// GetRDMANicsCountResult performs the RDMA NIC count check and returns structured result
func GetRDMANicsCountResult() (*RDMANicsCountResult, error) {
	// Step 1: Get shape from IMDS
	shape, err := executor.GetCurrentShape()
	if err != nil {
		return &RDMANicsCountResult{
			NumRDMANics: 0,
			Status:      "FAIL",
		}, fmt.Errorf("failed to get shape from IMDS: %w", err)
	}

	// Step 2: Look up for Shape and corresponding RDMA NICs - get count and PCI IDs
	expectedCount, expectedPCIIDs, err := getExpectedRDMANicConfig(shape)
	if err != nil {
		return &RDMANicsCountResult{
			NumRDMANics: 0,
			Status:      "FAIL",
		}, fmt.Errorf("failed to get expected RDMA NIC configuration: %w", err)
	}

	// Step 3: Using PCI IDs from Step 2, find the RDMA NICs count from OS using lspci
	actualCount, err := getActualRDMANicCount(expectedPCIIDs)
	if err != nil {
		return &RDMANicsCountResult{
			NumRDMANics: actualCount,
			Status:      "FAIL",
		}, fmt.Errorf("failed to get actual RDMA NIC count: %w", err)
	}

	// Step 4: Compare expected vs actual
	if expectedCount == actualCount {
		return &RDMANicsCountResult{
			NumRDMANics: actualCount,
			Status:      "PASS",
		}, nil
	} else {
		return &RDMANicsCountResult{
			NumRDMANics: actualCount,
			Status:      "FAIL",
		}, fmt.Errorf("RDMA NIC count mismatch: expected %d, actual %d", expectedCount, actualCount)
	}
}

func RunRDMANicsCount() error {
	logger.Info("=== RDMA NICs Count Check ===")
	rep := reporter.GetReporter()

	// Step 1: Get shape from IMDS
	logger.Info("Step 1: Getting shape from IMDS...")
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Error("RDMA NIC Count Check: FAIL - Could not get shape from IMDS:", err)
		rep.AddRDMAResult("FAIL", 0, err)
		return fmt.Errorf("failed to get shape from IMDS: %w", err)
	}
	logger.Info("Current shape from IMDS:", shape)

	// Step 2: Look up for Shape and corresponding RDMA NICs - get count and PCI IDs
	logger.Info("Step 2: Getting expected RDMA NIC count and PCI IDs from shapes.json...")
	expectedCount, expectedPCIIDs, err := getExpectedRDMANicConfig(shape)
	if err != nil {
		logger.Error("RDMA NIC Count Check: FAIL - Could not get expected RDMA NIC configuration:", err)
		rep.AddRDMAResult("FAIL", 0, err)
		return fmt.Errorf("failed to get expected RDMA NIC configuration: %w", err)
	}
	logger.Info("Expected RDMA NIC count for shape", shape+":", expectedCount)
	logger.Debugf("Expected PCI IDs: %v", expectedPCIIDs)

	// Step 3: Using PCI IDs from Step 2, find the RDMA NICs count from OS using lspci
	logger.Info("Step 3: Checking actual RDMA NIC count using lspci...")
	actualCount, err := getActualRDMANicCount(expectedPCIIDs)
	if err != nil {
		logger.Error("RDMA NIC Count Check: FAIL - Could not get actual RDMA NIC count:", err)
		rep.AddRDMAResult("FAIL", actualCount, err)
		return fmt.Errorf("failed to get actual RDMA NIC count: %w", err)
	}
	logger.Info("Actual RDMA NIC count from lspci:", actualCount)

	// Step 4: Compare expected vs actual
	logger.Info("Step 4: Comparing expected vs actual RDMA NIC counts...")
	if expectedCount == actualCount {
		logger.Info("RDMA NIC Count Check: PASS - Expected:", expectedCount, "Actual:", actualCount)
		rep.AddRDMAResult("PASS", actualCount, nil)
		return nil
	} else {
		if actualCount < expectedCount {
			missingCount := expectedCount - actualCount
			logger.Error("RDMA NIC Count Check: FAIL - Missing", missingCount, "RDMA NICs")
		} else {
			extraCount := actualCount - expectedCount
			logger.Error("RDMA NIC Count Check: FAIL - Found", extraCount, "extra RDMA NICs")
		}
		logger.Error("RDMA NIC Count Check: FAIL - Expected:", expectedCount, "Actual:", actualCount)
		err = fmt.Errorf("RDMA NIC count mismatch: expected %d, actual %d", expectedCount, actualCount)
		rep.AddRDMAResult("FAIL", actualCount, err)
		return err
	}
}

// DemoRDMANicsCountResult demonstrates the expected output format
func DemoRDMANicsCountResult() {
	// Example result for H100.8 shape with successful check
	result := RDMANicsCountResult{
		NumRDMANics: 16,
		Status:      "PASS",
	}

	// Format as requested: 'rdma_nic_count': [{'num_rdma_nics': 16, 'status': 'PASS'}]
	resultMap := map[string][]RDMANicsCountResult{
		"rdma_nic_count": {result},
	}

	jsonOutput, _ := json.MarshalIndent(resultMap, "", "  ")
	fmt.Printf("Expected output format:\n%s\n", string(jsonOutput))
}
