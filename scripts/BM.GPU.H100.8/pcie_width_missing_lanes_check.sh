#!/bin/bash

# PCIe Width Missing Lanes Check
#
# This script checks the PCIe link width for GPU and RDMA interfaces to detect missing lanes.
# It verifies that all PCIe links are operating at their expected width configuration.
#
# Expected output for BM.GPU.H100.8:
# - GPU/NVSwitch: 4x Width x2 (ok), 8x Width x16 (ok)  
# - RDMA: 2x Width x8 (ok), 16x Width x16 (ok)
#
# Author: Oracle Cloud Infrastructure

set -euo pipefail

# Global variables
SCRIPT_NAME="pcie_width_missing_lanes_check"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
IS_TERMINAL=false

# Check if we're running in a terminal
if [[ -t 1 ]]; then
    IS_TERMINAL=true
fi

# Function to log messages
log_info() {
    if [[ "$IS_TERMINAL" == "true" ]]; then
        echo "INFO: $*" >&2
    fi
}

log_error() {
    if [[ "$IS_TERMINAL" == "true" ]]; then
        echo "ERROR: $*" >&2
    fi
}

# Function to run command with timeout
run_command() {
    local cmd="$1"
    local timeout_duration=30
    
    if timeout "$timeout_duration" bash -c "$cmd" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to parse lspci width output
parse_lspci_width_output() {
    local output="$1"
    local -A width_counts
    local -A speed_counts
    local state_errors=()
    
    while IFS= read -r line; do
        if [[ -z "$line" ]]; then
            continue
        fi
        
        # Parse line format: "count           LnkSta: Speed 16GT/s (ok), Width x16 (ok)"
        # Use simpler parsing with sed/grep for better compatibility
        if echo "$line" | grep -q "LnkSta:.*Speed.*Width"; then
            local count=$(echo "$line" | sed -n 's/^[[:space:]]*\([0-9]\+\).*/\1/p')
            local speed=$(echo "$line" | sed -n 's/.*Speed[[:space:]]\+\([^[:space:]]\+\).*/\1/p')
            local speed_state=$(echo "$line" | sed -n 's/.*Speed[[:space:]]\+[^[:space:]]\+[[:space:]]*(\([^)]*\)).*/\1/p')
            local width=$(echo "$line" | sed -n 's/.*Width[[:space:]]\+x\([0-9]\+\).*/\1/p')
            local width_state=$(echo "$line" | sed -n 's/.*Width[[:space:]]\+x[0-9]\+[[:space:]]*(\([^)]*\)).*/\1/p')
            
            # Only process if we got valid values
            if [[ -n "$count" && -n "$width" && -n "$speed" ]]; then
                # Count widths only if width state is ok
                if [[ "$width_state" == "ok" ]]; then
                    width_counts["Width x${width}"]="$count"
                else
                    state_errors+=("$count devices have width state '$width_state' instead of 'ok'")
                fi
                
                # Count speeds only if speed state is ok
                if [[ "$speed_state" == "ok" ]]; then
                    speed_counts["Speed ${speed}"]="$count"
                else
                    state_errors+=("$count devices have speed state '$speed_state' instead of 'ok'")
                fi
            fi
        fi
    done <<< "$output"
    
    # Output format: width_counts|speed_counts|state_errors
    local width_json="{"
    local first=true
    for width in "${!width_counts[@]}"; do
        if [[ "$first" == "false" ]]; then
            width_json="$width_json, "
        fi
        width_json="$width_json\"$width\": ${width_counts[$width]}"
        first=false
    done
    width_json="$width_json}"
    
    local speed_json="{"
    first=true
    for speed in "${!speed_counts[@]}"; do
        if [[ "$first" == "false" ]]; then
            speed_json="$speed_json, "
        fi
        speed_json="$speed_json\"$speed\": ${speed_counts[$speed]}"
        first=false
    done
    speed_json="$speed_json}"
    
    local errors_json="["
    first=true
    for error in "${state_errors[@]}"; do
        if [[ "$first" == "false" ]]; then
            errors_json="$errors_json, "
        fi
        errors_json="$errors_json\"$error\""
        first=false
    done
    errors_json="$errors_json]"
    
    echo "$width_json|$speed_json|$errors_json"
}

# Function to check GPU/NVSwitch PCIe width, speed, and state
check_gpu_nvswitch_pcie_width() {
    local cmd="sudo lspci -vvv | egrep '^[0-9,a-f]|LnkSta:' | grep -A 1 -i nvidia | grep LnkSta | sort | uniq -c"
    local output
    local gpu_success=true
    local error_messages=()
    
    log_info "Checking GPU/NVSwitch PCIe width, speed, and state..."
    
    if ! output=$(run_command "$cmd"); then
        echo "false|Failed to execute lspci command for NVIDIA devices|{}|{}|[]"
        return
    fi
    
    if [[ -z "$output" ]]; then
        echo "false|No NVIDIA PCIe devices found|{}|{}|[]"
        return
    fi
    
    # Parse width, speed, and state data
    local parse_result
    parse_result=$(parse_lspci_width_output "$output")
    IFS='|' read -r width_json speed_json errors_json <<< "$parse_result"
    
    # Expected for BM.GPU.H100.8: 4x Width x2, 8x Width x16
    local expected_x2=4
    local expected_x16=8
    # Expected speeds: 4x Speed 16GT/s, 8x Speed 32GT/s
    local expected_16gt=4
    local expected_32gt=8
    
    # Extract actual width counts from JSON-like format
    local actual_x2=0
    local actual_x16=0
    if [[ $width_json == *'"Width x2":'* ]]; then
        actual_x2=$(echo "$width_json" | sed -n 's/.*"Width x2": *\([0-9]\+\).*/\1/p')
    fi
    if [[ $width_json == *'"Width x16":'* ]]; then
        actual_x16=$(echo "$width_json" | sed -n 's/.*"Width x16": *\([0-9]\+\).*/\1/p')
    fi
    
    # Extract actual speed counts
    local actual_16gt=0
    local actual_32gt=0
    if [[ $speed_json == *'"Speed 16GT/s":'* ]]; then
        actual_16gt=$(echo "$speed_json" | sed -n 's/.*"Speed 16GT\/s": *\([0-9]\+\).*/\1/p')
    fi
    if [[ $speed_json == *'"Speed 32GT/s":'* ]]; then
        actual_32gt=$(echo "$speed_json" | sed -n 's/.*"Speed 32GT\/s": *\([0-9]\+\).*/\1/p')
    fi
    
    # Check width counts
    if [[ $actual_x2 -ne $expected_x2 ]]; then
        error_messages+=("GPU/NVSwitch PCIe width mismatch: expected ${expected_x2}x Width x2, got ${actual_x2}x Width x2")
        gpu_success=false
    fi
    if [[ $actual_x16 -ne $expected_x16 ]]; then
        error_messages+=("GPU/NVSwitch PCIe width mismatch: expected ${expected_x16}x Width x16, got ${actual_x16}x Width x16")
        gpu_success=false
    fi
    
    # Check speed counts
    if [[ $actual_16gt -ne $expected_16gt ]]; then
        error_messages+=("GPU/NVSwitch PCIe speed mismatch: expected ${expected_16gt}x Speed 16GT/s, got ${actual_16gt}x Speed 16GT/s")
        gpu_success=false
    fi
    if [[ $actual_32gt -ne $expected_32gt ]]; then
        error_messages+=("GPU/NVSwitch PCIe speed mismatch: expected ${expected_32gt}x Speed 32GT/s, got ${actual_32gt}x Speed 32GT/s")
        gpu_success=false
    fi
    
    # Check state errors
    if [[ $errors_json != "[]" ]]; then
        # Parse state errors from JSON array
        local state_error_count
        state_error_count=$(echo "$errors_json" | grep -o '"[^"]*"' | wc -l)
        if [[ $state_error_count -gt 0 ]]; then
            error_messages+=("GPU/NVSwitch state errors detected")
            gpu_success=false
        fi
    fi
    
    # Join error messages
    local final_error=""
    if [[ ${#error_messages[@]} -gt 0 ]]; then
        final_error=$(IFS='; '; echo "${error_messages[*]}")
    fi
    
    # Format output: success|error_message|width_json|speed_json|errors_json
    echo "$gpu_success|$final_error|$width_json|$speed_json|$errors_json"
}

# Function to check RDMA PCIe width, speed, and state
check_rdma_pcie_width() {
    local cmd="sudo lspci -vvv | egrep '^[0-9,a-f]|LnkSta:' | grep -A 1 -i mellanox | grep LnkSta | sort | uniq -c"
    local output
    local rdma_success=true
    local error_messages=()
    
    log_info "Checking RDMA PCIe width, speed, and state..."
    
    if ! output=$(run_command "$cmd"); then
        echo "false|Failed to execute lspci command for Mellanox devices|{}|{}|[]"
        return
    fi
    
    if [[ -z "$output" ]]; then
        echo "false|No Mellanox PCIe devices found|{}|{}|[]"
        return
    fi
    
    # Parse width, speed, and state data
    local parse_result
    parse_result=$(parse_lspci_width_output "$output")
    IFS='|' read -r width_json speed_json errors_json <<< "$parse_result"
    
    # Expected for BM.GPU.H100.8: 2x Width x8, 16x Width x16
    local expected_x8=2
    local expected_x16=16
    # Expected speeds: 2x Speed 16GT/s, 16x Speed 32GT/s
    local expected_16gt=2
    local expected_32gt=16
    
    # Extract actual width counts from JSON-like format
    local actual_x8=0
    local actual_x16=0
    if [[ $width_json == *'"Width x8":'* ]]; then
        actual_x8=$(echo "$width_json" | sed -n 's/.*"Width x8": *\([0-9]\+\).*/\1/p')
    fi
    if [[ $width_json == *'"Width x16":'* ]]; then
        actual_x16=$(echo "$width_json" | sed -n 's/.*"Width x16": *\([0-9]\+\).*/\1/p')
    fi
    
    # Extract actual speed counts
    local actual_16gt=0
    local actual_32gt=0
    if [[ $speed_json == *'"Speed 16GT/s":'* ]]; then
        actual_16gt=$(echo "$speed_json" | sed -n 's/.*"Speed 16GT\/s": *\([0-9]\+\).*/\1/p')
    fi
    if [[ $speed_json == *'"Speed 32GT/s":'* ]]; then
        actual_32gt=$(echo "$speed_json" | sed -n 's/.*"Speed 32GT\/s": *\([0-9]\+\).*/\1/p')
    fi
    
    # Check width counts
    if [[ $actual_x8 -ne $expected_x8 ]]; then
        error_messages+=("RDMA PCIe width mismatch: expected ${expected_x8}x Width x8, got ${actual_x8}x Width x8")
        rdma_success=false
    fi
    if [[ $actual_x16 -ne $expected_x16 ]]; then
        error_messages+=("RDMA PCIe width mismatch: expected ${expected_x16}x Width x16, got ${actual_x16}x Width x16")
        rdma_success=false
    fi
    
    # Check speed counts
    if [[ $actual_16gt -ne $expected_16gt ]]; then
        error_messages+=("RDMA PCIe speed mismatch: expected ${expected_16gt}x Speed 16GT/s, got ${actual_16gt}x Speed 16GT/s")
        rdma_success=false
    fi
    if [[ $actual_32gt -ne $expected_32gt ]]; then
        error_messages+=("RDMA PCIe speed mismatch: expected ${expected_32gt}x Speed 32GT/s, got ${actual_32gt}x Speed 32GT/s")
        rdma_success=false
    fi
    
    # Check state errors
    if [[ $errors_json != "[]" ]]; then
        # Parse state errors from JSON array
        local state_error_count
        state_error_count=$(echo "$errors_json" | grep -o '"[^"]*"' | wc -l)
        if [[ $state_error_count -gt 0 ]]; then
            error_messages+=("RDMA state errors detected")
            rdma_success=false
        fi
    fi
    
    # Join error messages
    local final_error=""
    if [[ ${#error_messages[@]} -gt 0 ]]; then
        final_error=$(IFS='; '; echo "${error_messages[*]}")
    fi
    
    # Format output: success|error_message|width_json|speed_json|errors_json
    echo "$rdma_success|$final_error|$width_json|$speed_json|$errors_json"
}

# Function to get OCI shape
get_oci_shape() {
    local shape
    
    # First try environment variable
    if [[ -n "${OCI_SHAPE:-}" ]]; then
        echo "$OCI_SHAPE"
        return
    fi
    
    # Try IMDS
    if shape=$(curl -s -m 10 http://169.254.169.254/opc/v1/instance/shape 2>/dev/null); then
        if [[ -n "$shape" ]]; then
            echo "$shape"
            return
        fi
    fi
    
    echo "UNKNOWN"
}

# Main function
main() {
    log_info "Starting PCIe width missing lanes check..."
    
    # Get OCI shape
    local shape
    shape=$(get_oci_shape)
    
    # Initialize variables
    local overall_success=true
    local error_messages=()
    local gpu_width_counts="{}"
    local gpu_speed_counts="{}"
    local gpu_state_errors="[]"
    local rdma_width_counts="{}"
    local rdma_speed_counts="{}"
    local rdma_state_errors="[]"
    
    # Check GPU/NVSwitch PCIe width, speed, and state
    local gpu_result
    gpu_result=$(check_gpu_nvswitch_pcie_width)
    IFS='|' read -r gpu_success gpu_error gpu_width_counts gpu_speed_counts gpu_state_errors <<< "$gpu_result"
    
    if [[ "$gpu_success" != "true" ]]; then
        overall_success=false
        error_messages+=("GPU/NVSwitch: $gpu_error")
    fi
    
    # Check RDMA PCIe width, speed, and state
    local rdma_result
    rdma_result=$(check_rdma_pcie_width)
    IFS='|' read -r rdma_success rdma_error rdma_width_counts rdma_speed_counts rdma_state_errors <<< "$rdma_result"
    
    if [[ "$rdma_success" != "true" ]]; then
        overall_success=false
        error_messages+=("RDMA: $rdma_error")
    fi
    
    # Determine final status
    local result_status="PASS"
    
    if [[ "$overall_success" != "true" ]]; then
        # Join error messages with "; "
        local joined_errors
        joined_errors=$(IFS='; '; echo "${error_messages[*]}")
        result_status="FAIL - ${joined_errors}"
    fi
    
    # Output result in JSON format using jq (like other scripts)
    jq -n \
        --arg status "$result_status" \
        '{
            "pcie_width_missing_lanes": {
                "status": $status
            }
        }'
    
    # Exit with appropriate code
    if [[ "$overall_success" == "true" ]]; then
        return 0
    else
        return 1
    fi
}

# Execute main function
main "$@"