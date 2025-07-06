package autodiscover

import (
	"fmt"

	"github.com/oracle/oci-dr-hpc-v2/internal/config"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/shapes"
)

// DiscoverRDMANicsWithFallback discovers RDMA NICs using shape information
func DiscoverRDMANicsWithFallback(shapeName string) []RdmaNic {
	logger.Infof("Discovering RDMA NICs for shape: %s", shapeName)

	// Try to get RDMA NICs from shapes configuration
	rdmaNics, err := discoverRDMANicsFromShape(shapeName)
	if err != nil {
		logger.Errorf("Failed to discover RDMA NICs from shape: %v", err)
		logger.Infof("Falling back to undefined RDMA NIC values")
		return []RdmaNic{
			{
				PCI:        "undefined",
				Interface:  "undefined",
				RdmaIP:     "undefined",
				DeviceName: "undefined",
				Model:      "undefined",
				Numa:       "undefined",
				GpuID:      "undefined",
				GpuPCI:     "undefined",
			},
		}
	}

	logger.Infof("Successfully discovered %d RDMA NICs from shape configuration", len(rdmaNics))
	for i, nic := range rdmaNics {
		logger.Debugf("RDMA NIC %d: PCI=%s, Device=%s, GPU_ID=%s", i, nic.PCI, nic.DeviceName, nic.GpuID)
	}
	return rdmaNics
}

// DiscoverVCNNicWithFallback discovers VCN NIC using shape information
func DiscoverVCNNicWithFallback(shapeName string) VcnNic {
	logger.Infof("Discovering VCN NIC for shape: %s", shapeName)

	// Try to get VCN NIC from shapes configuration
	vcnNic, err := discoverVCNNicFromShape(shapeName)
	if err != nil {
		logger.Errorf("Failed to discover VCN NIC from shape: %v", err)
		logger.Infof("Falling back to undefined VCN NIC values")
		return VcnNic{
			PrivateIP:  "undefined",
			PCI:        "undefined",
			Interface:  "undefined",
			DeviceName: "undefined",
			Model:      "undefined",
		}
	}

	logger.Infof("Successfully discovered VCN NIC from shape configuration")
	logger.Debugf("VCN NIC: PCI=%s, Device=%s, Model=%s", vcnNic.PCI, vcnNic.DeviceName, vcnNic.Model)
	return vcnNic
}

// discoverRDMANicsFromShape discovers RDMA NICs from the shapes.json configuration
func discoverRDMANicsFromShape(shapeName string) ([]RdmaNic, error) {
	// Get the shapes configuration file path
	shapesFile := config.GetShapesFilePath()

	// Create shape manager
	shapeManager, err := shapes.NewShapeManager(shapesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create shape manager: %w", err)
	}

	// Get RDMA NICs for the shape
	rdmaNics, err := shapeManager.GetRDMANics(shapeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get RDMA NICs for shape %s: %w", shapeName, err)
	}

	// Convert shapes.RDMANic to autodiscover.RdmaNic
	var result []RdmaNic
	for _, rdmaNic := range rdmaNics {
		convertedNic := RdmaNic{
			PCI:        rdmaNic.PCI,
			Interface:  rdmaNic.Interface,
			RdmaIP:     "undefined", // IP addresses are not in shapes.json, would need runtime discovery
			DeviceName: rdmaNic.DeviceName,
			Model:      rdmaNic.Model,
			Numa:       "undefined", // NUMA info not in shapes.json, would need runtime discovery
			GpuID:      rdmaNic.GPUID.String(),
			GpuPCI:     rdmaNic.GPUPCI,
		}

		// If interface is empty, set to undefined
		if convertedNic.Interface == "" {
			convertedNic.Interface = "undefined"
		}

		result = append(result, convertedNic)
	}

	return result, nil
}

// discoverVCNNicFromShape discovers the first VCN NIC from the shapes.json configuration
func discoverVCNNicFromShape(shapeName string) (VcnNic, error) {
	// Get the shapes configuration file path
	shapesFile := config.GetShapesFilePath()

	// Create shape manager
	shapeManager, err := shapes.NewShapeManager(shapesFile)
	if err != nil {
		return VcnNic{}, fmt.Errorf("failed to create shape manager: %w", err)
	}

	// Get VCN NICs for the shape
	vcnNics, err := shapeManager.GetVCNNics(shapeName)
	if err != nil {
		return VcnNic{}, fmt.Errorf("failed to get VCN NICs for shape %s: %w", shapeName, err)
	}

	if len(vcnNics) == 0 {
		return VcnNic{}, fmt.Errorf("no VCN NICs found for shape %s", shapeName)
	}

	// Use the first VCN NIC
	vcnNic := vcnNics[0]

	// Convert shapes.VCNNic to autodiscover.VcnNic
	result := VcnNic{
		PrivateIP:  "undefined", // IP addresses are not in shapes.json, would need runtime discovery
		PCI:        vcnNic.PCI,
		Interface:  vcnNic.Interface,
		DeviceName: vcnNic.DeviceName,
		Model:      vcnNic.Model,
	}

	// If interface is empty, set to undefined
	if result.Interface == "" {
		result.Interface = "undefined"
	}

	return result, nil
}
