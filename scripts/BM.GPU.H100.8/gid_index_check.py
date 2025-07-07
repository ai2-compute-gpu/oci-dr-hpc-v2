""""
# This script checks the GID index on a system to ensure it matches expected values.
# It is designed to run on systems with multiple GPUs, such as H100 GPU systems,
# and provides a health check for GPU configuration consistency.
# The GID index is crucial for RDMA (Remote Direct Memory Access) operations,
# and discrepancies can lead to performance issues."""

import subprocess
import shlex
import json


# Execute a system command safely with automatic shell detection.
def run_cmd(cmd=None):
    # Split command into arguments for security (prevents injection)
    cmd_split = shlex.split(cmd)

    try:
        # Auto-detect if shell features are needed
        if '|' in cmd_split or '&&' in cmd_split:
            # Use shell mode for pipes and logical operators
            results = subprocess.run(cmd, shell=True, stdout=subprocess.PIPE,
                                     stderr=subprocess.STDOUT, check=True, encoding='utf8')
        else:
            # Use safer non-shell mode for simple commands
            results = subprocess.run(cmd_split, shell=False, stdout=subprocess.PIPE,
                                     stderr=subprocess.STDOUT, check=True, encoding='utf8')

        # Split output into individual lines for easier processing
        output = results.stdout.splitlines()
    except subprocess.CalledProcessError as e_process_error:
        # Return error information if command fails
        return [f"Error: {cmd} {e_process_error.returncode} {e_process_error.output}"]

    return output


# Check if a string represents a valid integer.
def isint(num):
    try:
        int(num)
        return True
    except ValueError:
        return False


# Run GID index check - some interfaces report a GID index of 4 instead of 0,1,2,3 as expected.
def run_gid_index_check():
    """ This change in GID indices can have an impact on workloads that expect the default values."""

    config = {
        "gid_index_check": [0, 1, 2, 3]
    }
    cmd = f'sudo show_gids | tail -n +3 | head -n -1'
    gid_index_result = run_cmd(cmd)
    gid_index_check_threshold = config["gid_index_check"]
    return parse_gid_index_results(gid_index_result, gid_index_check_threshold)


# parse results from gid and checks if all the ids confirms to given threshold
def parse_gid_index_results(gid_index_result=None, gid_index_check_threshold=[0, 1, 2, 3]):
    result = {
        "gid_index":
            {"status": "PASS"}
    }
    if len(gid_index_result) == 0:
        result = {
            "gid_index":
                {"status": "FAIL"}
        }
    for line in gid_index_result:
        gid_index = line.split()[2]
        if isint(gid_index):
            if int(gid_index) not in gid_index_check_threshold:
                result["gid_index"]["status"] = "FAIL"
    return result


# Main entry point for the RX discards health check script.
def main():
    print("Health check is in progress ...")
    result = run_gid_index_check()
    print(json.dumps(result, indent=2))


# Script entry point - only execute main() when script is run directly
if __name__ == "__main__":
    main()
