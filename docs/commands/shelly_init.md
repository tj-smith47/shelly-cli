## shelly init

Initialize shelly CLI for first-time use

### Synopsis

Initialize the shelly CLI with a guided setup wizard.

This command helps new users get started by:
  - Creating a configuration file with sensible defaults
  - Discovering Shelly devices on your network
  - Registering discovered devices with friendly names
  - Installing shell completions for tab completion
  - Optionally setting up Shelly Cloud access

INTERACTIVE MODE (default):
  Run without flags to use the interactive wizard.

NON-INTERACTIVE MODE:
  Automatically enabled when any device or config flags are provided.
  Discovery, completions, aliases, and cloud are opt-in via flags.

Use --check to verify your current setup without making changes.

```
shelly init [flags]
```

### Examples

```
  # Interactive setup wizard
  shelly init

  # Check current setup without changes
  shelly init --check

  # Headless: register devices directly
  shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

  # Headless: device with authentication
  shelly init --device secure=192.168.1.102:admin:secret

  # Headless: import from JSON file
  shelly init --devices-json devices.json

  # Headless: inline JSON
  shelly init --devices-json '{"name":"kitchen","address":"192.168.1.100"}'

  # Headless: with discovery and completions
  shelly init --discover --discover-modes mdns,http --completions bash,zsh

  # Headless: full CI/CD setup
  shelly init \
    --device kitchen=192.168.1.100 \
    --theme dracula \
    --api-mode local \
    --no-color

  # Headless: with cloud credentials
  shelly init --cloud-email user@example.com --cloud-password secret
```

### Options

```
      --aliases                     Install default command aliases (opt-in)
      --api-mode string             API mode: local,cloud,auto (default: local)
      --check                       Verify current setup without making changes
      --cloud-email string          Shelly Cloud email (enables cloud setup)
      --cloud-password string       Shelly Cloud password (enables cloud setup)
      --completions string          Install completions for shells: bash,zsh,fish,powershell (comma-separated)
      --device stringArray          Device spec: name=ip[:user:pass] (repeatable)
      --devices-json stringArray    JSON device(s): file path, array, or single object (repeatable)
      --discover                    Enable device discovery (opt-in in non-interactive mode)
      --discover-modes string       Discovery modes: mdns,coiot,http,ble,all (comma-separated) (default "mdns")
      --discover-timeout duration   Discovery timeout (default 15s)
      --force                       Overwrite existing configuration
  -h, --help                        help for init
      --network string              Subnet for HTTP probe discovery (e.g., 192.168.1.0/24)
      --no-color                    Disable colors in output
      --output-format string        Set output format: table,json,yaml (default: table)
      --theme string                Set theme (default: dracula)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

