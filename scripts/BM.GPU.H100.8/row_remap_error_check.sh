#!/bin/bash

# Row Remap Error Check Script
# This script checks for GPU row remap errors by querying remapped rows failure count
# Row remap failures indicate memory errors in the GPU

echo "Starting row remap error health check..."
echo "This will take about 1 minute to complete."

# Check nvidia-smi driver version first
driver_version=$(nvidia-smi --query-gpu=driver_version --format=csv,noheader,nounits 2>/dev/null | head -1 | grep -o '^[0-9]*')

if [ -z "$driver_version" ]; then
    echo "Error: Could not determine nvidia-smi driver version"
    jq -n '{"row_remap_error": {"status": "FAIL", "error": "could not determine nvidia-smi driver version"}}'
    exit 1
fi

if [ "$driver_version" -lt 550 ]; then
    echo "Driver version $driver_version is below required version 550"
    jq -n --arg version "$driver_version" '{"row_remap_error": {"status": ("Not applicable for nvidia-smi driver : " + $version)}}'
    exit 0
fi

# Expected GPU bus IDs for H100 shape
expected_bus_ids=(
    "00000000:0F:00.0"
    "00000000:2D:00.0"
    "00000000:44:00.0"
    "00000000:5B:00.0"
    "00000000:89:00.0"
    "00000000:A8:00.0"
    "00000000:C0:00.0"
    "00000000:D8:00.0"
)

# Run nvidia-smi command to get row remap information
echo "Checking for GPU row remap errors..."
output=$(nvidia-smi --query-remapped-rows=gpu_bus_id,remapped_rows.failure --format=csv,noheader 2>&1)

# Check if command failed
if [ $? -ne 0 ]; then
    echo "Error: Could not run nvidia-smi command"
    jq -n '{"row_remap_error": {"status": "FAIL", "error": "nvidia-smi command failed"}}'
    exit 1
fi

# Check if output is empty
if [ -z "$output" ]; then
    echo "Error: No nvidia-smi output received"
    jq -n '{"row_remap_error": {"status": "FAIL", "error": "No nvidia-smi output received"}}'
    exit 1
fi

# Arrays to track results
declare -A found_bus_ids
failed_bus_ids=()
missing_bus_ids=()

# Parse the CSV output
while IFS= read -r line; do
    line=$(echo "$line" | xargs)  # Trim whitespace
    
    # Skip empty lines or error lines
    if [[ -z "$line" || "$line" == Error:* ]]; then
        continue
    fi
    
    # Parse CSV format: gpu_bus_id, remapped_rows.failure
    IFS=',' read -r bus_id failure_count <<< "$line"
    bus_id=$(echo "$bus_id" | xargs)  # Trim whitespace
    failure_count=$(echo "$failure_count" | xargs)  # Trim whitespace
    
    # Mark this bus ID as found
    found_bus_ids["$bus_id"]=1
    
    # Check if failure count is not 0
    if [[ "$failure_count" != "0" ]]; then
        failed_bus_ids+=("$bus_id")
        echo "Found row remap failure on GPU: $bus_id (failures: $failure_count)"
    fi
done <<< "$output"

# Check for missing GPUs
for expected_id in "${expected_bus_ids[@]}"; do
    if [[ -z "${found_bus_ids[$expected_id]}" ]]; then
        missing_bus_ids+=("$expected_id")
        echo "Missing expected GPU: $expected_id"
    fi
done

# Determine overall status
status="PASS"
result_json='{"row_remap_error": {"status": "PASS"}}'

# Check for failures
if [[ ${#failed_bus_ids[@]} -gt 0 || ${#missing_bus_ids[@]} -gt 0 ]]; then
    status="FAIL"
    
    # Build JSON result with failure details
    if [[ ${#failed_bus_ids[@]} -gt 0 && ${#missing_bus_ids[@]} -gt 0 ]]; then
        # Both failed and missing GPUs
        failed_list=$(printf '%s\n' "${failed_bus_ids[@]}" | jq -R . | jq -s .)
        missing_list=$(printf '%s\n' "${missing_bus_ids[@]}" | jq -R . | jq -s .)
        result_json=$(jq -n \
            --argjson failed "$failed_list" \
            --argjson missing "$missing_list" \
            '{"row_remap_error": {"status": "FAIL", "failed_gpus": $failed, "missing_gpus": $missing}}')
    elif [[ ${#failed_bus_ids[@]} -gt 0 ]]; then
        # Only failed GPUs
        failed_list=$(printf '%s\n' "${failed_bus_ids[@]}" | jq -R . | jq -s .)
        result_json=$(jq -n \
            --argjson failed "$failed_list" \
            '{"row_remap_error": {"status": "FAIL", "failed_gpus": $failed}}')
    else
        # Only missing GPUs
        missing_list=$(printf '%s\n' "${missing_bus_ids[@]}" | jq -R . | jq -s .)
        result_json=$(jq -n \
            --argjson missing "$missing_list" \
            '{"row_remap_error": {"status": "FAIL", "missing_gpus": $missing}}')
    fi
else
    echo "No row remap errors found"
fi

# Print the final result
echo "$result_json"