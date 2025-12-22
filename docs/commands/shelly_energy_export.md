## shelly energy export

Export energy data to file

### Synopsis

Export historical energy consumption data to a file.

Supports multiple output formats:
  - CSV: Comma-separated values (default)
  - JSON: Structured JSON format
  - YAML: Human-readable YAML format

The exported data includes timestamp, voltage, current, power, and energy
measurements for the specified time range.

```
shelly energy export <device> [id] [flags]
```

### Examples

```
  # Export last 24 hours as CSV
  shelly energy export shelly-3em-pro > data.csv

  # Export specific time range as JSON
  shelly energy export shelly-em --format json --from "2025-01-01" --to "2025-01-07" --output energy.json

  # Export last week as YAML
  shelly energy export shelly-3em-pro 0 --format yaml --period week --output weekly.yaml
```

### Options

```
  -f, --format string   Output format (csv, json, yaml) (default "csv")
      --from string     Start time (RFC3339 or YYYY-MM-DD)
  -h, --help            help for export
  -o, --output string   Output file (default: stdout)
  -p, --period string   Time period (hour, day, week, month)
      --to string       End time (RFC3339 or YYYY-MM-DD)
      --type string     Component type (auto, em, em1) (default "auto")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

