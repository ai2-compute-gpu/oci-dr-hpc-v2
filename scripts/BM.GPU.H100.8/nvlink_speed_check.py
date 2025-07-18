"""
Script to check NVLink speed and presence on NVIDIA GPUs.
This script runs the command `nvidia-smi nvlink -s` to gather NVLink
speed and presence information, and parses the output to determine if the NVLink
is functioning correctly based on expected speed and count."""
import shlex
import subprocess
import json
import re


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


def parse_nvlink_results(results="undefined", expected_speed=25.0,
                         expected_count="undefined"):
    result = {
        "gpu":
            {"nvlink": "PASS"}
    }
    nvlink_dict = {}
    count = 0
    gpu_num = None
    if len(results) == 0:
        result["gpu"]["nvlink"] = "FAIL - check GPU "

    if not isinstance(expected_speed, float):
        numeric_part = re.search(r'(\d+\.?\d+)', str(expected_speed))
        if numeric_part:
            # If numeric part is found, convert it to float
            expected_speed = float(numeric_part.group())
        else:
            expected_speed = float(expected_speed)

    for line in results:
        gpu_line = re.search('GPU\s+(\d+):\s+(?:NVIDIA|HGX)', line)
        link_line = re.search('\s+Link\s+(\d+):\s+(.*)\s+GB/s', line)
        if gpu_line:
            gpu_num = gpu_line.group(1)
            count = 0
            link_speed = None
        elif link_line:
            if "inactive" not in line:
                link_speed = link_line.group(2)
                if float(link_speed) >= expected_speed:
                    count += 1
        else:
            # This works b/c the output of `nvidia-smi nvlink -s` should only be a GPU line or a Link line.
            # Anything else means there's a problem.
            result["gpu"]["nvlink"] = f"FAIL - unexpected entry in nvidia-smi nvlink -s output, output: {line}"

        if count:
            nvlink_dict[gpu_num] = count

    fail_list = []
    for gpu in nvlink_dict:
        current_count = nvlink_dict[gpu]
        if current_count != expected_count or len(results) == 0:
            fail_list.append(gpu)
    if fail_list:
        result["gpu"]["nvlink"] = "FAIL - check GPU " + ",".join(iter(fail_list))
    return result


def run_nvlink_speed_check():
    """ Check NVLink presence and speed """
    expected_speed = 26
    expected_count = 18

    cmd = "nvidia-smi nvlink -s"
    raw_result = run_cmd(cmd)
    result = parse_nvlink_results(raw_result, expected_speed, expected_count)
    return result

def main(argv=None):
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_nvlink_speed_check()
    print(json.dumps(result, indent=2))


# Run the main function
if __name__ == "__main__":
    main()