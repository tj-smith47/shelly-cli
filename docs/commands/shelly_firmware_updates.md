## shelly firmware updates

Interactive firmware update workflow

### Synopsis

Check for and apply firmware updates with an interactive workflow.

By default, runs in interactive mode - displays devices with available updates
and prompts for selection. Use --all with --yes for non-interactive batch updates.

Supports both native Shelly devices and plugin-managed devices (Tasmota, etc.).

```
shelly firmware updates [flags]
```

### Examples

```
  # Interactive mode - check and select devices
  shelly firmware updates

  # Update all devices to stable (non-interactive)
  shelly firmware updates --all --yes

  # Update all devices to beta (non-interactive)
  shelly firmware updates --all --beta --yes

  # Update specific devices to stable
  shelly firmware updates --devices=kitchen,bedroom --stable

  # Update only Tasmota devices
  shelly firmware updates --all --platform=tasmota --yes

  # Update specific devices interactively
  shelly firmware updates --devices=kitchen,bedroom
```

### Options

```
      --all               Update all devices with available updates
      --beta              Use beta/development release channel
      --devices string    Comma-separated list of specific devices to update
  -h, --help              help for updates
      --parallel int      Number of devices to update in parallel (default 3)
      --platform string   Only update devices of this platform (e.g., tasmota)
      --stable            Use stable release channel (default)
  -y, --yes               Skip confirmation prompt
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

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

