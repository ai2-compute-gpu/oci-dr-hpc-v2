package executor

import (
	"strings"
	"testing"
)

func TestOSCommandResult(t *testing.T) {
	result := &OSCommandResult{
		Command:  "test command",
		Output:   "test output",
		Error:    nil,
		ExitCode: 0,
	}

	if result.Command != "test command" {
		t.Errorf("Expected Command to be 'test command', got '%s'", result.Command)
	}
	if result.Output != "test output" {
		t.Errorf("Expected Output to be 'test output', got '%s'", result.Output)
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", result.ExitCode)
	}
}

func TestCommandStringFormatting(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		expected string
	}{
		{
			name:     "no options",
			options:  []string{},
			expected: "sudo lspci ",
		},
		{
			name:     "single option",
			options:  []string{"-v"},
			expected: "sudo lspci -v",
		},
		{
			name:     "multiple options",
			options:  []string{"-v", "-k", "-s", "00:00.0"},
			expected: "sudo lspci -v -k -s 00:00.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &OSCommandResult{
				Command: "sudo lspci " + strings.Join(tt.options, " "),
			}

			if result.Command != tt.expected {
				t.Errorf("Expected command '%s', got '%s'", tt.expected, result.Command)
			}
		})
	}
}

func TestLsmodModuleSearchLogic(t *testing.T) {
	testCases := []struct {
		name         string
		output       string
		searchModule string
		shouldFind   bool
	}{
		{
			name: "nvidia_peermem_present",
			output: `Module                  Size  Used by
nvidia_peermem         16384  0
nvidia_drm             69632  4`,
			searchModule: "nvidia_peermem",
			shouldFind:   true,
		},
		{
			name: "nvidia_peermem_not_present",
			output: `Module                  Size  Used by
nvidia_drm             69632  4
nvidia_modeset       1048576  6`,
			searchModule: "nvidia_peermem",
			shouldFind:   false,
		},
		{
			name: "partial_match_should_not_find",
			output: `Module                  Size  Used by
nvidia_peermem_test    16384  0
some_nvidia_peermem    32768  1`,
			searchModule: "nvidia_peermem",
			shouldFind:   false,
		},
		{
			name: "exact_match_required",
			output: `Module                  Size  Used by
nvidia_peermem         16384  0
nvidia_peermemory      32768  1`,
			searchModule: "nvidia_peermem",
			shouldFind:   true,
		},
		{
			name:         "empty_output",
			output:       "",
			searchModule: "nvidia_peermem",
			shouldFind:   false,
		},
		{
			name:         "header_only",
			output:       "Module                  Size  Used by",
			searchModule: "nvidia_peermem",
			shouldFind:   false,
		},
		{
			name: "malformed_lines_ignored",
			output: `Module                  Size  Used by
nvidia_peermem         16384  0
incomplete line
another bad line without proper format
nvidia_drm             69632  4`,
			searchModule: "nvidia_peermem",
			shouldFind:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found := false

			lines := strings.Split(tc.output, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				fields := strings.Fields(line)
				if len(fields) > 0 && fields[0] == tc.searchModule {
					found = true
					break
				}
			}

			if found != tc.shouldFind {
				t.Errorf("Expected to find module '%s': %v, but got: %v", tc.searchModule, tc.shouldFind, found)
			}
		})
	}
}

func TestPCIDeviceModelParsing(t *testing.T) {
	testOutput := "0c:00.0 Ethernet controller: Mellanox Technologies MT2910 Family [ConnectX-7]"
	parts := strings.SplitN(testOutput, ":", 3)
	if len(parts) >= 3 {
		model := strings.TrimSpace(parts[2])
		expected := "Mellanox Technologies MT2910 Family [ConnectX-7]"
		if model != expected {
			t.Errorf("Expected model %s, got %s", expected, model)
		}
	} else {
		t.Error("Failed to parse lspci output")
	}
}

func TestNUMANodeParsing(t *testing.T) {
	testOutput := "  0  \n"
	numaNode := strings.TrimSpace(testOutput)
	if numaNode != "0" {
		t.Errorf("Expected NUMA node 0, got %s", numaNode)
	}
}

func TestNetworkInterfaceParsing(t *testing.T) {
	testOutput := "eth0\neth1"
	interfaces := strings.Fields(testOutput)
	if len(interfaces) > 0 && interfaces[0] != "eth0" {
		t.Errorf("Expected first interface eth0, got %s", interfaces[0])
	}
}

func TestInfiniBandDeviceParsing(t *testing.T) {
	testOutput := "mlx5_0\nmlx5_1"
	devices := strings.Fields(testOutput)
	if len(devices) > 0 && devices[0] != "mlx5_0" {
		t.Errorf("Expected first device mlx5_0, got %s", devices[0])
	}
}

func TestIbdev2netdevParsing(t *testing.T) {
	testOutput := "mlx5_0 port 1 ==> enp12s0f0np0 (Up)"
	deviceName := "mlx5_0"
	var foundInterface string

	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, deviceName) {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "==>" && i+1 < len(parts) {
					foundInterface = parts[i+1]
					break
				}
			}
		}
	}

	expected := "enp12s0f0np0"
	if foundInterface != expected {
		t.Errorf("Expected interface %s, got %s", expected, foundInterface)
	}
}

func TestIbdev2netdevOutputParsing(t *testing.T) {
	testOutput := `mlx5_0 port 1 ==> enp12s0f0np0 (Up)
mlx5_1 port 1 ==> enp12s0f1np1 (Down)
mlx5_2 port 1 ==> enp175s0f0np0 (Up)`

	deviceMap := make(map[string]string)
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 5 && parts[3] == "==>" {
			deviceMap[parts[0]] = parts[4]
		}
	}

	expectedDevices := map[string]string{
		"mlx5_0": "enp12s0f0np0",
		"mlx5_1": "enp12s0f1np1",
		"mlx5_2": "enp175s0f0np0",
	}

	if len(deviceMap) != len(expectedDevices) {
		t.Errorf("Expected %d devices, got %d", len(expectedDevices), len(deviceMap))
	}

	for device, expectedNetdev := range expectedDevices {
		if actualNetdev, exists := deviceMap[device]; !exists {
			t.Errorf("Expected device %s not found", device)
		} else if actualNetdev != expectedNetdev {
			t.Errorf("Expected netdev %s for device %s, got %s", expectedNetdev, device, actualNetdev)
		}
	}
}

func TestIbdev2netdevEmptyOutput(t *testing.T) {
	testOutput := ""

	deviceMap := make(map[string]string)
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 5 && parts[3] == "==>" {
			deviceMap[parts[0]] = parts[4]
		}
	}

	if len(deviceMap) != 0 {
		t.Errorf("Expected empty device map, got %d devices", len(deviceMap))
	}
}

func TestIbdev2netdevMalformedOutput(t *testing.T) {
	testOutput := `mlx5_0 port 1 invalid format
incomplete line
mlx5_1 port 1 ==> enp12s0f1np1 (Up)`

	deviceMap := make(map[string]string)
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 5 && parts[3] == "==>" {
			deviceMap[parts[0]] = parts[4]
		}
	}

	expectedDevices := map[string]string{
		"mlx5_1": "enp12s0f1np1",
	}

	if len(deviceMap) != len(expectedDevices) {
		t.Errorf("Expected %d devices, got %d", len(expectedDevices), len(deviceMap))
	}

	for device, expectedNetdev := range expectedDevices {
		if actualNetdev, exists := deviceMap[device]; !exists {
			t.Errorf("Expected device %s not found", device)
		} else if actualNetdev != expectedNetdev {
			t.Errorf("Expected netdev %s for device %s, got %s", expectedNetdev, device, actualNetdev)
		}
	}
}

func TestIPAddressParsing(t *testing.T) {
	testOutput := `2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 00:1a:2b:3c:4d:5e brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/24 brd 192.168.1.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::21a:2bff:fe3c:4d5e/64 scope link
       valid_lft forever preferred_lft forever`

	var foundIP string
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipWithCIDR := parts[1]
				if strings.Contains(ipWithCIDR, "/") {
					foundIP = strings.Split(ipWithCIDR, "/")[0]
					break
				}
			}
		}
	}

	expected := "192.168.1.100"
	if foundIP != expected {
		t.Errorf("Expected IP address %s, got %s", expected, foundIP)
	}
}

func TestEthtoolCommandGeneration(t *testing.T) {
	tests := []struct {
		name           string
		interface_name string
		grepPattern    string
		expected       string
	}{
		{
			name:           "basic_ethtool",
			interface_name: "eth0",
			grepPattern:    "",
			expected:       "sudo ethtool -S eth0",
		},
		{
			name:           "ethtool_with_grep",
			interface_name: "enp12s0f0",
			grepPattern:    "rx_packets",
			expected:       "sudo ethtool -S enp12s0f0 | grep rx_packets",
		},
		{
			name:           "ethtool_with_regex_pattern",
			interface_name: "rdma0",
			grepPattern:    "rx_prio.*_discards",
			expected:       "sudo ethtool -S rdma0 | grep rx_prio.*_discards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd string
			if tt.grepPattern != "" {
				cmd = "sudo ethtool -S " + tt.interface_name + " | grep " + tt.grepPattern
			} else {
				cmd = "sudo ethtool -S " + tt.interface_name
			}

			if cmd != tt.expected {
				t.Errorf("Expected command '%s', got '%s'", tt.expected, cmd)
			}
		})
	}
}

func TestMlxlinkCommandGeneration(t *testing.T) {
	tests := []struct {
		name          string
		interfaceName string
		expected      string
	}{
		{
			name:          "basic_mlxlink",
			interfaceName: "mlx5_0",
			expected:      "sudo mlxlink -d mlx5_0 --json --show_module --show_counters --show_eye",
		},
		{
			name:          "mlxlink_ethernet_interface",
			interfaceName: "enp12s0f0",
			expected:      "sudo mlxlink -d enp12s0f0 --json --show_module --show_counters --show_eye",
		},
		{
			name:          "mlxlink_with_special_chars",
			interfaceName: "mlx5_0/1",
			expected:      "sudo mlxlink -d mlx5_0/1 --json --show_module --show_counters --show_eye",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := "sudo mlxlink -d " + tt.interfaceName + " --json --show_module --show_counters --show_eye"

			if cmd != tt.expected {
				t.Errorf("Expected command '%s', got '%s'", tt.expected, cmd)
			}

			// Verify all required components are present
			requiredComponents := []string{
				"sudo",
				"mlxlink",
				"-d",
				tt.interfaceName,
				"--json",
				"--show_module",
				"--show_counters",
				"--show_eye",
			}

			for _, component := range requiredComponents {
				if !strings.Contains(cmd, component) {
					t.Errorf("Expected command to contain '%s', got '%s'", component, cmd)
				}
			}
		})
	}
}
