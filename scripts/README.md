# Scripts Directory

This directory contains a collection of Python and Bash scripts organized by OCI GPU shapes types for running Level 1, Level 2, and Level 3 tests.

## Directory Structure

The scripts are organized by Oracle Cloud Infrastructure (OCI) bare metal GPU shapes:

- **BM.GPU.B200.8/** - Scripts for B200 GPU instances (8 GPUs)
- **BM.GPU.GB200.4/** - Scripts for GB200 GPU instances (4 GPUs)  
- **BM.GPU.H100.8/** - Scripts for H100 GPU instances (8 GPUs)
- **BM.GPU.H200.8/** - Scripts for H200 GPU instances (8 GPUs)

## Test Levels

### Level 1 Tests (Health Checks - Passive - It does not impact running workloads)
- **gpu_count_check.sh** - Verify expected GPU count
- **rdma_nic_count_check.sh** - Check RDMA network interface count
- **pcie_error_check.sh** - Scan for PCIe errors
- **walk_pcie_check.sh** - Comprehensive PCIe topology validation
- **max_acc_check.py** - Python script for maximum acceleration testing
- **max_acc_check.sh** - Bash wrapper for acceleration checks
- **gpu_clock_set_to_max.sh** - Set GPU clocks to maximum performance


### Level 2 Tests (Performance Validation - Active - It may impact running workloads)
TBD

### Level 3 Tests (Advanced Diagnostics- Active - It may impact running workloads)
TBD


## Usage

Scripts are designed to be executed by the main `oci-dr-hpc-v2` application based on the detected OCI shape and selected test level. They can also be run independently for manual testing and validation.

Each script is tailored to the specific hardware configuration and capabilities of its corresponding OCI shape.

@rekharoy