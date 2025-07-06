# Deployment Guide for OCI DR HPC v2

This guide covers how to deploy the OCI DR HPC v2 diagnostic tool for customers.

## Installation

### 1. Binary Installation

Place the compiled binary in `/usr/local/bin/`:
```bash
sudo cp oci-dr-hpc /usr/local/bin/
sudo chmod +x /usr/local/bin/oci-dr-hpc
```

### 2. Configuration Files

#### Main Configuration File
Install the main configuration file:
```bash
sudo cp config/oci-dr-hpc.yaml /etc/oci-dr-hpc.yaml
```

#### Shapes Configuration File
Install the shapes configuration file:
```bash
sudo cp internal/shapes/shapes.json /etc/oci-dr-hpc-shapes.json
```

Or use the provided installation script:
```bash
sudo ./scripts/install-shapes.sh
```

### 3. Log Directory
Create the log directory:
```bash
sudo mkdir -p /var/log/oci-dr-hpc
sudo chmod 755 /var/log/oci-dr-hpc
```

## File Locations

After installation, the tool will use these file locations:

| File | Location | Purpose |
|------|----------|---------|
| Binary | `/usr/local/bin/oci-dr-hpc` | Main executable |
| Config | `/etc/oci-dr-hpc.yaml` | Main configuration |
| Shapes | `/etc/oci-dr-hpc-shapes.json` | Hardware shapes configuration |
| Logs | `/var/log/oci-dr-hpc/oci-dr-hpc.log` | Application logs |

## Configuration

### Environment Variables

You can override configuration settings with environment variables:

```bash
# Override shapes file location
export OCI_DR_HPC_SHAPES_FILE="/custom/path/shapes.json"

# Override log level
export OCI_DR_HPC_LOGGING_LEVEL="debug"

# Override log file
export OCI_DR_HPC_LOGGING_FILE="/custom/path/app.log"
```

### Config File Options

Edit `/etc/oci-dr-hpc.yaml` to customize:

```yaml
# Output format (json|table|friendly)
output: table

# Logging configuration
logging:
  file: "/var/log/oci-dr-hpc/oci-dr-hpc.log"
  level: "info"

# Shapes configuration file path
shapes_file: "/etc/oci-dr-hpc-shapes.json"
```

## Usage

### Basic Usage
```bash
# Run all Level 1 tests
oci-dr-hpc level1

# Run specific test
oci-dr-hpc level1 --test=gpu_count_check

# Show version
oci-dr-hpc --version
```

### Testing Installation
```bash
# Test shapes file path resolution
oci-dr-hpc test-shapes

# Run autodiscovery
oci-dr-hpc autodiscover
```

## Troubleshooting

### Shapes File Not Found
If you see errors like "no such file or directory" for shapes.json:

1. **Check file location:**
   ```bash
   ls -la /etc/oci-dr-hpc-shapes.json
   ```

2. **Install the shapes file:**
   ```bash
   sudo ./scripts/install-shapes.sh
   ```

3. **Or set custom location:**
   ```bash
   export OCI_DR_HPC_SHAPES_FILE="/path/to/your/shapes.json"
   ```

### Permission Issues
If you see permission errors:

```bash
# Fix log directory permissions
sudo mkdir -p /var/log/oci-dr-hpc
sudo chmod 755 /var/log/oci-dr-hpc

# Fix shapes file permissions
sudo chmod 644 /etc/oci-dr-hpc-shapes.json
```

## Development vs Production

The tool automatically handles different environments:

- **Development**: Uses `internal/shapes/shapes.json` (source code)
- **Production**: Uses `/etc/oci-dr-hpc-shapes.json` (system-wide)

This ensures the tool works in both environments without code changes.

## Updating Shapes Configuration

To update the shapes configuration:

1. **Replace the file:**
   ```bash
   sudo cp new-shapes.json /etc/oci-dr-hpc-shapes.json
   ```

2. **Or update the config:**
   ```yaml
   shapes_file: "/path/to/new-shapes.json"
   ```

3. **Restart the application** (if running as a service)

The tool will automatically pick up the new configuration on the next run. 