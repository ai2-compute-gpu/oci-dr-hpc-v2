package autodiscover

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/spf13/viper"
)

type RdmaNic struct {
	PCI        string `json:"pci"`
	Interface  string `json:"interface"`
	RdmaIP     string `json:"rdma_ip"`
	DeviceName string `json:"device_name"`
	Model      string `json:"model"`
	Numa       string `json:"numa"`
	GpuID      string `json:"gpu_id"`
	GpuPCI     string `json:"gpu_pci"`
}

type GPU struct {
	PCI   string `json:"pci"`
	Model string `json:"model"`
	ID    int    `json:"id"`
}

type VcnNic struct {
	PrivateIP  string `json:"private_ip"`
	PCI        string `json:"pci"`
	Interface  string `json:"interface"`
	DeviceName string `json:"device_name"`
	Model      string `json:"model"`
}

type MapHost struct {
	Hostname         string    `json:"hostname"`
	Ocid             string    `json:"ocid"`
	FriendlyHostname string    `json:"friendly_hostname"`
	Shape            string    `json:"shape"`
	Serial           string    `json:"serial"`
	Rack             string    `json:"rack"`
	RdmaNics         []RdmaNic `json:"rdma_nics"`
	Gpus             []GPU     `json:"gpu"`
	InCluster        bool      `json:"in_cluster"`
	VcnNic           VcnNic    `json:"vcn_nic"`
}

// formatJSON formats the hardware discovery data as JSON
func formatJSON(mapHost *MapHost) (string, error) {
	jsonData, err := json.MarshalIndent(mapHost, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData) + "\n", nil
}

// formatTable formats the hardware discovery data as a table
func formatTable(mapHost *MapHost) (string, error) {
	var output strings.Builder

	output.WriteString("┌─────────────────────────────────────────────────────────────────┐\n")
	output.WriteString("│                    HARDWARE DISCOVERY RESULTS                  │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// System Information
	output.WriteString("│ SYSTEM INFORMATION                                              │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Hostname: %s", mapHost.Hostname)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Shape: %s", mapHost.Shape)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Serial: %s", mapHost.Serial)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("Rack: %s", mapHost.Rack)))
	output.WriteString(fmt.Sprintf("│ %-63s │\n", fmt.Sprintf("In Cluster: %t", mapHost.InCluster)))
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// GPU Information
	output.WriteString("│ GPU DEVICES                                                     │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	if len(mapHost.Gpus) > 0 {
		for _, gpu := range mapHost.Gpus {
			output.WriteString(fmt.Sprintf("│ GPU %d: %-54s │\n", gpu.ID, gpu.Model))
			output.WriteString(fmt.Sprintf("│   PCI: %-56s │\n", gpu.PCI))
		}
	} else {
		output.WriteString("│ No GPU devices detected                                         │\n")
	}
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// RDMA NIC Information
	output.WriteString("│ RDMA NETWORK INTERFACES                                         │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	if len(mapHost.RdmaNics) > 0 {
		for i, rdma := range mapHost.RdmaNics {
			output.WriteString(fmt.Sprintf("│ RDMA NIC %d: %-51s │\n", i+1, rdma.Interface))
			output.WriteString(fmt.Sprintf("│   PCI: %-56s │\n", rdma.PCI))
			output.WriteString(fmt.Sprintf("│   IP: %-57s │\n", rdma.RdmaIP))
			output.WriteString(fmt.Sprintf("│   Model: %-54s │\n", rdma.Model))
		}
	} else {
		output.WriteString("│ No RDMA NICs detected                                           │\n")
	}
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	// VCN NIC Information
	output.WriteString("│ VCN NETWORK INTERFACE                                           │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString(fmt.Sprintf("│ Interface: %-52s │\n", mapHost.VcnNic.Interface))
	output.WriteString(fmt.Sprintf("│ PCI: %-59s │\n", mapHost.VcnNic.PCI))
	output.WriteString(fmt.Sprintf("│ Private IP: %-51s │\n", mapHost.VcnNic.PrivateIP))
	output.WriteString(fmt.Sprintf("│ Model: %-57s │\n", mapHost.VcnNic.Model))

	output.WriteString("└─────────────────────────────────────────────────────────────────┘\n")
	return output.String(), nil
}

// formatFriendly formats the hardware discovery data in a user-friendly format
func formatFriendly(mapHost *MapHost) (string, error) {
	var output strings.Builder

	output.WriteString("🔍 Hardware Discovery Results\n")
	output.WriteString("=" + strings.Repeat("=", 50) + "\n\n")

	// System Information
	output.WriteString("🖥️  System Information\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	output.WriteString(fmt.Sprintf("   Hostname: %s\n", mapHost.Hostname))
	output.WriteString(fmt.Sprintf("   Shape: %s\n", mapHost.Shape))
	output.WriteString(fmt.Sprintf("   Serial Number: %s\n", mapHost.Serial))
	output.WriteString(fmt.Sprintf("   Rack Location: %s\n", mapHost.Rack))
	clusterStatus := "Yes"
	if !mapHost.InCluster {
		clusterStatus = "No"
	}
	output.WriteString(fmt.Sprintf("   In Cluster: %s\n\n", clusterStatus))

	// GPU Information
	output.WriteString("🎮 GPU Devices\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	if len(mapHost.Gpus) > 0 {
		for _, gpu := range mapHost.Gpus {
			output.WriteString(fmt.Sprintf("   ✅ GPU %d: %s\n", gpu.ID, gpu.Model))
			output.WriteString(fmt.Sprintf("      PCI Address: %s\n", gpu.PCI))
		}
		output.WriteString(fmt.Sprintf("\n   Total GPUs detected: %d\n\n", len(mapHost.Gpus)))
	} else {
		output.WriteString("   ❌ No GPU devices detected\n\n")
	}

	// RDMA Network Information
	output.WriteString("🌐 RDMA Network Interfaces\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	if len(mapHost.RdmaNics) > 0 {
		for i, rdma := range mapHost.RdmaNics {
			output.WriteString(fmt.Sprintf("   ✅ RDMA NIC %d: %s\n", i+1, rdma.Interface))
			output.WriteString(fmt.Sprintf("      PCI Address: %s\n", rdma.PCI))
			output.WriteString(fmt.Sprintf("      IP Address: %s\n", rdma.RdmaIP))
			output.WriteString(fmt.Sprintf("      Device: %s\n", rdma.DeviceName))
			output.WriteString(fmt.Sprintf("      Model: %s\n", rdma.Model))
			output.WriteString(fmt.Sprintf("      NUMA Node: %s\n", rdma.Numa))
			if rdma.GpuID != "" {
				output.WriteString(fmt.Sprintf("      Linked to GPU %s (%s)\n", rdma.GpuID, rdma.GpuPCI))
			}
			output.WriteString("\n")
		}
		output.WriteString(fmt.Sprintf("   Total RDMA NICs detected: %d\n\n", len(mapHost.RdmaNics)))
	} else {
		output.WriteString("   ❌ No RDMA NICs detected\n\n")
	}

	// VCN Network Information
	output.WriteString("🔗 VCN Network Interface\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	output.WriteString(fmt.Sprintf("   ✅ Interface: %s\n", mapHost.VcnNic.Interface))
	output.WriteString(fmt.Sprintf("      PCI Address: %s\n", mapHost.VcnNic.PCI))
	output.WriteString(fmt.Sprintf("      Private IP: %s\n", mapHost.VcnNic.PrivateIP))
	output.WriteString(fmt.Sprintf("      Device: %s\n", mapHost.VcnNic.DeviceName))
	output.WriteString(fmt.Sprintf("      Model: %s\n\n", mapHost.VcnNic.Model))

	// Summary
	output.WriteString("📊 Discovery Summary\n")
	output.WriteString("   " + strings.Repeat("-", 30) + "\n")
	output.WriteString(fmt.Sprintf("   GPUs: %d\n", len(mapHost.Gpus)))
	output.WriteString(fmt.Sprintf("   RDMA NICs: %d\n", len(mapHost.RdmaNics)))
	output.WriteString("   VCN NICs: 1\n")

	totalDevices := len(mapHost.Gpus) + len(mapHost.RdmaNics) + 1
	output.WriteString(fmt.Sprintf("   Total devices: %d\n\n", totalDevices))

	if len(mapHost.Gpus) > 0 && len(mapHost.RdmaNics) > 0 {
		output.WriteString("   🎉 Hardware discovery completed successfully!\n")
		output.WriteString("   Your HPC system appears to be properly configured.\n")
	} else {
		output.WriteString("   ⚠️  Some hardware components may be missing.\n")
		output.WriteString("   Please verify your system configuration.\n")
	}

	return output.String(), nil
}

func Run() {
	logger.Info("Running autodiscover...")

	// Gather real system information
	sysInfo := GatherSystemInfoPartial()

	// Mocked data for now; replace with real discovery logic
	mapHost := MapHost{
		Hostname:         sysInfo.Hostname,
		Ocid:             sysInfo.OCID,
		FriendlyHostname: sysInfo.FriendlyHostname,
		Shape:            sysInfo.Shape,
		Serial:           sysInfo.Serial,
		Rack:             sysInfo.Rack,
		InCluster:        true,
		Gpus: []GPU{
			{PCI: "0000:0f:00.0", Model: "NVIDIA H100 80GB HBM3", ID: 0},
			{PCI: "0000:2d:00.0", Model: "NVIDIA H100 80GB HBM3", ID: 1},
		},
		RdmaNics: []RdmaNic{
			{
				PCI:        "0000:0c:00.0",
				Interface:  "rdma0",
				RdmaIP:     "192.168.3.179",
				DeviceName: "mlx5_0",
				Model:      "Mellanox Technologies MT2910 Family [ConnectX-7]",
				Numa:       "0",
				GpuID:      "0",
				GpuPCI:     "0000:0f:00.0",
			},
		},
		VcnNic: VcnNic{
			PrivateIP:  "10.0.11.179",
			PCI:        "0000:1f:00.0",
			Interface:  "eth0",
			DeviceName: "eth0",
			Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
		},
	}

	// Get output format from configuration
	outputFormat := viper.GetString("output")
	if outputFormat == "" {
		outputFormat = "json" // Default to JSON format for autodiscover
	}

	// Generate output in the specified format
	var output string
	var err error
	switch outputFormat {
	case "json":
		output, err = formatJSON(&mapHost)
	case "table":
		output, err = formatTable(&mapHost)
	case "friendly":
		output, err = formatFriendly(&mapHost)
	default:
		logger.Errorf("Unsupported output format: %s", outputFormat)
		return
	}

	if err != nil {
		logger.Errorf("Failed to format output: %v", err)
		return
	}

	// Output to console
	fmt.Print(output)

	// Determine output file name
	var outputFile string
	if userOutputFile := viper.GetString("output-file"); userOutputFile != "" {
		outputFile = userOutputFile
	} else {
		outputFile = fmt.Sprintf("map_host_%s.json", strings.ToLower(sysInfo.Hostname))
	}

	// Create output directory if it doesn't exist (for user-specified paths)
	if strings.Contains(outputFile, "/") {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Errorf("Failed to create output directory: %v", err)
			return
		}
	}

	// For file output, always use JSON format regardless of console format
	var fileOutput string
	if outputFormat == "json" {
		fileOutput = output
	} else {
		fileOutput, err = formatJSON(&mapHost)
		if err != nil {
			logger.Errorf("Failed to format JSON for file output: %v", err)
			return
		}
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(fileOutput), 0644); err != nil {
		logger.Errorf("Failed to write output file: %v", err)
		return
	}

	logger.Infof("Autodiscovery complete. Output written to %s", outputFile)
}
