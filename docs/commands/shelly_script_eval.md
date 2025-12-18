## shelly script eval

Evaluate JavaScript code

### Synopsis

Evaluate a JavaScript expression in the context of a running script.

The script must be running for eval to work. The code argument can be
multiple words that will be joined together.

```
shelly script eval <device> <id> <code> [flags]
```

### Examples

```
  # Evaluate a simple expression
  shelly script eval living-room 1 "1 + 2"

  # Print a message
  shelly script eval living-room 1 "print('Hello from CLI!')"

  # Call a function defined in the script
  shelly script eval living-room 1 "myFunction()"
```

### Options

```
  -h, --help   help for eval
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

* [shelly script](shelly_script.md)	 - Manage device scripts

