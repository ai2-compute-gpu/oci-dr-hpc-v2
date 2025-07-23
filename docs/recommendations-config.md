# HPC Recommendations Configuration System

## Overview

The OCI DR HPC v2 recommender supports a flexible, JSON-based configuration system for defining diagnostic recommendations. This allows you to customize recommendations for each test type without modifying the source code.

## Configuration File Locations

The recommender looks for configuration files in the following order (first found wins):

1. `./recommendations.json` (current directory override - highest priority)
2. `~/.config/oci-dr-hpc/recommendations.json` (user-specific config)
3. `/etc/oci-dr-hpc/recommendations.json` (system configuration)
4. `/usr/share/oci-dr-hpc/recommendations.json` (default system config)
5. `/etc/oci-dr-hpc-recommendations.json` (legacy location)
6. `configs/recommendations.json` (development location)

This hierarchy allows for:
- **Local overrides**: Project-specific configs in working directory
- **User customization**: Personal configs in home directory
- **System administration**: System-wide configs in /etc
- **Package defaults**: Immutable defaults in /usr/share

## Configuration File Format

### Basic Structure

```json
{
  "recommendations": {
    "test_name": {
      "fail": {
        "type": "critical|warning|info",
        "fault_code": "HPCGPU-XXXX-XXXX",
        "issue": "Description of the issue",
        "suggestion": "Suggested remediation steps",
        "commands": ["command1", "command2"],
        "references": ["url1", "url2"]
      },
      "pass": {
        "type": "info",
        "issue": "Description when test passes",
        "suggestion": "What this means",
        "commands": ["verification commands"]
      }
    }
  },
  "summary_templates": {
    "no_issues": "Message when all tests pass",
    "has_issues": "Message when issues found with {variable} substitution"
  }
}
```

### Currently Supported Test Names

The following test names are currently implemented and supported:

- `gpu_count_check` - GPU hardware detection and count verification
- `gpu_clk_check` - GPU clock speed validation with threshold checking
- `pcie_error_check` - PCIe bus error detection
- `rdma_nics_count` - RDMA/InfiniBand NIC count verification

### Fault Code System

Fault codes follow the pattern: `HPCGPU-XXXX-XXXX`

- **HPCGPU**: Product identifier for HPC GPU diagnostics
- **First XXXX**: Test category (0001=GPU, 0002=PCIe, 0003=RDMA)
- **Second XXXX**: Specific error type (0001=count mismatch, etc.)

#### Current Fault Codes

| Fault Code | Test | Description |
|------------|------|-------------|
| `HPCGPU-0001-0001` | gpu_count_check | GPU count mismatch |
| `HPCGPU-0002-0001` | pcie_error_check | PCIe errors detected |
| `HPCGPU-0003-0001` | rdma_nics_count | RDMA NIC count mismatch |
| `HPCGPU-0011-0001` | gpu_clk_check | GPU clock speeds below threshold |

### Variable Substitution

The following variables can be used in `issue` and `suggestion` fields:

- `{gpu_count}` - Number of GPUs detected
- `{clock_speed}` - GPU clock speed information for gpu_clk_check
- `{num_rdma_nics}` - Number of RDMA NICs detected
- `{total_issues}` - Total number of issues (summary only)
- `{critical_count}` - Number of critical issues (summary only)
- `{warning_count}` - Number of warning issues (summary only)

### Recommendation Types

- **critical** - Issues that prevent proper system operation
- **warning** - Issues that may impact performance but don't prevent operation
- **info** - Informational messages about successful tests

## Current Configuration Example

The current `configs/recommendations.json` contains:

```json
{
  "recommendations": {
    "gpu_count_check": {
      "fail": {
        "type": "critical",
        "fault_code": "HPCGPU-0001-0001",
        "issue": "GPU count mismatch detected. Expected count not met (found: {gpu_count})",
        "suggestion": "Verify GPU hardware installation and driver status",
        "commands": [
          "nvidia-smi",
          "lspci | grep -i nvidia",
          "dmesg | grep -i nvidia",
          "sudo nvidia-smi -pm 1"
        ],
        "references": [
          "https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/",
          "https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm"
        ]
      },
      "pass": {
        "type": "info",
        "issue": "GPU count check passed ({gpu_count} GPUs detected)",
        "suggestion": "GPU hardware is properly detected and configured",
        "commands": [
          "nvidia-smi -q",
          "nvidia-smi topo -m"
        ]
      }
    },
    "pcie_error_check": {
      "fail": {
        "type": "critical",
        "fault_code": "HPCGPU-0002-0001",
        "issue": "PCIe errors detected in system logs",
        "suggestion": "Check PCIe bus health and reseat hardware if necessary",
        "commands": [
          "dmesg | grep -i pcie",
          "dmesg | grep -i 'corrected error'",
          "dmesg | grep -i 'uncorrectable error'",
          "lspci -tv",
          "sudo pcieport-error-inject"
        ],
        "references": [
          "https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm",
          "https://www.kernel.org/doc/Documentation/PCI/pci-error-recovery.txt"
        ]
      },
      "pass": {
        "type": "info",
        "issue": "PCIe error check passed",
        "suggestion": "PCIe bus appears healthy with no errors detected",
        "commands": [
          "lspci -tv",
          "dmesg | tail -50"
        ]
      }
    },
    "gpu_clk_check": {
      "fail": {
        "type": "critical",
        "fault_code": "HPCGPU-0011-0001",
        "issue": "GPU clock speeds below acceptable threshold ({clock_speed})",
        "suggestion": "Verify GPU performance state and check for thermal throttling",
        "commands": [
          "nvidia-smi --query-gpu=clocks.current.graphics --format=csv,noheader,nounits",
          "nvidia-smi -q -d CLOCK",
          "nvidia-smi --query-gpu=temperature.gpu,power.draw --format=csv,noheader",
          "nvidia-smi --query-gpu=pstate --format=csv,noheader"
        ],
        "references": [
          "https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/",
          "https://developer.nvidia.com/nvidia-system-management-interface"
        ]
      },
      "pass": {
        "type": "info",
        "issue": "GPU clock speed check passed ({clock_speed})",
        "suggestion": "GPU clock speeds are within acceptable range",
        "commands": [
          "nvidia-smi --query-gpu=clocks.current.graphics --format=csv,noheader,nounits",
          "nvidia-smi -q -d CLOCK"
        ]
      }
    },
    "rdma_nics_count": {
      "fail": {
        "type": "warning",
        "fault_code": "HPCGPU-0003-0001",
        "issue": "RDMA NIC count mismatch (found: {num_rdma_nics})",
        "suggestion": "Verify RDMA hardware installation and driver configuration",
        "commands": [
          "ibstat",
          "ibv_devices",
          "lspci | grep -i mellanox",
          "rdma link show",
          "systemctl status openibd"
        ],
        "references": [
          "https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/configuringrdma.htm",
          "https://docs.mellanox.com/display/MLNXOFEDv461000/"
        ]
      },
      "pass": {
        "type": "info",
        "issue": "RDMA NIC count check passed ({num_rdma_nics} NICs detected)",
        "suggestion": "RDMA hardware is properly detected and configured",
        "commands": [
          "ibstat",
          "ibv_devinfo",
          "rdma link show"
        ]
      }
    }
  },
  "summary_templates": {
    "no_issues": "All diagnostic tests passed. Your HPC environment appears healthy.",
    "has_issues": "Found {total_issues} issue(s) requiring attention: {critical_count} critical, {warning_count} warning"
  }
}
```

## Installation and Setup

### Building and Installing

The application provides multiple installation methods:

#### Development Installation
```bash
# Build and install for current user
make install-dev
```

#### System-wide Installation
```bash
# Build and install system-wide (requires sudo)
make install
```

#### Package Installation
```bash
# Build RPM package
make rpm

# Build DEB packages
make deb-ubuntu    # Ubuntu-specific package
make deb-debian    # Debian-specific package

# Then install with package manager
sudo rpm -i dist/oci-dr-hpc-v2-*.rpm
# or
sudo dpkg -i dist/oci-dr-hpc-v2-*.deb
```

#### Other Useful Targets
```bash
# Just build the binary
make build

# Run tests
make test

# Run tests with coverage
make coverage

# Clean build artifacts
make clean

# Remove installation
make uninstall
```

### Installed File Locations

After installation, files are placed according to the Filesystem Hierarchy Standard:

- **Binary**: `/usr/bin/oci-dr-hpc-v2` (system) or `~/.local/bin/oci-dr-hpc-v2` (user)
- **Default config**: `/usr/share/oci-dr-hpc/recommendations.json`
- **System config**: `/etc/oci-dr-hpc/recommendations.json`
- **User config**: `~/.config/oci-dr-hpc/recommendations.json`

## Using Custom Configurations

### 1. Create Your Configuration File

Create a `recommendations.json` file with your custom recommendations:

```bash
# Copy the default configuration
cp configs/recommendations.json my-custom-recommendations.json

# Edit to customize
vi my-custom-recommendations.json
```

### 2. Place Configuration File

Option A: Use current directory (highest priority)
```bash
cp my-custom-recommendations.json ./recommendations.json
```

Option B: Use user-specific location
```bash
mkdir -p ~/.config/oci-dr-hpc
cp my-custom-recommendations.json ~/.config/oci-dr-hpc/recommendations.json
```

Option C: Use system-wide location
```bash
sudo mkdir -p /etc/oci-dr-hpc
sudo cp my-custom-recommendations.json /etc/oci-dr-hpc/recommendations.json
```

### 3. Run Recommender

```bash
# The recommender will automatically load your custom configuration
oci-dr-hpc-v2 recommender -r test_results.json --output friendly
```

## Adding New Test Types

To add support for a new test type:

### 1. Add Configuration Entry

```json
{
  "recommendations": {
    "your_new_test": {
      "fail": {
        "type": "warning",
        "fault_code": "HPCGPU-0004-0001",
        "issue": "Your test failed with {custom_variable}",
        "suggestion": "Here's how to fix it",
        "commands": ["diagnostic-command", "fix-command"],
        "references": ["https://docs.example.com/"]
      },
      "pass": {
        "type": "info",
        "issue": "Your test passed",
        "suggestion": "Everything looks good"
      }
    }
  }
}
```

### 2. Update Variable Substitution (if needed)

If your test requires new variables, modify the `applyVariableSubstitution` function in `internal/recommender/config.go`:

```go
// Add new variable substitutions
result = strings.ReplaceAll(result, "{custom_variable}", fmt.Sprintf("%d", testResult.CustomField))
```

### 3. Update Test Processing

Add your test to the `testMappings` in `generateRecommendations` function in `internal/recommender/recommender.go`:

```go
testMappings := []struct {
    testName string
    results  []TestResult
}{
    {"gpu_count_check", results.GPUCountCheck},
    {"pcie_error_check", results.PCIeErrorCheck},
    {"rdma_nics_count", results.RDMANicsCount},
    {"your_new_test", results.YourNewTest}, // Add this line
}
```

## Output Formats

The recommendation system supports multiple output formats:

### JSON Output
```bash
oci-dr-hpc-v2 recommender -r results.json --output json
```

Produces structured JSON with all recommendation details.

### Table Output
```bash
oci-dr-hpc-v2 recommender -r results.json --output table
```

Produces formatted table output for reports.

### Friendly Output
```bash
oci-dr-hpc-v2 recommender -r results.json --output friendly
```

Produces user-friendly output with emojis and clear formatting.

## Fallback Mode

If the configuration file cannot be loaded, the recommender automatically falls back to built-in recommendations. This ensures the system continues to work even if the configuration is missing or corrupted.

Fallback mode indicators:
- Log message: "Using fallback recommendations due to config load failure"
- Summary includes "(fallback mode)"
- Reduced set of commands and references

## Usage Examples

### Basic Usage

```bash
# Run Level 1 diagnostic tests
oci-dr-hpc-v2 level1

# Run specific tests
oci-dr-hpc-v2 level1 --test=gpu_count_check,gpu_clk_check,rdma_nics_count

# List available tests
oci-dr-hpc-v2 level1 --list-tests

# Generate hardware discovery
oci-dr-hpc-v2 autodiscover

# Analyze test results
oci-dr-hpc-v2 recommender -r results.json

# Specify output format
oci-dr-hpc-v2 recommender -r results.json --output friendly
oci-dr-hpc-v2 recommender -r results.json --output json
oci-dr-hpc-v2 recommender -r results.json --output table
```

### Debug Configuration

```bash
# See where configuration is loaded from
oci-dr-hpc-v2 recommender -r results.json --verbose

# Debug configuration search paths
export OCI_DR_HPC_LOGGING_LEVEL="debug"
oci-dr-hpc-v2 recommender -r results.json
```

## Best Practices

### 1. Version Control Configuration

Keep your recommendation configurations in version control:

```bash
git add configs/recommendations.json
git commit -m "Add custom HPC recommendations"
```

### 2. Test Configuration Changes

Always test configuration changes with sample data:

```bash
# Test with sample results
oci-dr-hpc-v2 recommender -r sample_results.json --output friendly
```

### 3. Validate JSON Syntax

Use a JSON validator to ensure your configuration is valid:

```bash
# Validate syntax
python -m json.tool configs/recommendations.json > /dev/null && echo "Valid JSON"
```

### 4. Document Custom Tests

Document any custom test types you add:

```json
{
  "_comment": "Custom test for checking network bandwidth",
  "recommendations": {
    "network_bandwidth_check": {
      // ... configuration
    }
  }
}
```

## Available Test Scripts

The system includes comprehensive test scripts in the `scripts/` directory, organized by OCI shape:

### BM.GPU.H100.8 Scripts
- `gpu_count_check.py/sh` - GPU hardware detection
- `gpu_driver_check.py/sh` - GPU driver verification
- `gpu_clk_check.py/sh` - GPU clock validation
- `gpu_mode_check.py/sh` - GPU mode checks
- `pcie_error_check.py/sh` - PCIe error detection
- `rdma_nic_count_check.py/sh` - RDMA NIC validation
- `rx_discards_check.py/sh` - Network RX discards monitoring
- `gid_index_check.py/sh` - GID index validation
- `max_acc_check.py/sh` - Max accelerator checks
- `sram_error_check.py/sh` - SRAM validation

These scripts provide detailed diagnostics and can be referenced or integrated into the recommendation system.

## Troubleshooting

### Configuration Not Loading

1. Check file exists and is readable
2. Verify JSON syntax with `python -m json.tool file.json`
3. Check log messages for specific error details
4. Ensure file permissions allow reading

### Missing Recommendations

1. Verify test name matches exactly (case-sensitive)
2. Check that both "fail" and "pass" sections exist
3. Ensure test results use correct status values ("PASS", "FAIL")

### Variable Substitution Not Working

1. Verify variable names are correct (case-sensitive)
2. Check that test results contain the expected fields
3. Ensure variables are enclosed in curly braces: `{variable_name}`

## Advanced Features

### Multi-format Result Support

The recommender supports both single-run and multi-run result formats:

```bash
# Works with single run results
oci-dr-hpc-v2 recommender -r single_run_results.json

# Works with appended multi-run results (uses latest run)
oci-dr-hpc-v2 recommender -r historical_results.json
```

### Integration with CI/CD

Include recommendation validation in your CI/CD pipeline:

```bash
# Validate configuration syntax
python -m json.tool configs/recommendations.json

# Test with sample data
oci-dr-hpc-v2 recommender -r tests/sample_results.json --output json > /dev/null
```

This ensures configuration changes don't break the recommendation system.