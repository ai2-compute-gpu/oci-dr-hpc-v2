#!/bin/bash

# Simple NVLink speed checker
# Expected values
EXPECTED_SPEED=26
EXPECTED_COUNT=18

# Run nvidia-smi nvlink command
output=$(nvidia-smi nvlink -s 2>&1)

# Check if command failed
if [ $? -ne 0 ]; then
    echo '{"gpu": {"nvlink": "FAIL - nvidia-smi command failed"}}' | jq '.'
    exit 1
fi

# Check if output is empty
if [ -z "$output" ]; then
    echo '{"gpu": {"nvlink": "FAIL - check GPU"}}' | jq '.'
    exit 1
fi

# Count good links for each GPU
current_gpu=""
link_count=0
declare -A gpu_counts
fail_gpus=()

while IFS= read -r line; do
    # Look for GPU line like "GPU 0: NVIDIA"
    if echo "$line" | grep -q "GPU [0-9]*:.*NVIDIA\|GPU [0-9]*:.*HGX"; then
        current_gpu=$(echo "$line" | grep -o "GPU [0-9]*" | grep -o "[0-9]*")
        link_count=0

    # Look for Link line like "    Link 0: 25.781 GB/s" (not inactive)
    elif echo "$line" | grep -q "Link [0-9]*:.*GB/s" && ! echo "$line" | grep -q "inactive"; then
        link_speed=$(echo "$line" | grep -o "[0-9]*\.[0-9]*" | head -1)
        # Check if speed is good enough (convert to integer for comparison)
        speed_int=${link_speed%.*}  # Remove decimal part
        if [[ $speed_int -ge $EXPECTED_SPEED ]]; then
            ((link_count++))
        fi
    fi

    # Save count for this GPU
    if [[ -n "$current_gpu" ]]; then
        gpu_counts[$current_gpu]=$link_count
    fi
done <<< "$output"

# Check if any GPU has wrong number of links
for gpu in "${!gpu_counts[@]}"; do
    if [[ ${gpu_counts[$gpu]} -ne $EXPECTED_COUNT ]]; then
        fail_gpus+=("$gpu")
    fi
done

# Print result
if [[ ${#fail_gpus[@]} -gt 0 ]]; then
    fail_list=$(IFS=','; echo "${fail_gpus[*]}")
    echo "{\"gpu\": {\"nvlink\": \"FAIL - check GPU $fail_list\"}}" | jq '.'
else
    echo '{"gpu": {"nvlink": "PASS"}}' | jq '.'
fi