#!/bin/bash

# GPU SRAM Memory Health Check Script
#
# This script monitors GPU memory health by checking SRAM (Static Random Access Memory)
# error counts using nvidia-smi. SRAM errors indicate potential GPU hardware issues:
#
# - Uncorrectable errors: Critical hardware failures requiring immediate attention
# - Correctable errors: ECC memory corrections indicating possible memory degradation
#
# The script compares error counts against predefined thresholds to determine GPU health status.


# Configuration: Set error count thresholds for health assessment
readonly UNCORRECTABLE_THRESHOLD=5      # Critical: Any uncorrectable errors indicate serious hardware issues
readonly CORRECTABLE_THRESHOLD=1000     # Warning: High correctable errors suggest memory degradation

# Start health check process
echo "Health check is in progress ..."

# Execute nvidia-smi command to get uncorrectable & correctable error data
uncorrectable_output=$(sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable 2>/dev/null)
correctable_output=$(sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable 2>/dev/null)


# Parse uncorrectable error counts from nvidia-smi output
declare -a uncorrectable_errors=()
parity_errors=0

# Process each line of uncorrectable error output
while IFS= read -r line; do
    # Only process lines that contain a colon (indicating key-value pairs)
    if [[ "$line" == *":"* ]]; then
        # Extract the value after the colon and remove whitespace
        error_value=$(echo "$line" | cut -d':' -f2 | tr -d ' ')

        # Check if the value is a valid number
        if [[ "$error_value" =~ ^[0-9]+$ ]]; then
            if [[ "$line" == *"Parity"* ]]; then
                # Store parity errors separately - they will be added to SEC-DED errors
                parity_errors=$error_value
            elif [[ "$line" == *"SEC-DED"* ]]; then
                # SEC-DED: Single Error Correction, Double Error Detection
                # Combine SEC-DED errors with parity errors for total uncorrectable count
                total_uncorrectable=$((error_value + parity_errors))
                uncorrectable_errors+=("$total_uncorrectable")
            else
                # Other types of uncorrectable errors
                uncorrectable_errors+=("$error_value")
            fi
        fi
    fi
done <<< "$uncorrectable_output"

# Initialize array for correctable error counts
declare -a correctable_errors=()

# Process each line of correctable error output
while IFS= read -r line; do
    # Only process lines that contain a colon
    if [[ "$line" == *":"* ]]; then
        # Extract the numeric value after the colon
        error_value=$(echo "$line" | cut -d':' -f2 | tr -d ' ')

        # Validate that the value is numeric
        if [[ "$error_value" =~ ^[0-9]+$ ]]; then
            correctable_errors+=("$error_value")
        fi
    fi
done <<< "$correctable_output"

# Analyze error counts and determine health status
health_status="PASS"

# Check if we have error data to analyze
if [[ ${#uncorrectable_errors[@]} -eq 0 ]] || [[ ${#correctable_errors[@]} -eq 0 ]]; then
    health_status="FAIL"
else
    # Check each uncorrectable error count against threshold
    for error_count in "${uncorrectable_errors[@]}"; do
        if [[ "$error_count" -gt "$UNCORRECTABLE_THRESHOLD" ]]; then
            health_status="FAIL"
            break  # Stop checking once we find a critical error
        fi
    done

    # Check each correctable error count against threshold
    for error_count in "${correctable_errors[@]}"; do
        if [[ "$error_count" -gt "$CORRECTABLE_THRESHOLD" ]]; then
            # Only set to WARN if not already FAIL
            if [[ "$health_status" == "PASS" ]]; then
                health_status="WARN - SRAM Correctable Exceeded Threshold"
            fi
        fi
    done
fi

# Also provide simple JSON output for basic monitoring systems
jq -n \
    --arg status "$health_status" \
    '{
        "sram": {
            "status": $status
        }
    }'