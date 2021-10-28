package lights

import (
	log "github.com/sirupsen/logrus"
	"iot-bedroom-pi/pkg/config"
	"iot-bedroom-pi/pkg/util"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"time"
)

type LightConfig struct {
	Name string
	R    gpio.Duty
	G    gpio.Duty
	W    gpio.Duty
	B    gpio.Duty
}

func DriveLights(cfg *config.Config, settings LightConfig) {
	util.DrivePWM(cfg.LedControlRedPin, settings.R, 5*physic.KiloHertz, nil)
	util.DrivePWM(cfg.LedControlGreenPin, settings.G, 5*physic.KiloHertz, nil)
	util.DrivePWM(cfg.LedControlWhitePin, settings.W, 5*physic.KiloHertz, nil)
	util.DrivePWM(cfg.LedControlBluePin, settings.B, 5*physic.KiloHertz, nil)
}

func RunLightsDemo(cfg *config.Config) {
	//log.Debug("Running PWM demo on Red")
	//lightDemoPwmPin(cfg.LedControlRedPin)
	//log.Debug("Running PWM demo on Green")
	//lightDemoPwmPin(cfg.LedControlGreenPin)
	log.Debug("Running PWM demo on White")
	lightDemoPwmPin(cfg.LedControlWhitePin)
	//log.Debug("Running PWM demo on Blue")
	//lightDemoPwmPin(cfg.LedControlBluePin)
}
func lightDemoPwmPin(pin gpio.PinOut) {
	// Binary on/off flashes
	util.Flash(pin, 250*time.Millisecond, nil)
	time.Sleep(250 * time.Millisecond)
	util.Flash(pin, 250*time.Millisecond, nil)
	time.Sleep(250 * time.Millisecond)

	// Variable brightness
	var err error
	lowBrightness, err := gpio.ParseDuty("10%")
	if err != nil {
		log.WithError(err).Fatal("Failed to parse duty 10%")
	}
	highBrightness, err := gpio.ParseDuty("80%")
	if err != nil {
		log.WithError(err).Fatal("Failed to parse duty 80%")
	}
	err = pin.PWM(lowBrightness, 5*physic.KiloHertz)
	if err != nil {
		log.WithError(err).WithField("pin", pin).Error("Failed to set PWM")
		return
	}
	time.Sleep(500 * time.Millisecond)
	err = pin.PWM(highBrightness, 5*physic.KiloHertz)
	if err != nil {
		log.WithError(err).WithField("pin", pin).Error("Failed to set PWM")
		return
	}
	time.Sleep(500 * time.Millisecond)

	pin.Out(gpio.Low)
}
