package executor

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInstanceMetadataStruct(t *testing.T) {
	metadata := &InstanceMetadata{
		ID:                  "ocid1.instance.oc1.phx.test123",
		DisplayName:         "test-instance",
		Hostname:            "test-hostname",
		CompartmentID:       "ocid1.compartment.oc1..test",
		TenantID:            "ocid1.tenancy.oc1..test",
		Region:              "phx",
		CanonicalRegionName: "us-phoenix-1",
		AvailabilityDomain:  "EMIr:PHX-AD-1",
		OciAdName:           "phx-ad-1",
		FaultDomain:         "FAULT-DOMAIN-1",
		Image:               "ocid1.image.oc1.phx.test456",
		Shape:               "BM.GPU.H100.8",
		State:               "Running",
		TimeCreated:         1600381928581,
		RegionInfo: RegionInfo{
			RealmKey:             "oc1",
			RealmDomainComponent: "oraclecloud.com",
			RegionKey:            "PHX",
			RegionIdentifier:     "us-phoenix-1",
		},
	}

	if metadata.ID != "ocid1.instance.oc1.phx.test123" {
		t.Errorf("Expected ID 'ocid1.instance.oc1.phx.test123', got '%s'", metadata.ID)
	}
	if metadata.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected Shape 'BM.GPU.H100.8', got '%s'", metadata.Shape)
	}
	if metadata.Region != "phx" {
		t.Errorf("Expected Region 'phx', got '%s'", metadata.Region)
	}
	if metadata.CanonicalRegionName != "us-phoenix-1" {
		t.Errorf("Expected CanonicalRegionName 'us-phoenix-1', got '%s'", metadata.CanonicalRegionName)
	}
	if metadata.RegionInfo.RealmKey != "oc1" {
		t.Errorf("Expected RealmKey 'oc1', got '%s'", metadata.RegionInfo.RealmKey)
	}
}

func TestIdentityMetadataStruct(t *testing.T) {
	metadata := &IdentityMetadata{
		Fingerprint: "test:fingerprint:12345",
		TenancyID:   "ocid1.tenancy.oc1..test",
	}

	if metadata.Fingerprint != "test:fingerprint:12345" {
		t.Errorf("Expected Fingerprint 'test:fingerprint:12345', got '%s'", metadata.Fingerprint)
	}
	if metadata.TenancyID != "ocid1.tenancy.oc1..test" {
		t.Errorf("Expected TenancyID 'ocid1.tenancy.oc1..test', got '%s'", metadata.TenancyID)
	}
}

func TestHostMetadataStruct(t *testing.T) {
	metadata := &HostMetadata{
		BuildingID:     "building:1725b321355f0314955d9e8d68cf2c54bc99db9ce21fecece62db989e763ac71",
		ID:             "c108fca4ead3550cf7bbf0be7c0dce30b01b6b772a48f9f8e13346891a57f8a2",
		NetworkBlockID: "922fd61aa1af7d80d4edb08bcf09d6ae6f0d0152bb99843885fcd732b22716d9",
		RackID:         "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a",
	}

	expectedBuildingID := "building:1725b321355f0314955d9e8d68cf2c54bc99db9ce21fecece62db989e763ac71"
	expectedHostID := "c108fca4ead3550cf7bbf0be7c0dce30b01b6b772a48f9f8e13346891a57f8a2"
	expectedNetworkBlockID := "922fd61aa1af7d80d4edb08bcf09d6ae6f0d0152bb99843885fcd732b22716d9"
	expectedRackID := "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a"

	if metadata.BuildingID != expectedBuildingID {
		t.Errorf("Expected BuildingID '%s', got '%s'", expectedBuildingID, metadata.BuildingID)
	}
	if metadata.ID != expectedHostID {
		t.Errorf("Expected ID '%s', got '%s'", expectedHostID, metadata.ID)
	}
	if metadata.NetworkBlockID != expectedNetworkBlockID {
		t.Errorf("Expected NetworkBlockID '%s', got '%s'", expectedNetworkBlockID, metadata.NetworkBlockID)
	}
	if metadata.RackID != expectedRackID {
		t.Errorf("Expected RackID '%s', got '%s'", expectedRackID, metadata.RackID)
	}
}

func TestVnicMetadataStruct(t *testing.T) {
	metadata := &VnicMetadata{
		MacAddr:         "02:00:17:05:d1:db",
		NicIndex:        0,
		PrivateIP:       "10.0.0.2",
		SubnetCidrBlock: "10.0.0.0/24",
		VirtualRouterIP: "10.0.0.1",
		VlanTag:         0,
		VnicID:          "ocid1.vnic.oc1.phx.test",
	}

	if metadata.MacAddr != "02:00:17:05:d1:db" {
		t.Errorf("Expected MacAddr '02:00:17:05:d1:db', got '%s'", metadata.MacAddr)
	}
	if metadata.NicIndex != 0 {
		t.Errorf("Expected NicIndex 0, got %d", metadata.NicIndex)
	}
	if metadata.PrivateIP != "10.0.0.2" {
		t.Errorf("Expected PrivateIP '10.0.0.2', got '%s'", metadata.PrivateIP)
	}
	if metadata.SubnetCidrBlock != "10.0.0.0/24" {
		t.Errorf("Expected SubnetCidrBlock '10.0.0.0/24', got '%s'", metadata.SubnetCidrBlock)
	}
}

func TestRegionInfoStruct(t *testing.T) {
	regionInfo := &RegionInfo{
		RealmKey:             "oc1",
		RealmDomainComponent: "oraclecloud.com",
		RegionKey:            "PHX",
		RegionIdentifier:     "us-phoenix-1",
	}

	if regionInfo.RealmKey != "oc1" {
		t.Errorf("Expected RealmKey 'oc1', got '%s'", regionInfo.RealmKey)
	}
	if regionInfo.RealmDomainComponent != "oraclecloud.com" {
		t.Errorf("Expected RealmDomainComponent 'oraclecloud.com', got '%s'", regionInfo.RealmDomainComponent)
	}
	if regionInfo.RegionKey != "PHX" {
		t.Errorf("Expected RegionKey 'PHX', got '%s'", regionInfo.RegionKey)
	}
	if regionInfo.RegionIdentifier != "us-phoenix-1" {
		t.Errorf("Expected RegionIdentifier 'us-phoenix-1', got '%s'", regionInfo.RegionIdentifier)
	}
}

func TestShapeConfigStruct(t *testing.T) {
	shapeConfig := &ShapeConfig{
		MaxVnicAttachments:        8,
		MemoryInGBs:               1024.0,
		NetworkingBandwidthInGbps: 100.0,
		Ocpus:                     128.0,
	}

	if shapeConfig.MaxVnicAttachments != 8 {
		t.Errorf("Expected MaxVnicAttachments 8, got %d", shapeConfig.MaxVnicAttachments)
	}
	if shapeConfig.MemoryInGBs != 1024.0 {
		t.Errorf("Expected MemoryInGBs 1024.0, got %f", shapeConfig.MemoryInGBs)
	}
	if shapeConfig.NetworkingBandwidthInGbps != 100.0 {
		t.Errorf("Expected NetworkingBandwidthInGbps 100.0, got %f", shapeConfig.NetworkingBandwidthInGbps)
	}
	if shapeConfig.Ocpus != 128.0 {
		t.Errorf("Expected Ocpus 128.0, got %f", shapeConfig.Ocpus)
	}
}

func TestAgentConfigStruct(t *testing.T) {
	agentConfig := &AgentConfig{
		AllPluginsDisabled: false,
		ManagementDisabled: false,
		MonitoringDisabled: true,
		PluginsConfig: []PluginConfig{
			{
				Name:         "Oracle Java Management Service",
				DesiredState: "ENABLED",
			},
		},
	}

	if agentConfig.AllPluginsDisabled != false {
		t.Errorf("Expected AllPluginsDisabled false, got %v", agentConfig.AllPluginsDisabled)
	}
	if agentConfig.MonitoringDisabled != true {
		t.Errorf("Expected MonitoringDisabled true, got %v", agentConfig.MonitoringDisabled)
	}
	if len(agentConfig.PluginsConfig) != 1 {
		t.Errorf("Expected 1 plugin config, got %d", len(agentConfig.PluginsConfig))
	}
	if agentConfig.PluginsConfig[0].Name != "Oracle Java Management Service" {
		t.Errorf("Expected plugin name 'Oracle Java Management Service', got '%s'", agentConfig.PluginsConfig[0].Name)
	}
}

func TestNewIMDSClientConstants(t *testing.T) {
	if IMDSBaseURL != "http://169.254.169.254/opc/v2" {
		t.Errorf("Expected IMDSBaseURL 'http://169.254.169.254/opc/v2', got '%s'", IMDSBaseURL)
	}
	if IMDSTimeout != 5*time.Second {
		t.Errorf("Expected IMDSTimeout 5s, got %v", IMDSTimeout)
	}
}

func TestInstanceMetadataJSONSerialization(t *testing.T) {
	metadata := &InstanceMetadata{
		ID:                  "ocid1.instance.oc1.phx.test123",
		Shape:               "BM.GPU.H100.8",
		Region:              "phx",
		CanonicalRegionName: "us-phoenix-1",
		RegionInfo: RegionInfo{
			RealmKey:         "oc1",
			RegionIdentifier: "us-phoenix-1",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled InstanceMetadata
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if unmarshaled.ID != metadata.ID {
		t.Errorf("Expected ID '%s' after unmarshal, got '%s'", metadata.ID, unmarshaled.ID)
	}
	if unmarshaled.Shape != metadata.Shape {
		t.Errorf("Expected Shape '%s' after unmarshal, got '%s'", metadata.Shape, unmarshaled.Shape)
	}
	if unmarshaled.RegionInfo.RealmKey != metadata.RegionInfo.RealmKey {
		t.Errorf("Expected RealmKey '%s' after unmarshal, got '%s'", metadata.RegionInfo.RealmKey, unmarshaled.RegionInfo.RealmKey)
	}
}

func TestInstanceMetadataJSONDeserialization(t *testing.T) {
	jsonData := `{
		"id": "ocid1.instance.oc1.phx.test123",
		"displayName": "test-instance",
		"hostname": "test-hostname",
		"compartmentId": "ocid1.compartment.oc1..test",
		"tenantId": "ocid1.tenancy.oc1..test",
		"region": "phx",
		"canonicalRegionName": "us-phoenix-1",
		"availabilityDomain": "EMIr:PHX-AD-1",
		"ociAdName": "phx-ad-1",
		"shape": "BM.GPU.H100.8",
		"state": "Running",
		"timeCreated": 1600381928581,
		"regionInfo": {
			"realmKey": "oc1",
			"realmDomainComponent": "oraclecloud.com",
			"regionKey": "PHX",
			"regionIdentifier": "us-phoenix-1"
		}
	}`

	var metadata InstanceMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if metadata.ID != "ocid1.instance.oc1.phx.test123" {
		t.Errorf("Expected ID 'ocid1.instance.oc1.phx.test123', got '%s'", metadata.ID)
	}
	if metadata.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected Shape 'BM.GPU.H100.8', got '%s'", metadata.Shape)
	}
	if metadata.Region != "phx" {
		t.Errorf("Expected Region 'phx', got '%s'", metadata.Region)
	}
	if metadata.TimeCreated != 1600381928581 {
		t.Errorf("Expected TimeCreated 1600381928581, got %d", metadata.TimeCreated)
	}
	if metadata.RegionInfo.RealmKey != "oc1" {
		t.Errorf("Expected RealmKey 'oc1', got '%s'", metadata.RegionInfo.RealmKey)
	}
}

func TestHostMetadataJSONDeserialization(t *testing.T) {
	jsonData := `{
		"buildingId": "building:1725b321355f0314955d9e8d68cf2c54bc99db9ce21fecece62db989e763ac71",
		"id": "c108fca4ead3550cf7bbf0be7c0dce30b01b6b772a48f9f8e13346891a57f8a2",
		"networkBlockId": "922fd61aa1af7d80d4edb08bcf09d6ae6f0d0152bb99843885fcd732b22716d9",
		"rackId": "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a"
	}`

	var metadata HostMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	expectedBuildingID := "building:1725b321355f0314955d9e8d68cf2c54bc99db9ce21fecece62db989e763ac71"
	expectedHostID := "c108fca4ead3550cf7bbf0be7c0dce30b01b6b772a48f9f8e13346891a57f8a2"
	expectedNetworkBlockID := "922fd61aa1af7d80d4edb08bcf09d6ae6f0d0152bb99843885fcd732b22716d9"
	expectedRackID := "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a"

	if metadata.BuildingID != expectedBuildingID {
		t.Errorf("Expected BuildingID '%s', got '%s'", expectedBuildingID, metadata.BuildingID)
	}
	if metadata.ID != expectedHostID {
		t.Errorf("Expected ID '%s', got '%s'", expectedHostID, metadata.ID)
	}
	if metadata.NetworkBlockID != expectedNetworkBlockID {
		t.Errorf("Expected NetworkBlockID '%s', got '%s'", expectedNetworkBlockID, metadata.NetworkBlockID)
	}
	if metadata.RackID != expectedRackID {
		t.Errorf("Expected RackID '%s', got '%s'", expectedRackID, metadata.RackID)
	}
}

func TestVnicMetadataJSONDeserialization(t *testing.T) {
	jsonData := `[{
		"macAddr": "02:00:17:05:d1:db",
		"nicIndex": 0,
		"privateIp": "10.0.0.2",
		"subnetCidrBlock": "10.0.0.0/24",
		"virtualRouterIp": "10.0.0.1",
		"vlanTag": 0,
		"vnicId": "ocid1.vnic.oc1.phx.test"
	}]`

	var vnics []VnicMetadata
	err := json.Unmarshal([]byte(jsonData), &vnics)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if len(vnics) != 1 {
		t.Fatalf("Expected 1 VNIC, got %d", len(vnics))
	}

	vnic := vnics[0]
	if vnic.MacAddr != "02:00:17:05:d1:db" {
		t.Errorf("Expected MacAddr '02:00:17:05:d1:db', got '%s'", vnic.MacAddr)
	}
	if vnic.NicIndex != 0 {
		t.Errorf("Expected NicIndex 0, got %d", vnic.NicIndex)
	}
	if vnic.PrivateIP != "10.0.0.2" {
		t.Errorf("Expected PrivateIP '10.0.0.2', got '%s'", vnic.PrivateIP)
	}
}

func TestIdentityMetadataJSONDeserialization(t *testing.T) {
	jsonData := `{
		"fingerprint": "test:fingerprint:12345",
		"tenancyId": "ocid1.tenancy.oc1..test"
	}`

	var metadata IdentityMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if metadata.Fingerprint != "test:fingerprint:12345" {
		t.Errorf("Expected Fingerprint 'test:fingerprint:12345', got '%s'", metadata.Fingerprint)
	}
	if metadata.TenancyID != "ocid1.tenancy.oc1..test" {
		t.Errorf("Expected TenancyID 'ocid1.tenancy.oc1..test', got '%s'", metadata.TenancyID)
	}
}

func TestShapeConfigJSONDeserialization(t *testing.T) {
	jsonData := `{
		"maxVnicAttachments": 8,
		"memoryInGBs": 1024.0,
		"networkingBandwidthInGbps": 100.0,
		"ocpus": 128.0
	}`

	var shapeConfig ShapeConfig
	err := json.Unmarshal([]byte(jsonData), &shapeConfig)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if shapeConfig.MaxVnicAttachments != 8 {
		t.Errorf("Expected MaxVnicAttachments 8, got %d", shapeConfig.MaxVnicAttachments)
	}
	if shapeConfig.MemoryInGBs != 1024.0 {
		t.Errorf("Expected MemoryInGBs 1024.0, got %f", shapeConfig.MemoryInGBs)
	}
	if shapeConfig.NetworkingBandwidthInGbps != 100.0 {
		t.Errorf("Expected NetworkingBandwidthInGbps 100.0, got %f", shapeConfig.NetworkingBandwidthInGbps)
	}
	if shapeConfig.Ocpus != 128.0 {
		t.Errorf("Expected Ocpus 128.0, got %f", shapeConfig.Ocpus)
	}
}

func TestPluginConfigStruct(t *testing.T) {
	plugin := PluginConfig{
		Name:         "Oracle Java Management Service",
		DesiredState: "ENABLED",
	}

	if plugin.Name != "Oracle Java Management Service" {
		t.Errorf("Expected Name 'Oracle Java Management Service', got '%s'", plugin.Name)
	}
	if plugin.DesiredState != "ENABLED" {
		t.Errorf("Expected DesiredState 'ENABLED', got '%s'", plugin.DesiredState)
	}
}

func TestComplexInstanceMetadataStruct(t *testing.T) {
	metadata := &InstanceMetadata{
		ID:    "ocid1.instance.oc1.phx.test123",
		Shape: "BM.GPU.H100.8",
		ShapeConfig: &ShapeConfig{
			MaxVnicAttachments:        8,
			MemoryInGBs:               1024.0,
			NetworkingBandwidthInGbps: 100.0,
			Ocpus:                     128.0,
		},
		AgentConfig: &AgentConfig{
			AllPluginsDisabled: false,
			ManagementDisabled: false,
			MonitoringDisabled: true,
		},
		Metadata: map[string]interface{}{
			"ssh_authorized_keys": "ssh-rsa AAAAB3NzaC1yc2E...",
			"user_data":           "#!/bin/bash\necho 'Hello World'",
		},
		DefinedTags: map[string]interface{}{
			"Environment": map[string]interface{}{
				"Type": "Production",
			},
		},
		FreeformTags: map[string]interface{}{
			"Project": "HPC-Cluster",
			"Owner":   "TeamA",
		},
	}

	// Test shape config
	if metadata.ShapeConfig.Ocpus != 128.0 {
		t.Errorf("Expected ShapeConfig.Ocpus 128.0, got %f", metadata.ShapeConfig.Ocpus)
	}

	// Test agent config
	if metadata.AgentConfig.MonitoringDisabled != true {
		t.Errorf("Expected AgentConfig.MonitoringDisabled true, got %v", metadata.AgentConfig.MonitoringDisabled)
	}

	// Test metadata map
	if metadata.Metadata["ssh_authorized_keys"] != "ssh-rsa AAAAB3NzaC1yc2E..." {
		t.Errorf("Expected ssh_authorized_keys to be set")
	}

	// Test freeform tags
	if metadata.FreeformTags["Project"] != "HPC-Cluster" {
		t.Errorf("Expected Project tag 'HPC-Cluster', got '%v'", metadata.FreeformTags["Project"])
	}
}
