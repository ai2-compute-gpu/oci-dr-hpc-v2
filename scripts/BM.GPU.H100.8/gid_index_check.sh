#!/bin/bash

# Simple GID Index Health Check Script
# This script checks the GID index on a system to ensure it matches expected values.
# It is designed to run on systems with multiple GPUs, such as H100 GPU systems,
# and provides a health check for GPU configuration consistency.
# The GID index is crucial for RDMA (Remote Direct Memory Access) operations,
# and discrepancies can lead to performance issues.

echo "Health check is in progress ..."

# Expected GID index values (0, 1, 2, 3 are the expected defaults)
expected_gid_indices=(0 1 2 3)

# Get GID information from the system
# This command gets all GID entries, skips the header and footer lines
gid_output=$(sudo show_gids | tail -n +3 | head -n -1 2>/dev/null)

# Start with PASS status
status="PASS"

# Check if we got any output from show_gids command
if [ -z "$gid_output" ]; then
    # No output means command failed or no GIDs found
    status="FAIL"
else
    # Check each line of GID output
    while IFS= read -r line; do
        # Skip empty lines
        if [ -n "$line" ]; then
            # Extract the GID index (3rd column in the output)
            # Example line: "DEV  PORT  INDEX  GID                                      IPv4            VER   DEV"
            gid_index=$(echo "$line" | awk '{print $3}')

            # Check if the GID index is a valid number
            if [[ "$gid_index" =~ ^[0-9]+$ ]]; then
                # Check if this GID index is in our expected list
                found=false
                for expected in "${expected_gid_indices[@]}"; do
                    if [ "$gid_index" -eq "$expected" ]; then
                        found=true
                        break
                    fi
                done

                # If GID index is not in expected list, mark as FAIL
                if [ "$found" = false ]; then
                    status="FAIL"
                    break
                fi
            else
                # Invalid GID index means parsing error
                status="FAIL"
                break
            fi
        fi
    done <<< "$gid_output"
fi

# Create and output JSON result using jq
jq -n \
    --arg status "$status" \
    '{
        "gid_index": {
            "status": $status
        }
    }'

