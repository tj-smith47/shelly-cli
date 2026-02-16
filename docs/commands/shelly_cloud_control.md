## shelly cloud control

Control a device via cloud

### Synopsis

Control a Shelly device through the Shelly Cloud API.

Supported actions:
  switch:  on, off, toggle
  cover:   open, close, stop, position=<0-100>
  light:   on, off, toggle, brightness=<0-100>

This command requires authentication with 'shelly cloud login'.

```
shelly cloud control <device-id> <action> [flags]
```

### Examples

```
  # Turn on a switch
  shelly cloud control abc123 on

  # Turn off switch on channel 1
  shelly cloud control abc123 off --channel 1

  # Toggle a switch
  shelly cloud control abc123 toggle

  # Set cover to 50%
  shelly cloud control abc123 position=50

  # Open cover
  shelly cloud control abc123 open

  # Set light brightness
  shelly cloud control abc123 brightness=75
```

### Options

```
      --channel int   Device channel/relay number
  -h, --help          help for control
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

