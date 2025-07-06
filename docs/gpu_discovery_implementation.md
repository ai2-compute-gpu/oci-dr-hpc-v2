# GPU Discovery Implementation

This document describes the implementation of real GPU discovery using nvidia-smi queries instead of hardcoded values in the OCI DR HPC v2 tool.

## Overview

The GPU discovery functionality has been enhanced to query real GPU information from nvidia-smi instead of using hardcoded mock data. This provides accurate, real-time GPU hardware information for the autodiscover process.

## Implementation Details

### New Functions Added

#### 1. **nvidia-smi Query Functions** (`internal/executor/nvidia_smi.go`)

```go
// GetGPUInfo queries nvidia-smi for comprehensive GPU information
func GetGPUInfo() ([]GPUInfo, error)

// parseGPUInfo parses the nvidia-smi output into GPUInfo structs  
func parseGPUInfo(output string) ([]GPUInfo, error)

// GetGPUCount queries nvidia-smi for the number of GPUs
func GetGPUCount() (int, error)

// IsNvidiaSMIAvailable checks if nvidia-smi is available and working
func IsNvidiaSMIAvailable() bool
```

#### 2. **GPU Discovery Functions** (`internal/autodiscover/gpu_discovery.go`)

```go
// DiscoverGPUs attempts to discover GPU information using nvidia-smi
func DiscoverGPUs() []GPU

// DiscoverGPUsWithFallback attempts to discover GPUs with fallback to mock data
func DiscoverGPUsWithFallback() []GPU
```

### Data Structure

The `GPUInfo` struct in `executor` and `GPU` struct in `autodiscover` are identical:

```go
type GPUInfo struct {
    PCI   string `json:"pci"`
    Model string `json:"model"`
    ID    int    `json:"id"`
}
```

## nvidia-smi Query Details

### Query Used
The implementation uses the following nvidia-smi query:
```bash
nvidia-smi --query-gpu=pci.bus_id,name,index --format=csv,noheader,nounits
```

### Sample Output
```
00000000:65:00.0, NVIDIA GeForce GTX 1650, 0
00000000:66:00.0, NVIDIA H100 80GB HBM3, 1
```

### Parsing Logic
- Splits CSV output by commas
- Extracts PCI address, GPU model name, and index
- Handles parsing errors gracefully
- Skips invalid lines and continues processing

## Error Handling Strategy

The implementation uses a multi-layered error handling approach:

1. **Check nvidia-smi availability**: Uses `exec.LookPath()` to verify nvidia-smi is installed
2. **Graceful fallback**: If nvidia-smi is not available, falls back to mock data for development
3. **Partial success**: If some GPUs fail to parse, continues with successfully parsed GPUs
4. **Empty results**: Returns empty slice if no GPUs are detected (valid for systems without GPUs)

## Integration with Autodiscover

### Before (Hardcoded)
```go
Gpus: []GPU{
    {PCI: "0000:0f:00.0", Model: "NVIDIA H100 80GB HBM3", ID: 0},
    {PCI: "0000:2d:00.0", Model: "NVIDIA H100 80GB HBM3", ID: 1},
},
```

### After (Real Discovery)
```go
// Discover real GPU information
discoveredGPUs := DiscoverGPUsWithFallback()
mapHost := MapHost{
    // ... other fields
    Gpus: discoveredGPUs,
}
```

## Testing

### Unit Tests
- **`TestGetGPUInfo`**: Tests real GPU discovery (skipped if nvidia-smi unavailable)
- **`TestParseGPUInfo`**: Tests parsing logic with various input scenarios
- **`TestGetGPUCount`**: Tests GPU count functionality
- **`TestDiscoverGPUs`**: Tests the autodiscover GPU discovery
- **`TestGPUStructConversion`**: Tests conversion between different GPU structs

### Test Coverage
- ✅ Real nvidia-smi integration
- ✅ Parsing edge cases (empty output, invalid format, mixed valid/invalid)
- ✅ Error handling (nvidia-smi not available, parsing failures)
- ✅ Fallback functionality
- ✅ Data structure validation

## Real vs. Mock Data Examples

### Test Environment (Mock Fallback)
```json
{
  "gpu": [
    {
      "pci": "0000:0f:00.0",
      "model": "NVIDIA H100 80GB HBM3 (Mock)",
      "id": 0
    }
  ]
}
```

### System with Real GPU
```json
{
  "gpu": [
    {
      "pci": "00000000:65:00.0", 
      "model": "NVIDIA GeForce GTX 1650",
      "id": 0
    }
  ]
}
```

### OCI Instance with Multiple GPUs
```json
{
  "gpu": [
    {
      "pci": "00000000:0F:00.0",
      "model": "NVIDIA H100 80GB HBM3", 
      "id": 0
    },
    {
      "pci": "00000000:2D:00.0",
      "model": "NVIDIA H100 80GB HBM3",
      "id": 1
    }
  ]
}
```

## Usage Examples

### Command Line
```bash
# Basic usage - now shows real GPU data
./oci-dr-hpc-v2 autodiscover

# View GPU information in different formats
./oci-dr-hpc-v2 autodiscover -o friendly
./oci-dr-hpc-v2 autodiscover -o table
./oci-dr-hpc-v2 autodiscover -o json
```

### Programmatic Access
```go
// Direct GPU discovery
gpus := autodiscover.DiscoverGPUs()
for _, gpu := range gpus {
    fmt.Printf("GPU %d: %s at %s\n", gpu.ID, gpu.Model, gpu.PCI)
}

// With fallback for development
gpus := autodiscover.DiscoverGPUsWithFallback()

// Low-level nvidia-smi access
gpuInfos, err := executor.GetGPUInfo()
if err != nil {
    log.Printf("GPU discovery failed: %v", err)
}
```

## Benefits

1. **Real Hardware Data**: Uses actual GPU information instead of mock data
2. **Accurate Inventory**: Provides correct PCI addresses, models, and counts
3. **Development Friendly**: Falls back to mock data when nvidia-smi unavailable
4. **Robust**: Handles various error conditions gracefully
5. **Well Tested**: Comprehensive test coverage for all scenarios
6. **Performance**: Efficient single nvidia-smi query gets all needed information

## Dependencies

- **nvidia-smi**: Must be installed and accessible in PATH for real GPU discovery
- **GPU Drivers**: Requires working NVIDIA GPU drivers
- **Permissions**: No special permissions required (nvidia-smi runs as user)

## Backward Compatibility

- ✅ **API Compatible**: Same JSON output structure
- ✅ **Fallback Behavior**: Continues to work without nvidia-smi
- ✅ **Output Formats**: All existing output formats (json, table, friendly) supported
- ✅ **File Output**: Same file output behavior maintained

The implementation seamlessly replaces hardcoded GPU data with real discovery while maintaining full backward compatibility and robust error handling. 