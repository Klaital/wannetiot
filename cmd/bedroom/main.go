package main

import (
	"github.com/klaital/wannetiot/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
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

	// Check each of the sensors once to validate the configuration
	chip, err := gpiod.NewChip(cfg.ChipID, gpiod.WithConsumer(cfg.NodeName))
	if err != nil {
		chips := gpiod.Chips()
		logger.WithError(err).WithFields(log.Fields{
			"ChipID": cfg.ChipID,
			"AvailableChips": chips,
		}).Fatal("Failed to init GPIO chip")
	}
	defer chip.Close()

	/// Temperature
	logger.WithField("TempPin", cfg.TempSensorPin).Debug("Initializing temperature sensor")
	//temperaturePin, err := chip.RequestLine(cfg.TempSensorPin, gpiod.AsInput)
	//defer temperaturePin.Close()
	//if err != nil {
	//	logger.WithError(err).WithField("TempPin", cfg.TempSensorPin).Fatal("Failed to load temperature sensor")
	//}
	//li, err := temperaturePin.Info()
	//if err != nil {
	//	logger.WithError(err).WithField("TempPin", cfg.TempSensorPin).Fatal("Failed to load temperature sensor info")
	//}
	//val, err := temperaturePin.Value()
	//if err != nil {
	//	logger.WithError(err).WithField("TempSensor", li).Fatal("Failed to load temperature sensor initial value")
	//}
	//logger.WithFields(log.Fields{
	//	"LineInfo": li,
	//	"InitialReading": val,
	//}).Info("Temperature pin verified")

	/// TODO: Air Quality
	/// TODO: RF Remote Control
	// TODO: Register interrupts for RF Remote

	// Main loop
	//for {
	//	// Sample Temperature
	//	temp, err := temperaturePin.Value()
	//	if err != nil {
	//		logger.WithError(err).Error("Failed to read temperature")
	//	}
	//	// TODO: Sample Air Quality
	//
	//	logger.WithFields(log.Fields{
	//		"temp": temp,
	//	}).Debug("Read sensors")
	//	// TODO: Write the metrics to influx
	//
	//	time.Sleep(cfg.SampleInterval * time.Millisecond)
	//}
}
