## shelly cloud login

Authenticate with Shelly Cloud

### Synopsis

Authenticate with the Shelly Cloud API.

This command authenticates you with the Shelly Cloud service using your
email and password. The access token is stored locally for future use.

You can provide credentials via:
  1. Command flags (--email, --password)
  2. Interactive prompts (if TTY available)
  3. Environment variables (SHELLY_CLOUD_EMAIL, SHELLY_CLOUD_PASSWORD)

```
shelly cloud login [flags]
```

### Examples

```
  # Interactive login
  shelly cloud login

  # Login with flags
  shelly cloud login --email user@example.com --password mypassword

  # Login with environment variables
  SHELLY_CLOUD_EMAIL=user@example.com SHELLY_CLOUD_PASSWORD=mypassword shelly cloud login
```

### Options

```
      --email string      Shelly Cloud email
  -h, --help              help for login
      --password string   Shelly Cloud password
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

