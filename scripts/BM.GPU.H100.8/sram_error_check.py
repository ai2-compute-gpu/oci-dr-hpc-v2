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


# Check if a string can be converted to an integer
def isint(num):
    try:
        int(num)
        return True
    except ValueError:
        return False


# Check GPU SRAM (Static RAM) memory errors using nvidia-smi
def run_sram_error_check():
    # Configuration thresholds for GPU memory error tolerance
    config = {
        "sram_error_check": {
            "sram_uncorrectable_threshold": 5,  # Critical: Any uncorrectable errors indicate serious hardware issues
            "sram_correctable_threshold": 1000  # Warning: High correctable errors suggest memory degradation
        },
    }

    # Get raw data by run command
    cmd_uncorrectable = f'sudo nvidia-smi -q | grep -A 3 Aggregate | grep Uncorrectable'
    cmd_correctable = f'sudo nvidia-smi -q | grep -A 3 Aggregate | grep Correctable'

    sram_uncorrectable_result = run_cmd(cmd_uncorrectable)
    sram_correctable_result = run_cmd(cmd_correctable)

    # Get threshold
    sram_uncorrectable_threshold = config["sram_error_check"]["sram_uncorrectable_threshold"]
    sram_correctable_threshold = config["sram_error_check"]["sram_correctable_threshold"]

    sram_uncorrectable_list = []
    sram_correctable_list = []
    parity = 0

    # Parity errors are added to SEC-DED errors for total uncorrectable count
    for line in sram_uncorrectable_result:
        if ":" in line:
            if "Parity" in line:
                parity = int(line.split(':')[1].strip())
            elif "SEC-DED" in line:
                sram_uncorrectable_list.append(str(int(line.split(':')[1].strip()) + parity))
            else:
                sram_uncorrectable_list.append(line.split(':')[1].strip())
    for line in sram_correctable_result:
        sram_correctable_list.append(line.split(':')[1].strip())

    return parse_sram_results(sram_uncorrectable_list, sram_correctable_list,
                              sram_uncorrectable_threshold, sram_correctable_threshold)


# Analyze GPU SRAM error counts against thresholds to determine system health status
def parse_sram_results(sram_uncorrectable_list=None,
                       sram_correctable_list=None, sram_uncorrectable_threshold=1,
                       sram_correctable_threshold=1000):
    """
    Health Assessment Logic:
        - FAIL: No error data available OR any uncorrectable errors exceed threshold
        - WARN: Correctable errors exceed threshold (indicates memory degradation)
        - PASS: All error counts within acceptable limits
    """
    result = {
        "sram":
            {"status": "PASS"}
    }

    if len(sram_uncorrectable_list) == 0 or len(sram_correctable_list) == 0:
        result = {
            "sram":
                {"status": "FAIL"}
        }

    for value in sram_uncorrectable_list:
        if isint(value):
            if int(value) > sram_uncorrectable_threshold:
                result["sram"]["status"] = "FAIL"

    for value in sram_correctable_list:
        if isint(value):
            if int(value) > sram_correctable_threshold:
                result["sram"]["status"] = "WARN - SRAM Correctable Exceeded Threshold"
    return result


def main():
    print("Health check is in progress ...")
    result = run_sram_error_check()
    print(json.dumps(result, indent=2))


# Script entry point - only execute main() when script is run directly
if __name__ == "__main__":
    main()
