#!/usr/bin/env bash
"""
Check if GPU is in MIG mode (Only for nvidia GPUs)
Multi-Instance GPU (MIG) mode allows GPUs to be securely partitioned into up to seven separate GPU Instances
for CUDA applications, providing multiple users with separate GPU resources for optimal GPU utilization.
"""

# Query NVIDIA GPUs for their MIG mode
lines=$(nvidia-smi --query-gpu=index,mig.mode.current --format=csv,noheader 2>&1)
if [[ $? -ne 0 ]]; then
  echo '{"gpu": {"MIG Mode": "UNKNOWN"}}' | jq
  exit 1
fi

fails=()
while IFS=, read -r idx mode; do
  idx=${idx//[[:space:]]/}
  mode=${mode//[[:space:]]/}
  if [[ "$mode" == "Enabled" ]]; then
    fails+=("$idx")
  elif [[ "$mode" != "Disabled" && "$mode" != "N/A" ]]; then
    echo '{"gpu": {"MIG Mode": "UNKNOWN"}}' | jq
    exit 1
  fi
done < <(printf '%s\n' "$lines")

if ((${#fails[@]})); then
  echo "{\"gpu\": {\"MIG Mode\": \"FAIL - MIG Mode enabled on GPUs ${fails[*]}\"}}"
elif [[ -n "$lines" ]]; then
  echo '{"gpu": {"MIG Mode": "PASS"}}' | jq
else
  echo '{"gpu": {"MIG Mode": "UNKNOWN"}}' | jq
  exit 1
fi
