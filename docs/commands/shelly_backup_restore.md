## shelly backup restore

Restore a device from backup

### Synopsis

Restore a Shelly device from a backup file.

By default, everything from the backup is restored including network
and authentication settings. Use --skip-* flags to exclude specific
sections.

```
shelly backup restore <device> <file> [flags]
```

### Examples

```
  # Full restore from backup
  shelly backup restore living-room backup.json

  # Dry run - show what would change
  shelly backup restore living-room backup.json --dry-run

  # Restore without network config (keep current WiFi)
  shelly backup restore living-room backup.json --skip-network

  # Restore without auth config
  shelly backup restore living-room backup.json --skip-auth

  # Restore encrypted backup
  shelly backup restore living-room backup.json --decrypt mysecret

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts

  # Clone another bulb's backup onto this device with a different static IP
  # (identity — MAC, serial, device ID — is never overwritten by restore)
  shelly backup restore new-bulb master-bath-1.json \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # Restore a sibling's backup straight onto a brand-new device at its factory
  # WiFi AP: hops the host onto the AP, applies the config + static IP, and the
  # device joins the LAN — no separate provisioning step (target name = "fr")
  shelly backup restore fr sr.json --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # Same, but the target runs older firmware than the backup: update it first so
  # the full restore lands on matched firmware and cannot reboot-loop. With --to-ap
  # the update runs AT the factory AP (where the device is stable) — the image is
  # fetched before the hop and re-served on the AP subnet — then the full restore.
  shelly backup restore fr sr.json --to-ap ShellyBulbDuo-D0DCFF --update-firmware \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0
```

### Options

```
      --allow-firmware-downgrade   Allow restoring a backup captured from newer firmware onto an older-firmware device (refused by default; this can trigger a reboot loop — prefer --update-firmware)
      --ap-ip string               Static host IP to use on the device's AP subnet during --to-ap (default 192.168.33.133)
  -d, --decrypt string             Password to decrypt backup
      --dns string                 Static IPv4 nameserver (optional, with --static-ip)
      --dry-run                    Show what would be restored without applying
      --firmware-url string        Firmware image for --update-firmware (default: derived from the backup's device model)
      --gateway string             Static IPv4 default gateway (with --static-ip)
  -h, --help                       help for restore
      --name string                Override the device name (defaults to the target identifier when it is a friendly alias)
      --netmask string             Static IPv4 subnet mask (with --static-ip)
      --password string            WiFi passphrase for the target network (optional: derived from this host's stored credentials when omitted; set to override or when derivation fails)
      --skip-auth                  Skip authentication configuration
      --skip-meters                Skip restoring meter/energy-meter configuration (e.g. overpower limits)
      --skip-network               Skip network configuration (WiFi, Ethernet)
      --skip-schedules             Skip schedule restoration
      --skip-scripts               Skip script restoration
      --skip-state                 Skip restoring live component state (color temperature, brightness); apply configuration only
      --skip-webhooks              Skip webhook restoration
      --ssid string                Override the WiFi SSID the device joins (defaults to the backup's network)
      --static-ip string           Override the backup's WiFi with this static IPv4 address
      --to-ap string               Restore onto a device at its factory WiFi AP with this SSID (hops host WiFi; the network override moves it onto the LAN)
      --update-firmware            When the backup is from newer firmware than the target, update the device to current stable firmware before restoring (Gen1; with --to-ap the update runs at the factory AP before the device joins, since corrupt firmware reboot-loops once on the LAN)
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

* [shelly backup](shelly_backup.md)	 - Backup and restore device configurations

