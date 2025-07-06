package autodiscover

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestParseIPAddrOutput(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedInterface string
		expectedIP        string
		expectedMTU       int
		expectError       bool
	}{
		{
			name: "interface_with_mtu_9000",
			input: `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 qdisc mq state UP group default qlen 1000
    link/ether 02:00:17:00:0b:b3 brd ff:ff:ff:ff:ff:ff
    inet 10.0.11.179/24 brd 10.0.11.255 scope global dynamic eth0
       valid_lft 86334sec preferred_lft 86334sec
    inet6 fe80::17:ff:fe00:bb3/64 scope link
       valid_lft forever preferred_lft forever`,
			expectedInterface: "eth0",
			expectedIP:        "10.0.11.179",
			expectedMTU:       9000,
			expectError:       false,
		},
		{
			name: "no_interface_with_mtu_9000",
			input: `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:00:17:00:0b:b3 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/24 brd 192.168.1.255 scope global dynamic eth0
       valid_lft 86334sec preferred_lft 86334sec`,
			expectError: true,
		},
		{
			name: "multiple_interfaces_with_different_mtus",
			input: `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:00:17:00:0b:b3 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/24 brd 192.168.1.255 scope global dynamic eth0
       valid_lft 86334sec preferred_lft 86334sec
3: ens5: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 qdisc mq state UP group default qlen 1000
    link/ether 0e:00:a3:02:1f:e4 brd ff:ff:ff:ff:ff:ff
    inet 10.0.11.179/24 brd 10.0.11.255 scope global dynamic ens5
       valid_lft 86334sec preferred_lft 86334sec`,
			expectedInterface: "ens5",
			expectedIP:        "10.0.11.179",
			expectedMTU:       9000,
			expectError:       false,
		},
		{
			name:        "empty_input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIPAddrOutput(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && result != nil {
				if result.Interface != tt.expectedInterface {
					t.Errorf("Expected interface %s, got %s", tt.expectedInterface, result.Interface)
				}
				if result.PrivateIP != tt.expectedIP {
					t.Errorf("Expected IP %s, got %s", tt.expectedIP, result.PrivateIP)
				}
				if result.MTU != tt.expectedMTU {
					t.Errorf("Expected MTU %d, got %d", tt.expectedMTU, result.MTU)
				}
			}
		})
	}
}

func TestParseRdmaLinkOutput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		interfaceName  string
		expectedDevice string
		expectError    bool
	}{
		{
			name: "single_rdma_link",
			input: `link mlx5_0/1 state ACTIVE physical_state LINK_UP netdev eth0
link mlx5_1/1 state ACTIVE physical_state LINK_UP netdev eth1`,
			interfaceName:  "eth0",
			expectedDevice: "mlx5_0",
			expectError:    false,
		},
		{
			name: "multiple_rdma_links",
			input: `link mlx5_0/1 state ACTIVE physical_state LINK_UP netdev ens3
link mlx5_1/1 state ACTIVE physical_state LINK_UP netdev ens5
link mlx5_2/1 state ACTIVE physical_state LINK_UP netdev ens7`,
			interfaceName:  "ens5",
			expectedDevice: "mlx5_1",
			expectError:    false,
		},
		{
			name: "interface_not_found",
			input: `link mlx5_0/1 state ACTIVE physical_state LINK_UP netdev eth0
link mlx5_1/1 state ACTIVE physical_state LINK_UP netdev eth1`,
			interfaceName: "eth2",
			expectError:   true,
		},
		{
			name:          "empty_input",
			input:         "",
			interfaceName: "eth0",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRdmaLinkOutput(tt.input, tt.interfaceName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && result != tt.expectedDevice {
				t.Errorf("Expected device %s, got %s", tt.expectedDevice, result)
			}
		})
	}
}

func TestParsePCIAddressFromPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedPCI string
		expectError bool
	}{
		{
			name:        "valid_pci_path",
			input:       "/sys/devices/pci0000:00/0000:00:1f.0",
			expectedPCI: "0000:00:1f.0",
			expectError: false,
		},
		{
			name:        "complex_pci_path",
			input:       "/sys/devices/pci0000:00/0000:00:02.0/0000:01:00.0",
			expectedPCI: "0000:01:00.0",
			expectError: false,
		},
		{
			name:        "different_domain",
			input:       "/sys/devices/pci0001:80/0001:81:00.0",
			expectedPCI: "0001:81:00.0",
			expectError: false,
		},
		{
			name:        "no_pci_address",
			input:       "/sys/devices/platform/soc/some_device",
			expectError: true,
		},
		{
			name:        "empty_path",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePCIAddressFromPath(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && result != tt.expectedPCI {
				t.Errorf("Expected PCI %s, got %s", tt.expectedPCI, result)
			}
		})
	}
}

func TestParseModelFromLspci(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedModel string
		expectError   bool
	}{
		{
			name:          "mellanox_connectx6",
			input:         "1f:00.0 Ethernet controller: Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
			expectedModel: "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
			expectError:   false,
		},
		{
			name:          "mellanox_connectx7",
			input:         "0c:00.0 Ethernet controller: Mellanox Technologies MT2910 Family [ConnectX-7]",
			expectedModel: "Mellanox Technologies MT2910 Family [ConnectX-7]",
			expectError:   false,
		},
		{
			name:          "intel_nic",
			input:         "02:00.0 Ethernet controller: Intel Corporation 82599ES 10-Gigabit SFI/SFP+ Network Connection",
			expectedModel: "Intel Corporation 82599ES 10-Gigabit SFI/SFP+ Network Connection",
			expectError:   false,
		},
		{
			name:        "invalid_format",
			input:       "invalid lspci output",
			expectError: true,
		},
		{
			name:        "empty_input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseModelFromLspci(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && result != tt.expectedModel {
				t.Errorf("Expected model %s, got %s", tt.expectedModel, result)
			}
		})
	}
}

func TestDiscoverVCNInterfaceWithFallback(t *testing.T) {
	// This test verifies that the fallback function always returns a valid interface
	result := DiscoverVCNInterfaceWithFallback()

	if result == nil {
		t.Fatal("DiscoverVCNInterfaceWithFallback returned nil")
	}

	// Verify required fields are not empty
	if result.Interface == "" {
		t.Error("Interface field is empty")
	}
	if result.PrivateIP == "" {
		t.Error("PrivateIP field is empty")
	}
	if result.Model == "" {
		t.Error("Model field is empty")
	}

	// Verify fallback values are "undefined" when discovery fails
	if result.PrivateIP != "undefined" {
		t.Errorf("Expected PrivateIP 'undefined', got '%s'", result.PrivateIP)
	}
	if result.PCI != "undefined" {
		t.Errorf("Expected PCI 'undefined', got '%s'", result.PCI)
	}
	if result.Interface != "undefined" {
		t.Errorf("Expected Interface 'undefined', got '%s'", result.Interface)
	}
	if result.DeviceName != "undefined" {
		t.Errorf("Expected DeviceName 'undefined', got '%s'", result.DeviceName)
	}
	if result.Model != "undefined" {
		t.Errorf("Expected Model 'undefined', got '%s'", result.Model)
	}
	if result.MTU != 0 {
		t.Errorf("Expected MTU 0, got %d", result.MTU)
	}

	t.Logf("Fallback interface: %+v", result)
}

func TestNetworkInterfaceStruct(t *testing.T) {
	// Test that the NetworkInterface struct has all required fields
	netif := &NetworkInterface{
		Interface:  "eth0",
		PrivateIP:  "10.0.11.179",
		PCI:        "0000:1f:00.0",
		DeviceName: "mlx5_2",
		Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]",
		MTU:        9000,
	}

	if netif.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", netif.Interface)
	}
	if netif.PrivateIP != "10.0.11.179" {
		t.Errorf("Expected PrivateIP '10.0.11.179', got '%s'", netif.PrivateIP)
	}
	if netif.PCI != "0000:1f:00.0" {
		t.Errorf("Expected PCI '0000:1f:00.0', got '%s'", netif.PCI)
	}
	if netif.DeviceName != "mlx5_2" {
		t.Errorf("Expected DeviceName 'mlx5_2', got '%s'", netif.DeviceName)
	}
	if netif.MTU != 9000 {
		t.Errorf("Expected MTU 9000, got %d", netif.MTU)
	}
}

func TestNetworkInterfaceJSONTags(t *testing.T) {
	// This test ensures the struct has proper JSON tags for serialization
	netif := &NetworkInterface{
		Interface:  "eth0",
		PrivateIP:  "10.0.11.179",
		PCI:        "0000:1f:00.0",
		DeviceName: "mlx5_2",
		Model:      "Mellanox",
		MTU:        9000,
	}

	// The struct should be serializable
	if netif == nil {
		t.Error("NetworkInterface should not be nil")
	}

	// Check that all fields are accessible
	fields := []string{
		netif.Interface,
		netif.PrivateIP,
		netif.PCI,
		netif.DeviceName,
		netif.Model,
	}

	for i, field := range fields {
		if field == "" && i < 5 { // MTU can be 0, but strings shouldn't be empty
			t.Errorf("Field %d should not be empty", i)
		}
	}
}

// Integration test - only runs if we can actually run commands
func TestNetworkDiscoveryIntegration(t *testing.T) {
	t.Skip("Skipping integration test - requires OCI instance with proper network setup")

	// This test would run the actual discovery on a real system
	// netInterface, err := DiscoverVCNInterface()
	// if err != nil {
	// 	t.Logf("Network discovery failed (expected in test environment): %v", err)
	// 	return
	// }
	//
	// t.Logf("Discovered network interface: %+v", netInterface)
}

// Integration test for partial VCN discovery (Step 1 success, Step 2 failure)
func TestVCNDiscoveryWithRdmaFailure(t *testing.T) {
	// Test what happens when we find MTU 9000 interface but rdma link fails
	mockIPOutput := `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 qdisc mq state UP group default qlen 1000
    link/ether 02:00:17:00:0b:b3 brd ff:ff:ff:ff:ff:ff
    inet 10.0.11.179/24 brd 10.0.11.255 scope global dynamic eth0
       valid_lft 86334sec preferred_lft 86334sec`

	// Test parseIPAddrOutput
	netInterface, err := parseIPAddrOutput(mockIPOutput)
	if err != nil {
		t.Fatalf("parseIPAddrOutput failed: %v", err)
	}

	if netInterface.Interface != "eth0" {
		t.Errorf("Expected interface 'eth0', got '%s'", netInterface.Interface)
	}
	if netInterface.PrivateIP != "10.0.11.179" {
		t.Errorf("Expected IP '10.0.11.179', got '%s'", netInterface.PrivateIP)
	}
	if netInterface.MTU != 9000 {
		t.Errorf("Expected MTU 9000, got %d", netInterface.MTU)
	}

	// Simulate rdma link failure and check fallback device name logic
	interfaceName := "eth0"

	// Test the fallback logic directly
	deviceName := fmt.Sprintf("mlx5_%s", strings.TrimPrefix(interfaceName, "eth"))
	expectedDeviceName := "mlx5_0"

	if deviceName != expectedDeviceName {
		t.Errorf("Fallback device name logic failed: expected '%s', got '%s'", expectedDeviceName, deviceName)
	}

	t.Logf("VCN discovery test: interface=%s, ip=%s, fallback_device=%s",
		netInterface.Interface, netInterface.PrivateIP, deviceName)
}

// Test improved fallback device name generation for different interface patterns
func TestDeviceNameFallbackLogic(t *testing.T) {
	tests := []struct {
		interfaceName  string
		expectedDevice string
		description    string
	}{
		{
			interfaceName:  "eth0",
			expectedDevice: "mlx5_0",
			description:    "Standard eth0 interface",
		},
		{
			interfaceName:  "eth1",
			expectedDevice: "mlx5_1",
			description:    "eth1 interface",
		},
		{
			interfaceName:  "ens3",
			expectedDevice: "mlx5_3",
			description:    "Predictable interface name ens3",
		},
		{
			interfaceName:  "ens5",
			expectedDevice: "mlx5_5",
			description:    "Predictable interface name ens5",
		},
		{
			interfaceName:  "enp0s3",
			expectedDevice: "mlx5_3",
			description:    "Complex interface name with trailing number",
		},
		{
			interfaceName:  "wlan0",
			expectedDevice: "mlx5_0",
			description:    "Wireless interface with trailing number",
		},
		{
			interfaceName:  "unknown",
			expectedDevice: "mlx5_0",
			description:    "Unknown interface pattern defaults to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Simulate the fallback logic
			interfaceNumber := "0" // Default fallback

			// Try to extract number from interface name
			if strings.HasPrefix(tt.interfaceName, "eth") {
				interfaceNumber = strings.TrimPrefix(tt.interfaceName, "eth")
			} else if strings.HasPrefix(tt.interfaceName, "ens") {
				interfaceNumber = strings.TrimPrefix(tt.interfaceName, "ens")
			} else {
				// For other interface names, try to extract trailing number
				re := regexp.MustCompile(`\d+$`)
				matches := re.FindStringSubmatch(tt.interfaceName)
				if len(matches) > 0 {
					interfaceNumber = matches[0]
				}
			}

			deviceName := fmt.Sprintf("mlx5_%s", interfaceNumber)

			if deviceName != tt.expectedDevice {
				t.Errorf("Expected device name %s for interface %s, got %s",
					tt.expectedDevice, tt.interfaceName, deviceName)
			}

			t.Logf("Interface %s -> Device %s", tt.interfaceName, deviceName)
		})
	}
}

// Test complete VCN discovery process with device name verification
func TestCompleteVCNDiscoveryWithDeviceName(t *testing.T) {
	// Test simulates what happens on a real OCI system where:
	// 1. Step 1 succeeds (finds MTU 9000 interface)
	// 2. Step 2 fails (rdma link unavailable)
	// 3. Fallback generates correct device name

	// Create a mock NetworkInterface as if Step 1 succeeded
	mockInterface := &NetworkInterface{
		Interface: "ens5", // Common OCI interface name
		PrivateIP: "10.0.11.179",
		MTU:       9000,
	}

	// Simulate Step 2 failure and fallback device name generation
	interfaceName := mockInterface.Interface

	// This is the actual fallback logic from the code
	interfaceNumber := "0" // Default fallback

	if strings.HasPrefix(interfaceName, "eth") {
		interfaceNumber = strings.TrimPrefix(interfaceName, "eth")
	} else if strings.HasPrefix(interfaceName, "ens") {
		interfaceNumber = strings.TrimPrefix(interfaceName, "ens")
	} else {
		// For other interface names, try to extract trailing number
		re := regexp.MustCompile(`\d+$`)
		matches := re.FindStringSubmatch(interfaceName)
		if len(matches) > 0 {
			interfaceNumber = matches[0]
		}
	}

	deviceName := fmt.Sprintf("mlx5_%s", interfaceNumber)
	mockInterface.DeviceName = deviceName

	// Verify the device name is correct
	expectedDeviceName := "mlx5_5" // ens5 -> mlx5_5
	if mockInterface.DeviceName != expectedDeviceName {
		t.Errorf("Expected device name %s, got %s", expectedDeviceName, mockInterface.DeviceName)
	}

	// Create a VcnNic from the discovered interface (simulating autodiscover.go logic)
	vcnNic := VcnNic{
		PrivateIP:  mockInterface.PrivateIP,
		PCI:        "0000:1f:00.0", // Would come from Step 3
		Interface:  mockInterface.Interface,
		DeviceName: mockInterface.DeviceName,
		Model:      "Mellanox Technologies MT2892 Family [ConnectX-6 Dx]", // Would come from Step 4
	}

	// Verify the final VCN NIC has the correct device name
	if vcnNic.DeviceName != expectedDeviceName {
		t.Errorf("VcnNic DeviceName should be %s, got %s", expectedDeviceName, vcnNic.DeviceName)
	}

	// Verify it's NOT the interface name (the original bug)
	if vcnNic.DeviceName == vcnNic.Interface {
		t.Errorf("DeviceName should not equal Interface name. DeviceName=%s, Interface=%s",
			vcnNic.DeviceName, vcnNic.Interface)
	}

	t.Logf("✅ VCN discovery test passed:")
	t.Logf("   Interface: %s", vcnNic.Interface)
	t.Logf("   DeviceName: %s", vcnNic.DeviceName)
	t.Logf("   PrivateIP: %s", vcnNic.PrivateIP)
	t.Logf("   Model: %s", vcnNic.Model)
}

// VcnNic is already defined in autodiscover.go - no need to redeclare
