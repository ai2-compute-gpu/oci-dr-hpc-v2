"""
# This script checks the state of each RDMA NIC link and validates link parameters.
"""

import subprocess
import shlex
import json
import shutil

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

# Function to parse mlxlink output and emit JSON
def parse_link_results(interface, raw_result, expected_speed,
                      raw_physical_errors_per_lane_threshold=-1,
                      effective_physical_errors_threshold=-1):
    # Join lines and parse JSON
    results = ''.join(raw_result)
    result = {"link": {"device": interface}}

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
            return val
        try:
            return [int(x) for x in val if x != "undefined"]
        except Exception:
            return []

    # If error, try to extract JSON
    if results.startswith("Error:"):
        index = results.find("{")
        if index != -1:
            results = results[index:]
        else:
            result["link"]["status"] = "FAIL - Invalid interface: {}".format(interface)
            return result

    if not results.strip():
        result["link"]["status"] = "FAIL - Invalid interface: {}".format(interface)
        return result

    try:
        output = json.loads(results)["result"]["output"]
    except Exception:
        result["link"]["status"] = "FAIL - Unable to parse mlxlink output"
        return result

    # Expected values
    expected_state = "Active"
    expected_phys_state = ["LinkUp", "ETH_AN_FSM_ENABLE"]

    # Extract fields
    speed = output["Operational Info"].get("Speed", "")
    state = output["Operational Info"].get("State", "")
    phys_state = output["Operational Info"].get("Physical state", "")
    status_opcode = output["Troubleshooting Info"].get("Status Opcode", "")
    recommendation = output["Troubleshooting Info"].get("Recommendation", "")
    effective_physical_errors = output["Physical Counters and BER Info"].get("Effective Physical Errors", "")
    effective_physical_ber = output["Physical Counters and BER Info"].get("Effective Physical BER", "")
    raw_physical_errors_per_lane = parse_raw_physical_errors_per_lane(
        output["Physical Counters and BER Info"].get("Raw Physical Errors Per Lane", []))
    raw_physical_ber = output["Physical Counters and BER Info"].get("Raw Physical BER", "")

    # Set initial FAILs
    result["link"]["link_speed"] = f"FAIL - {speed}, expected {expected_speed}"
    result["link"]["link_state"] = f"FAIL - {state}, expected {expected_state}"
    result["link"]["physical_state"] = f"FAIL - {phys_state}, expected {expected_phys_state}"
    result["link"]["link_status"] = f"FAIL - {recommendation}"
    result["link"]["effective_physical_errors"] = "PASS"
    result["link"]["effective_physical_ber"] = f"FAIL - {effective_physical_ber}"
    result["link"]["raw_physical_errors_per_lane"] = "PASS"
    result["link"]["raw_physical_ber"] = f"FAIL - {raw_physical_ber}"

    # Set PASS if matches
    if expected_speed in speed:
        result["link"]["link_speed"] = "PASS"
    if state == expected_state:
        result["link"]["link_state"] = "PASS"
    if phys_state in expected_phys_state:
        result["link"]["physical_state"] = "PASS"
    if status_opcode == "0":
        result["link"]["link_status"] = "PASS"
    if isfloat(effective_physical_ber) and float(effective_physical_ber) < 1E-12:
        result["link"]["effective_physical_ber"] = "PASS"
    if isfloat(raw_physical_ber) and float(raw_physical_ber) < 1E-5:
        result["link"]["raw_physical_ber"] = "PASS"
    if isint(effective_physical_errors) and int(effective_physical_errors) > effective_physical_errors_threshold:
        result["link"]["effective_physical_errors"] = f"FAIL - {effective_physical_errors}"
    try:
        for lane_error in raw_physical_errors_per_lane:
            if lane_error == "undefined":
                continue
            lane_error = int(lane_error)
            if lane_error > raw_physical_errors_per_lane_threshold:
                errors_per_lane_summary = ' '.join([str(err) for err in raw_physical_errors_per_lane])
                result["link"]["raw_physical_errors_per_lane"] = f"WARN - {errors_per_lane_summary}"
                break
    except Exception:
        result["link"]["raw_physical_errors_per_lane"] = "WARN - Unknown"

    return result

# Main function to check all interfaces
def run_link_check():
    # Expected RDMA mlx device names for this shape
    expected_mlx_devices = [
        "mlx5_0", "mlx5_1", "mlx5_3", "mlx5_4", "mlx5_5", "mlx5_6", "mlx5_7", "mlx5_8",
        "mlx5_9", "mlx5_10", "mlx5_12", "mlx5_13", "mlx5_14", "mlx5_15", "mlx5_16", "mlx5_17"
    ]
    
    config = {
        "link_check": {
            "speed": "200G",
            "effective_physical_errors": 0,
            "raw_physical_errors_per_lane": 10000
        }
    }

    expected_speed = config["link_check"]["speed"]
    expected_effective_physical_errors = config["link_check"]["effective_physical_errors"]
    expected_raw_physical_errors_per_lane = config["link_check"]["raw_physical_errors_per_lane"]

    # Find binaries in the OS
    ibdev2netdev_bin = shutil.which("ibdev2netdev")
    mlxlink_bin = shutil.which("mlxlink")
    if not ibdev2netdev_bin or not mlxlink_bin:
        raise FileNotFoundError("Required binaries 'ibdev2netdev' or 'mlxlink' not found in PATH.")

    # Get device list and map mlx devices to OS interface names
    cmd = f'sudo {ibdev2netdev_bin}'
    raw_result = run_cmd(cmd)
    device_to_interface_map = {}
    for line in raw_result:
        if len(line.split()) >= 5:
            device = line.split()[0]
            interface = line.split()[4]
            device_to_interface_map[device] = interface

    # Find OS interface names for expected mlx devices and create failure results for missing ones
    interfaces_to_check = []
    interface_results = []
    
    for expected_device in expected_mlx_devices:
        if expected_device in device_to_interface_map:
            os_interface = device_to_interface_map[expected_device]
            interfaces_to_check.append(os_interface)
        else:
            # Create a failure result for the missing device
            failure_result = {
                "device": expected_device,
                "link_speed": f"FAIL - Device {expected_device} not found",
                "link_state": f"FAIL - Device {expected_device} not found",
                "physical_state": f"FAIL - Device {expected_device} not found",
                "link_status": f"FAIL - Device {expected_device} not found",
                "effective_physical_errors": f"FAIL - Device {expected_device} not found",
                "effective_physical_ber": f"FAIL - Device {expected_device} not found",
                "raw_physical_errors_per_lane": f"FAIL - Device {expected_device} not found",
                "raw_physical_ber": f"FAIL - Device {expected_device} not found"
            }
            interface_results.append(failure_result)

    if not interfaces_to_check and not interface_results:
        raise RuntimeError("No expected RDMA devices found on the system")

    for interface in interfaces_to_check:
        # Find the device name for this interface
        device_name = None
        for device, iface in device_to_interface_map.items():
            if iface == interface:
                device_name = device
                break
        
        if device_name:
            cmd = f'sudo {mlxlink_bin} -d {device_name} --json --show_module --show_counters --show_eye'
            raw_result = run_cmd(cmd)
        else:
            raw_result = []

        result = parse_link_results(
            interface,
            raw_result,
            expected_speed,
            expected_raw_physical_errors_per_lane,
            expected_effective_physical_errors
        )
        interface_results.append(result["link"])

    return {"link": interface_results}

# Main function to call run_link_check and print results
def main(argv=None):
    print("Health check is in progress ...")
    result = run_link_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()