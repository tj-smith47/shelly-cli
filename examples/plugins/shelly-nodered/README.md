# shelly-nodered

Export Shelly devices to Node-RED flow configurations.

## Installation

```bash
# Install directly
shelly plugin install ./shelly-nodered

# Or copy to plugins directory
cp shelly-nodered ~/.config/shelly/plugins/
chmod +x ~/.config/shelly/plugins/shelly-nodered
```

## Dependencies

- **jq** - JSON processing (`apt install jq` or `brew install jq`)
- **Node-RED** - Flow-based automation platform
- **node-red-dashboard** (optional) - For dashboard generation

## Usage

```bash
# Generate basic flows for all devices
shelly nodered flows

# Generate reusable Shelly control subflow
shelly nodered subflow

# Generate dashboard flows (requires node-red-dashboard)
shelly nodered dashboard

# List devices that would be exported
shelly nodered list

# Generate to specific file
shelly nodered flows my-flows.json
```

## Commands

| Command | Description |
|---------|-------------|
| `flows [file]` | Generate device control flows |
| `subflow [file]` | Generate reusable Shelly control subflow |
| `dashboard [file]` | Generate dashboard flows with UI |
| `list` | List devices available for export |
| `help` | Show help |
| `version` | Show version |

## Generated Flows

### Basic Device Flows (`flows`)

Each device gets a flow with:
- **Inject node** - Manual trigger button
- **HTTP Request node** - Calls Shelly RPC API
- **Debug node** - Shows response

```json
{
    "id": "kitchen_light_inject",
    "type": "inject",
    "name": "Toggle Kitchen Light",
    "topic": "",
    "payload": "",
    "wires": [["kitchen_light_request"]]
}
```

### Reusable Subflow (`subflow`)

A parameterized subflow that works with any Shelly device:
- Input: `msg.device` (IP), `msg.method` (RPC method), `msg.params`
- Output: API response

**Usage in Node-RED:**
1. Import the subflow
2. Use in any flow with a function node setting device/method
3. Connect to your processing nodes

### Dashboard Flows (`dashboard`)

Generates Node-RED Dashboard UI with:
- **Switch** - On/Off toggle per device
- **Text** - Status display
- **Button** - Toggle control

Requires the `node-red-dashboard` package installed.

## Node-RED Setup

1. Install Node-RED:
   ```bash
   npm install -g node-red
   ```

2. Start Node-RED:
   ```bash
   node-red
   ```

3. Generate flows:
   ```bash
   shelly nodered flows
   ```

4. Import into Node-RED:
   - Open http://localhost:1880
   - Menu (☰) → Import
   - Select `shelly-flows.json`
   - Click Import

5. Deploy flows:
   - Click Deploy button

## Example: Complete Setup

```bash
# Generate all components
shelly nodered flows flows.json
shelly nodered subflow subflow.json
shelly nodered dashboard dashboard.json

# Combine into single import file
jq -s 'add' flows.json subflow.json dashboard.json > shelly-complete.json
```

## Customization

After importing, customize nodes in Node-RED:
- Add error handling with catch nodes
- Connect to MQTT for persistence
- Add notification nodes (email, Telegram, etc.)
- Create automation rules with function nodes

## Device API Methods

The generated flows use these Shelly Gen2+ RPC methods:

| Method | Description |
|--------|-------------|
| `Switch.Toggle` | Toggle switch state |
| `Switch.Set` | Set switch on/off |
| `Switch.GetStatus` | Get current state |
| `Light.Toggle` | Toggle light |
| `Cover.Open` | Open cover/roller |
| `Cover.Close` | Close cover/roller |
| `Shelly.GetStatus` | Get full device status |

For Gen1 devices, the plugin uses REST endpoints:
- `/relay/0?turn=toggle`
- `/relay/0?turn=on`
- `/relay/0?turn=off`

## Environment Variables

Automatically provided by Shelly CLI:

- `SHELLY_DEVICES_JSON` - JSON of registered devices
- `SHELLY_CONFIG_PATH` - Path to config file
- `SHELLY_NO_COLOR` - Disable colored output

## Links

- [Node-RED Documentation](https://nodered.org/docs/)
- [Node-RED Dashboard](https://flows.nodered.org/node/node-red-dashboard)
- [Shelly Gen2+ API Reference](https://shelly-api-docs.shelly.cloud/gen2/)
