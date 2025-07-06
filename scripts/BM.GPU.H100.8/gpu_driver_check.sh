#!/usr/bin/env bash
# Validate that the GPU driver version is supported and not blacklisted.
# Blacklisted and supported versions for H100
BLACKLIST=("470.57.02")
SUPPORTED=("450.119.03" "450.142.0" "470.103.01" "470.129.06" "470.141.03" "510.47.03" "535.104.12" "550.90.12")

# Run nvidia-smi command and capture output for driver version
if ! version="$(nvidia-smi --query-gpu=driver_version --format=csv,noheader 2>&1)"; then
  echo '{"gpu":{"driver_version":"FAIL - nvidia-smi error"}}' | jq
  exit 1
fi

# Take first non-empty line if multiple GPUs
version="$(echo "$version" | awk 'NF{print; exit}')"

# Check presence
if [[ -z "$version" ]]; then
  echo '{"gpu":{"driver_version":"FAIL"}}' | jq
  exit 1
fi

# Check for mismatch: ensure all GPUs have same version
all_versions=( $(nvidia-smi --query-gpu=driver_version --format=csv,noheader) )
first="${all_versions[0]}"
mismatch=false
for v in "${all_versions[@]}"; do
  if [[ "$v" != "$first" ]]; then
    mismatch=true
    break
  fi
done
if $mismatch; then
  echo '{"gpu":{"driver_version":"FAIL - Driver versions are mismatched"}}' | jq
  exit 1
fi

# Compare against blacklist/supported
status="WARN - unsupported driver"
for b in "${BLACKLIST[@]}"; do
  if [[ "$version" == "$b" ]]; then
    status="FAIL"
    break
  fi
done
if [[ "$status" != "FAIL" ]]; then
  for s in "${SUPPORTED[@]}"; do
    if [[ "$version" == "$s" ]]; then
      status="PASS"
      break
    fi
  done
fi

# Output JSON result
printf '{"gpu":{"driver_version":"%s"}}\n' "$status" | jq
if [[ "$status" == "FAIL" ]]; then
  exit 1
fi
