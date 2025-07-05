// Package shapes provides functionality to read and query OCI shape configurations
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// ShapeConfig represents the entire shapes configuration
type ShapeConfig struct {
	Version      string            `json:"version"`
	RDMANetwork  []RDMANetwork     `json:"rdma-network"`
	RDMASettings []RDMAShapeConfig `json:"rdma-settings"`
}

// RDMANetwork represents RDMA network configuration
type RDMANetwork struct {
	DefaultSettings DefaultSettings `json:"default-settings"`
	SubnetSettings  SubnetSettings  `json:"subnet-settings"`
	ARPSettings     ARPSettings     `json:"arp-settings"`
}

// DefaultSettings represents default RDMA network settings
type DefaultSettings struct {
	RDMANetwork          string `json:"rdma_network"`
	OverwriteConfigFiles bool   `json:"overwrite_config_files"`
	ModifySubnet         bool   `json:"modify_subnet"`
	ModifyARP            bool   `json:"modify_arp"`
}

// SubnetSettings represents subnet configuration
type SubnetSettings struct {
	Netmask                  string `json:"netmask"`
	OverrideNetconfigNetmask string `json:"override_netconfig_netmask"`
}

// ARPSettings represents ARP configuration
type ARPSettings struct {
	RPFilter    string `json:"rp_filter"`
	ARPIgnore   string `json:"arp_ignore"`
	ARPAnnounce string `json:"arp_announce"`
}

// RDMAShapeConfig represents configuration for a group of shapes
type RDMAShapeConfig struct {
	Shapes   []string      `json:"shapes"`
	Model    string        `json:"model"`
	Settings ShapeSettings `json:"settings"`
}

// ShapeSettings represents all settings for a shape
type ShapeSettings struct {
	Ring                RingSettings           `json:"ring"`
	Channels            string                 `json:"channels"`
	TxQueueLength       string                 `json:"tx_queue_length"`
	MTU                 string                 `json:"mtu"`
	RateToSetOnFirstCNP string                 `json:"rate_to_set_on_first_cnp"`
	DSCPIP              string                 `json:"dscp_ip"`
	DSCPRDMA            string                 `json:"dscp_rdma"`
	DSCPGPU             string                 `json:"dscp_gpu"`
	DSCPCNP             string                 `json:"dscp_cnp"`
	DSCPRDMATos         string                 `json:"dscp_rdma_tos"`
	DSCPGPUTos          string                 `json:"dscp_gpu_tos"`
	DSCPDefaultTos      string                 `json:"dscp_default_tos"`
	DSCPIPTC            string                 `json:"dscp_ip_tc"`
	DSCPRDMATC          string                 `json:"dscp_rdma_tc"`
	DSCPGPUTC           string                 `json:"dscp_gpu_tc"`
	DSCPCNPTC           string                 `json:"dscp_cnp_tc"`
	DSCPDefaultTC       string                 `json:"dscp_default_tc"`
	ROCEAccl            map[string]interface{} `json:"roce_accl"`
	Buffer              []string               `json:"buffer"`
	PFC                 []string               `json:"pfc"`
	Prio2Buffer         []string               `json:"prio2buffer"`
	Prio2TC             []string               `json:"prio2tc"`
	Sysctl              map[string]interface{} `json:"sysctl"`
	MLXConfig           map[string]interface{} `json:"mlxconfig"`
	PCIe                map[string]interface{} `json:"pcie"`
}

// RingSettings represents ring buffer settings
type RingSettings struct {
	RX string `json:"rx"`
	TX string `json:"tx"`
}

// ShapeManager manages shape configurations
type ShapeManager struct {
	config   *ShapeConfig
	filePath string
}

// NewShapeManager creates a new ShapeManager
func NewShapeManager(filePath string) (*ShapeManager, error) {
	manager := &ShapeManager{
		filePath: filePath,
	}

	if err := manager.LoadConfig(); err != nil {
		return nil, err
	}

	return manager, nil
}

// LoadConfig loads the shapes configuration from the JSON file
func (sm *ShapeManager) LoadConfig() error {
	logger.Info("Loading shapes configuration from:", sm.filePath)

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		logger.Errorf("Failed to read shapes file: %v", err)
		return fmt.Errorf("failed to read shapes file: %w", err)
	}

	var config ShapeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Errorf("Failed to parse shapes JSON: %v", err)
		return fmt.Errorf("failed to parse shapes JSON: %w", err)
	}

	sm.config = &config
	logger.Infof("Successfully loaded %d RDMA shape configurations", len(config.RDMASettings))

	return nil
}

// GetAllShapes returns a list of all supported shapes
func (sm *ShapeManager) GetAllShapes() []string {
	var allShapes []string

	for _, rdmaConfig := range sm.config.RDMASettings {
		allShapes = append(allShapes, rdmaConfig.Shapes...)
	}

	return allShapes
}

// GetShapeConfig returns the configuration for a specific shape
func (sm *ShapeManager) GetShapeConfig(shapeName string) (*RDMAShapeConfig, error) {
	for _, rdmaConfig := range sm.config.RDMASettings {
		for _, shape := range rdmaConfig.Shapes {
			if shape == shapeName {
				logger.Debugf("Found configuration for shape: %s", shapeName)
				return &rdmaConfig, nil
			}
		}
	}

	logger.Errorf("Shape not found: %s", shapeName)
	return nil, fmt.Errorf("shape %s not found in configuration", shapeName)
}

// GetShapesByModel returns all shapes that use a specific model
func (sm *ShapeManager) GetShapesByModel(model string) []string {
	var shapes []string

	for _, rdmaConfig := range sm.config.RDMASettings {
		if strings.Contains(rdmaConfig.Model, model) {
			shapes = append(shapes, rdmaConfig.Shapes...)
		}
	}

	return shapes
}

// GetSupportedModels returns a list of all supported models
func (sm *ShapeManager) GetSupportedModels() []string {
	var models []string
	seen := make(map[string]bool)

	for _, rdmaConfig := range sm.config.RDMASettings {
		if !seen[rdmaConfig.Model] {
			models = append(models, rdmaConfig.Model)
			seen[rdmaConfig.Model] = true
		}
	}

	return models
}

// IsShapeSupported checks if a shape is supported
func (sm *ShapeManager) IsShapeSupported(shapeName string) bool {
	_, err := sm.GetShapeConfig(shapeName)
	return err == nil
}

// GetRDMANetworkConfig returns the RDMA network configuration
func (sm *ShapeManager) GetRDMANetworkConfig() []RDMANetwork {
	return sm.config.RDMANetwork
}

// GetShapeSettings returns just the settings for a specific shape
func (sm *ShapeManager) GetShapeSettings(shapeName string) (*ShapeSettings, error) {
	config, err := sm.GetShapeConfig(shapeName)
	if err != nil {
		return nil, err
	}

	return &config.Settings, nil
}

// GetShapeModel returns the model for a specific shape
func (sm *ShapeManager) GetShapeModel(shapeName string) (string, error) {
	config, err := sm.GetShapeConfig(shapeName)
	if err != nil {
		return "", err
	}

	return config.Model, nil
}

// GetGPUShapes returns all shapes that contain "GPU" in their name
func (sm *ShapeManager) GetGPUShapes() []string {
	var gpuShapes []string

	allShapes := sm.GetAllShapes()
	for _, shape := range allShapes {
		if strings.Contains(shape, "GPU") {
			gpuShapes = append(gpuShapes, shape)
		}
	}

	return gpuShapes
}

// GetHPCShapes returns all shapes that contain "HPC" in their name
func (sm *ShapeManager) GetHPCShapes() []string {
	var hpcShapes []string

	allShapes := sm.GetAllShapes()
	for _, shape := range allShapes {
		if strings.Contains(shape, "HPC") {
			hpcShapes = append(hpcShapes, shape)
		}
	}

	return hpcShapes
}

// SearchShapes searches for shapes by partial name match
func (sm *ShapeManager) SearchShapes(query string) []string {
	var matchingShapes []string
	query = strings.ToUpper(query)

	allShapes := sm.GetAllShapes()
	for _, shape := range allShapes {
		if strings.Contains(strings.ToUpper(shape), query) {
			matchingShapes = append(matchingShapes, shape)
		}
	}

	return matchingShapes
}

// GetShapeInfo returns comprehensive information about a shape
func (sm *ShapeManager) GetShapeInfo(shapeName string) (*ShapeInfo, error) {
	config, err := sm.GetShapeConfig(shapeName)
	if err != nil {
		return nil, err
	}

	info := &ShapeInfo{
		Name:     shapeName,
		Model:    config.Model,
		Settings: config.Settings,
		IsGPU:    strings.Contains(shapeName, "GPU"),
		IsHPC:    strings.Contains(shapeName, "HPC"),
	}

	return info, nil
}

// ShapeInfo represents comprehensive information about a shape
type ShapeInfo struct {
	Name     string        `json:"name"`
	Model    string        `json:"model"`
	Settings ShapeSettings `json:"settings"`
	IsGPU    bool          `json:"is_gpu"`
	IsHPC    bool          `json:"is_hpc"`
}

// String returns a string representation of ShapeInfo
func (si *ShapeInfo) String() string {
	return fmt.Sprintf("Shape: %s, Model: %s, GPU: %v, HPC: %v",
		si.Name, si.Model, si.IsGPU, si.IsHPC)
}
