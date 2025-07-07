# OCI DR HPC v2

A comprehensive diagnostic and repair tool for High Performance Computing (HPC) environments with GPU and RDMA support on Oracle Cloud Infrastructure (OCI).

## 🚀 Features

- **🎮 GPU Diagnostics**: Check GPU count, driver status, and hardware health using nvidia-smi
- **🔗 RDMA Network Testing**: Validate RDMA NIC count, PCI addresses, and connectivity  
- **⚡ PCIe Error Detection**: Scan system logs for PCIe-related hardware errors
- **🔍 Hardware Autodiscovery**: Generate logical hardware models automatically from system detection
- **📊 Multiple Output Formats**: Support for table, JSON, and friendly human-readable output
- **⚙️ Flexible Configuration**: Support for config files, environment variables, and CLI flags
- **🏗️ Smart Path Resolution**: Automatic detection of development vs production environments
- **📦 Customer-Ready Deployment**: Installation scripts and system-wide configuration support

## 📁 Project Structure

```
oci-dr-hpc-v2/
├── cmd/                    # CLI command definitions (Cobra framework)
│   ├── root.go            # Main CLI entry point and config initialization
│   ├── level1.go          # Level 1 diagnostic commands
│   └── autodiscover.go    # Hardware autodiscovery commands
├── config/                # Configuration files
│   └── oci-dr-hpc.yaml   # Default configuration template
├── docs/                  # Documentation
│   ├── deployment.md     # Customer deployment guide
│   ├── imds.md          # IMDS (Instance Metadata Service) documentation
│   └── *.md             # Additional documentation
├── internal/              # Internal application logic
│   ├── autodiscover/     # Hardware discovery and modeling
│   │   ├── gpu_discovery.go      # GPU detection logic
│   │   ├── network_discovery.go  # Network/RDMA discovery
│   │   └── system_info.go        # System information gathering
│   ├── config/           # Configuration management (Viper integration)
│   │   └── config.go     # Config loading with smart path resolution
│   ├── executor/         # System command execution
│   │   ├── nvidia_smi.go # NVIDIA GPU command execution
│   │   ├── os_commands.go # OS-level commands (lspci, dmesg)
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
│   ├── reporter/         # Test result reporting and output formatting
│   │   └── reporter.go   # Multi-format result reporting
│   └── shapes/           # OCI shape configuration management
│       ├── shapes.go     # Shape manager and query interface
│       ├── shapes.json   # Hardware shape definitions (development)
│       └── README.md     # Shapes package documentation
├── scripts/              # Installation and utility scripts
│   ├── install-shapes.sh # Shapes file installation script
│   └── BM.GPU.*/         # Shape-specific reference scripts
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── Makefile            # Build automation
```

## 🏗️ Architecture Overview

### Core Packages

- **`cmd/`**: CLI interface using Cobra framework for command handling
- **`internal/config/`**: Configuration management with Viper, supporting files and environment variables
- **`internal/executor/`**: System command execution layer (nvidia-smi, lspci, IMDS, etc.)
- **`internal/level1_tests/`**: Core diagnostic test implementations
- **`internal/shapes/`**: OCI hardware shape definitions and query interface
- **`internal/autodiscover/`**: Hardware discovery and logical model generation
- **`internal/logger/`**: Structured logging with configurable output levels
- **`internal/reporter/`**: Multi-format result reporting (table, JSON, friendly)

### Configuration System

The application uses a sophisticated configuration system with the following priority order:

1. **CLI Flags** (highest priority)
2. **Environment Variables** (`OCI_DR_HPC_*` prefix)
3. **Configuration Files** (`/etc/oci-dr-hpc.yaml` or user-specified)
4. **Smart Defaults** (development vs production detection)

## 📦 Installation

### For Customers (Production Deployment)

#### 1. Binary Installation
```bash
# Install the main binary
sudo cp oci-dr-hpc /usr/local/bin/
sudo chmod +x /usr/local/bin/oci-dr-hpc
```

#### 2. Configuration Files
```bash
# Install main configuration
sudo cp config/oci-dr-hpc.yaml /etc/oci-dr-hpc.yaml

# Install shapes configuration using provided script
sudo ./scripts/install-shapes.sh

# Or manually install shapes file
sudo cp internal/shapes/shapes.json /etc/oci-dr-hpc-shapes.json
sudo chmod 644 /etc/oci-dr-hpc-shapes.json
```

#### 3. System Directories
```bash
# Create log directory
sudo mkdir -p /var/log/oci-dr-hpc
sudo chmod 755 /var/log/oci-dr-hpc

# Verify installation
oci-dr-hpc --version
```

### For Developers

The tool automatically detects environments and prioritizes production paths:

```bash
# Clone and build
git clone <repository-url>
cd oci-dr-hpc-v2
go build -o oci-dr-hpc main.go

# Run directly (uses /etc/oci-dr-hpc-shapes.json if exists, falls back to internal/shapes/shapes.json)
./oci-dr-hpc level1
```

## ⚙️ Configuration

### File Locations

| Component | Development Path | Production Path | Purpose |
|-----------|-----------------|-----------------|---------|
| **Main Config** | `config/oci-dr-hpc.yaml` | `/etc/oci-dr-hpc.yaml` | Application configuration |
| **Shapes Config** | `internal/shapes/shapes.json` | `/etc/oci-dr-hpc-shapes.json` | Hardware shape definitions |
| **Binary** | `./oci-dr-hpc` | `/usr/local/bin/oci-dr-hpc` | Executable |
| **Logs** | Console/file | `/var/log/oci-dr-hpc/oci-dr-hpc.log` | Application logs |

### Smart Path Resolution

The application automatically resolves file paths using this logic:

```go
// For shapes.json file:
1. Check environment variable: OCI_DR_HPC_SHAPES_FILE
2. Check config file setting: shapes_file
3. Check production path: /etc/oci-dr-hpc-shapes.json (if exists)
4. Fall back to development path: internal/shapes/shapes.json (if exists)
```

### Environment Variables

Override any configuration setting with environment variables:

```bash
# Override shapes file location
export OCI_DR_HPC_SHAPES_FILE="/custom/path/shapes.json"

# Override logging configuration
export OCI_DR_HPC_LOGGING_LEVEL="debug"
export OCI_DR_HPC_LOGGING_FILE="/custom/path/app.log"

# Override output format
export OCI_DR_HPC_OUTPUT="json"
export OCI_DR_HPC_VERBOSE="true"
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

# Generate hardware discovery model
oci-dr-hpc autodiscover

# Show version and build information
oci-dr-hpc --version
```

### Output Format Options

```bash
# Table format (default) - human-readable
oci-dr-hpc level1 --output=table

# JSON format - machine-readable
oci-dr-hpc level1 --output=json

# Friendly format - detailed human-readable
oci-dr-hpc level1 --output=friendly

# Save output to file
oci-dr-hpc level1 --output=json --output-file=results.json
```

### Verbose and Debug Mode

```bash
# Enable verbose output
oci-dr-hpc level1 --verbose

# Enable debug logging
oci-dr-hpc level1 --verbose
export OCI_DR_HPC_LOGGING_LEVEL="debug"
```

## 🧪 Available Diagnostic Tests

### Level 1 Tests (Production Ready)

| Test Name | Description | Checks |
|-----------|-------------|---------|
| **`gpu_count_check`** | Verify GPU count matches shape specification | Uses nvidia-smi and shapes.json |
| **`pcie_error_check`** | Scan system logs for PCIe errors | Parses dmesg output for hardware errors |
| **`rdma_nics_count`** | Validate RDMA NIC count and PCI addresses | Uses lspci and shapes.json |

### Example Test Execution

```bash
# Run single test with verbose output
oci-dr-hpc level1 --test=gpu_count_check --verbose

# Output:
# INFO: === GPU Count Check ===
# INFO: Step 1: Getting shape from IMDS...
# INFO: Loading shapes configuration from: /etc/oci-dr-hpc-shapes.json
# INFO: Step 2: Getting expected GPU count from shapes.json...
# INFO: Expected GPU count for shape BM.GPU.H100.8: 8
# INFO: Step 3: Getting actual GPU count from nvidia-smi...
# INFO: Actual GPU count from nvidia-smi: 8
# INFO: GPU Count Check: PASS - Expected: 8, Actual: 8
```

## 🔧 Development

### Building

```bash
# Build for current platform
go build -o oci-dr-hpc main.go

# Build with version information
go build -ldflags "-X github.com/oracle/oci-dr-hpc-v2/cmd.version=v1.0.0" -o oci-dr-hpc main.go

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o oci-dr-hpc-linux-amd64 main.go
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
# Check current shapes file location
oci-dr-hpc test-shapes  # (if test command available)

# Install shapes file using script
sudo ./scripts/install-shapes.sh

# Or set custom location
export OCI_DR_HPC_SHAPES_FILE="/path/to/shapes.json"
oci-dr-hpc level1

# Or check file exists
ls -la /etc/oci-dr-hpc-shapes.json
```

### IMDS Connection Issues

**Problem**: `IMDS request failed: context deadline exceeded`

**Cause**: Running outside Oracle Cloud Infrastructure or network connectivity issues

**Solution**: This is expected when running outside OCI. Tests will fail gracefully.

### Permission Issues

**Problem**: Permission denied errors for log files or config files

**Solutions**:
```bash
# Fix log directory permissions
sudo mkdir -p /var/log/oci-dr-hpc
sudo chmod 755 /var/log/oci-dr-hpc

# Fix config file permissions
sudo chmod 644 /etc/oci-dr-hpc.yaml
sudo chmod 644 /etc/oci-dr-hpc-shapes.json

# Run with appropriate privileges
sudo oci-dr-hpc level1  # If system-level access needed
```

### GPU Detection Issues

**Problem**: `nvidia-smi not available` or GPU count mismatch

**Solutions**:
```bash
# Check NVIDIA drivers
nvidia-smi

# Check if running on correct instance type
oci-dr-hpc autodiscover  # See detected hardware

# Verify shapes configuration
cat /etc/oci-dr-hpc-shapes.json | grep -A 10 "BM.GPU.H100.8"
```

## 📚 Additional Documentation

- **[Deployment Guide](docs/deployment.md)**: Complete customer deployment instructions
- **[Host Metadata and IMDS Documentation](docs/host_metadata_and_imds.md)**: Instance Metadata Service integration and host metadata details
- **[Shapes Package](internal/shapes/README.md)**: Hardware shape configuration management
- **[Installation Notes](docs/installation_notes_*.md)**: OS-specific installation guides

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
**Platforms**: Linux (Oracle Linux, Ubuntu)
**Architecture**: x86_64