"""
RDMA NIC Count Check Script

This script verifies that the correct number of RDMA (Remote Direct Memory Access) 
Network Interface Cards are present on an H100 GPU system. It checks specific PCI 
bus locations for Mellanox RDMA controllers and validates the count against expected values.

RDMA enables high-speed, low-latency network communication between GPUs in distributed 
computing environments, which is critical for multi-node AI/ML workloads.
"""

import subprocess
import shlex
import json


def run_cmd(cmd):
    cmd_split = shlex.split(cmd)
    try:
        results = subprocess.run(cmd_split, shell=False, stdout=subprocess.PIPE,
                                 stderr=subprocess.STDOUT, check=True, encoding='utf8')
        output = results.stdout.splitlines()
    except subprocess.CalledProcessError as e_process_error:
        return [f"Error: {cmd} {e_process_error.returncode} {e_process_error.output}"]
    return output


# Main function to check RDMA NIC count on the system.
def run_rdma_nic_count_check():
    # Define the path to lspci command (Linux PCI utilities)
    lspci_bin = "/usr/sbin/lspci"

    # Define expected RDMA PCI device locations for H100 GPU systems
    # These are the specific PCI bus addresses where RDMA NICs should be found
    # Each H100 system typically has 16 RDMA NICs (8 pairs) for high-bandwidth networking
    config = {
        "rdma_pci_ids": [
            "0000:0c:00.0",  # RDMA NIC 1A
            "0000:0c:00.1",  # RDMA NIC 1B
            "0000:2a:00.0",  # RDMA NIC 2A
            "0000:2a:00.1",  # RDMA NIC 2B
            "0000:41:00.0",  # RDMA NIC 3A
            "0000:41:00.1",  # RDMA NIC 3B
            "0000:58:00.0",  # RDMA NIC 4A
            "0000:58:00.1",  # RDMA NIC 4B
            "0000:86:00.0",  # RDMA NIC 5A
            "0000:86:00.1",  # RDMA NIC 5B
            "0000:a5:00.0",  # RDMA NIC 6A
            "0000:a5:00.1",  # RDMA NIC 6B
            "0000:bd:00.0",  # RDMA NIC 7A
            "0000:bd:00.1",  # RDMA NIC 7B
            "0000:d5:00.0",  # RDMA NIC 8A
            "0000:d5:00.1"   # RDMA NIC 8B
        ]
    }

    # Extract the list of PCI IDs from the configuration
    rdma_pci_ids_list = config["rdma_pci_ids"]

    # Initialize list to store results from each PCI device query
    list_of_results = []

    # Check each expected RDMA device location
    for device in rdma_pci_ids_list:
        cmd = f'sudo {lspci_bin} -v -s {device}'

        # Execute the command and store the output
        raw_result = run_cmd(cmd)
        list_of_results.append(raw_result)

    # TODO: bhrajan - How to determine host is not in cluster
    result = parse_rdma_nic_count_results(results=list_of_results,
                                          expected_nics_count=len(rdma_pci_ids_list))
    return result


# Parse all the results and determine pass/fail status
def parse_rdma_nic_count_results(results, expected_nics_count, is_host_not_in_cluster=False):
    # Handle special case where host doesn't support RDMA
    if is_host_not_in_cluster:
        return {"rdma_nic_count": {"status": "PASS - This host doesn't support RDMA link."}}

    # Initialize counter for detected RDMA NICs
    num_rdma_nics = 0

    # Search through all lspci output results
    for line in results:  # For each PCI device queried
        for sub_line in line:  # For each line of that device's lspci output
            # Look for the Mellanox manufacturer identifier
            # Mellanox Technologies makes the RDMA controllers used in H100 systems
            if "controller: Mellanox Technologies" in sub_line:
                num_rdma_nics += 1

    # Compare actual count with expected count and determine result
    if num_rdma_nics == expected_nics_count:
        result = {"rdma_nic_count": {"status": "PASS", "num_rdma_nics": num_rdma_nics}}
    else:
        result = {"rdma_nic_count": {"status": "FAIL", "num_rdma_nics": num_rdma_nics}}

    return result


# Main entry point for the RDMA NIC count check script.
def main(argv=None):
    print("Health check is in progress ...")
    result = run_rdma_nic_count_check()
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
