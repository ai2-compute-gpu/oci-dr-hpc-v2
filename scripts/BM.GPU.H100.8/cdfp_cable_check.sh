#!/bin/bash

# CDFP Cable Check Script
# This script checks if CDFP cables are correctly cabled by validating GPU PCI addresses 
# and module IDs against expected configurations for H100 GPUs.
# Uses nvidia-smi to query both PCI bus IDs and module IDs for accurate hardware validation.

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

# Define expected GPU module IDs for H100 GPUs (matches shapes.json configuration)
expected_gpu_module_ids=(
    "2" "4" "3" "1"
    "7" "5" "8" "6"
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

# Parse PCI bus IDs and module IDs from nvidia-smi -q output
actual_pci_bus_ids=()
actual_gpu_module_ids=()

# First pass: collect all PCI addresses in order
while IFS= read -r line; do
    line=$(echo "$line" | xargs)  # Trim whitespace
    
    # Look for Bus ID lines - format: "Bus Id                        : 00000000:0F:00.0"
    if [[ "$line" == *"Bus Id"* && "$line" == *":"* ]]; then
        # Extract Bus ID (take last 3 colon-separated parts)
        if [[ "$line" =~ :.*:.*:.*\. ]]; then
            bus_id=$(echo "$line" | sed 's/.*\([0-9a-fA-F]\{8\}:[0-9a-fA-F]\{2\}:[0-9a-fA-F]\{2\}\.[0-9a-fA-F]\).*/\1/')
            if [ -n "$bus_id" ] && [ "$bus_id" != "::" ]; then
                normalized_pci=$(normalize_pci_address "$bus_id")
                actual_pci_bus_ids+=("$normalized_pci")
            fi
        fi
    fi
done <<< "$gpu_query_output"

# Second pass: collect all module IDs using the exact working command
# nvidia-smi -q | grep -i "Module ID" | awk '{print $4}'
module_output=$(nvidia-smi -q | grep -i "Module ID" | awk '{print $4}')

# Parse module IDs
while IFS= read -r line; do
    if [ -n "$line" ]; then
        actual_gpu_module_ids+=("$line")
    fi
done <<< "$module_output"

# If no module IDs found, use sequential numbering
if [ ${#actual_gpu_module_ids[@]} -eq 0 ]; then
    echo "No module IDs found, using sequential numbering"
    for ((i=1; i<=${#actual_pci_bus_ids[@]}; i++)); do
        actual_gpu_module_ids+=("$i")
    done
fi

# Check if we found any GPUs
if [ ${#actual_pci_bus_ids[@]} -eq 0 ]; then
    echo "Error: No GPUs found"
    jq -n '{"gpu": {"cdfp": "FAIL - No GPUs detected"}}'
    exit 1
fi

# Check if PCI address count matches module ID count
if [ ${#actual_pci_bus_ids[@]} -ne ${#actual_gpu_module_ids[@]} ]; then
    echo "Error: Mismatch between PCI address count (${#actual_pci_bus_ids[@]}) and module ID count (${#actual_gpu_module_ids[@]})"
    jq -n '{"gpu": {"cdfp": "FAIL - Missing input data"}}'
    exit 1
fi

# Create associative arrays for mapping
declare -A expected_mapping
declare -A actual_mapping

# Populate expected mapping
for i in "${!expected_pci_bus_ids[@]}"; do
    expected_pci_normalized=$(normalize_pci_address "${expected_pci_bus_ids[$i]}")
    expected_mapping["$expected_pci_normalized"]="${expected_gpu_module_ids[$i]}"
done

# Populate actual mapping
for i in "${!actual_pci_bus_ids[@]}"; do
    actual_mapping["${actual_pci_bus_ids[$i]}"]="${actual_gpu_module_ids[$i]}"
done

# Validate each expected PCI and GPU module ID pair
fail_list=()
for expected_pci in "${!expected_mapping[@]}"; do
    expected_module_id="${expected_mapping[$expected_pci]}"
    actual_module_id="${actual_mapping[$expected_pci]}"
    
    if [ -z "$actual_module_id" ]; then
        fail_list+=("Expected GPU with PCI Address $expected_pci not found")
    elif [ "$actual_module_id" != "$expected_module_id" ]; then
        fail_list+=("Mismatch for PCI $expected_pci: Expected GPU module ID $expected_module_id, found $actual_module_id")
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