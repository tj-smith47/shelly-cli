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
