## shelly rgbw set

Set RGBW parameters

### Synopsis

Set parameters of an RGBW light component on the specified device.

You can set color values (red, green, blue), white channel, brightness, and on/off state.
Values not specified will be left unchanged.

```
shelly rgbw set <device> [flags]
```

### Examples

```
  # Set RGBW color to red
  shelly rgbw set living-room --red 255 --green 0 --blue 0

  # Set RGBW with white channel
  shelly rgbw set living-room -r 255 -g 200 -b 150 --white 128

  # Set RGBW with brightness
  shelly rgbw color living-room -r 0 -g 255 -b 128 --brightness 75

  # Set only white channel
  shelly rgbw set living-room --white 200
```

### Options

```
  -b, --blue int         Blue value (0-255) (default -1)
      --brightness int   Brightness (0-100) (default -1)
  -g, --green int        Green value (0-255) (default -1)
  -h, --help             help for set
  -i, --id int           RGBW component ID (default 0)
      --on               Turn on
  -r, --red int          Red value (0-255) (default -1)
  -w, --white int        White channel value (0-100) (default -1)
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

* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs

