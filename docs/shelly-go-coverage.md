# Shelly-Go Library Coverage Analysis

> **Reference:** The shelly-go library is at `/db/appdata/shelly-go/` and provides comprehensive Shelly device support.

## Currently Exposed in CLI

| Feature | shelly-go Package | CLI Commands | Status |
|---------|-------------------|--------------|--------|
| Discovery | `discovery/` | `discover mdns/ble/coiot/scan` | Complete |
| Switch Control | `gen2/components/switch.go` | `switch on/off/toggle/status/list` | Complete |
| Cover Control | `gen2/components/cover.go` | `cover open/close/stop/position/status` | Complete |
| Light Control | `gen2/components/light.go` | `light on/off/toggle/set/status` | Complete |
| RGB Control | `gen2/components/rgb.go` | `rgb on/off/toggle/set/status` | Complete |
| Input Status | `gen2/components/input.go` | `input list/status/trigger` | Complete |
| Batch Operations | `helpers/batch.go` | `batch on/off/toggle/command` | Complete |
| Scene Management | `helpers/scenes.go` | `scene list/create/activate/etc` | Complete |
| Device Groups | `helpers/groups.go` | `group list/create/add/etc` | Complete |

## Missing from CLI (High Priority)

| Feature | shelly-go Package | CLI Phase | Priority |
|---------|-------------------|-----------|----------|
| **Firmware** | `firmware/` | Phase 6 | **Critical** |
| **Cloud API** | `cloud/` | Phase 9 | **Critical** |
| **Energy Monitoring** | `gen2/components/em.go`, `pm.go` | Phase 11 | High |
| **Scripts** | `gen2/components/script.go` | Phase 7 | High |
| **Schedules** | `gen2/components/schedule.go` | Phase 8 | High |
| **KVS Storage** | `gen2/components/kvs.go` | Phase 18 | High |
| **MQTT Config** | `gen2/components/mqtt.go` | Phase 5 | High |
| **Webhook Config** | `gen2/components/webhook.go` | Phase 5 | High |
| **WiFi Config** | `gen2/components/wifi.go` | Phase 5 | High |

## Missing from CLI (Medium Priority)

| Feature | shelly-go Package | CLI Phase | Priority |
|---------|-------------------|-----------|----------|
| BTHome Devices | `gen2/components/bthome*.go` | Phase 21 | High |
| Zigbee Gateway | `zigbee/` | Phase 21 | High |
| Events/Real-time | `events/` | Phase 11 | Medium |
| Provisioning | `provisioning/` | Phase 19 | Medium |
| Device Profiles | `profiles/` | New Phase | Medium |
| Ethernet Config | `gen2/components/ethernet.go` | Phase 5 | Medium |
| Thermostat | `gen2/components/thermostat.go` | Phase 22C | Medium |
| Temperature | `gen2/components/temperature.go` | Phase 22B | Medium |
| Humidity | `gen2/components/humidity.go` | Phase 22B | Medium |
| Gen1 Devices | `gen1/` | Phase 22A | Medium |

## Missing from CLI (Low Priority)

| Feature | shelly-go Package | CLI Phase | Priority |
|---------|-------------------|-----------|----------|
| Matter | `matter/` | Phase 22 | Medium |
| LoRa | `lora/` | Phase 21 | Low |
| Z-Wave | `zwave/` | Phase 21 | Low |
| ModBus | `gen2/components/modbus.go` | New Phase | Low |
| Flood Sensor | `gen2/components/flood.go` | Phase 22B | Low |
| Smoke Sensor | `gen2/components/smoke.go` | Phase 22B | Low |
| Illuminance | `gen2/components/illuminance.go` | Phase 22B | Low |
| Voltmeter | `gen2/components/voltmeter.go` | Phase 22B | Low |
| Virtual Components | `gen2/components/virtual.go` | New Phase | Low |

## Gen2 Components Coverage Summary

```
/db/appdata/shelly-go/gen2/components/
├── switch.go        ✓ CLI: switch/*
├── cover.go         ✓ CLI: cover/*
├── light.go         ✓ CLI: light/*
├── rgb.go           ✓ CLI: rgb/*
├── rgbw.go          ✗ CLI: needs rgb/* extension
├── input.go         ✓ CLI: input/*
├── script.go        ✗ CLI: Phase 7
├── schedule.go      ✗ CLI: Phase 8
├── kvs.go           ✗ CLI: Phase 18
├── webhook.go       ✗ CLI: Phase 5
├── mqtt.go          ✗ CLI: Phase 5
├── wifi.go          ✗ CLI: Phase 5
├── ethernet.go      ✗ CLI: Phase 5
├── cloud.go         ✗ CLI: Phase 9
├── em.go / em1.go   ✗ CLI: Phase 11
├── pm.go / pm1.go   ✗ CLI: Phase 11
├── bthome*.go       ✗ CLI: Phase 21
├── thermostat.go    ✗ CLI: Phase 22C
├── temperature.go   ✗ CLI: Phase 22B
├── humidity.go      ✗ CLI: Phase 22B
├── flood.go         ✗ CLI: Phase 22B (low)
├── smoke.go         ✗ CLI: Phase 22B (low)
├── illuminance.go   ✗ CLI: Phase 22B (low)
├── voltmeter.go     ✗ CLI: Phase 22B (low)
├── modbus.go        ✗ CLI: New Phase (low)
├── virtual.go       ✗ CLI: New Phase (low)
├── sys.go           ✗ CLI: device info/status (implicit)
├── ble.go           ✗ CLI: discover ble (implicit)
├── devicepower.go   ✗ CLI: device status (implicit)
├── htui.go          ✗ CLI: TUI (implicit)
├── ui.go            ✗ CLI: TUI (implicit)
├── ws.go            ✗ CLI: internal transport
├── sensoraddon.go   ✗ CLI: New Phase (low)
```

## Major Feature Packages

| Package | Description | CLI Status |
|---------|-------------|------------|
| `cloud/` | Shelly Cloud API client, auth, events, websocket | Phase 9 |
| `firmware/` | Check, update, rollback firmware | Phase 6 |
| `events/` | Event bus, handlers, filters | Phase 11 |
| `provisioning/` | BLE transmitter, bulk provisioning | Phase 19 |
| `helpers/` | Batch, groups, scenes, schedule | Partial (batch/groups/scenes done) |
| `profiles/` | Device profiles, capability detection | New Phase |
| `zigbee/` | Full Zigbee gateway support | Phase 21 |
| `matter/` | Enable, disable, pairing codes | Phase 22 |
| `lora/` | Config, send, receive | Phase 21 |
| `zwave/` | Wave device support | Phase 21 |
| `gen1/` | Full Gen1 device support | Phase 22A |
