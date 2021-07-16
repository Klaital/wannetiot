# Wannet IoT Architecture

1. The IoT devices will be on a separate (W)LAN from the rest of the home network.
1. The main controlling hub will be hosted on klaital.com as the network bridge.
1. The sensors will be run on Raspberry Pi devices natively.
1. The sensor software will be held in this repo, along with build scripts.
1. Deployment will consist of SSH commands to copy a new binary, update the env vars if needed, and restart the service from the build host on klaital.com
1. Sensors will write their telemetry to an InfluxDB hosted on klaital.com
1. Home automation will be operated using HomeAssistant running on klaital.com. It may not be accessible from outside the home network.
