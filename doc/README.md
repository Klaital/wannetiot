# Wannet IoT tools

This repo is a suite of software for running on raspberry pis or other monitoring gear around the house.

# Locations

## Front Door

1. Presence: Is someone at the door?
    - Camera
    - IR Motion detector
2. Door state: Is the door open?
3. Environment: temperature/humidity/air quality

## Back Door

Same as Front Door

## Attic

1. Environment: temperature/humidity/air quality

# Hardware

## Networking

We need an additional private wifi network dedicated for IoT.
This network should be bridged on a single server (probaly klaital.com) used to recieve
sensor data, send out updated builds and relay commands.

A MoCA adapter can enable using the in-wall coax wiring to extend the wired network upstairs.
https://www.amazon.com/Actiontec-Ethernet-Adapter-without-Routers/dp/B008EQ4BQG

Microtik seems like a good router choice:
https://multilink.us/mikrotik-hap-ac3-1/

## Platform

- Raspberry Pi
- Arduino

# Software

## Builds + Deployment

Builds should be automatically deployed to the devices over the LAN.
We can run a buildbot/jenkins/custom scripts on klaital.com to compile the software and ssh out to the Pi's to update/restart them.

### Architecture #1 - microk8s
Easy enough to run a k8s master on klaital.com, with the worker nodes on each Pi.
https://ubuntu.com/tutorials/how-to-kubernetes-cluster-on-raspberry-pi#4-installing-microk8s

This would require each worker to either auto-discover the sensors present, or some way to force a specific config onto specific workers.

Pro:
- Deployments are easy
- Any language/runtime easy to support via docker
  Cons:
- CPU/RAM overhead from k8s

### Architecture #2 - run code natively

Go code is really lightweight anyway. We could just ssh new binaries to the IPs and restart the services using runit or something.

Pro:
- no overhead from docker/k8s
  Cons:
    - Docker isolates code from the serial bus used for most sensors. Connecting them may be challenging.
    - More management overhead - we need a database specifying which builds/configs go to which IPs
    - Like hell I'm dealing with anything other than go exe's outside of docker environments.

## Monitoring + Storage

- Influxdb running on klaital.com will be the primary metrics database and GUI.
- Grafana could add additional visualization support.

## Automation

HomeAssistant running on klaital.com will provide smarthome automation tooling.

