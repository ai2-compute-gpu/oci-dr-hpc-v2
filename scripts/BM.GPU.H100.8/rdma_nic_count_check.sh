#!/bin/bash

##
## Check RDMA NIC count
##

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

KNOWN_RDMA_NIC_DEVICES=("0000:0c:00.0" "0000:0c:00.1" "0000:2a:00.0" "0000:2a:00.1" "0000:41:00.0" "0000:41:00.1" "0000:58:00.0" "0000:58:00.1" "0000:86:00.0" "0000:86:00.1" "0000:a5:00.0" "0000:a5:00.1" "0000:bd:00.0" "0000:bd:00.1" "0000:d5:00.0" "0000:d5:00.1")

rdma_nic_count_check() {
    # Result array for missing elements
    missing=()
    
    for device in "${KNOWN_RDMA_NIC_DEVICES[@]}"; do
        local cmd_output=$(sudo lspci -v -s "$device")
        
        if [[ -z "$cmd_output" ]]; then
            missing+=("$device")
        else
            local rdma_nic_check=$(echo "$cmd_output" | grep -c "Mellanox Technologies")
            if [[ $rdma_nic_check -eq 0 ]]; then
                missing+=("$device")
            fi
        fi
    done

    # Print missing elements
    if [ ${#missing[@]} -eq 0 ]; then
        echo "All RDMA NIC devices are known and accounted for."
        return
    else
        for device in "${missing[@]}"; do
            echo "Missing RDMA NIC device: $device"
        done
    fi
}

rdma_nic_count_check