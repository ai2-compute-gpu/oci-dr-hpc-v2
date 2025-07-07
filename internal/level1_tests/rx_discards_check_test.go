package level1_tests

import (
	"encoding/json"
	"testing"
)

func TestGetRXDiscardsConfig(t *testing.T) {
	config := getRXDiscardsConfig()

	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	if len(config.Interfaces) != 16 {
		t.Errorf("Expected 16 interfaces, got %d", len(config.Interfaces))
	}

	if config.Interfaces[0] != "rdma0" {
		t.Errorf("Expected first interface 'rdma0', got '%s'", config.Interfaces[0])
	}

	if config.Interfaces[15] != "rdma15" {
		t.Errorf("Expected last interface 'rdma15', got '%s'", config.Interfaces[15])
	}
}

func TestParseRXDiscardsResults(t *testing.T) {
	// Test empty results (should fail)
	result := parseRXDiscardsResults("rdma0", []string{}, 100.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL for empty results")
	}
	if result.RXDiscards.Device != "rdma0" {
		t.Errorf("Expected device 'rdma0', got '%s'", result.RXDiscards.Device)
	}

	// Test valid results below threshold (should pass)
	result = parseRXDiscardsResults("rdma1", []string{"rx_prio0_discards: 50"}, 100.0)
	if result.RXDiscards.Status != "PASS" {
		t.Error("Expected PASS for results below threshold")
	}

	// Test results above threshold (should fail)
	result = parseRXDiscardsResults("rdma2", []string{"rx_prio0_discards: 150"}, 100.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL for results above threshold")
	}

	// Test invalid format (should fail)
	result = parseRXDiscardsResults("rdma3", []string{"rx_prio0_discards: abc"}, 100.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL for invalid number format")
	}

	// Test exact threshold value (should pass)
	result = parseRXDiscardsResults("rdma4", []string{"rx_prio0_discards: 100"}, 100.0)
	if result.RXDiscards.Status != "PASS" {
		t.Error("Expected PASS for exact threshold value")
	}

	// Test multiple lines with one exceeding threshold
	result = parseRXDiscardsResults("rdma5", []string{"rx_prio0_discards: 50", "rx_prio1_discards: 150"}, 100.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL when any line exceeds threshold")
	}
}

func TestRxDiscardTestConfig(t *testing.T) {
	config := &RxDiscardTestConfig{
		Threshold: 100.5,
		IsEnabled: true,
		Shape:     "BM.GPU.H100.8",
	}

	if config.Threshold != 100.5 {
		t.Errorf("Expected threshold 100.5, got %f", config.Threshold)
	}

	if !config.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}

	if config.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape 'BM.GPU.H100.8', got '%s'", config.Shape)
	}
}

func TestRXDiscardsResultJSON(t *testing.T) {
	result := RXDiscardsResult{
		RXDiscards: RXDiscardsDevice{
			Device: "rdma0",
			Status: "PASS",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled RXDiscardsResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	if unmarshaled.RXDiscards.Device != result.RXDiscards.Device {
		t.Errorf("Expected device '%s', got '%s'", result.RXDiscards.Device, unmarshaled.RXDiscards.Device)
	}

	if unmarshaled.RXDiscards.Status != result.RXDiscards.Status {
		t.Errorf("Expected status '%s', got '%s'", result.RXDiscards.Status, unmarshaled.RXDiscards.Status)
	}
}

func TestParseRXDiscardsResultsEdgeCases(t *testing.T) {
	// Test empty lines mixed with valid data
	result := parseRXDiscardsResults("rdma0", []string{"", "rx_prio0_discards: 50", ""}, 100.0)
	if result.RXDiscards.Status != "PASS" {
		t.Error("Expected PASS with empty lines in data")
	}

	// Test malformed line without colon
	result = parseRXDiscardsResults("rdma1", []string{"invalid line format"}, 100.0)
	if result.RXDiscards.Status != "PASS" {
		t.Error("Expected PASS for malformed line without colon")
	}

	// Test line with multiple colons
	result = parseRXDiscardsResults("rdma2", []string{"rx_prio0_discards: test: 50"}, 100.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL for invalid number after multiple colons")
	}

	// Test zero threshold
	result = parseRXDiscardsResults("rdma3", []string{"rx_prio0_discards: 1"}, 0.0)
	if result.RXDiscards.Status != "FAIL" {
		t.Error("Expected FAIL when value exceeds zero threshold")
	}

	// Test negative numbers
	result = parseRXDiscardsResults("rdma4", []string{"rx_prio0_discards: -5"}, 100.0)
	if result.RXDiscards.Status != "PASS" {
		t.Error("Expected PASS for negative number below threshold")
	}
}

func TestRXDiscardsStructures(t *testing.T) {
	// Test RXDiscardsDevice
	device := RXDiscardsDevice{
		Device: "rdma10",
		Status: "FAIL",
	}

	if device.Device != "rdma10" {
		t.Errorf("Expected device 'rdma10', got '%s'", device.Device)
	}

	if device.Status != "FAIL" {
		t.Errorf("Expected status 'FAIL', got '%s'", device.Status)
	}

	// Test RXDiscardsResult
	result := RXDiscardsResult{
		RXDiscards: device,
	}

	if result.RXDiscards.Device != "rdma10" {
		t.Errorf("Expected nested device 'rdma10', got '%s'", result.RXDiscards.Device)
	}

	if result.RXDiscards.Status != "FAIL" {
		t.Errorf("Expected nested status 'FAIL', got '%s'", result.RXDiscards.Status)
	}
}
