package autodiscover

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// NetworkInterface represents a discovered network interface
type NetworkInterface struct {
	Interface  string `json:"interface"`
	PrivateIP  string `json:"private_ip"`
	PCI        string `json:"pci"`
	DeviceName string `json:"device_name"`
	Model      string `json:"model"`
	MTU        int    `json:"mtu"`
}

// DiscoverVCNInterface discovers VCN network interface using the 4-step process:
// 1. Use ip a to find network card with MTU 9000, get IP and interface
// 2. Use rdma link to get the device_name (mlx5_X)
// 3. Use readlink to get PCI address from /sys/class/infiniband/device_name/device
// 4. Use lspci to get the model
func DiscoverVCNInterface() (*NetworkInterface, error) {
	logger.Info("Discovering VCN network interface...")

	// Step 1: Find interface with MTU 9000 using ip addr
	interfaceInfo, err := findInterfaceWithMTU9000()
	if err != nil {
		return nil, fmt.Errorf("failed to find interface with MTU 9000: %w", err)
	}

	logger.Infof("Found interface %s with MTU 9000 and IP %s", interfaceInfo.Interface, interfaceInfo.PrivateIP)

	// Step 2: Get device name using rdma link
	deviceName, err := getDeviceNameFromRdmaLink(interfaceInfo.Interface)
	if err != nil {
		logger.Errorf("Failed to get device name from rdma link: %v", err)

		// Fallback: Generate device name from interface name
		// eth0 -> mlx5_0, eth1 -> mlx5_1, ens3 -> mlx5_3, etc.
		interfaceNumber := "0" // Default fallback

		// Try to extract number from interface name
		if strings.HasPrefix(interfaceInfo.Interface, "eth") {
			interfaceNumber = strings.TrimPrefix(interfaceInfo.Interface, "eth")
		} else if strings.HasPrefix(interfaceInfo.Interface, "ens") {
			interfaceNumber = strings.TrimPrefix(interfaceInfo.Interface, "ens")
		} else {
			// For other interface names, try to extract trailing number
			re := regexp.MustCompile(`\d+$`)
			matches := re.FindStringSubmatch(interfaceInfo.Interface)
			if len(matches) > 0 {
				interfaceNumber = matches[0]
			}
		}

		deviceName = fmt.Sprintf("mlx5_%s", interfaceNumber)
		logger.Infof("Using fallback device name: %s (from interface %s)", deviceName, interfaceInfo.Interface)
	} else {
		logger.Infof("Successfully got device name from rdma link: %s", deviceName)
	}

	interfaceInfo.DeviceName = deviceName
	logger.Infof("Set DeviceName to: %s", interfaceInfo.DeviceName)

	// Step 3: Get PCI address using readlink
	pciAddress, err := getPCIAddressFromDevice(deviceName)
	if err != nil {
		logger.Errorf("Failed to get PCI address for device %s: %v", deviceName, err)
		// Fallback: use a placeholder
		interfaceInfo.PCI = "unknown"
	} else {
		interfaceInfo.PCI = pciAddress
	}

	// Step 4: Get model using lspci
	if interfaceInfo.PCI != "unknown" {
		model, err := getModelFromLspci(interfaceInfo.PCI)
		if err != nil {
			logger.Errorf("Failed to get model from lspci: %v", err)
			interfaceInfo.Model = "Unknown Mellanox Device"
		} else {
			interfaceInfo.Model = model
		}
	} else {
		interfaceInfo.Model = "Unknown Mellanox Device"
	}

	logger.Infof("Successfully discovered VCN interface: %s (%s)", interfaceInfo.Interface, interfaceInfo.Model)
	return interfaceInfo, nil
}

// Step 1: Find interface with MTU 9000 and get IP address
func findInterfaceWithMTU9000() (*NetworkInterface, error) {
	logger.Info("Step 1: Finding interface with MTU 9000...")

	result, err := executor.RunIPAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to run ip addr: %w", err)
	}

	return parseIPAddrOutput(result.Output)
}

// parseIPAddrOutput parses ip addr output to find interface with MTU 9000
func parseIPAddrOutput(output string) (*NetworkInterface, error) {
	lines := strings.Split(output, "\n")

	var currentInterface *NetworkInterface

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Match interface definition line: "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 ..."
		if matched, _ := regexp.MatchString(`^\d+:\s+\w+:.*mtu\s+9000`, line); matched {
			// Extract interface name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				interfaceName := strings.TrimSuffix(parts[1], ":")

				// Extract MTU
				mtu := 9000
				for i, part := range parts {
					if part == "mtu" && i+1 < len(parts) {
						if parsedMTU, err := strconv.Atoi(parts[i+1]); err == nil {
							mtu = parsedMTU
						}
						break
					}
				}

				currentInterface = &NetworkInterface{
					Interface: interfaceName,
					MTU:       mtu,
				}

				logger.Infof("Found interface %s with MTU %d", interfaceName, mtu)
			}
		}

		// Match IP address line: "inet 10.0.11.179/24 ..."
		if currentInterface != nil && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			// Extract IP address
			re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentInterface.PrivateIP = matches[1]
				logger.Infof("Found IP address %s for interface %s", currentInterface.PrivateIP, currentInterface.Interface)
				return currentInterface, nil
			}
		}
	}

	return nil, fmt.Errorf("no interface with MTU 9000 found")
}

// Step 2: Get device name using rdma link
func getDeviceNameFromRdmaLink(interfaceName string) (string, error) {
	logger.Infof("Step 2: Getting device name for interface %s using rdma link...", interfaceName)

	result, err := executor.RunRdmaLink()
	if err != nil {
		return "", fmt.Errorf("failed to run rdma link: %w", err)
	}

	return parseRdmaLinkOutput(result.Output, interfaceName)
}

// parseRdmaLinkOutput parses rdma link output to find device name
func parseRdmaLinkOutput(output, interfaceName string) (string, error) {
	lines := strings.Split(output, "\n")

	logger.Infof("Parsing rdma link output for interface %s:", interfaceName)
	logger.Infof("RDMA output:\n%s", output)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		logger.Infof("Checking rdma line: %s", line)

		// Parse rdma link output format: "link mlx5_0/1 state ACTIVE physical_state LINK_UP netdev eth0"
		// Or alternative format: "1/1: mlx5_0/1: state ACTIVE physical_state LINK_UP netdev eth0"
		if strings.Contains(line, interfaceName) {
			logger.Infof("Found matching line for interface %s: %s", interfaceName, line)

			fields := strings.Fields(line)
			for i, field := range fields {
				// Look for "link" followed by device name, or device name pattern directly
				if field == "link" && i+1 < len(fields) {
					linkInfo := fields[i+1]
					// Extract device name (mlx5_0 from mlx5_0/1)
					deviceName := strings.Split(linkInfo, "/")[0]
					logger.Infof("Found device name %s for interface %s (format: link mlx5_X/Y)", deviceName, interfaceName)
					return deviceName, nil
				}

				// Alternative: look for mlx5_X/Y pattern directly
				if strings.Contains(field, "mlx5_") && strings.Contains(field, "/") {
					deviceName := strings.Split(field, "/")[0]
					logger.Infof("Found device name %s for interface %s (format: mlx5_X/Y)", deviceName, interfaceName)
					return deviceName, nil
				}
			}
		}
	}

	logger.Errorf("Device name not found for interface %s in rdma link output", interfaceName)
	return "", fmt.Errorf("device name not found for interface %s in rdma link output", interfaceName)
}

// Step 3: Get PCI address using readlink
func getPCIAddressFromDevice(deviceName string) (string, error) {
	logger.Infof("Step 3: Getting PCI address for device %s...", deviceName)

	syspath := fmt.Sprintf("/sys/class/infiniband/%s/device", deviceName)
	result, err := executor.RunReadlink(syspath, "-f")
	if err != nil {
		return "", fmt.Errorf("failed to run readlink for %s: %w", syspath, err)
	}

	return parsePCIAddressFromPath(result.Output)
}

// parsePCIAddressFromPath extracts PCI address from device path
func parsePCIAddressFromPath(devicePath string) (string, error) {
	if devicePath == "" {
		return "", fmt.Errorf("empty device path")
	}

	// Extract PCI address from path like /sys/devices/pci0000:00/0000:00:1f.0
	// We want the last component that looks like a PCI address
	pathParts := strings.Split(devicePath, "/")
	for i := len(pathParts) - 1; i >= 0; i-- {
		part := pathParts[i]
		// Match PCI address pattern: 0000:00:1f.0 (domain:bus:device.function)
		if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]+$`, part); matched {
			logger.Infof("Extracted PCI address: %s", part)
			return part, nil
		}
	}

	return "", fmt.Errorf("no PCI address found in path: %s", devicePath)
}

// Step 4: Get model using lspci
func getModelFromLspci(pciAddress string) (string, error) {
	logger.Infof("Step 4: Getting model for PCI address %s...", pciAddress)

	result, err := executor.RunLspciByPCI(pciAddress, false)
	if err != nil {
		return "", fmt.Errorf("failed to run lspci for PCI %s: %w", pciAddress, err)
	}

	return parseModelFromLspci(result.Output)
}

// parseModelFromLspci extracts model information from lspci output
func parseModelFromLspci(output string) (string, error) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse lspci output: "1f:00.0 Ethernet controller: Mellanox Technologies MT2892 Family [ConnectX-6 Dx]"
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) >= 2 {
			// Skip the device type and extract the model
			modelParts := strings.SplitN(parts[1], ": ", 2)
			if len(modelParts) >= 2 {
				model := strings.TrimSpace(modelParts[1])
				logger.Infof("Extracted model: %s", model)
				return model, nil
			} else {
				// Fallback: use the entire description after the first colon
				model := strings.TrimSpace(parts[1])
				logger.Infof("Extracted model (fallback): %s", model)
				return model, nil
			}
		}
	}

	return "", fmt.Errorf("no model information found in lspci output: %s", output)
}

// DiscoverVCNInterfaceWithFallback attempts to discover VCN interface with graceful fallback
func DiscoverVCNInterfaceWithFallback() *NetworkInterface {
	logger.Info("Discovering VCN interface with fallback...")

	netInterface, err := DiscoverVCNInterface()
	if err != nil {
		logger.Errorf("VCN interface discovery failed: %v", err)
		logger.Info("Using undefined VCN interface data - discovery failed")

		return &NetworkInterface{
			PrivateIP:  "undefined",
			PCI:        "undefined",
			Interface:  "undefined",
			DeviceName: "undefined",
			Model:      "undefined",
			MTU:        0,
		}
	}

	logger.Info("Successfully discovered VCN interface")
	return netInterface
}
