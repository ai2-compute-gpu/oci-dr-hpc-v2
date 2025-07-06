package shapes

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewShapeManager(t *testing.T) {
	// Test with the actual shapes.json file
	shapesFile := filepath.Join(".", "shapes.json")

	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	if manager == nil {
		t.Fatal("ShapeManager should not be nil")
	}

	if manager.config == nil {
		t.Fatal("Config should not be nil")
	}
}

func TestGetAllShapes(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	shapes := manager.GetAllShapes()
	if len(shapes) == 0 {
		t.Fatal("Should have at least one shape")
	}

	// Check for some expected shapes
	expectedShapes := []string{
		"BM.GPU.H100.8",
		"BM.GPU.H200.8",
		"BM.GPU.B200.8",
		"BM.GPU.GB200.4",
	}

	allShapesMap := make(map[string]bool)
	for _, shape := range shapes {
		allShapesMap[shape] = true
	}

	for _, expected := range expectedShapes {
		if !allShapesMap[expected] {
			t.Errorf("Expected shape %s not found in shapes list", expected)
		}
	}

	t.Logf("Found %d total shapes", len(shapes))
}

func TestGetShapeConfig(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	tests := []struct {
		name        string
		shapeName   string
		expectError bool
	}{
		{
			name:        "valid H100 shape",
			shapeName:   "BM.GPU.H100.8",
			expectError: false,
		},
		{
			name:        "valid B200 shape",
			shapeName:   "BM.GPU.B200.8",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := manager.GetShapeConfig(tt.shapeName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if config == nil {
					t.Fatal("Config should not be nil for valid shape")
				}
				if config.Model == "" {
					t.Error("Model should not be empty")
				}

				// Check that the shape is in the shapes list
				found := false
				for _, shape := range config.Shapes {
					if shape == tt.shapeName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Shape %s not found in config.Shapes", tt.shapeName)
				}
			}
		})
	}
}

func TestGetShapesByModel(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	tests := []struct {
		name         string
		model        string
		expectShapes bool
	}{
		{
			name:         "ConnectX-7 model",
			model:        "ConnectX-7",
			expectShapes: true,
		},
		{
			name:         "ConnectX-5 model",
			model:        "ConnectX-5",
			expectShapes: true,
		},
		{
			name:         "Non-existent model",
			model:        "NonExistentModel",
			expectShapes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shapes := manager.GetShapesByModel(tt.model)

			if tt.expectShapes && len(shapes) == 0 {
				t.Errorf("Expected shapes for model %s but got none", tt.model)
			}
			if !tt.expectShapes && len(shapes) > 0 {
				t.Errorf("Expected no shapes for model %s but got %d", tt.model, len(shapes))
			}

			if len(shapes) > 0 {
				t.Logf("Found %d shapes for model %s: %v", len(shapes), tt.model, shapes)
			}
		})
	}
}

func TestGetSupportedModels(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	models := manager.GetSupportedModels()
	if len(models) == 0 {
		t.Fatal("Should have at least one supported model")
	}

	// Check for expected models
	expectedModels := []string{
		"ConnectX-7",
		"ConnectX-5",
	}

	modelMap := make(map[string]bool)
	for _, model := range models {
		modelMap[model] = true
	}

	for _, expected := range expectedModels {
		found := false
		for model := range modelMap {
			if strings.Contains(model, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model containing %s not found", expected)
		}
	}

	t.Logf("Found %d supported models: %v", len(models), models)
}

func TestIsShapeSupported(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	tests := []struct {
		name      string
		shapeName string
		expected  bool
	}{
		{
			name:      "supported H100 shape",
			shapeName: "BM.GPU.H100.8",
			expected:  true,
		},
		{
			name:      "supported B200 shape",
			shapeName: "BM.GPU.B200.8",
			expected:  true,
		},
		{
			name:      "unsupported shape",
			shapeName: "BM.INVALID.SHAPE",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsShapeSupported(tt.shapeName)
			if result != tt.expected {
				t.Errorf("Expected %v for shape %s, got %v", tt.expected, tt.shapeName, result)
			}
		})
	}
}

func TestGetGPUShapes(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	gpuShapes := manager.GetGPUShapes()
	if len(gpuShapes) == 0 {
		t.Fatal("Should have at least one GPU shape")
	}

	// All returned shapes should contain "GPU"
	for _, shape := range gpuShapes {
		if !strings.Contains(shape, "GPU") {
			t.Errorf("Shape %s does not contain 'GPU'", shape)
		}
	}

	t.Logf("Found %d GPU shapes", len(gpuShapes))
}

func TestGetHPCShapes(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	hpcShapes := manager.GetHPCShapes()
	if len(hpcShapes) == 0 {
		t.Fatal("Should have at least one HPC shape")
	}

	// All returned shapes should contain "HPC"
	for _, shape := range hpcShapes {
		if !strings.Contains(shape, "HPC") {
			t.Errorf("Shape %s does not contain 'HPC'", shape)
		}
	}

	t.Logf("Found %d HPC shapes", len(hpcShapes))
}

func TestSearchShapes(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	tests := []struct {
		name          string
		query         string
		expectResults bool
	}{
		{
			name:          "search for H100",
			query:         "H100",
			expectResults: true,
		},
		{
			name:          "search for B200",
			query:         "B200",
			expectResults: true,
		},
		{
			name:          "case insensitive search",
			query:         "gpu",
			expectResults: true,
		},
		{
			name:          "search for non-existent",
			query:         "NONEXISTENT",
			expectResults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := manager.SearchShapes(tt.query)

			if tt.expectResults && len(results) == 0 {
				t.Errorf("Expected results for query %s but got none", tt.query)
			}
			if !tt.expectResults && len(results) > 0 {
				t.Errorf("Expected no results for query %s but got %d", tt.query, len(results))
			}

			if len(results) > 0 {
				t.Logf("Found %d shapes for query '%s': %v", len(results), tt.query, results)
			}
		})
	}
}

func TestGetShapeInfo(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	shapeName := "BM.GPU.H100.8"
	info, err := manager.GetShapeInfo(shapeName)
	if err != nil {
		t.Fatalf("Failed to get shape info: %v", err)
	}

	if info.Name != shapeName {
		t.Errorf("Expected name %s, got %s", shapeName, info.Name)
	}

	if info.Model == "" {
		t.Error("Model should not be empty")
	}

	if !info.IsGPU {
		t.Error("H100 shape should be marked as GPU")
	}

	if info.IsHPC {
		t.Error("H100 shape should not be marked as HPC")
	}

	t.Logf("Shape info: %s", info.String())
}

func TestGetShapeSettings(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	shapeName := "BM.GPU.H100.8"
	settings, err := manager.GetShapeSettings(shapeName)
	if err != nil {
		t.Fatalf("Failed to get shape settings: %v", err)
	}

	if settings == nil {
		t.Fatal("Settings should not be nil")
	}

	// Check some expected settings fields
	if settings.MTU == "" {
		t.Error("MTU should not be empty")
	}

	if settings.Channels == "" {
		t.Error("Channels should not be empty")
	}

	// Buffer may be empty for some shapes like H100
	t.Logf("MTU: %s, Channels: %s, Buffer length: %d, Ring RX: %s, Ring TX: %s",
		settings.MTU, settings.Channels, len(settings.Buffer), settings.Ring.RX, settings.Ring.TX)
}

func TestGetShapeModel(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	shapeName := "BM.GPU.H100.8"
	model, err := manager.GetShapeModel(shapeName)
	if err != nil {
		t.Fatalf("Failed to get shape model: %v", err)
	}

	if model == "" {
		t.Error("Model should not be empty")
	}

	// H100 shapes should use ConnectX-7
	if !strings.Contains(model, "ConnectX-7") {
		t.Errorf("Expected ConnectX-7 model for H100, got %s", model)
	}

	t.Logf("Model for %s: %s", shapeName, model)
}

func TestGetRDMANetworkConfig(t *testing.T) {
	shapesFile := filepath.Join(".", "shapes.json")
	manager, err := NewShapeManager(shapesFile)
	if err != nil {
		t.Fatalf("Failed to create ShapeManager: %v", err)
	}

	rdmaConfigs := manager.GetRDMANetworkConfig()
	if len(rdmaConfigs) == 0 {
		t.Fatal("Should have at least one RDMA network config")
	}

	// Check the first config has expected fields
	config := rdmaConfigs[0]
	if config.DefaultSettings.RDMANetwork == "" {
		t.Error("RDMA network should not be empty")
	}

	t.Logf("RDMA Network: %s", config.DefaultSettings.RDMANetwork)
}
