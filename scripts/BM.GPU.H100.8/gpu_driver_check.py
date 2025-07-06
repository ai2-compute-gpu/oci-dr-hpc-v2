"""
Validate that the GPU driver version is supported and is not blacklisted.
For H100:
Blacklisted versions: "470.57.02"
Supported versions: "450.119.03", "450.142.0", "470.103.01", "470.129.06", "470.141.03", "510.47.03", "535.104.12", "550.90.12"
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

# Function to call nvidia-smi to determine GPU driver version
def run_gpu_driver_check():
    config = {
        "gpu_driver_check": {
            "state": "enable",
            "blacklisted_versions": ["470.57.02"],
            "supported_versions": ["450.119.03", "450.142.0",
                                   "470.103.01", "470.129.06", "470.141.03",
                                   "510.47.03", "535.104.12", "550.90.12"]
        }
    }
    bad_driver_list = config["gpu_driver_check"]["blacklisted_versions"]
    supported_driver_list = config["gpu_driver_check"]["supported_versions"]

    cmd = f'nvidia-smi --query-gpu=driver_version --format=csv,noheader'
    output = run_cmd(cmd)
    result = parse_gpu_driver_results(output, bad_driver_list, supported_driver_list)
    return result

# Function to parse nvidia-smi output and emit JSON
def parse_gpu_driver_results(results="undefined", driver_blacklist="undefined",
                             supported_driver="undefined"):
    result = {
        "gpu":
            {"driver_version": "FAIL"}
    }
    results = [item for item in results if item != '']
    if len(results) == 0:
        return result
    current_version = results[0]
    all_same_version = results.count(current_version) == len(results)
    # This should never happen, but let's double check all driver versions are the same
    if not all_same_version:
        result["gpu"]["driver_version"] = "FAIL - Driver versions are mismatched"
    else:
        if current_version not in driver_blacklist:
            if current_version in supported_driver:
                result["gpu"]["driver_version"] = "PASS"
            else:
                result["gpu"]["driver_version"] = "WARN - unsupported driver"
    return result

# Main function to call run_max_acc_check and parse the results
def main(argv=None):
    result = run_gpu_driver_check()
    print(json.dumps(result, indent=2))


# Run the main function
if __name__ == "__main__":
    main()
