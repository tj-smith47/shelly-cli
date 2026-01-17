---
title: "shelly cloud login"
description: "shelly cloud login"
---

## shelly cloud login

Authenticate with Shelly Cloud

### Synopsis

Authenticate with the Shelly Cloud API.

The access token is stored locally for future use.

Three authentication methods are available:

1. Browser OAuth (default):
   Opens your web browser to the Shelly Cloud login page. After you log in,
   the authorization code is automatically captured. This is the most secure
   method as your password is never stored locally.

2. Auth Key (--key, --server):
   Use the authorization key from the Shelly mobile app. Find it in:
   User Settings → Authorization cloud key. You must also provide the
   server URL shown with the key.

3. Email/Password (--email, --password):
   Provide your Shelly Cloud email and password via flags or environment
   variables (SHELLY_CLOUD_EMAIL, SHELLY_CLOUD_PASSWORD).

```
shelly cloud login [flags]
```

### Examples

```
  # OAuth browser flow (default, most secure)
  shelly cloud login

  # Browser flow without auto-opening browser
  shelly cloud login --no-browser

  # Auth key from Shelly App
  shelly cloud login --key MTZkZGM3dWlk... --server shelly-59-eu.shelly.cloud

  # Email/password login
  shelly cloud login --email user@example.com --password mypassword
```

### Options

```
      --email string       Shelly Cloud email
  -h, --help               help for login
      --key string         Authorization key from Shelly App
      --no-browser         Don't auto-open browser, just print the URL
      --password string    Shelly Cloud password
      --port int           Port for OAuth callback server (default: auto-select)
      --server string      Server URL for auth key (e.g., shelly-59-eu.shelly.cloud)
      --timeout duration   Timeout waiting for OAuth callback (default 5m0s)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

