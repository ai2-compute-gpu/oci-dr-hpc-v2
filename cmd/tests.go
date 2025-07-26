package cmd

import "github.com/oracle/oci-dr-hpc-v2/internal/level1_tests"

// TestDefinition represents a single diagnostic test
type TestDefinition struct {
	Name string
	Fn   func() error
}

// TestDefinitionWithDescription represents a diagnostic test with description
type TestDefinitionWithDescription struct {
	Name        string
	Description string
	Fn          func() error
}

// GetLevel1Tests returns all Level 1 diagnostic tests
func GetLevel1Tests() []TestDefinition {
	return []TestDefinition{
		{"gpu_count_check", level1_tests.RunGPUCountCheck},
		{"pcie_error_check", level1_tests.RunPCIeErrorCheck},
		{"rdma_nics_count", level1_tests.RunRDMANicsCount},
		{"rx_discards_check", level1_tests.RunRXDiscardsCheck},
		{"gid_index_check", level1_tests.RunGIDIndexCheck},
		{"link_check", level1_tests.RunLinkCheck},
		{"eth_link_check", level1_tests.RunEthLinkCheck},
		{"auth_check", level1_tests.RunAuthCheck},
		{"sram_error_check", level1_tests.RunSRAMCheck},
		{"gpu_mode_check", level1_tests.RunGPUModeCheck},
		{"gpu_driver_check", level1_tests.RunGPUDriverCheck},
		{"gpu_clk_check", level1_tests.RunGPUClkCheck},
		{"peermem_module_check", level1_tests.RunPeermemModuleCheck},
		{"nvlink_speed_check", level1_tests.RunNVLinkSpeedCheck},
		{"eth0_presence_check", level1_tests.RunEth0PresenceCheck},
		{"cdfp_cable_check", level1_tests.RunCDFPCableCheck},
		{"fabricmanager_check", level1_tests.RunFabricManagerCheck},
		{"hca_error_check", level1_tests.RunHCAErrorCheck},
		{"missing_interface_check", level1_tests.RunMissingInterfaceCheck},
		{"gpu_xid_check", level1_tests.RunGPUXIDCheck},
		{"max_acc_check", level1_tests.RunMaxAccCheck},
		{"row_remap_error_check", level1_tests.RunRowRemapErrorCheck},
	}
}

// GetLevel1TestsWithDescriptions returns all Level 1 diagnostic tests with descriptions
func GetLevel1TestsWithDescriptions() []TestDefinitionWithDescription {
	return []TestDefinitionWithDescription{
		{"gpu_count_check", "Check GPU count using nvidia-smi", level1_tests.RunGPUCountCheck},
		{"pcie_error_check", "Check for PCIe errors in system logs", level1_tests.RunPCIeErrorCheck},
		{"rdma_nics_count", "Check RDMA NICs count", level1_tests.RunRDMANicsCount},
		{"rx_discards_check", "Check Network Interface for rx discard", level1_tests.RunRXDiscardsCheck},
		{"gid_index_check", "Check device GID Index are in range", level1_tests.RunGIDIndexCheck},
		{"link_check", "Check RDMA link state and parameters", level1_tests.RunLinkCheck},
		{"eth_link_check", "Check Ethernet link state and parameters for 100GbE RoCE interfaces", level1_tests.RunEthLinkCheck},
		{"auth_check", "Check authentication status of RDMA interfaces using wpa_cli", level1_tests.RunAuthCheck},
		{"sram_error_check", "Check SRAM correctable and uncorrectable errors", level1_tests.RunSRAMCheck},
		{"gpu_mode_check", "Check if GPU is in Multi-Instance GPU (MIG) mode", level1_tests.RunGPUModeCheck},
		{"gpu_driver_check", "Check GPU driver version compatibility", level1_tests.RunGPUDriverCheck},
		{"gpu_clk_check", "Check GPU clock speeds are within acceptable range", level1_tests.RunGPUClkCheck},
		{"peermem_module_check", "Check for presence of peermem module", level1_tests.RunPeermemModuleCheck},
		{"nvlink_speed_check", "Check for presence and speed for nvlink", level1_tests.RunNVLinkSpeedCheck},
		{"eth0_presence_check", "Check if eth0 network interface is present", level1_tests.RunEth0PresenceCheck},
		{"cdfp_cable_check", "Check CDFP cable connections between GPUs", level1_tests.RunCDFPCableCheck},
		{"fabricmanager_check", "Check if nvidia-fabricmanager service is running", level1_tests.RunFabricManagerCheck},
		{"hca_error_check", "Check for MLX5 HCA fatal errors in system logs", level1_tests.RunHCAErrorCheck},
		{"missing_interface_check", "Check for missing PCIe interfaces (revision ff)", level1_tests.RunMissingInterfaceCheck},
		{"gpu_xid_check", "Check for NVIDIA GPU XID errors in system logs", level1_tests.RunGPUXIDCheck},
		{"max_acc_check", "Check MAX_ACC_OUT_READ and ADVANCED_PCI_SETTINGS configuration", level1_tests.RunMaxAccCheck},
		{"row_remap_error_check", "Check for GPU row remap errors using nvidia-smi", level1_tests.RunRowRemapErrorCheck},
	}
}