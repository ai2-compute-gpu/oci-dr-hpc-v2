"""
Check if GPU is in MIG mode (Only for nvidia GPUs)
Multi-Instance GPU (MIG) mode allows GPUs to be securely partitioned into up to seven separate GPU Instances
for CUDA applications, providing multiple users with separate GPU resources for optimal GPU utilization.
"""

import shlex
import subprocess
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

# Function to check if GPU is in MIG mode (Only for nvidia GPUs)
def run_gpu_mode_check():
    cmd = f'nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader'
    output = run_cmd(cmd)
    result = parse_gpu_mode_results(output)
    return result

# Function to parse nvidia-smi output and emit JSON
def parse_gpu_mode_results(results="undefined"):
    result = {
        "gpu":
            {"MIG Mode": "UNKNOWN"}
    }
    wrong_mode = False
    wrong_mode_index = []
    filtered_results = [item for item in results if item != '']
    for line in filtered_results:
        try:
            index, mode = line.split(",")
            if not any(name in mode for name in ("Enabled", "Disabled", "N/A")) and not index.isnumeric():
                wrong_mode = True
                break
        except ValueError:
            wrong_mode = True
            break
        else:
            if "Enabled" in mode:
                wrong_mode = True
                wrong_mode_index.append(index)
    if wrong_mode and wrong_mode_index:
        result["gpu"]["MIG Mode"] = f"FAIL - MIG Mode enabled on GPUs {','.join(wrong_mode_index)}"
    if not wrong_mode and results:
        result["gpu"]["MIG Mode"] = "PASS"
    return result


# Main function to call run_gpu_mode_check and parse the results
def main(argv=None):
    result = run_gpu_mode_check()
    print(json.dumps(result, indent=2))


# Run the main function
if __name__ == "__main__":
    main()
