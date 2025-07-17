#!/bin/bash

# Simple Peermem Module Health Check Script
# This script checks for the presence of the nvidia_peermem module on the system.
# The nvidia_peermem module is required for GPU-to-GPU peer memory access,
# which is crucial for high-performance computing applications.
# Missing this module can lead to performance degradation in multi-GPU setups.

echo "Health check is in progress ..."

# Module name to check for
module_name="nvidia_peermem"

# Get loaded modules information
# Use lsmod to list all currently loaded kernel modules
module_output=$(/usr/sbin/lsmod 2>/dev/null)

# Start with FAIL status
status="FAIL"

# Check if we got any output from lsmod command
if [ -n "$module_output" ]; then
    # Check each line of module output
    while IFS= read -r line; do
        # Skip empty lines
        if [ -n "$line" ]; then
            # Extract the module name (1st column in the output)
            # Example line: "nvidia_peermem    16384  0"
            current_module=$(echo "$line" | awk '{print $1}')

            # Check if this matches our target module
            if [ "$current_module" = "$module_name" ]; then
                status="PASS"
                break
            fi
        fi
    done <<< "$module_output"
fi

# Create and output JSON result using jq
jq -n \
    --arg status "$status" \
    '{
        "gpu": {
            "nvidia_peermem": $status
        }
    }'