#!/usr/bin/env python3
"""
Network Interface RX Discards Health Check Script

This script monitors network interface health by checking for receive (RX) packet 
discards on both standalone servers and Oracle Container Engine (OKE) nodes. 
High discard counts can indicate network congestion, buffer overflows, or hardware issues.

The script supports two deployment modes:
- Standalone servers: Uses standard Ethernet interfaces (enp*)
- OKE nodes: Uses RDMA-capable interfaces (rdma*)
"""

import sys
import argparse
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


# Define and display command-line arguments for the script.
def parse_arguments(parser):
    # Add boolean flag for OKE node identification
    parser.add_argument('-o', '--is-oke-node',
                        action='store_true',  # Sets to True when present, False when absent
                        dest='oke_node_flag',  # Variable name in parsed args
                        help='Is this a OKE node')

    # Display usage information
    print("""
        OPTIONAL ARGUMENTS:
            -o, --is-oke-node      Specify if this node is part of OKE infrastructure
    """)


# Execute RX discards health check across all relevant network interfaces.
def run_rx_discards_check(is_oke_node_flag=None):
    # Configuration defining interface lists and thresholds
    config = {
        "interfaces": {
            # Standard Ethernet interfaces for standalone servers
            "standalone": [
                "enp12s0f0",  # PCIe slot 12, function 0, port 0
                "enp12s0f1",  # PCIe slot 12, function 0, port 1
                "enp42s0f0",  # PCIe slot 42, function 0, port 0
                "enp42s0f1",  # PCIe slot 42, function 0, port 1
                "enp65s0f0",  # PCIe slot 65, function 0, port 0
                "enp65s0f1",  # PCIe slot 65, function 0, port 1
                "enp88s0f0",  # PCIe slot 88, function 0, port 0
                "enp88s0f1",  # PCIe slot 88, function 0, port 1
                "enp134s0f0",  # PCIe slot 134, function 0, port 0
                "enp134s0f1",  # PCIe slot 134, function 0, port 1
                "enp165s0f0",  # PCIe slot 165, function 0, port 0
                "enp165s0f1",  # PCIe slot 165, function 0, port 1
                "enp189s0f0",  # PCIe slot 189, function 0, port 0
                "enp189s0f1",  # PCIe slot 189, function 0, port 1
                "enp213s0f0",  # PCIe slot 213, function 0, port 0
                "enp213s0f1"  # PCIe slot 213, function 0, port 1
            ],
            # RDMA interfaces for Oracle Container Engine (OKE) nodes
            "oke_node": [
                "rdma0", "rdma1", "rdma2", "rdma3",  # RDMA interfaces 0-3
                "rdma4", "rdma5", "rdma6", "rdma7",  # RDMA interfaces 4-7
                "rdma8", "rdma9", "rdma10", "rdma11",  # RDMA interfaces 8-11
                "rdma12", "rdma13", "rdma14", "rdma15"  # RDMA interfaces 12-15
            ]
        },
        # Threshold for considering RX discards problematic
        # Values above this indicate potential network issues
        "rx_discards_check_threshold": 100
    }

    # Select interface list based on node type
    interfaces_list = config["interfaces"]["oke_node"] if is_oke_node_flag else config["interfaces"]["standalone"]
    rx_discards_check_threshold = config["rx_discards_check_threshold"]

    # Process each interface and collect results
    rx_discards_results = []
    for interface in interfaces_list:
        cmd = f"sudo ethtool -S {interface} | grep rx_prio.*_discards"

        # Execute the command and get raw output
        raw_result = run_cmd(cmd)

        # Parse the results and determine pass/fail status
        result = parse_rx_discards_results(interface, raw_result, rx_discards_check_threshold, True)
        rx_discards_results.append(result)

    return rx_discards_results


# Parse ethtool output to determine if RX discards exceed acceptable thresholds.
def parse_rx_discards_results(interface="undefined", results=None,
                              rx_discards_check_threshold=-1, in_cluster_flag=False):
    # TODO: bhrajan auto discover in_cluster_flag
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

    # Only perform detailed checking for cluster nodes
    if in_cluster_flag:
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
    parser = argparse.ArgumentParser()
    parse_arguments(parser)
    args = parser.parse_args()
    print("Health check is in progress ...")
    result = run_rx_discards_check(args.oke_node_flag)
    print(json.dumps(result, indent=2))


# Script entry point - only execute main() when script is run directly
if __name__ == "__main__":
    main()
