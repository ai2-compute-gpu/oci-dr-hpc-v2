# Hardware Autodiscovery Algorithm

*@rekharoy*

## Overview

The OCI DR HPC v2 autodiscovery system creates a comprehensive hardware inventory of High-Performance Computing (HPC) environments by combining static topology information from Oracle Cloud Infrastructure (OCI) shape configurations with real-time system discovery from the operating system.

## Architecture

The autodiscovery process follows a hybrid approach that maximizes accuracy while maintaining reliability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   shapes.json   │    │  OCI IMDS API   │    │  OS Discovery   │
│  (Static Topo)  │    │ (Metadata Svc)  │    │ (Runtime Info)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                         ┌─────────────────┐
                         │   Autodiscover  │
                         │    Algorithm    │
                         └─────────────────┘
                                 │
                    ┌─────────────────────────┐
                    │    Hardware Map         │
                    │ (JSON/Table/Friendly)   │
                    └─────────────────────────┘
```

## Core Algorithm

### Main Discovery Flow

```go
func Run() {
    // 1. System Information Discovery
    sysInfo := GatherSystemInfoPartial()
    
    // 2. Hardware Component Discovery  
    discoveredGPUs := DiscoverGPUsWithFallback()
    discoveredRDMANics := DiscoverRDMANicsWithFallback(sysInfo.Shape)
    discoveredVCNNic := DiscoverVCNNicWithFallback(sysInfo.Shape)
    
    // 3. Cluster Detection
    inCluster := sysInfo.NetworkBlockId != "" && sysInfo.NetworkBlockId != "unknown"
    
    // 4. Hardware Map Construction
    mapHost := MapHost{
        // System metadata + discovered hardware
    }
    
    // 5. Output Generation (JSON/Table/Friendly)
    // 6. File Output (always JSON format)
}
```

## Data Sources

### 1. OCI Instance Metadata Service (IMDS)

**Purpose**: Provides OCI-specific instance metadata
**Endpoint**: `http://169.254.169.254/opc/v2/`

**Retrieved Information**:
- `instance/id` → Instance OCID
- `instance/hostname` → Instance hostname  
- `instance/shape` → OCI shape name (e.g., BM.GPU.H100.8)
- `host/rackId` → Physical rack identifier
- `host/networkBlockId` → Cluster network block ID
- `host/buildingId` → Data center building ID

**Example IMDS Response**:
```json
{
  "buildingId": "building:8c78b091ea762ea0f270678bd297f75bc5c5054410f5c1b192704169f025bc1d",
  "id": "3525493a7853a44a72dd76ada9be95c635e53cc818bf2b5630c5c7328d6faf7a", 
  "networkBlockId": "9877ad75a930eec924e1bb971688e0c338ab613cba5afc9793b5b4754902477d",
  "rackId": "53c6b7740e043b1becb2d5654df67738ac2edc46edcb7eb838a4504cf99ba835"
}
```

### 2. Shapes Configuration (shapes.json)

**Purpose**: Provides static hardware topology for each OCI shape
**Location**: `/etc/oci-dr-hpc-shapes.json`

**Contains**:
- PCI addresses for GPUs, RDMA NICs, VCN NICs
- Expected device models and configurations
- GPU-to-RDMA NIC mappings for topology understanding

**Example Shape Configuration**:
```json
{
  "shape": "BM.GPU.H100.8",
  "gpu": [
    {
      "pci": "0000:0f:00.0",
      "model": "NVIDIA H100 80GB HBM3", 
      "id": 0,
      "module_id": 2
    }
  ],
  "rdma-nics": [
    {
      "pci": "0000:0c:00.0",
      "device_name": "mlx5_0",
      "model": "Mellanox Technologies MT2910 Family [ConnectX-7]",
      "gpu_pci": "0000:0f:00.0",
      "gpu_id": "0"
    }
  ]
}
```

### 3. Operating System Discovery

**Purpose**: Discovers real-time hardware state and runtime configuration

**Methods Used**:
- `lspci` - PCI device enumeration and details
- `nvidia-smi` - GPU device discovery and status
- `/sys/bus/pci/devices/` - Sysfs PCI device information
- `ibdev2netdev` - InfiniBand to network interface mapping
- `ip addr` - Network interface IP address discovery
- `dmidecode` - System serial number extraction

## Component Discovery Algorithms

### GPU Discovery

```
DiscoverGPUsWithFallback()
├── CheckNvidiaSMI() 
│   └── Test nvidia-smi availability
├── GetGPUInfo()
│   ├── nvidia-smi --query-gpu=pci.bus_id,name,index
│   ├── Parse CSV output
│   └── Map to GPU structs
└── Fallback: Return empty list if nvidia-smi unavailable
```

**GPU Data Structure**:
```go
type GPU struct {
    PCI   string `json:"pci"`     // e.g., "0000:0f:00.0"
    Model string `json:"model"`   // e.g., "NVIDIA H100 80GB HBM3"  
    ID    string `json:"id"`      // e.g., "0"
}
```

### RDMA NIC Discovery

**Hybrid Discovery Process**:
```
DiscoverRDMANicsWithFallback(shapeName)
├── Get static topology from shapes.json
│   ├── PCI addresses
│   ├── GPU mappings (gpu_pci, gpu_id)
│   └── Expected device names
├── OS Discovery for runtime values
│   ├── GetPCIDeviceModel(pci) → lspci device description
│   ├── GetInfiniBandDeviceName(pci) → /sys/.../infiniband/*/
│   ├── GetNetworkInterfaceName(pci) → /sys/.../net/*/
│   ├── GetPCIDeviceNUMANode(pci) → /sys/.../numa_node
│   └── GetRDMADeviceIP(device) → ibdev2netdev + ip addr
└── Fallback: Use shapes.json values or "undefined"
```

**RDMA NIC Data Structure**:
```go
type RdmaNic struct {
    PCI        string `json:"pci"`         // From shapes.json
    Interface  string `json:"interface"`   // From OS discovery
    RdmaIP     string `json:"rdma_ip"`     // From OS discovery  
    DeviceName string `json:"device_name"` // From OS discovery
    Model      string `json:"model"`       // From OS discovery
    Numa       string `json:"numa"`        // From OS discovery
    GpuID      string `json:"gpu_id"`      // From shapes.json
    GpuPCI     string `json:"gpu_pci"`     // From shapes.json
}
```

### VCN NIC Discovery

**Similar hybrid approach as RDMA NICs**:
```
DiscoverVCNNicWithFallback(shapeName)
├── Get PCI address from shapes.json
├── OS Discovery
│   ├── GetPCIDeviceModel(pci) → Device model
│   ├── GetInfiniBandDeviceName(pci) → IB device name
│   ├── GetNetworkInterfaceName(pci) → Network interface
│   └── getInterfaceIP(interface) → Private IP address
└── Fallback to shapes.json or "undefined"
```

### System Information Discovery

```
GatherSystemInfoPartial()
├── GetHostname() → os.Hostname()
├── IMDS Queries
│   ├── GetCurrentInstanceOCID()
│   ├── GetCurrentShape() 
│   ├── GetCurrentRackID()
│   ├── GetCurrentNetworkBlockID()
│   └── GetCurrentBuildingID()
├── GetSerialNumber() → dmidecode chassis-serial-number
└── Continue on errors (partial mode)
```

## Cluster Detection Algorithm

The system determines if an instance is part of a cluster based on network block ID availability:

```go
// Cluster detection logic
inCluster := sysInfo.NetworkBlockId != "" && sysInfo.NetworkBlockId != "unknown"
```

**Logic**:
- `in_cluster: true` → Valid networkBlockId from IMDS (instance is clustered)
- `in_cluster: false` → Missing/unknown networkBlockId (standalone instance)

## Output Formats

### JSON Format (Default)
```json
{
  "hostname": "bio-2334xlg08t",
  "shape": "BM.GPU.H100.8", 
  "rack": "8d93acc296b77c923d0778079061b64094d55b3fbe4eb54460655e916cddf34a",
  "network_block_id": "9877ad75a930eec924e1bb971688e0c338ab613cba5afc9793b5b4754902477d",
  "building_id": "building:8c78b091ea762ea0f270678bd297f75bc5c5054410f5c1b192704169f025bc1d",
  "in_cluster": true,
  "gpu": [...],
  "rdma_nics": [...],
  "vcn_nic": {...}
}
```

### Table Format
```
┌─────────────────────────────────────────────────────────────────┐
│                    HARDWARE DISCOVERY RESULTS                  │
├─────────────────────────────────────────────────────────────────┤
│ SYSTEM INFORMATION                                              │
│ Hostname: bio-2334xlg08t                                        │
│ Shape: BM.GPU.H100.8                                            │
│ Network Block: 9877ad75a930eec924e1bb971688e0c338ab613cb...     │
└─────────────────────────────────────────────────────────────────┘
```

### Friendly Format
```
🔍 Hardware Discovery Results
🖥️  System Information
   Hostname: bio-2334xlg08t
   Shape: BM.GPU.H100.8
🎮 GPU Devices  
   ✅ GPU 0: NVIDIA H100 80GB HBM3
🌐 RDMA Network Interfaces
   ✅ RDMA NIC 1: enp12s0f0np0
```

## Error Handling & Fallback Strategy

### Graceful Degradation
1. **IMDS Failure** → Use "unknown" values, continue discovery
2. **nvidia-smi Missing** → Return empty GPU list
3. **OS Discovery Failure** → Fall back to shapes.json values
4. **Missing shapes.json** → Use "undefined" placeholders

### Logging Strategy
- **INFO**: Normal discovery progress
- **ERROR**: Failed operations (with fallback)
- **DEBUG**: Detailed discovery information

## File Output

**Always saves JSON format** to:
- Default: `map_host_<hostname>.json`
- Custom: User-specified path via `--output-file`

**Directory Creation**: Automatically creates parent directories for custom paths

## Performance Considerations

### Parallel Discovery
- System info, GPU, RDMA, and VCN discovery run independently
- Failed operations don't block other discoveries
- Timeout mechanisms for external commands

### Caching
- Shapes configuration loaded once per execution
- IMDS client reuses HTTP connections
- OS command results used immediately (no caching)

## Security Considerations

- **No secrets exposure**: Only hardware topology information
- **Read-only operations**: No system modifications
- **Privilege escalation**: Uses `sudo` only for `lspci` and `dmidecode`
- **Network access**: Limited to OCI IMDS endpoints

## Extension Points

### Adding New Hardware Types
1. Define data structure in `autodiscover/` package
2. Implement discovery function with fallback
3. Add to shapes.json configuration
4. Update output formatters

### Custom Discovery Sources
1. Implement discovery interface
2. Add to hybrid discovery chain
3. Maintain fallback compatibility

## Dependencies

### System Requirements
- Linux operating system
- `lspci` (pciutils package)
- `dmidecode` (dmidecode package)  
- `ip` (iproute2 package)

### Optional Dependencies
- `nvidia-smi` (for GPU discovery)
- `ibdev2netdev` (for RDMA IP discovery)
- `rdma` tools (for advanced RDMA info)

### Go Dependencies
- Cobra (CLI framework)
- Viper (configuration management)
- Standard library (net/http, os/exec, etc.)

## Testing Strategy

### Unit Test Coverage
- **Mock-based testing**: Tests parsing logic without hardware requirements
- **Edge case handling**: Invalid inputs, missing commands, network failures
- **Data structure validation**: JSON serialization/deserialization
- **Fallback verification**: Ensures graceful degradation

### Integration Testing
- **End-to-end discovery**: Full workflow validation
- **Multi-format output**: JSON, table, and friendly format generation
- **Error simulation**: Network timeouts, missing tools, invalid shapes

---

*This documentation covers the complete autodiscovery algorithm and implementation details for OCI DR HPC v2.*