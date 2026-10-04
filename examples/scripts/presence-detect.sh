#!/usr/bin/env bash
#
# presence-detect.sh - Network presence-based home automation
#
# Detects when family members arrive/leave by pinging their phones on the
# local network. This goes beyond Shelly app by integrating with network
# state that Shelly devices can't detect themselves.
#
# Usage: ./presence-detect.sh [--daemon]
#
# Configure DEVICES with phone IPs and associated actions below.
# Run with --daemon to continuously monitor (or add to cron).

set -euo pipefail

# Configuration: Map phone IPs to device actions
# Format: "phone_ip:device:action_on_arrive:action_on_leave"
TRACKED_DEVICES=(
    "192.168.1.100:porch-light:on:off"       # Alice's phone
    "192.168.1.101:garage-light:on:off"      # Bob's phone
    "192.168.1.102:entryway-light:on:off"    # Guest phone
)

# How many successful pings = "home"
PING_THRESHOLD=2
# How many failed pings = "away"
AWAY_THRESHOLD=3
# Seconds between checks
CHECK_INTERVAL=30
# State file to track presence
STATE_FILE="${HOME}/.shelly-presence-state"

declare -A presence_state
declare -A away_count

log() {
    printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

# Load previous state
load_state() {
    if [[ -f "${STATE_FILE}" ]]; then
        while IFS='=' read -r key value; do
            presence_state["${key}"]="${value}"
        done < "${STATE_FILE}"
    fi
}

# Save state
save_state() {
    : > "${STATE_FILE}"
    for key in "${!presence_state[@]}"; do
        printf '%s=%s\n' "${key}" "${presence_state[${key}]}" >> "${STATE_FILE}"
    done
}

# Ping a device, return 0 if reachable
is_reachable() {
    local ip="${1}"
    ping -c 1 -W 1 "${ip}" &>/dev/null
}

# Execute shelly action
do_action() {
    local device="${1}"
    local action="${2}"

    case "${action}" in
        on)
            shelly on "${device}" -q 2>/dev/null || log "  Warning: Failed to turn on ${device}"
            ;;
        off)
            shelly off "${device}" -q 2>/dev/null || log "  Warning: Failed to turn off ${device}"
            ;;
        *)
            log "  Unknown action: ${action}"
            ;;
    esac
}

check_presence() {
    for entry in "${TRACKED_DEVICES[@]}"; do
        IFS=':' read -r phone_ip device action_arrive action_leave <<< "${entry}"

        local prev_state="${presence_state[${phone_ip}]:-unknown}"
        local current_away="${away_count[${phone_ip}]:-0}"

        if is_reachable "${phone_ip}"; then
            # Device is reachable
            away_count["${phone_ip}"]=0

            if [[ "${prev_state}" != "home" ]]; then
                log "ARRIVAL: ${phone_ip} detected -> turning ${device} ${action_arrive}"
                do_action "${device}" "${action_arrive}"
                presence_state["${phone_ip}"]="home"
            fi
        else
            # Device not reachable
            ((current_away++)) || true
            away_count["${phone_ip}"]="${current_away}"

            if [[ "${current_away}" -ge "${AWAY_THRESHOLD}" && "${prev_state}" != "away" ]]; then
                log "DEPARTURE: ${phone_ip} gone -> turning ${device} ${action_leave}"
                do_action "${device}" "${action_leave}"
                presence_state["${phone_ip}"]="away"
            fi
        fi
    done

    save_state
}

# One-time check
single_check() {
    load_state
    log "Checking presence..."
    check_presence
    log "Done"
}

# Continuous monitoring
daemon_mode() {
    load_state
    log "Starting presence detection daemon (Ctrl+C to stop)"
    log "Monitoring ${#TRACKED_DEVICES[@]} device(s)"

    while true; do
        check_presence
        sleep "${CHECK_INTERVAL}"
    done
}

# Main
case "${1:-}" in
    --daemon)
        daemon_mode
        ;;
    *)
        single_check
        ;;
esac
