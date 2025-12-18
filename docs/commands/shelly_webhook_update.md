## shelly webhook update

Update a webhook

### Synopsis

Update an existing webhook.

Only specified fields will be updated. Use --enable or --disable to
change the webhook's active state.

```
shelly webhook update <device> <webhook-id> [flags]
```

### Examples

```
  # Change webhook URL
  shelly webhook update living-room 1 --url "http://new-url.com"

  # Disable a webhook
  shelly webhook update living-room 1 --disable

  # Enable and change event
  shelly webhook update living-room 1 --enable --event "switch.off"
```

### Options

```
      --disable           Disable webhook
      --enable            Enable webhook
      --event string      Event type
  -h, --help              help for update
      --name string       Webhook name
      --url stringArray   Webhook URL (replaces all URLs)
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

* [shelly webhook](shelly_webhook.md)	 - Manage device webhooks

