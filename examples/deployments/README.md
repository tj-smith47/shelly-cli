# Shelly CLI Deployments

Examples for deploying shelly-cli in various environments.

## Kubernetes Deployment

Deploy shelly-cli as a Prometheus metrics exporter in Kubernetes:

```yaml
# shelly-exporter.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: shelly-config
  namespace: monitoring
data:
  config.yaml: |
    devices:
      kitchen-light:
        address: 192.168.1.50
      living-room:
        address: 192.168.1.51
      garage:
        address: 192.168.1.52
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app: shelly-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: shelly-exporter
  template:
    metadata:
      labels:
        app: shelly-exporter
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
          volumeMounts:
            - name: config
              mountPath: /root/.config/shelly
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
      volumes:
        - name: config
          configMap:
            name: shelly-config
---
apiVersion: v1
kind: Service
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app: shelly-exporter
spec:
  ports:
    - port: 9090
      targetPort: 9090
      name: metrics
  selector:
    app: shelly-exporter
---
# ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: shelly-exporter
  namespace: monitoring
  labels:
    app: shelly-exporter
spec:
  selector:
    matchLabels:
      app: shelly-exporter
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

Apply with:

```bash
kubectl apply -f shelly-exporter.yaml
```

## Docker

```bash
# Run as Prometheus exporter
docker run -d \
  -p 9090:9090 \
  -v ~/.config/shelly:/root/.config/shelly \
  --name shelly-exporter \
  ghcr.io/tj-smith47/shelly-cli:latest \
  metrics prometheus --port 9090
```

## Docker Compose

```yaml
version: '3.8'

services:
  shelly-exporter:
    image: ghcr.io/tj-smith47/shelly-cli:latest
    command: ["metrics", "prometheus", "--port", "9090"]
    ports:
      - "9090:9090"
    volumes:
      - ./config.yaml:/root/.config/shelly/config.yaml:ro
    restart: unless-stopped

  # Optional: Prometheus to scrape metrics
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    depends_on:
      - shelly-exporter
```

Example `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'shelly'
    static_configs:
      - targets: ['shelly-exporter:9090']
    scrape_interval: 30s
```

## Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'shelly'
    static_configs:
      - targets: ['shelly-exporter:9090']  # or localhost:9090
    scrape_interval: 30s
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
