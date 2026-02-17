## shelly discover http

Discover devices via HTTP subnet scanning

### Synopsis

Discover Shelly devices by probing HTTP endpoints on a subnet.

If no subnet is provided, attempts to detect the local network(s).
When multiple subnets are detected, an interactive prompt lets you
choose which to scan. Use --all-networks to scan all without prompting.

This method is slower than mDNS or CoIoT but works when multicast
is blocked or devices are on different VLANs.

The scan probes each IP address in the subnet range for Shelly device
HTTP endpoints. Progress is shown in real-time. Discovered devices
can be automatically registered with --register.

Use --skip-existing (enabled by default) to avoid re-registering
devices that are already in your registry.

Output is formatted as a table showing: ID, Address, Model, Generation,
Protocol, and Auth status.

```
shelly discover http [subnet...] [flags]
```

### Examples

```
  # Scan default network (auto-detect)
  shelly discover http

  # Scan specific subnet
  shelly discover http 192.168.1.0/24

  # Scan multiple subnets
  shelly discover http 192.168.1.0/24 10.0.0.0/24

  # Scan all detected subnets without prompting
  shelly discover http --all-networks

  # Use --network flag (repeatable)
  shelly discover http --network 192.168.1.0/24 --network 10.0.0.0/24

  # Scan a /16 network (large, use longer timeout)
  shelly discover http 10.0.0.0/16 --timeout 30m

  # Auto-register discovered devices
  shelly discover http --register

  # Using 'scan' alias
  shelly discover scan --timeout 5m

  # Force re-register all discovered devices
  shelly discover http --register --skip-existing=false

  # Combine flags for initial network setup
  shelly discover http 192.168.1.0/24 --register --timeout 10m
```

### Options

```
      --all-networks          Scan all detected subnets without prompting
  -h, --help                  help for http
      --network stringArray   Subnet(s) to scan (repeatable, auto-detected if not specified)
      --register              Automatically register discovered devices
      --skip-existing         Skip devices already registered (default true)
  -t, --timeout duration      Scan timeout (default 2m0s)
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly discover](shelly_discover.md)	 - Discover Shelly devices on the network

