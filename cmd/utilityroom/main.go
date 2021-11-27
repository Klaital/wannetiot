package main

import (
	"github.com/klaital/wannetiot/pkg/config"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/host/v3"
	"time"
)

func main() {
	var err error
	cfg := config.New()
	logger := cfg.Logger.WithFields(log.Fields{
		"op": "utilityroom.main",
	})

	logger.Infof("Starting up")

	// Initialize the Thermocouple and Water Sensor pins
	//cfg.InitPins()
	//defer cfg.HaltPins()
	_, err = host.Init()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize host devices")
	}
	if _, err = driverreg.Init(); err != nil {
		logger.WithError(err).Fatal("Failed to init driverreg")
	}

	//logger.Debug("Initializing water sensors")
	//err = cfg.InitWaterSensors()
	//if err != nil {
	//	logger.WithError(err).Fatal("Failed to initialize water sensors")
	//}
	//defer cfg.HaltWaterSensors()
	err = cfg.InitThermocouples()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize thermocouples")
	}
	defer cfg.HaltThermocouples()

	err = cfg.InitializeAdc()
	if err != nil {
		logger.WithError(err).Fatal("Failed to init ADC")
	}
	defer cfg.HaltAdc()

	logger.Debug("Starting ticker")
	ticker := time.NewTicker(cfg.PollInterval)
	for {
		<-ticker.C

		if cfg.Thermocouple1 != nil {
			// Read the thermocouple indicating dryer exhaust temperature
			t, err := cfg.Thermocouple1.GetTemp()
			if err != nil {
				logger.WithError(err).Error("Failed to read thermocouple 1")
			} else {
				// TODO: send the telemetry to Influx
				logger.WithFields(log.Fields{
					"Thermocouple_F": t.Thermocouple.Fahrenheit(),
					"Internal_F":     t.Internal.Fahrenheit(),
				}).Debug("Thermocouple1 data")
			}
			// TODO: sound an alarm if the thermocouple detects a fire
		}

		// TODO: read the water sensors

		// FIXME: remove this after testing
		if cfg.TmpADC != nil {
			raw, err := cfg.TmpADC.ReadChannel(0)
			if err != nil {
				logger.WithError(err).Fatal("Failed to read ADC raw")
			}
			pct, err := cfg.TmpADC.ReadChannelAsPct(0)
			if err != nil {
				logger.WithError(err).Fatal("Failed to read ADC pct")
			}
			vlt, err := cfg.TmpADC.ReadChannelAsVoltage(0)
			if err != nil {
				logger.WithError(err).Fatal("Failed to read ADC vlt")
			}
			logger.WithFields(log.Fields{
				"raw": raw,
				"pct": pct,
				"vlt": vlt,
			}).Debug("Read ADC channel 0")
		}

	}
}
