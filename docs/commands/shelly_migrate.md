## shelly migrate

Migrate configuration between devices

### Synopsis

Migrate configuration from one Shelly device to another.

Reads the current configuration from the source device and applies it to
the target device. By default, everything is migrated including network
and authentication settings.

When network settings are migrated, the source device is factory reset
after a successful migration to prevent IP conflicts on the network.
Use --skip-network to keep both devices online with their current
network settings, or --reset-source=false to skip the factory reset
(warning: this may cause IP conflicts).

Use --dry-run to preview what would change without applying.

```
shelly migrate <source-device> <target-device> [flags]
```

### Examples

```
  # Preview migration (dry run)
  shelly migrate living-room bedroom --dry-run

  # Full migration (factory resets source after)
  shelly migrate living-room bedroom --yes

  # Migrate without network config (no factory reset needed)
  shelly migrate living-room bedroom --skip-network

  # Migrate network but skip factory reset (may cause IP conflict)
  shelly migrate living-room bedroom --reset-source=false

  # Force migration between different device types
  shelly migrate living-room bedroom --force --yes

  # Clone config onto a new bulb with a distinct static IP (keeps both online,
  # source is not reset since there is no IP conflict)
  shelly migrate master-bath-1 new-bulb \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0

  # Clone a live sibling straight onto a brand-new device at its factory WiFi AP:
  # hops the host onto the AP, applies the config + static IP, the device joins
  # the LAN, and the source is left untouched (target name = "fr")
  shelly migrate sr fr --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1
```

### Options

```
      --allow-firmware-downgrade   Allow migrating a backup captured from newer firmware onto an older-firmware target (refused by default; this can trigger a reboot loop — prefer --update-firmware)
      --ap-ip string               Static host IP to use on the target's AP subnet during --to-ap (default 192.168.33.133)
      --dns string                 Static IPv4 nameserver (optional, with --static-ip)
      --dry-run                    Show what would be changed without applying
      --firmware-url string        Firmware image for --update-firmware (default: derived from the source device model)
      --force                      Force migration between different device types
      --gateway string             Static IPv4 default gateway (with --static-ip)
  -h, --help                       help for migrate
      --name string                Override the target device name (defaults to the target identifier when it is a friendly alias)
      --netmask string             Static IPv4 subnet mask (with --static-ip)
      --password string            WiFi passphrase for the target network (optional: derived from this host's stored credentials when omitted; set to override or when derivation fails)
      --reset-source               Factory reset source device after migration (default true)
      --skip-auth                  Skip authentication configuration
      --skip-meters                Skip migrating meter/energy-meter configuration (e.g. overpower limits)
      --skip-network               Skip network configuration (WiFi, Ethernet)
      --skip-schedules             Skip schedule migration
      --skip-scripts               Skip script migration
      --skip-state                 Skip migrating live component state (color temperature, brightness); apply configuration only
      --skip-webhooks              Skip webhook migration
      --ssid string                Override the WiFi SSID the target joins (defaults to the source's network)
      --static-ip string           Assign this static IPv4 to the target instead of copying the source's IP
      --to-ap string               Migrate onto a target at its factory WiFi AP with this SSID (hops host WiFi; source is never reset)
      --update-firmware            When the source runs newer firmware than the target, update the target to current stable firmware before migrating (Gen1; with --to-ap the update runs on the LAN after the target joins)
  -y, --yes                        Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly migrate diff](shelly_migrate_diff.md)	 - Show differences between device and backup
* [shelly migrate validate](shelly_migrate_validate.md)	 - Validate a backup file

