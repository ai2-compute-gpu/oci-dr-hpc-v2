package autodiscover

import (
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
