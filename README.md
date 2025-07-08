# OCI DR HPC v2

A comprehensive diagnostic and repair tool for High Performance Computing (HPC) environments with GPU and RDMA support on Oracle Cloud Infrastructure (OCI).

## 🚀 Features

- **🎮 GPU Diagnostics**: Check GPU count, driver status, and hardware health using nvidia-smi
- **🔗 RDMA Network Testing**: Validate RDMA NIC count, PCI addresses, and connectivity with hybrid discovery
- **⚡ PCIe Error Detection**: Scan system logs for PCIe-related hardware errors
- **🔍 Hardware Autodiscovery**: Generate logical hardware models with IMDS integration and cluster detection
- **📊 Multiple Output Formats**: Support for table, JSON, and friendly human-readable output
- **🔧 Smart Recommendations**: JSON-configurable diagnostic recommendations with fault codes
- **⚙️ Flexible Configuration**: Support for config files, environment variables, and CLI flags
- **🏗️ Smart Path Resolution**: Automatic detection of development vs production environments
- **📦 Customer-Ready Deployment**: Makefile-based installation with filesystem hierarchy compliance
- **🐛 Debug-Friendly**: Comprehensive logging with config path visibility for troubleshooting
- **🏗️ Cross-Platform Build**: Support for AMD64 and ARM64 architectures with cross-compilation

## 🌐 Cross-Platform Support

The OCI DR HPC v2 tool supports multiple architectures with automatic detection and cross-compilation capabilities:

### Supported Architectures
- **AMD64 (x86_64)**: Intel and AMD 64-bit processors
- **ARM64 (aarch64)**: ARM 64-bit processors (AWS Graviton, Apple Silicon, etc.)

### Build System Features
- **Automatic Architecture Detection**: Detects your system architecture and builds accordingly
- **Cross-Compilation**: Build binaries for both architectures from any supported system
- **Multi-Architecture Packages**: Create RPM and DEB packages for both architectures simultaneously
- **Smart Binary Naming**: Architecture-specific binaries (`oci-dr-hpc-v2-amd64`, `oci-dr-hpc-v2-arm64`)

### Quick Cross-Compilation Example
```bash
# Build for both architectures on any system
make build-all

# Results:
# build/oci-dr-hpc-v2-amd64    # For x86_64 systems
# build/oci-dr-hpc-v2-arm64    # For ARM64 systems

# Create packages for distribution
make all-cross               # Build everything for both architectures
```

## 📁 Project Structure

```
oci-dr-hpc-v2/
├── cmd/                    # CLI command definitions (Cobra framework)
│   ├── root.go            # Main CLI entry point and config initialization
│   ├── level1.go          # Level 1 diagnostic commands
│   ├── autodiscover.go    # Hardware autodiscovery commands
│   └── recommender.go     # Recommendation analysis commands
├── configs/               # Configuration files
│   ├── oci-dr-hpc.yaml   # Default application configuration
│   └── recommendations.json # Diagnostic recommendations with fault codes
├── docs/                  # Documentation
│   ├── autodiscovery.md  # Autodiscovery algorithm documentation (@rekharoy)
│   ├── recommendations-config.md # Recommendation system documentation
│   ├── deployment.md     # Customer deployment guide
│   ├── imds.md          # IMDS (Instance Metadata Service) documentation
│   └── *.md             # Additional documentation
├── internal/              # Internal application logic
│   ├── autodiscover/     # Hardware discovery and modeling
│   │   ├── autodiscover.go       # Main autodiscovery logic with IMDS integration
│   │   ├── gpu_discovery.go      # GPU detection using nvidia-smi
│   │   ├── network_discovery.go  # Hybrid RDMA/VCN NIC discovery
│   │   └── system_info.go        # System info with networkBlockId and buildingId
│   ├── config/           # Configuration management (Viper integration)
│   │   └── config.go     # Config loading with smart path resolution
│   ├── executor/         # System command execution
│   │   ├── nvidia_smi.go # NVIDIA GPU command execution
│   │   ├── os_commands.go # OS-level commands with runtime hardware discovery
│   │   ├── imds.go       # Instance Metadata Service queries
│   │   └── mlxlink.go    # Mellanox network diagnostics
│   ├── level1_tests/     # Level 1 diagnostic test implementations
│   │   ├── gpu_count_check.go     # GPU count validation
│   │   ├── pcie_error_check.go    # PCIe error detection
│   │   └── rdma_nics_count.go     # RDMA NIC validation
│   ├── level2_tests/     # Level 2 diagnostic tests (placeholder)
│   ├── level3_tests/     # Level 3 diagnostic tests (placeholder)
│   ├── logger/           # Centralized logging system
│   │   └── logger.go     # Structured logging with configurable levels
│   ├── recommender/      # Intelligent recommendation system
│   │   ├── recommender.go# Multi-format recommendation analysis
│   │   └── config.go     # JSON-based recommendation configuration
│   ├── reporter/         # Test result reporting and output formatting
│   │   └── reporter.go   # Multi-format result reporting
│   └── shapes/           # OCI shape configuration management
│       ├── shapes.go     # Shape manager and query interface
│       ├── shapes.json   # Hardware shape definitions (development)
│       └── README.md     # Shapes package documentation
├── scripts/              # Installation and utility scripts
│   ├── setup-logging.sh # Log directory and permissions setup
│   └── BM.GPU.*/         # Shape-specific reference scripts
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── Makefile            # Build automation with FPM packaging
```

## 🏗️ Architecture Overview

### Core Packages

- **`cmd/`**: CLI interface using Cobra framework for command handling and subcommands
- **`internal/config/`**: Configuration management with Viper, supporting files and environment variables
- **`internal/executor/`**: System command execution layer with IMDS, nvidia-smi, lspci, and OS discovery
- **`internal/level1_tests/`**: Core diagnostic test implementations for GPU, PCIe, and RDMA
- **`internal/shapes/`**: OCI hardware shape definitions and query interface
- **`internal/autodiscover/`**: Hardware discovery with hybrid approach (shapes.json + runtime OS)
- **`internal/recommender/`**: JSON-configurable recommendation engine with fault codes
- **`internal/logger/`**: Structured logging with configurable output levels and debug visibility
- **`internal/reporter/`**: Multi-format result reporting (table, JSON, friendly)

### Configuration System

The application uses a sophisticated configuration system with the following priority order:

1. **CLI Flags** (highest priority)
2. **Environment Variables** (`OCI_DR_HPC_*` prefix)
3. **Configuration Files** (`/etc/oci-dr-hpc.yaml` or user-specified)
4. **Smart Defaults** (development vs production detection)

## 📦 Installation

### Development Installation
```bash
# Build and install for current user
make install-dev

# Binary installed to: ~/.local/bin/oci-dr-hpc-v2
# Config installed to: ~/.config/oci-dr-hpc/recommendations.json
```

### Production Installation
```bash
# Build and install system-wide
sudo make install

# Binary installed to: /usr/bin/oci-dr-hpc-v2
# Default config: /usr/share/oci-dr-hpc/recommendations.json  
# System config: /etc/oci-dr-hpc/recommendations.json
```

### Package Installation
```bash
# Build and install RPM package (auto-detected architecture)
make rpm
sudo rpm -i dist/oci-dr-hpc-v2-*.rpm

# Or build and install DEB package (auto-detected architecture)
make deb
sudo dpkg -i dist/oci-dr-hpc-v2-*.deb
```

### Cross-Platform Package Building
```bash
# Build packages for both AMD64 and ARM64 architectures
make all-cross      # Build everything for both architectures

# Build specific architecture packages
make rpm-amd64      # RPM for x86_64 systems
make rpm-arm64      # RPM for ARM64 systems

make deb-ubuntu-amd64    # Ubuntu DEB for x86_64
make deb-ubuntu-arm64    # Ubuntu DEB for ARM64

make deb-debian-amd64    # Debian DEB for x86_64  
make deb-debian-arm64    # Debian DEB for ARM64

# Build all packages for both architectures
make rpm-all        # All RPM packages
make deb-all        # All DEB packages
```

### Architecture-Specific Installation
```bash
# After building with make all-cross, install the appropriate package:

# For Intel/AMD x86_64 systems:
sudo rpm -i dist/oci-dr-hpc-v2-*-amd64-*.rpm
# or
sudo dpkg -i dist/oci-dr-hpc-v2-*-amd64*.deb

# For ARM64 systems (AWS Graviton, Apple Silicon Linux VMs):
sudo rpm -i dist/oci-dr-hpc-v2-*-arm64-*.rpm  
# or
sudo dpkg -i dist/oci-dr-hpc-v2-*-arm64*.deb

# Check your system architecture:
uname -m
# x86_64 = use amd64 packages
# aarch64 = use arm64 packages
```

### Build Targets
```bash
# Single architecture (auto-detected)
make build       # Build binary for detected architecture
make test        # Run tests
make clean       # Clean build artifacts
make uninstall   # Remove installation

# Cross-compilation
make build-amd64    # Build for x86_64 (Intel/AMD)
make build-arm64    # Build for ARM64 (AWS Graviton, Apple Silicon)
make build-all      # Build for both architectures

# Comprehensive help
make help          # Show all available targets
```

## ⚙️ Configuration

### File Locations

| Component | Development Path | Production Path | Purpose |
|-----------|-----------------|-----------------|---------|
| **Main Config** | `configs/oci-dr-hpc.yaml` | `/etc/oci-dr-hpc.yaml` | Application configuration |
| **Shapes Config** | `internal/shapes/shapes.json` | `/etc/oci-dr-hpc-shapes.json` | Hardware shape definitions |
| **Recommendations** | `configs/recommendations.json` | `/usr/share/oci-dr-hpc/recommendations.json` | Diagnostic recommendations with fault codes |
| **Binary** | `./oci-dr-hpc-v2` | `/usr/bin/oci-dr-hpc-v2` | Executable |
| **Logs** | Console/file | `/var/log/oci-dr-hpc/oci-dr-hpc.log` | Application logs |

### Smart Path Resolution

The application automatically resolves file paths using this logic:

```go
// For shapes.json file:
1. Check environment variable: OCI_DR_HPC_SHAPES_FILE
2. Check config file setting: shapes_file
3. Check production path: /etc/oci-dr-hpc-shapes.json (if exists)
4. Fall back to development path: internal/shapes/shapes.json (if exists)

// For recommendations.json file:
1. Check current directory: ./recommendations.json (highest priority override)
2. Check user config: ~/.config/oci-dr-hpc/recommendations.json
3. Check system config: /etc/oci-dr-hpc/recommendations.json
4. Check system data: /usr/share/oci-dr-hpc/recommendations.json
5. Check legacy location: /etc/oci-dr-hpc-recommendations.json
6. Fall back to development: configs/recommendations.json
```

### Environment Variables

Override any configuration setting with environment variables:

```bash
# Override shapes file location
export OCI_DR_HPC_SHAPES_FILE="/custom/path/shapes.json"

# Override logging configuration (enables config path visibility)
export OCI_DR_HPC_LOGGING_LEVEL="debug"
export OCI_DR_HPC_LOGGING_FILE="/custom/path/app.log"

# Override output format
export OCI_DR_HPC_OUTPUT="json"
export OCI_DR_HPC_VERBOSE="true"

# Debug configuration loading
export OCI_DR_HPC_LOGGING_LEVEL="debug"  # Shows config search paths
```

### Configuration File Format

```yaml
# /etc/oci-dr-hpc.yaml
verbose: false
output: table
level: L1

logging:
  file: "/var/log/oci-dr-hpc/oci-dr-hpc.log"
  level: "info"

# Shapes configuration file path
shapes_file: "/etc/oci-dr-hpc-shapes.json"
```

## 🚀 Usage

### Core Commands

```bash
# Run all Level 1 diagnostic tests
oci-dr-hpc level1

# Run specific tests
oci-dr-hpc level1 --test=gpu_count_check,rdma_nics_count

# List available tests
oci-dr-hpc level1 --list-tests

# Generate hardware discovery model with IMDS integration
oci-dr-hpc-v2 autodiscover

# Generate hardware discovery in different formats
oci-dr-hpc-v2 autodiscover --output json
oci-dr-hpc-v2 autodiscover --output table  
oci-dr-hpc-v2 autodiscover --output friendly

# Analyze test results and get recommendations with fault codes
oci-dr-hpc-v2 recommender -r results.json

# Get recommendations in different formats
oci-dr-hpc-v2 recommender -r results.json --output friendly
oci-dr-hpc-v2 recommender -r results.json --output json
oci-dr-hpc-v2 recommender -r results.json --output table

# Show version and build information
oci-dr-hpc-v2 --version
```

### Output Format Options

```bash
# Table format (default) - human-readable
oci-dr-hpc-v2 level1 --output=table

# JSON format - machine-readable
oci-dr-hpc-v2 level1 --output=json

# Friendly format - detailed human-readable
oci-dr-hpc-v2 level1 --output=friendly

# Save output to file (appends by default)
oci-dr-hpc-v2 level1 --output=json --output-file=results.json

# Append to existing file (default behavior)
oci-dr-hpc-v2 level1 --output=json --output-file=results.json --append

# Overwrite existing file
oci-dr-hpc-v2 level1 --output=json --output-file=results.json --append=false
```

### File Append Format

When using the `--append` flag (default behavior), the tool creates a JSON file with multiple test runs:

```json
{
  "test_runs": [
    {
      "run_id": "run_1704067200",
      "timestamp": "2024-01-01T10:00:00Z",
      "test_results": {
        "gpu_count_check": [
          {
            "status": "PASS",
            "gpu_count": 8,
            "timestamp_utc": "2024-01-01T10:00:00Z"
          }
        ],
        "pcie_error_check": [
          {
            "status": "PASS",
            "timestamp_utc": "2024-01-01T10:00:00Z"
          }
        ],
        "rdma_nics_count": [
          {
            "status": "PASS",
            "num_rdma_nics": 16,
            "timestamp_utc": "2024-01-01T10:00:00Z"
          }
        ]
      }
    },
    {
      "run_id": "run_1704070800",
      "timestamp": "2024-01-01T11:00:00Z",
      "test_results": {
        "gpu_count_check": [
          {
            "status": "PASS",
            "gpu_count": 8,
            "timestamp_utc": "2024-01-01T11:00:00Z"
          }
        ]
      }
    }
  ]
}
```

This format allows you to:
- **Track test history** over time
- **Compare results** between different runs
- **Analyze trends** in system health
- **Maintain historical records** for auditing

### Recommendations and Analysis

The recommender module analyzes test results and provides actionable recommendations for fixing issues:

```bash
# Analyze results and get recommendations with fault codes
oci-dr-hpc-v2 recommender -r results.json

# Works with both single and appended result formats
oci-dr-hpc-v2 recommender -r historical_results.json  # Uses latest run from appended format

# Debug configuration loading (shows where recommendations.json is loaded from)
oci-dr-hpc-v2 recommender -r results.json --verbose
```

#### Recommendation Types

| Type | Description | Example |
|------|-------------|---------|
| **Critical** 🚨 | Issues requiring immediate attention | GPU count mismatch, PCIe errors |
| **Warning** ⚠️ | Issues that should be addressed | RDMA NIC count discrepancy |
| **Info** ℹ️ | Informational status and suggestions | Successful test confirmations |

#### Sample Recommendation Output

```
======================================================================
🔍 HPC DIAGNOSTIC RECOMMENDATIONS
======================================================================

📊 SUMMARY: ⚠️ Found 2 issue(s) requiring attention: 1 critical, 1 warning
   • Total Issues: 2
   • Critical: 1
   • Warning: 1
   • Info: 1

----------------------------------------------------------------------
📋 DETAILED RECOMMENDATIONS
----------------------------------------------------------------------

🚨 1. CRITICAL [gpu_count_check]
   Fault Code: HPCGPU-0001-0001
   Issue: GPU count mismatch detected. Expected count not met (found: 6)
   Suggestion: Verify GPU hardware installation and driver status
   Commands to run:
     $ nvidia-smi
     $ lspci | grep -i nvidia
     $ dmesg | grep -i nvidia
     $ sudo nvidia-smi -pm 1
   References:
     - https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/
     - https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm

⚠️ 2. WARNING [rdma_nics_count]
   Fault Code: HPCGPU-0003-0001
   Issue: RDMA NIC count mismatch (found: 14)
   Suggestion: Verify RDMA hardware installation and driver configuration
   Commands to run:
     $ ibstat
     $ ibv_devices
     $ lspci | grep -i mellanox
     $ rdma link show
     $ systemctl status openibd
   References:
     - https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/configuringrdma.htm
     - https://docs.mellanox.com/display/MLNXOFEDv461000/
```

### Verbose and Debug Mode

```bash
# Enable verbose output
oci-dr-hpc-v2 level1 --verbose

# Enable debug logging (shows config loading paths)
oci-dr-hpc-v2 level1 --verbose
export OCI_DR_HPC_LOGGING_LEVEL="debug"

# See where recommendations.json is loaded from
oci-dr-hpc-v2 recommender -r results.json --verbose
# Output: INFO: Loading recommendation config from: /usr/share/oci-dr-hpc/recommendations.json
```

## 🧪 Available Diagnostic Tests

### Level 1 Tests (Production Ready)

| Test Name | Description | Checks | Fault Code |
|-----------|-------------|---------|------------|
| **`gpu_count_check`** | Verify GPU count matches shape specification | Uses nvidia-smi and shapes.json | HPCGPU-0001-0001 |
| **`pcie_error_check`** | Scan system logs for PCIe errors | Parses dmesg output for hardware errors | HPCGPU-0002-0001 |
| **`rdma_nics_count`** | Validate RDMA NIC count and PCI addresses | Uses hybrid discovery (shapes.json + OS) | HPCGPU-0003-0001 |

### Example Test Execution

```bash
# Run single test with verbose output
oci-dr-hpc-v2 level1 --test=gpu_count_check --verbose

# Output:
# INFO: === GPU Count Check ===
# INFO: Step 1: Getting shape from IMDS...
# INFO: Loading shapes configuration from: /etc/oci-dr-hpc-shapes.json
# INFO: Step 2: Getting expected GPU count from shapes.json...
# INFO: Expected GPU count for shape BM.GPU.H100.8: 8
# INFO: Step 3: Getting actual GPU count from nvidia-smi...
# INFO: Actual GPU count from nvidia-smi: 8
# INFO: GPU Count Check: PASS - Expected: 8, Actual: 8

# Run autodiscovery with verbose output (shows IMDS integration)
oci-dr-hpc-v2 autodiscover --verbose
# INFO: Discovering system information from IMDS and OS...
# INFO: Loading recommendation config from: /usr/share/oci-dr-hpc/recommendations.json
# INFO: Cluster detection: in_cluster=true (networkBlockId: ocid1.networkblock.oc1...)
```

## 🔧 Development

### Building

```bash
# Build for detected architecture using Makefile (recommended)
make build

# Cross-compilation using Makefile (recommended)
make build-amd64    # Build for x86_64
make build-arm64    # Build for ARM64
make build-all      # Build for both architectures

# Manual building for current platform
go build -o oci-dr-hpc-v2 main.go

# Manual cross-compilation (use Makefile instead)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=v1.0.0 -s -w" -o oci-dr-hpc-v2-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=v1.0.0 -s -w" -o oci-dr-hpc-v2-arm64 .
```

### Package Building

```bash
# Build packages for detected architecture
make rpm            # RPM package
make deb-ubuntu     # Ubuntu DEB package  
make deb-debian     # Debian DEB package

# Cross-platform package building
make rpm-all        # RPM packages for both architectures
make deb-all        # DEB packages for both architectures
make all-cross      # Everything for both architectures

# View all available build targets
make help
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/level1_tests/
go test -v ./internal/shapes/
go test -v ./internal/config/
go test -v ./internal/recommender/
go test -v ./internal/autodiscover/

# Run tests with race detection
go test -race ./...
```

### Adding New Tests

1. **Create test file** in `internal/levelX_tests/`:
   ```go
   func RunNewTest() error {
       logger.Info("=== New Test ===")
       // Implementation
       return nil
   }
   ```

2. **Add to test runner** in `cmd/levelX.go`:
   ```go
   {"new_test", level1_tests.RunNewTest},
   ```

3. **Update documentation** in README.md and relevant docs

### Code Structure Guidelines

- **Use structured logging**: `logger.Info()`, `logger.Error()`, `logger.Debug()`
- **Follow error handling**: Wrap errors with context using `fmt.Errorf()`
- **Use configuration system**: Access paths via `config.GetShapesFilePath()`
- **Test thoroughly**: Add unit tests for new functionality
- **Document changes**: Update README.md and package documentation

## 🔧 Troubleshooting

### Shapes File Issues

**Problem**: `no such file or directory: shapes.json`

**Solutions**:
```bash
# Install using Makefile
sudo make install  # Installs shapes.json to /etc/oci-dr-hpc-shapes.json

# Or set custom location
export OCI_DR_HPC_SHAPES_FILE="/path/to/shapes.json"
oci-dr-hpc-v2 level1

# Check file exists
ls -la /etc/oci-dr-hpc-shapes.json

# Debug shapes loading (with verbose mode)
oci-dr-hpc-v2 level1 --verbose
```

### IMDS Connection Issues

**Problem**: `IMDS request failed: context deadline exceeded`

**Cause**: Running outside Oracle Cloud Infrastructure or network connectivity issues

**Solution**: This is expected when running outside OCI. Tests will fail gracefully.

### Permission Issues

**Problem**: Permission denied errors for log files or config files

**Solutions**:
```bash
# Use setup script included in package installation
sudo ./scripts/setup-logging.sh

# Or manually fix log directory permissions
sudo mkdir -p /var/log/oci-dr-hpc
sudo chmod 755 /var/log/oci-dr-hpc

# Fix config file permissions
sudo chmod 644 /etc/oci-dr-hpc.yaml
sudo chmod 644 /etc/oci-dr-hpc-shapes.json
sudo chmod 644 /usr/share/oci-dr-hpc/recommendations.json

# Run with appropriate privileges
sudo oci-dr-hpc-v2 level1  # If system-level access needed
```

### GPU Detection Issues

**Problem**: `nvidia-smi not available` or GPU count mismatch

**Solutions**:
```bash
# Check NVIDIA drivers
nvidia-smi

# Check if running on correct instance type (shows IMDS integration)
oci-dr-hpc-v2 autodiscover  # See detected hardware with cluster detection

# Verify shapes configuration
cat /etc/oci-dr-hpc-shapes.json | grep -A 10 "BM.GPU.H100.8"

# Debug GPU detection with verbose output
oci-dr-hpc-v2 level1 --test=gpu_count_check --verbose
```

### Configuration Issues

**Problem**: Recommendations not loading or showing unexpected behavior

**Solutions**:
```bash
# Debug configuration loading (shows where config is loaded from)
oci-dr-hpc-v2 recommender -r results.json --verbose
# Output: INFO: Loading recommendation config from: /usr/share/oci-dr-hpc/recommendations.json

# Check if custom config exists
ls -la ./recommendations.json ~/.config/oci-dr-hpc/recommendations.json

# Validate JSON syntax
python -m json.tool /usr/share/oci-dr-hpc/recommendations.json

# Test with debug logging to see search paths
export OCI_DR_HPC_LOGGING_LEVEL="debug"
oci-dr-hpc-v2 recommender -r results.json
```

## 📚 Additional Documentation

- **[Recommender System](docs/recommender-system.md)**: Complete guide to the intelligent diagnostic recommendation engine
- **[Autodiscovery Algorithm](docs/autodiscovery.md)**: Comprehensive guide to hardware discovery (@rekharoy)
- **[Recommendations Configuration](docs/recommendations-config.md)**: JSON-based recommendation system configuration
- **[Deployment Guide](docs/deployment.md)**: Complete customer deployment instructions
- **[Host Metadata and IMDS Documentation](docs/host_metadata_and_imds.md)**: Instance Metadata Service integration
- **[Shapes Package](internal/shapes/README.md)**: Hardware shape configuration management

## 🤝 Contributing

1. **Environment Setup**: Ensure Go 1.21.5+ is installed
2. **Code Quality**: Run `go test ./...` and `go vet ./...` before submitting
3. **Formatting**: Use `go fmt ./...` to format code
4. **Documentation**: Update README.md and relevant documentation for changes
5. **Testing**: Add unit tests for new functionality
6. **Dependencies**: Minimize external dependencies, prefer standard library

## 📄 License

Oracle Cloud Infrastructure Diagnostic and Repair for HPC v2

## 📧 Support

For support and questions:
- **Technical Issues**: Create an issue in the repository
- **Oracle Support**: [bob.r.booth@oracle.com](mailto:bob.r.booth@oracle.com)

---

**Version**: Development
**Go Version**: 1.21.5+
**Platforms**: Linux (Oracle Linux, Ubuntu, Debian, RHEL)
**Architectures**: AMD64 (x86_64), ARM64 (aarch64)
**Cross-Compilation**: Supported for both architectures from any build system