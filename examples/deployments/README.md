# Shelly CLI Deployments

Examples for deploying shelly-cli as a Prometheus metrics exporter.

## Deployment Options

| Method | Directory | Description |
|--------|-----------|-------------|
| [Kubernetes](kubernetes/) | `kubernetes/` | Kustomize manifests for K8s deployment |
| [Docker](docker/) | `docker/` | Docker Compose with optional Prometheus/Grafana |

## Configuration

Generate a config file with `shelly init` or create one manually.

See [example configs](../config/) for reference.

### Headless Init

Use `shelly init` flags to configure without interactive prompts:

```bash
# Register devices directly
shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

# Device with authentication
shelly init --device secure=192.168.1.102:admin:secret

# Bulk import from JSON
shelly init --devices-json devices.json
```

See the [Docker](docker/) and [Kubernetes](kubernetes/) READMEs for container-specific examples using init containers and run args.

## Quick Start

### Kubernetes (Kustomize)

```bash
# Use as remote base with your config
cat <<EOF > kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/tj-smith47/shelly-cli//examples/deployments/kubernetes
configMapGenerator:
  - name: shelly-config
    behavior: replace
    files:
      - config.yaml=~/.config/shelly/config.yaml
EOF

kustomize build . | kubectl apply -f -
```

### Docker Compose

```bash
cd docker/
docker compose up -d
```

### Docker

```bash
docker run -d \
  -p 9090:9090 \
  -v ~/.config/shelly:/root/.config/shelly:ro \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  metrics prometheus --port 9090
```

## Metrics Exported

| Metric | Type | Description |
|--------|------|-------------|
| `shelly_device_online` | gauge | Device reachability (0/1) |
| `shelly_wifi_rssi` | gauge | WiFi signal strength (dBm) |
| `shelly_uptime_seconds` | counter | Device uptime |
| `shelly_temperature_celsius` | gauge | Device temperature |
| `shelly_switch_on` | gauge | Switch state (0/1) |
| `shelly_power_watts` | gauge | Active power consumption |
| `shelly_voltage_volts` | gauge | Voltage reading |
| `shelly_current_amps` | gauge | Current reading |
| `shelly_energy_wh_total` | counter | Total energy consumed |
| `shelly_frequency_hz` | gauge | AC frequency |
| `shelly_ram_free_bytes` | gauge | Free RAM |
| `shelly_ram_total_bytes` | gauge | Total RAM |

## Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'shelly'
    static_configs:
      - targets: ['shelly-exporter:9090']  # or localhost:9090
    scrape_interval: 30s
```
