// =============================================================================
// internal/discovery/discovery.go - HPC environment discovery
package discovery

import (
	"fmt"
	"os"
	"runtime"
)

// SystemInfo holds basic system information
type SystemInfo struct {
	OS       string
	Arch     string
	Hostname string
}

// DiscoverSystem discovers basic system information
func DiscoverSystem() (*SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	return &SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}, nil
}

// DiscoverHPCEnvironment discovers HPC-specific environment details
func DiscoverHPCEnvironment() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	sysInfo, err := DiscoverSystem()
	if err != nil {
		return nil, err
	}

	info["system"] = sysInfo
	info["cpu_count"] = runtime.NumCPU()

	return info, nil
}
