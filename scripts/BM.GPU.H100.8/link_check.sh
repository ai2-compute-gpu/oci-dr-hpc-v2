#!/bin/bash
# This script checks the state of each RDMA NIC link and validates link parameters.

echo "Health check is in progress ..."

# Expected RDMA mlx device names for this shape
EXPECTED_MLX_DEVICES=(
  "mlx5_0" "mlx5_1" "mlx5_3" "mlx5_4" "mlx5_5" "mlx5_6" "mlx5_7" "mlx5_8"
  "mlx5_9" "mlx5_10" "mlx5_12" "mlx5_13" "mlx5_14" "mlx5_15" "mlx5_16" "mlx5_17"
)
EXPECTED_SPEED="200G"
EXPECTED_WIDTH="4x"
EXPECTED_EFFECTIVE_PHYSICAL_ERRORS=0
EXPECTED_RAW_PHYSICAL_ERRORS_PER_LANE=10000

IBDEV2NETDEV_BIN=$(command -v ibdev2netdev)
MLXLINK_BIN=$(command -v mlxlink)

if [[ -z "$IBDEV2NETDEV_BIN" || -z "$MLXLINK_BIN" ]]; then
  echo "Required binaries 'ibdev2netdev' or 'mlxlink' not found in PATH." >&2
  exit 1
fi

declare -A DEVICE_TO_INTERFACE_MAP

# Get device list and map mlx devices to OS interface names
while read -r line; do
  set -- $line
  device=$1
  interface=${5:-}
  if [[ -n "$interface" ]]; then
    DEVICE_TO_INTERFACE_MAP["$device"]=$interface
  fi
done < <(sudo "$IBDEV2NETDEV_BIN")

# Find OS interface names for expected mlx devices and create failure results for missing ones
INTERFACES_TO_CHECK=()
RESULTS=()

for expected_device in "${EXPECTED_MLX_DEVICES[@]}"; do
  if [[ -n "${DEVICE_TO_INTERFACE_MAP[$expected_device]}" ]]; then
    os_interface="${DEVICE_TO_INTERFACE_MAP[$expected_device]}"
    INTERFACES_TO_CHECK+=("$os_interface")
  else
    # Create a failure result for the missing device
    RESULTS+=($(jq -n \
      --arg device "$expected_device" \
      '{device: $device, link_speed: ("FAIL - Device " + $device + " not found"), link_state: ("FAIL - Device " + $device + " not found"), physical_state: ("FAIL - Device " + $device + " not found"), link_width: ("FAIL - Device " + $device + " not found"), link_status: ("FAIL - Device " + $device + " not found"), effective_physical_errors: ("FAIL - Device " + $device + " not found"), effective_physical_ber: ("FAIL - Device " + $device + " not found"), raw_physical_errors_per_lane: ("FAIL - Device " + $device + " not found"), raw_physical_ber: ("FAIL - Device " + $device + " not found")}'))
  fi
done

if [[ ${#INTERFACES_TO_CHECK[@]} -eq 0 && ${#RESULTS[@]} -eq 0 ]]; then
  echo "No expected RDMA devices found on the system" >&2
  exit 1
fi

for interface in "${INTERFACES_TO_CHECK[@]}"; do
  # Find the device name for this interface
  device=""
  for dev in "${!DEVICE_TO_INTERFACE_MAP[@]}"; do
    if [[ "${DEVICE_TO_INTERFACE_MAP[$dev]}" == "$interface" ]]; then
      device="$dev"
      break
    fi
  done
  
  if [[ -n "$device" ]]; then
    output=$(sudo "$MLXLINK_BIN" -d "$device" --json --show_module --show_counters --show_eye 2>&1)
    # Try to extract JSON from output
    json_part=$(echo "$output" | awk '/{/{flag=1}flag')
    if [[ -z "$json_part" ]]; then
      RESULTS+=("{\"device\": \"$interface\", \"status\": \"FAIL - Invalid interface: $interface\"}")
      continue
    fi

    # Extract fields using jq
    speed=$(echo "$json_part" | jq -r '.result.output["Operational Info"].Speed // ""')
    state=$(echo "$json_part" | jq -r '.result.output["Operational Info"].State // ""')
    phys_state=$(echo "$json_part" | jq -r '.result.output["Operational Info"]["Physical state"] // ""')
    width=$(echo "$json_part" | jq -r '.result.output["Operational Info"].Width // ""')
    status_opcode=$(echo "$json_part" | jq -r '.result.output["Troubleshooting Info"]["Status Opcode"] // ""')
    recommendation=$(echo "$json_part" | jq -r '.result.output["Troubleshooting Info"].Recommendation // ""')
    effective_physical_errors=$(echo "$json_part" | jq -r '.result.output["Physical Counters and BER Info"]["Effective Physical Errors"] // ""')
    effective_physical_ber=$(echo "$json_part" | jq -r '.result.output["Physical Counters and BER Info"]["Effective Physical BER"] // ""')
    raw_physical_errors_per_lane=$(echo "$json_part" | jq -r '.result.output["Physical Counters and BER Info"]["Raw Physical Errors Per Lane"] ' | tr -d '"')
    raw_physical_ber=$(echo "$json_part" | jq -r '.result.output["Physical Counters and BER Info"]["Raw Physical BER"] // ""')

    # Set initial FAILs
    link_speed="FAIL - $speed, expected $EXPECTED_SPEED"
    link_state="FAIL - $state, expected Active"
    physical_state="FAIL - $phys_state, expected LinkUp/ETH_AN_FSM_ENABLE"
    link_width="FAIL - $width, expected $EXPECTED_WIDTH"
    link_status="FAIL - $recommendation"
    effective_physical_errors_status="PASS"
    effective_physical_ber_status="FAIL - $effective_physical_ber"
    raw_physical_errors_per_lane_status="PASS"
    raw_physical_ber_status="FAIL - $raw_physical_ber"

    # Set PASS if matches
    [[ "$speed" == *"$EXPECTED_SPEED"* ]] && link_speed="PASS"
    [[ "$state" == "Active" ]] && link_state="PASS"
    [[ "$phys_state" == "LinkUp" || "$phys_state" == "ETH_AN_FSM_ENABLE" ]] && physical_state="PASS"
    [[ "$width" == "$EXPECTED_WIDTH" ]] && link_width="PASS"
    [[ "$status_opcode" == "0" ]] && link_status="PASS"
    awk_float_lt() { awk -v n1="$1" -v n2="$2" 'BEGIN { if (n1+0 < n2+0) exit 0; exit 1 }'; }
    awk_float_gt() { awk -v n1="$1" -v n2="$2" 'BEGIN { if (n1+0 > n2+0) exit 0; exit 1 }'; }
    if [[ "$effective_physical_ber" =~ ^[0-9.eE+-]+$ ]] && awk_float_lt "$effective_physical_ber" "1e-12"; then
      effective_physical_ber_status="PASS"
    fi
    if [[ "$raw_physical_ber" =~ ^[0-9.eE+-]+$ ]] && awk_float_lt "$raw_physical_ber" "1e-5"; then
      raw_physical_ber_status="PASS"
    fi
    if [[ "$effective_physical_errors" =~ ^[0-9]+$ ]] && [[ "$effective_physical_errors" -gt "$EXPECTED_EFFECTIVE_PHYSICAL_ERRORS" ]]; then
      effective_physical_errors_status="FAIL - $effective_physical_errors"
    fi
    # Check raw_physical_errors_per_lane
    if [[ -n "$raw_physical_errors_per_lane" ]]; then
      IFS=',' read -ra lane_errors <<< "$raw_physical_errors_per_lane"
      for lane_error in "${lane_errors[@]}"; do
        lane_error=$(echo "$lane_error" | xargs)
        [[ "$lane_error" == "undefined" ]] && continue
        if [[ "$lane_error" =~ ^[0-9]+$ ]] && [[ "$lane_error" -gt "$EXPECTED_RAW_PHYSICAL_ERRORS_PER_LANE" ]]; then
          raw_physical_errors_per_lane_status="WARN - $raw_physical_errors_per_lane"
          break
        fi
      done
    fi

    # Compose JSON for this device
    RESULTS+=("$(jq -n \
      --arg device "$interface" \
      --arg link_speed "$link_speed" \
      --arg link_state "$link_state" \
      --arg physical_state "$physical_state" \
      --arg link_width "$link_width" \
      --arg link_status "$link_status" \
      --arg effective_physical_errors "$effective_physical_errors_status" \
      --arg effective_physical_ber "$effective_physical_ber_status" \
      --arg raw_physical_errors_per_lane "$raw_physical_errors_per_lane_status" \
      --arg raw_physical_ber "$raw_physical_ber_status" \
      '{device: $device, link_speed: $link_speed, link_state: $link_state, physical_state: $physical_state, link_width: $link_width, link_status: $link_status, effective_physical_errors: $effective_physical_errors, effective_physical_ber: $effective_physical_ber, raw_physical_errors_per_lane: $raw_physical_errors_per_lane, raw_physical_ber: $raw_physical_ber}')")
  else
    RESULTS+=("{\"device\": \"$interface\", \"status\": \"FAIL - Invalid interface: $interface\"}")
  fi
done

# Print results as JSON object using jq
jq -n --argjson results "$(printf '%s\n' "${RESULTS[@]}" | jq -s '.')" '{link: $results}'
