#!/usr/bin/env bash
# This script checks the GPU clock speed to determine if it is within the expected threshold of the maximum setting.

MAX=1980
MIN=$(( MAX*90/100 ))  # 90% threshold

gpu=0
fails=0
min_allowed=""

while IFS= read -r line; do
  # Handle errors
  if [[ "$line" == *"couldn't communicate with the NVIDIA driver"* ]]; then
    echo '{"gpu":{"max_clock_speed":"FAIL - NVIDIA driver not loaded"}}' | jq
    exit 1
  fi
  if [[ "$line" == Error* ]]; then
    echo '{"gpu":{"max_clock_speed":"FAIL - cannot run nvidia-smi"}}' | jq
    exit 1
  fi

  # Read current speed
  speed="${line%%[^0-9]*}"
  (( speed < MIN )) && fails=1
  if (( speed >= MIN && speed < MAX )); then
    [[ -z $min_allowed || speed -lt min_allowed ]] && min_allowed=$speed
  fi
  (( gpu++ ))
done < <(nvidia-smi --query-gpu=clocks.current.graphics --format=csv,noheader 2>&1)

# Output result
if (( fails )); then
  echo "{\"gpu\":{\"max_clock_speed\":\"FAIL - check GPU\"}}" | jq
  exit 1
else
  [[ -z $min_allowed ]] && min_allowed=$MAX
  echo "{\"gpu\":{\"max_clock_speed\":\"PASS - Expected $MAX, allowed $min_allowed\"}}" | jq
fi