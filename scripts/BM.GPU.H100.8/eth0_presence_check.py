"""
# This script checks if the eth0 network interface is present on the system by running
# the `ip addr | grep eth0` command and parsing the output. It returns a JSON object
# indicating the status of the eth0 presence check, which can be either "PASS" or "FAIL".
# The script is designed to be run in a HPC environment where the user has the necessary
# permissions to run network interface commands.
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

# Function to parse ip addr output and check for eth0 presence
def parse_eth0_presence_results(ip_result="undefined"):
    result = {
         "eth0_presence":
            {"status": "FAIL"}
    }
    
    if len(ip_result) == 0:
        result = {
            "eth0_presence":
                {"status": "FAIL"}
        }
        return result
    
    for line in ip_result:
        if "eth0" in line:
            result["eth0_presence"]["status"] = "PASS"
            break
    
    return result

# Function to run eth0 presence check
def run_eth0_presence_check():
    """ Run eth0 presence check - checking if eth0 interface exists"""
    cmd = "ip addr | grep eth0"
    ip_result = run_cmd(cmd)
    return parse_eth0_presence_results(ip_result)

# Main function to call run_eth0_presence_check and parse the results
def main(argv=None):
    print("Eth0 presence check is in progress and the result will be provided within 1 minute.")
    result = run_eth0_presence_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()