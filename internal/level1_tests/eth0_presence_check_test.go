package level1_tests

import (
	"strings"
	"testing"
)

// Helper function to create test IP addr output
func createTestIPAddrOutput(includeEth0 bool) string {
	output := "1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN\n"
	output += "    inet 127.0.0.1/8 scope host lo\n"

	if includeEth0 {
		output += "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP\n"
		output += "    inet 10.0.0.5/24 brd 10.0.0.255 scope global eth0\n"
	}

	output += "3: eth1: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN\n"
	return output
}

// Test parsing logic that mimics checkEth0Present
func parseIPAddrForTesting(ipOutput string, interfaceName string) bool {
	output := strings.TrimSpace(ipOutput)
	if output == "" {
		return false
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, interfaceName) {
			return true
		}
	}

	return false
}

func TestEth0PresenceCheckTestConfig(t *testing.T) {
	config := &Eth0PresenceCheckTestConfig{
		IsEnabled: true,
	}

	if !config.IsEnabled {
		t.Errorf("Expected IsEnabled=true, got %v", config.IsEnabled)
	}
}

func TestEth0ParsingLogic(t *testing.T) {
	tests := []struct {
		name        string
		ipOutput    string
		expectFound bool
		description string
	}{
		{
			name:        "eth0_present",
			ipOutput:    createTestIPAddrOutput(true),
			expectFound: true,
			description: "Should find eth0 when present",
		},
		{
			name:        "eth0_not_present",
			ipOutput:    createTestIPAddrOutput(false),
			expectFound: false,
			description: "Should not find eth0 when not present",
		},
		{
			name:        "empty_output",
			ipOutput:    "",
			expectFound: false,
			description: "Should handle empty output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := parseIPAddrForTesting(tt.ipOutput, "eth0")
			if found != tt.expectFound {
				t.Errorf("%s: Expected found=%v, got %v", tt.description, tt.expectFound, found)
			}
		})
	}
}

func TestPrintEth0PresenceCheck(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintEth0PresenceCheck panicked: %v", r)
		}
	}()

	PrintEth0PresenceCheck()
	t.Log("PrintEth0PresenceCheck completed successfully")
}