## shelly alias

Manage command aliases

### Synopsis

Create, list, and manage command aliases.

Aliases allow you to create shortcuts for frequently used commands.
They support argument interpolation with $1, $2, etc., and $@ for all arguments.

Shell command aliases are prefixed with ! and execute in your shell.

### Examples

```
  # List all aliases
  shelly alias list

  # Create a simple alias
  shelly alias set lights "batch on living-room kitchen bedroom"

  # Create an alias with arguments
  shelly alias set sw "switch $1 $2"
  # Usage: shelly sw on kitchen

  # Create a shell alias
  shelly alias set backup '!tar -czf shelly-backup.tar.gz ~/.config/shelly'

  # Delete an alias
  shelly alias delete lights

  # Export aliases to file
  shelly alias export aliases.yaml

  # Import aliases from file
  shelly alias import aliases.yaml
```

### Options

```
  -h, --help   help for alias
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly alias delete](shelly_alias_delete.md)	 - Delete a alias
* [shelly alias export](shelly_alias_export.md)	 - Export aliases to a YAML file
* [shelly alias import](shelly_alias_import.md)	 - Import aliases from a YAML file
* [shelly alias list](shelly_alias_list.md)	 - List aliass
* [shelly alias set](shelly_alias_set.md)	 - Create or update a command alias

