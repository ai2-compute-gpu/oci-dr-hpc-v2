"""
# This script checks if nvidia-fabricmanager service is enabled and running on the system.
# nvidia-fabricmanager is required for optimal GPU performance on systems with multiple GPUs
# and high-speed interconnects like NVLink and NVSwitch.
"""

import subprocess
import shlex
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

# Function to check nvidia-fabricmanager service status
def check_fabricmanager_service():
    """Check if nvidia-fabricmanager service is active and running."""
    cmd = 'systemctl status nvidia-fabricmanager'
    
    result = run_cmd(cmd)
    if not result:
        return False, "No output from systemctl status command"
    
    # Check for active (running) status
    for line in result:
        if "active (running)" in line.lower():
            return True, "nvidia-fabricmanager service is active and running"
    
    # If not active, check what the actual status is
    status_info = []
    for line in result:
        line = line.strip()
        if "Active:" in line or "Main PID:" in line or "Status:" in line:
            status_info.append(line)
    
    if status_info:
        return False, "; ".join(status_info)
    else:
        return False, "nvidia-fabricmanager service is not active"

# Function to run fabricmanager check
def run_fabricmanager_check():
    """Main function to check nvidia-fabricmanager service status."""
    is_running, details = check_fabricmanager_service()
    
    result = {
        "gpu": {
            "fabricmanager-service": "PASS" if is_running else "FAIL"
        }
    }
    
    return result, details

# Function to parse fabricmanager results
def parse_fabricmanager_results(raw_result=None, details=""):
    """Parse the fabricmanager check results and return in expected format."""
    result = {
        "gpu": {
            "fabricmanager-service": "FAIL"
        }
    }

    if raw_result and "gpu" in raw_result:
        if raw_result["gpu"].get("fabricmanager-service") == "PASS":
            result["gpu"]["fabricmanager-service"] = "PASS"

    return result

# Main function to call run_fabricmanager_check and parse the results  
def main(argv=None):
    print("Health check is in progress ...")
    result, details = run_fabricmanager_check()
    parsed_result = parse_fabricmanager_results(result, details)
    print(json.dumps(parsed_result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()