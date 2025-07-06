#!/bin/bash

# GPU Count Check Script
# This script checks if the correct H100 GPUs are present by comparing
# their PCI bus IDs with a predefined list of expected bus IDs

echo "Health check is in progress ..."

# Define expected PCI Bus IDs for H100 GPUs (converted to lowercase for comparison)
expected_pci_bus_ids=(
    "00000000:0f:00.0"
    "00000000:2d:00.0"
    "00000000:44:00.0"
    "00000000:89:00.0"
    "00000000:5b:00.0"
    "00000000:a8:00.0"
    "00000000:c0:00.0"
    "00000000:d8:00.0"
)

# Get the expected GPU count
expected_gpu_count=${#expected_pci_bus_ids[@]}

# Run nvidia-smi to get the actual PCI bus IDs
nvidia_output=$(nvidia-smi --query-gpu=pci.bus_id --format=csv,noheader)

# Check if the command failed
if [ $? -ne 0 ]; then
    echo "Error running nvidia-smi: $nvidia_output"
    jq -n '{"gpu": {"device_count": "FAIL"}}'
    exit 1
fi

# Convert nvidia-smi output to array, removing empty lines
actual_pci_bus_ids=()
while IFS= read -r line; do
    if [ -n "$line" ]; then
        # Convert to lowercase for comparison
        actual_pci_bus_ids+=("${line,,}")
    fi
done <<< "$nvidia_output"

# Get actual GPU count
actual_gpu_count=${#actual_pci_bus_ids[@]}

# Check for GPUs with wrong bus IDs
fail_list=()
for actual_bus_id in "${actual_pci_bus_ids[@]}"; do
    found_match=false
    for expected_bus_id in "${expected_pci_bus_ids[@]}"; do
        if [ "$actual_bus_id" = "$expected_bus_id" ]; then
            found_match=true
            break
        fi
    done

    # If this bus ID is not in our expected list, add it to fail list
    if [ "$found_match" = false ]; then
        fail_list+=("$actual_bus_id")
    fi
done

# Determine final result
if [ ${#fail_list[@]} -eq 0 ] && [ "$actual_gpu_count" -eq "$expected_gpu_count" ]; then
    # All GPUs found with correct bus IDs
    echo "SUCCESS: All GPUs present with correct bus IDs"
    result_status="PASS"

elif [ "$actual_gpu_count" -eq "$expected_gpu_count" ]; then
    # Same number of GPUs but wrong bus IDs
    echo "ERROR: Found GPUs with wrong bus IDs"
    fail_type="wrong busid"
    fail_ids=$(IFS=','; echo "${fail_list[*]}")
    result_status="FAIL - $fail_type: $fail_ids"

else
    # Different number of GPUs (missing GPUs)
    echo "ERROR: Missing GPUs"
    fail_type="missing"

    # Find which expected bus IDs are missing
    missing_ids=()
    for expected_bus_id in "${expected_pci_bus_ids[@]}"; do
        found_match=false
        for actual_bus_id in "${actual_pci_bus_ids[@]}"; do
            if [ "$actual_bus_id" = "$expected_bus_id" ]; then
                found_match=true
                break
            fi
        done
        if [ "$found_match" = false ]; then
            missing_ids+=("$expected_bus_id")
        fi
    done

    fail_ids=$(IFS=','; echo "${missing_ids[*]}")
    result_status="FAIL - $fail_type: $fail_ids"
fi

# Output result in JSON format
jq -n \
    --arg status "$result_status" \
    '{
        "gpu": {
            "device_count": $status
        }
    }'