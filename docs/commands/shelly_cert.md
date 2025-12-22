## shelly cert

Manage device TLS certificates

### Synopsis

Manage TLS certificates for Gen2+ Shelly devices.

Devices support custom CA certificates for secure MQTT and cloud connections.
Use these commands to view or install certificates on devices.

### Examples

```
  # Show TLS configuration
  shelly cert show kitchen

  # Install a CA certificate
  shelly cert install kitchen --ca /path/to/ca.pem
```

### Options

```
  -h, --help   help for cert
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly cert install](shelly_cert_install.md)	 - Install a certificate on a device
* [shelly cert show](shelly_cert_show.md)	 - Show device TLS configuration

