""" 
Simple Peermem Module Health Check Script
This script checks for the presence of the nvidia_peermem module on the system.
The nvidia_peermem module is required for GPU-to-GPU peer memory access,
which is crucial for high-performance computing applications.
Missing this module can lead to performance degradation in multi-GPU setups.
"""

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


def run_peermem_module_check():
    """ Check for presence of peermem module """
    cmd = "/usr/sbin/lsmod"
    module_name = "nvidia_peermem"
    raw_result_list = run_cmd(cmd)
    result = parse_module_results(raw_result_list, module_name)
    return result


def parse_module_results(results, module_name="undefined"):
    result = {
        "gpu":
            {module_name: "FAIL"}
    }
    for line in results:
        if len(line) > 0:
            name = line.split()[0]
            if name == module_name:
                result["gpu"][module_name] = "PASS"
    return result


# Main entry point for the RX discards health check script.
def main():
    print("Health check is in progress ...")
    result = run_peermem_module_check()
    print(json.dumps(result, indent=2))


# Script entry point - only execute main() when script is run directly
if __name__ == "__main__":
    main()
