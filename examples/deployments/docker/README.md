# Docker Deployment

Run shelly-cli as a Prometheus metrics exporter using Docker or Docker Compose.

## Prerequisites

- Docker 20.10+
- Docker Compose v2 (optional)
- Shelly config at `~/.config/shelly/config.yaml` (generate with `shelly init`)

## Quick Start

### Using Docker Compose

```bash
# Start just the exporter
docker compose up -d shelly-exporter

# Start with full monitoring stack (Prometheus + Grafana)
docker compose --profile full up -d

# View logs
docker compose logs -f shelly-exporter

# Stop services
docker compose down
```

### Using Docker Directly

```bash
docker run -d \
  -p 9090:9090 \
  -v ~/.config/shelly:/root/.config/shelly:ro \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  metrics prometheus --port 9090
```

## Configuration

Generate a config file with `shelly init` or create one manually.

See [example configs](https://github.com/tj-smith47/shelly-cli/tree/master/examples/config) for reference.

### Container Init Setup

Initialize config inside the container using run args:

```bash
# One-liner: init then run metrics
docker run -d \
  -p 9090:9090 \
  -v shelly-config:/root/.config/shelly \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  sh -c 'shelly init --device=kitchen=192.168.1.100 --device=bedroom=192.168.1.101 --no-color && shelly metrics prometheus --port 9090'

# Or init first, then run
docker run --rm \
  -v shelly-config:/root/.config/shelly \
  ghcr.io/tj-smith47/shelly-cli:latest \
  init --device=kitchen=192.168.1.100 --no-color

docker run -d \
  -p 9090:9090 \
  -v shelly-config:/root/.config/shelly \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  metrics prometheus --port 9090
```

Import devices from a JSON file:

```bash
docker run --rm \
  -v shelly-config:/root/.config/shelly \
  -v ./devices.json:/config/devices.json:ro \
  ghcr.io/tj-smith47/shelly-cli:latest \
  init --devices-json /config/devices.json --no-color
```

### Network Mode

For device discovery to work, you may need host network mode:

```bash
docker run -d \
  --network host \
  -v ~/.config/shelly:/root/.config/shelly:ro \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  metrics prometheus --port 9090
```

## Files

| File | Description |
|------|-------------|
| `docker-compose.yml` | Docker Compose configuration |
| `prometheus.yml` | Prometheus scrape configuration (for full profile) |

## Profiles

The Docker Compose file supports profiles:

- **Default** (no profile): Just the shelly-exporter
- **full**: Includes Prometheus and Grafana

```bash
# Exporter only
docker compose up -d

# Full monitoring stack
docker compose --profile full up -d
```

## Verification

```bash
# Check container status
docker compose ps

# Test metrics endpoint
curl http://localhost:9090/metrics

# View exporter logs
docker compose logs shelly-exporter
```

## Grafana Setup

When using the full profile, Grafana is available at http://localhost:3000

1. Login with admin/admin
2. Add Prometheus data source: http://prometheus:9090
3. Import or create dashboards for Shelly metrics
