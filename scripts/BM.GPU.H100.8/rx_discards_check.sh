#!/bin/bash

# Network Interface RX Discards Health Check Script
# This script checks if network interfaces have too many dropped packets

interfaces=(
    "enp12s0f0"  "enp12s0f1"
    "enp42s0f0"  "enp42s0f1"
    "enp65s0f0"  "enp65s0f1"
    "enp88s0f0"  "enp88s0f1"
    "enp134s0f0" "enp134s0f1"
    "enp165s0f0" "enp165s0f1"
    "enp189s0f0" "enp189s0f1"
    "enp213s0f0" "enp213s0f1"
)


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
