"""
# This script checks the authentication status of RDMA interfaces using wpa_cli.
"""

import subprocess
import shlex
import json
import shutil
import time

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

# Function to parse authentication status for RDMA interfaces
def parse_auth_results(interface, wpa_cli_output):
    
    result = {
        "device": interface,
        "auth_status": "FAIL - Unable to check authentication"
    }
    
    # Check if wpa_cli command succeeded and has output
    if not wpa_cli_output or any("Error:" in line for line in wpa_cli_output):
        result["auth_status"] = "FAIL - Unable to run wpa_cli command"
        return result
    
    # Look for specific authenticated status in the output
    for line in wpa_cli_output:
        if "Supplicant PAE state=AUTHENTICATED" in line:
            # If we find the authenticated state, consider it authenticated
            result["auth_status"] = "PASS"
            return result
    
    # If no authenticated status found, authentication failed
    result["auth_status"] = "FAIL - Interface not authenticated"
    return result

# Main function to check RDMA interface authentication
def run_auth_check():
    # Find required binaries
    ibdev2netdev_bin = shutil.which("ibdev2netdev")
    wpa_cli_bin = shutil.which("wpa_cli")
    
    if not ibdev2netdev_bin:
        raise FileNotFoundError("Required binary 'ibdev2netdev' not found in PATH.")
    
    if not wpa_cli_bin:
        raise FileNotFoundError("Required binary 'wpa_cli' not found in PATH.")

    # RDMA device names for BM.GPU.H100.8 (based on shapes.json)
    # These are the ConnectX-7 devices used for RDMA, excluding VCN devices
    rdma_devices = [
        "mlx5_0", "mlx5_1", "mlx5_3", "mlx5_4", "mlx5_5", "mlx5_6", 
        "mlx5_7", "mlx5_8", "mlx5_9", "mlx5_10", "mlx5_12", "mlx5_13", 
        "mlx5_14", "mlx5_15", "mlx5_16", "mlx5_17"
    ]
    
    # Get device to interface mapping from ibdev2netdev
    device_dict = {}
    cmd = f'sudo {ibdev2netdev_bin}'
    raw_result = run_cmd(cmd)
    for line in raw_result:
        if len(line.split()) >= 5:
            device = line.split()[0]
            interface = line.split()[4]
            device_dict[device] = interface

    interface_results = []
    
    # Check only the RDMA devices specified for H100
    for device in rdma_devices:
        if device in device_dict:
            interface = device_dict[device]
            print(f"Checking RDMA device {device} (interface {interface})...")
            
            # Run wpa_cli status command for this interface
            cmd = f'sudo {wpa_cli_bin} -i {interface} status'
            raw_result = run_cmd(cmd)
            
            # Use the full output for authentication check
            filtered_result = raw_result
            
            result = parse_auth_results(interface, filtered_result)
            interface_results.append(result)
        else:
            print(f"Warning: RDMA device {device} not found in device mapping")

    return interface_results

# Main function to call run_auth_check and print results
def main(argv=None):
    print("RDMA interface authentication check is in progress ...")
    result = run_auth_check()
    print(json.dumps({"auth_check": result}, indent=2))

# Run the main function
if __name__ == "__main__":
    main()