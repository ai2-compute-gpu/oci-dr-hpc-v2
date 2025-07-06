#!/bin/bash

# RDMA NIC Count Check Script
# This script verifies that the correct number of RDMA (Remote Direct Memory Access)
# Network Interface Cards are present on an H100 GPU system. It checks specific PCI
# bus locations for Mellanox RDMA controllers and validates the count against expected values.
#
# RDMA enables high-speed, low-latency network communication between GPUs in distributed
# computing environments, which is critical for multi-node AI/ML workloads.

echo "Health check is in progress ..."

# Define expected RDMA PCI device locations for H100 GPU systems
# These are the specific PCI bus addresses where RDMA NICs should be found
# Each H100 system typically has 16 RDMA NICs (8 pairs) for high-bandwidth networking
expected_rdma_pci_ids=(
    "0000:0c:00.0"   # RDMA NIC 1A
    "0000:0c:00.1"   # RDMA NIC 1B
    "0000:2a:00.0"   # RDMA NIC 2A
    "0000:2a:00.1"   # RDMA NIC 2B
    "0000:41:00.0"   # RDMA NIC 3A
    "0000:41:00.1"   # RDMA NIC 3B
    "0000:58:00.0"   # RDMA NIC 4A
    "0000:58:00.1"   # RDMA NIC 4B
    "0000:86:00.0"   # RDMA NIC 5A
    "0000:86:00.1"   # RDMA NIC 5B
    "0000:a5:00.0"   # RDMA NIC 6A
    "0000:a5:00.1"   # RDMA NIC 6B
    "0000:bd:00.0"   # RDMA NIC 7A
    "0000:bd:00.1"   # RDMA NIC 7B
    "0000:d5:00.0"   # RDMA NIC 8A
    "0000:d5:00.1"   # RDMA NIC 8B
)

# Get the expected RDMA NIC count (should be 16 for H100 systems)
expected_nics_count=${#expected_rdma_pci_ids[@]}

# Initialize counter for detected RDMA NICs
detected_rdma_nics=0

# Check each expected RDMA device location
for device in "${expected_rdma_pci_ids[@]}"; do
    lspci_output=$(sudo lspci -v -s "$device" 2>&1)

    # Check if the command failed
    if [ $? -ne 0 ]; then
        echo "  Error running lspci for device $device: $lspci_output"
        continue
    fi

    # Look through each line of the lspci output for this device
    while IFS= read -r line; do
        # Look for the Mellanox manufacturer identifier
        # Mellanox Technologies makes the RDMA controllers used in H100 systems
        if [[ "$line" == *"controller: Mellanox Technologies"* ]]; then
            detected_rdma_nics=$((detected_rdma_nics + 1))
            break  # Found the controller for this device, move to next device
        fi
    done <<< "$lspci_output"
done

# Determine if the check passed or failed
if [ "$detected_rdma_nics" -eq "$expected_nics_count" ]; then
    status="PASS"
else
    if [ "$detected_rdma_nics" -lt "$expected_nics_count" ]; then
        missing_count=$((expected_nics_count - detected_rdma_nics))
        echo "Missing $missing_count RDMA NICs"
    else
        extra_count=$((detected_rdma_nics - expected_nics_count))
        echo "Found $extra_count extra RDMA NICs"
    fi
    status="FAIL"
fi

# Output the final result in JSON format using jq for proper formatting
jq -n \
    --arg status "$status" \
    --argjson num_nics "$detected_rdma_nics" \
    '{
        "rdma_nic_count": {
            "status": $status,
            "num_rdma_nics": $num_nics
        }
    }'