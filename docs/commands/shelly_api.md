## shelly api

Execute API calls on Shelly devices

### Synopsis

Execute API calls on Shelly devices.

This command provides direct access to the device API using either:
  - RPC methods: "Shelly.GetDeviceInfo", "Switch.Set" (Gen2+ only)
  - REST paths: "/status", "/relay/0?turn=on", "/rpc/Shelly.GetStatus" (all generations)

The command auto-detects the call type based on input format:
  - Starts with "/" → HTTP GET to that path (works for all generations)
  - Contains "." → JSON-RPC call via WebSocket (Gen2+ only)

For Gen2+ devices, you can use either format:
  - RPC: Shelly.GetStatus (uses WebSocket JSON-RPC)
  - Path: /rpc/Shelly.GetStatus (uses HTTP GET)

Use 'shelly api methods <device>' to list available RPC methods (Gen2+ only).

```
shelly api <device> <method|path> [params_json] [flags]
```

### Examples

```
  # Gen2+ RPC methods (JSON-RPC via WebSocket)
  shelly api living-room Shelly.GetDeviceInfo
  shelly api living-room Switch.GetStatus '{"id":0}'
  shelly api living-room Switch.Set '{"id":0,"on":true}'

  # Path-based HTTP calls (all generations)
  shelly api backyard /status                      # Gen1 status
  shelly api backyard /relay/0?turn=on             # Gen1 relay control
  shelly api living-room /rpc/Shelly.GetStatus     # Gen2+ via HTTP
  shelly api living-room /rpc/Switch.Set?id=0&on=true  # Gen2+ via HTTP

  # Raw output (no formatting)
  shelly api living-room Shelly.GetStatus --raw

  # Using 'rpc' alias
  shelly rpc living-room Shelly.GetDeviceInfo
```

### Options

```
  -h, --help   help for api
      --raw    Output raw JSON without formatting
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
* [shelly api methods](shelly_api_methods.md)	 - List available RPC methods (Gen2+ only)

