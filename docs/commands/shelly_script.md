## shelly script

Manage device scripts

### Synopsis

Manage JavaScript scripts on Gen2+ Shelly devices.

Scripts allow you to extend device functionality with custom JavaScript code.
Scripts can respond to events, automate actions, and interact with sensors.

Note: Scripts are only available on Gen2+ devices.

### Examples

```
  # List scripts on a device
  shelly script list living-room

  # Get script code
  shelly script get living-room 1

  # Create a new script
  shelly script create living-room --name "My Script" --file script.js

  # Start/stop a script
  shelly script start living-room 1
  shelly script stop living-room 1

  # Evaluate code on a running script
  shelly script eval living-room 1 "print('Hello!')"
```

### Options

```
  -h, --help   help for script
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
* [shelly script create](shelly_script_create.md)	 - Create a new script
* [shelly script delete](shelly_script_delete.md)	 - Delete a script
* [shelly script download](shelly_script_download.md)	 - Download script to file
* [shelly script eval](shelly_script_eval.md)	 - Evaluate JavaScript code
* [shelly script get](shelly_script_get.md)	 - Get script code or status
* [shelly script list](shelly_script_list.md)	 - List scripts on a device
* [shelly script start](shelly_script_start.md)	 - Start a script
* [shelly script stop](shelly_script_stop.md)	 - Stop a running script
* [shelly script template](shelly_script_template.md)	 - Manage script templates
* [shelly script update](shelly_script_update.md)	 - Update a script
* [shelly script upload](shelly_script_upload.md)	 - Upload script from file

