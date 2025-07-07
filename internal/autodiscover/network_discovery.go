package autodiscover

import (
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/config"
	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
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

	// Convert shapes.RDMANic to autodiscover.RdmaNic with OS discovery
	var result []RdmaNic
	for _, rdmaNic := range rdmaNics {
		logger.Infof("Processing RDMA NIC with PCI: %s", rdmaNic.PCI)
		
		// Get static values from shapes.json
		convertedNic := RdmaNic{
			PCI:     rdmaNic.PCI,
			GpuID:   rdmaNic.GPUID.String(),
			GpuPCI:  rdmaNic.GPUPCI,
		}

		// Discover runtime values from OS
		// Get device name from InfiniBand subsystem
		if deviceName, err := executor.GetInfiniBandDeviceName(rdmaNic.PCI); err == nil && deviceName != "" {
			convertedNic.DeviceName = deviceName
		} else {
			// Fallback to shapes.json value
			convertedNic.DeviceName = rdmaNic.DeviceName
			if convertedNic.DeviceName == "" {
				convertedNic.DeviceName = "undefined"
			}
		}

		// Get device model from lspci
		if model, err := executor.GetPCIDeviceModel(rdmaNic.PCI); err == nil && model != "" {
			convertedNic.Model = model
		} else {
			// Fallback to shapes.json value
			convertedNic.Model = rdmaNic.Model
			if convertedNic.Model == "" {
				convertedNic.Model = "undefined"
			}
		}

		// Get network interface name
		if interfaceName, err := executor.GetNetworkInterfaceName(rdmaNic.PCI); err == nil && interfaceName != "" {
			convertedNic.Interface = interfaceName
		} else {
			// Fallback to shapes.json value
			convertedNic.Interface = rdmaNic.Interface
			if convertedNic.Interface == "" {
				convertedNic.Interface = "undefined"
			}
		}

		// Get NUMA node
		if numaNode, err := executor.GetPCIDeviceNUMANode(rdmaNic.PCI); err == nil && numaNode != "" && numaNode != "unknown" {
			convertedNic.Numa = numaNode
		} else {
			convertedNic.Numa = "undefined"
		}

		// Get RDMA IP address
		if convertedNic.DeviceName != "undefined" && convertedNic.DeviceName != "" {
			if rdmaIP, err := executor.GetRDMADeviceIP(convertedNic.DeviceName); err == nil && rdmaIP != "" {
				convertedNic.RdmaIP = rdmaIP
			} else {
				logger.Debugf("Could not get RDMA IP for device %s: %v", convertedNic.DeviceName, err)
				convertedNic.RdmaIP = "undefined"
			}
		} else {
			convertedNic.RdmaIP = "undefined"
		}

		logger.Infof("Discovered RDMA NIC: PCI=%s, Device=%s, Interface=%s, Model=%s, NUMA=%s, IP=%s", 
			convertedNic.PCI, convertedNic.DeviceName, convertedNic.Interface, 
			convertedNic.Model, convertedNic.Numa, convertedNic.RdmaIP)

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

	// Convert shapes.VCNNic to autodiscover.VcnNic with OS discovery
	logger.Infof("Processing VCN NIC with PCI: %s", vcnNic.PCI)
	
	// Get static values from shapes.json
	result := VcnNic{
		PCI: vcnNic.PCI,
	}

	// Discover runtime values from OS
	// Get device name from InfiniBand subsystem (try first)
	if deviceName, err := executor.GetInfiniBandDeviceName(vcnNic.PCI); err == nil && deviceName != "" {
		result.DeviceName = deviceName
	} else {
		// Fallback to shapes.json value
		result.DeviceName = vcnNic.DeviceName
		if result.DeviceName == "" {
			result.DeviceName = "undefined"
		}
	}

	// Get device model from lspci
	if model, err := executor.GetPCIDeviceModel(vcnNic.PCI); err == nil && model != "" {
		result.Model = model
	} else {
		// Fallback to shapes.json value
		result.Model = vcnNic.Model
		if result.Model == "" {
			result.Model = "undefined"
		}
	}

	// Get network interface name
	if interfaceName, err := executor.GetNetworkInterfaceName(vcnNic.PCI); err == nil && interfaceName != "" {
		result.Interface = interfaceName
	} else {
		// Fallback to shapes.json value
		result.Interface = vcnNic.Interface
		if result.Interface == "" {
			result.Interface = "undefined"
		}
	}

	// Get private IP address
	if result.Interface != "undefined" && result.Interface != "" {
		if privateIP, err := getInterfaceIP(result.Interface); err == nil && privateIP != "" {
			result.PrivateIP = privateIP
		} else {
			logger.Debugf("Could not get private IP for interface %s: %v", result.Interface, err)
			result.PrivateIP = "undefined"
		}
	} else {
		result.PrivateIP = "undefined"
	}

	logger.Infof("Discovered VCN NIC: PCI=%s, Device=%s, Interface=%s, Model=%s, IP=%s", 
		result.PCI, result.DeviceName, result.Interface, result.Model, result.PrivateIP)

	return result, nil
}

// getInterfaceIP returns the IP address of a network interface
func getInterfaceIP(interfaceName string) (string, error) {
	logger.Infof("Getting IP address for interface: %s", interfaceName)

	result, err := executor.RunIPAddr("show", interfaceName)
	if err != nil {
		logger.Errorf("Failed to get IP for interface %s: %v", interfaceName, err)
		return "", fmt.Errorf("failed to get IP for interface %s: %w", interfaceName, err)
	}

	// Parse ip addr output to find IPv4 address
	// Look for "inet 192.168.1.1/24"
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Extract just the IP address (remove /24 suffix)
				ipWithCIDR := parts[1]
				if strings.Contains(ipWithCIDR, "/") {
					ip := strings.Split(ipWithCIDR, "/")[0]
					logger.Infof("Found IP address: %s", ip)
					return ip, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find IP address for interface %s", interfaceName)
}
