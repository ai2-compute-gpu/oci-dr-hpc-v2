# OCI Shapes Configuration Package

This package provides functionality to read and query Oracle Cloud Infrastructure (OCI) shape configurations from a JSON file. It supports querying RDMA settings, network configurations, and shape-specific parameters for HPC and GPU workloads.

## Features

- 📋 **Load and parse** shapes configuration from JSON
- 🔍 **Query specific shapes** and their settings
- 🎮 **Filter shapes** by type (GPU, HPC)
- 🔌 **Search shapes** by model or partial name
- ⚙️ **Access detailed settings** like RDMA, network, and PCIe configurations
- 🧪 **Comprehensive testing** with 100% test coverage

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/oracle/oci-dr-hpc-v2/internal/shapes"
)

func main() {
    // Initialize the ShapeManager
    manager, err := shapes.NewShapeManager("shapes.json")
    if err != nil {
        log.Fatalf("Failed to load shapes: %v", err)
    }
    
    // Get all GPU shapes
    gpuShapes := manager.GetGPUShapes()
    fmt.Printf("Found %d GPU shapes\n", len(gpuShapes))
    
    // Query specific shape
    if manager.IsShapeSupported("BM.GPU.H100.8") {
        info, _ := manager.GetShapeInfo("BM.GPU.H100.8")
        fmt.Printf("Shape: %s, Model: %s\n", info.Name, info.Model)
    }
}
```

## API Reference

### ShapeManager

#### Creation
```go
manager, err := shapes.NewShapeManager("path/to/shapes.json")
```

#### Query Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `GetAllShapes()` | Get all supported shapes | `[]string` |
| `GetGPUShapes()` | Get GPU shapes only | `[]string` |
| `GetHPCShapes()` | Get HPC shapes only | `[]string` |
| `GetShapeConfig(name)` | Get full config for a shape | `*RDMAShapeConfig, error` |
| `GetShapeInfo(name)` | Get comprehensive shape info | `*ShapeInfo, error` |
| `GetShapeSettings(name)` | Get settings for a shape | `*ShapeSettings, error` |
| `GetShapeModel(name)` | Get model for a shape | `string, error` |
| `IsShapeSupported(name)` | Check if shape is supported | `bool` |

#### Search Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `SearchShapes(query)` | Search shapes by partial name | `[]string` |
| `GetShapesByModel(model)` | Get shapes using specific model | `[]string` |
| `GetSupportedModels()` | Get all supported models | `[]string` |

#### Configuration Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `GetRDMANetworkConfig()` | Get RDMA network configuration | `[]RDMANetwork` |

### Data Structures

#### ShapeInfo
```go
type ShapeInfo struct {
    Name     string        `json:"name"`
    Model    string        `json:"model"`
    Settings ShapeSettings `json:"settings"`
    IsGPU    bool          `json:"is_gpu"`
    IsHPC    bool          `json:"is_hpc"`
}
```

#### ShapeSettings
```go
type ShapeSettings struct {
    Ring          RingSettings           `json:"ring"`
    Channels      string                 `json:"channels"`
    MTU           string                 `json:"mtu"`
    DSCPRDMA      string                 `json:"dscp_rdma"`
    DSCPGPU       string                 `json:"dscp_gpu"`
    ROCEAccl      map[string]interface{} `json:"roce_accl"`
    Buffer        []string               `json:"buffer"`
    PFC           []string               `json:"pfc"`
    // ... other fields
}
```

## Usage Examples

### 1. List All Shapes
```go
allShapes := manager.GetAllShapes()
for _, shape := range allShapes {
    fmt.Printf("Shape: %s\n", shape)
}
```

### 2. Find GPU Shapes
```go
gpuShapes := manager.GetGPUShapes()
fmt.Printf("Found %d GPU shapes:\n", len(gpuShapes))
for _, shape := range gpuShapes {
    fmt.Printf("  • %s\n", shape)
}
```

### 3. Query Shape Details
```go
shapeName := "BM.GPU.H100.8"
if manager.IsShapeSupported(shapeName) {
    info, err := manager.GetShapeInfo(shapeName)
    if err == nil {
        fmt.Printf("Name: %s\n", info.Name)
        fmt.Printf("Model: %s\n", info.Model)
        fmt.Printf("MTU: %s\n", info.Settings.MTU)
        fmt.Printf("Channels: %s\n", info.Settings.Channels)
    }
}
```

### 4. Search for Shapes
```go
// Search by partial name
h100Shapes := manager.SearchShapes("H100")

// Get shapes by model
cx7Shapes := manager.GetShapesByModel("ConnectX-7")
```

### 5. Get Network Configuration
```go
settings, err := manager.GetShapeSettings("BM.GPU.H100.8")
if err == nil {
    fmt.Printf("MTU: %s\n", settings.MTU)
    fmt.Printf("Ring RX: %s, TX: %s\n", settings.Ring.RX, settings.Ring.TX)
    fmt.Printf("DSCP RDMA: %s\n", settings.DSCPRDMA)
}
```

### 6. Check RDMA Configuration
```go
rdmaConfigs := manager.GetRDMANetworkConfig()
if len(rdmaConfigs) > 0 {
    config := rdmaConfigs[0]
    fmt.Printf("RDMA Network: %s\n", config.DefaultSettings.RDMANetwork)
    fmt.Printf("Netmask: %s\n", config.SubnetSettings.Netmask)
}
```

## Supported Shapes

The package supports the following OCI shape families:

### GPU Shapes
- **H100 Family**: `BM.GPU.H100.8`, `BM.GPU.H100T.8`
- **H200 Family**: `BM.GPU.H200.8`
- **B200 Family**: `BM.GPU.B200.8`
- **GB200 Family**: `BM.GPU.GB200.4`
- **A100 Family**: `BM.GPU.A100-v2.8`
- **L40S Family**: `BM.GPU.L40S.4`
- **Legacy GPU**: `BM.GPU.T1.2`, `BM.GPU.GU1.2`, etc.

### HPC Shapes
- **HPC2 Family**: `BM.HPC2.36`
- **HPC E5 Family**: `BM.HPC.E5.128`, `BM.HPC.E5.144`, `BM.HPC.E5-RBR.144`
- **Optimized**: `BM.Optimized3.36`

### Network Models
- **ConnectX-7**: Latest generation for H100/H200/B200 shapes
- **ConnectX-6 Dx**: For HPC E5 shapes  
- **ConnectX-5 Ex**: For legacy GPU and HPC2 shapes

## Testing

Run the comprehensive test suite:

```bash
go test -v ./internal/shapes
```

The tests cover:
- ✅ JSON parsing and validation
- ✅ Shape querying and filtering
- ✅ Search functionality
- ✅ Configuration access
- ✅ Error handling
- ✅ Edge cases

## Demo

Run the interactive demo:

```bash
go run cmd/shapes_demo.go internal/shapes/shapes.json
```

This shows all the package capabilities with real data from your shapes.json file.

## Configuration File Format

The package expects a JSON file with this structure:

```json
{
  "version": "__VERSION__",
  "rdma-network": [...],
  "rdma-settings": [
    {
      "shapes": ["BM.GPU.H100.8", "BM.GPU.H200.8"],
      "model": "Mellanox Technologies MT2910 Family [ConnectX-7]",
      "settings": {
        "mtu": "4220",
        "channels": "8",
        "ring": {"rx": "8192", "tx": "8192"},
        "dscp_rdma": "26",
        "dscp_gpu": "10",
        ...
      }
    }
  ]
}
```

## Integration

This package integrates with other components:

- **Logger**: Uses `internal/logger` for structured logging
- **Config**: Can be used with `internal/config` for application settings
- **CLI**: Can be integrated into CLI commands for shape querying
- **Tests**: Follows the same testing patterns as other packages

## Performance

- **Fast Loading**: JSON parsing optimized for large configuration files
- **Efficient Queries**: In-memory lookups with O(1) access patterns
- **Memory Efficient**: Lazy loading and minimal memory overhead
- **Concurrent Safe**: Thread-safe read operations after initialization