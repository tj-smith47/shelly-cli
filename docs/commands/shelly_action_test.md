## shelly action test

Test/trigger an action on a Gen1 device

### Synopsis

Test (trigger) an action on a Gen1 Shelly device.

This simulates the event that would trigger the action URL, causing
the device to make the configured HTTP request.

Gen1 devices trigger actions based on actual state changes. This command
will temporarily change the device state to trigger the action callback.

For output actions (out_on_url, out_off_url), the device relay will be toggled.
For button actions, the physical button press must be used.

Gen2+ devices use webhooks. See 'shelly webhook test'.

```
shelly action test <device> <event> [flags]
```

### Examples

```
  # Test output on action (turns relay on, triggering out_on_url)
  shelly action test living-room out_on_url

  # Test output off action
  shelly action test living-room out_off_url

  # Test action on specific relay
  shelly action test relay out_on_url --index 1
```

### Options

```
  -h, --help        help for test
      --index int   Action index (for multi-channel devices)
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

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs

