#!/bin/bash

# Missing Interface Check Script
# This script checks for missing PCIe interfaces by looking for devices with revision 'ff'
# It outputs either "PASS" if no missing interfaces are found, or "FAIL" if any are detected

echo "Starting missing interface health check..."
echo "This will take about 1 minute to complete."

# Run the lspci command to check for missing interfaces
# lspci shows PCI devices, grep filters for revision 'ff' (missing/failed devices), wc counts them
echo "Checking for missing PCIe interfaces..."
missing_count=$(lspci | grep -i 'rev ff' | wc -l 2>&1)

# Check if the command failed
if [ $? -ne 0 ]; then
    echo "Error: Could not run lspci command"
    jq -n '{"missing_interface": {"status": "FAIL", "error": "Command failed"}}'
    exit 1
fi

# Start with PASS status - we'll change to FAIL if we find missing interfaces
status="PASS"

# Check if any missing interfaces were found
if [ "$missing_count" -gt 0 ]; then
    echo "Found $missing_count missing interface(s)"
    status="FAIL"
else
    echo "No missing interfaces found"
fi

# Print the final result
jq -n \
      --arg status "$status" \
      --argjson missing_count "$missing_count" \
      '{
        "missing_interface": {
          "status": $status,
          "missing_count": $missing_count
        }
      }'