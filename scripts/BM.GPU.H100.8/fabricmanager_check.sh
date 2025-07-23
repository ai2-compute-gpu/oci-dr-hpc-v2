#!/bin/bash

# Fabricmanager Service Check Script
# This script checks if nvidia-fabricmanager service is enabled and running on the system.
# nvidia-fabricmanager is required for optimal GPU performance on systems with multiple GPUs
# and high-speed interconnects like NVLink and NVSwitch.

echo "Health check is in progress ..."

# Function to check nvidia-fabricmanager service status
check_fabricmanager_service() {
    # Use systemctl to check the service status
    service_status=$(systemctl status nvidia-fabricmanager 2>/dev/null)
    exit_code=$?
    
    # Check if command failed (service doesn't exist or systemctl not available)
    if [ $exit_code -ne 0 ]; then
        echo "ERROR: Failed to check nvidia-fabricmanager service status"
        return 1
    fi
    
    # Check if service is active (running)
    if echo "$service_status" | grep -q "active (running)"; then
        echo "SUCCESS: nvidia-fabricmanager service is active and running"
        return 0
    else
        # Get the actual status for debugging
        actual_status=$(echo "$service_status" | grep "Active:" | head -1)
        echo "ERROR: nvidia-fabricmanager service is not running - $actual_status"
        return 1
    fi
}

# Main check
if check_fabricmanager_service; then
    result_status="PASS"
else
    result_status="FAIL"
fi

# Output result in JSON format
jq -n \
    --arg status "$result_status" \
    '{
        "gpu": {
            "fabricmanager-service": $status
        }
    }'