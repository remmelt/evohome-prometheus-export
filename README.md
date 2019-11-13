# evohome-prometheus-export
Exposes prometheus formatted data for your Evohome temperature controlled zones.

## Building
On a system with docker installed:
```
make build
```
## Installation
Run the resulting Docker image in k8s/Nomad/etc. Set the SERVER_PORT env var.

## Configure Prometheus
Add the following to your prometheus.yml file
```
  - job_name: 'evohome'
    scrape_interval: 3m
    static_configs:
      - targets: ['<hostname>:8080']
    metrics_path: /zoneTemperatures
```
