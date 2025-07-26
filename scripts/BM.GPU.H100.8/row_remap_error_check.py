"""
Script to check for GPU row remap errors on NVIDIA GPUs.
This script runs the command `nvidia-smi --query-remapped-rows=gpu_bus_id,remapped_rows.failure --format=csv,noheader`
to gather row remap failure information, and parses the output to determine if any GPUs
have row remap failures. Row remap failures indicate memory errors in the GPU.
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


def parse_row_remap_results(results="undefined"):
    result = {
        "row_remap_error":
            {"status": "PASS"}
    }
    
    # Expected GPU bus IDs for H100 shape
    expected_bus_ids = {
        "00000000:0F:00.0",
        "00000000:2D:00.0", 
        "00000000:44:00.0",
        "00000000:5B:00.0",
        "00000000:89:00.0",
        "00000000:A8:00.0",
        "00000000:C0:00.0",
        "00000000:D8:00.0"
    }
    
    if len(results) == 0:
        result["row_remap_error"]["status"] = "FAIL"
        result["row_remap_error"]["error"] = "No nvidia-smi output received"
        return result
    
    found_bus_ids = set()
    failed_bus_ids = []
    
    for line in results:
        line = line.strip()
        if not line or line.startswith("Error:"):
            continue
            
        # Parse CSV format: gpu_bus_id, remapped_rows.failure
        parts = line.split(',')
        if len(parts) >= 2:
            bus_id = parts[0].strip()
            failure_count = parts[1].strip()
            
            found_bus_ids.add(bus_id)
            
            # Check if failure count is not 0
            try:
                if int(failure_count) != 0:
                    failed_bus_ids.append(bus_id)
            except ValueError:
                # If we can't parse the failure count, treat as failure
                failed_bus_ids.append(bus_id)
    
    # Check if all expected bus IDs are present
    missing_bus_ids = expected_bus_ids - found_bus_ids
    if missing_bus_ids:
        result["row_remap_error"]["status"] = "FAIL"
        result["row_remap_error"]["missing_gpus"] = list(missing_bus_ids)
    
    # Check if any GPUs have row remap failures
    if failed_bus_ids:
        result["row_remap_error"]["status"] = "FAIL"
        result["row_remap_error"]["failed_gpus"] = failed_bus_ids
    
    return result


def get_nvidia_driver_version():
    """ Get nvidia-smi driver version """
    cmd = "nvidia-smi --query-gpu=driver_version --format=csv,noheader,nounits"
    raw_result = run_cmd(cmd)
    
    if len(raw_result) == 0 or raw_result[0].startswith("Error:"):
        return None
    
    # Get the first line and extract version
    version_line = raw_result[0].strip()
    # Extract numeric part (e.g., "550.54.15" -> 550)
    import re
    version_match = re.search(r'(\d+)', version_line)
    if version_match:
        return int(version_match.group(1))
    return None


def run_row_remap_error_check():
    """ Check for GPU row remap errors """
    # Check nvidia-smi driver version first
    driver_version = get_nvidia_driver_version()
    if driver_version is None:
        return {
            "row_remap_error": {
                "status": "FAIL",
                "error": "could not determine nvidia-smi driver version"
            }
        }
    
    if driver_version < 550:
        return {
            "row_remap_error": {
                "status": f"Not applicable for nvidia-smi driver : {driver_version}"
            }
        }
    
    cmd = "nvidia-smi --query-remapped-rows=gpu_bus_id,remapped_rows.failure --format=csv,noheader"
    raw_result = run_cmd(cmd)
    result = parse_row_remap_results(raw_result)
    return result


def main():
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_row_remap_error_check()
    print(json.dumps(result, indent=2))


# Run the main function
if __name__ == "__main__":
    main()