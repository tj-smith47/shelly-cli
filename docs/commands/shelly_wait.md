## shelly wait

Wait until a device is online or its output is on or off

### Synopsis

Wait until a device responds, or until its output is on or off, then exit.

The device is checked every --interval until the condition holds or --timeout
passes. The exit code is zero once the condition holds and non-zero on
timeout, so the command can gate the next step of a script.

Without --state the command waits for the device to answer. That is useful
after anything that takes a device off the network for a while:
  - A firmware update
  - A reboot or factory reset
  - A power cycle or a WiFi change

With --state on or --state off the command waits until every switch, light
and RGB output of the device is in that state (or only the one given by --id).
That is useful after a timer, a schedule, a scene or a physical button is
expected to change the output.

```
shelly wait <device> [flags]
```

### Examples

```
  # Continue once the device is back after a firmware update
  shelly firmware update kitchen --yes && shelly wait kitchen

  # Give a slow device up to five minutes
  shelly wait garage --timeout 5m

  # Check more often
  shelly wait 192.168.1.100 --interval 500ms

  # Use in a script without output
  if shelly wait kitchen --online --timeout 120s -q; then
    shelly on kitchen
  fi

  # Continue once the light has switched off
  shelly wait hallway --state off --timeout 60s

  # Watch one output of a multi-channel device
  shelly wait bathroom --state on --id 1
```

### Options

```
  -h, --help                help for wait
      --id int              Component ID to watch with --state (omit to watch all) (default -1)
      --interval duration   Time between checks (default 2s)
      --online              Wait until the device responds (default true)
      --state string        Wait until the device output is in this state: on, off
      --timeout duration    How long to wait before giving up (default 2m0s)
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

