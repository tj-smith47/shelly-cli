## shelly batch command

Send RPC command to devices

### Synopsis

Send a raw RPC command to multiple devices simultaneously.

The method is the RPC method name (e.g., "Switch.Set", "Shelly.GetStatus").
Params should be a JSON object (e.g., '{"id":0,"on":true}').

Target devices can be specified multiple ways:
  - As arguments: device names or addresses after the method/params
  - Via stdin: pipe device names (one per line or space-separated)
  - Via group: --group flag targets all devices in a group
  - Via all: --all flag targets all registered devices

Priority: explicit args > stdin > group > all

Results are output as JSON or YAML (use -o yaml). Each result includes
the device name and either the response or error message.

```
shelly batch command <method> [params-json] [device...] [flags]
```

### Examples

```
  # Get status from all devices in a group
  shelly batch command "Shelly.GetStatus" --group living-room

  # Turn on switch 0 on specific devices
  shelly batch command "Switch.Set" '{"id":0,"on":true}' light-1 light-2

  # Set brightness on all devices
  shelly batch command "Light.Set" '{"id":0,"brightness":50}' --all

  # Using alias
  shelly batch rpc "Switch.Toggle" '{"id":0}' --group bedroom

  # Output as YAML
  shelly batch command "Shelly.GetDeviceInfo" --all -o yaml

  # Pipe device names from a file
  cat devices.txt | shelly batch command "Shelly.GetStatus"

  # Pipe from device list command
  shelly device list -o json | jq -r '.[].name' | shelly batch command "Shelly.Reboot"

  # Get status of Gen2+ devices and extract uptime
  shelly device list -o json | jq -r '.[] | select(.generation >= 2) | .name' | \
    shelly batch command "Shelly.GetStatus" | jq '.[] | {device, uptime: .response.sys.uptime}'

  # Check firmware versions across all devices
  shelly batch command "Shelly.GetDeviceInfo" --all | jq '.[] | {device, fw: .response.fw_id}'
```

### Options

```
  -a, --all                Target all registered devices
  -c, --concurrent int     Max concurrent operations (default 5)
  -g, --group string       Target device group
  -h, --help               help for command
  -o, --output string      Output format: json, yaml (default "json")
  -t, --timeout duration   Timeout per device (default 10s)
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

* [shelly batch](shelly_batch.md)	 - Execute commands on multiple devices

