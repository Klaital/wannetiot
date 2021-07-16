# IoT Nodes

# Summary

## Platforms

**Microcontrollers:**
1. Arduino has a nano formfactor device with builtin wifi chip
   - $15-20
   - https://store.arduino.cc/usa/nano-33-iot-with-headers
    - https://www.amazon.com/Arduino-Nano-BLE-headers-Mounted/dp/B07WXKWNQF
1. RasPi has a nano microcontroller, but it needs a separate wifi module to be wired in and programmed.
   - $4 for the controller, $15 for the wifi module
   - https://www.raspberrypi.org/forums/viewtopic.php?t=300153
1. Coming soon: Arduino-made nanocontroller powered by the RPi chip.
   - $25
   - onboard mic, 6-axis sensors
   - https://store.arduino.cc/usa/nano-rp2040-connect-with-headers
   
## Sensors

| Sensor | Qty |
| ------ | --- |
| Temperature/Humidity | 5 | 
| Outdoor Temp/Humidity | 1 |
| Soil Temp | 1 |
| Air Quality | 3 | 
| VOC | 3 | 
| Motion Detector | 4 |
| Door State | 3 |

## Bedroom


**Platform:** rpi

**Device IP:** tbd

**Sensors:**
1. Temperature/Humidity
1. Air Quality 
1. VOC (Carbon Monoxide)
1. Presence/motion
1. RF Remote control

**Controlled Widgets:**
1. Bedroom Lights (color, brightness, on/off)
1. Slack pager

## Laundry Room
**Platform:** rpi

**Device IP:** TBD

**Sensors:**
1. Temperature
1. Flooding under washing machine
1. Maybe separate flooding sensor deployed to the water heater pan?

## Front Door:
**Platform:** arduino with battery 

**Device IP:** TBD

**Sensors:**
1. Temperature/Humidity
1. Door state
1. Motion

**Controlled Widgets:**
1. Nightlight

## Back Door:
**Platform:** arduino

**Device IP:** TBD

**Sensors:**
1. Temperature/Humidity
1. Air Quality
1. VOC (Carbon Monoxide)
1. Door state
1. Motion

## Bonus Room:
**Platform:** arduino

**Device IP:** TBD

**Sensors:**
1. Window vibration

## Rec Room Windows:
**Platform:** arduino

**Device IP:** TBD

**Sensors:**
1. Window vibration for both windows
1. Arduino nano 33 iot has an onboard 6axis IMU, could be used for window strike detection.

## Attic:

With a roof penetration somewhere, this could become a general-purpose weather station.
https://projects.raspberrypi.org/en/projects/build-your-own-weather-station

**Platform:** rpi

**Device IP:** TBD

**Sensors:**
1. Temperature/Humidity
1. Presence/motion?

## Garage:

**Platform:** arduino

**Device IP:** TBD

**Sensors:**
1. Temperature/Humidity
1. VOC (Carbon Monoxide)
1. Presence/motion
1. Garage door state

## Shed:
**Platform:** rpi

**Device IP:** TBD

**Sensors:**
1. Outdoor Temperature/Humidity
1. Outdoor Air Quality
1. Motion
1. Camera
1. Soil Temperature

