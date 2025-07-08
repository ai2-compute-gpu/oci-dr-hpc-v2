#!/bin/bash
# Custom GPU Count Check Script (Bash)
# ====================================
#
# This script demonstrates how to create a custom Level 1 test that:
# - Reads configuration from test_limits.json
# - Detects current OCI shape
# - Validates GPU count against expected values
# - Produces output compatible with the recommender system
#
# Usage:
#     oci-dr-hpc-v2 custom-script --script examples/custom-scripts/gpu_count_check.sh \
#                    --limits-file internal/test_limits/test_limits.json \
#                    --recommendations-file configs/recommendations.json \
#                    --output json
#
# Exit Codes:
#     0 - Test passed or skipped
#     1 - Test failed  
#     2 - Configuration/execution error

set -euo pipefail

# Global variables
TEST_NAME="gpu_count_check"
TEST_CATEGORY="LEVEL_1"
IS_TERMINAL=false
START_TIME=$(date +%s)

# Check if output is going to terminal
if [ -t 1 ]; then
    IS_TERMINAL=true
fi

# Logging function with terminal detection
log() {
    local message="$1"
    local level="${2:-INFO}"
    local icon=""
    
    case "$level" in
        "ERROR") icon=$([ "$IS_TERMINAL" = true ] && echo "❌" || echo "[ERROR]") ;;
        "WARN")  icon=$([ "$IS_TERMINAL" = true ] && echo "⚠️" || echo "[WARN]") ;;
        *)       icon=$([ "$IS_TERMINAL" = true ] && echo "ℹ️" || echo "[INFO]") ;;
    esac
    
    echo "$icon $message" >&2
}

# Function to run commands safely
run_command() {
    local cmd="$1"
    local timeout="${2:-30}"
    
    if timeout "$timeout" bash -c "$cmd" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to get command output safely
get_command_output() {
    local cmd="$1"
    local timeout="${2:-30}"
    
    timeout "$timeout" bash -c "$cmd" 2>/dev/null || echo ""
}

# Function to load test limits using jq
load_test_limits() {
    local limits_file="$1"
    
    if [ ! -f "$limits_file" ]; then
        log "Test limits file not found: $limits_file" "ERROR"
        return 1
    fi
    
    log "Loading test limits from: $limits_file"
    
    # Validate JSON
    if ! jq empty "$limits_file" 2>/dev/null; then
        log "Invalid JSON in test limits file" "ERROR"
        return 1
    fi
    
    return 0
}

# Function to get current OCI shape
get_current_shape() {
    local shape=""
    
    # Try environment variable first
    if [ -n "${OCI_SHAPE:-}" ]; then
        shape="$OCI_SHAPE"
        log "Using shape from environment: $shape"
        echo "$shape"
        return 0
    fi
    
    # Try IMDS
    log "Querying OCI shape from IMDS..."
    
    # Try IMDS v2 first
    local imds_response
    imds_response=$(get_command_output "curl -s -H 'Authorization: Bearer Oracle' -L http://169.254.169.254/opc/v2/instance/" 10)
    
    if [ -n "$imds_response" ]; then
        shape=$(echo "$imds_response" | grep '"shape"' | cut -d':' -f2 | tr -d ' ",')
        if [ -n "$shape" ]; then
            log "Detected shape from IMDS v2: $shape"
            echo "$shape"
            return 0
        fi
    fi
    
    # Try IMDS v1 fallback
    shape=$(get_command_output "curl -s http://169.254.169.254/opc/v1/instance/shape" 5)
    if [ -n "$shape" ]; then
        log "Detected shape from IMDS v1: $shape"
        echo "$shape"
        return 0
    fi
    
    log "Could not determine OCI shape from IMDS" "ERROR"
    return 1
}

# Function to check if test is enabled
is_test_enabled() {
    local limits_file="$1"
    local shape="$2"
    
    local enabled
    enabled=$(jq -r ".test_limits[\"$shape\"][\"$TEST_NAME\"].enabled" "$limits_file" 2>/dev/null)
    
    if [ "$enabled" = "true" ]; then
        local category
        category=$(jq -r ".test_limits[\"$shape\"][\"$TEST_NAME\"].test_category" "$limits_file" 2>/dev/null)
        log "Test $TEST_NAME for shape $shape: enabled=true, category=$category"
        echo "true"
        return 0
    else
        log "Test $TEST_NAME for shape $shape: enabled=false" "WARN"
        echo "false"
        return 0
    fi
}

# Function to get expected GPU count based on shape
get_expected_gpu_count() {
    local shape="$1"
    local expected_count=0
    
    case "$shape" in
        "BM.GPU.H100.8"|"BM.GPU.H200.8"|"BM.GPU.B200.8")
            expected_count=8
            ;;
        "BM.GPU.GB200.4")
            expected_count=4
            ;;
        *)
            expected_count=0
            ;;
    esac
    
    log "Expected GPU count for $shape: $expected_count"
    echo "$expected_count"
}

# Function to get actual GPU count from nvidia-smi
get_actual_gpu_count() {
    log "Querying actual GPU count from nvidia-smi..."
    
    # Check if nvidia-smi is available
    if ! command -v nvidia-smi >/dev/null 2>&1; then
        log "nvidia-smi not found in PATH" "ERROR"
        return 1
    fi
    
    # Query GPU count
    local gpu_output
    gpu_output=$(get_command_output "nvidia-smi --query-gpu=name --format=csv,noheader" 10)
    
    if [ -z "$gpu_output" ]; then
        log "nvidia-smi returned no output" "ERROR"
        return 1
    fi
    
    # Count non-empty lines
    local actual_count
    actual_count=$(echo "$gpu_output" | grep -c '^[[:space:]]*[^[:space:]]' || echo "0")
    
    log "Actual GPU count from nvidia-smi: $actual_count"
    
    if [ "$actual_count" -gt 0 ]; then
        local first_gpu
        first_gpu=$(echo "$gpu_output" | head -1 | tr -d '\r\n')
        log "GPU model detected: $first_gpu"
    fi
    
    echo "$actual_count"
}

# Function to create JSON output
create_json_output() {
    local status="$1"
    local shape="$2"
    local expected_count="$3"
    local actual_count="$4"
    local message="$5"
    local execution_time="$6"
    local details="${7:-{}}"
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Create test result JSON using printf to avoid jq issues
    # Escape quotes in message for JSON safety
    local escaped_message=$(echo "$message" | sed 's/"/\\"/g')
    
    printf '{
  "test_suite": "custom_gpu_count_check",
  "timestamp": "%s",
  "execution_time_seconds": %d,
  "test_results": {
    "gpu_count_check": [
      {
        "test_name": "%s",
        "test_category": "%s",
        "timestamp": "%s",
        "execution_time_seconds": %d,
        "status": "%s",
        "expected_count": %d,
        "actual_count": %d,
        "shape": "%s",
        "message": "%s",
        "details": %s
      }
    ]
  },
  "summary": {
    "total_tests": 1,
    "passed": %d,
    "failed": %d,
    "skipped": %d,
    "errors": %d
  }
}' \
    "$timestamp" \
    "$execution_time" \
    "$TEST_NAME" \
    "$TEST_CATEGORY" \
    "$timestamp" \
    "$execution_time" \
    "$status" \
    "$expected_count" \
    "$actual_count" \
    "$shape" \
    "$escaped_message" \
    "$details" \
    $([ "$status" = "PASS" ] && echo 1 || echo 0) \
    $([ "$status" = "FAIL" ] && echo 1 || echo 0) \
    $([ "$status" = "SKIP" ] && echo 1 || echo 0) \
    $([ "$status" = "ERROR" ] && echo 1 || echo 0)
}

# Main test execution function
run_gpu_count_test() {
    local limits_file="${OCI_DR_HPC_LIMITS_FILE:-}"
    local status="UNKNOWN"
    local shape="UNKNOWN"
    local expected_count=0
    local actual_count=0
    local message=""
    local details='{}'
    local exit_code=2
    
    # Find limits file if not provided
    if [ -z "$limits_file" ]; then
        local possible_paths=(
            "./test_limits.json"
            "/etc/oci-dr-hpc-test-limits.json"
            "$HOME/.config/oci-dr-hpc/test_limits.json"
            "internal/test_limits/test_limits.json"
        )
        
        for path in "${possible_paths[@]}"; do
            if [ -f "$path" ]; then
                limits_file="$path"
                break
            fi
        done
    fi
    
    if [ -z "$limits_file" ] || [ ! -f "$limits_file" ]; then
        status="ERROR"
        message="Test limits file not found"
        log "$message" "ERROR"
    else
        # Step 1: Load test limits
        log "Step 1: Loading test configuration..."
        if ! load_test_limits "$limits_file"; then
            status="ERROR"
            message="Failed to load test limits"
        else
            # Step 2: Get current shape
            log "Step 2: Detecting OCI shape..."
            if shape=$(get_current_shape); then
                # Step 3: Check if test is enabled
                log "Step 3: Checking if test is enabled..."
                local enabled
                enabled=$(is_test_enabled "$limits_file" "$shape")
                
                if [ "$enabled" = "false" ]; then
                    status="SKIP"
                    message="GPU count check not enabled for shape $shape"
                    log "$message" "WARN"
                    exit_code=0
                else
                    # Step 4: Get expected GPU count
                    log "Step 4: Determining expected GPU count..."
                    expected_count=$(get_expected_gpu_count "$shape")
                    
                    if [ "$expected_count" -eq 0 ]; then
                        status="SKIP"
                        message="No GPU count mapping defined for shape $shape"
                        log "$message" "WARN"
                        exit_code=0
                    else
                        # Step 5: Get actual GPU count
                        log "Step 5: Querying actual GPU count..."
                        if actual_count=$(get_actual_gpu_count); then
                            # Step 6: Compare and determine result
                            log "Step 6: Comparing expected vs actual GPU counts..."
                            if [ "$actual_count" -eq "$expected_count" ]; then
                                status="PASS"
                                message="GPU count check passed: found $actual_count GPUs as expected"
                                log "$message"
                                exit_code=0
                            else
                                status="FAIL"
                                if [ "$actual_count" -lt "$expected_count" ]; then
                                    local missing=$((expected_count - actual_count))
                                    message="GPU count check failed: missing $missing GPUs (expected $expected_count, found $actual_count)"
                                    details="{\"missing_gpus\": $missing}"
                                else
                                    local extra=$((actual_count - expected_count))
                                    message="GPU count check failed: found $extra extra GPUs (expected $expected_count, found $actual_count)"
                                    details="{\"extra_gpus\": $extra}"
                                fi
                                log "$message" "ERROR"
                                exit_code=1
                            fi
                        else
                            status="ERROR"
                            message="Failed to get actual GPU count from nvidia-smi"
                            log "$message" "ERROR"
                            exit_code=2
                        fi
                    fi
                fi
            else
                status="ERROR"
                message="Failed to detect OCI shape"
                exit_code=2
            fi
        fi
    fi
    
    # Calculate execution time
    local end_time
    end_time=$(date +%s)
    local execution_time=$((end_time - START_TIME))
    
    # Print summary
    log ""
    if [ "$IS_TERMINAL" = true ]; then
        log "📊 TEST SUMMARY"
    else
        log "TEST SUMMARY"
    fi
    log "$(printf '%*s' 30 '' | tr ' ' '-')"
    log "Test: $TEST_NAME"
    log "Shape: $shape"
    log "Status: $status"
    log "Expected GPUs: $expected_count"
    log "Actual GPUs: $actual_count"
    log "Execution Time: ${execution_time}s"
    log "Message: $message"
    
    # Print JSON output
    log ""
    if [ "$IS_TERMINAL" = true ]; then
        log "📄 JSON OUTPUT:"
    else
        log "JSON OUTPUT:"
    fi
    
    create_json_output "$status" "$shape" "$expected_count" "$actual_count" "$message" "$execution_time" "$details"
    
    # Print final status
    log ""
    case "$status" in
        "PASS")
            if [ "$IS_TERMINAL" = true ]; then
                log "✅ GPU count check PASSED"
            else
                log "[SUCCESS] GPU count check PASSED"
            fi
            ;;
        "SKIP")
            if [ "$IS_TERMINAL" = true ]; then
                log "⚠️ GPU count check SKIPPED"
            else
                log "[SKIP] GPU count check SKIPPED"
            fi
            ;;
        "FAIL")
            if [ "$IS_TERMINAL" = true ]; then
                log "❌ GPU count check FAILED"
            else
                log "[FAIL] GPU count check FAILED"
            fi
            ;;
        "ERROR")
            if [ "$IS_TERMINAL" = true ]; then
                log "💥 GPU count check ERROR"
            else
                log "[ERROR] GPU count check ERROR"
            fi
            ;;
    esac
    
    return $exit_code
}

# Main execution
main() {
    # Check for required tools
    if ! command -v jq >/dev/null 2>&1; then
        log "jq is required but not installed" "ERROR"
        echo '{"error": "jq command not found", "status": "ERROR"}' 
        return 2
    fi
    
    if [ "$IS_TERMINAL" = true ]; then
        log "🚀 Starting GPU Count Check Test"
    else
        log "Starting GPU Count Check Test"
    fi
    log "$(printf '%*s' 50 '' | tr ' ' '=')"
    
    # Run the test
    run_gpu_count_test
}

# Execute main function
main "$@" 