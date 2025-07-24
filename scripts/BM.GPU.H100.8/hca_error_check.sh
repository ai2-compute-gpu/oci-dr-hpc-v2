#!/bin/bash

# HCA Error Check Script
# This script checks for HCA (Host Channel Adapter) fatal errors by reading MLX5 messages from dmesg
# It outputs either "PASS" if no fatal errors are found, or "FAIL" if MLX5 fatal errors are detected

echo "Starting HCA error check..."
echo "This will take about 1 minute to complete."

# Run the dmesg command to get MLX5 fatal error messages
# dmesg -T shows timestamped kernel messages, filtered for mlx5 and Fatal errors
echo "Checking for MLX5 fatal errors..."
hca_output=$(sudo dmesg -T | grep mlx5 | grep 'Fatal' 2>&1)

# Check if the dmesg command failed
if [ $? -ne 0 ]; then
    echo "Error: Could not run dmesg command"
    jq -n '{"hca_error": {"status": "FAIL"}}'
    exit 1
fi

# For HCA check: if we get ANY output, that means fatal errors were found
# This is opposite logic from PCIe check
if [ -n "$hca_output" ]; then
    # Found fatal errors - check fails
    echo "Found MLX5 fatal errors:"
    echo "$hca_output"
    status="FAIL"
else
    # No fatal errors found - check passes
    echo "No MLX5 fatal errors found"
    status="PASS"
fi

# Print the final result
jq -n \
      --arg status "$status" \
      '{
        "hca_error": {
          "status": $status
        }
      }'