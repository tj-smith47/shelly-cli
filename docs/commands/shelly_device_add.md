## shelly device add

Add a device to the registry

### Synopsis

Add a Shelly device to the local registry.

The device will be verified and its generation/model auto-detected
unless --no-verify is specified.

The name is used as a friendly identifier for the device in other commands.
Names with spaces will be normalized to dashes (e.g., "Master Bathroom"
becomes "master-bathroom" as the key).

Credentials for a password-protected device are stored with the device and
used by every command run after it. Give them as --password (the user
defaults to admin, the user Shelly devices use), as --password-stdin to keep
the password out of the shell history, or as --auth user:pass. When the device
has authentication enabled the credentials are checked against it first, and
a device that rejects them is not added. --no-verify skips that check.

A name that is already registered is refused unless --force is given, which
replaces the registration (address, credentials, platform, model). Aliases
of the replaced device are kept.

--platform registers a device managed by a plugin, such as tasmota for the
shelly-tasmota plugin. The plugin's detect hook verifies the device instead
of the Shelly API.

```
shelly device add <name> <address> [flags]
```

### Examples

```
  # Add a device (auto-detects generation and model)
  shelly device add kitchen 192.168.1.100

  # Add a password-protected device
  shelly device add kitchen 192.168.1.100 --user admin --password secret

  # Read the password from stdin (prompts without echo on a terminal)
  shelly device add kitchen 192.168.1.100 --password-stdin < ~/.shelly-kitchen-password

  # Replace an existing registration with a new address or password
  shelly device add kitchen 192.168.1.110 --force --password-stdin

  # Add a Tasmota device through the shelly-tasmota plugin
  shelly device add garage-plug 192.168.1.50 --platform tasmota

  # Add without verification (offline)
  shelly device add offline-device 192.168.1.102 --no-verify --generation 2

  # Short form
  shelly dev add bedroom 192.168.1.103
```

### Options

```
      --auth string       Authentication credentials (user:pass)
  -f, --force             Replace an existing device of the same name
  -g, --generation int    Device generation (auto-detected if omitted)
  -h, --help              help for add
      --no-verify         Skip the connectivity check, auto-detection and the credentials check
      --password string   Device password
      --password-stdin    Read the device password from stdin
      --platform string   Device platform managed by a plugin (e.g. tasmota); default shelly
      --user string       Username for --password or --password-stdin (default admin)
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

