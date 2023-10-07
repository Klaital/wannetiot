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

var Full LightConfig = LightSettingsFull()
var Off LightConfig = LightConfig{Name: "LIGHTS_OFF"}
var dimmerPct float64 = 10.0

// LightSettingLow generates the LightConfig for dim Soft White Light.
func LightSettingLow() LightConfig {
	//redDuty, err := gpio.ParseDuty("1%")
	//if err != nil {
	//	log.WithError(err).Error("Failed to parse duty setting")
	//	redDuty = gpio.DutyMax / 100
	//}
	//white, err := gpio.ParseDuty("10%")
	//if err != nil {
	//	log.WithError(err).Error("Failed to parse duty setting")
	//	white = gpio.DutyMax / 10
	//}
	return LightConfig{
		Name: "LIGHTS_LOW",
		R:    gpio.Duty(float64(Full.R) * dimmerPct),
		G:    gpio.Duty(float64(Full.G) * dimmerPct),
		W:    gpio.Duty(float64(Full.W) * dimmerPct),
		B:    gpio.Duty(float64(Full.B) * dimmerPct),
	}
}
