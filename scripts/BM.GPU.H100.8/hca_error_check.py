"""
# This script checks for HCA (Host Channel Adapter) errors on a specified host by running 
# the `sudo dmesg -T | grep mlx5 | grep 'Fatal'` command and parsing the output for any 
# fatal MLX5-related error messages. It returns a JSON object indicating the status of 
# the HCA error check, which can be either "PASS" or "FAIL".
# The script is designed to be run in a HPC environment where SSH access to the host is
# available and the user has the necessary permissions to run `dmesg` with sudo.
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

# Function to parse dmesg output and check for HCA errors
def parse_hca_error_results(dmesg_result="undefined"):
    result = {
         "hca_error":
            {"status": "PASS"}
    }
    
    # If command returned any output, that means fatal errors were found
    if len(dmesg_result) > 0:
        result["hca_error"]["status"] = "FAIL"
    
    return result

# Function to run HCA error check
def run_hca_error_check():
    """ Run HCA check - checking if each node has MLX5 fatal errors"""
    cmd = "sudo dmesg -T | grep mlx5 | grep 'Fatal'"
    dmesg_result = run_cmd(cmd)
    return parse_hca_error_results(dmesg_result)

# Main function to call run_hca_error_check and parse the results
def main(argv=None):
    print("HCA error check is in progress and the result will be provided within 1 minute.")
    result = run_hca_error_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()