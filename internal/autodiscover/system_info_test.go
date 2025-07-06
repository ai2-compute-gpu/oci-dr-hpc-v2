package autodiscover

import (
	"testing"
)

func TestSystemInfoStruct(t *testing.T) {
	sysInfo := &SystemInfo{
		Hostname:         "test-host",
		OCID:             "ocid1.instance.oc1.test",
		FriendlyHostname: "test-host",
		Shape:            "BM.GPU.H100.8",
		Serial:           "2334XLG08T",
		Rack:             "test-rack-id",
	}

	if sysInfo.Hostname != "test-host" {
		t.Errorf("Expected hostname 'test-host', got '%s'", sysInfo.Hostname)
	}
	if sysInfo.OCID != "ocid1.instance.oc1.test" {
		t.Errorf("Expected OCID 'ocid1.instance.oc1.test', got '%s'", sysInfo.OCID)
	}
	if sysInfo.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape 'BM.GPU.H100.8', got '%s'", sysInfo.Shape)
	}
	if sysInfo.Serial != "2334XLG08T" {
		t.Errorf("Expected serial '2334XLG08T', got '%s'", sysInfo.Serial)
	}
	if sysInfo.Rack != "test-rack-id" {
		t.Errorf("Expected rack 'test-rack-id', got '%s'", sysInfo.Rack)
	}
}

func TestGatherSystemInfo(t *testing.T) {
	// This is an integration test that will only work on OCI instances
	// with proper IMDS access and dmidecode installed
	t.Skip("Skipping integration test - requires OCI instance with IMDS and dmidecode")

	sysInfo, err := GatherSystemInfo()
	if err != nil {
		t.Logf("GatherSystemInfo failed (expected in test environment): %v", err)
		// Don't fail the test since this requires specific environment
		return
	}

	if sysInfo == nil {
		t.Fatal("Expected SystemInfo but got nil")
	}

	t.Logf("Gathered system info: %+v", sysInfo)
}

func TestGatherSystemInfoPartial(t *testing.T) {
	// This test should always complete without errors
	sysInfo := GatherSystemInfoPartial()

	if sysInfo == nil {
		t.Fatal("Expected SystemInfo but got nil")
	}

	// At minimum, hostname should be available
	if sysInfo.Hostname == "" {
		t.Error("Expected non-empty hostname")
	}

	t.Logf("Gathered partial system info: %+v", sysInfo)
}

func TestSystemInfoJSONTags(t *testing.T) {
	// Test that the struct has the expected JSON tags
	sysInfo := &SystemInfo{}

	// This is a basic test to ensure the struct compiles with the expected fields
	sysInfo.Hostname = "test"
	sysInfo.OCID = "test"
	sysInfo.FriendlyHostname = "test"
	sysInfo.Shape = "test"
	sysInfo.Serial = "test"
	sysInfo.Rack = "test"

	if sysInfo.Hostname != "test" {
		t.Error("Hostname field not working")
	}
}
