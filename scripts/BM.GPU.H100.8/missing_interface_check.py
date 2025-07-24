"""
# This script checks for missing PCIe interfaces on a specified host by running the `lspci` command
# and parsing the output for any devices with revision 'ff' (indicating missing/failed devices).
# It returns a JSON object indicating the status of the missing interface check, which can be
# either "PASS" or "FAIL". The script is designed to be run in a HPC environment where the user
# has the necessary permissions to run `lspci`.
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
        output = results.stdout.strip()
    except subprocess.CalledProcessError as e_process_error:
        return f"Error: {cmd} {e_process_error.returncode} {e_process_error.output}"
    return output

# Function to parse lspci output and check for missing interfaces
def parse_missing_interface_results(lspci_result="0"):
    result = {
         "missing_interface":
            {"status": "PASS", "missing_count": 0}
    }
    
    try:
        missing_count = int(lspci_result.strip())
        result["missing_interface"]["missing_count"] = missing_count
        
        if missing_count > 0:
            result["missing_interface"]["status"] = "FAIL"
            
    except (ValueError, AttributeError):
        # If we can't parse the count or get an error, treat as failure
        result["missing_interface"]["status"] = "FAIL"
        result["missing_interface"]["error"] = "Unable to parse missing interface count"
        
    return result

# Function to run the missing interface check
def run_missing_interface_check():
    """ Run missing interface check - checking for PCIe devices with revision 'ff' """
    cmd = "lspci | grep -i 'rev ff' | wc -l"
    lspci_result = run_cmd(cmd)
    return parse_missing_interface_results(lspci_result)

# Main function to call run_missing_interface_check and parse the results
def main():
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_missing_interface_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()