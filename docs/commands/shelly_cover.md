## shelly cover

Control cover/roller components

### Synopsis

Control cover and roller shutter components on Shelly devices.

### Examples

```
  # Open a cover
  shelly cover open bedroom

  # Close a cover
  shelly cv close living-room

  # Set cover to 50% position
  shelly cover position bedroom --pos 50
```

### Options

```
  -h, --help   help for cover
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly cover calibrate](shelly_cover_calibrate.md)	 - Calibrate cover
* [shelly cover close](shelly_cover_close.md)	 - Close cover
* [shelly cover list](shelly_cover_list.md)	 - List cover components
* [shelly cover open](shelly_cover_open.md)	 - Open cover
* [shelly cover position](shelly_cover_position.md)	 - Set cover position
* [shelly cover status](shelly_cover_status.md)	 - Show cover status
* [shelly cover stop](shelly_cover_stop.md)	 - Stop cover

