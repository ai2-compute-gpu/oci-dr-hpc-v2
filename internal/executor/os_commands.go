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
