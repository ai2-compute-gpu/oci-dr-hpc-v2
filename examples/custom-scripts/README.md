# Custom GPU Count Check Scripts

This directory contains example custom scripts that demonstrate how to create Level 1 tests that integrate with the `oci-dr-hpc-v2` framework.

## 📋 **Overview**

These scripts demonstrate how to create custom tests that:
- ✅ Read configuration from `test_limits.json`
- ✅ Use OCI shape detection via IMDS
- ✅ Produce output compatible with the recommender system
- ✅ Follow clean coding practices
- ✅ Handle errors gracefully
- ✅ Support both terminal and captured output

## 📁 **Files**

### `gpu_count_check.py`
Python implementation of GPU count validation test.

### `gpu_count_check.sh` 
Bash implementation of GPU count validation test.

## 🚀 **Usage Examples**

### **Basic Usage**
```bash
# Run Python version
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --output json

# Run Bash version  
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.sh \
  --output friendly
```

### **With Configuration Files**
```bash
# Use explicit configuration files
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --limits-file internal/test_limits/test_limits.json \
  --recommendations-file configs/recommendations.json \
  --output json

# Use system-wide configuration files
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.sh \
  --limits-file /etc/oci-dr-hpc-test-limits.json \
  --recommendations-file /usr/share/oci-dr-hpc/recommendations.json \
  --output table
```

### **Testing on Different Shapes**
```bash
# Override shape for testing
OCI_SHAPE="BM.GPU.H100.8" oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --output json

# Test with different shapes
OCI_SHAPE="BM.GPU.GB200.4" oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.sh \
  --output friendly
```

## 📊 **Output Format**

The scripts produce JSON output compatible with the recommender system:

```json
{
  "test_suite": "custom_gpu_count_check",
  "timestamp": "2024-01-01T12:00:00Z",
  "execution_time_seconds": 2.34,
  "test_results": {
    "gpu_count_check": [
      {
        "test_name": "gpu_count_check",
        "test_category": "LEVEL_1",
        "timestamp": "2024-01-01T12:00:00Z",
        "execution_time_seconds": 2.34,
        "status": "PASS",
        "expected_count": 8,
        "actual_count": 8,
        "shape": "BM.GPU.H100.8",
        "message": "GPU count check passed: found 8 GPUs as expected",
        "details": {}
      }
    ]
  },
  "summary": {
    "total_tests": 1,
    "passed": 1,
    "failed": 0,
    "skipped": 0,
    "errors": 0
  }
}
```

## 🔧 **Integration with Recommender**

The output format is designed to work seamlessly with the recommender system:

```bash
# Run test and pipe to recommender
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --output json > gpu_test_results.json

# Generate recommendations
oci-dr-hpc-v2 recommender \
  --results-file gpu_test_results.json \
  --output friendly
```

## ⚙️ **Configuration**

### **Test Limits Structure**

The scripts read from `test_limits.json` to determine if tests are enabled:

```json
{
  "test_limits": {
    "BM.GPU.H100.8": {
      "gpu_count_check": {
        "enabled": true,
        "test_category": "LEVEL_1"
      }
    },
    "BM.GPU.GB200.4": {
      "gpu_count_check": {
        "enabled": true,
        "test_category": "LEVEL_1"
      }
    }
  }
}
```

### **Expected GPU Counts**

The scripts have built-in mappings for OCI shapes:

| Shape | Expected GPUs |
|-------|---------------|
| `BM.GPU.H100.8` | 8 |
| `BM.GPU.H200.8` | 8 |
| `BM.GPU.B200.8` | 8 |
| `BM.GPU.GB200.4` | 4 |

### **Environment Variables**

- `OCI_SHAPE`: Override shape detection (useful for testing)
- `OCI_DR_HPC_LIMITS_FILE`: Path to test limits file

## 📋 **Test Status Values**

- **`PASS`**: Test passed - GPU count matches expectation
- **`FAIL`**: Test failed - GPU count mismatch
- **`SKIP`**: Test skipped - not enabled or no mapping for shape
- **`ERROR`**: Test error - configuration or execution issue

## 🎯 **Exit Codes**

- **`0`**: Success (PASS or SKIP)
- **`1`**: Test failure (FAIL)
- **`2`**: Configuration or execution error (ERROR)

## 🛠️ **Requirements**

### **Python Script (`gpu_count_check.py`)**
- Python 3.6+
- `nvidia-smi` command available
- Access to IMDS endpoint (for shape detection)

### **Bash Script (`gpu_count_check.sh`)**
- Bash 4.0+
- `jq` command for JSON processing
- `nvidia-smi` command available
- `curl` command for IMDS queries
- Access to IMDS endpoint (for shape detection)

## 🧪 **Testing Your Scripts**

### **Local Testing**
```bash
# Test Python script directly
python3 examples/custom-scripts/gpu_count_check.py
echo "Exit code: $?"

# Test Bash script directly
bash examples/custom-scripts/gpu_count_check.sh
echo "Exit code: $?"
```

### **Framework Testing**
```bash
# Test through framework
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --output json

# Validate JSON output
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --output json | jq '.test_results.gpu_count_check[0].status'
```

## 📝 **Creating Your Own Custom Scripts**

Use these scripts as templates for creating your own custom tests:

### **1. Copy and Modify**
```bash
# Copy the template
cp examples/custom-scripts/gpu_count_check.py my_custom_test.py

# Modify for your specific test
# - Change test_name
# - Update test logic
# - Modify expected values
# - Adjust error handling
```

### **2. Key Components to Implement**

#### **Configuration Loading**
```python
def load_test_limits(self, limits_file=None):
    # Load and validate test_limits.json
    # Check if test is enabled for current shape
```

#### **Shape Detection**
```python  
def get_current_shape(self):
    # Query OCI IMDS for current shape
    # Support environment variable override
```

#### **Test Logic**
```python
def run_test(self):
    # Implement your specific test logic
    # Return structured test result
```

#### **Output Formatting**
```python
def format_output(self, test_result):
    # Format for recommender compatibility
    # Include test_results and summary sections
```

### **3. Best Practices**

- ✅ **Use structured error handling**
- ✅ **Support both terminal and captured output** 
- ✅ **Include detailed logging and messages**
- ✅ **Follow exit code conventions**
- ✅ **Validate configuration files**
- ✅ **Handle timeouts appropriately**
- ✅ **Provide meaningful error messages**

### **4. Output Structure Requirements**

Your custom script output should include:

```json
{
  "test_suite": "your_custom_test_name",
  "timestamp": "ISO8601_timestamp",
  "execution_time_seconds": number,
  "test_results": {
    "your_test_name": [
      {
        "test_name": "string",
        "test_category": "LEVEL_1|LEVEL_2|LEVEL_3",
        "status": "PASS|FAIL|SKIP|ERROR",
        "message": "descriptive_message",
        // ... other test-specific fields
      }
    ]
  },
  "summary": {
    "total_tests": 1,
    "passed": 0_or_1,
    "failed": 0_or_1,
    "skipped": 0_or_1,
    "errors": 0_or_1
  }
}
```

## 🚨 **Troubleshooting**

### **Common Issues**

1. **`nvidia-smi not found`**
   ```bash
   # Check if nvidia-smi is in PATH
   which nvidia-smi
   
   # Check if NVIDIA drivers are installed
   lsmod | grep nvidia
   ```

2. **`jq command not found` (Bash script)**
   ```bash
   # Install jq
   sudo apt-get install jq  # Ubuntu/Debian
   sudo yum install jq      # RHEL/CentOS
   ```

3. **`Could not determine OCI shape`**
   ```bash
   # Test IMDS connectivity
   curl -s http://169.254.169.254/opc/v1/instance/shape
   
   # Use environment variable override
   export OCI_SHAPE="BM.GPU.H100.8"
   ```

4. **`Test limits file not found`**
   ```bash
   # Check file location
   find . -name "test_limits.json"
   
   # Use explicit path
   oci-dr-hpc-v2 custom-script \
     --script your_script.py \
     --limits-file /path/to/test_limits.json
   ```

### **Debug Mode**

Add debug output to your scripts:

```bash
# Enable verbose logging
export OCI_DR_HPC_LOGGING_LEVEL="debug"

# Run with explicit configuration
oci-dr-hpc-v2 custom-script \
  --script examples/custom-scripts/gpu_count_check.py \
  --limits-file internal/test_limits/test_limits.json \
  --output friendly \
  --verbose
```

## 🤝 **Contributing**

If you create useful custom scripts, consider contributing them back to the project:

1. Follow the same structure and patterns
2. Include comprehensive error handling
3. Add documentation and examples
4. Test on multiple OCI shapes
5. Submit a pull request

## 📚 **Additional Resources**

- [Main Project Documentation](../../README.md)
- [Test Limits Configuration](../../internal/test_limits/README.md)
- [Recommender System](../../docs/recommender-system.md)
- [Template Scripts](../../templates/custom-scripts/README.md) 