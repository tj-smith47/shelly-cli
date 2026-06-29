## shelly group on

Turn on all group members

### Synopsis

Turn on every device in a group.

The action is fanned out to all members concurrently and a per-member result
summary is printed. Works across mixed Gen1 and Gen2+ members. Omit --id to
control every controllable component on each member.

```
shelly group on <group> [flags]
```

### Examples

```
  # Turn on every member of a group
  shelly group on guest-bath-bulbs

  # Turn on only component 1 on each member
  shelly group on living-room --id 1
```

### Options

```
  -c, --concurrent int   Max concurrent operations (default 5)
  -h, --help             help for on
      --id int           Component ID to control (omit to control all) (default -1)
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

* [shelly group](shelly_group.md)	 - Manage device groups

