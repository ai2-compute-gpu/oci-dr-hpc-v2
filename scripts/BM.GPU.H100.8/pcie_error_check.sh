#!/bin/bash

##
## Run PCIe check - checking if each node has PCIe error
##

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

local error_found=false

# Excecute dmesg
local dmesg_output=$(sudo dmesg)

# Parse the output line by line
while IFS= read -r line; do
  if [[ "$line" =~ capabilities ]]; then
      continue
  fi
  # Check for any pcie port error
  if echo "$line" | grep -q '.*pcieport.*[Ee]rror'; then
    echo "ERROR $line"
    error_found=true
  fi
done <<< "$dmesg_output"


if [ "$error_found" == "false" ]; then
  echo "No pcie error detected"
  return 0
fi

