---
title: "Script Examples"
description: "Shell scripts for automation and batch operations"
weight: 30
---

Shell scripts demonstrating advanced automation patterns with Shelly CLI.

## Usage

```bash
# Make executable
chmod +x examples/scripts/bulk-update.sh

# Run with dry-run first
./examples/scripts/bulk-update.sh --dry-run

# Run for real
./examples/scripts/bulk-update.sh
```

---

## bulk-update.sh

bulk-update.sh - Bulk firmware update with safety checks

```bash
#!/usr/bin/env bash
#
# bulk-update.sh - Bulk firmware update with safety checks
#
# Updates firmware on multiple devices with proper error handling,
# progress tracking, and optional backup capability.
#
# Usage: ./bulk-update.sh [--dry-run] [--sequential] [--skip-backup] [--yes]
#
# Options:
#   --dry-run      Check for updates without applying
#   --sequential   Update one device at a time (safer, slower)
#   --skip-backup  Skip backup before update (not recommended)
#   --yes          Skip confirmation prompt

set -euo pipefail

# Configuration
DRY_RUN=false
SEQUENTIAL=false
SKIP_BACKUP=false
SKIP_CONFIRM=false
BACKUP_DIR="${HOME}/shelly-backups/pre-update-$(date '+%Y%m%d-%H%M%S')"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "${1}" in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --sequential)
            SEQUENTIAL=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
            shift
            ;;
        --yes|-y)
            SKIP_CONFIRM=true
            shift
            ;;
        *)
            printf 'Unknown option: %s\n' "${1}"
            exit 1
            ;;
    esac
done

log() {
    printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

error() {
    printf '[%s] ERROR: %s\n' "$(date '+%H:%M:%S')" "$*" >&2
}

# Check if shelly CLI is available
if ! command -v shelly &>/dev/null; then
    printf 'Error: shelly CLI not found in PATH\n'
    exit 1
fi

log "Checking for firmware updates..."

# Get list of devices with updates available
updates_json=$(shelly firmware check --all -o json 2>/dev/null || printf '[]')

if [[ "${updates_json}" == "[]" ]]; then
    log "No devices found or unable to check updates"
    exit 0
fi

# Filter devices with updates
devices_with_updates=$(printf '%s' "${updates_json}" | jq -r '.[] | select(.update_available == true) | .name' 2>/dev/null || true)

if [[ -z "${devices_with_updates}" ]]; then
    log "All devices are up to date!"
    exit 0
fi

# Count updates
update_count=$(printf '%s' "${devices_with_updates}" | wc -l | tr -d ' ')
log "Found ${update_count} device(s) with available updates:"
printf '\n'

# Show what will be updated
printf '%s' "${updates_json}" | jq -r '.[] | select(.update_available == true) | "  \(.name): \(.current_version) → \(.new_version)"' 2>/dev/null || true
printf '\n'

if ${DRY_RUN}; then
    log "Dry run mode - no updates will be applied"
    exit 0
fi

# Confirm with user unless --yes
if ! ${SKIP_CONFIRM}; then
    read -p "Proceed with updates? (y/N) " -n 1 -r
    printf '\n'
    if [[ ! ${REPLY} =~ ^[Yy]$ ]]; then
        log "Update cancelled"
        exit 0
    fi
fi

# Create backup unless skipped
if ! ${SKIP_BACKUP}; then
    log "Creating backups in ${BACKUP_DIR}..."
    mkdir -p "${BACKUP_DIR}"

    if shelly backup create --all --dir "${BACKUP_DIR}" -q 2>/dev/null; then
        log "Backups created successfully"
    else
        error "Backup failed! Use --skip-backup to proceed anyway"
        exit 1
    fi
fi

# Perform updates
log "Starting firmware updates..."
success_count=0
fail_count=0

while IFS= read -r device; do
    [[ -z "${device}" ]] && continue

    log "Updating ${device}..."

    if shelly firmware update "${device}" --yes -q 2>/dev/null; then
        log "  ${device} updated successfully"
        ((success_count++)) || true
    else
        error "  Failed to update ${device}"
        ((fail_count++)) || true
    fi

    # Wait between updates if sequential
    if ${SEQUENTIAL}; then
        log "  Waiting for device to reboot..."
        sleep 30

        # Wait for device to come back online
        if shelly wait "${device}" --online --timeout 120s -q 2>/dev/null; then
            log "  ${device} is back online"
        else
            error "  ${device} did not come back online within timeout"
        fi
    fi
done <<< "${devices_with_updates}"

printf '\n'
log "Update complete!"
log "  Successful: ${success_count}"
log "  Failed: ${fail_count}"

if [[ ${fail_count} -gt 0 ]]; then
    error "Some updates failed. Backups available in: ${BACKUP_DIR}"
    exit 1
fi
```

---

## presence-detect.sh

presence-detect.sh - Network presence-based home automation

```bash
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
```

---

## weather-automation.sh

weather-automation.sh - Weather-based device automation

```bash
#!/usr/bin/env bash
#
# weather-automation.sh - Weather-based device automation
#
# Adjusts devices based on current weather conditions using wttr.in API.
# Examples:
# - Close blinds/covers when sunny and hot
# - Turn on lights when cloudy/overcast
# - Reduce AC when temperature drops
#
# This integrates external weather data with Shelly - something the
# devices and app cannot do on their own.
#
# Usage: ./weather-automation.sh [--daemon] [--location CITY]
#
# Requires: curl, jq

set -euo pipefail

# Configuration - customize for your setup
LOCATION="${LOCATION:-}"  # Auto-detect if empty
CHECK_INTERVAL=1800  # 30 minutes

# Device mappings
COVER_DEVICE="living-room-blinds"   # Cover/roller for sun protection
LIGHT_DEVICE="living-room-light"    # Light for cloudy days
AC_DEVICE="ac-outlet"               # Smart plug for AC

# Thresholds
HOT_TEMP=28          # Celsius - close blinds above this + sunny
CLOUDY_LUX=50        # Close blinds when cloud cover % is below this (sunny)
OVERCAST_LUX=70      # Turn on lights when cloud cover % is above this

log() {
    printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

# Get weather data from wttr.in
get_weather() {
    local location="${1:-}"
    local url="https://wttr.in/${location}?format=j1"

    curl -s "${url}" 2>/dev/null
}

# Extract values from weather JSON
parse_weather() {
    local json="${1}"

    temp=$(printf '%s' "${json}" | jq -r '.current_condition[0].temp_C' 2>/dev/null || printf 'N/A')
    feels_like=$(printf '%s' "${json}" | jq -r '.current_condition[0].FeelsLikeC' 2>/dev/null || printf 'N/A')
    cloud_cover=$(printf '%s' "${json}" | jq -r '.current_condition[0].cloudcover' 2>/dev/null || printf '50')
    uv_index=$(printf '%s' "${json}" | jq -r '.current_condition[0].uvIndex' 2>/dev/null || printf '0')
    description=$(printf '%s' "${json}" | jq -r '.current_condition[0].weatherDesc[0].value' 2>/dev/null || printf 'Unknown')
    location_name=$(printf '%s' "${json}" | jq -r '.nearest_area[0].areaName[0].value' 2>/dev/null || printf 'Unknown')
}

# Apply automation rules based on weather
apply_rules() {
    log "Current weather in ${location_name}:"
    log "  Temperature: ${temp}°C (feels like ${feels_like}°C)"
    log "  Conditions: ${description}"
    log "  Cloud cover: ${cloud_cover}%"
    log "  UV index: ${uv_index}"

    local hour
    hour=$(date '+%H')

    # Only apply rules during daytime (7 AM - 9 PM)
    if [[ "${hour}" -lt 7 || "${hour}" -gt 21 ]]; then
        log "Outside daytime hours, skipping automation"
        return
    fi

    # Rule 1: Hot and sunny -> close blinds
    if [[ "${temp}" -gt "${HOT_TEMP}" && "${cloud_cover}" -lt "${CLOUDY_LUX}" ]]; then
        log "Hot and sunny - closing blinds for sun protection"
        shelly cover close "${COVER_DEVICE}" -q 2>/dev/null || log "  Warning: Could not close blinds"
    fi

    # Rule 2: Overcast/dark -> turn on lights
    if [[ "${cloud_cover}" -gt "${OVERCAST_LUX}" ]]; then
        log "Overcast conditions - turning on lights"
        shelly on "${LIGHT_DEVICE}" -q 2>/dev/null || log "  Warning: Could not turn on lights"
    else
        # Clear skies during day - natural light is enough
        log "Clear skies - turning off supplemental lights"
        shelly off "${LIGHT_DEVICE}" -q 2>/dev/null || true
    fi

    # Rule 3: Temperature dropped significantly -> reduce AC
    if [[ "${temp}" -lt 22 ]]; then
        log "Temperature comfortable - reducing AC"
        shelly off "${AC_DEVICE}" -q 2>/dev/null || true
    fi

    log "Automation rules applied"
}

# Single check
single_check() {
    log "Fetching weather data..."

    local weather_json
    weather_json=$(get_weather "${LOCATION}")

    if [[ -z "${weather_json}" || "${weather_json}" == "null" ]]; then
        log "Error: Could not fetch weather data"
        exit 1
    fi

    parse_weather "${weather_json}"
    apply_rules
}

# Daemon mode - continuous monitoring
daemon_mode() {
    log "Starting weather automation daemon (Ctrl+C to stop)"
    log "Checking every $((CHECK_INTERVAL / 60)) minutes"

    while true; do
        single_check
        printf '\n'
        sleep "${CHECK_INTERVAL}"
    done
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "${1}" in
        --daemon)
            daemon_mode
            exit 0
            ;;
        --location)
            LOCATION="${2}"
            shift 2
            ;;
        *)
            printf 'Unknown option: %s\n' "${1}"
            exit 1
            ;;
    esac
done

# Default: single check
single_check
```

---

## workstation-sync.sh

workstation-sync.sh - Sync desk lights with computer activity

```bash
#!/usr/bin/env bash
#
# workstation-sync.sh - Sync desk lights with computer activity
#
# Automatically adjusts desk/office lights based on your computer's state:
# - Screen unlocks -> lights on at appropriate brightness for time of day
# - Screen locks -> lights dim/off
# - Meeting starts (calendar) -> bias lighting mode
#
# This integrates your computer's state with Shelly - something the app can't do.
#
# macOS: Uses system events
# Linux: Uses dbus/logind
#
# Usage: ./workstation-sync.sh [--install]
#
# --install: Set up as a launch agent (macOS) or systemd service (Linux)

set -euo pipefail

# Configuration
DESK_LIGHT="desk-lamp"
MONITOR_BACKLIGHT="monitor-bias"  # Optional LED strip behind monitor
OFFICE_MAIN="office-ceiling"

# Brightness levels by time of day
get_brightness() {
    local hour
    hour=$(date '+%H')

    if [[ "${hour}" -lt 7 ]]; then
        printf '20'    # Early morning: dim
    elif [[ "${hour}" -lt 9 ]]; then
        printf '60'    # Morning: medium
    elif [[ "${hour}" -lt 17 ]]; then
        printf '100'   # Daytime: full
    elif [[ "${hour}" -lt 20 ]]; then
        printf '70'    # Evening: slightly reduced
    else
        printf '40'    # Night: dim
    fi
}

log() {
    printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

on_unlock() {
    local brightness
    brightness=$(get_brightness)

    log "Screen unlocked - setting up workspace (brightness: ${brightness}%)"

    # Turn on desk light with appropriate brightness
    if shelly light set "${DESK_LIGHT}" --brightness "${brightness}" -q 2>/dev/null; then
        log "  Desk light: ${brightness}%"
    else
        shelly on "${DESK_LIGHT}" -q 2>/dev/null || true
    fi

    # Optional: Turn on bias lighting
    shelly on "${MONITOR_BACKLIGHT}" -q 2>/dev/null || true
}

on_lock() {
    log "Screen locked - reducing lights"

    # Dim desk light significantly or turn off
    if shelly light set "${DESK_LIGHT}" --brightness 10 -q 2>/dev/null; then
        log "  Desk light: dimmed to 10%"
    else
        shelly off "${DESK_LIGHT}" -q 2>/dev/null || true
    fi

    # Turn off bias lighting
    shelly off "${MONITOR_BACKLIGHT}" -q 2>/dev/null || true
}

# macOS: Monitor screen lock/unlock events
monitor_macos() {
    log "Monitoring screen lock events (macOS)..."
    log "Press Ctrl+C to stop"

    # Use ioreg to detect screen state changes
    local prev_state="unknown"

    while true; do
        # Check if screen is locked
        if /usr/libexec/PlistBuddy -c "Print :IOConsoleUsers:0:CGSSessionScreenIsLocked" /dev/stdin 2>/dev/null <<< "$(ioreg -n Root -d1 -a)" | grep -q "true"; then
            if [[ "${prev_state}" != "locked" ]]; then
                on_lock
                prev_state="locked"
            fi
        else
            if [[ "${prev_state}" != "unlocked" ]]; then
                on_unlock
                prev_state="unlocked"
            fi
        fi
        sleep 2
    done
}

# Linux: Monitor screen lock via dbus
monitor_linux() {
    log "Monitoring screen lock events (Linux)..."
    log "Press Ctrl+C to stop"

    # Use gdbus to monitor screensaver signals
    gdbus monitor --session --dest org.gnome.ScreenSaver --object-path /org/gnome/ScreenSaver 2>/dev/null | while read -r line; do
        if [[ "${line}" == *"ActiveChanged"*"true"* ]]; then
            on_lock
        elif [[ "${line}" == *"ActiveChanged"*"false"* ]]; then
            on_unlock
        fi
    done
}

# Detect OS and run appropriate monitor
case "$(uname)" in
    Darwin)
        monitor_macos
        ;;
    Linux)
        monitor_linux
        ;;
    *)
        log "Unsupported OS: $(uname)"
        exit 1
        ;;
esac
```

---

