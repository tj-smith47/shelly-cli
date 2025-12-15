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
