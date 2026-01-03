## shelly webhook server

Start a local webhook receiver server

### Synopsis

Start a local HTTP server to receive and log webhooks from Shelly devices.

This is useful for testing and debugging webhook configurations. The server
logs all incoming requests with their headers, query parameters, and body.

The server will display its URL which can be used to configure device webhooks.

```
shelly webhook server [flags]
```

### Examples

```
  # Start server on default port 8080
  shelly webhook server

  # Start on a specific port
  shelly webhook server --port 9000

  # Start with JSON logging for piping
  shelly webhook server --log-json

  # Auto-configure devices to send webhooks here
  shelly webhook server --auto-config --device kitchen --device bedroom
```

### Options

```
      --auto-config        Auto-configure devices to use this server
      --device strings     Devices to auto-configure (with --auto-config)
  -h, --help               help for server
      --interface string   Network interface to bind to (default "0.0.0.0")
      --log-json           Log webhooks as JSON (for piping)
  -p, --port int           Port to listen on (default 8080)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
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

* [shelly webhook](shelly_webhook.md)	 - Manage device webhooks

