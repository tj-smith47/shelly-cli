## shelly cloud token

Show or manage cloud token

### Synopsis

Show the current Shelly Cloud access token.

This command displays the access token for debugging purposes.
Be careful not to share or expose your token.

```
shelly cloud token [flags]
```

### Examples

```
  # Show the current token
  shelly cloud token

  # Copy token to clipboard (Linux)
  shelly cloud token | xclip -selection clipboard
```

### Options

```
  -h, --help   help for token
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

