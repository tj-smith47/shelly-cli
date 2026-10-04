## shelly scene add-action

Add a device action to a scene

### Synopsis

Add a device action to a scene.

An action is one RPC call that 'shelly scene activate' sends to a device.
The method is the RPC method name, and params is its parameters as a JSON
object. Actions run in the order they were added.

The device does not have to be reachable when the action is added: nothing is
sent until the scene is activated.

```
shelly scene add-action <scene> <device> <method> [params] [flags]
```

### Examples

```
  # Turn a switch on when the scene is activated
  shelly scene add-action movie-night living-room Switch.Set '{"id":0,"on":true}'

  # Dim a light
  shelly scene add-action movie-night lamp Light.Set '{"id":0,"on":true,"brightness":20}'

  # A method that takes no parameters
  shelly scene add-action bedtime hallway Shelly.Reboot

  # Using alias
  shelly scene add movie-night tv-plug Switch.Set '{"id":0,"on":false}'
```

### Options

```
  -h, --help   help for add-action
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

* [shelly scene](shelly_scene.md)	 - Manage device scenes

