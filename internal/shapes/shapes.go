// Package shapes provides functionality to read and query OCI shape configurations
package shapes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// ShapeConfig represents the entire shapes configuration
type ShapeConfig struct {
	Version      string            `json:"version"`
	RDMANetwork  []RDMANetwork     `json:"rdma-network"`
	RDMASettings []RDMAShapeConfig `json:"rdma-settings"`
	HPCShapes    []HPCShape        `json:"hpc-shapes"`
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

// HPCShape represents a hardware shape configuration
type HPCShape struct {
	Shape    string      `json:"shape"`
	GPU      interface{} `json:"gpu"` // Can be bool or []GPUSpec
	VCNNics  []VCNNic    `json:"vcn-nics"`
	RDMANics []RDMANic   `json:"rdma-nics"`
}

// VCNNic represents a VCN network interface configuration
type VCNNic struct {
	PCI        string `json:"pci"`
	Interface  string `json:"interface"`
	DeviceName string `json:"device_name"`
	Model      string `json:"model"`
}

// RDMANic represents an RDMA network interface configuration
type RDMANic struct {
	PCI        string        `json:"pci"`
	Interface  string        `json:"interface"`
	DeviceName string        `json:"device_name"`
	Model      string        `json:"model"`
	GPUPCI     string        `json:"gpu_pci,omitempty"`
	GPUID      FlexibleGPUID `json:"gpu_id,omitempty"`
}

// FlexibleGPUID is a custom type that can handle both string and number JSON values
type FlexibleGPUID string

// UnmarshalJSON implements custom JSON unmarshaling for FlexibleGPUID
func (fgid *FlexibleGPUID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*fgid = FlexibleGPUID(str)
		return nil
	}

	// Try to unmarshal as number
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*fgid = FlexibleGPUID(strconv.FormatFloat(num, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleGPUID", data)
}

// String returns the string representation
func (fgid FlexibleGPUID) String() string {
	return string(fgid)
}

// GPUSpec represents a GPU specification in the shapes configuration
type GPUSpec struct {
	PCI      string `json:"pci"`
	Model    string `json:"model"`
	ID       int    `json:"id"`
	ModuleID int    `json:"module_id"`
}

// GetHPCShape returns the HPC shape configuration for a specific shape
func (sm *ShapeManager) GetHPCShape(shapeName string) (*HPCShape, error) {
	for _, hpcShape := range sm.config.HPCShapes {
		if hpcShape.Shape == shapeName {
			logger.Debugf("Found HPC shape configuration for: %s", shapeName)
			return &hpcShape, nil
		}
	}

	logger.Errorf("HPC shape not found: %s", shapeName)
	return nil, fmt.Errorf("HPC shape %s not found in configuration", shapeName)
}

// GetVCNNics returns VCN NICs for a specific shape
func (sm *ShapeManager) GetVCNNics(shapeName string) ([]VCNNic, error) {
	hpcShape, err := sm.GetHPCShape(shapeName)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Found %d VCN NICs for shape %s", len(hpcShape.VCNNics), shapeName)
	return hpcShape.VCNNics, nil
}

// GetRDMANics returns RDMA NICs for a specific shape
func (sm *ShapeManager) GetRDMANics(shapeName string) ([]RDMANic, error) {
	hpcShape, err := sm.GetHPCShape(shapeName)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Found %d RDMA NICs for shape %s", len(hpcShape.RDMANics), shapeName)
	return hpcShape.RDMANics, nil
}

// GetAllHPCShapes returns all available HPC shapes
func (sm *ShapeManager) GetAllHPCShapes() []string {
	var shapes []string
	for _, hpcShape := range sm.config.HPCShapes {
		shapes = append(shapes, hpcShape.Shape)
	}
	return shapes
}

// GetGPUSpecs returns GPU specifications for a specific shape
func (sm *ShapeManager) GetGPUSpecs(shapeName string) ([]GPUSpec, error) {
	hpcShape, err := sm.GetHPCShape(shapeName)
	if err != nil {
		return nil, err
	}

	// Handle the case where GPU is a boolean (false = no GPUs)
	if gpuBool, ok := hpcShape.GPU.(bool); ok {
		if !gpuBool {
			return []GPUSpec{}, nil // No GPUs for this shape
		}
		return nil, fmt.Errorf("GPU field is true but no GPU specifications found for shape %s", shapeName)
	}

	// Handle the case where GPU is an array of GPU specifications
	if gpuArray, ok := hpcShape.GPU.([]interface{}); ok {
		var gpuSpecs []GPUSpec
		for _, gpuInterface := range gpuArray {
			if gpuMap, ok := gpuInterface.(map[string]interface{}); ok {
				gpuSpec := GPUSpec{}
				
				if pci, exists := gpuMap["pci"]; exists {
					if pciStr, ok := pci.(string); ok {
						gpuSpec.PCI = pciStr
					}
				}
				
				if model, exists := gpuMap["model"]; exists {
					if modelStr, ok := model.(string); ok {
						gpuSpec.Model = modelStr
					}
				}
				
				if id, exists := gpuMap["id"]; exists {
					if idFloat, ok := id.(float64); ok {
						gpuSpec.ID = int(idFloat)
					}
				}
				
				if moduleID, exists := gpuMap["module_id"]; exists {
					if moduleIDFloat, ok := moduleID.(float64); ok {
						gpuSpec.ModuleID = int(moduleIDFloat)
					}
				}
				
				gpuSpecs = append(gpuSpecs, gpuSpec)
			}
		}
		
		logger.Debugf("Found %d GPU specifications for shape %s", len(gpuSpecs), shapeName)
		return gpuSpecs, nil
	}

	return nil, fmt.Errorf("GPU field has unexpected type for shape %s", shapeName)
}

// GetGPUPCIAddresses returns a list of GPU PCI addresses for a specific shape
func (sm *ShapeManager) GetGPUPCIAddresses(shapeName string) ([]string, error) {
	gpuSpecs, err := sm.GetGPUSpecs(shapeName)
	if err != nil {
		return nil, err
	}

	var pciAddresses []string
	for _, gpu := range gpuSpecs {
		pciAddresses = append(pciAddresses, gpu.PCI)
	}

	return pciAddresses, nil
}

// GetGPUIndices returns a list of GPU indices for a specific shape
func (sm *ShapeManager) GetGPUIndices(shapeName string) ([]string, error) {
	gpuSpecs, err := sm.GetGPUSpecs(shapeName)
	if err != nil {
		return nil, err
	}

	var indices []string
	for _, gpu := range gpuSpecs {
		indices = append(indices, strconv.Itoa(gpu.ID))
	}

	return indices, nil
}

// GetGPUModuleIDs returns a list of GPU module IDs for a specific shape
func (sm *ShapeManager) GetGPUModuleIDs(shapeName string) ([]string, error) {
	gpuSpecs, err := sm.GetGPUSpecs(shapeName)
	if err != nil {
		return nil, err
	}

	var moduleIDs []string
	for _, gpu := range gpuSpecs {
		moduleIDs = append(moduleIDs, strconv.Itoa(gpu.ModuleID))
	}

	return moduleIDs, nil
}

// HasGPUs returns true if the shape has GPU configurations
func (sm *ShapeManager) HasGPUs(shapeName string) (bool, error) {
	hpcShape, err := sm.GetHPCShape(shapeName)
	if err != nil {
		return false, err
	}

	// Handle the case where GPU is a boolean
	if gpuBool, ok := hpcShape.GPU.(bool); ok {
		return gpuBool, nil
	}

	// Handle the case where GPU is an array
	if gpuArray, ok := hpcShape.GPU.([]interface{}); ok {
		return len(gpuArray) > 0, nil
	}

	return false, nil
}

// getPackageDir returns the directory where this package is located
func getPackageDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get current file path")
	}
	return filepath.Dir(filename), nil
}

// getDefaultShapesPath returns the default path to shapes.json
func getDefaultShapesPath() (string, error) {
	packageDir, err := getPackageDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(packageDir, "shapes.json"), nil
}

// GetDefaultShapeManager returns a ShapeManager using the default shapes.json file
func GetDefaultShapeManager() (*ShapeManager, error) {
	shapesPath, err := getDefaultShapesPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get default shapes path: %w", err)
	}

	return NewShapeManager(shapesPath)
}
