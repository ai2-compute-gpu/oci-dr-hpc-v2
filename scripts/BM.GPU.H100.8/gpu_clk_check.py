"""
# This script checks the GPU clock speed to determine if it is within the expected threshold of the maximum setting.
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
def run_gpu_clk_check():
    """ Check if the clock speed is set to max """
    config = {
        "gpu_clk_check": {
            "state": "disable",
            "clock_speed": "1980"
        },
        "gpu_pci_ids": [
            "0000:0f:00.0",
            "0000:2d:00.0",
            "0000:44:00.0",
            "0000:5b:00.0",
            "0000:89:00.0",
            "0000:a8:00.0",
            "0000:c0:00.0",
            "0000:d8:00.0"
        ]
    }
    max_clock_speed = config["gpu_clk_check"]["clock_speed"]
    cmd = f'nvidia-smi --query-gpu=clocks.current.graphics --format=csv,noheader'
    output = run_cmd(cmd)
    result = parse_gpu_clk_results(output, max_clock_speed, len(config["gpu_pci_ids"]))
    return result


# Function to parse GPU clock to determine if is within the acceptable range
def parse_gpu_clk_results(results="undefined", max_clock_speed="undefined", gpu_count=8):
    result = {
        "gpu":
            {"max_clock_speed": "PASS"}
    }
    fail_list = []
    warn_list = []
    gpu = 0
    original_max_clock_speed = max_clock_speed
    allowed_max_clock_speed = ""
    if not results:
        result["gpu"]["max_clock_speed"] = "FAIL - check GPU "
    for line in results:
        if "couldn't communicate with the NVIDIA driver" in line:
            result["gpu"]["max_clock_speed"] = "FAIL - NVIDIA driver is not loaded"
            return result
        if "Error" in line:
            result["gpu"]["max_clock_speed"] = "FAIL - not able to run command 'nvidia-smi " \
                                               "--query-gpu=clocks.current.graphics --format=csv' "
            return result
    # Allow for an offset for H100
    max_clock_speed = str(int(original_max_clock_speed) - round(int(original_max_clock_speed) * 0.10))
    current_speed = ""
    allowed_speed = ""
    for line in results:
        if len(line) > 0:
            current_speed = line.split()[0]
            # Include the smallest allowed clock among all GPU clocks for allowed message
            if int(current_speed) >= int(max_clock_speed) and int(current_speed) < int(original_max_clock_speed):
                allowed_speed = str(min(int(allowed_speed), int(current_speed))) if allowed_speed else current_speed
            try:
                if int(current_speed) < int(max_clock_speed):
                    fail_list.append(str(gpu))
            except ValueError:
                fail_list.append(str(gpu))
        gpu += 1
    if fail_list:
        result["gpu"]["max_clock_speed"] = "FAIL - check GPU " + ",".join(iter(fail_list))
    else:
        if warn_list:
            result["gpu"]["max_clock_speed"] = f"WARN - Clock speed is lower than max clock speed. Check " + ", ".join(iter(warn_list))
        if results:
            if not allowed_speed:
                allowed_speed = current_speed
            result["gpu"]["max_clock_speed"] = f"PASS - Expected {original_max_clock_speed}, " f"allowed {allowed_speed}"
    return result

# Main function to call run_gpu_clk_check and parse the results
def main(argv=None):
    result = run_gpu_clk_check()
    print(json.dumps(result, indent=2))

# Run the main function
if __name__ == "__main__":
    main()
