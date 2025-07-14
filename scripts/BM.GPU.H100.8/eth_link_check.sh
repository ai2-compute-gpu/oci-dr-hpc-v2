#!/bin/bash
# This script checks the state of each 100GbE RoCE NIC (non-RDMA Ethernet interfaces).

echo "Ethernet link health check is in progress ..."

# Configuration for BM.GPU.H100.8
EXPECTED_SPEED="100G"
EXPECTED_WIDTH="4x"
EXPECTED_EFFECTIVE_PHYSICAL_ERRORS=0
EXPECTED_RAW_PHYSICAL_ERRORS_PER_LANE=10000
EXPECTED_EFFECTIVE_PHYSICAL_BER="1e-12"
EXPECTED_RAW_PHYSICAL_BER="1e-5"

IBDEV2NETDEV_BIN=$(command -v ibdev2netdev)
MLXLINK_BIN=$(command -v mlxlink)
MST_BIN=$(command -v mst)

if [[ -z "$IBDEV2NETDEV_BIN" || -z "$MLXLINK_BIN" || -z "$MST_BIN" ]]; then
  echo "Required binaries 'ibdev2netdev', 'mlxlink', or 'mst' not found in PATH." >&2
  exit 1
fi

# Hard-coded VCN device names for BM.GPU.H100.8 (non-RDMA Ethernet interfaces)
# Based on shapes.json configuration for H100
VCN_DEVICES=("mlx5_2" "mlx5_11")

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

# Check only the VCN devices specified for H100
for device in "${VCN_DEVICES[@]}"; do
  interface="${DEVICE_TO_INTERFACE_MAP[$device]:-}"
  
  if [[ -n "$interface" ]]; then
    echo "Checking VCN device $device (interface $interface)..."
    output=$(sudo "$MLXLINK_BIN" -d "$device" --json --show_module --show_counters --show_eye 2>&1)
    
    # Try to extract JSON from output
    json_part=$(echo "$output" | awk '/{/{flag=1}flag')
    if [[ -z "$json_part" ]]; then
      RESULTS+=("$(jq -n \
        --arg device "$interface" \
        '{device: $device, eth_link_speed: "FAIL - Unable to parse mlxlink output", eth_link_state: "FAIL - Unable to parse mlxlink output", physical_state: "FAIL - Unable to parse mlxlink output", eth_link_width: "FAIL - Unable to parse mlxlink output", eth_link_status: "FAIL - Unable to parse mlxlink output", effective_physical_errors: "FAIL - Unable to parse mlxlink output", effective_physical_ber: "FAIL - Unable to parse mlxlink output", raw_physical_errors_per_lane: "FAIL - Unable to parse mlxlink output", raw_physical_ber: "FAIL - Unable to parse mlxlink output"}')")
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
    eth_link_speed="FAIL - $speed, expected $EXPECTED_SPEED"
    eth_link_state="FAIL - $state, expected Active"
    physical_state="FAIL - $phys_state, expected LinkUp/ETH_AN_FSM_ENABLE"
    eth_link_width="FAIL - $width, expected $EXPECTED_WIDTH"
    eth_link_status="FAIL - $recommendation"
    effective_physical_errors_status="PASS"
    effective_physical_ber_status="FAIL - $effective_physical_ber"
    raw_physical_errors_per_lane_status="PASS"
    raw_physical_ber_status="FAIL - $raw_physical_ber"

    # Set PASS if matches
    [[ "$speed" == *"$EXPECTED_SPEED"* ]] && eth_link_speed="PASS"
    [[ "$state" == "Active" ]] && eth_link_state="PASS"
    [[ "$phys_state" == "LinkUp" || "$phys_state" == "ETH_AN_FSM_ENABLE" ]] && physical_state="PASS"
    [[ "$width" == "$EXPECTED_WIDTH" ]] && eth_link_width="PASS"
    [[ "$status_opcode" == "0" ]] && eth_link_status="PASS"
    
    # Helper functions for float comparison
    awk_float_lt() { awk -v n1="$1" -v n2="$2" 'BEGIN { if (n1+0 < n2+0) exit 0; exit 1 }'; }
    awk_float_gt() { awk -v n1="$1" -v n2="$2" 'BEGIN { if (n1+0 > n2+0) exit 0; exit 1 }'; }
    
    if [[ "$effective_physical_ber" =~ ^[0-9.eE+-]+$ ]] && awk_float_lt "$effective_physical_ber" "$EXPECTED_EFFECTIVE_PHYSICAL_BER"; then
      effective_physical_ber_status="PASS"
    fi
    if [[ "$raw_physical_ber" =~ ^[0-9.eE+-]+$ ]] && awk_float_lt "$raw_physical_ber" "$EXPECTED_RAW_PHYSICAL_BER"; then
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
      --arg eth_link_speed "$eth_link_speed" \
      --arg eth_link_state "$eth_link_state" \
      --arg physical_state "$physical_state" \
      --arg eth_link_width "$eth_link_width" \
      --arg eth_link_status "$eth_link_status" \
      --arg effective_physical_errors "$effective_physical_errors_status" \
      --arg effective_physical_ber "$effective_physical_ber_status" \
      --arg raw_physical_errors_per_lane "$raw_physical_errors_per_lane_status" \
      --arg raw_physical_ber "$raw_physical_ber_status" \
      '{device: $device, eth_link_speed: $eth_link_speed, eth_link_state: $eth_link_state, physical_state: $physical_state, eth_link_width: $eth_link_width, eth_link_status: $eth_link_status, effective_physical_errors: $effective_physical_errors, effective_physical_ber: $effective_physical_ber, raw_physical_errors_per_lane: $raw_physical_errors_per_lane, raw_physical_ber: $raw_physical_ber}')")
  else
    echo "Warning: VCN device $device not found in device mapping"
  fi
done

# Print results as JSON object using jq
jq -n --argjson results "$(printf '%s\n' "${RESULTS[@]}" | jq -s '.')" '{eth_link: $results}'