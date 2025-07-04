#!/bin/bash

##
## Check if all known GPU PCI bus IDs are present in the system.
##
KNOWN_PCI_BUSIDS=("00000000:0F:00.0" "00000000:2D:00.0" "00000000:44:00.0" "00000000:5B:00.0" "00000000:89:00.0" "00000000:A8:00.0" "00000000:C0:00.0" "00000000:D8:00.0")

gpu_count_check() {
    cmd_output=$(nvidia-smi --query-gpu=pci.bus_id --format=csv,noheader)

    # Check if the command output is empty
    if [[ -z "$cmd_output" ]]; then
        echo "Nvidia SMI command failed or returned no output. Please reboot the system and try again."
        exit 1
    fi

    mapfile -t pci_bus_ids <<< "$cmd_output"

    if [ ${#pci_bus_ids[@]} -eq 0 ]; then
        echo "No GPU found. Please check the system configuration."
        exit 1
    fi

    # Result array for missing elements
    missing=()

    for item in "${KNOWN_PCI_BUSIDS[@]}"; do
        found=false
        for existing in "${pci_bus_ids[@]}"; do
            if [[ "$item" == "$existing" ]]; then
                found=true
                break
            fi
        done
        if [[ "$found" == false ]]; then
            missing+=("$item")
        fi
    done

    # Print missing elements
    if [ ${#missing[@]} -eq 0 ]; then
        echo "All GPU bus IDs are known and accounted for."
        return
    else
        for item in "${missing[@]}"; do
            echo "Missing PCI bus ID $item"
        done
    fi
}


gpu_count_check



