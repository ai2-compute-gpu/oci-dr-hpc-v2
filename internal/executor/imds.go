package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

const (
	// IMDSBaseURL is the base URL for OCI Instance Metadata Service
	IMDSBaseURL = "http://169.254.169.254/opc/v2"
	// IMDSTimeout is the default timeout for IMDS requests
	IMDSTimeout = 5 * time.Second
)

// IMDSClient represents a client for OCI Instance Metadata Service
type IMDSClient struct {
	httpClient *http.Client
	baseURL    string
}

// InstanceMetadata represents the instance metadata structure according to OCI IMDS v2 spec
type InstanceMetadata struct {
	ID                  string                 `json:"id"`
	DisplayName         string                 `json:"displayName"`
	Hostname            string                 `json:"hostname"`
	CompartmentID       string                 `json:"compartmentId"`
	TenantID            string                 `json:"tenantId"`
	Region              string                 `json:"region"`
	CanonicalRegionName string                 `json:"canonicalRegionName"`
	AvailabilityDomain  string                 `json:"availabilityDomain"`
	OciAdName           string                 `json:"ociAdName"`
	FaultDomain         string                 `json:"faultDomain"`
	Image               string                 `json:"image"`
	Shape               string                 `json:"shape"`
	ShapeConfig         *ShapeConfig           `json:"shapeConfig,omitempty"`
	State               string                 `json:"state"`
	TimeCreated         int64                  `json:"timeCreated"` // UNIX timestamp in milliseconds
	Metadata            map[string]interface{} `json:"metadata"`
	AgentConfig         *AgentConfig           `json:"agentConfig,omitempty"`
	RegionInfo          RegionInfo             `json:"regionInfo"`
	DefinedTags         map[string]interface{} `json:"definedTags,omitempty"`
	FreeformTags        map[string]interface{} `json:"freeformTags,omitempty"`
}

// RegionInfo represents the region information structure
type RegionInfo struct {
	RealmKey             string `json:"realmKey"`
	RealmDomainComponent string `json:"realmDomainComponent"`
	RegionKey            string `json:"regionKey"`
	RegionIdentifier     string `json:"regionIdentifier"`
}

// ShapeConfig represents the shape configuration structure
type ShapeConfig struct {
	MaxVnicAttachments        int     `json:"maxVnicAttachments"`
	MemoryInGBs               float64 `json:"memoryInGBs"`
	NetworkingBandwidthInGbps float64 `json:"networkingBandwidthInGbps"`
	Ocpus                     float64 `json:"ocpus"`
}

// AgentConfig represents the agent configuration structure
type AgentConfig struct {
	AllPluginsDisabled bool           `json:"allPluginsDisabled"`
	ManagementDisabled bool           `json:"managementDisabled"`
	MonitoringDisabled bool           `json:"monitoringDisabled"`
	PluginsConfig      []PluginConfig `json:"pluginsConfig,omitempty"`
}

// PluginConfig represents individual plugin configuration
type PluginConfig struct {
	Name         string `json:"name"`
	DesiredState string `json:"desiredState"`
}

// VnicMetadata represents VNIC metadata structure
type VnicMetadata struct {
	MacAddr         string `json:"macAddr"`
	NicIndex        int    `json:"nicIndex"`
	PrivateIP       string `json:"privateIp"`
	SubnetCidrBlock string `json:"subnetCidrBlock"`
	VirtualRouterIP string `json:"virtualRouterIp"`
	VlanTag         int    `json:"vlanTag"`
	VnicID          string `json:"vnicId"`
}

// IdentityMetadata represents the identity metadata structure
type IdentityMetadata struct {
	CertPem         string `json:"cert.pem"`
	IntermediatePem string `json:"intermediate.pem"`
	KeyPem          string `json:"key.pem"`
	Fingerprint     string `json:"fingerprint"`
	TenancyID       string `json:"tenancyId"`
}

// HostMetadata represents the host metadata structure
type HostMetadata struct {
	BuildingID     string `json:"buildingId"`
	ID             string `json:"id"`
	NetworkBlockID string `json:"networkBlockId"`
	RackID         string `json:"rackId"`
}

// NewIMDSClient creates a new IMDS client
func NewIMDSClient() *IMDSClient {
	return &IMDSClient{
		httpClient: &http.Client{
			Timeout: IMDSTimeout,
		},
		baseURL: IMDSBaseURL,
	}
}

// NewIMDSClientWithTimeout creates a new IMDS client with custom timeout
func NewIMDSClientWithTimeout(timeout time.Duration) *IMDSClient {
	return &IMDSClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: IMDSBaseURL,
	}
}

// makeRequest makes an HTTP request to the IMDS endpoint
func (c *IMDSClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	logger.Debugf("Making IMDS request to: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create IMDS request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required Authorization header for OCI IMDS v2
	req.Header.Set("Authorization", "Bearer Oracle")
	req.Header.Set("User-Agent", "rekharoy-oci-dr-hpc-v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Errorf("IMDS request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("IMDS request returned status %d", resp.StatusCode)
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read IMDS response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debugf("IMDS response received: %d bytes", len(body))
	return body, nil
}

// GetInstanceMetadata retrieves instance metadata from IMDS
func (c *IMDSClient) GetInstanceMetadata() (*InstanceMetadata, error) {
	logger.Info("Retrieving instance metadata from IMDS")

	body, err := c.makeRequest("instance")
	if err != nil {
		return nil, fmt.Errorf("failed to get instance metadata: %w", err)
	}

	var metadata InstanceMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		logger.Errorf("Failed to parse instance metadata: %v", err)
		return nil, fmt.Errorf("failed to parse instance metadata: %w", err)
	}

	logger.Infof("Successfully retrieved instance metadata for shape: %s", metadata.Shape)
	return &metadata, nil
}

// GetIdentityMetadata retrieves identity metadata from IMDS
func (c *IMDSClient) GetIdentityMetadata() (*IdentityMetadata, error) {
	logger.Info("Retrieving identity metadata from IMDS")

	body, err := c.makeRequest("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get identity metadata: %w", err)
	}

	var metadata IdentityMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		logger.Errorf("Failed to parse identity metadata: %v", err)
		return nil, fmt.Errorf("failed to parse identity metadata: %w", err)
	}

	logger.Info("Successfully retrieved identity metadata")
	return &metadata, nil
}

// GetHostMetadata retrieves host metadata from IMDS
func (c *IMDSClient) GetHostMetadata() (*HostMetadata, error) {
	logger.Info("Retrieving host metadata from IMDS")

	body, err := c.makeRequest("host")
	if err != nil {
		return nil, fmt.Errorf("failed to get host metadata: %w", err)
	}

	var metadata HostMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		logger.Errorf("Failed to parse host metadata: %v", err)
		return nil, fmt.Errorf("failed to parse host metadata: %w", err)
	}

	logger.Info("Successfully retrieved host metadata")
	return &metadata, nil
}

// GetRackID retrieves just the rack ID from host metadata
func (c *IMDSClient) GetRackID() (string, error) {
	logger.Info("Retrieving rack ID from IMDS")

	body, err := c.makeRequest("host/rackId")
	if err != nil {
		return "", fmt.Errorf("failed to get rack ID: %w", err)
	}

	rackID := string(body)
	logger.Infof("Successfully retrieved rack ID: %s", rackID)
	return rackID, nil
}

// GetBuildingID retrieves just the building ID from host metadata
func (c *IMDSClient) GetBuildingID() (string, error) {
	logger.Info("Retrieving building ID from IMDS")

	body, err := c.makeRequest("host/buildingId")
	if err != nil {
		return "", fmt.Errorf("failed to get building ID: %w", err)
	}

	buildingID := string(body)
	logger.Infof("Successfully retrieved building ID: %s", buildingID)
	return buildingID, nil
}

// GetHostID retrieves just the host ID from host metadata
func (c *IMDSClient) GetHostID() (string, error) {
	logger.Info("Retrieving host ID from IMDS")

	body, err := c.makeRequest("host/id")
	if err != nil {
		return "", fmt.Errorf("failed to get host ID: %w", err)
	}

	hostID := string(body)
	logger.Infof("Successfully retrieved host ID: %s", hostID)
	return hostID, nil
}

// GetNetworkBlockID retrieves just the network block ID from host metadata
func (c *IMDSClient) GetNetworkBlockID() (string, error) {
	logger.Info("Retrieving network block ID from IMDS")

	body, err := c.makeRequest("host/networkBlockId")
	if err != nil {
		return "", fmt.Errorf("failed to get network block ID: %w", err)
	}

	networkBlockID := string(body)
	logger.Infof("Successfully retrieved network block ID: %s", networkBlockID)
	return networkBlockID, nil
}

// GetVnicMetadata retrieves VNIC metadata from IMDS
func (c *IMDSClient) GetVnicMetadata() ([]VnicMetadata, error) {
	logger.Info("Retrieving VNIC metadata from IMDS")

	body, err := c.makeRequest("vnics")
	if err != nil {
		return nil, fmt.Errorf("failed to get VNIC metadata: %w", err)
	}

	var vnics []VnicMetadata
	if err := json.Unmarshal(body, &vnics); err != nil {
		logger.Errorf("Failed to parse VNIC metadata: %v", err)
		return nil, fmt.Errorf("failed to parse VNIC metadata: %w", err)
	}

	logger.Infof("Successfully retrieved VNIC metadata for %d VNICs", len(vnics))
	return vnics, nil
}

// GetPrimaryVnic retrieves the primary VNIC metadata (nicIndex 0)
func (c *IMDSClient) GetPrimaryVnic() (*VnicMetadata, error) {
	logger.Info("Retrieving primary VNIC metadata from IMDS")

	vnics, err := c.GetVnicMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get primary VNIC: %w", err)
	}

	for _, vnic := range vnics {
		if vnic.NicIndex == 0 {
			return &vnic, nil
		}
	}

	return nil, fmt.Errorf("primary VNIC not found")
}

// GetShape retrieves just the shape name from instance metadata
func (c *IMDSClient) GetShape() (string, error) {
	logger.Info("Retrieving shape from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get shape: %w", err)
	}

	return metadata.Shape, nil
}

// GetRegion retrieves just the region from instance metadata
func (c *IMDSClient) GetRegion() (string, error) {
	logger.Info("Retrieving region from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get region: %w", err)
	}

	return metadata.Region, nil
}

// GetAvailabilityDomain retrieves just the availability domain from instance metadata
func (c *IMDSClient) GetAvailabilityDomain() (string, error) {
	logger.Info("Retrieving availability domain from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get availability domain: %w", err)
	}

	return metadata.AvailabilityDomain, nil
}

// GetHostname retrieves just the hostname from instance metadata
func (c *IMDSClient) GetHostname() (string, error) {
	logger.Info("Retrieving hostname from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	return metadata.Hostname, nil
}

// GetCanonicalRegionName retrieves the full region identifier from instance metadata
func (c *IMDSClient) GetCanonicalRegionName() (string, error) {
	logger.Info("Retrieving canonical region name from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get canonical region name: %w", err)
	}

	return metadata.CanonicalRegionName, nil
}

// GetOciAdName retrieves the OCI AD name from instance metadata
func (c *IMDSClient) GetOciAdName() (string, error) {
	logger.Info("Retrieving OCI AD name from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get OCI AD name: %w", err)
	}

	return metadata.OciAdName, nil
}

// GetImageOCID retrieves the image OCID from instance metadata
func (c *IMDSClient) GetImageOCID() (string, error) {
	logger.Info("Retrieving image OCID from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get image OCID: %w", err)
	}

	return metadata.Image, nil
}

// GetInstanceOCID retrieves the instance OCID from instance metadata
func (c *IMDSClient) GetInstanceOCID() (string, error) {
	logger.Info("Retrieving instance OCID from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get instance OCID: %w", err)
	}

	return metadata.ID, nil
}

// GetCompartmentOCID retrieves the compartment OCID from instance metadata
func (c *IMDSClient) GetCompartmentOCID() (string, error) {
	logger.Info("Retrieving compartment OCID from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get compartment OCID: %w", err)
	}

	return metadata.CompartmentID, nil
}

// GetInstanceState retrieves the instance state from instance metadata
func (c *IMDSClient) GetInstanceState() (string, error) {
	logger.Info("Retrieving instance state from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get instance state: %w", err)
	}

	return metadata.State, nil
}

// GetRegionInfo retrieves the region information from instance metadata
func (c *IMDSClient) GetRegionInfo() (*RegionInfo, error) {
	logger.Info("Retrieving region info from IMDS")

	metadata, err := c.GetInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get region info: %w", err)
	}

	return &metadata.RegionInfo, nil
}

// GetRawMetadata retrieves raw metadata from a specific endpoint
func (c *IMDSClient) GetRawMetadata(endpoint string) ([]byte, error) {
	logger.Infof("Retrieving raw metadata from endpoint: %s", endpoint)
	return c.makeRequest(endpoint)
}

// IsRunningOnOCI checks if the application is running on an OCI instance
func (c *IMDSClient) IsRunningOnOCI() bool {
	logger.Debug("Checking if running on OCI instance")

	// Try to get instance metadata with a short timeout
	client := &IMDSClient{
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		baseURL: c.baseURL,
	}

	_, err := client.GetInstanceMetadata()
	return err == nil
}

// GetInstanceInfo retrieves comprehensive instance information
func (c *IMDSClient) GetInstanceInfo() (map[string]interface{}, error) {
	logger.Info("Retrieving comprehensive instance information")

	info := make(map[string]interface{})

	// Get instance metadata
	instanceMeta, err := c.GetInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance metadata: %w", err)
	}
	info["instance"] = instanceMeta

	// Get identity metadata
	identityMeta, err := c.GetIdentityMetadata()
	if err != nil {
		logger.Errorf("Failed to get identity metadata: %v", err)
		// Don't fail completely, just log the error
		info["identity"] = nil
	} else {
		info["identity"] = identityMeta
	}

	logger.Info("Successfully retrieved comprehensive instance information")
	return info, nil
}

// Convenience functions for common use cases

// GetCurrentShape is a convenience function to get the current instance shape
func GetCurrentShape() (string, error) {
	client := NewIMDSClient()
	return client.GetShape()
}

// GetCurrentRegion is a convenience function to get the current region
func GetCurrentRegion() (string, error) {
	client := NewIMDSClient()
	return client.GetRegion()
}

// GetCurrentInstanceMetadata is a convenience function to get current instance metadata
func GetCurrentInstanceMetadata() (*InstanceMetadata, error) {
	client := NewIMDSClient()
	return client.GetInstanceMetadata()
}

// GetCurrentIdentityMetadata is a convenience function to get current identity metadata
func GetCurrentIdentityMetadata() (*IdentityMetadata, error) {
	client := NewIMDSClient()
	return client.GetIdentityMetadata()
}

// GetCurrentVnicMetadata is a convenience function to get current VNIC metadata
func GetCurrentVnicMetadata() ([]VnicMetadata, error) {
	client := NewIMDSClient()
	return client.GetVnicMetadata()
}

// GetCurrentPrimaryVnic is a convenience function to get the primary VNIC metadata
func GetCurrentPrimaryVnic() (*VnicMetadata, error) {
	client := NewIMDSClient()
	return client.GetPrimaryVnic()
}

// IsRunningOnOCIInstance is a convenience function to check if running on OCI
func IsRunningOnOCIInstance() bool {
	client := NewIMDSClient()
	return client.IsRunningOnOCI()
}

// Additional convenience functions for new fields

// GetCurrentHostname is a convenience function to get the current hostname
func GetCurrentHostname() (string, error) {
	client := NewIMDSClient()
	return client.GetHostname()
}

// GetCurrentCanonicalRegionName is a convenience function to get the full region name
func GetCurrentCanonicalRegionName() (string, error) {
	client := NewIMDSClient()
	return client.GetCanonicalRegionName()
}

// GetCurrentOciAdName is a convenience function to get the OCI AD name
func GetCurrentOciAdName() (string, error) {
	client := NewIMDSClient()
	return client.GetOciAdName()
}

// GetCurrentImageOCID is a convenience function to get the image OCID
func GetCurrentImageOCID() (string, error) {
	client := NewIMDSClient()
	return client.GetImageOCID()
}

// GetCurrentInstanceOCID is a convenience function to get the instance OCID
func GetCurrentInstanceOCID() (string, error) {
	client := NewIMDSClient()
	return client.GetInstanceOCID()
}

// GetCurrentCompartmentOCID is a convenience function to get the compartment OCID
func GetCurrentCompartmentOCID() (string, error) {
	client := NewIMDSClient()
	return client.GetCompartmentOCID()
}

// GetCurrentInstanceState is a convenience function to get the instance state
func GetCurrentInstanceState() (string, error) {
	client := NewIMDSClient()
	return client.GetInstanceState()
}

// GetCurrentRegionInfo is a convenience function to get the region info
func GetCurrentRegionInfo() (*RegionInfo, error) {
	client := NewIMDSClient()
	return client.GetRegionInfo()
}

// GetCurrentTenantID is a convenience function to get the tenant ID
func GetCurrentTenantID() (string, error) {
	client := NewIMDSClient()
	metadata, err := client.GetInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get tenant ID: %w", err)
	}
	return metadata.TenantID, nil
}

// GetCurrentShapeConfig is a convenience function to get the shape configuration
func GetCurrentShapeConfig() (*ShapeConfig, error) {
	client := NewIMDSClient()
	metadata, err := client.GetInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get shape config: %w", err)
	}
	return metadata.ShapeConfig, nil
}

// GetCurrentAgentConfig is a convenience function to get the agent configuration
func GetCurrentAgentConfig() (*AgentConfig, error) {
	client := NewIMDSClient()
	metadata, err := client.GetInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent config: %w", err)
	}
	return metadata.AgentConfig, nil
}

// Convenience functions for host metadata

// GetCurrentHostMetadata is a convenience function to get current host metadata
func GetCurrentHostMetadata() (*HostMetadata, error) {
	client := NewIMDSClient()
	return client.GetHostMetadata()
}

// GetCurrentRackID is a convenience function to get the current rack ID
func GetCurrentRackID() (string, error) {
	client := NewIMDSClient()
	return client.GetRackID()
}

// GetCurrentBuildingID is a convenience function to get the current building ID
func GetCurrentBuildingID() (string, error) {
	client := NewIMDSClient()
	return client.GetBuildingID()
}

// GetCurrentHostID is a convenience function to get the current host ID
func GetCurrentHostID() (string, error) {
	client := NewIMDSClient()
	return client.GetHostID()
}

// GetCurrentNetworkBlockID is a convenience function to get the current network block ID
func GetCurrentNetworkBlockID() (string, error) {
	client := NewIMDSClient()
	return client.GetNetworkBlockID()
}
