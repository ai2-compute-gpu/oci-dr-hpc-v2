#!/bin/bash

# Network Interface RX Discards Health Check Script
# This script checks if network interfaces have too many dropped packets
# It works on two types of servers:
# - Standalone servers: Use regular Ethernet ports (enp12s0f0, etc.)
# - OKE nodes: Use RDMA high-speed ports (rdma0, rdma1, etc.)

# Check what type of server this is
is_oke_node=false

# Look at command line arguments to see if user specified OKE node
for arg in "$@"; do
    if [[ "$arg" == "-o" ]] || [[ "$arg" == "--is-oke-node" ]]; then
        is_oke_node=true
        echo "Running in OKE node mode"
        break
    elif [[ "$arg" == "-h" ]] || [[ "$arg" == "--help" ]]; then
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  -o, --is-oke-node    Check OKE/RDMA interfaces instead of regular Ethernet"
        echo "  -h, --help           Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0                   # Check regular Ethernet interfaces"
        echo "  $0 --is-oke-node     # Check OKE RDMA interfaces"
        exit 0
    fi
done

if [[ "$is_oke_node" == "false" ]]; then
    echo "Running in standalone server mode"
fi

# Define which network interfaces to check
if [[ "$is_oke_node" == "true" ]]; then
    # OKE nodes use RDMA interfaces for high-speed networking
    interfaces=(
        "rdma0"  "rdma1"  "rdma2"  "rdma3"
        "rdma4"  "rdma5"  "rdma6"  "rdma7"
        "rdma8"  "rdma9"  "rdma10" "rdma11"
        "rdma12" "rdma13" "rdma14" "rdma15"
    )
else
    # Standalone servers use regular Ethernet interfaces
    interfaces=(
        "enp12s0f0"  "enp12s0f1"    # PCIe slot 12, ports 0 and 1
        "enp42s0f0"  "enp42s0f1"    # PCIe slot 42, ports 0 and 1
        "enp65s0f0"  "enp65s0f1"    # PCIe slot 65, ports 0 and 1
        "enp88s0f0"  "enp88s0f1"    # PCIe slot 88, ports 0 and 1
        "enp134s0f0" "enp134s0f1"   # PCIe slot 134, ports 0 and 1
        "enp165s0f0" "enp165s0f1"   # PCIe slot 165, ports 0 and 1
        "enp189s0f0" "enp189s0f1"   # PCIe slot 189, ports 0 and 1
        "enp213s0f0" "enp213s0f1"   # PCIe slot 213, ports 0 and 1
    )
fi

# Set the threshold for acceptable packet drops
# If an interface drops more than this many packets, we consider it problematic
rx_discards_threshold=100

# Start the health check
echo "Health check is in progress ..."

# Initialize empty array to store results for jq
results_array=()

# Check each network interface one by one
for interface in "${interfaces[@]}"; do
    # Default status is PASS - we'll change it to FAIL if we find problems
    status="PASS"

    # Capture the output of the ethtool command
    ethtool_output=$(sudo ethtool -S "$interface" 2>&1 | grep "rx_prio.*_discards" || true)

    # Check if the command worked and found discard statistics
    if [[ -z "$ethtool_output" ]]; then
        status="FAIL"

        # Process each line of discard statistics
        # Each line looks like: "rx_prio0_discards: 42"
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                echo "    $line"

                # Extract the number after the colon
                # Remove all spaces and split on colon, take the second part
                discard_value=$(echo "$line" | tr -d ' ' | cut -d':' -f2)

                # Check if the discard value is a valid number
                if [[ "$discard_value" =~ ^[0-9]+$ ]]; then
                    if [[ "$discard_value" -gt "$rx_discards_threshold" ]]; then
                        status="FAIL"
                        break  # Stop checking this interface since we found a problem
                    fi
                else
                    status="FAIL"
                    break  # Stop checking this interface since we found a problem
                fi
            fi
        done <<< "$ethtool_output"
    fi

    # Add the result to our array as a JSON string
    results_array+=("{\"rx_discards\":{\"device\":\"$interface\",\"status\":\"$status\"}}")
done

# Use jq to create properly formatted JSON output
printf '%s\n' "${results_array[@]}" | jq -s '.'
