"""
# This script checks the GPU count on a system by querying the PCI bus IDs of NVIDIA GPUs
# using the `nvidia-smi` command. It compares the detected GPU bus IDs against a predefined list
# of expected PCI bus IDs for H100 GPUs. The output is formatted as JSON, indicating whether the GPU count check passed or failed.          
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

# Function to parse nvidia-smi output and emit JSON
def run_gpu_count_check():
    # Define the path to nvidia-smi
    nvidia_smi_bin = "/usr/bin/nvidia-smi"

    # Define PCI Bus IDs for H100
    config = {
        "pci_bus_ids": [
            "00000000:0F:00.0",
            "00000000:2D:00.0",
            "00000000:44:00.0",
            "00000000:89:00.0",
            "00000000:5B:00.0",
            "00000000:A8:00.0",
            "00000000:C0:00.0",
            "00000000:D8:00.0"
        ]
    }

    known_pci_busid = config["pci_bus_ids"]
    cmd = f'{nvidia_smi_bin} --query-gpu=pci.bus_id --format=csv,noheader'
    raw_result = run_cmd(cmd)
    result = parse_gpu_count_results(raw_result, known_pci_busid)
    return result

# Function to parse GPU count results and check against expected bus IDs
def parse_gpu_count_results(raw_result="undefined", known_gpu_busids="undefined"):
    result = {
        "gpu":
            {"device_count": "FAIL"}
    }
    fail_list = []
    gpu_bus_ids= [busid.lower() for busid in known_gpu_busids]
    expected_gpu_count = len(gpu_bus_ids)
    gpu_count_result = len([item for item in raw_result if item != ''])
    parsed_results = []

    for line in raw_result:
        parsed_results.append(line)
        if len(line) > 0 and line.lower() not in gpu_bus_ids:
            fail_list.append(line)
    if not fail_list and expected_gpu_count == gpu_count_result:
        result["gpu"]["device_count"] = "PASS"
    else:
        if expected_gpu_count == gpu_count_result:
            fail_type = "wrong busid"
            fail_ids = ",".join(iter(fail_list))
        else:
            fail_type = "missing"
            fail_ids = ",".join(iter(set(gpu_bus_ids) - set(parsed_results)))
        result["gpu"]["device_count"] = f"FAIL - {fail_type}: {fail_ids}"

    return result

# Main function to call run_gpu_count_check and parse the results
def main(argv=None):
    print("Health check is in progress ...")
    result = run_gpu_count_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()
