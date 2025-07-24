#!/bin/bash

# GPU XID Error Check Script
# This script checks for NVIDIA GPU XID errors in system logs by scanning dmesg output.
# XID errors indicate various GPU hardware/software issues, with critical errors being
# more severe than warnings. The script parses XID error codes and outputs JSON results.

echo "GPU XID error check is in progress ..."

# Function to get XID severity and description
get_xid_info() {
    local xid_code="$1"
    case "$xid_code" in
        1|2|3|4|5|6|7|8|9|10|11|12|13) echo "Critical|Hardware/Software error" ;;
        31) echo "Critical|GPU memory page fault" ;;
        48) echo "Critical|Double Bit ECC Error" ;;
        56|57|58) echo "Critical|Display/Memory interface error" ;;
        62|63|64|65) echo "Critical|ECC/Controller error" ;;
        68|69) echo "Critical|NVDEC/Graphics Engine error" ;;
        73|74) echo "Critical|NVENC2/NVLINK Error" ;;
        79|80|81) echo "Critical|GPU bus/data error" ;;
        92|94|95) echo "Critical|ECC error" ;;
        109) echo "Critical|Context Switch Timeout Error" ;;
        119|120|121) echo "Critical|GSP/C2C Error" ;;
        14|15|16|18|19|24|25|26|27|28|29|30) echo "Warn|Display/Processing warning" ;;
        32|33|34|35|36|37|38|39|40|41|42|43|44|45|46|47) echo "Warn|Processing warning" ;;
        59|60|61|66|67|70|71|72|75|76|77|78|82|83|84|85|86|87|88|89) echo "Warn|Driver/Hardware warning" ;;
        93|110|140|143) echo "Warn|System warning" ;;
        *) echo "Warn|Unknown XID error" ;;
    esac
}

# Get dmesg output
dmesg_output=$(sudo dmesg 2>/dev/null)

# Check if dmesg command failed
if [ $? -ne 0 ]; then
    echo "Error: Failed to get dmesg output"
    jq -n '{
        "gpu_xid": {
            "status": "ERROR",
            "message": "Failed to get dmesg output"
        }
    }'
    exit 2
fi

# Initialize counters and arrays
critical_errors=()
warning_errors=()
critical_count=0
warning_count=0

# Check if any XID errors exist in dmesg
if echo "$dmesg_output" | grep -q "NVRM: Xid"; then
    # Parse XID errors - look for common critical XID codes
    for xid_code in 1 2 3 4 5 6 7 8 9 10 11 12 13 31 48 56 57 58 62 63 64 65 68 69 73 74 79 80 81 92 94 95 109 119 120 121 14 15 16 18 19 24 25 26 27 28 29 30 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 59 60 61 66 67 70 71 72 75 76 77 78 82 83 84 85 86 87 88 89 93 110 140 143; do
        # Look for this specific XID code in dmesg
        xid_matches=$(echo "$dmesg_output" | grep -o "NVRM: Xid (PCI:[^:]*: $xid_code," | wc -l)
        
        if [ "$xid_matches" -gt 0 ]; then
            # Get PCI addresses for this XID code
            pci_addresses=$(echo "$dmesg_output" | grep "NVRM: Xid (PCI:[^:]*: $xid_code," | sed -n 's/.*NVRM: Xid (PCI:\([^:]*\): '"$xid_code"',.*/\1/p' | sort -u)
            
            # Get severity and description
            xid_info_raw=$(get_xid_info "$xid_code")
            IFS='|' read -r severity description <<< "$xid_info_raw"
            
            # Create XID info object
            xid_info="{\"xid_code\": \"$xid_code\", \"description\": \"$description\", \"severity\": \"$severity\", \"count\": $xid_matches, \"pci_addresses\": [$(echo "$pci_addresses" | sed 's/.*/"&"/' | paste -sd ',' -)]}"
            
            if [ "$severity" = "Critical" ]; then
                critical_errors+=("$xid_info")
                ((critical_count++))
            else
                warning_errors+=("$xid_info")
                ((warning_count++))
            fi
        fi
    done
fi

# Determine result status and create JSON output
if [ $critical_count -gt 0 ]; then
    # Critical errors found - test fails
    status="FAIL"
    message="Critical XID errors detected: $critical_count critical, $warning_count warnings"
    exit_code=1
elif [ $warning_count -gt 0 ]; then
    # Only warnings found
    status="WARN"
    message="Warning XID errors detected: $warning_count warnings"
    exit_code=0
else
    # No XID errors found
    if echo "$dmesg_output" | grep -q "NVRM: Xid"; then
        # XID messages found but no recognized error codes
        status="WARN"
        message="XID messages found but no recognized error codes"
        exit_code=0
    else
        # No XID errors at all
        status="PASS"
        message="No XID errors found in system logs"
        exit_code=0
    fi
fi

# Build JSON output
json_output="{\"gpu_xid\": {\"status\": \"$status\", \"message\": \"$message\""

# Add error arrays if they exist
if [ $critical_count -gt 0 ]; then
    critical_array=$(IFS=','; echo "${critical_errors[*]}")
    json_output="$json_output, \"critical_errors\": [$critical_array]"
fi

if [ $warning_count -gt 0 ]; then
    warning_array=$(IFS=','; echo "${warning_errors[*]}")
    json_output="$json_output, \"warning_errors\": [$warning_array]"
fi

json_output="$json_output}}"

# Output JSON result
echo "$json_output" | jq .

# Exit with appropriate code
exit $exit_code