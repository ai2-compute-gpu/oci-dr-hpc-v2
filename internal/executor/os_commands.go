package executor

import (
	"fmt"
	"os"
	"os/exec"
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
