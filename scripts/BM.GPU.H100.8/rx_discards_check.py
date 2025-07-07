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


# Execute RX discards health check across all relevant network interfaces.
def run_rx_discards_check():
    # Configuration defining interface lists and thresholds
    config = {
        "interfaces": [
                "enp12s0f0",
                "enp12s0f1",
                "enp42s0f0",
                "enp42s0f1",
                "enp65s0f0",
                "enp65s0f1",
                "enp88s0f0",
                "enp88s0f1",
                "enp134s0f0",
                "enp134s0f1",
                "enp165s0f0",
                "enp165s0f1",
                "enp189s0f0",
                "enp189s0f1",
                "enp213s0f0",
                "enp213s0f1"
        ],
        # Threshold for considering RX discards problematic
        # Values above this indicate potential network issues
        "rx_discards_check_threshold": 100
    }

    # Select interface list based on node type
    interfaces_list = config["interfaces"]
    rx_discards_check_threshold = config["rx_discards_check_threshold"]

    # Process each interface and collect results
    rx_discards_results = []
    for interface in interfaces_list:
        cmd = f"sudo ethtool -S {interface} | grep rx_prio.*_discards"

        # Execute the command and get raw output
        raw_result = run_cmd(cmd)

        # Parse the results and determine pass/fail status
        result = parse_rx_discards_results(interface, raw_result, rx_discards_check_threshold)
        rx_discards_results.append(result)

    return rx_discards_results


# Parse RX discards results for a single interface and determine pass/fail status.
def parse_rx_discards_results(interface="undefined", results=None,
                              rx_discards_check_threshold=-1):
    # Initialize default values
    if results is None:
        results = []

    # Create default result structure with PASS status
    result = {
        "rx_discards": {
            "device": interface,
            "status": "PASS"
        }
    }

    # Check if ethtool command returned any results
    if len(results) == 0:
        # No results means interface doesn't exist or ethtool failed
        result = {
            "rx_discards": {
                "device": interface,
                "status": "FAIL"
            }
        }

    # Ensure device name is set correctly
    result["rx_discards"]["device"] = interface

    # Process each line of ethtool output
    for line in results:
        if len(line) > 0:
            # Parse ethtool output format: "stat_name: value"
            # Remove spaces and split on colon to extract the numeric value
            discards = line.replace(' ', '').split(':')[1]

            # Validate that the discard count is a valid integer
            if isint(discards):
                # Check if discard count exceeds threshold
                if int(discards) > rx_discards_check_threshold:
                    result["rx_discards"]["status"] = "FAIL"
                    break  # Exit early on first failure
            else:
                # Invalid discard value indicates parsing error or interface issue
                result["rx_discards"]["status"] = "FAIL"
                break  # Exit early on parsing failure

    return result


# Main entry point for the RX discards health check script.
def main():
    print("Health check is in progress ...")
    result = run_rx_discards_check()
    print(json.dumps(result, indent=2))


# Script entry point - only execute main() when script is run directly
if __name__ == "__main__":
    main()
