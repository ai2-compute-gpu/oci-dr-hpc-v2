# System Information Implementation

This document describes the implementation of system information gathering from multiple sources in the OCI DR HPC v2 tool.

## Overview

The autodiscover functionality now gathers real system information from multiple sources instead of using hardcoded mock data:

- **Hostname**: Retrieved from OS using `os.Hostname()`
- **OCID**: Retrieved from IMDS `instance/id` endpoint  
- **Friendly Hostname**: Same as hostname (from OS)
- **Shape**: Retrieved from IMDS `instance/shape` endpoint
- **Serial Number**: Retrieved using `dmidecode -s chassis-serial-number`
- **Rack ID**: Retrieved from IMDS `host/rackId` endpoint

## Implementation Details

### Files Modified/Created

1. **`internal/executor/os_commands.go`**:
   - Added `GetHostname()` function
   - Added `GetSerialNumber()` function using dmidecode

2. **`internal/executor/os_commands_test.go`**:
   - Added tests for the new OS command functions

3. **`internal/executor/imds.go`**:
   - Added `HostMetadata` struct
   - Added `GetHostMetadata()` function
   - Added individual functions: `GetRackID()`, `GetBuildingID()`, `GetHostID()`, `GetNetworkBlockID()`
   - Added convenience functions for easy access

4. **`internal/executor/imds_test.go`**:
   - Added comprehensive tests for host metadata functionality

5. **`internal/autodiscover/system_info.go`** (NEW):
   - `SystemInfo` struct defining the system information model
   - `GatherSystemInfo()` function that collects from all sources
   - `GatherSystemInfoPartial()` function that continues on errors

6. **`internal/autodiscover/system_info_test.go`** (NEW):
   - Tests for the system information gathering functionality

7. **`internal/autodiscover/autodiscover.go`**:
   - Updated to use real system information instead of hardcoded values

8. **`docs/host_metadata.md`** (NEW):
   - Documentation for the host metadata functionality

## API Functions

### OS Commands (`internal/executor/os_commands.go`)

```go
// GetHostname retrieves the system hostname using os.Hostname()
func GetHostname() (string, error)

// GetSerialNumber retrieves the chassis serial number using dmidecode
func GetSerialNumber() (*OSCommandResult, error)
```

### IMDS Host Metadata (`internal/executor/imds.go`)

```go
// GetHostMetadata retrieves complete host metadata from IMDS
func (c *IMDSClient) GetHostMetadata() (*HostMetadata, error)

// Individual field getters
func (c *IMDSClient) GetRackID() (string, error)
func (c *IMDSClient) GetBuildingID() (string, error)
func (c *IMDSClient) GetHostID() (string, error)
func (c *IMDSClient) GetNetworkBlockID() (string, error)

// Convenience functions
func GetCurrentHostMetadata() (*HostMetadata, error)
func GetCurrentRackID() (string, error)
func GetCurrentBuildingID() (string, error)
func GetCurrentHostID() (string, error)
func GetCurrentNetworkBlockID() (string, error)
```

### System Information (`internal/autodiscover/system_info.go`)

```go
// GatherSystemInfo collects system information from multiple sources
func GatherSystemInfo() (*SystemInfo, error)

// GatherSystemInfoPartial collects system information and continues even if some sources fail
func GatherSystemInfoPartial() *SystemInfo
```

## Data Sources Mapping

| Field | Source | Command/Endpoint | Fallback |
|-------|--------|------------------|----------|
| Hostname | OS | `os.Hostname()` | "unknown" |
| OCID | IMDS | `/opc/v2/instance` (id field) | "unknown" |
| FriendlyHostname | OS | `os.Hostname()` | "unknown" |
| Shape | IMDS | `/opc/v2/instance` (shape field) | "unknown" |
| Serial | dmidecode | `sudo dmidecode -s chassis-serial-number` | "unknown" |
| Rack | IMDS | `/opc/v2/host/rackId` | "unknown" |

## Error Handling

The implementation uses two strategies for error handling:

1. **`GatherSystemInfo()`**: Returns an error if any source fails, but still returns partial data
2. **`GatherSystemInfoPartial()`**: Continues on errors and sets failed fields to "unknown"

The autodiscover command uses the partial strategy to ensure it always produces output.

## Testing

### Unit Tests
- OS command functions tested with both success and failure scenarios
- IMDS functions tested with mock server
- System info gathering tested in isolation

### Integration Tests
- Real system information gathering (skipped in non-OCI environments)
- End-to-end autodiscover functionality

## Usage Examples

### Command Line
```bash
# Run autodiscover with real system information
./oci-dr-hpc-v2 autodiscover

# Specify output format
./oci-dr-hpc-v2 autodiscover -o friendly

# Save to custom file
./oci-dr-hpc-v2 autodiscover -f /tmp/my-system.json
```

### Programmatic Access
```go
// Get system information
sysInfo := autodiscover.GatherSystemInfoPartial()
fmt.Printf("Hostname: %s\n", sysInfo.Hostname)
fmt.Printf("Shape: %s\n", sysInfo.Shape)
fmt.Printf("Serial: %s\n", sysInfo.Serial)

// Get individual components
hostname, _ := executor.GetHostname()
rackID, _ := executor.GetCurrentRackID()
```

## Expected Output on OCI Instance

When running on an actual OCI instance, the output would look like:

```json
{
  "hostname": "bio-2334xlg08t",
  "ocid": "ocid1.instance.oc1.us-chicago-1.anxxeljr7rhxvoacvwyntbjri74hfv7abuyprpnnscgebwmdu264bpskcxtq",
  "friendly_hostname": "bio-2334xlg08t",
  "shape": "BM.GPU.H100.8",
  "serial": "2334XLG08T",
  "rack": "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a"
}
```

## Benefits

1. **Real Data**: Uses actual system information instead of hardcoded values
2. **Resilient**: Continues to work even when some data sources are unavailable
3. **Comprehensive**: Gathers information from multiple authoritative sources
4. **Testable**: Well-tested with both unit and integration tests
5. **Documented**: Clear API and usage documentation 