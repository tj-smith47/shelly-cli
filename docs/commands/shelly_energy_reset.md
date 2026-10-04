## shelly energy reset

Reset energy monitor counters

### Synopsis

Reset the accumulated energy counters of a component that meters power.

Works on EM (3-phase) energy monitors, PM and PM1 power meters, and
switches, covers and lights that meter their load (Plus 1PM, Plus 2PM,
Plug, dimmers, RGBW PM). The component is chosen as 'shelly energy
status' chooses it: the first one that meters power, or the one with
the given ID and --type.

EM1 energy monitors and Gen1 meters have no counter reset.

```
shelly energy reset <device> [id] [flags]
```

### Examples

```
  # Reset all counters for EM component 0
  shelly energy reset shelly-3em-pro 0

  # Reset specific counter types
  shelly energy reset shelly-3em-pro 0 --types active,reactive

  # Reset the energy total of switch channel 1 on a Plus 2PM
  shelly energy reset kitchen 1 --type switch

  # Reset with device alias
  shelly energy reset basement-em
```

### Options

```
  -h, --help            help for reset
      --type string     Component type: auto, or one of em, em1, pm, pm1, switch, cover, light, rgb, rgbw, cct, meter (Gen1), emeter (Gen1) (default "auto")
      --types strings   Counter types to reset (leave empty for all)
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

