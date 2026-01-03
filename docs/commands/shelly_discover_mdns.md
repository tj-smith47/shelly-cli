## shelly discover mdns

Discover devices using mDNS/Zeroconf

### Synopsis

Discover Shelly devices using mDNS/Zeroconf.

mDNS (Multicast DNS) is the fastest discovery method. Devices broadcast
their presence on the local network using the _shelly._tcp.local service.
This works best for Gen2+ devices.

Note: mDNS requires multicast support on your network. If devices aren't
found, try 'shelly discover scan' which probes addresses directly.

Output is formatted as a table showing: ID, Address, Model, Generation,
Protocol, and Auth status.

```
shelly discover mdns [flags]
```

### Examples

```
  # Basic mDNS discovery
  shelly discover mdns

  # With longer timeout for slow networks
  shelly discover mdns --timeout 30s

  # Auto-register discovered devices
  shelly discover mdns --register

  # Register but skip devices already in registry
  shelly discover mdns --register --skip-existing

  # Force re-register all discovered devices
  shelly discover mdns --register --skip-existing=false

  # Using aliases
  shelly discover zeroconf --timeout 20s
  shelly discover bonjour --register
```

### Options

```
  -h, --help               help for mdns
      --register           Auto-register discovered devices
      --skip-existing      Skip devices already registered (default true)
  -t, --timeout duration   Discovery timeout (default 10s)
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

* [shelly discover](shelly_discover.md)	 - Discover Shelly devices on the network

