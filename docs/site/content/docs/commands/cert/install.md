---
title: "shelly cert install"
description: "shelly cert install"
---

## shelly cert install

Install a certificate on a device

### Synopsis

Install a TLS certificate on a Gen2+ Shelly device.

Supports installing CA certificates for MQTT/cloud TLS verification,
as well as client certificates for mutual TLS authentication.

```
shelly cert install <device> [flags]
```

### Examples

```
  # Install CA certificate
  shelly cert install kitchen --ca /path/to/ca.pem

  # Install client certificate and key
  shelly cert install kitchen --client-cert cert.pem --client-key key.pem
```

### Options

```
      --ca string            CA certificate file (PEM format)
      --client-cert string   Client certificate file (PEM format)
      --client-key string    Client private key file (PEM format)
  -h, --help                 help for install
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

* [shelly cert](shelly_cert.md)	 - Manage device TLS certificates

