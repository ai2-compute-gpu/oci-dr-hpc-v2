package level1_tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Mock shapes.json content for testing RDMA NICs
const mockRDMAShapesJSON = `{
  "version": "test-version",
  "rdma-network": [],
  "rdma-settings": [],
  "hpc-shapes": [
    {
      "shape": "BM.GPU.H100.8",
      "gpu": [
        {"pci": "0000:0f:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 0, "module_id": 2},
        {"pci": "0000:2d:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 1, "module_id": 4},
        {"pci": "0000:44:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 2, "module_id": 3},
        {"pci": "0000:5b:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 3, "module_id": 1},
        {"pci": "0000:89:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 4, "module_id": 7},
        {"pci": "0000:a8:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 5, "module_id": 5},
        {"pci": "0000:c0:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 6, "module_id": 8},
        {"pci": "0000:d8:00.0", "model": "NVIDIA H100 80GB HBM3", "id": 7, "module_id": 6}
      ],
      "vcn-nics": [],
      "rdma-nics": [
        {"pci": "0000:0c:00.0", "interface": "", "device_name": "mlx5_0", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:0f:00.0", "gpu_id": "0"},
        {"pci": "0000:0c:00.1", "interface": "", "device_name": "mlx5_1", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:0f:00.0", "gpu_id": "0"},
        {"pci": "0000:2a:00.0", "interface": "", "device_name": "mlx5_3", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:2d:00.0", "gpu_id": "1"},
        {"pci": "0000:2a:00.1", "interface": "", "device_name": "mlx5_4", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:2d:00.0", "gpu_id": "1"},
        {"pci": "0000:41:00.0", "interface": "", "device_name": "mlx5_5", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:44:00.0", "gpu_id": "2"},
        {"pci": "0000:41:00.1", "interface": "", "device_name": "mlx5_6", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:44:00.0", "gpu_id": "2"},
        {"pci": "0000:58:00.0", "interface": "", "device_name": "mlx5_7", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:5b:00.0", "gpu_id": "3"},
        {"pci": "0000:58:00.1", "interface": "", "device_name": "mlx5_8", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:5b:00.0", "gpu_id": "3"},
        {"pci": "0000:86:00.0", "interface": "", "device_name": "mlx5_9", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:89:00.0", "gpu_id": "4"},
        {"pci": "0000:86:00.1", "interface": "", "device_name": "mlx5_10", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:89:00.0", "gpu_id": "4"},
        {"pci": "0000:a5:00.0", "interface": "", "device_name": "mlx5_12", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:a8:00.0", "gpu_id": "5"},
        {"pci": "0000:a5:00.1", "interface": "", "device_name": "mlx5_13", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:a8:00.0", "gpu_id": "5"},
        {"pci": "0000:bd:00.0", "interface": "", "device_name": "mlx5_14", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:c0:00.0", "gpu_id": "6"},
        {"pci": "0000:bd:00.1", "interface": "", "device_name": "mlx5_15", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:c0:00.0", "gpu_id": "6"},
        {"pci": "0000:d5:00.0", "interface": "", "device_name": "mlx5_16", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:d8:00.0", "gpu_id": "7"},
        {"pci": "0000:d5:00.1", "interface": "", "device_name": "mlx5_17", "model": "Mellanox Technologies MT2910 Family [ConnectX-7]", "gpu_pci": "0000:d8:00.0", "gpu_id": "7"}
      ]
    },
    {
      "shape": "BM.GPU4.8",
      "gpu": [
        {"pci": "0000:0f:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 0, "module_id": 7},
        {"pci": "0000:15:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 1, "module_id": 5}
      ],
      "vcn-nics": [],
      "rdma-nics": [
        {"pci": "0000:0c:00.0", "interface": "", "device_name": "mlx5_6", "model": "Mellanox Technologies MT28800 Family [ConnectX-5 Ex]", "gpu_pci": "0000:0f:00.0", "gpu_id": "0"},
        {"pci": "0000:0c:00.1", "interface": "", "device_name": "mlx5_7", "model": "Mellanox Technologies MT28800 Family [ConnectX-5 Ex]", "gpu_pci": "0000:0f:00.0", "gpu_id": "0"}
      ]
    },
    {
      "shape": "BM.HPC.E5.128",
      "gpu": false,
      "vcn-nics": [],
      "rdma-nics": []
    }
  ]
}`

// createTempRDMAShapesFile creates a temporary shapes.json file for testing
func createTempRDMAShapesFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	shapesFile := filepath.Join(tempDir, "shapes.json")

	err := os.WriteFile(shapesFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp shapes file: %v", err)
	}

	return shapesFile
}

// testGetExpectedRDMANicConfig is a test helper that mimics the actual function
func testGetExpectedRDMANicConfig(shapeName string, shapesFilePath string) (int, []string, error) {
	// Read the shapes.json file
	data, err := os.ReadFile(shapesFilePath)
	if err != nil {
		return 0, nil, err
	}

	// Parse the JSON
	var shapesConfig ShapesConfigRDMA
	if err := json.Unmarshal(data, &shapesConfig); err != nil {
		return 0, nil, err
	}

	// Find the shape in hpc-shapes
	for _, shapeHW := range shapesConfig.HPCShapes {
		if shapeHW.Shape == shapeName {
			count := shapeHW.GetRDMANicCount()
			pciIDs := shapeHW.GetRDMANicPCIIDs()
			return count, pciIDs, nil
		}
	}

	return 0, nil, os.ErrNotExist
}

func TestGetExpectedRDMANicConfig(t *testing.T) {
	tests := []struct {
		name          string
		shapeName     string
		shapesContent string
		expectedCount int
		expectedPCIs  int
		expectError   bool
	}{
		{
			name:          "BM.GPU.H100.8 should return 16 RDMA NICs",
			shapeName:     "BM.GPU.H100.8",
			shapesContent: mockRDMAShapesJSON,
			expectedCount: 16,
			expectedPCIs:  16,
			expectError:   false,
		},
		{
			name:          "BM.GPU4.8 should return 2 RDMA NICs",
			shapeName:     "BM.GPU4.8",
			shapesContent: mockRDMAShapesJSON,
			expectedCount: 2,
			expectedPCIs:  2,
			expectError:   false,
		},
		{
			name:          "BM.HPC.E5.128 should return 0 RDMA NICs",
			shapeName:     "BM.HPC.E5.128",
			shapesContent: mockRDMAShapesJSON,
			expectedCount: 0,
			expectedPCIs:  0,
			expectError:   false,
		},
		{
			name:          "Non-existent shape should return error",
			shapeName:     "BM.NONEXISTENT",
			shapesContent: mockRDMAShapesJSON,
			expectedCount: 0,
			expectedPCIs:  0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary shapes file
			shapesFile := createTempRDMAShapesFile(t, tt.shapesContent)

			// Test the function
			count, pciIDs, err := testGetExpectedRDMANicConfig(tt.shapeName, shapesFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if count != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
				}
				if len(pciIDs) != tt.expectedPCIs {
					t.Errorf("Expected %d PCI IDs, got %d", tt.expectedPCIs, len(pciIDs))
				}
			}
		})
	}
}

func TestRDMANicStructUnmarshaling(t *testing.T) {
	// Test RDMA NIC struct marshaling/unmarshaling
	nic := RDMANic{
		PCI:        "0000:0c:00.0",
		Interface:  "enp12s0f0",
		DeviceName: "mlx5_0",
		Model:      "Mellanox Technologies MT2910 Family [ConnectX-7]",
		GPUPCI:     "0000:0f:00.0",
		GPUID:      "0",
	}

	data, err := json.Marshal(nic)
	if err != nil {
		t.Fatalf("Failed to marshal RDMA NIC struct: %v", err)
	}

	var unmarshaledNic RDMANic
	err = json.Unmarshal(data, &unmarshaledNic)
	if err != nil {
		t.Fatalf("Failed to unmarshal RDMA NIC struct: %v", err)
	}

	if unmarshaledNic != nic {
		t.Errorf("RDMA NIC struct mismatch after marshal/unmarshal: got %+v, expected %+v", unmarshaledNic, nic)
	}
}

func TestShapeHardwareRDMAMethods(t *testing.T) {
	// Test ShapeHardwareRDMA struct methods
	rdmaNics := []RDMANic{
		{PCI: "0000:0c:00.0", DeviceName: "mlx5_0"},
		{PCI: "0000:0c:00.1", DeviceName: "mlx5_1"},
		{PCI: "0000:2a:00.0", DeviceName: "mlx5_3"},
	}

	shape := ShapeHardwareRDMA{
		Shape:    "TEST.SHAPE",
		RDMANics: &rdmaNics,
	}

	// Test GetRDMANicCount
	count := shape.GetRDMANicCount()
	if count != 3 {
		t.Errorf("Expected RDMA NIC count 3, got %d", count)
	}

	// Test GetRDMANicPCIIDs
	pciIDs := shape.GetRDMANicPCIIDs()
	expectedPCIs := []string{"0000:0c:00.0", "0000:0c:00.1", "0000:2a:00.0"}
	if len(pciIDs) != len(expectedPCIs) {
		t.Errorf("Expected %d PCI IDs, got %d", len(expectedPCIs), len(pciIDs))
	}

	for i, expected := range expectedPCIs {
		if i >= len(pciIDs) || pciIDs[i] != expected {
			t.Errorf("Expected PCI ID %s at index %d, got %s", expected, i, pciIDs[i])
		}
	}

	// Test with nil RDMANics
	emptyShape := ShapeHardwareRDMA{
		Shape:    "EMPTY.SHAPE",
		RDMANics: nil,
	}

	if emptyShape.GetRDMANicCount() != 0 {
		t.Errorf("Expected 0 RDMA NICs for empty shape, got %d", emptyShape.GetRDMANicCount())
	}

	if len(emptyShape.GetRDMANicPCIIDs()) != 0 {
		t.Errorf("Expected empty PCI ID list for empty shape, got %v", emptyShape.GetRDMANicPCIIDs())
	}
}

func TestRDMANicsCountResult(t *testing.T) {
	// Test RDMANicsCountResult struct
	result := RDMANicsCountResult{
		NumRDMANics: 16,
		Status:      "PASS",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal RDMANicsCountResult: %v", err)
	}

	var unmarshaledResult RDMANicsCountResult
	err = json.Unmarshal(data, &unmarshaledResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal RDMANicsCountResult: %v", err)
	}

	if unmarshaledResult != result {
		t.Errorf("RDMANicsCountResult mismatch after marshal/unmarshal: got %+v, expected %+v", unmarshaledResult, result)
	}
}

func TestShapeHardwareRDMAUnmarshaling(t *testing.T) {
	// Test that our structs can properly unmarshal the JSON
	var shapesConfig ShapesConfigRDMA
	err := json.Unmarshal([]byte(mockRDMAShapesJSON), &shapesConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal mock shapes JSON: %v", err)
	}

	// Verify we have the expected shapes
	expectedShapes := map[string]int{
		"BM.GPU.H100.8": 16,
		"BM.GPU4.8":     2,
		"BM.HPC.E5.128": 0,
	}

	if len(shapesConfig.HPCShapes) != len(expectedShapes) {
		t.Errorf("Expected %d shapes, got %d", len(expectedShapes), len(shapesConfig.HPCShapes))
	}

	for _, shape := range shapesConfig.HPCShapes {
		expectedCount, exists := expectedShapes[shape.Shape]
		if !exists {
			t.Errorf("Unexpected shape found: %s", shape.Shape)
			continue
		}

		actualCount := shape.GetRDMANicCount()
		if actualCount != expectedCount {
			t.Errorf("Shape %s: expected %d RDMA NICs, got %d", shape.Shape, expectedCount, actualCount)
		}

		// Verify PCI IDs are properly extracted
		pciIDs := shape.GetRDMANicPCIIDs()
		if len(pciIDs) != expectedCount {
			t.Errorf("Shape %s: expected %d PCI IDs, got %d", shape.Shape, expectedCount, len(pciIDs))
		}

		// For H100.8, verify some specific PCI addresses
		if shape.Shape == "BM.GPU.H100.8" {
			expectedPCIs := []string{"0000:0c:00.0", "0000:0c:00.1", "0000:2a:00.0", "0000:2a:00.1"}
			for i, expectedPCI := range expectedPCIs {
				if i < len(pciIDs) && pciIDs[i] != expectedPCI {
					t.Errorf("H100.8 PCI ID at index %d: expected %s, got %s", i, expectedPCI, pciIDs[i])
				}
			}
		}
	}
}

// Benchmark tests for performance
func BenchmarkJSONParsingRDMA(b *testing.B) {
	jsonData := []byte(mockRDMAShapesJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var shapesConfig ShapesConfigRDMA
		json.Unmarshal(jsonData, &shapesConfig)
	}
}

func BenchmarkRDMAShapeSearch(b *testing.B) {
	var shapesConfig ShapesConfigRDMA
	json.Unmarshal([]byte(mockRDMAShapesJSON), &shapesConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, shapeHW := range shapesConfig.HPCShapes {
			if shapeHW.Shape == "BM.GPU.H100.8" {
				_ = shapeHW.GetRDMANicCount()
				_ = shapeHW.GetRDMANicPCIIDs()
				break
			}
		}
	}
}
