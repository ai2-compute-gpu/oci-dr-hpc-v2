# HPC Recommendations Configuration System

## Overview

The OCI DR HPC v2 recommender now supports a flexible, JSON-based configuration system for defining diagnostic recommendations. This allows you to customize recommendations for each test type without modifying the source code.

## Configuration File Locations

The recommender looks for configuration files in the following order (first found wins):

1. `~/.config/oci-dr-hpc/recommendations.json` (user-specific config)
2. `./recommendations.json` (current directory override)
3. `/etc/oci-dr-hpc/recommendations.json` (system configuration)
4. `/usr/share/oci-dr-hpc/recommendations.json` (default system config)
5. `/etc/oci-dr-hpc-recommendations.json` (legacy location)
6. `configs/recommendations.json` (development location)

This hierarchy allows for:
- **User customization**: Personal configs in home directory
- **Project overrides**: Local configs in working directory
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

### Supported Test Names

- `gpu_count_check` - GPU hardware detection and count verification
- `pcie_error_check` - PCIe bus error detection
- `rdma_nics_count` - RDMA/InfiniBand NIC count verification
- `memory_check` - System memory configuration validation
- `cpu_check` - CPU configuration and performance checks

### Variable Substitution

The following variables can be used in `issue` and `suggestion` fields:

- `{gpu_count}` - Number of GPUs detected
- `{num_rdma_nics}` - Number of RDMA NICs detected
- `{total_issues}` - Total number of issues (summary only)
- `{critical_count}` - Number of critical issues (summary only)
- `{warning_count}` - Number of warning issues (summary only)

### Recommendation Types

- **critical** - Issues that prevent proper system operation
- **warning** - Issues that may impact performance but don't prevent operation
- **info** - Informational messages about successful tests

## Example Configurations

### Basic Configuration (Default)

The current `configs/recommendations.json` contains the basic configuration with support for:
- GPU count checks  
- PCIe error detection
- RDMA NIC count verification

This file is actively used by the system and contains production-ready recommendations.

### Extended Configuration Example

For more comprehensive diagnostics, you can extend the configuration with additional test types:

```json
{
  "recommendations": {
    "gpu_count_check": {
      "fail": {
        "type": "critical",
        "issue": "GPU count mismatch detected. Expected count not met (found: {gpu_count})",
        "suggestion": "Verify GPU hardware installation and driver status. Check if all GPUs are properly seated.",
        "commands": [
          "nvidia-smi",
          "lspci | grep -i nvidia",
          "dmesg | grep -i nvidia",
          "nvidia-smi -L",
          "sudo lshw -C display"
        ],
        "references": [
          "https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/",
          "https://developer.nvidia.com/nvidia-system-management-interface"
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
    "memory_check": {
      "fail": {
        "type": "critical",
        "issue": "Insufficient memory detected for HPC workloads",
        "suggestion": "Verify system memory configuration and check for memory errors",
        "commands": [
          "free -h",
          "dmidecode --type memory",
          "dmesg | grep -i memory"
        ],
        "references": [
          "https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm"
        ]
      },
      "pass": {
        "type": "info",
        "issue": "Memory configuration check passed",
        "suggestion": "System memory is properly configured for HPC workloads",
        "commands": ["free -h"]
      }
    }
  },
  "summary_templates": {
    "no_issues": "🎉 All diagnostic tests passed! Your HPC environment is ready for high-performance computing workloads.",
    "has_issues": "⚠️ Found {total_issues} issue(s) requiring attention: {critical_count} critical, {warning_count} warning. Please review the recommendations below."
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
sudo make install
```

#### Package Installation
```bash
# Build RPM package
make rpm

# Build DEB package
make deb

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

Option A: Use current directory
```bash
cp my-custom-recommendations.json ./recommendations.json
```

Option B: Use system-wide location
```bash
sudo cp my-custom-recommendations.json /etc/oci-dr-hpc-recommendations.json
```

Option C: Use project configs directory
```bash
cp my-custom-recommendations.json configs/recommendations.json
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

Add your test to the `testMappings` in `generateRecommendations` function:

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

### Conditional Recommendations

You can create different recommendations based on context by using multiple configuration files and switching between them:

```bash
# Development environment
cp configs/dev-recommendations.json ./recommendations.json

# Production environment  
cp configs/prod-recommendations.json ./recommendations.json
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