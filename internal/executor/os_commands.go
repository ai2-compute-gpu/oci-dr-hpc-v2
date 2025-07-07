package executor

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// OSCommandResult represents the result of executing an OS command
type OSCommandResult struct {
	Command  string
	Output   string
	Error    error
	ExitCode int
}

// RunLspci executes lspci command with specified options
func RunLspci(options ...string) (*OSCommandResult, error) {
	logger.Info("Running lspci command...")

	// Build command arguments - prepend lspci to sudo args
	args := append([]string{"lspci"}, options...)

	cmd := exec.Command("sudo", args...)
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: fmt.Sprintf("sudo lspci %s", strings.Join(options, " ")),
		Output:  string(output),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("lspci command failed: %v", err)
		logger.Debugf("lspci output: %s", result.Output)
		return result, err
	}

	logger.Info("lspci command completed successfully")
	logger.Debugf("lspci output: %s", result.Output)

	return result, nil
}

// RunLspciForDevice executes lspci for a specific device
func RunLspciForDevice(deviceID string, verbose bool) (*OSCommandResult, error) {
	logger.Infof("Running lspci for device: %s", deviceID)

	args := []string{"-s", deviceID}
	if verbose {
		args = append(args, "-v")
	}

	return RunLspci(args...)
}

// RunDmesg executes dmesg command with specified options
func RunDmesg(options ...string) (*OSCommandResult, error) {
	logger.Info("Running dmesg command...")

	// Build command arguments - prepend dmesg to sudo args
	args := append([]string{"dmesg"}, options...)

	cmd := exec.Command("sudo", args...)
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: fmt.Sprintf("sudo dmesg %s", strings.Join(options, " ")),
		Output:  string(output),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("dmesg command failed: %v", err)
		logger.Debugf("dmesg output: %s", result.Output)
		return result, err
	}

	logger.Info("dmesg command completed successfully")
	logger.Debugf("dmesg output length: %d characters", len(result.Output))

	return result, nil
}

// GetHostname retrieves the system hostname using os.Hostname()
func GetHostname() (string, error) {
	logger.Info("Getting system hostname...")

	hostname, err := os.Hostname()
	if err != nil {
		logger.Errorf("Failed to get hostname: %v", err)
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	logger.Infof("Successfully retrieved hostname: %s", hostname)
	return hostname, nil
}

// GetSerialNumber retrieves the chassis serial number using dmidecode
func GetSerialNumber() (*OSCommandResult, error) {
	logger.Info("Running dmidecode to get chassis serial number...")

	cmd := exec.Command("sudo", "dmidecode", "-s", "chassis-serial-number")
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: "sudo dmidecode -s chassis-serial-number",
		Output:  strings.TrimSpace(string(output)),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("dmidecode command failed: %v", err)
		logger.Debugf("dmidecode output: %s", result.Output)
		return result, err
	}

	logger.Infof("Successfully retrieved chassis serial number: %s", result.Output)
	return result, nil
}

// RunIPAddr executes ip addr command to get network interface information
func RunIPAddr(options ...string) (*OSCommandResult, error) {
	logger.Info("Running ip addr command...")

	// Build command arguments
	args := append([]string{"addr"}, options...)

	cmd := exec.Command("ip", args...)
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: fmt.Sprintf("ip addr %s", strings.Join(options, " ")),
		Output:  string(output),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("ip addr command failed: %v", err)
		logger.Debugf("ip addr output: %s", result.Output)
		return result, err
	}

	logger.Info("ip addr command completed successfully")
	logger.Debugf("ip addr output length: %d characters", len(result.Output))

	return result, nil
}

// RunRdmaLink executes rdma link command to get RDMA device information
func RunRdmaLink(options ...string) (*OSCommandResult, error) {
	logger.Info("Running rdma link command...")

	cmd := exec.Command("rdma", append([]string{"link"}, options...)...)
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: fmt.Sprintf("rdma link %s", strings.Join(options, " ")),
		Output:  string(output),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("rdma link command failed: %v", err)
		logger.Debugf("rdma link output: %s", result.Output)
		return result, err
	}

	logger.Info("rdma link command completed successfully")
	logger.Debugf("rdma link output: %s", result.Output)

	return result, nil
}

// RunReadlink executes readlink command to resolve symbolic links
func RunReadlink(path string, options ...string) (*OSCommandResult, error) {
	logger.Infof("Running readlink for path: %s", path)

	// Build command arguments
	args := append(options, path)

	cmd := exec.Command("readlink", args...)
	output, err := cmd.CombinedOutput()

	result := &OSCommandResult{
		Command: fmt.Sprintf("readlink %s %s", strings.Join(options, " "), path),
		Output:  strings.TrimSpace(string(output)),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("readlink command failed: %v", err)
		logger.Debugf("readlink output: %s", result.Output)
		return result, err
	}

	logger.Infof("readlink command completed successfully: %s", result.Output)
	return result, nil
}

// RunLspciByPCI executes lspci for a specific PCI address and returns device information
func RunLspciByPCI(pciAddress string, verbose bool) (*OSCommandResult, error) {
	logger.Infof("Running lspci for PCI address: %s", pciAddress)

	// Extract just the PCI ID from full path if needed
	// e.g., /sys/devices/pci0000:00/0000:00:1f.0 -> 00:1f.0
	pciID := pciAddress
	if strings.Contains(pciAddress, "/") {
		parts := strings.Split(pciAddress, "/")
		for _, part := range parts {
			// Match PCI address pattern: 0000:00:1f.0 (domain:bus:device.function)
			if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]+$`, part); matched {
				pciID = part
				break
			}
		}
	}

	// Remove domain prefix if present (0000:)
	if strings.Count(pciID, ":") == 2 {
		parts := strings.SplitN(pciID, ":", 2)
		if len(parts) == 2 {
			pciID = parts[1]
		}
	}

	args := []string{"-s", pciID}
	if verbose {
		args = append(args, "-v")
	}

	return RunLspci(args...)
}

// GetPCIDeviceModel returns the model/description of a PCI device
func GetPCIDeviceModel(pciAddress string) (string, error) {
	logger.Infof("Getting PCI device model for: %s", pciAddress)

	result, err := RunLspciByPCI(pciAddress, false)
	if err != nil {
		return "", fmt.Errorf("failed to run lspci: %w", err)
	}

	// Parse lspci output to extract device description
	// Format: "00:1f.0 Ethernet controller: Intel Corporation ..."
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by colon to get device description
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			// Remove the device type (e.g., "Ethernet controller")
			model := strings.TrimSpace(parts[2])
			logger.Infof("Found PCI device model: %s", model)
			return model, nil
		}
	}

	return "", fmt.Errorf("could not parse PCI device model from lspci output")
}

// GetPCIDeviceNUMANode returns the NUMA node of a PCI device
func GetPCIDeviceNUMANode(pciAddress string) (string, error) {
	logger.Infof("Getting NUMA node for PCI device: %s", pciAddress)

	// Check /sys/bus/pci/devices/[pci]/numa_node
	sysPath := fmt.Sprintf("/sys/bus/pci/devices/%s/numa_node", pciAddress)
	cmd := exec.Command("cat", sysPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Errorf("Failed to read NUMA node from %s: %v", sysPath, err)
		return "unknown", nil // Don't fail completely, just return unknown
	}

	numaNode := strings.TrimSpace(string(output))
	logger.Infof("Found NUMA node: %s", numaNode)
	return numaNode, nil
}

// GetNetworkInterfaceName returns the network interface name for a PCI device
func GetNetworkInterfaceName(pciAddress string) (string, error) {
	logger.Infof("Getting network interface name for PCI device: %s", pciAddress)

	// Check /sys/bus/pci/devices/[pci]/net/*/
	netPath := fmt.Sprintf("/sys/bus/pci/devices/%s/net", pciAddress)
	cmd := exec.Command("ls", netPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Debugf("No network interface found for PCI device %s: %v", pciAddress, err)
		return "", nil // This is normal for some devices
	}

	interfaces := strings.Fields(string(output))
	if len(interfaces) > 0 {
		interfaceName := interfaces[0] // Take the first interface
		logger.Infof("Found network interface: %s", interfaceName)
		return interfaceName, nil
	}

	return "", nil
}

// GetInfiniBandDeviceName returns the InfiniBand device name for a PCI device
func GetInfiniBandDeviceName(pciAddress string) (string, error) {
	logger.Infof("Getting InfiniBand device name for PCI device: %s", pciAddress)

	// Check /sys/bus/pci/devices/[pci]/infiniband/*/
	ibPath := fmt.Sprintf("/sys/bus/pci/devices/%s/infiniband", pciAddress)
	cmd := exec.Command("ls", ibPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Debugf("No InfiniBand device found for PCI device %s: %v", pciAddress, err)
		return "", nil // This is normal for non-IB devices
	}

	devices := strings.Fields(string(output))
	if len(devices) > 0 {
		deviceName := devices[0] // Take the first device
		logger.Infof("Found InfiniBand device: %s", deviceName)
		return deviceName, nil
	}

	return "", nil
}

// GetRDMADeviceIP returns the IP address of an RDMA device
func GetRDMADeviceIP(deviceName string) (string, error) {
	logger.Infof("Getting RDMA device IP for: %s", deviceName)

	// Use ibdev2netdev to map IB device to network interface
	cmd := exec.Command("ibdev2netdev")
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Errorf("Failed to run ibdev2netdev: %v", err)
		return "", fmt.Errorf("failed to run ibdev2netdev: %w", err)
	}

	// Parse ibdev2netdev output
	// Format: "mlx5_0 port 1 ==> enp12s0f0np0 (Up)"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, deviceName) {
			// Extract network interface name
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "==>" && i+1 < len(parts) {
					netInterface := parts[i+1]
					// Get IP address for this interface
					return getInterfaceIP(netInterface)
				}
			}
		}
	}

	return "", fmt.Errorf("could not find network interface for device %s", deviceName)
}

// getInterfaceIP returns the IP address of a network interface
func getInterfaceIP(interfaceName string) (string, error) {
	logger.Infof("Getting IP address for interface: %s", interfaceName)

	cmd := exec.Command("ip", "addr", "show", interfaceName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Errorf("Failed to get IP for interface %s: %v", interfaceName, err)
		return "", fmt.Errorf("failed to get IP for interface %s: %w", interfaceName, err)
	}

	// Parse ip addr output to find IPv4 address
	// Look for "inet 192.168.1.1/24"
	lines := strings.Split(string(output), "\n")
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

// RunEthtoolStats executes ethtool -S command to get interface statistics with optional grep pattern
func RunEthtoolStats(interfaceName string, grepPattern string) (*OSCommandResult, error) {
	logger.Infof("Running ethtool -S for interface: %s", interfaceName)

	var cmd string
	if grepPattern != "" {
		// Use grep to filter the output
		cmd = fmt.Sprintf("sudo ethtool -S %s | grep %s", interfaceName, grepPattern)
	} else {
		cmd = fmt.Sprintf("sudo ethtool -S %s", interfaceName)
	}

	// Execute the command using shell since we're using pipes
	cmdExec := exec.Command("bash", "-c", cmd)
	output, err := cmdExec.CombinedOutput()

	result := &OSCommandResult{
		Command: cmd,
		Output:  string(output),
		Error:   err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		logger.Errorf("ethtool command failed: %v", err)
		logger.Debugf("ethtool output: %s", result.Output)
		return result, err
	}

	logger.Info("ethtool command completed successfully")
	logger.Debugf("ethtool output for %s: %s", interfaceName, result.Output)

	return result, nil
}
