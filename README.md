# evohome-prometheus-export
Exposes prometheus formatted data for your Evohome temperature controlled zones.

## Building
On a system with docker installed:
```
cd docker/build
./build.sh
cd ..
sudo docker build -t <image tag>
```
## Installation
* Copy the systemd unit file docker/evohome-prometheus-export.service to /etc/systemd/system
* Copy the secrets.env file somewhere appropriate on your system and set your Evohome username and password.
* Set the permissions of the secrets.env to be readable and writable by root only.
* Update the path to the secrets.env file in the evohome-prometheus-export.service file.
* Update the docker image name in the evohome-prometheus-export.service file.
* ```sudo systemctl enable evohome-prometheus-export```
* To run: ```sudo systemctl start evohome-prometheus-export```

## Configure Prometheus
Add the following to your prometheus.yml file
```
  - job_name: 'evohome'
    scrape_interval: 3m
    static_configs:
      - targets: ['<hostname>:8080']
    metrics_path: /zoneTemperatures
```
