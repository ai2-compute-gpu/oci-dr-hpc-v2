"""
On H100 Systems with DGX OS 6.0, the MAX_ACC_OUT_READ configuration for the CX-7 controllers
can be too low and result in reduced data transfer rates.
This script validates that RDMA NICs are accessible and have expected settings for optimal data transfer rates.
MAX_ACC_OUT_READ must be set to either 0, 44, or 128.
ADVANCED_PCI_SETTINGS must be set to True.
"""

import shlex
import subprocess
import json

# Function to run command and capture output
def run_cmd(cmd):
    cmd_split = shlex.split(cmd)
    try:
        results = subprocess.run(cmd_split, shell=False, stdout=subprocess.PIPE,
                                 stderr=subprocess.STDOUT, check=True, encoding='utf8')
        output = results.stdout.splitlines()
    except subprocess.CalledProcessError as e_process_error:
        return [f"Error: {cmd} {e_process_error.returncode} {e_process_error.output}"]
    return output

# Function to parse mlxconfig output and emit JSON
def run_max_acc_check():
    # Define the path to mlxconfig
    mlxconfig_bin = "/usr/bin/mlxconfig"

    # Define PCI IDs for H100
    config = {
        "pci_ids": [
            "0000:0c:00.0",
            "0000:2a:00.0",
            "0000:41:00.0",
            "0000:58:00.0",
            "0000:86:00.0",
            "0000:a5:00.0",
            "0000:bd:00.0",
            "0000:d5:00.0"
            ]
    }
    pci_ids = config["pci_ids"]

    pci_config_results = []
    for pci in pci_ids:
        cmd = f'sudo {mlxconfig_bin} -d {pci} query'
        output = run_cmd(cmd)
        result = parse_acc_results(pci, output)
        pci_config_results.append(result["pcie_config"])
    result = dict(pcie_config=pci_config_results)
    return result

# Main function to loop through PCI IDs and collect JSON objects
def parse_acc_results(pci_id="undefined", results="undefined"):
    result = {
        "pcie_config":
            {"pci_busid": pci_id,
             "max_acc_out": "FAIL",
             "advanced_pci_settings": "FAIL"}
    }

    for line in results:
        if "MAX_ACC_OUT_READ" in line and "0" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"
        elif "MAX_ACC_OUT_READ" in line and "44" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"
        elif "MAX_ACC_OUT_READ" in line and "128" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"

        if "ADVANCED_PCI_SETTINGS" in line and "True" in line:
            result["pcie_config"]["advanced_pci_settings"] = "PASS"
    return result

# Main function to call run_max_acc_check and parse the results
def main(argv=None):
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_max_acc_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()

