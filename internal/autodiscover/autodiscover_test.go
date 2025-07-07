package autodiscover

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	// Create a test MapHost
	testMapHost := &MapHost{
		Hostname:         "test-host",
		Ocid:             "ocid1.instance.oc1.test",
		FriendlyHostname: "test-host",
		Shape:            "BM.GPU.H100.8",
		Serial:           "TEST123",
		Rack:             "rack123",
		NetworkBlockId:   "netblock123",
		BuildingId:       "building123",
		InCluster:        true,
		Gpus: []GPU{
			{
				PCI:   "0000:0f:00.0",
				Model: "NVIDIA H100 80GB HBM3",
				ID:    "0",
			},
		},
		RdmaNics: []RdmaNic{
			{
				PCI:        "0000:0c:00.0",
				Interface:  "eth1",
				RdmaIP:     "192.168.1.100",
				DeviceName: "mlx5_0",
				Model:      "Mellanox Technologies MT2910 Family [ConnectX-7]",
				Numa:       "0",
				GpuID:      "0",
				GpuPCI:     "0000:0f:00.0",
			},
		},
		VcnNic: VcnNic{
			PrivateIP:  "10.0.0.100",
			PCI:        "0000:1f:00.0",
			Interface:  "eth0",
			DeviceName: "mlx5_2",
			Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
		},
	}

	// Test JSON formatting
	result, err := formatJSON(testMapHost)
	if err != nil {
		t.Fatalf("formatJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var unmarshaled MapHost
	if err := json.Unmarshal([]byte(result), &unmarshaled); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Verify key fields
	if unmarshaled.Hostname != "test-host" {
		t.Errorf("Expected hostname 'test-host', got '%s'", unmarshaled.Hostname)
	}
	if unmarshaled.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape 'BM.GPU.H100.8', got '%s'", unmarshaled.Shape)
	}
	if len(unmarshaled.Gpus) != 1 {
		t.Errorf("Expected 1 GPU, got %d", len(unmarshaled.Gpus))
	}
	if len(unmarshaled.RdmaNics) != 1 {
		t.Errorf("Expected 1 RDMA NIC, got %d", len(unmarshaled.RdmaNics))
	}
}

func TestFormatTable(t *testing.T) {
	// Create a test MapHost
	testMapHost := &MapHost{
		Hostname:         "test-host",
		Shape:            "BM.GPU.H100.8",
		Serial:           "TEST123",
		Rack:             "rack123",
		NetworkBlockId:   "netblock123",
		BuildingId:       "building123",
		InCluster:        true,
		Gpus: []GPU{
			{
				PCI:   "0000:0f:00.0",
				Model: "NVIDIA H100 80GB HBM3",
				ID:    "0",
			},
		},
		RdmaNics: []RdmaNic{
			{
				PCI:        "0000:0c:00.0",
				Interface:  "eth1",
				DeviceName: "mlx5_0",
				Model:      "Mellanox ConnectX-7",
			},
		},
		VcnNic: VcnNic{
			Interface:  "eth0",
			PCI:        "0000:1f:00.0",
			DeviceName: "mlx5_2",
			Model:      "Mellanox ConnectX-6",
		},
	}

	// Test table formatting
	result, err := formatTable(testMapHost)
	if err != nil {
		t.Fatalf("formatTable failed: %v", err)
	}

	// Verify table contains expected content
	expectedStrings := []string{
		"HARDWARE DISCOVERY RESULTS",
		"SYSTEM INFORMATION",
		"test-host",
		"BM.GPU.H100.8",
		"TEST123",
		"rack123",
		"netblock123",
		"building123",
		"GPU DEVICES",
		"RDMA NETWORK INTERFACES",
		"VCN NETWORK INTERFACE",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Table output missing expected string: %s", expected)
		}
	}
}

func TestFormatFriendly(t *testing.T) {
	// Create a test MapHost with multiple GPUs and RDMA NICs
	testMapHost := &MapHost{
		Hostname:         "test-host",
		Shape:            "BM.GPU.H100.8",
		Serial:           "TEST123",
		Rack:             "rack123",
		NetworkBlockId:   "netblock123",
		BuildingId:       "building123",
		InCluster:        true,
		Gpus: []GPU{
			{PCI: "0000:0f:00.0", Model: "NVIDIA H100", ID: "0"},
			{PCI: "0000:2d:00.0", Model: "NVIDIA H100", ID: "1"},
		},
		RdmaNics: []RdmaNic{
			{
				PCI:        "0000:0c:00.0",
				Interface:  "eth1",
				DeviceName: "mlx5_0",
				Model:      "Mellanox ConnectX-7",
				GpuID:      "0",
			},
			{
				PCI:        "0000:0c:00.1",
				Interface:  "eth2",
				DeviceName: "mlx5_1",
				Model:      "Mellanox ConnectX-7",
				GpuID:      "1",
			},
		},
		VcnNic: VcnNic{
			Interface: "eth0",
			PCI:       "0000:1f:00.0",
		},
	}

	// Test friendly formatting
	result, err := formatFriendly(testMapHost)
	if err != nil {
		t.Fatalf("formatFriendly failed: %v", err)
	}

	// Verify friendly format contains expected content
	expectedStrings := []string{
		"🔍 Hardware Discovery Results",
		"🖥️  System Information",
		"🎮 GPU Devices",
		"🌐 RDMA Network Interfaces",
		"🔗 VCN Network Interface",
		"📊 Discovery Summary",
		"Total GPUs detected: 2",
		"Total RDMA NICs detected: 2",
		"Total devices: 5", // 2 GPUs + 2 RDMA + 1 VCN
		"Hardware discovery completed successfully",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Friendly output missing expected string: %s", expected)
		}
	}
}

func TestFormatFriendlyWithMissingHardware(t *testing.T) {
	// Create a test MapHost with no GPUs or RDMA NICs
	testMapHost := &MapHost{
		Hostname:         "test-host",
		Shape:            "BM.Standard",
		InCluster:        false,
		Gpus:             []GPU{}, // No GPUs
		RdmaNics:         []RdmaNic{}, // No RDMA NICs
		VcnNic: VcnNic{
			Interface: "eth0",
			PCI:       "0000:1f:00.0",
		},
	}

	// Test friendly formatting
	result, err := formatFriendly(testMapHost)
	if err != nil {
		t.Fatalf("formatFriendly failed: %v", err)
	}

	// Verify it shows missing hardware warning
	expectedStrings := []string{
		"❌ No GPU devices detected",
		"❌ No RDMA NICs detected",
		"⚠️  Some hardware components may be missing",
		"Please verify your system configuration",
		"GPUs: 0",
		"RDMA NICs: 0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Friendly output missing expected string: %s", expected)
		}
	}
}

func TestMapHostStructure(t *testing.T) {
	// Test that MapHost can be marshaled and unmarshaled correctly
	original := MapHost{
		Hostname:         "test-host",
		Ocid:             "ocid1.test",
		FriendlyHostname: "test-host",
		Shape:            "BM.GPU.H100.8",
		Serial:           "TEST123",
		Rack:             "rack123",
		NetworkBlockId:   "netblock123",
		BuildingId:       "building123",
		InCluster:        true,
		Gpus: []GPU{
			{PCI: "0000:0f:00.0", Model: "NVIDIA H100", ID: "0"},
		},
		RdmaNics: []RdmaNic{
			{
				PCI:        "0000:0c:00.0",
				Interface:  "eth1",
				RdmaIP:     "192.168.1.100",
				DeviceName: "mlx5_0",
				Model:      "Mellanox ConnectX-7",
				Numa:       "0",
				GpuID:      "0",
				GpuPCI:     "0000:0f:00.0",
			},
		},
		VcnNic: VcnNic{
			PrivateIP:  "10.0.0.100",
			PCI:        "0000:1f:00.0",
			Interface:  "eth0",
			DeviceName: "mlx5_2",
			Model:      "Mellanox ConnectX-6",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal MapHost: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled MapHost
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MapHost: %v", err)
	}

	// Verify all fields
	if unmarshaled.Hostname != original.Hostname {
		t.Errorf("Hostname mismatch: expected %s, got %s", original.Hostname, unmarshaled.Hostname)
	}
	if unmarshaled.NetworkBlockId != original.NetworkBlockId {
		t.Errorf("NetworkBlockId mismatch: expected %s, got %s", original.NetworkBlockId, unmarshaled.NetworkBlockId)
	}
	if unmarshaled.BuildingId != original.BuildingId {
		t.Errorf("BuildingId mismatch: expected %s, got %s", original.BuildingId, unmarshaled.BuildingId)
	}
	if len(unmarshaled.Gpus) != len(original.Gpus) {
		t.Errorf("GPU count mismatch: expected %d, got %d", len(original.Gpus), len(unmarshaled.Gpus))
	}
	if len(unmarshaled.RdmaNics) != len(original.RdmaNics) {
		t.Errorf("RDMA NIC count mismatch: expected %d, got %d", len(original.RdmaNics), len(unmarshaled.RdmaNics))
	}
}

func TestRdmaNicStructure(t *testing.T) {
	rdmaNic := RdmaNic{
		PCI:        "0000:0c:00.0",
		Interface:  "eth1",
		RdmaIP:     "192.168.1.100",
		DeviceName: "mlx5_0",
		Model:      "Mellanox Technologies MT2910 Family [ConnectX-7]",
		Numa:       "0",
		GpuID:      "0",
		GpuPCI:     "0000:0f:00.0",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(rdmaNic)
	if err != nil {
		t.Fatalf("Failed to marshal RdmaNic: %v", err)
	}

	var unmarshaled RdmaNic
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal RdmaNic: %v", err)
	}

	// Verify all fields
	if unmarshaled.PCI != rdmaNic.PCI {
		t.Errorf("PCI mismatch: expected %s, got %s", rdmaNic.PCI, unmarshaled.PCI)
	}
	if unmarshaled.GpuID != rdmaNic.GpuID {
		t.Errorf("GpuID mismatch: expected %s, got %s", rdmaNic.GpuID, unmarshaled.GpuID)
	}
	if unmarshaled.RdmaIP != rdmaNic.RdmaIP {
		t.Errorf("RdmaIP mismatch: expected %s, got %s", rdmaNic.RdmaIP, unmarshaled.RdmaIP)
	}
}

func TestVcnNicStructure(t *testing.T) {
	vcnNic := VcnNic{
		PrivateIP:  "10.0.0.100",
		PCI:        "0000:1f:00.0",
		Interface:  "eth0",
		DeviceName: "mlx5_2",
		Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(vcnNic)
	if err != nil {
		t.Fatalf("Failed to marshal VcnNic: %v", err)
	}

	var unmarshaled VcnNic
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal VcnNic: %v", err)
	}

	// Verify all fields
	if unmarshaled.PrivateIP != vcnNic.PrivateIP {
		t.Errorf("PrivateIP mismatch: expected %s, got %s", vcnNic.PrivateIP, unmarshaled.PrivateIP)
	}
	if unmarshaled.Interface != vcnNic.Interface {
		t.Errorf("Interface mismatch: expected %s, got %s", vcnNic.Interface, unmarshaled.Interface)
	}
}

func TestGPUStructure(t *testing.T) {
	gpu := GPU{
		PCI:   "0000:0f:00.0",
		Model: "NVIDIA H100 80GB HBM3",
		ID:    "0",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(gpu)
	if err != nil {
		t.Fatalf("Failed to marshal GPU: %v", err)
	}

	var unmarshaled GPU
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal GPU: %v", err)
	}

	// Verify all fields
	if unmarshaled.PCI != gpu.PCI {
		t.Errorf("PCI mismatch: expected %s, got %s", gpu.PCI, unmarshaled.PCI)
	}
	if unmarshaled.Model != gpu.Model {
		t.Errorf("Model mismatch: expected %s, got %s", gpu.Model, unmarshaled.Model)
	}
	if unmarshaled.ID != gpu.ID {
		t.Errorf("ID mismatch: expected %s, got %s", gpu.ID, unmarshaled.ID)
	}
}

func TestInClusterLogic(t *testing.T) {
	testCases := []struct {
		name               string
		networkBlockId     string
		expectedInCluster  bool
		description        string
	}{
		{
			name:               "Valid network block ID",
			networkBlockId:     "9877ad75a930eec924e1bb971688e0c338ab613cba5afc9793b5b4754902477d",
			expectedInCluster:  true,
			description:        "Should be in cluster with valid network block ID",
		},
		{
			name:               "Empty network block ID",
			networkBlockId:     "",
			expectedInCluster:  false,
			description:        "Should not be in cluster with empty network block ID",
		},
		{
			name:               "Unknown network block ID",
			networkBlockId:     "unknown",
			expectedInCluster:  false,
			description:        "Should not be in cluster with unknown network block ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the logic from the autodiscover function
			inCluster := tc.networkBlockId != "" && tc.networkBlockId != "unknown"
			
			if inCluster != tc.expectedInCluster {
				t.Errorf("Expected InCluster=%v, got %v (%s)", tc.expectedInCluster, inCluster, tc.description)
			}
		})
	}
}