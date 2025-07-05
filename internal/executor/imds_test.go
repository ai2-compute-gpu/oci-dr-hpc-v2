package executor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock IMDS server for testing
func createMockIMDSServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer Oracle" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Route based on path
		switch r.URL.Path {
		case "/opc/v2/instance":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"id": "ocid1.instance.oc1.phx.test123",
				"displayName": "test-instance",
				"hostname": "test-hostname",
				"compartmentId": "ocid1.compartment.oc1..test",
				"region": "phx",
				"canonicalRegionName": "us-phoenix-1",
				"availabilityDomain": "EMIr:PHX-AD-1",
				"ociAdName": "phx-ad-1",
				"faultDomain": "FAULT-DOMAIN-1",
				"image": "ocid1.image.oc1.phx.test456",
				"shape": "BM.GPU.H100.8",
				"state": "Running",
				"timeCreated": 1600381928581,
				"metadata": {
					"ssh_authorized_keys": "example-ssh-key"
				},
				"regionInfo": {
					"realmKey": "oc1",
					"realmDomainComponent": "oraclecloud.com",
					"regionKey": "PHX",
					"regionIdentifier": "us-phoenix-1"
				},
				"definedTags": {},
				"freeformTags": {}
			}`))
		case "/opc/v2/identity":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"fingerprint": "test:fingerprint:12345",
				"tenancyId": "ocid1.tenancy.oc1..test"
			}`))
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
}

func TestNewIMDSClient(t *testing.T) {
	client := NewIMDSClient()
	if client == nil {
		t.Fatal("NewIMDSClient should not return nil")
	}
	if client.baseURL != IMDSBaseURL {
		t.Errorf("Expected baseURL %s, got %s", IMDSBaseURL, client.baseURL)
	}
	if client.httpClient.Timeout != IMDSTimeout {
		t.Errorf("Expected timeout %v, got %v", IMDSTimeout, client.httpClient.Timeout)
	}
}

func TestNewIMDSClientWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewIMDSClientWithTimeout(timeout)
	if client == nil {
		t.Fatal("NewIMDSClientWithTimeout should not return nil")
	}
	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestGetInstanceMetadata(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	metadata, err := client.GetInstanceMetadata()
	if err != nil {
		t.Fatalf("GetInstanceMetadata failed: %v", err)
	}

	if metadata.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape BM.GPU.H100.8, got %s", metadata.Shape)
	}
	if metadata.Region != "phx" {
		t.Errorf("Expected region phx, got %s", metadata.Region)
	}
	if metadata.CanonicalRegionName != "us-phoenix-1" {
		t.Errorf("Expected canonical region us-phoenix-1, got %s", metadata.CanonicalRegionName)
	}
	if metadata.AvailabilityDomain != "EMIr:PHX-AD-1" {
		t.Errorf("Expected AD EMIr:PHX-AD-1, got %s", metadata.AvailabilityDomain)
	}
	if metadata.OciAdName != "phx-ad-1" {
		t.Errorf("Expected OCI AD name phx-ad-1, got %s", metadata.OciAdName)
	}
	if metadata.Hostname != "test-hostname" {
		t.Errorf("Expected hostname test-hostname, got %s", metadata.Hostname)
	}
	if metadata.Image != "ocid1.image.oc1.phx.test456" {
		t.Errorf("Expected image ocid1.image.oc1.phx.test456, got %s", metadata.Image)
	}
	if metadata.TimeCreated != 1600381928581 {
		t.Errorf("Expected timeCreated 1600381928581, got %d", metadata.TimeCreated)
	}
	if metadata.RegionInfo.RealmKey != "oc1" {
		t.Errorf("Expected realm key oc1, got %s", metadata.RegionInfo.RealmKey)
	}
	if metadata.RegionInfo.RegionIdentifier != "us-phoenix-1" {
		t.Errorf("Expected region identifier us-phoenix-1, got %s", metadata.RegionInfo.RegionIdentifier)
	}
}

func TestGetIdentityMetadata(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	metadata, err := client.GetIdentityMetadata()
	if err != nil {
		t.Fatalf("GetIdentityMetadata failed: %v", err)
	}

	if metadata.TenancyID != "ocid1.tenancy.oc1..test" {
		t.Errorf("Expected tenancy ocid1.tenancy.oc1..test, got %s", metadata.TenancyID)
	}
	if metadata.Fingerprint != "test:fingerprint:12345" {
		t.Errorf("Expected fingerprint test:fingerprint:12345, got %s", metadata.Fingerprint)
	}
}

func TestGetShape(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	shape, err := client.GetShape()
	if err != nil {
		t.Fatalf("GetShape failed: %v", err)
	}

	if shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape BM.GPU.H100.8, got %s", shape)
	}
}

func TestGetRegion(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	region, err := client.GetRegion()
	if err != nil {
		t.Fatalf("GetRegion failed: %v", err)
	}

	if region != "phx" {
		t.Errorf("Expected region phx, got %s", region)
	}
}

func TestGetAvailabilityDomain(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	ad, err := client.GetAvailabilityDomain()
	if err != nil {
		t.Fatalf("GetAvailabilityDomain failed: %v", err)
	}

	if ad != "EMIr:PHX-AD-1" {
		t.Errorf("Expected AD EMIr:PHX-AD-1, got %s", ad)
	}
}

func TestMakeRequestUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return unauthorized for this test
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	_, err := client.makeRequest("instance")
	if err == nil {
		t.Fatal("Expected error for unauthorized request")
	}
}

func TestMakeRequestNotFound(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	_, err := client.makeRequest("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent endpoint")
	}
}

func TestIsRunningOnOCI(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	isOCI := client.IsRunningOnOCI()
	if !isOCI {
		t.Error("Expected IsRunningOnOCI to return true with mock server")
	}

	// Test with unreachable server
	client.baseURL = "http://localhost:99999/opc/v2"
	isOCI = client.IsRunningOnOCI()
	if isOCI {
		t.Error("Expected IsRunningOnOCI to return false with unreachable server")
	}
}

func TestGetInstanceInfo(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	info, err := client.GetInstanceInfo()
	if err != nil {
		t.Fatalf("GetInstanceInfo failed: %v", err)
	}

	if info["instance"] == nil {
		t.Error("Expected instance metadata in info")
	}
	if info["identity"] == nil {
		t.Error("Expected identity metadata in info")
	}

	instanceMeta, ok := info["instance"].(*InstanceMetadata)
	if !ok {
		t.Error("Expected instance metadata to be *InstanceMetadata")
	} else if instanceMeta.Shape != "BM.GPU.H100.8" {
		t.Errorf("Expected shape BM.GPU.H100.8, got %s", instanceMeta.Shape)
	}
}

func TestGetRawMetadata(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	data, err := client.GetRawMetadata("instance")
	if err != nil {
		t.Fatalf("GetRawMetadata failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty raw metadata")
	}
}

func TestGetHostname(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	hostname, err := client.GetHostname()
	if err != nil {
		t.Fatalf("GetHostname failed: %v", err)
	}

	if hostname != "test-hostname" {
		t.Errorf("Expected hostname test-hostname, got %s", hostname)
	}
}

func TestGetCanonicalRegionName(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	regionName, err := client.GetCanonicalRegionName()
	if err != nil {
		t.Fatalf("GetCanonicalRegionName failed: %v", err)
	}

	if regionName != "us-phoenix-1" {
		t.Errorf("Expected canonical region us-phoenix-1, got %s", regionName)
	}
}

func TestGetOciAdName(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	adName, err := client.GetOciAdName()
	if err != nil {
		t.Fatalf("GetOciAdName failed: %v", err)
	}

	if adName != "phx-ad-1" {
		t.Errorf("Expected OCI AD name phx-ad-1, got %s", adName)
	}
}

func TestGetImageOCID(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	imageOCID, err := client.GetImageOCID()
	if err != nil {
		t.Fatalf("GetImageOCID failed: %v", err)
	}

	if imageOCID != "ocid1.image.oc1.phx.test456" {
		t.Errorf("Expected image OCID ocid1.image.oc1.phx.test456, got %s", imageOCID)
	}
}

func TestGetInstanceOCID(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	instanceOCID, err := client.GetInstanceOCID()
	if err != nil {
		t.Fatalf("GetInstanceOCID failed: %v", err)
	}

	if instanceOCID != "ocid1.instance.oc1.phx.test123" {
		t.Errorf("Expected instance OCID ocid1.instance.oc1.phx.test123, got %s", instanceOCID)
	}
}

func TestGetInstanceState(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	state, err := client.GetInstanceState()
	if err != nil {
		t.Fatalf("GetInstanceState failed: %v", err)
	}

	if state != "Running" {
		t.Errorf("Expected state Running, got %s", state)
	}
}

func TestGetRegionInfo(t *testing.T) {
	server := createMockIMDSServer()
	defer server.Close()

	client := &IMDSClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL + "/opc/v2",
	}

	regionInfo, err := client.GetRegionInfo()
	if err != nil {
		t.Fatalf("GetRegionInfo failed: %v", err)
	}

	if regionInfo.RealmKey != "oc1" {
		t.Errorf("Expected realm key oc1, got %s", regionInfo.RealmKey)
	}
	if regionInfo.RegionKey != "PHX" {
		t.Errorf("Expected region key PHX, got %s", regionInfo.RegionKey)
	}
	if regionInfo.RegionIdentifier != "us-phoenix-1" {
		t.Errorf("Expected region identifier us-phoenix-1, got %s", regionInfo.RegionIdentifier)
	}
	if regionInfo.RealmDomainComponent != "oraclecloud.com" {
		t.Errorf("Expected realm domain oraclecloud.com, got %s", regionInfo.RealmDomainComponent)
	}
}