## shelly export ansible

Export devices as Ansible inventory

### Synopsis

Export devices as an Ansible inventory YAML file.

Creates an Ansible-compatible inventory with device groups based on
model type. Use @all to export all registered devices.

```
shelly export ansible <devices...> [file] [flags]
```

### Examples

```
  # Export to stdout
  shelly export ansible @all

  # Export to file
  shelly export ansible @all inventory.yaml

  # Export specific devices
  shelly export ansible living-room bedroom inventory.yaml

  # Specify group name
  shelly export ansible @all --group-name shelly_devices
```

### Options

```
      --group-name string   Ansible group name for devices (default "shelly")
  -h, --help                help for ansible
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly export](shelly_export.md)	 - Export fleet data for infrastructure tools

