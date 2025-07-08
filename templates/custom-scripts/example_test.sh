#!/bin/bash
# Example Custom Test Script Template (Shell Script)
# ===================================================
#
# This is a template script for creating custom diagnostic tests that can be executed
# using the oci-dr-hpc-v2 custom-script command.
#
# Usage:
#     oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.sh
#     oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.sh --output json
#     oci-dr-hpc-v2 custom-script --script templates/custom-scripts/example_test.sh --limits-file internal/test_limits/test_limits.json --recommendations-file configs/recommendations.json
#
# Exit Codes:
#     0 - Success (all tests passed)
#     1 - Failure (one or more tests failed)
#     2 - Error (script execution error)

set -e  # Exit on error

# Colors for output (only if outputting to terminal)
if [ -t 1 ]; then
    # Output is going to terminal
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    # Output is being captured/redirected - no colors
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"
    
    echo -e "${BLUE}🧪 Running test: $test_name${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command"; then
        if [ "$expected_result" = "success" ]; then
            echo -e "${GREEN}✅ PASS: $test_name${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}❌ FAIL: $test_name (expected failure but got success)${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        if [ "$expected_result" = "failure" ]; then
            echo -e "${GREEN}✅ PASS: $test_name (expected failure)${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}❌ FAIL: $test_name${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
}

# Function to skip a test
skip_test() {
    local test_name="$1"
    local reason="$2"
    
    echo -e "${YELLOW}⚠️ SKIP: $test_name - $reason${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check system requirements
check_system_requirements() {
    echo -e "${BLUE}📋 Checking system requirements...${NC}"
    
    # Check if we're running on Linux
    run_test "linux_kernel_check" "[ $(uname -s) = 'Linux' ]" "success"
    
    # Check available memory (at least 1GB)
    run_test "memory_check" "[ $(free -m | awk 'NR==2{print $2}') -gt 1000 ]" "success"
    
    # Check if /tmp is writable
    run_test "tmp_writable_check" "[ -w /tmp ]" "success"
    
    # Check if basic commands are available
    if command_exists "ps"; then
        run_test "ps_command_check" "ps aux > /dev/null" "success"
    else
        skip_test "ps_command_check" "ps command not available"
    fi
}

# Function to check network connectivity
check_network_connectivity() {
    echo -e "${BLUE}🌐 Checking network connectivity...${NC}"
    
    # Check if ping command is available
    if command_exists "ping"; then
        # Test ping to localhost
        run_test "localhost_ping_check" "ping -c 1 -W 5 127.0.0.1 > /dev/null" "success"
        
        # Test ping to external host (if possible)
        if ping -c 1 -W 5 8.8.8.8 > /dev/null 2>&1; then
            run_test "external_ping_check" "ping -c 1 -W 5 8.8.8.8 > /dev/null" "success"
        else
            skip_test "external_ping_check" "No external network access"
        fi
    else
        skip_test "ping_checks" "ping command not available"
    fi
}

# Function to check GPU availability
check_gpu_availability() {
    echo -e "${BLUE}🖥️ Checking GPU availability...${NC}"
    
    if command_exists "nvidia-smi"; then
        # Check if nvidia-smi works
        run_test "nvidia_smi_check" "nvidia-smi > /dev/null" "success"
        
        # Check GPU count
        local gpu_count=$(nvidia-smi --query-gpu=count --format=csv,noheader 2>/dev/null | wc -l)
        if [ "$gpu_count" -gt 0 ]; then
            echo -e "${GREEN}ℹ️ Found $gpu_count GPU(s)${NC}"
            run_test "gpu_count_check" "[ $gpu_count -gt 0 ]" "success"
        else
            run_test "gpu_count_check" "false" "success"
        fi
    else
        skip_test "gpu_checks" "nvidia-smi not available"
    fi
}

# Function to run custom tests
run_custom_tests() {
    echo -e "${BLUE}🧪 Running custom tests...${NC}"
    
    # Example: Check if a specific file exists
    run_test "config_file_check" "[ -f /etc/hostname ]" "success"
    
    # Example: Check if a specific service is running (modify as needed)
    if command_exists "systemctl"; then
        if systemctl is-active --quiet ssh 2>/dev/null; then
            run_test "ssh_service_check" "systemctl is-active --quiet ssh" "success"
        else
            skip_test "ssh_service_check" "SSH service not running or not available"
        fi
    else
        skip_test "systemctl_checks" "systemctl not available"
    fi
    
    # Example: Check disk usage
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 90 ]; then
        run_test "disk_usage_check" "[ $disk_usage -lt 90 ]" "success"
    else
        run_test "disk_usage_check" "[ $disk_usage -lt 90 ]" "success"
    fi
    
    # Example: Custom performance test
    local start_time=$(date +%s%N)
    sleep 0.1  # Simulate some work
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
    
    echo -e "${BLUE}ℹ️ Performance test took ${duration}ms${NC}"
    run_test "performance_test" "[ $duration -lt 1000 ]" "success"
}

# Function to generate JSON output
generate_json_report() {
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local success_rate=0
    
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    fi
    
    cat << EOF
{
  "test_suite": "custom_shell_script_template",
  "timestamp": "$timestamp",
  "summary": {
    "total_tests": $TOTAL_TESTS,
    "passed": $PASSED_TESTS,
    "failed": $FAILED_TESTS,
    "skipped": $SKIPPED_TESTS,
    "success_rate": $success_rate
  },
  "status": "$( [ $FAILED_TESTS -eq 0 ] && echo "PASS" || echo "FAIL" )"
}
EOF
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Starting Custom Shell Script Test${NC}"
    if [ -t 1 ]; then
        echo "=" | tr ' ' '=' | head -c 50; echo
    else
        echo "=================================================="
    fi
    
    local start_time=$(date +%s)
    
    # Run test categories
    check_system_requirements
    echo
    check_network_connectivity
    echo
    check_gpu_availability
    echo
    run_custom_tests
    
    local end_time=$(date +%s)
    local execution_time=$((end_time - start_time))
    
    # Print summary
    echo
    if [ -t 1 ]; then
        echo "=" | tr ' ' '=' | head -c 50; echo
    else
        echo "=================================================="
    fi
    echo -e "${BLUE}📊 TEST SUMMARY${NC}"
    if [ -t 1 ]; then
        echo "=" | tr ' ' '=' | head -c 50; echo
    else
        echo "=================================================="
    fi
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Skipped: $SKIPPED_TESTS"
    
    if [ $TOTAL_TESTS -gt 0 ]; then
        local success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
        echo "Success Rate: $success_rate%"
    fi
    
    echo "Execution Time: ${execution_time}s"
    echo
    
    # Generate JSON report
    echo -e "${BLUE}📄 JSON OUTPUT:${NC}"
    generate_json_report
    echo
    
    # Return appropriate exit code
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}❌ $FAILED_TESTS test(s) failed${NC}"
        return 1
    fi
}

# Execute main function
main "$@" 