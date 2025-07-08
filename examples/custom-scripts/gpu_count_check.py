#!/usr/bin/env python3
"""
Custom GPU Count Check Script (Python)
======================================

This script demonstrates how to create a custom Level 1 test that:
- Reads configuration from test_limits.json
- Detects current OCI shape
- Validates GPU count against expected values
- Produces output compatible with the recommender system

Usage:
    oci-dr-hpc-v2 custom-script --script examples/custom-scripts/gpu_count_check.py \
                   --limits-file internal/test_limits/test_limits.json \
                   --recommendations-file configs/recommendations.json \
                   --output json

Exit Codes:
    0 - Test passed
    1 - Test failed  
    2 - Configuration/execution error
"""

import json
import sys
import subprocess
import os
import time
from datetime import datetime

class GPUCountChecker:
    def __init__(self):
        self.is_terminal = sys.stdout.isatty()
        self.test_name = "gpu_count_check"
        self.test_category = "LEVEL_1"
        
        # Expected GPU counts by shape
        self.shape_gpu_mapping = {
            "BM.GPU.H100.8": 8,
            "BM.GPU.H200.8": 8, 
            "BM.GPU.B200.8": 8,
            "BM.GPU.GB200.4": 4
        }
    
    def log(self, message, level="INFO"):
        """Log messages with appropriate formatting."""
        if level == "ERROR":
            icon = "❌" if self.is_terminal else "[ERROR]"
        elif level == "WARN":
            icon = "⚠️" if self.is_terminal else "[WARN]"
        else:
            icon = "ℹ️" if self.is_terminal else "[INFO]"
        
        print(f"{icon} {message}")
    
    def run_command(self, cmd, timeout=30):
        """Execute a system command safely."""
        try:
            result = subprocess.run(
                cmd, shell=True, capture_output=True, text=True, timeout=timeout
            )
            return {
                "success": result.returncode == 0,
                "stdout": result.stdout.strip(),
                "stderr": result.stderr.strip(),
                "returncode": result.returncode
            }
        except subprocess.TimeoutExpired:
            return {
                "success": False,
                "stdout": "",
                "stderr": "Command timeout",
                "returncode": 124
            }
        except Exception as e:
            return {
                "success": False,
                "stdout": "",
                "stderr": str(e),
                "returncode": 1
            }
    
    def load_test_limits(self, limits_file=None):
        """Load test limits configuration."""
        if not limits_file:
            # Try default locations
            possible_paths = [
                "./test_limits.json",
                "/etc/oci-dr-hpc-test-limits.json",
                "~/.config/oci-dr-hpc/test_limits.json",
                "internal/test_limits/test_limits.json"
            ]
            
            for path in possible_paths:
                expanded_path = os.path.expanduser(path)
                if os.path.exists(expanded_path):
                    limits_file = expanded_path
                    break
        
        if not limits_file or not os.path.exists(limits_file):
            raise FileNotFoundError("Test limits file not found")
        
        self.log(f"Loading test limits from: {limits_file}")
        
        try:
            with open(limits_file, 'r') as f:
                limits_data = json.load(f)
            return limits_data
        except Exception as e:
            raise Exception(f"Failed to load test limits: {e}")
    
    def get_current_shape(self):
        """Get current OCI shape from IMDS or environment."""
        # Try environment variable first (for testing)
        if "OCI_SHAPE" in os.environ:
            shape = os.environ["OCI_SHAPE"]
            self.log(f"Using shape from environment: {shape}")
            return shape
        
        # Try IMDS
        self.log("Querying OCI shape from IMDS...")
        cmd = "curl -s -H 'Authorization: Bearer Oracle' -L http://169.254.169.254/opc/v2/instance/ | grep shape"
        result = self.run_command(cmd, timeout=10)
        
        if result["success"] and result["stdout"]:
            try:
                # Parse shape from IMDS response 
                for line in result["stdout"].split('\n'):
                    if '"shape"' in line:
                        shape = line.split(':')[1].strip().strip('",')
                        self.log(f"Detected shape from IMDS: {shape}")
                        return shape
            except Exception as e:
                self.log(f"Failed to parse IMDS response: {e}", "WARN")
        
        # Fallback - try alternative IMDS endpoint
        cmd = "curl -s http://169.254.169.254/opc/v1/instance/shape"
        result = self.run_command(cmd, timeout=5)
        if result["success"] and result["stdout"]:
            shape = result["stdout"].strip()
            self.log(f"Detected shape from IMDS v1: {shape}")
            return shape
        
        raise Exception("Could not determine OCI shape from IMDS")
    
    def is_test_enabled(self, limits_data, shape):
        """Check if GPU count test is enabled for the given shape."""
        try:
            test_config = limits_data["test_limits"][shape][self.test_name]
            enabled = test_config.get("enabled", False)
            category = test_config.get("test_category", "UNKNOWN")
            
            self.log(f"Test {self.test_name} for shape {shape}: enabled={enabled}, category={category}")
            return enabled, category
        except KeyError:
            self.log(f"No test configuration found for shape {shape}", "WARN")
            return False, "UNKNOWN"
    
    def get_expected_gpu_count(self, shape):
        """Get expected GPU count for the given shape."""
        expected_count = self.shape_gpu_mapping.get(shape, 0)
        self.log(f"Expected GPU count for {shape}: {expected_count}")
        return expected_count
    
    def get_actual_gpu_count(self):
        """Get actual GPU count from nvidia-smi."""
        self.log("Querying actual GPU count from nvidia-smi...")
        
        # Check if nvidia-smi is available
        check_cmd = "which nvidia-smi"
        check_result = self.run_command(check_cmd)
        if not check_result["success"]:
            raise Exception("nvidia-smi not found in PATH")
        
        # Query GPU count
        cmd = "nvidia-smi --query-gpu=name --format=csv,noheader"
        result = self.run_command(cmd)
        
        if not result["success"]:
            raise Exception(f"nvidia-smi failed: {result['stderr']}")
        
        # Count GPUs
        gpu_lines = [line.strip() for line in result["stdout"].split('\n') if line.strip()]
        actual_count = len(gpu_lines)
        
        self.log(f"Actual GPU count from nvidia-smi: {actual_count}")
        if actual_count > 0:
            self.log(f"GPU models detected: {', '.join(gpu_lines[:3])}{'...' if len(gpu_lines) > 3 else ''}")
        
        return actual_count
    
    def run_test(self, limits_file=None):
        """Run the GPU count check test."""
        start_time = time.time()
        test_result = {
            "test_name": self.test_name,
            "test_category": self.test_category,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "execution_time_seconds": 0,
            "status": "UNKNOWN",
            "expected_count": 0,
            "actual_count": 0,
            "shape": "UNKNOWN",
            "message": "",
            "details": {}
        }
        
        try:
            # Step 1: Load test limits
            self.log("Step 1: Loading test configuration...")
            limits_data = self.load_test_limits(limits_file)
            
            # Step 2: Get current shape
            self.log("Step 2: Detecting OCI shape...")
            shape = self.get_current_shape()
            test_result["shape"] = shape
            
            # Step 3: Check if test is enabled
            self.log("Step 3: Checking if test is enabled...")
            enabled, category = self.is_test_enabled(limits_data, shape)
            test_result["test_category"] = category
            
            if not enabled:
                test_result["status"] = "SKIP"
                test_result["message"] = f"GPU count check not enabled for shape {shape}"
                self.log(test_result["message"], "WARN")
                return test_result
            
            # Step 4: Get expected GPU count
            self.log("Step 4: Determining expected GPU count...")
            expected_count = self.get_expected_gpu_count(shape)
            test_result["expected_count"] = expected_count
            
            if expected_count == 0:
                test_result["status"] = "SKIP"
                test_result["message"] = f"No GPU count mapping defined for shape {shape}"
                self.log(test_result["message"], "WARN")
                return test_result
            
            # Step 5: Get actual GPU count
            self.log("Step 5: Querying actual GPU count...")
            actual_count = self.get_actual_gpu_count()
            test_result["actual_count"] = actual_count
            
            # Step 6: Compare and determine result
            self.log("Step 6: Comparing expected vs actual GPU counts...")
            if actual_count == expected_count:
                test_result["status"] = "PASS"
                test_result["message"] = f"GPU count check passed: found {actual_count} GPUs as expected"
                self.log(test_result["message"])
            else:
                test_result["status"] = "FAIL"
                if actual_count < expected_count:
                    missing = expected_count - actual_count
                    test_result["message"] = f"GPU count check failed: missing {missing} GPUs (expected {expected_count}, found {actual_count})"
                    test_result["details"]["missing_gpus"] = missing
                else:
                    extra = actual_count - expected_count
                    test_result["message"] = f"GPU count check failed: found {extra} extra GPUs (expected {expected_count}, found {actual_count})"
                    test_result["details"]["extra_gpus"] = extra
                
                self.log(test_result["message"], "ERROR")
        
        except Exception as e:
            test_result["status"] = "ERROR"
            test_result["message"] = f"GPU count check error: {str(e)}"
            test_result["details"]["error"] = str(e)
            self.log(test_result["message"], "ERROR")
        
        finally:
            test_result["execution_time_seconds"] = round(time.time() - start_time, 2)
        
        return test_result
    
    def format_output(self, test_result):
        """Format output for recommender system compatibility."""
        # Format compatible with existing reporter system
        output = {
            "test_suite": "custom_gpu_count_check",
            "timestamp": test_result["timestamp"],
            "execution_time_seconds": test_result["execution_time_seconds"],
            "test_results": {
                "gpu_count_check": [test_result]
            },
            "summary": {
                "total_tests": 1,
                "passed": 1 if test_result["status"] == "PASS" else 0,
                "failed": 1 if test_result["status"] == "FAIL" else 0,
                "skipped": 1 if test_result["status"] == "SKIP" else 0,
                "errors": 1 if test_result["status"] == "ERROR" else 0
            }
        }
        
        return output

def main():
    """Main execution function."""
    checker = GPUCountChecker()
    
    # Check for limits file from environment or argument
    limits_file = os.environ.get("OCI_DR_HPC_LIMITS_FILE")
    
    try:
        checker.log("🚀 Starting GPU Count Check Test" if checker.is_terminal else "Starting GPU Count Check Test")
        checker.log("=" * 50)
        
        # Run the test
        test_result = checker.run_test(limits_file)
        
        # Format output
        output = checker.format_output(test_result)
        
        # Print summary
        checker.log("")
        checker.log("📊 TEST SUMMARY" if checker.is_terminal else "TEST SUMMARY")
        checker.log("-" * 30)
        checker.log(f"Test: {test_result['test_name']}")
        checker.log(f"Shape: {test_result['shape']}")
        checker.log(f"Status: {test_result['status']}")
        checker.log(f"Expected GPUs: {test_result['expected_count']}")
        checker.log(f"Actual GPUs: {test_result['actual_count']}")
        checker.log(f"Execution Time: {test_result['execution_time_seconds']}s")
        checker.log(f"Message: {test_result['message']}")
        
        # Print JSON output
        checker.log("")
        checker.log("📄 JSON OUTPUT:" if checker.is_terminal else "JSON OUTPUT:")
        print(json.dumps(output, indent=2))
        
        # Return appropriate exit code
        if test_result["status"] == "PASS":
            checker.log("\n✅ GPU count check PASSED" if checker.is_terminal else "\n[SUCCESS] GPU count check PASSED")
            return 0
        elif test_result["status"] in ["SKIP"]:
            checker.log("\n⚠️ GPU count check SKIPPED" if checker.is_terminal else "\n[SKIP] GPU count check SKIPPED")
            return 0  # Skip is not a failure
        elif test_result["status"] == "FAIL":
            checker.log("\n❌ GPU count check FAILED" if checker.is_terminal else "\n[FAIL] GPU count check FAILED")
            return 1
        else:  # ERROR
            checker.log("\n💥 GPU count check ERROR" if checker.is_terminal else "\n[ERROR] GPU count check ERROR")
            return 2
            
    except Exception as e:
        checker.log(f"Script execution error: {e}", "ERROR")
        error_output = {
            "test_suite": "custom_gpu_count_check",
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "error": str(e),
            "status": "ERROR"
        }
        print(json.dumps(error_output, indent=2))
        return 2

if __name__ == "__main__":
    sys.exit(main()) 