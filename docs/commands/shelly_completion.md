## shelly completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for shelly.

To load completions:

Bash:

  $ source <(shelly completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ shelly completion bash > /etc/bash_completion.d/shelly
  # macOS:
  $ shelly completion bash > $(brew --prefix)/etc/bash_completion.d/shelly

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ shelly completion zsh > "${fpath[1]}/_shelly"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ shelly completion fish | source

  # To load completions for each session, execute once:
  $ shelly completion fish > ~/.config/fish/completions/shelly.fish

PowerShell:

  PS> shelly completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> shelly completion powershell > shelly.ps1
  # and source this file from your PowerShell profile.

Use 'shelly completion install' to automatically install completions.

### Examples

```
  # Generate bash completions
  shelly completion bash

  # Generate and install completions automatically
  shelly completion install

  # Install for specific shell
  shelly completion install --shell zsh
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly completion bash](shelly_completion_bash.md)	 - Generate bash completion script
* [shelly completion fish](shelly_completion_fish.md)	 - Generate fish completion script
* [shelly completion install](shelly_completion_install.md)	 - Install shell completions
* [shelly completion powershell](shelly_completion_powershell.md)	 - Generate PowerShell completion script
* [shelly completion zsh](shelly_completion_zsh.md)	 - Generate zsh completion script

