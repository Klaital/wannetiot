package lights

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
)

// LightSettingsFull generates the LightConfig for full-power Soft White Light.
func LightSettingsFull() LightConfig {
	redDuty, err := gpio.ParseDuty("60%")
	if err != nil {
		log.WithError(err).Error("Failed to parse duty setting")
		redDuty = gpio.DutyMax / 5
	}
	greenDuty, err := gpio.ParseDuty("10%")
	if err != nil {
		log.WithError(err).Error("Failed to parse duty setting")
		greenDuty = gpio.DutyMax / 5
	}
	return LightConfig{
		Name: "LIGHTS_FULL",
		R:    redDuty,
		G:    greenDuty,
		W:    gpio.DutyMax,
		B:    0,
	}
}

var LightsFull LightConfig = LightSettingsFull()
var LightsOff LightConfig = LightConfig{Name: "LIGHTS_OFF"}

// LightSettingLow generates the LightConfig for dim Soft White Light.
func LightSettingLow() LightConfig {
	redDuty, err := gpio.ParseDuty("1%")
	if err != nil {
		log.WithError(err).Error("Failed to parse duty setting")
		redDuty = gpio.DutyMax / 100
	}
	white, err := gpio.ParseDuty("10%")
	if err != nil {
		log.WithError(err).Error("Failed to parse duty setting")
		white = gpio.DutyMax / 10
	}
	return LightConfig{
		Name: "LIGHTS_LOW",
		R:    redDuty,
		G:    0,
		W:    white,
		B:    0,
	}
}
