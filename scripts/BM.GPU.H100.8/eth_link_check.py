"""
# This script checks the state of each 100GbE RoCE NIC (non-RDMA Ethernet interfaces).
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

# Helper functions
def isfloat(val):
    try:
        float(val)
        return True
    except Exception:
        return False

def isint(val):
    try:
        int(val)
        return True
    except Exception:
        return False

def parse_raw_physical_errors_per_lane(val):
    if isinstance(val, list):
        return [x for x in val if x != "undefined"]
    try:
        return [int(x) for x in val if x != "undefined"]
    except Exception:
        return []

# Function to parse mlxlink output and emit JSON for Ethernet interfaces
def parse_eth_link_results(interface, results, expected_speed, expected_width,
                          raw_physical_errors_per_lane_threshold,
                          effective_physical_errors_threshold,
                          raw_physical_ber_threshold,
                          effective_physical_ber_threshold):
    
    expected_state = "Active"
    expected_phys_state = ["LinkUp", "ETH_AN_FSM_ENABLE"]

    try:
        output = json.loads(''.join(results))["result"]["output"]
    except json.JSONDecodeError:
        return {
            "device": interface,
            "eth_link_speed": f"FAIL - Unable to parse mlxlink output",
            "eth_link_state": f"FAIL - Unable to parse mlxlink output",
            "physical_state": f"FAIL - Unable to parse mlxlink output",
            "eth_link_width": f"FAIL - Unable to parse mlxlink output",
            "eth_link_status": f"FAIL - Unable to parse mlxlink output",
            "effective_physical_errors": "FAIL - Unable to parse mlxlink output",
            "effective_physical_ber": f"FAIL - Unable to parse mlxlink output",
            "raw_physical_errors_per_lane": "FAIL - Unable to parse mlxlink output",
            "raw_physical_ber": f"FAIL - Unable to parse mlxlink output"
        }
    
    speed = output["Operational Info"]["Speed"]
    state = output["Operational Info"]["State"]
    phys_state = output["Operational Info"]["Physical state"]
    width = output["Operational Info"]["Width"]
    status_opcode = output["Troubleshooting Info"]["Status Opcode"]
    recommendation = output["Troubleshooting Info"]["Recommendation"]
    effective_physical_errors = output["Physical Counters and BER Info"]["Effective Physical Errors"]
    effective_physical_ber = output["Physical Counters and BER Info"]["Effective Physical BER"]
    raw_physical_errors_per_lane = parse_raw_physical_errors_per_lane(
        output["Physical Counters and BER Info"]["Raw Physical Errors Per Lane"])
    raw_physical_ber = output["Physical Counters and BER Info"]["Raw Physical BER"]

    result = {
        "device": interface,
        "eth_link_speed": f"FAIL - {speed}, expected {expected_speed}",
        "eth_link_state": f"FAIL - {state}, expected {expected_state}",
        "physical_state": f"FAIL - {phys_state}, expected {expected_phys_state}",
        "eth_link_width": f"FAIL - {width}, expected {expected_width}",
        "eth_link_status": f"FAIL - {recommendation}",
        "effective_physical_errors": "PASS",
        "effective_physical_ber": f"FAIL - {effective_physical_ber}",
        "raw_physical_errors_per_lane": "PASS",
        "raw_physical_ber": f"FAIL - {raw_physical_ber}"
    }
    
    # Set PASS conditions
    if expected_speed in speed:
        result["eth_link_speed"] = "PASS"
    if state == expected_state:
        result["eth_link_state"] = "PASS"
    if phys_state in expected_phys_state:
        result["physical_state"] = "PASS"
    if width == expected_width:
        result["eth_link_width"] = "PASS"
    if status_opcode == "0":
        result["eth_link_status"] = "PASS"
    if isfloat(effective_physical_ber):
        if float(effective_physical_ber) < effective_physical_ber_threshold:
            result["effective_physical_ber"] = "PASS"
    if isfloat(raw_physical_ber):
        if float(raw_physical_ber) < raw_physical_ber_threshold:
            result["raw_physical_ber"] = "PASS"
    if isint(effective_physical_errors):
        if int(effective_physical_errors) > effective_physical_errors_threshold:
            result["effective_physical_errors"] = f"FAIL - {effective_physical_errors}"
    
    try:
        for lane_error in raw_physical_errors_per_lane:
            if lane_error == "undefined":
                continue
            lane_error = int(lane_error)
            if lane_error > raw_physical_errors_per_lane_threshold:
                errors_per_lane_summary = ' '.join([str(err) for err in raw_physical_errors_per_lane])
                result["raw_physical_errors_per_lane"] = f"WARN - {errors_per_lane_summary}"
                break
    except:
        result["raw_physical_errors_per_lane"] = "WARN - Unknown"

    return result

# Main function to check all Ethernet interfaces
def run_eth_link_check():
    # Configuration for BM.GPU.H100.8
    config = {
        "eth_link_check": {
            "speed": "100G",
            "width": "4x", 
            "effective_physical_errors": 0,
            "raw_physical_errors_per_lane": 10000,
            "effective_physical_ber": 1E-12,
            "raw_physical_ber": 1E-5
        }
    }

    expected_speed = config["eth_link_check"]["speed"]
    expected_width = config["eth_link_check"]["width"]
    effective_physical_errors_threshold = config["eth_link_check"]["effective_physical_errors"]
    raw_physical_errors_per_lane_threshold = config["eth_link_check"]["raw_physical_errors_per_lane"]
    effective_physical_ber_threshold = config["eth_link_check"]["effective_physical_ber"]
    raw_physical_ber_threshold = config["eth_link_check"]["raw_physical_ber"]

    # Find binaries in the OS
    ibdev2netdev_bin = shutil.which("ibdev2netdev")
    mlxlink_bin = shutil.which("mlxlink")
    mst_bin = shutil.which("mst")
    
    if not ibdev2netdev_bin or not mlxlink_bin or not mst_bin:
        raise FileNotFoundError("Required binaries 'ibdev2netdev', 'mlxlink', or 'mst' not found in PATH.")

    # Hard-coded VCN device names for BM.GPU.H100.8 (non-RDMA Ethernet interfaces)
    # Based on shapes.json configuration for H100
    vcn_devices = ["mlx5_2", "mlx5_11"]
    
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
    
    # Check only the VCN devices specified for H100
    for device in vcn_devices:
        if device in device_dict:
            interface = device_dict[device]
            print(f"Checking VCN device {device} (interface {interface})...")
            
            cmd = f'sudo {mlxlink_bin} -d {device} --json --show_module --show_counters --show_eye'
            raw_result = run_cmd(cmd)
            
            result = parse_eth_link_results(
                interface, raw_result, expected_speed, expected_width,
                raw_physical_errors_per_lane_threshold,
                effective_physical_errors_threshold,
                raw_physical_ber_threshold,
                effective_physical_ber_threshold
            )
            interface_results.append(result)
        else:
            print(f"Warning: VCN device {device} not found in device mapping")

    return interface_results

# Main function to call run_eth_link_check and print results
def main(argv=None):
    print("Ethernet link health check is in progress ...")
    result = run_eth_link_check()
    print(json.dumps({"eth_link": result}, indent=2))

# Run the main function
if __name__ == "__main__":
    main()