"""
# This script checks for PCIe errors on a specified host by running the `dmesg` command
# and parsing the output for any PCIe-related error messages. It returns a JSON object
# indicating the status of the PCIe error check, which can be either "PASS" or "FAIL".
# The script is designed to be run in a HPC environment where SSH access to the host is
# available and the user has the necessary permissions to run `dmesg` with sudo.
"""

import shlex
import subprocess
import json
import re

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

# Function to parse dmesg output and check for PCIe errors
def parse_pcie_error_results(dmesg_result="undefined"):
    result = {
         "pcie_error":
            {"status": "PASS"}
    }
    if len(dmesg_result) == 0:
        result = {
            "pcie_error":
                {"status": "FAIL"}
        }
    for line in dmesg_result:
        if re.search('capabilities', line):
            continue
        else:
            pcie_error = re.search('.*pcieport.*[E|e]rror', line)
            if pcie_error:
                result["pcie_error"]["status"] = "FAIL"
    return result

# Function to loop through dmesg output and parse for PCIe errors
def run_pcie_error_check():
    """ Run PCIe check - checking if each node has PCIe error"""
    cmd = "sudo dmesg"
    dmesg_result = run_cmd(cmd)
    return parse_pcie_error_results(dmesg_result)

# Main function to call run_pcie_error_check and parse the results
def main(argv=None):
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_pcie_error_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()

