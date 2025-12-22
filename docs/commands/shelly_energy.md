## shelly energy

Energy monitoring operations (EM/EM1 components)

### Synopsis

Manage and monitor professional energy monitoring components.

This command works with EM (3-phase) and EM1 (single-phase) energy monitor
components found on professional Shelly devices like:
  - Shelly Pro 3EM (3-phase energy monitor)
  - Shelly Pro EM (single-phase or dual-phase monitor)
  - Shelly Pro EM-50 (professional energy monitor)

These components provide real-time measurements including:
  - Voltage, current, power (active/apparent)
  - Power factor and frequency
  - Per-phase data for 3-phase monitors
  - Total power and neutral current

For power meters with energy totals (PM/PM1 components), use 'shelly power'.

### Examples

```
  # List energy monitor components
  shelly energy list kitchen

  # Get current energy status
  shelly em status pro3em

  # View energy history
  shelly energy history kitchen --period day
```

### Options

```
  -h, --help   help for energy
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly energy compare](shelly_energy_compare.md)	 - Compare energy usage between devices
* [shelly energy dashboard](shelly_energy_dashboard.md)	 - Show energy dashboard for all devices
* [shelly energy export](shelly_energy_export.md)	 - Export energy data to file
* [shelly energy history](shelly_energy_history.md)	 - Show energy consumption history
* [shelly energy list](shelly_energy_list.md)	 - List energy monitoring components
* [shelly energy reset](shelly_energy_reset.md)	 - Reset energy monitor counters
* [shelly energy status](shelly_energy_status.md)	 - Show energy monitor status

