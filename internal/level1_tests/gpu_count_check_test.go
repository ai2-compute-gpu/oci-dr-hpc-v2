package level1_tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock shapes.json content for testing
const mockShapesJSON = `{
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
      "rdma-nics": []
    },
    {
      "shape": "BM.GPU4.8",
      "gpu": [
        {"pci": "0000:0f:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 0, "module_id": 7},
        {"pci": "0000:15:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 1, "module_id": 5},
        {"pci": "0000:51:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 2, "module_id": 8},
        {"pci": "0000:54:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 3, "module_id": 6},
        {"pci": "0000:8d:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 4, "module_id": 3},
        {"pci": "0000:92:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 5, "module_id": 1},
        {"pci": "0000:d6:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 6, "module_id": 4},
        {"pci": "0000:da:00.0", "model": "NVIDIA A100-SXM4-40GB", "id": 7, "module_id": 2}
      ],
      "vcn-nics": [],
      "rdma-nics": []
    },
    {
      "shape": "BM.HPC.E5.128",
      "gpu": false,
      "vcn-nics": [],
      "rdma-nics": []
    },
    {
      "shape": "BM.GPU.L40S.4",
      "gpu": [
        {"pci": "0000:16:00.0", "model": "NVIDIA L40S", "id": 0},
        {"pci": "0000:38:00.0", "model": "NVIDIA L40S", "id": 1},
        {"pci": "0000:82:00.0", "model": "NVIDIA L40S", "id": 2},
        {"pci": "0000:ac:00.0", "model": "NVIDIA L40S", "id": 3}
      ],
      "vcn-nics": [],
      "rdma-nics": []
    }
  ]
}`

// createTempShapesFile creates a temporary shapes.json file for testing
func createTempShapesFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	shapesFile := filepath.Join(tempDir, "shapes.json")

	err := os.WriteFile(shapesFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp shapes file: %v", err)
	}

	return shapesFile
}

// testGetExpectedGPUCount is a test helper that mimics the actual function
func testGetExpectedGPUCount(shapeName string, shapesFilePath string) (int, error) {
	// Read the shapes.json file
	data, err := os.ReadFile(shapesFilePath)
	if err != nil {
		return 0, err
	}

	// Parse the JSON
	var shapesConfig ShapesConfig
	if err := json.Unmarshal(data, &shapesConfig); err != nil {
		return 0, err
	}

	// Find the shape in hpc-shapes
	for _, shapeHW := range shapesConfig.HPCShapes {
		if shapeHW.Shape == shapeName {
			return shapeHW.GetGPUCount(), nil
		}
	}

	return 0, os.ErrNotExist
}

func TestGetExpectedGPUCountLogic(t *testing.T) {
	tests := []struct {
		name          string
		shapeName     string
		shapesContent string
		expectedCount int
		expectError   bool
		errorContains string
	}{
		{
			name:          "BM.GPU.H100.8 should return 8 GPUs",
			shapeName:     "BM.GPU.H100.8",
			shapesContent: mockShapesJSON,
			expectedCount: 8,
			expectError:   false,
		},
		{
			name:          "BM.GPU4.8 should return 8 GPUs",
			shapeName:     "BM.GPU4.8",
			shapesContent: mockShapesJSON,
			expectedCount: 8,
			expectError:   false,
		},
		{
			name:          "BM.GPU.L40S.4 should return 4 GPUs",
			shapeName:     "BM.GPU.L40S.4",
			shapesContent: mockShapesJSON,
			expectedCount: 4,
			expectError:   false,
		},
		{
			name:          "BM.HPC.E5.128 should return 0 GPUs (HPC shape)",
			shapeName:     "BM.HPC.E5.128",
			shapesContent: mockShapesJSON,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Non-existent shape should return error",
			shapeName:     "BM.GPU.NONEXISTENT",
			shapesContent: mockShapesJSON,
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "Empty shape name should return error",
			shapeName:     "",
			shapesContent: mockShapesJSON,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary shapes file
			shapesFile := createTempShapesFile(t, tt.shapesContent)

			// Test the function
			count, err := testGetExpectedGPUCount(tt.shapeName, shapesFile)

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
			}
		})
	}
}

func TestMalformedJSON(t *testing.T) {
	malformedJSON := `{"invalid": json}`
	shapesFile := createTempShapesFile(t, malformedJSON)

	_, err := testGetExpectedGPUCount("BM.GPU.H100.8", shapesFile)
	if err == nil {
		t.Error("Expected error for malformed JSON")
	}
}

func TestFileNotFound(t *testing.T) {
	_, err := testGetExpectedGPUCount("BM.GPU.H100.8", "/nonexistent/path/shapes.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetActualGPUCountParsing(t *testing.T) {
	tests := []struct {
		name            string
		nvidiaSMIOutput string
		expectedCount   int
		description     string
	}{
		{
			name: "8 GPUs output",
			nvidiaSMIOutput: `NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3
NVIDIA H100 80GB HBM3`,
			expectedCount: 8,
			description:   "Should count 8 H100 GPUs",
		},
		{
			name: "4 GPUs output",
			nvidiaSMIOutput: `NVIDIA L40S
NVIDIA L40S
NVIDIA L40S
NVIDIA L40S`,
			expectedCount: 4,
			description:   "Should count 4 L40S GPUs",
		},
		{
			name:            "Empty output",
			nvidiaSMIOutput: "",
			expectedCount:   0,
			description:     "Should return 0 for empty output",
		},
		{
			name:            "Whitespace only output",
			nvidiaSMIOutput: "   \n\t  \n  ",
			expectedCount:   0,
			description:     "Should return 0 for whitespace-only output",
		},
		{
			name:            "Single GPU output",
			nvidiaSMIOutput: `NVIDIA GeForce RTX 3080`,
			expectedCount:   1,
			description:     "Should count 1 GPU",
		},
		{
			name: "Mixed GPU types output",
			nvidiaSMIOutput: `NVIDIA A100-SXM4-40GB
NVIDIA Tesla V100-SXM2-16GB
NVIDIA GeForce RTX 3090`,
			expectedCount: 3,
			description:   "Should count mixed GPU types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parsing logic directly (mimics getActualGPUCount logic)
			output := strings.TrimSpace(tt.nvidiaSMIOutput)
			var count int
			if output == "" {
				count = 0
			} else {
				lines := strings.Split(output, "\n")
				count = len(lines)
			}

			if count != tt.expectedCount {
				t.Errorf("%s: expected %d GPUs, got %d", tt.description, tt.expectedCount, count)
			}
		})
	}
}

func TestShapeHardwareStructs(t *testing.T) {
	// Test that our structs can properly unmarshal the JSON
	var shapesConfig ShapesConfig
	err := json.Unmarshal([]byte(mockShapesJSON), &shapesConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal mock shapes JSON: %v", err)
	}

	// Verify we have the expected shapes
	expectedShapes := map[string]int{
		"BM.GPU.H100.8": 8,
		"BM.GPU4.8":     8,
		"BM.HPC.E5.128": 0,
		"BM.GPU.L40S.4": 4,
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

		actualCount := shape.GetGPUCount()

		if actualCount != expectedCount {
			t.Errorf("Shape %s: expected %d GPUs, got %d", shape.Shape, expectedCount, actualCount)
		}
	}
}

func TestGPUStruct(t *testing.T) {
	// Test GPU struct marshaling/unmarshaling
	gpu := GPU{
		PCI:      "0000:0f:00.0",
		Model:    "NVIDIA H100 80GB HBM3",
		ID:       0,
		ModuleID: 2,
	}

	data, err := json.Marshal(gpu)
	if err != nil {
		t.Fatalf("Failed to marshal GPU struct: %v", err)
	}

	var unmarshaledGPU GPU
	err = json.Unmarshal(data, &unmarshaledGPU)
	if err != nil {
		t.Fatalf("Failed to unmarshal GPU struct: %v", err)
	}

	if unmarshaledGPU != gpu {
		t.Errorf("GPU struct mismatch after marshal/unmarshal: got %+v, expected %+v", unmarshaledGPU, gpu)
	}
}

func TestPrintGPUCountCheck(t *testing.T) {
	// Test that PrintGPUCountCheck doesn't panic
	// This is a simple function that just logs, so we just verify it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintGPUCountCheck panicked: %v", r)
		}
	}()

	PrintGPUCountCheck()
}

// Test edge cases for shape name parsing
func TestShapeNameEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		shapeName  string
		shouldFind bool
	}{
		{
			name:       "Exact match",
			shapeName:  "BM.GPU.H100.8",
			shouldFind: true,
		},
		{
			name:       "Case sensitive mismatch",
			shapeName:  "bm.gpu.h100.8",
			shouldFind: false,
		},
		{
			name:       "Partial match",
			shapeName:  "H100.8",
			shouldFind: false,
		},
		{
			name:       "Empty string",
			shapeName:  "",
			shouldFind: false,
		},
		{
			name:       "Special characters",
			shapeName:  "BM.GPU.H100.8!@#",
			shouldFind: false,
		},
	}

	shapesFile := createTempShapesFile(t, mockShapesJSON)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read and parse the test file
			data, err := os.ReadFile(shapesFile)
			if err != nil {
				t.Fatalf("Failed to read shapes file: %v", err)
			}

			var shapesConfig ShapesConfig
			if err := json.Unmarshal(data, &shapesConfig); err != nil {
				t.Fatalf("Failed to parse shapes JSON: %v", err)
			}

			// Look for the shape
			found := false
			for _, shapeHW := range shapesConfig.HPCShapes {
				if shapeHW.Shape == tt.shapeName {
					found = true
					break
				}
			}

			if found != tt.shouldFind {
				t.Errorf("Shape %s: expected found=%v, got found=%v", tt.shapeName, tt.shouldFind, found)
			}
		})
	}
}

// Test the actual shapes.json file exists and is readable
func TestRealShapesFileExists(t *testing.T) {
	shapesPath := "internal/shapes/shapes.json"

	// Check if the file exists
	if _, err := os.Stat(shapesPath); os.IsNotExist(err) {
		t.Skip("Real shapes.json file not found, skipping integration test")
	}

	// Try to read and parse it
	data, err := os.ReadFile(shapesPath)
	if err != nil {
		t.Fatalf("Failed to read real shapes.json: %v", err)
	}

	var shapesConfig ShapesConfig
	if err := json.Unmarshal(data, &shapesConfig); err != nil {
		t.Fatalf("Failed to parse real shapes.json: %v", err)
	}

	// Verify it has some shapes
	if len(shapesConfig.HPCShapes) == 0 {
		t.Error("Real shapes.json should contain at least one shape")
	}
}

// Benchmark tests for performance
func BenchmarkJSONParsing(b *testing.B) {
	jsonData := []byte(mockShapesJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var shapesConfig ShapesConfig
		json.Unmarshal(jsonData, &shapesConfig)
	}
}

func BenchmarkShapeSearch(b *testing.B) {
	var shapesConfig ShapesConfig
	json.Unmarshal([]byte(mockShapesJSON), &shapesConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, shapeHW := range shapesConfig.HPCShapes {
			if shapeHW.Shape == "BM.GPU.H100.8" {
				_ = shapeHW.GetGPUCount()
				break
			}
		}
	}
}
