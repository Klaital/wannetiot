package util

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"time"
)


func Flash(pin gpio.PinOut, d time.Duration, logger *log.Entry) {
	if logger == nil{
		logger = log.NewEntry(log.New())
	}
	logger = logger.WithFields(log.Fields{
		"pin": pin.String(),
		"duration": d.String(),
		"op": "Flash",
	})
	if err := pin.Out(gpio.High); err != nil {
		logger.WithError(err).Error("Failed to drive pin high")
	}
	time.Sleep(d)
	if err := pin.Out(gpio.Low); err != nil {
		logger.WithError(err).Error("Failed to drive pin low")
	}
}
