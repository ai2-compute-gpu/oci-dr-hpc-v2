package executor

import (
	"fmt"
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
