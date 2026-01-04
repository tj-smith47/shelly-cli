## shelly batch

Execute commands on multiple devices

### Synopsis

Execute commands on multiple devices simultaneously.

Batch operations can target:
- A device group (created with 'shelly group create')
- Multiple devices specified by name or IP
- All registered devices

Failed operations are reported but don't stop the batch.

### Examples

```
  # Turn on all devices in a group
  shelly batch on --group living-room

  # Turn off specific devices
  shelly batch off light-1 light-2 switch-1

  # Toggle all devices
  shelly batch toggle --all

  # Send custom RPC command to a group
  shelly batch command --group office "Switch.Set" '{"id":0,"on":true}'
```

### Options

```
  -h, --help   help for batch
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly batch command](shelly_batch_command.md)	 - Send RPC command to devices
* [shelly batch off](shelly_batch_off.md)	 - Turn off switches
* [shelly batch on](shelly_batch_on.md)	 - Turn on switches
* [shelly batch toggle](shelly_batch_toggle.md)	 - Toggle switches

