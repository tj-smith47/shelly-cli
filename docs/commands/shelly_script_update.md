## shelly script update

Update a script

### Synopsis

Update an existing script on a Gen2+ Shelly device.

You can update the script name, code, or enabled status.
Use --append to add code to the existing script instead of replacing it.

```
shelly script update <device> <id> [flags]
```

### Examples

```
  # Update script name
  shelly script update living-room 1 --name "New Name"

  # Update script code
  shelly script update living-room 1 --file script.js

  # Append code to existing script
  shelly script update living-room 1 --code "// More code" --append

  # Enable/disable script
  shelly script update living-room 1 --enable
```

### Options

```
      --append        Append code instead of replacing
      --code string   Script code (inline)
      --enable        Enable the script
  -f, --file string   Script code file
  -h, --help          help for update
      --name string   Script name
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly script](shelly_script.md)	 - Manage device scripts

