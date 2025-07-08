#!/usr/bin/env python3
"""
Example Custom Test Script Template (Python)
============================================

This is a template script for creating custom diagnostic tests that can be executed
using the oci-dr-hpc-v2 custom-script command.

Usage:
    oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.py
    oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.py --output json
    oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.py --limits-file internal/test_limits/test_limits.json --recommendations-file configs/recommendations.json

Exit Codes:
    0 - Success (all tests passed)
    1 - Failure (one or more tests failed)
    2 - Error (script execution error)
"""

import json
import sys
import time
import subprocess
import os
import shutil
from datetime import datetime

class CustomTestRunner:
    def __init__(self):
        self.results = []
        self.start_time = time.time()
        
    def run_test(self, test_name, test_func, expected_result="PASS"):
        """Run a single test and record the result."""
        print(f"🧪 Running test: {test_name}")
        
        try:
            result = test_func()
            if result:
                status = "PASS"
                print(f"✅ PASS: {test_name}")
            else:
                status = "FAIL"
                print(f"❌ FAIL: {test_name}")
                
            self.results.append({
                "test_name": test_name,
                "status": status,
                "message": f"Test {status.lower()}ed",
                "timestamp": datetime.utcnow().isoformat() + "Z"
            })
            
        except Exception as e:
            status = "ERROR"
            print(f"💥 ERROR: {test_name} - {str(e)}")
            self.results.append({
                "test_name": test_name,
                "status": status,
                "message": f"Test error: {str(e)}",
                "timestamp": datetime.utcnow().isoformat() + "Z"
            })
            
        return status == "PASS"
    
    def skip_test(self, test_name, reason):
        """Skip a test with a reason."""
        print(f"⚠️ SKIP: {test_name} - {reason}")
        self.results.append({
            "test_name": test_name,
            "status": "SKIP",
            "message": f"Skipped: {reason}",
            "timestamp": datetime.utcnow().isoformat() + "Z"
        })
        
    def run_command(self, cmd, timeout=30):
        """Execute a system command and return the result."""
        try:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=timeout)
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
                "stderr": "Command timed out",
                "returncode": 124
            }
        except Exception as e:
            return {
                "success": False,
                "stdout": "",
                "stderr": str(e),
                "returncode": 1
            }
    
    def check_system_requirements(self):
        """Check basic system requirements."""
        print("📋 Checking system requirements...")
        
        # Check Python version
        def python_version_test():
            return sys.version_info >= (3, 6)
        
        self.run_test("python_version_check", python_version_test)
        
        # Check available disk space
        def disk_space_test():
            try:
                total, used, free = shutil.disk_usage("/")
                free_gb = free / (1024**3)
                return free_gb > 1.0
            except:
                return False
        
        self.run_test("disk_space_check", disk_space_test)
        
        # Check if /tmp is writable
        def tmp_writable_test():
            try:
                test_file = "/tmp/test_write_permission"
                with open(test_file, "w") as f:
                    f.write("test")
                os.remove(test_file)
                return True
            except:
                return False
        
        self.run_test("tmp_writable_check", tmp_writable_test)
        
    def check_network_connectivity(self):
        """Check network connectivity."""
        print("🌐 Checking network connectivity...")
        
        # Check localhost connectivity
        def localhost_test():
            result = self.run_command("ping -c 1 -W 5 127.0.0.1")
            return result["success"]
        
        self.run_test("localhost_ping_check", localhost_test)
        
        # Check external connectivity (if available)
        def external_test():
            result = self.run_command("ping -c 1 -W 5 8.8.8.8")
            return result["success"]
        
        if self.run_command("ping -c 1 -W 1 8.8.8.8")["success"]:
            self.run_test("external_ping_check", external_test)
        else:
            self.skip_test("external_ping_check", "No external network access")
    
    def check_gpu_availability(self):
        """Check GPU availability."""
        print("🖥️ Checking GPU availability...")
        
        # Check if nvidia-smi is available
        nvidia_result = self.run_command("which nvidia-smi")
        if nvidia_result["success"]:
            def nvidia_smi_test():
                result = self.run_command("nvidia-smi")
                return result["success"]
            
            self.run_test("nvidia_smi_check", nvidia_smi_test)
            
            # Check GPU count
            def gpu_count_test():
                result = self.run_command("nvidia-smi --query-gpu=count --format=csv,noheader")
                if result["success"]:
                    try:
                        count = len(result["stdout"].split('\n'))
                        print(f"ℹ️ Found {count} GPU(s)")
                        return count > 0
                    except:
                        return False
                return False
            
            self.run_test("gpu_count_check", gpu_count_test)
        else:
            self.skip_test("gpu_availability_check", "nvidia-smi not available")
    
    def run_custom_tests(self):
        """Run custom tests."""
        print("🧪 Running custom tests...")
        
        # Example: Check if a specific file exists
        def config_file_test():
            return os.path.exists("/etc/hostname")
        
        self.run_test("config_file_check", config_file_test)
        
        # Example: Performance test
        def performance_test():
            start = time.time()
            time.sleep(0.1)  # Simulate work
            duration = time.time() - start
            print(f"ℹ️ Performance test took {duration:.3f}s")
            return duration < 1.0
        
        self.run_test("performance_test", performance_test)
        
        # Example: Memory test
        def memory_test():
            try:
                # Simple memory allocation test
                test_data = [i for i in range(1000)]
                return len(test_data) == 1000
            except:
                return False
        
        self.run_test("memory_allocation_test", memory_test)
        
    def generate_summary(self):
        """Generate test summary."""
        total_tests = len(self.results)
        passed_tests = len([r for r in self.results if r["status"] == "PASS"])
        failed_tests = len([r for r in self.results if r["status"] == "FAIL"])
        skipped_tests = len([r for r in self.results if r["status"] == "SKIP"])
        error_tests = len([r for r in self.results if r["status"] == "ERROR"])
        
        execution_time = time.time() - self.start_time
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        summary = {
            "test_suite": "custom_python_template",
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "execution_time_seconds": round(execution_time, 2),
            "summary": {
                "total_tests": total_tests,
                "passed": passed_tests,
                "failed": failed_tests,
                "skipped": skipped_tests,
                "errors": error_tests,
                "success_rate": round(success_rate, 1)
            },
            "test_results": self.results
        }
        
        return summary, failed_tests == 0 and error_tests == 0
    
    def print_summary(self, summary):
        """Print test summary."""
        print("\n" + "=" * 50)
        print("📊 TEST SUMMARY")
        print("=" * 50)
        print(f"Total Tests: {summary['summary']['total_tests']}")
        print(f"Passed: {summary['summary']['passed']}")
        print(f"Failed: {summary['summary']['failed']}")
        print(f"Skipped: {summary['summary']['skipped']}")
        print(f"Errors: {summary['summary']['errors']}")
        print(f"Success Rate: {summary['summary']['success_rate']}%")
        print(f"Execution Time: {summary['execution_time_seconds']}s")
        
        print("\n📋 DETAILED RESULTS:")
        print("-" * 50)
        for result in self.results:
            status_icon = {
                "PASS": "✅",
                "FAIL": "❌",
                "SKIP": "⚠️",
                "ERROR": "💥"
            }.get(result["status"], "❓")
            
            print(f"{status_icon} {result['test_name']}: {result['message']}")
        
        print("\n📄 JSON OUTPUT:")
        print(json.dumps(summary, indent=2))

def main():
    """Main test execution function."""
    print("🚀 Starting Custom Python Test Script")
    print("=" * 50)
    
    try:
        runner = CustomTestRunner()
        
        # Run different test categories
        runner.check_system_requirements()
        print()
        runner.check_network_connectivity()
        print()
        runner.check_gpu_availability()
        print()
        runner.run_custom_tests()
        
        # Generate and print summary
        summary, success = runner.generate_summary()
        runner.print_summary(summary)
        
        # Return appropriate exit code
        if success:
            print("\n✅ All tests passed!")
            return 0
        else:
            print(f"\n❌ Some tests failed or had errors")
            return 1
            
    except Exception as e:
        print(f"\n💥 Script execution error: {e}")
        error_report = {
            "test_suite": "custom_python_template",
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "error": str(e),
            "status": "ERROR"
        }
        print(json.dumps(error_report, indent=2))
        return 2

if __name__ == "__main__":
    sys.exit(main()) 