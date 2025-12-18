## shelly group add

Add devices to a group

### Synopsis

Add one or more devices to a group.

Devices can be specified by their registered name or IP address.
Devices can belong to multiple groups.

```
shelly group add <group> <device>... [flags]
```

### Examples

```
  # Add a single device to a group
  shelly group add living-room light-1

  # Add multiple devices
  shelly group add living-room light-1 light-2 switch-1

  # Add by IP address
  shelly group add office 192.168.1.100

  # Short form
  shelly grp add bedroom lamp
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly group](shelly_group.md)	 - Manage device groups

