package autodiscover

import (
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/executor"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// SystemInfo represents comprehensive system information gathered from multiple sources
type SystemInfo struct {
	Hostname         string `json:"hostname"`          // from os_commands.GetHostname()
	OCID             string `json:"ocid"`              // from IMDS instance/id
	FriendlyHostname string `json:"friendly_hostname"` // from os_commands.GetHostname() (same as hostname)
	Shape            string `json:"shape"`             // from IMDS instance/shape
	Serial           string `json:"serial"`            // from dmidecode chassis-serial-number
	Rack             string `json:"rack"`              // from IMDS host/rackId
}

// GatherSystemInfo collects system information from multiple sources
func GatherSystemInfo() (*SystemInfo, error) {
	logger.Info("Gathering comprehensive system information...")

	sysInfo := &SystemInfo{}
	var errors []string

	// Get hostname from OS
	hostname, err := executor.GetHostname()
	if err != nil {
		logger.Errorf("Failed to get hostname: %v", err)
		errors = append(errors, fmt.Sprintf("hostname: %v", err))
	} else {
		sysInfo.Hostname = hostname
		sysInfo.FriendlyHostname = hostname // Both are the same from OS hostname
		logger.Infof("Got hostname: %s", hostname)
	}

	// Get OCID from IMDS
	ocid, err := executor.GetCurrentInstanceOCID()
	if err != nil {
		logger.Errorf("Failed to get OCID: %v", err)
		errors = append(errors, fmt.Sprintf("ocid: %v", err))
	} else {
		sysInfo.OCID = ocid
		logger.Infof("Got OCID: %s", ocid)
	}

	// Get shape from IMDS
	shape, err := executor.GetCurrentShape()
	if err != nil {
		logger.Errorf("Failed to get shape: %v", err)
		errors = append(errors, fmt.Sprintf("shape: %v", err))
	} else {
		sysInfo.Shape = shape
		logger.Infof("Got shape: %s", shape)
	}

	// Get serial number from dmidecode
	serialResult, err := executor.GetSerialNumber()
	if err != nil {
		logger.Errorf("Failed to get serial number: %v", err)
		errors = append(errors, fmt.Sprintf("serial: %v", err))
	} else {
		sysInfo.Serial = serialResult.Output
		logger.Infof("Got serial number: %s", serialResult.Output)
	}

	// Get rack ID from IMDS
	rackID, err := executor.GetCurrentRackID()
	if err != nil {
		logger.Errorf("Failed to get rack ID: %v", err)
		errors = append(errors, fmt.Sprintf("rack: %v", err))
	} else {
		sysInfo.Rack = rackID
		logger.Infof("Got rack ID: %s", rackID)
	}

	// If we have any errors, return them as a combined error
	if len(errors) > 0 {
		logger.Errorf("Some system information could not be gathered: %s", strings.Join(errors, ", "))
		return sysInfo, fmt.Errorf("failed to gather some system information: %s", strings.Join(errors, ", "))
	}

	logger.Info("Successfully gathered all system information")
	return sysInfo, nil
}

// GatherSystemInfoPartial collects system information and continues even if some sources fail
func GatherSystemInfoPartial() *SystemInfo {
	logger.Info("Gathering system information (partial mode - continue on errors)...")

	sysInfo := &SystemInfo{}

	// Get hostname from OS
	if hostname, err := executor.GetHostname(); err != nil {
		logger.Errorf("Failed to get hostname: %v", err)
		sysInfo.Hostname = "unknown"
		sysInfo.FriendlyHostname = "unknown"
	} else {
		sysInfo.Hostname = hostname
		sysInfo.FriendlyHostname = hostname
		logger.Infof("Got hostname: %s", hostname)
	}

	// Get OCID from IMDS
	if ocid, err := executor.GetCurrentInstanceOCID(); err != nil {
		logger.Errorf("Failed to get OCID: %v", err)
		sysInfo.OCID = "unknown"
	} else {
		sysInfo.OCID = ocid
		logger.Infof("Got OCID: %s", ocid)
	}

	// Get shape from IMDS
	if shape, err := executor.GetCurrentShape(); err != nil {
		logger.Errorf("Failed to get shape: %v", err)
		sysInfo.Shape = "unknown"
	} else {
		sysInfo.Shape = shape
		logger.Infof("Got shape: %s", shape)
	}

	// Get serial number from dmidecode
	if serialResult, err := executor.GetSerialNumber(); err != nil {
		logger.Errorf("Failed to get serial number: %v", err)
		sysInfo.Serial = "unknown"
	} else {
		sysInfo.Serial = serialResult.Output
		logger.Infof("Got serial number: %s", serialResult.Output)
	}

	// Get rack ID from IMDS
	if rackID, err := executor.GetCurrentRackID(); err != nil {
		logger.Errorf("Failed to get rack ID: %v", err)
		sysInfo.Rack = "unknown"
	} else {
		sysInfo.Rack = rackID
		logger.Infof("Got rack ID: %s", rackID)
	}

	logger.Info("Completed gathering system information in partial mode")
	return sysInfo
}
