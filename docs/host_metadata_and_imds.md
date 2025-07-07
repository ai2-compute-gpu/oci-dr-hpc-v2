# Host Metadata and IMDS Support

This document describes the host metadata functionality and Instance Metadata Service (IMDS) integration in the OCI DR HPC v2 tool. This feature allows you to retrieve host-specific information and general instance metadata from the Oracle Cloud Infrastructure Instance Metadata Service (IMDS).

## Overview

The IMDS functionality provides access to comprehensive instance and host-specific information:

### Host Metadata
- **Building ID**: Identifies the datacenter building where the host is located
- **Host ID**: Unique identifier for the physical host  
- **Network Block ID**: Identifies the network block/segment for the host
- **Rack ID**: Identifies the specific rack where the host is located

### General Instance Metadata
- **Instance Information**: Shape, state, availability domain, region details
- **Network Configuration**: VNICs, IP addresses, subnets, virtual router IPs
- **Identity Information**: Cryptographic material for instance principals
- **Agent Configuration**: OCI agent plugin status and configuration

## API Endpoints

### Host Metadata Endpoints
- `GET /opc/v2/host` - Returns complete host metadata as JSON
- `GET /opc/v2/host/rackId` - Returns only the rack ID
- `GET /opc/v2/host/buildingId` - Returns only the building ID  
- `GET /opc/v2/host/id` - Returns only the host ID
- `GET /opc/v2/host/networkBlockId` - Returns only the network block ID

### General IMDS Endpoints
- `GET /opc/v2/instance/` - Returns complete instance metadata
- `GET /opc/v2/vnics/` - Returns VNIC configuration details
- `GET /opc/v2/identity/` - Returns instance identity and cryptographic material

## Usage Examples

### Using curl Commands

#### Host Metadata with curl

```bash
# Get complete host metadata
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host

# Get specific host metadata fields
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host/rackId
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host/buildingId
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host/id
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host/networkBlockId
```

#### General IMDS with curl

```bash
# Get complete instance metadata
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/

# Get VNIC information
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/vnics/

# Get identity information
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/identity/

# Get specific instance fields
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/shape
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/id
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/region
```

### Using the Go IMDS Client

#### Host Metadata with Go

```go
package main

import (
    "fmt"
    "log"
    "github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

func main() {
    // Create IMDS client
    client := executor.NewIMDSClient()
    
    // Get complete host metadata
    hostMeta, err := client.GetHostMetadata()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Building ID: %s\n", hostMeta.BuildingID)
    fmt.Printf("Host ID: %s\n", hostMeta.ID) 
    fmt.Printf("Network Block ID: %s\n", hostMeta.NetworkBlockID)
    fmt.Printf("Rack ID: %s\n", hostMeta.RackID)
    
    // Get individual fields
    rackID, err := client.GetRackID()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Rack ID: %s\n", rackID)
}
```

#### General IMDS with Go

```go
package main

import (
    "fmt"
    "log"
    "github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

func main() {
    // Create IMDS client
    client := executor.NewIMDSClient()
    
    // Get instance shape
    shape, err := client.GetShape()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Instance shape: %s\n", shape)
    
    // Get instance ID
    instanceID, err := client.GetInstanceID()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Instance ID: %s\n", instanceID)
}
```

#### Using Convenience Functions

```go
package main

import (
    "fmt"
    "log"
    "github.com/oracle/oci-dr-hpc-v2/internal/executor"
)

func main() {
    // Get host metadata using convenience functions
    hostMeta, err := executor.GetCurrentHostMetadata()
    if err != nil {
        log.Fatal(err)
    }
    
    rackID, err := executor.GetCurrentRackID()
    if err != nil {
        log.Fatal(err)
    }
    
    buildingID, err := executor.GetCurrentBuildingID()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Rack: %s, Building: %s\n", rackID, buildingID)
}
```

## Data Structures

### HostMetadata

```go
type HostMetadata struct {
    BuildingID     string `json:"buildingId"`
    ID             string `json:"id"`
    NetworkBlockID string `json:"networkBlockId"`
    RackID         string `json:"rackId"`
}
```

### Instance Metadata (Sample)

```json
{
  "agentConfig": {
    "allPluginsDisabled": false,
    "managementDisabled": false,
    "monitoringDisabled": false,
    "pluginsConfig": [
      {
        "desiredState": "ENABLED",
        "name": "Compute HPC RDMA Authentication"
      },
      {
        "desiredState": "ENABLED",
        "name": "Compute HPC RDMA Auto-Configuration"
      },
      {
        "desiredState": "ENABLED",
        "name": "Compute RDMA GPU Monitoring"
      }
    ]
  },
  "availabilityDomain": "kWVD:US-rekharoy-1-AD-3",
  "canonicalRegionName": "us-rekharoy-1",
  "compartmentId": "ocid1.compartment.oc1..rekharoy",
  "definedTags": {
    "Oracle-Tag": {
      "CreatedBy": "ocid1.instance.oc1.us-chicago-1.rekharoy",
      "CreatedOn": "2025-06-12T02:32:22.888Z"
    }
  },
  "displayName": "BIO-2334XLG08T",
  "faultDomain": "FAULT-DOMAIN-1",
  "freeformTags": {
    "hostSerial": "2334XLG08T"
  },
  "hostname": "bio-2334xlg08t",
  "id": "ocid1.instance.oc1.us-chicago-1.rekharoy",
  "image": "ocid1.image.oc1.us-chicago-1.rekharoy",
  "metadata": {
    "ssh_authorized_keys": "rekharoy"
  },
  "ociAdName": "us-rekharoy-1-ad-1",
  "region": "us-rekharoy-1",
  "regionInfo": {
    "realmDomainComponent": "oraclecloud.com",
    "realmKey": "oc1",
    "regionIdentifier": "us-rekharoy-1",
    "regionKey": "rekharoy"
  },
  "shape": "BM.GPU.H100.8",
  "shapeConfig": {
    "maxVnicAttachments": 256,
    "memoryInGBs": 2048.0,
    "networkingBandwidthInGbps": 100.0,
    "ocpus": 112.0
  },
  "state": "Running",
  "tenantId": "ocid1.tenancy.oc1..rekharoy",
  "timeCreated": 1749695543362
}
```

### VNIC Metadata (Sample)

```json
[
  {
    "macAddr": "rekharoy:3f:d2:b3:0b:0c",
    "nicIndex": 0,
    "privateIp": "10.0.11.179",
    "subnetCidrBlock": "10.0.8.0/21",
    "virtualRouterIp": "10.0.8.1",
    "vlanTag": 0,
    "vnicId": "ocid1.vnic.oc1.us-chicago-1.rekharoy"
  }
]
```

### Identity Metadata (Sample)

The identity endpoint provides cryptographic material for instance principals. Key components include:

* `cert.pem`: Instance certificate
* `intermediate.pem`: Intermediate CA certificate
* `key.pem`: Private key for identity
* `fingerprint`: Certificate fingerprint (short identifier, useful for validation)
* `tenancyId`: OCID of the tenancy this instance belongs to

```json
{
  "cert.pem": "-----BEGIN CERTIFICATE-----\n...<rekharoy>...\n-----END CERTIFICATE-----\n",
  "intermediate.pem": "-----BEGIN CERTIFICATE-----\n...<rekharoy>...\n-----END CERTIFICATE-----\n",
  "key.pem": "-----BEGIN RSA PRIVATE KEY-----\n...<rekharoy>...\n-----END RSA PRIVATE KEY-----\n",
  "fingerprint": "test:fingerprint:12345",
  "tenancyId": "ocid1.tenancy.oc1..rekharoy"
}
```

> 🔒 **Important**: Do not expose or log the full `key.pem`. It grants identity-based authentication for the instance.

## Available Functions

### Client Methods

#### Host Metadata Methods
- `GetHostMetadata() (*HostMetadata, error)` - Get complete host metadata
- `GetRackID() (string, error)` - Get rack ID only
- `GetBuildingID() (string, error)` - Get building ID only  
- `GetHostID() (string, error)` - Get host ID only
- `GetNetworkBlockID() (string, error)` - Get network block ID only

#### General IMDS Methods
- `GetShape() (string, error)` - Get instance shape
- `GetInstanceID() (string, error)` - Get instance ID
- `GetRegion() (string, error)` - Get region
- `GetAvailabilityDomain() (string, error)` - Get availability domain

### Convenience Functions

#### Host Metadata Convenience Functions
- `GetCurrentHostMetadata() (*HostMetadata, error)` - Get complete host metadata
- `GetCurrentRackID() (string, error)` - Get current rack ID
- `GetCurrentBuildingID() (string, error)` - Get current building ID
- `GetCurrentHostID() (string, error)` - Get current host ID  
- `GetCurrentNetworkBlockID() (string, error)` - Get current network block ID

#### General IMDS Convenience Functions
- `GetCurrentShape() (string, error)` - Get current instance shape
- `GetCurrentInstanceID() (string, error)` - Get current instance ID
- `GetCurrentRegion() (string, error)` - Get current region

## Example Output

### Host Metadata Output

When running on an OCI HPC instance, the host metadata output looks like this:

```json
{
  "buildingId": "building:1725b321355f0314955d9e8d68cf2c54bc99db9ce21fecece62db989e763ac71",
  "id": "c108fca4ead3550cf7bbf0be7c0dce30b01b6b772a48f9f8e13346891a57f8a2", 
  "networkBlockId": "922fd61aa1af7d80d4edb08bcf09d6ae6f0d0152bb99843885fcd732b22716d9",
  "rackId": "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a"
}
```

## Testing

The functionality includes comprehensive unit tests:

```bash
# Run host metadata tests specifically
go test -v ./internal/executor -run TestGetHost

# Run all IMDS tests
go test ./internal/executor

# Test with curl commands
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host
```

## Requirements

- Must be running on an OCI instance with IMDS enabled
- Requires appropriate IAM permissions to access IMDS
- Network connectivity to the IMDS endpoint (169.254.169.254)
- For curl commands: `curl` must be installed and available

## Best Practices

- Use IMDSv2 only; disable IMDSv1 via Console or API
- Always include `Authorization: Bearer Oracle` header when using curl
- Cache metadata values to avoid hitting rate limits
- Secure `key.pem` if used (rotate regularly and never share)
- Use Go convenience functions for programmatic access within the application
- Use curl commands for quick testing and debugging

## Error Handling

All functions return appropriate error messages for common failure scenarios:

- Network connectivity issues
- IMDS authentication failures  
- Invalid response formats
- Non-OCI environments
- Rate limiting issues

The functions follow Go error handling conventions and provide descriptive error messages for debugging.

## Common Use Cases

### In OCI DR HPC v2 Tool

1. **Shape Detection**: Automatically detect instance shape for hardware validation
2. **Rack Identification**: Identify physical rack location for cluster mapping
3. **Network Configuration**: Retrieve network settings for connectivity tests
4. **Building/Datacenter Mapping**: Map instances to physical infrastructure

### Example Integration

```bash
# Quick host metadata check
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host | jq .

# Get shape for validation
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/shape

# Get rack ID for cluster mapping
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/host/rackId
```

## Reference

- [Oracle Cloud Infrastructure Instance Metadata Service](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/gettingmetadata.htm)
- [OCI HPC Documentation](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/runninghpcclusters.htm) 