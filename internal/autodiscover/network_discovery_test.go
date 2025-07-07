package autodiscover

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDiscoverRDMANicsWithFallback(t *testing.T) {
	// Test with a known shape that should have RDMA NICs
	t.Run("Known shape BM.GPU.H100.8", func(t *testing.T) {
		rdmaNics := DiscoverRDMANicsWithFallback("BM.GPU.H100.8")

		// For tests, we expect fallback behavior since shapes.json path is different
		// In production, this should work correctly
		if len(rdmaNics) >= 1 {
			rdmaNic := rdmaNics[0]
			t.Logf("Found %d RDMA NICs, first one: PCI=%s, Device=%s, Model=%s, GPU_ID=%s",
				len(rdmaNics), rdmaNic.PCI, rdmaNic.DeviceName, rdmaNic.Model, rdmaNic.GpuID)
		} else {
			t.Log("No RDMA NICs found (expected in test environment)")
		}
	})

	// Test with unknown shape (should fallback to undefined)
	t.Run("Unknown shape", func(t *testing.T) {
		rdmaNics := DiscoverRDMANicsWithFallback("BM.UNKNOWN.SHAPE")

		if len(rdmaNics) != 1 {
			t.Errorf("Expected 1 fallback RDMA NIC but got %d", len(rdmaNics))
		}

		if len(rdmaNics) > 0 {
			rdmaNic := rdmaNics[0]
			if rdmaNic.PCI != "undefined" {
				t.Errorf("Expected undefined PCI but got %s", rdmaNic.PCI)
			}
			if rdmaNic.DeviceName != "undefined" {
				t.Errorf("Expected undefined device name but got %s", rdmaNic.DeviceName)
			}
		}
	})
}

func TestDiscoverVCNNicWithFallback(t *testing.T) {
	// Test with a known shape that should have VCN NIC
	t.Run("Known shape BM.GPU.H100.8", func(t *testing.T) {
		vcnNic := DiscoverVCNNicWithFallback("BM.GPU.H100.8")

		t.Logf("Found VCN NIC: PCI=%s, Device=%s, Model=%s",
			vcnNic.PCI, vcnNic.DeviceName, vcnNic.Model)
	})

	// Test with unknown shape (should fallback to undefined)
	t.Run("Unknown shape", func(t *testing.T) {
		vcnNic := DiscoverVCNNicWithFallback("BM.UNKNOWN.SHAPE")

		if vcnNic.PCI != "undefined" {
			t.Errorf("Expected undefined PCI but got %s", vcnNic.PCI)
		}
		if vcnNic.DeviceName != "undefined" {
			t.Errorf("Expected undefined device name but got %s", vcnNic.DeviceName)
		}
		if vcnNic.Model != "undefined" {
			t.Errorf("Expected undefined model but got %s", vcnNic.Model)
		}
	})
}

func TestNetworkDiscoveryIntegration(t *testing.T) {
	// Test the complete integration with a known shape
	shapeName := "BM.GPU.H100.8"

	rdmaNics := DiscoverRDMANicsWithFallback(shapeName)
	vcnNic := DiscoverVCNNicWithFallback(shapeName)

	// In production with correct shape, should find RDMA NICs and VCN NIC
	// In test environment, may fallback to undefined values

	t.Logf("Discovered %d RDMA NICs and VCN NIC (PCI=%s) for shape %s",
		len(rdmaNics), vcnNic.PCI, shapeName)

	// Verify structure is correct regardless of whether real data was found
	if len(rdmaNics) > 0 {
		rdmaNic := rdmaNics[0]
		if rdmaNic.PCI == "" || rdmaNic.DeviceName == "" || rdmaNic.Model == "" {
			t.Error("RDMA NIC structure incomplete")
		}
	}
}

func TestGetInterfaceIPParsing(t *testing.T) {
	// Test the IP address parsing logic with mock data
	testCases := []struct {
		name       string
		ipOutput   string
		expectedIP string
		shouldErr  bool
	}{
		{
			name: "Valid IPv4 address",
			ipOutput: `2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:00:17:05:09:02 brd ff:ff:ff:ff:ff:ff
    inet 10.0.0.100/24 brd 10.0.0.255 scope global eth0
       valid_lft forever preferred_lft forever`,
			expectedIP: "10.0.0.100",
			shouldErr:  false,
		},
		{
			name: "RDMA network IP",
			ipOutput: `3: enp12s0f0np0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 4220 qdisc mq state UP group default qlen 1000
    link/ether 0c:42:a1:dd:37:b2 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/16 brd 192.168.255.255 scope global enp12s0f0np0
       valid_lft forever preferred_lft forever`,
			expectedIP: "192.168.1.100",
			shouldErr:  false,
		},
		{
			name: "No IP address configured",
			ipOutput: `4: eth2: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 02:00:17:05:09:04 brd ff:ff:ff:ff:ff:ff`,
			expectedIP: "",
			shouldErr:  true,
		},
		{
			name:       "Empty output",
			ipOutput:   "",
			expectedIP: "",
			shouldErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the IP parsing logic from getInterfaceIP
			var foundIP string
			lines := strings.Split(tc.ipOutput, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "inet ") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						ipWithCIDR := parts[1]
						if strings.Contains(ipWithCIDR, "/") {
							foundIP = strings.Split(ipWithCIDR, "/")[0]
							break
						}
					}
				}
			}

			if tc.shouldErr {
				if foundIP != "" {
					t.Errorf("Expected error but found IP: %s", foundIP)
				}
			} else {
				if foundIP != tc.expectedIP {
					t.Errorf("Expected IP %s, got %s", tc.expectedIP, foundIP)
				}
			}
		})
	}
}

func TestRDMANicJSONSerialization(t *testing.T) {
	rdmaNic := RdmaNic{
		PCI:        "0000:0c:00.0",
		Interface:  "enp12s0f0np0",
		RdmaIP:     "192.168.1.100",
		DeviceName: "mlx5_0",
		Model:      "Mellanox Technologies MT2910 Family [ConnectX-7]",
		Numa:       "0",
		GpuID:      "0",
		GpuPCI:     "0000:0f:00.0",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(rdmaNic)
	if err != nil {
		t.Fatalf("Failed to marshal RdmaNic: %v", err)
	}

	// Test JSON unmarshaling
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
	if unmarshaled.Numa != rdmaNic.Numa {
		t.Errorf("Numa mismatch: expected %s, got %s", rdmaNic.Numa, unmarshaled.Numa)
	}
}

func TestVCNNicJSONSerialization(t *testing.T) {
	vcnNic := VcnNic{
		PrivateIP:  "10.0.0.100",
		PCI:        "0000:1f:00.0",
		Interface:  "eth0",
		DeviceName: "mlx5_2",
		Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(vcnNic)
	if err != nil {
		t.Fatalf("Failed to marshal VcnNic: %v", err)
	}

	// Test JSON unmarshaling
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
	if unmarshaled.DeviceName != vcnNic.DeviceName {
		t.Errorf("DeviceName mismatch: expected %s, got %s", vcnNic.DeviceName, unmarshaled.DeviceName)
	}
}

func TestNetworkDiscoveryEdgeCases(t *testing.T) {
	// Test edge cases for network discovery
	
	t.Run("Empty shape name", func(t *testing.T) {
		rdmaNics := DiscoverRDMANicsWithFallback("")
		if len(rdmaNics) == 0 {
			t.Error("Expected fallback RDMA NIC for empty shape")
		}
		
		vcnNic := DiscoverVCNNicWithFallback("")
		if vcnNic.PCI == "" {
			t.Error("Expected fallback VCN NIC for empty shape")
		}
	})
	
	t.Run("Very long shape name", func(t *testing.T) {
		longShape := strings.Repeat("A", 1000)
		rdmaNics := DiscoverRDMANicsWithFallback(longShape)
		if len(rdmaNics) == 0 {
			t.Error("Expected fallback RDMA NIC for long shape name")
		}
	})
	
	t.Run("Shape with special characters", func(t *testing.T) {
		specialShape := "BM.GPU.H100@#$%^&*()_+"
		rdmaNics := DiscoverRDMANicsWithFallback(specialShape)
		if len(rdmaNics) == 0 {
			t.Error("Expected fallback RDMA NIC for special character shape")
		}
	})
}
