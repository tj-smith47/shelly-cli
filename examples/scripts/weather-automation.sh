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
