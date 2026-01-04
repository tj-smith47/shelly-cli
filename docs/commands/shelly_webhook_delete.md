## shelly webhook delete

Delete a webhook

### Synopsis

Delete a webhook by ID.

Use 'shelly webhook list' to see webhook IDs.

```
shelly webhook delete <device> <webhook-id> [flags]
```

### Examples

```
  # Delete a webhook
  shelly webhook delete <device> 1

  # Delete without confirmation
  shelly webhook delete <device> 1 --yes
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly webhook](shelly_webhook.md)	 - Manage device webhooks

