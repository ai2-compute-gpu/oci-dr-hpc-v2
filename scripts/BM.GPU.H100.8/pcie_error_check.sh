#!/bin/bash

# PCIe Error Check Script
# This script checks for PCIe errors by reading system messages and looking for error patterns
# It outputs either "PASS" if no errors are found, or "FAIL" if PCIe errors are detected

echo "Starting PCIe health check..."
echo "This will take about 1 minute to complete."

# Run the dmesg command to get system messages
# dmesg shows kernel ring buffer messages including hardware errors
echo "Getting system messages..."
dmesg_output=$(sudo dmesg 2>&1)

# Check if the dmesg command failed
if [ $? -ne 0 ]; then
    echo "Error: Could not run dmesg command"
    jq -n '{"pcie_error": {"status": "FAIL"}}'
    exit 1
fi

# Check if dmesg output is empty
if [ -z "$dmesg_output" ]; then
    echo "Error: No system messages found"
    jq -n '{"pcie_error": {"status": "FAIL"}}'
    exit 1
fi

# Start with PASS status - we'll change to FAIL if we find errors
status="PASS"

# Look through each line of the dmesg output
echo "Checking for PCIe errors..."
while IFS= read -r line; do
    # Skip lines that contain "capabilities" - these are not error messages
    if [[ "$line" == *"capabilities"* ]]; then
        continue
    fi
    
    # Look for lines that contain both "pcieport" and "error" (case insensitive)
    # pcieport = PCIe port driver messages
    # error = indicates an actual error occurred
    if [[ "$line" =~ .*pcieport.*[Ee]rror ]]; then
        echo "Found PCIe error: $line"
        status="FAIL"
        break  # Stop checking once we find the first error
    fi
done <<< "$dmesg_output"

# Print the final result
jq -n \
      --arg status "$status" \
      '{
        "pcie_error": {
          "status": $status
        }
      }'