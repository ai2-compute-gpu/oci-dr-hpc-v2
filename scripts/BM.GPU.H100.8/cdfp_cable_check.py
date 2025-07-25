"""
# This script checks if CDFP cables are correctly cabled by validating GPU PCI addresses 
# and module IDs against expected configurations. It queries nvidia-smi for both PCI bus IDs
# and module IDs, then compares them against predefined mappings to ensure proper cabling.
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

# Function to normalize PCI addresses
def normalize_pci_address(pci_addr):
    """Normalize PCI address format."""
    normalized = pci_addr.lower()
    # Handle cases where PCI address starts with "000000"
    if pci_addr.startswith("000000"):
        normalized = "00" + pci_addr[6:].lower()
    return normalized

# Function to get GPU PCI addresses and module IDs
def get_gpu_info():
    """Get GPU PCI addresses and corresponding GPU module IDs using nvidia-smi -q."""
    # Use nvidia-smi -q to get detailed GPU information
    query_cmd = 'nvidia-smi -q'
    
    query_result = run_cmd(query_cmd)
    if not query_result or any("Error:" in line for line in query_result):
        return [], []
    
    pci_addresses = []
    module_ids = []
    
    # First pass: collect all PCI addresses in order
    for line in query_result:
        line = line.strip()
        
        # Look for Bus ID lines
        if "Bus Id" in line and ":" in line:
            # Extract Bus ID from lines like "    Bus Id                        : 00000000:0F:00.0"
            try:
                # Split on colons and take the last 3 parts to reconstruct the PCI address
                parts = line.split(":")
                if len(parts) >= 4:
                    # Strip whitespace from each part before reconstructing
                    bus_id = parts[-3].strip() + ":" + parts[-2].strip() + ":" + parts[-1].strip()
                    if bus_id and bus_id != "::" and bus_id != "":
                        # Normalize the PCI address
                        pci_addresses.append(normalize_pci_address(bus_id))
            except:
                continue
    
    # Second pass: collect all module IDs using the exact working command
    # nvidia-smi -q | grep -i "Module ID" | awk '{print $4}'
    module_cmd = 'bash -c "nvidia-smi -q | grep -i \\"Module ID\\" | awk \'{print $4}\'"'
    module_result = run_cmd(module_cmd)
    
    for line in module_result:
        line = line.strip()
        if line and not line.startswith("Error"):
            module_ids.append(line)
    
    # If no module IDs found, use sequential numbering
    if len(module_ids) == 0:
        print("No module IDs found, using sequential numbering")
        for i in range(len(pci_addresses)):
            module_ids.append(str(i + 1))
    
    return pci_addresses, module_ids

# Function to run CDFP cable check
def run_cdfp_cable_check():
    # Define expected PCI Bus IDs and Module IDs for H100
    config = {
        "gpu_pci_ids": [
            "00000000:0f:00.0",
            "00000000:2d:00.0", 
            "00000000:44:00.0",
            "00000000:5b:00.0",
            "00000000:89:00.0",
            "00000000:a8:00.0",
            "00000000:c0:00.0",
            "00000000:d8:00.0"
        ],
        "gpu_module_ids": [
            "2", "4", "3", "1",
            "7", "5", "8", "6"
        ]
    }
    
    expected_pci_ids = config["gpu_pci_ids"]
    expected_gpu_module_ids = config["gpu_module_ids"]
    
    # Get actual GPU information
    pci_result, module_id_result = get_gpu_info()
    
    # Parse the CDFP results
    result = parse_cdfp_results(pci_result, module_id_result, expected_pci_ids, expected_gpu_module_ids)
    return result

# Function to parse CDFP cable results
def parse_cdfp_results(pci_result=None, module_id_result=None, pci_expected=None, module_id_expected=None):
    result = {
        "gpu": {
            "cdfp": "PASS"
        }
    }

    if not pci_result or not module_id_result or len(pci_result) == 0 or len(module_id_result) == 0:
        result["gpu"]["cdfp"] = "FAIL - Missing input data"
        return result

    # Normalize PCI results
    pci_result_list = []
    for pci in pci_result:
        normalized = normalize_pci_address(pci)
        pci_result_list.append(normalized)

    # Create dictionaries for mapping
    expected_mapping = dict(zip([normalize_pci_address(pci) for pci in pci_expected], module_id_expected))
    actual_mapping = dict(zip(pci_result_list, module_id_result))

    # Validate each expected PCI and module ID pair
    fail_list = []
    for expected_pci, expected_module_id in expected_mapping.items():
        actual_module_id = actual_mapping.get(expected_pci)
        
        if actual_module_id is None:
            fail_list.append(f"Expected GPU with PCI Address {expected_pci} not found")
        elif actual_module_id != expected_module_id:
            fail_list.append(f"Mismatch for PCI {expected_pci}: Expected GPU module ID {expected_module_id}, found {actual_module_id}")

    if fail_list:
        result["gpu"]["cdfp"] = "FAIL - " + ", ".join(fail_list)

    return result

# Main function to call run_cdfp_cable_check and parse the results  
def main(argv=None):
    print("Health check is in progress ...")
    result = run_cdfp_cable_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()