package main

import (
	"github.com/klaital/wannetiot/pkg/config"
	"github.com/klaital/wannetiot/pkg/sensors"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
	"time"
)

func main() {
	// Initialization
	log.Printf("Loading configuration")
	cfg := config.Load()
	logger := cfg.LogContext
	logger.Info("Config loaded")
	influxClient := cfg.ConnectInflux()
	defer influxClient.Close()

	logger.WithField("InfluxHost", cfg.InfluxHost).Info("Influxdb connection initialized")

	// Initialize the ADC
	chip, err := gpiod.NewChip(cfg.ChipID, gpiod.WithConsumer(cfg.NodeName))
	if err != nil {
		chips := gpiod.Chips()
		logger.WithError(err).WithFields(log.Fields{
			"ChipID":         cfg.ChipID,
			"AvailableChips": chips,
		}).Fatal("Failed to init GPIO chip")
	}
	defer chip.Close()
	logger.WithField("TempPin", cfg.TempSensorPin).Debug("Initializing temperature sensor")
	adc := sensors.NewAdc(cfg.LogContext, chip)

	// Check each of the sensors once to validate the configuration

	/// Temperature
	tempSensor := sensors.NewTMP36(cfg, adc, cfg.TempSensorPin)
	logger.WithField("temperature", tempSensor.Read()).Info("Initial temperature reading")

	/// TODO: Air Quality
	/// TODO: RF Remote Control
	// TODO: Register interrupts for RF Remote

	// Main loop
	var temperature float64
	for {
		// Sample Temperature
		tempC := tempSensor.Read()
		temperature = tempSensor.ReadFahrenheit()

		// TODO: Sample Air Quality

		logger.WithFields(log.Fields{
			"temp": temperature,
			"tempC": tempC,
		}).Debug("Read sensors")
		// TODO: Write the metrics to influx

		time.Sleep(cfg.SampleInterval)
	}
}
