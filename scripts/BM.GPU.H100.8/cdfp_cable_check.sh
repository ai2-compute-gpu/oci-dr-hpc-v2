#!/bin/bash

# CDFP Cable Check Script
# This script checks if CDFP cables are correctly cabled by validating GPU PCI addresses 
# and module IDs against expected configurations for H100 GPUs

echo "Health check is in progress ..."

# Define expected PCI Bus IDs for H100 GPUs (converted to lowercase for comparison)
expected_pci_bus_ids=(
    "00000000:0f:00.0"
    "00000000:2d:00.0"
    "00000000:44:00.0"
    "00000000:5b:00.0"
    "00000000:89:00.0"
    "00000000:a8:00.0"
    "00000000:c0:00.0"
    "00000000:d8:00.0"
)

# Define expected GPU indices for H100 GPUs  
expected_gpu_indices=(
    "0" "1" "2" "3"
    "4" "5" "6" "7"
)

# Function to normalize PCI address
normalize_pci_address() {
    local pci_addr="$1"
    local normalized="${pci_addr,,}"  # Convert to lowercase
    
    # Handle cases where PCI address starts with "000000"
    if [[ "$pci_addr" == 000000* ]]; then
        normalized="00${pci_addr:6}"
        normalized="${normalized,,}"
    fi
    
    echo "$normalized"
}

# Get GPU information using nvidia-smi -q
gpu_query_output=$(nvidia-smi -q)
query_exit_code=$?

# Check if command failed
if [ $query_exit_code -ne 0 ]; then
    echo "Error running nvidia-smi -q command"
    jq -n '{"gpu": {"cdfp": "FAIL - Missing input data"}}'
    exit 1
fi

# Extract PCI bus IDs using the approach: nvidia-smi -q | grep -i "Bus Id" | awk '{print $4}'
pci_output=$(echo "$gpu_query_output" | grep -i "Bus Id" | awk '{print $4}')

# Extract GPU indices by parsing the GPU sections in order
gpu_index=0
actual_gpu_indices=()
actual_pci_bus_ids=()

# Parse PCI bus IDs and assign GPU indices
gpu_index=0
while IFS= read -r line; do
    if [ -n "$line" ]; then
        # Normalize PCI address
        normalized_pci=$(normalize_pci_address "$line")
        actual_pci_bus_ids+=("$normalized_pci")
        actual_gpu_indices+=("$gpu_index")
        ((gpu_index++))
    fi
done <<< "$pci_output"

# Check if we found any GPUs
if [ ${#actual_pci_bus_ids[@]} -eq 0 ]; then
    echo "Error: No GPUs found"
    jq -n '{"gpu": {"cdfp": "FAIL - No GPUs detected"}}'
    exit 1
fi

# Create associative arrays for mapping
declare -A expected_mapping
declare -A actual_mapping

# Populate expected mapping
for i in "${!expected_pci_bus_ids[@]}"; do
    expected_pci_normalized=$(normalize_pci_address "${expected_pci_bus_ids[$i]}")
    expected_mapping["$expected_pci_normalized"]="${expected_gpu_indices[$i]}"
done

# Populate actual mapping
for i in "${!actual_pci_bus_ids[@]}"; do
    actual_mapping["${actual_pci_bus_ids[$i]}"]="${actual_gpu_indices[$i]}"
done

# Validate each expected PCI and GPU index pair
fail_list=()
for expected_pci in "${!expected_mapping[@]}"; do
    expected_index="${expected_mapping[$expected_pci]}"
    actual_index="${actual_mapping[$expected_pci]}"
    
    if [ -z "$actual_index" ]; then
        fail_list+=("Expected GPU with PCI Address $expected_pci not found")
    elif [ "$actual_index" != "$expected_index" ]; then
        fail_list+=("Mismatch for PCI $expected_pci: Expected GPU index $expected_index, found $actual_index")
    fi
done

# Determine final result
if [ ${#fail_list[@]} -eq 0 ]; then
    # All CDFP cables correctly connected
    echo "SUCCESS: All CDFP cables correctly connected"
    result_status="PASS"
else
    # CDFP cable mismatches found
    echo "ERROR: CDFP cable mismatches detected"
    fail_details=$(IFS=', '; echo "${fail_list[*]}")
    result_status="FAIL - $fail_details"
fi

# Output result in JSON format
jq -n \
    --arg status "$result_status" \
    '{
        "gpu": {
            "cdfp": $status
        }
    }'