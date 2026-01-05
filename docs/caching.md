# Caching

The Shelly CLI includes a file-based caching system that improves response times and enables offline operation.

## Cache Location

Cache files are stored in the standard user cache directory:

| Platform | Location |
|----------|----------|
| Linux | `~/.cache/shelly/` |
| macOS | `~/Library/Caches/shelly/` |
| Windows | `%LocalAppData%\shelly\cache\` |

Cache entries are organized by data type and device:
```
~/.cache/shelly/
├── firmware/
│   ├── kitchen.json
│   └── living-room.json
├── automation/
│   └── schedules/
│       └── kitchen.json
└── meta.json
```

## TTL Values

Different data types have different time-to-live (TTL) values based on how frequently the data changes:

| Data Type | TTL | Description |
|-----------|-----|-------------|
| Device Info | 24 hours | Hardware info (model, MAC address) |
| Components | 24 hours | Component list (switches, lights, etc.) |
| System | 1 hour | System settings |
| Firmware | 1 hour | Firmware version and updates |
| Security | 1 hour | Authentication settings |
| BLE | 1 hour | Bluetooth settings |
| Protocols (MQTT, Modbus) | 1 hour | Protocol configurations |
| WiFi | 30 minutes | WiFi status and configuration |
| Cloud | 30 minutes | Cloud connection status |
| Smart Home (Matter, Zigbee) | 30 minutes | Smart home protocol status |
| Inputs | 10 minutes | Input component status |
| Automation | 5 minutes | Schedules, webhooks, virtuals, KVS, scripts |

## Global Flags

Two global flags control cache behavior:

### `--refresh`

Bypass the cache and fetch fresh data directly from the device.

```bash
# Force fresh data fetch
shelly device info kitchen --refresh

# Check current firmware without using cache
shelly firmware check garage --refresh
```

### `--offline`

Only read from cache, never contact the device. Returns an error if no cached data exists.

```bash
# View cached device info (fast, no network)
shelly device info kitchen --offline

# List schedules from cache
shelly schedule list kitchen --offline
```

These flags are mutually exclusive - using both together produces an error.

## Cache Commands

### `shelly cache show`

Display cache statistics including total entries, size, device count, and entries by type.

```bash
# Show cache statistics
shelly cache show

# Output in JSON format
shelly cache show -o json
```

**Aliases:** `s`, `stats`, `status`

### `shelly cache clear`

Clear cached data with various scopes.

```bash
# Clear all cache (requires confirmation)
shelly cache clear --all

# Clear without confirmation
shelly cache clear --all --yes

# Clear cache for a specific device
shelly cache clear --device kitchen

# Clear specific data type for a device
shelly cache clear --device kitchen --type firmware

# Clear only expired entries
shelly cache clear --expired
```

**Aliases:** `c`, `rm`, `clean`

**Flags:**
- `--all, -a`: Clear all cached data
- `--device, -d`: Clear cache for a specific device
- `--type, -t`: Clear specific data type (requires `--device`)
- `--expired, -e`: Clear only expired entries
- `--yes, -y`: Skip confirmation prompt

## Cache Invalidation

The cache is automatically invalidated when you perform mutations:

| Command | Invalidates |
|---------|-------------|
| `firmware update` | Device info, firmware |
| `schedule create/update/delete` | Schedules |
| `webhook create/update/delete` | Webhooks |
| `virtual create/delete` | Virtuals |
| `kvs set/delete` | KVS |
| `script create/delete/put` | Scripts |
| `wifi set` | WiFi |
| `mqtt set/disable` | MQTT |
| `cloud enable/disable` | Cloud |

## Verbose Mode

In verbose mode (`-v`), cache operations are logged:

```bash
shelly device info kitchen -v
# Output includes: "cache hit for kitchen/deviceinfo (cached 5m ago)"
```

Cache hits show:
- The device and data type served from cache
- How long ago the data was cached

## Implementation Details

- **Atomic writes**: Cache files are written atomically using temp files and rename
- **JSON format**: Entries are stored as human-readable JSON with metadata
- **Version tracking**: Cache format includes a version number for future migrations
- **Automatic cleanup**: Expired entries can be cleaned up with `cache clear --expired`

### Cache Entry Format

Each cache entry contains:
```json
{
  "version": 1,
  "device": "kitchen",
  "device_id": "shellyplus1-a1b2c3d4e5f6",
  "data_type": "firmware",
  "cached_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-15T11:30:00Z",
  "data": { ... }
}
```

## Disabling Cache

To always fetch fresh data, use `--refresh` with your commands. There is no global setting to disable caching, as the cache provides offline capability.

For scripts that need guaranteed fresh data:
```bash
#!/bin/bash
shelly device info "$1" --refresh
```
