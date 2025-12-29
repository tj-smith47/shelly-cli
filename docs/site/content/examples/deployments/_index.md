---
title: "Deployment Examples"
description: "Docker and Kubernetes deployment examples"
weight: 50
---

Container and orchestration examples for running Shelly CLI as a metrics exporter.


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


---

## Docker Compose

```yaml
# Shelly CLI Prometheus Exporter
#
# Configuration: Generate config with `shelly init` or place manually at ~/.config/shelly/config.yaml
# Example configs: https://github.com/tj-smith47/shelly-cli/tree/master/examples/config

services:
  shelly-exporter:
    image: ghcr.io/tj-smith47/shelly-cli:latest
    container_name: shelly-exporter
    command: ["metrics", "prometheus", "--port", "9090"]
    ports:
      - "9090:9090"
    volumes:
      - ${XDG_CONFIG_HOME:-$HOME/.config}/shelly:/root/.config/shelly:ro
    restart: unless-stopped
    networks:
      - monitoring
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9090/metrics"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Optional: Prometheus to scrape metrics
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
    restart: unless-stopped
    networks:
      - monitoring
    depends_on:
      - shelly-exporter
    profiles:
      - full

  # Optional: Grafana for dashboards
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    restart: unless-stopped
    networks:
      - monitoring
    depends_on:
      - prometheus
    profiles:
      - full

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus-data:
  grafana-data:
```


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


---

## Kubernetes


Deploy shelly-cli as a Prometheus metrics exporter in Kubernetes using Kustomize.

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- kustomize (or kubectl with kustomize support)
- Shelly config file (generate with `shelly init`)

## Quick Start

### Using as Remote Base (Recommended)

Create an overlay that references this as a remote base:

```yaml
# my-shelly/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/tj-smith47/shelly-cli//examples/deployments/kubernetes

# Provide your config file
configMapGenerator:
  - name: shelly-config
    behavior: replace
    files:
      - config.yaml=~/.config/shelly/config.yaml
```

Then apply:

```bash
kustomize build my-shelly/ | kubectl apply -f -
```

### Local Clone

```bash
# Clone and provide your config
git clone https://github.com/tj-smith47/shelly-cli.git
cd shelly-cli/examples/deployments/kubernetes

# Copy your config (or symlink)
cp ~/.config/shelly/config.yaml .

# Apply
kustomize build . | kubectl apply -f -
```

## Configuration

Generate a config file with `shelly init` or create one manually.

See [example configs](https://github.com/tj-smith47/shelly-cli/tree/master/examples/config) for reference.

### Init Container Setup

Use an init container to configure devices at pod startup:

```yaml
# In your deployment overlay
spec:
  template:
    spec:
      initContainers:
        - name: init-config
          image: ghcr.io/tj-smith47/shelly-cli:latest
          command:
            - shelly
            - init
            - --device=kitchen=192.168.1.100
            - --device=bedroom=192.168.1.101
            - --api-mode=local
            - --no-color
          volumeMounts:
            - name: shelly-config
              mountPath: /root/.config/shelly
      volumes:
        - name: shelly-config
          emptyDir: {}
```

Or import from a ConfigMap containing device JSON:

```yaml
initContainers:
  - name: init-config
    image: ghcr.io/tj-smith47/shelly-cli:latest
    command: ["sh", "-c"]
    args:
      - shelly init --devices-json /config/devices.json --no-color
    volumeMounts:
      - name: devices-json
        mountPath: /config
```

### Image Version

To use a specific version in your overlay:

```yaml
images:
  - name: ghcr.io/tj-smith47/shelly-cli
    newTag: v1.0.0
```

### Prometheus Operator

If you have Prometheus Operator installed, add the ServiceMonitor to your overlay:

```yaml
resources:
  - https://github.com/tj-smith47/shelly-cli//examples/deployments/kubernetes
  - https://github.com/tj-smith47/shelly-cli//examples/deployments/kubernetes/servicemonitor.yaml
```

## Files

| File | Description |
|------|-------------|
| `kustomization.yaml` | Kustomize configuration with configMapGenerator |
| `namespace.yaml` | Creates the monitoring namespace |
| `deployment.yaml` | Exporter deployment |
| `service.yaml` | ClusterIP service for scraping |
| `servicemonitor.yaml` | Prometheus Operator ServiceMonitor (optional) |

## Verification

```bash
# Check deployment status
kubectl get pods -n monitoring -l app=shelly-exporter

# View logs
kubectl logs -n monitoring -l app=shelly-exporter

# Test metrics endpoint
kubectl port-forward -n monitoring svc/shelly-exporter 9090:9090
curl http://localhost:9090/metrics
```

## Prometheus Configuration

If not using Prometheus Operator, add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'shelly'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - monitoring
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
        regex: shelly-exporter
        action: keep
```

Or use static config:

```yaml
scrape_configs:
  - job_name: 'shelly'
    static_configs:
      - targets: ['shelly-exporter.monitoring.svc.cluster.local:9090']
    scrape_interval: 30s
```

### deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app.kubernetes.io/name: shelly-exporter
    app.kubernetes.io/component: exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: shelly-exporter
  template:
    metadata:
      labels:
        app: shelly-exporter
        app.kubernetes.io/name: shelly-exporter
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: shelly-exporter
          image: ghcr.io/tj-smith47/shelly-cli:latest
          args:
            - metrics
            - prometheus
            - --port
            - "9090"
          ports:
            - containerPort: 9090
              name: metrics
              protocol: TCP
          volumeMounts:
            - name: config
              mountPath: /root/.config/shelly
              readOnly: true
          resources:
            requests:
              cpu: 10m
              memory: 32Mi
            limits:
              cpu: 100m
              memory: 64Mi
          livenessProbe:
            httpGet:
              path: /metrics
              port: 9090
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /metrics
              port: 9090
            initialDelaySeconds: 5
            periodSeconds: 10
          securityContext:
            readOnlyRootFilesystem: true
            runAsNonRoot: false
            allowPrivilegeEscalation: false
      volumes:
        - name: config
          configMap:
            name: shelly-config
      securityContext:
        seccompProfile:
          type: RuntimeDefault
```

### service.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app.kubernetes.io/name: shelly-exporter
    app.kubernetes.io/component: exporter
spec:
  type: ClusterIP
  ports:
    - port: 9090
      targetPort: 9090
      name: metrics
      protocol: TCP
  selector:
    app: shelly-exporter
```

### servicemonitor.yaml

```yaml
# ServiceMonitor for Prometheus Operator
# Only apply if you have Prometheus Operator installed
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app.kubernetes.io/name: shelly-exporter
    app.kubernetes.io/component: monitoring
spec:
  selector:
    matchLabels:
      app: shelly-exporter
  namespaceSelector:
    matchNames:
      - monitoring
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
      scrapeTimeout: 10s
```

