#!/bin/bash

##
## Check if GPU clock is set to max
##

MAX_CLOCK_SPEED=1980

gpu_clock_set_to_max() {
    # Get the current clock speed
    current_clock_list=$(nvidia-smi --query-gpu=clocks.current.graphics --format=csv,noheader,nounits)

    # Check if the command output is empty
    if [[ -z "$current_clock_list" ]]; then
        echo "Nvidia SMI command failed or returned no output. Please reboot the system and try again."
        exit 1
    fi
    
    local gpu_num=0
    local allowed_clock_speed=$(echo "$MAX_CLOCK_SPEED * 0.9" | bc) # 10% less than max clock speed
    

    for current_clock in $current_clock_list; do
        # Check if the current clock speed is less than the allowed clock speed
        echo "Checking GPU ${gpu_num} clock speed: $current_clock MHz against allowed threshold: $allowed_clock_speed MHz"
        if (( $(echo "$current_clock < $allowed_clock_speed" | bc -l) )); then
            echo ""
            echo "Error: GPU ${gpu_num} clock speed is set to $current_clock MHz, which is below the allowed threshold of $ALLOWED_CLOCK_SPEED MHz."
            echo "Set GPU clock speed to maximum ($MAX_CLOCK_SPEED MHz) with following commands:"
            
            echo "sudo nvidia-smi -pm 1 # Enable persistence mode"
            echo "sudo nvidia-smi -i ${gpu_num} -ac 1980 # Example command to set clock speed"  
            echo "sudo nvidia-smi -i ${gpu_num} -acp UNRESTRICTED # Example command to set power limit"
        
          else
            echo "GPU ${gpu_num} clock speed is set to $current_clock MHz, which is within the allowed range."
        fi
        ((gpu_num++))
    done

}

gpu_clock_set_to_max

