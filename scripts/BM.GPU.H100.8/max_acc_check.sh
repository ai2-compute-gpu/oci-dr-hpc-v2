#!/bin/bash

# Define the path to mlxconfig
mlxconfig_bin="/usr/bin/mlxconfig"

# Define PCI IDs for H100
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

# Function to run command and capture output
run_cmd() {
    local cmd="$1"
    output=$(eval "$cmd" 2>&1)
    echo "$output"
}

# Function to parse mlxconfig output and emit JSON
parse_acc_results() {
    local pci_id="$1"
    local results="$2"

    local max_acc_out="FAIL"
    local advanced_pci_settings="FAIL"

    if echo "$results" | grep -qE "MAX_ACC_OUT_READ.*(0|44|128)"; then
        max_acc_out="PASS"
    fi

    if echo "$results" | grep -q "ADVANCED_PCI_SETTINGS.*True"; then
        advanced_pci_settings="PASS"
    fi

    jq -n \
        --arg pci_busid "$pci_id" \
        --arg max_acc_out "$max_acc_out" \
        --arg advanced_pci_settings "$advanced_pci_settings" \
        '{
            pci_busid: $pci_busid,
            max_acc_out: $max_acc_out,
            advanced_pci_settings: $advanced_pci_settings
        }'
}

# Main function to loop through PCI IDs and collect JSON objects
run_max_acc_check() {
    echo "Health check is in progress and the result will be provided within 1 minute."
    json_array="[]"

    for pci in "${pci_ids[@]}"; do
        cmd="sudo $mlxconfig_bin -d $pci query"
        output=$(run_cmd "$cmd")
        result=$(parse_acc_results "$pci" "$output")

        # Append result to JSON array
        json_array=$(echo "$json_array" | jq --argjson obj "$result" '. + [$obj]')
    done

    jq -n --argjson pcie_config "$json_array" '{pcie_config: $pcie_config}'

}

# Run the main function
run_max_acc_check