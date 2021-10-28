package util

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

// DrivePWM sets the pin off when the duty given is 0,
// otherwise to the given frequency and duty cycle. Errors are logged.
func DrivePWM(pin gpio.PinOut, d gpio.Duty, f physic.Frequency, logger *log.Entry) {
	if logger == nil {
		logger = log.NewEntry(log.New())
	}
	logger = logger.WithFields(log.Fields{
		"op":"drivePWM",
		"pin": pin.String(),
		"duty": d.String(),
		"freq": f.String(),
	})
	if pin == nil {
		return
	}
	var err error
	if d == 0 {
		err = pin.Out(gpio.Low)
	} else {
		err = pin.PWM(d, f)
	}
	if err != nil {
		log.WithError(err).WithField("pin", pin.String()).Error("Failed to drive LED pin")
	}
}
