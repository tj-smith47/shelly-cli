# Kubernetes Deployment

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
