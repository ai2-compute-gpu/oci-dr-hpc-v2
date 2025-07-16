#!/bin/bash
# This script checks the authentication status of RDMA interfaces using wpa_cli.

echo "RDMA interface authentication check is in progress ..."

IBDEV2NETDEV_BIN=$(command -v ibdev2netdev)
WPA_CLI_BIN=$(command -v wpa_cli)

if [[ -z "$IBDEV2NETDEV_BIN" ]]; then
  echo "Required binary 'ibdev2netdev' not found in PATH." >&2
  exit 1
fi

if [[ -z "$WPA_CLI_BIN" ]]; then
  echo "Required binary 'wpa_cli' not found in PATH." >&2
  exit 1
fi

# RDMA device names for BM.GPU.H100.8 (based on shapes.json)
# These are the ConnectX-7 devices used for RDMA, excluding VCN devices
RDMA_DEVICES=("mlx5_0" "mlx5_1" "mlx5_3" "mlx5_4" "mlx5_5" "mlx5_6" "mlx5_7" "mlx5_8" "mlx5_9" "mlx5_10" "mlx5_12" "mlx5_13" "mlx5_14" "mlx5_15" "mlx5_16" "mlx5_17")

declare -A DEVICE_TO_INTERFACE_MAP

# Get device to interface mapping from ibdev2netdev
while read -r line; do
  set -- $line
  device=$1
  interface=${5:-}
  if [[ -n "$interface" ]]; then
    DEVICE_TO_INTERFACE_MAP["$device"]=$interface
  fi
done < <(sudo "$IBDEV2NETDEV_BIN")

RESULTS=()

# Check only the RDMA devices specified for H100
for device in "${RDMA_DEVICES[@]}"; do
  interface="${DEVICE_TO_INTERFACE_MAP[$device]:-}"
  
  if [[ -n "$interface" ]]; then
    echo "Checking RDMA device $device (interface $interface)..."
    
    # Run wpa_cli status command for this interface
    output=$(sudo "$WPA_CLI_BIN" -i "$interface" status 2>&1)
    
    # Check for specific authenticated status in the output
    supplicant_status=""
    if echo "$output" | grep -q "Supplicant PAE state=AUTHENTICATED"; then
      supplicant_status="PASS"
    elif [[ $? -ne 0 ]] || echo "$output" | grep -q -i "error\|failed\|no such"; then
      supplicant_status="FAIL - Unable to run wpa_cli command"
    else
      supplicant_status="FAIL - Interface not authenticated"
    fi

    # Compose JSON for this device
    RESULTS+=("$(jq -n \
      --arg device "$interface" \
      --arg auth_status "$supplicant_status" \
      '{device: $device, auth_status: $auth_status}')")
  else
    echo "Warning: RDMA device $device not found in device mapping"
  fi
done

# Print results as JSON object using jq
jq -n --argjson results "$(printf '%s\n' "${RESULTS[@]}" | jq -s '.')" '{auth_check: $results}'