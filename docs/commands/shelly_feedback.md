## shelly feedback

Report issues or request features

### Synopsis

Report issues, bugs, or request features via GitHub.

This command helps create well-formatted GitHub issues with automatic
system information collection.

Issue types:
  bug     - Report a bug or unexpected behavior
  feature - Request a new feature
  device  - Report a device compatibility issue

The command opens your browser to create a GitHub issue with
pre-populated system information.

```
shelly feedback [flags]
```

### Examples

```
  # Report a bug
  shelly feedback --type bug

  # Request a feature
  shelly feedback --type feature --title "Add XYZ support"

  # Report device compatibility issue
  shelly feedback --type device --device kitchen-light

  # Preview the issue without opening browser
  shelly feedback --type bug --dry-run

  # View existing issues
  shelly feedback --issues
```

### Options

```
      --attach-log      Include CLI log info in report
      --device string   Device name/IP for device compatibility issues
      --dry-run         Preview issue without opening browser
  -h, --help            help for feedback
      --issues          Open GitHub issues page instead
      --title string    Issue title
  -t, --type string     Issue type: bug, feature, or device
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

