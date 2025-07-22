#!/bin/bash

# Eth0 Presence Check Script
# This script checks if the eth0 network interface is present on the system
# It outputs either "PASS" if eth0 is found, or "FAIL" if eth0 is not present

echo "Starting eth0 presence check..."
echo "This will take about 1 minute to complete."

# Run the ip addr command and grep for eth0
# ip addr shows all network interfaces on the system
echo "Checking for eth0 interface..."
ip_output=$(ip addr | grep eth0 2>&1)

# Check if the ip command failed
if [ $? -ne 0 ]; then
    echo "Error: Could not run ip addr command"
    jq -n '{"eth0_presence": {"status": "FAIL"}}'
    exit 1
fi

# Start with FAIL status - we'll change to PASS if we find eth0
status="FAIL"

# Check if eth0 was found in the output
if [ -n "$ip_output" ]; then
    echo "Found eth0 interface: $ip_output"
    status="PASS"
else
    echo "eth0 interface not found"
fi

# Print the final result
jq -n \
      --arg status "$status" \
      '{
        "eth0_presence": {
          "status": $status
        }
      }'