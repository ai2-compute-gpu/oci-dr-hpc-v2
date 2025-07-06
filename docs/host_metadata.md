# Host Metadata Support

This document describes the host metadata functionality added to the OCI DR HPC v2 tool. This feature allows you to retrieve host-specific information from the Oracle Cloud Infrastructure Instance Metadata Service (IMDS).

## Overview

The host metadata functionality provides access to the following host-specific information:

- **Building ID**: Identifies the datacenter building where the host is located
- **Host ID**: Unique identifier for the physical host  
- **Network Block ID**: Identifies the network block/segment for the host
- **Rack ID**: Identifies the specific rack where the host is located

## API Endpoints

The functionality maps to these IMDS endpoints:

- `GET /opc/v2/host` - Returns complete host metadata as JSON
- `GET /opc/v2/host/rackId` - Returns only the rack ID
- `GET /opc/v2/host/buildingId` - Returns only the building ID  
- `GET /opc/v2/host/id` - Returns only the host ID
- `GET /opc/v2/host/networkBlockId` - Returns only the network block ID

## Usage Examples

### Using the IMDS Client

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

### Using Convenience Functions

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

## Available Functions

### Client Methods

- `GetHostMetadata() (*HostMetadata, error)` - Get complete host metadata
- `GetRackID() (string, error)` - Get rack ID only
- `GetBuildingID() (string, error)` - Get building ID only  
- `GetHostID() (string, error)` - Get host ID only
- `GetNetworkBlockID() (string, error)` - Get network block ID only

### Convenience Functions

- `GetCurrentHostMetadata() (*HostMetadata, error)` - Get complete host metadata
- `GetCurrentRackID() (string, error)` - Get current rack ID
- `GetCurrentBuildingID() (string, error)` - Get current building ID
- `GetCurrentHostID() (string, error)` - Get current host ID  
- `GetCurrentNetworkBlockID() (string, error)` - Get current network block ID

## Example Output

When running on an OCI HPC instance, the output looks like this:

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
```

## Requirements

- Must be running on an OCI instance with IMDS enabled
- Requires appropriate IAM permissions to access IMDS
- Network connectivity to the IMDS endpoint (169.254.169.254)

## Error Handling

All functions return appropriate error messages for common failure scenarios:

- Network connectivity issues
- IMDS authentication failures  
- Invalid response formats
- Non-OCI environments

The functions follow Go error handling conventions and provide descriptive error messages for debugging. 