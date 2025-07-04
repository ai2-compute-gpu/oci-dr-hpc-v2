#!/bin/bash

# Function to run commands and capture the output
run_cmd() {
    cmd="$1"
    result=$(eval "$cmd" 2>&1)
    ret_code=$?

    # If the command fails, return the error message
    if [ $ret_code -ne 0 ]; then
        echo "Error: $cmd $ret_code $result"
        return 1
    fi

    # Otherwise, return the output of the command
    echo "$result"
}

# Function to parse results and determine pass/fail status
parse_acc_results() {
    host="$1"
    pci_id="$2"
    results="$3"
    shape="$4"

    result="host=$host"
    pcie_config="pci_busid=$pci_id max_acc_out=FAIL advanced_pci_settings=FAIL"

    # Check for "MAX_ACC_OUT_READ" values to set max_acc_out
    if echo "$results" | grep -q "MAX_ACC_OUT_READ.*0"; then
        pcie_config=$(echo "$pcie_config" | sed 's/max_acc_out=FAIL/max_acc_out=PASS/')
    elif echo "$results" | grep -q "MAX_ACC_OUT_READ.*44"; then
        pcie_config=$(echo "$pcie_config" | sed 's/max_acc_out=FAIL/max_acc_out=PASS/')
    elif echo "$results" | grep -q "MAX_ACC_OUT_READ.*128"; then
        pcie_config=$(echo "$pcie_config" | sed 's/max_acc_out=FAIL/max_acc_out=PASS/')
    fi

    # Check for "ADVANCED_PCI_SETTINGS" value to set advanced_pci_settings
    if echo "$results" | grep -q "ADVANCED_PCI_SETTINGS.*True"; then
        pcie_config=$(echo "$pcie_config" | sed 's/advanced_pci_settings=FAIL/advanced_pci_settings=PASS/')
    fi

    # Return the result as key-value pairs
    echo "$result pcie_config={$pcie_config}"
}

# Main function to run max_acc_check for each PCI device
run_max_acc_check() {
    host="hpc-node-1"
    mlxconfig_bin="/usr/bin/mlxconfig"
    shape="BM.GPU.H100.8"
    pci_ids=(
        "0000:0c:00.0"
        "0000:2a:00.0"
        "0000:41:00.0"
        "0000:58:00.0"
        "0000:86:00.0"
        "0000:a5:00.0"
        "0000:bd:00.0"
        "0000:d5:00.0"
    )

    pci_config_results=""

    for pci in "${pci_ids[@]}"; do
        cmd="sudo $mlxconfig_bin -d $pci query"
        output=$(run_cmd "$cmd")

        # Parse the results for this PCI device
        result=$(parse_acc_results "$host" "$pci" "$output" "$shape")
        pci_config_results="$pci_config_results $result"
    done

    # Return the result in a key-value format
    echo "host=$host pcie_config=[$pci_config_results]"
}

# Run the main function and print the result
result=$(run_max_acc_check)
echo "$result"
