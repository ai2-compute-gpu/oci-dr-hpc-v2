# oci-dr-hpc-v2

Oracle Cloud Infrastructure Diagnostic and Repair tool for HPC environments with GPU and RDMA support.

## Overview

A Go-based CLI application that performs diagnostic and repair operations for HPC environments, specifically designed for Oracle Cloud Infrastructure GPU and RDMA configurations.

## OCI Support Email
For support and issues, please contact: [bob.r.booth@oracle.com](mailto:bob.r.booth@oracle.com)

## Project Structure

```
├── cmd/           # CLI command definitions (Cobra framework)
├── config/        # Configuration files
├── internal/      # Internal application logic
│   ├── config/    # Configuration management (Viper)
│   ├── executor/  # Command execution (nvidia-smi, lspci, dmesg)
│   ├── level1_tests/ # Level 1 diagnostic tests
│   └── logger/    # Custom logging implementation
├── scripts/       # Shell scripts for different GPU shapes
│   ├── BM.GPU.B200.8/
│   ├── BM.GPU.GB200.4/
│   ├── BM.GPU.H100.8/
│   └── BM.GPU.H200.8/
└── Makefile       # Build automation
```

## Requirements

### System Requirements
- **Operating Systems**: Oracle Linux 9.5, Ubuntu 22.04
- **Go Version**: 1.21.5 or higher
- **Architecture**: x86_64

### Build Dependencies
- `git`
- `make` 
- `golang` (1.21.5+)

### Runtime Dependencies
- `nvidia-smi` (for GPU diagnostics)
- `lspci` (for PCIe diagnostics)
- `dmesg` (for system message diagnostics)
- `sudo` access (for system-level operations)

## Installation

### From Source

1. **Install Dependencies**
   ```bash
   # Oracle Linux/RHEL
   sudo dnf update && sudo dnf install -y git make golang
   
   # Ubuntu/Debian  
   sudo apt update && sudo apt install -y git make golang-go
   ```

2. **Clone and Build**
   ```bash
   git clone https://github.com/oracle/oci-dr-hpc-v2.git
   cd oci-dr-hpc-v2
   make build
   ```

3. **Install Binary**
   ```bash
   sudo cp build/oci-dr-hpc-v2 /usr/bin/
   sudo cp config/oci-dr-hpc.yaml /etc/
   ```

### From Package

#### RPM Package (Oracle Linux/RHEL)
```bash
# Build RPM
make rpm

# Install
sudo rpm -i dist/oci-dr-hpc-v2-*.rpm
```

#### DEB Package (Ubuntu/Debian)
```bash
# Build DEB
make deb

# Install
sudo dpkg -i dist/oci-dr-hpc-v2_*.deb
```

## Usage

### Basic Commands
```bash
# Show help
oci-dr-hpc-v2 --help

# Run with verbose output
oci-dr-hpc-v2 --verbose

# Set test level and output format
oci-dr-hpc-v2 --level L2 --output json

# Use custom config file
oci-dr-hpc-v2 --config /path/to/config.yaml
```

### Available Flags
- `--config string`: Configuration file path
- `--level string`: Test level (L1|L2|L3), default: L1
- `--output string`: Output format (json|table|friendly), default: table
- `--verbose`: Enable verbose output
- `--version`: Show version information

## Configuration

Configuration priority (highest to lowest):
1. CLI flags
2. Environment variables
3. Configuration file
4. Default values

### Configuration File Locations
- `/etc/oci-dr-hpc.yaml` (system-wide)
- `~/.oci-dr-hpc.yaml` (user-specific)
- Custom path via `--config` flag

### Example Configuration
```yaml
# /etc/oci-dr-hpc.yaml
verbose: false
output: table
level: L1

logging:
  level: "info"
  file: "/var/log/oci-dr-hpc/oci-dr-hpc.log"
```

### Environment Variables

Override any configuration using environment variables with `OCI_DR_HPC_` prefix:

```bash
# Logging control
export OCI_DR_HPC_LOGGING_LEVEL=debug    # debug|info|error
export OCI_DR_HPC_LOGGING_FILE=/tmp/debug.log

# Application settings
export OCI_DR_HPC_VERBOSE=true
export OCI_DR_HPC_OUTPUT=json            # json|table|friendly
export OCI_DR_HPC_LEVEL=L2               # L1|L2|L3

# Run with environment overrides
oci-dr-hpc-v2
```

### Log Levels
- **debug**: All messages (INFO, ERROR, DEBUG)
- **info**: INFO and ERROR messages only (filters DEBUG)
- **error**: ERROR messages only

## Development

### Build Commands
```bash
# Clean, test, and build
make all

# Build only
make build

# Run tests
make test

# Generate coverage report
make coverage

# Clean build artifacts
make clean
```

### Testing
```bash
# Run all unit tests
go test -v ./...

# Run tests with coverage
make coverage
# Opens coverage.html in build/ directory
```

### Project Architecture
- **Cobra**: CLI framework for command handling
- **Viper**: Configuration management with file and environment support
- **Custom Logger**: Structured logging with configurable levels
- **Modular Design**: Separate packages for config, logging, execution, and tests

## Scripts Directory

The `scripts/` directory contains shell scripts organized by OCI GPU shape types:
- `BM.GPU.B200.8/`: Scripts for B200 GPU shapes
- `BM.GPU.GB200.4/`: Scripts for GB200 GPU shapes  
- `BM.GPU.H100.8/`: Scripts for H100 GPU shapes
- `BM.GPU.H200.8/`: Scripts for H200 GPU shapes

These scripts provide alternative implementations and reference tests for the Go-based CLI functionality.

## Contributing

1. Ensure Go 1.21.5+ is installed
2. Run `make test` before submitting changes
3. Follow existing code patterns and conventions
4. Add unit tests for new functionality

@rekharoy