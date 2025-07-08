# Custom Script Templates

This directory contains template scripts that demonstrate how to create custom diagnostic tests for the `oci-dr-hpc-v2` tool.

## Overview

The `custom-script` command in `oci-dr-hpc-v2` allows you to execute custom diagnostic scripts while leveraging the application's configuration system for test limits and recommendations. These templates provide a starting point for creating your own custom tests.

## Available Templates

### 1. Python Template (`example_test.py`)

A comprehensive Python script template that demonstrates:
- Test execution framework with proper error handling
- System requirements checking
- Network connectivity tests
- GPU availability detection
- Custom test examples
- JSON output formatting
- Proper exit code handling

### 2. Shell Script Template (`example_test.sh`)

A shell script template that demonstrates:
- Bash-based test execution
- Colored output for better readability
- Test result tracking
- System validation checks
- JSON report generation
- Exit code handling

## Basic Usage

### Running Templates

```bash
# Run Python template with table output
oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.py --output table

# Run shell script template with JSON output
oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.sh --output json

# Run with configuration files
oci-dr-hpc-v2 custom-script \
  --script templates/custom-scripts/example_test.py \
  --limits-file internal/test_limits/test_limits.json \
  --recommendations-file configs/recommendations.json \
  --output friendly
```

### Available Output Formats

- **`table`**: Formatted table output (default)
- **`json`**: Machine-readable JSON output
- **`friendly`**: Human-readable output with emojis and formatting

## Creating Your Own Custom Scripts

### 1. Choose Your Language

Both Python and shell scripts are supported. Choose based on your needs:
- **Python**: Better for complex logic, data processing, API calls
- **Shell**: Better for system commands, file operations, simple tests

### 2. Script Structure

Your custom script should:
1. **Accept no arguments** (configuration is handled by the framework)
2. **Use proper exit codes**:
   - `0`: Success (all tests passed)
   - `1`: Failure (one or more tests failed)
   - `2`: Error (script execution error)
3. **Output results to stdout**
4. **Include proper error handling**

### 3. Testing Best Practices

#### Test Categories
Organize your tests into logical categories:
- **System Requirements**: Basic system checks
- **Hardware Validation**: GPU, network, storage checks
- **Performance Tests**: Timing, throughput, latency tests
- **Configuration Validation**: File existence, permissions, settings
- **Custom Business Logic**: Application-specific tests

#### Test Result Structure
Each test should have:
- **Test Name**: Unique identifier
- **Status**: PASS, FAIL, SKIP, or ERROR
- **Message**: Descriptive result message
- **Timestamp**: When the test was executed
- **Additional Data**: Any relevant test-specific information

### 4. Example Test Implementation

#### Python Example
```python
def check_gpu_memory():
    """Check if GPU has sufficient memory."""
    try:
        result = subprocess.run(
            ["nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits"],
            capture_output=True, text=True, timeout=10
        )
        
        if result.returncode == 0:
            memory_mb = int(result.stdout.strip())
            required_mb = 8000  # 8GB minimum
            
            return {
                "status": "PASS" if memory_mb >= required_mb else "FAIL",
                "message": f"GPU memory: {memory_mb}MB (required: {required_mb}MB)",
                "actual_memory": memory_mb,
                "required_memory": required_mb
            }
        else:
            return {
                "status": "FAIL",
                "message": "Failed to query GPU memory",
                "error": result.stderr
            }
    except Exception as e:
        return {
            "status": "ERROR",
            "message": f"GPU memory check failed: {str(e)}"
        }
```

#### Shell Example
```bash
check_rdma_devices() {
    local test_name="rdma_device_check"
    local expected_devices=8
    
    if command -v ibv_devices >/dev/null 2>&1; then
        local device_count=$(ibv_devices 2>/dev/null | wc -l)
        
        if [ "$device_count" -ge "$expected_devices" ]; then
            echo "✅ PASS: $test_name - Found $device_count RDMA devices"
            return 0
        else
            echo "❌ FAIL: $test_name - Found $device_count devices, expected $expected_devices"
            return 1
        fi
    else
        echo "⚠️ SKIP: $test_name - ibv_devices command not available"
        return 0
    fi
}
```

## Configuration Integration

### Using Test Limits

If you specify a `--limits-file`, your script can access test limits and thresholds:

```python
# The limits file is validated and loaded by the framework
# Your script can reference standard limits for consistency
```

### Using Recommendations

If you specify a `--recommendations-file`, your script can provide recommendations:

```python
# The recommendations file is validated and loaded by the framework  
# Your script can reference standard fault codes and recommendations
```

## Advanced Features

### Environment Variables

The framework sets these environment variables:
- `OCI_DR_HPC_SCRIPT_PATH`: Path to your script
- `OCI_DR_HPC_LIMITS_FILE`: Path to limits file (if specified)
- `OCI_DR_HPC_RECOMMENDATIONS_FILE`: Path to recommendations file (if specified)

### Integration with HPC Scripts

You can reference existing HPC test scripts:

```python
# Call existing HPC test scripts
result = subprocess.run([
    "python3", "scripts/BM.GPU.H100.8/gpu_count_check.py"
], capture_output=True, text=True)
```

### JSON Output Format

The framework captures your script output and wraps it in a standard format:

```json
{
  "script_path": "templates/custom-scripts/example_test.py",
  "status": "PASS",
  "output": "... your script output ...",
  "exit_code": 0,
  "execution_time_seconds": 2.34,
  "timestamp_utc": "2024-01-01T12:00:00Z",
  "configs_used": {
    "limits_file": "internal/test_limits/test_limits.json",
    "recommendations_file": "configs/recommendations.json"
  }
}
```

## Testing Your Scripts

### Local Testing

```bash
# Test your script directly
python3 templates/custom-scripts/example_test.py
echo "Exit code: $?"

# Test with the framework
oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.py --output json
```

### Validation Checklist

- [ ] Script executes without errors
- [ ] Proper exit codes are returned
- [ ] Output is properly formatted
- [ ] All test cases are covered
- [ ] Error handling is implemented
- [ ] Script completes in reasonable time
- [ ] JSON output is valid (if applicable)

## Best Practices

1. **Keep scripts focused** - One script per test category
2. **Use descriptive test names** - Make it clear what each test does
3. **Implement proper error handling** - Don't let exceptions crash your script
4. **Provide meaningful output** - Help users understand what was tested
5. **Use consistent exit codes** - Follow the standard convention
6. **Add timeouts** - Prevent scripts from hanging indefinitely
7. **Make tests independent** - Each test should be self-contained
8. **Document your tests** - Include comments explaining complex logic

## Troubleshooting

### Common Issues

1. **Permission Denied**: Make sure your script is executable
   ```bash
   chmod +x your_script.py
   ```

2. **Command Not Found**: Check if required tools are installed
   ```bash
   which nvidia-smi
   which python3
   ```

3. **Script Timeout**: Add timeouts to long-running operations
   ```python
   subprocess.run(cmd, timeout=30)
   ```

4. **JSON Parse Error**: Ensure your JSON output is valid
   ```bash
   python3 -m json.tool your_output.json
   ```

### Getting Help

- Check the main project documentation
- Review existing test scripts in the `scripts/` directory
- Look at the source code in `internal/custom-script/`
- Run with `--verbose` flag for detailed logging

## Contributing

If you create useful custom scripts, consider contributing them back to the project:

1. Place your script in the appropriate directory
2. Add proper documentation
3. Include test cases
4. Submit a pull request

## License

These templates are provided under the same license as the main project. 