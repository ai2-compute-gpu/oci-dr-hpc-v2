import shlex
import subprocess
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


def run_max_acc_check():
    mlxconfig_bin = "/usr/bin/mlxconfig"
    config = {
        "pci_ids": [
            "0000:0c:00.0",
            "0000:2a:00.0",
            "0000:41:00.0",
            "0000:58:00.0",
            "0000:86:00.0",
            "0000:a5:00.0",
            "0000:bd:00.0",
            "0000:d5:00.0"
            ]
    }
    pci_ids = config["pci_ids"]

    pci_config_results = []
    for pci in pci_ids:
        cmd = f'sudo {mlxconfig_bin} -d {pci} query'
        output = run_cmd(cmd)
        result = parse_acc_results(pci, output)
        pci_config_results.append(result["pcie_config"])
    result = dict(pcie_config=pci_config_results)
    return result

def parse_acc_results(pci_id="undefined", results="undefined"):
    result = {
        "pcie_config":
            {"pci_busid": pci_id,
             "max_acc_out": "FAIL",
             "advanced_pci_settings": "FAIL"}
    }

    for line in results:
        if "MAX_ACC_OUT_READ" in line and "0" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"
        elif "MAX_ACC_OUT_READ" in line and "44" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"
        elif "MAX_ACC_OUT_READ" in line and "128" in line:
            result["pcie_config"]["max_acc_out"] = "PASS"

        if "ADVANCED_PCI_SETTINGS" in line and "True" in line:
            result["pcie_config"]["advanced_pci_settings"] = "PASS"
    return result

def main(argv=None):
    print("Health check is in progress and the result will be provided within 1 minute.")
    result = run_max_acc_check()
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()

