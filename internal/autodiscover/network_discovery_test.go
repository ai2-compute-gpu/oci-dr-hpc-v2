package autodiscover

import (
	"testing"
)

func TestDiscoverRDMANicsWithFallback(t *testing.T) {
	// Test with a known shape that should have RDMA NICs
	t.Run("Known shape BM.GPU.H100.8", func(t *testing.T) {
		rdmaNics := DiscoverRDMANicsWithFallback("BM.GPU.H100.8")

		if len(rdmaNics) == 0 {
			t.Error("Expected RDMA NICs for BM.GPU.H100.8 but got none")
		}

		// Check first RDMA NIC has expected properties
		if len(rdmaNics) > 0 {
			rdmaNic := rdmaNics[0]
			if rdmaNic.PCI == "" {
				t.Error("Expected PCI address but got empty string")
			}
			if rdmaNic.DeviceName == "" {
				t.Error("Expected device name but got empty string")
			}
			if rdmaNic.Model == "" {
				t.Error("Expected model but got empty string")
			}

			t.Logf("Found %d RDMA NICs, first one: PCI=%s, Device=%s, Model=%s, GPU_ID=%s",
				len(rdmaNics), rdmaNic.PCI, rdmaNic.DeviceName, rdmaNic.Model, rdmaNic.GpuID)
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

		if vcnNic.PCI == "" {
			t.Error("Expected PCI address but got empty string")
		}
		if vcnNic.DeviceName == "" {
			t.Error("Expected device name but got empty string")
		}
		if vcnNic.Model == "" {
			t.Error("Expected model but got empty string")
		}

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

	// Should find multiple RDMA NICs for H100
	if len(rdmaNics) == 0 {
		t.Error("Expected RDMA NICs for H100 shape")
	}

	// Should find VCN NIC
	if vcnNic.PCI == "undefined" {
		t.Error("Expected actual VCN NIC for H100 shape")
	}

	// Verify RDMA NICs have GPU associations
	foundGPUAssociation := false
	for _, rdmaNic := range rdmaNics {
		if rdmaNic.GpuID != "" && rdmaNic.GpuID != "undefined" {
			foundGPUAssociation = true
			break
		}
	}

	if !foundGPUAssociation {
		t.Error("Expected at least one RDMA NIC to have GPU association")
	}

	t.Logf("Successfully discovered %d RDMA NICs and 1 VCN NIC for shape %s",
		len(rdmaNics), shapeName)
}
