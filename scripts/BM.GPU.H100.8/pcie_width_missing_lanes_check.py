#!/usr/bin/env python3

"""
PCIe Width Missing Lanes Check

This script checks the PCIe link width for GPU and RDMA interfaces to detect missing lanes.
It verifies that all PCIe links are operating at their expected width configuration.

Expected output for BM.GPU.H100.8:
- GPU/NVSwitch: 4x Width x2 (ok), 8x Width x16 (ok)  
- RDMA: 2x Width x8 (ok), 16x Width x16 (ok)

Author: Oracle Cloud Infrastructure
"""

import json
import os
import subprocess
import sys
from datetime import datetime
from typing import Dict, List, Tuple


def run_command(cmd: str) -> Tuple[int, str, str]:
    """
    Execute a shell command and return exit code, stdout, stderr.
    
    Args:
        cmd: Command to execute
        
    Returns:
        Tuple of (exit_code, stdout, stderr)
    """
    try:
        result = subprocess.run(
            cmd, 
            shell=True, 
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE, 
            universal_newlines=True, 
            timeout=30
        )
        return result.returncode, result.stdout.strip(), result.stderr.strip()
    except subprocess.TimeoutExpired:
        return 124, "", "Command timed out"
    except Exception as e:
        return 1, "", str(e)


def parse_lspci_width_output(output: str) -> Tuple[Dict[str, int], Dict[str, int], List[str]]:
    """
    Parse lspci output to extract PCIe width, speed, and state information.
    
    Args:
        output: Raw output from lspci command
        
    Returns:
        Tuple of (width_counts, speed_counts, state_errors)
    """
    width_counts = {}
    speed_counts = {}
    state_errors = []
    
    lines = output.strip().split('\n')
    for line in lines:
        line = line.strip()
        if not line:
            continue
            
        # Parse line format: "count           LnkSta: Speed 16GT/s (ok), Width x16 (ok)"
        import re
        match = re.match(r'^\s*(\d+)\s+LnkSta:\s*Speed\s+([^\s]+)\s*\(([^)]+)\),\s*Width\s+x(\d+)\s*\(([^)]+)\)', line)
        if match:
            try:
                count = int(match.group(1))
                speed = match.group(2)
                speed_state = match.group(3)
                width = match.group(4)
                width_state = match.group(5)
                
                # Count widths only if width state is ok
                width_key = f"Width x{width}"
                if width_state == "ok":
                    width_counts[width_key] = width_counts.get(width_key, 0) + count
                else:
                    state_errors.append(f"{count} devices have width state '{width_state}' instead of 'ok'")
                
                # Count speeds only if speed state is ok
                speed_key = f"Speed {speed}"
                if speed_state == "ok":
                    speed_counts[speed_key] = speed_counts.get(speed_key, 0) + count
                else:
                    state_errors.append(f"{count} devices have speed state '{speed_state}' instead of 'ok'")
                    
            except (ValueError, IndexError):
                continue
                
    return width_counts, speed_counts, state_errors


def check_gpu_nvswitch_pcie_width() -> Tuple[bool, str, Dict[str, int], Dict[str, int], List[str]]:
    """
    Check PCIe width, speed, and state for GPU and NVSwitch interfaces.
    
    Returns:
        Tuple of (success, error_message, width_counts, speed_counts, state_errors)
    """
    cmd = "sudo lspci -vvv | egrep '^[0-9,a-f]|LnkSta:' | grep -A 1 -i nvidia | grep LnkSta | sort | uniq -c"
    
    exit_code, stdout, stderr = run_command(cmd)
    
    if exit_code != 0:
        return False, f"Failed to execute lspci command: {stderr}", {}, {}, []
    
    if not stdout.strip():
        return False, "No NVIDIA PCIe devices found", {}, {}, []
    
    width_counts, speed_counts, state_errors = parse_lspci_width_output(stdout)
    
    # Expected for BM.GPU.H100.8: 4x Width x2, 8x Width x16
    expected_width_counts = {
        "Width x2": 4,
        "Width x16": 8
    }
    
    # Expected for BM.GPU.H100.8: 4x Speed 16GT/s, 8x Speed 32GT/s
    expected_speed_counts = {
        "Speed 16GT/s": 4,
        "Speed 32GT/s": 8
    }
    
    error_messages = []
    
    # Check width counts
    for width, expected_count in expected_width_counts.items():
        actual_count = width_counts.get(width, 0)
        if actual_count != expected_count:
            error_messages.append(f"GPU/NVSwitch PCIe width mismatch: expected {expected_count}x {width}, got {actual_count}x")
    
    # Check speed counts
    for speed, expected_count in expected_speed_counts.items():
        actual_count = speed_counts.get(speed, 0)
        if actual_count != expected_count:
            error_messages.append(f"GPU/NVSwitch PCIe speed mismatch: expected {expected_count}x {speed}, got {actual_count}x")
    
    # Add state errors
    if state_errors:
        error_messages.extend([f"GPU/NVSwitch state error: {err}" for err in state_errors])
    
    if error_messages:
        return False, "; ".join(error_messages), width_counts, speed_counts, state_errors
    
    return True, "", width_counts, speed_counts, state_errors


def check_rdma_pcie_width() -> Tuple[bool, str, Dict[str, int], Dict[str, int], List[str]]:
    """
    Check PCIe width, speed, and state for RDMA interfaces.
    
    Returns:
        Tuple of (success, error_message, width_counts, speed_counts, state_errors)
    """
    cmd = "sudo lspci -vvv | egrep '^[0-9,a-f]|LnkSta:' | grep -A 1 -i mellanox | grep LnkSta | sort | uniq -c"
    
    exit_code, stdout, stderr = run_command(cmd)
    
    if exit_code != 0:
        return False, f"Failed to execute lspci command: {stderr}", {}, {}, []
    
    if not stdout.strip():
        return False, "No Mellanox PCIe devices found", {}, {}, []
    
    width_counts, speed_counts, state_errors = parse_lspci_width_output(stdout)
    
    # Expected for BM.GPU.H100.8: 2x Width x8, 16x Width x16
    expected_width_counts = {
        "Width x8": 2,
        "Width x16": 16
    }
    
    # Expected for BM.GPU.H100.8: 2x Speed 16GT/s, 16x Speed 32GT/s
    expected_speed_counts = {
        "Speed 16GT/s": 2,
        "Speed 32GT/s": 16
    }
    
    error_messages = []
    
    # Check width counts
    for width, expected_count in expected_width_counts.items():
        actual_count = width_counts.get(width, 0)
        if actual_count != expected_count:
            error_messages.append(f"RDMA PCIe width mismatch: expected {expected_count}x {width}, got {actual_count}x")
    
    # Check speed counts
    for speed, expected_count in expected_speed_counts.items():
        actual_count = speed_counts.get(speed, 0)
        if actual_count != expected_count:
            error_messages.append(f"RDMA PCIe speed mismatch: expected {expected_count}x {speed}, got {actual_count}x")
    
    # Add state errors
    if state_errors:
        error_messages.extend([f"RDMA state error: {err}" for err in state_errors])
    
    if error_messages:
        return False, "; ".join(error_messages), width_counts, speed_counts, state_errors
    
    return True, "", width_counts, speed_counts, state_errors


def get_oci_shape() -> str:
    """
    Get the current OCI shape from IMDS or environment variable.
    
    Returns:
        OCI shape string
    """
    # First try environment variable (useful for testing)
    shape = os.environ.get("OCI_SHAPE")
    if shape:
        return shape
    
    # Try IMDS
    try:
        cmd = "curl -s -m 10 http://169.254.169.254/opc/v1/instance/shape"
        exit_code, stdout, stderr = run_command(cmd)
        if exit_code == 0 and stdout.strip():
            return stdout.strip()
    except Exception:
        pass
    
    return "UNKNOWN"


def main():
    """Main function to run PCIe width missing lanes check."""
    
    overall_success = True
    error_messages = []
    
    # Check GPU/NVSwitch PCIe width, speed, and state
    gpu_success, gpu_error, gpu_width_counts, gpu_speed_counts, gpu_state_errors = check_gpu_nvswitch_pcie_width()
    
    if not gpu_success:
        overall_success = False
        error_messages.append(f"GPU/NVSwitch: {gpu_error}")
    
    # Check RDMA PCIe width, speed, and state
    rdma_success, rdma_error, rdma_width_counts, rdma_speed_counts, rdma_state_errors = check_rdma_pcie_width()
    
    if not rdma_success:
        overall_success = False
        error_messages.append(f"RDMA: {rdma_error}")
    
    # Create result in simple format matching other scripts
    if overall_success:
        status = "PASS"
    else:
        status = f"FAIL - {'; '.join(error_messages)}"
    
    result = {
        "pcie_width_missing_lanes": {
            "status": status
        }
    }
    
    print(json.dumps(result, indent=2))
    
    # Exit with appropriate code
    return 0 if overall_success else 1


if __name__ == "__main__":
    sys.exit(main())