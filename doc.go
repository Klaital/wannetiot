/*
iot-bedroom-pi is a program to be run on a Raspberry Pi in my bedroom, controlling some sensors and other devices.

It does:

	- Read SDS011 Air Quality sensor
	- Read AM2302 Temperature & Humidity sensor
	- Drive RGWB LED strip via power transistors over PWM
	- Accept control inputs via RF remote controls
	- Accept control inputs via custom-built Control Panels

The connected controllers can:
	- Control light settings: on/dim/off
	- Send a pager notification to Slack

The sensors are read on a regular interval, and the results sent to a centralized InfluxDB.
The control inputs should be handled as an interrupt, but as they are buffered
through a hardware latch, some latency is acceptable.
 */
package main
