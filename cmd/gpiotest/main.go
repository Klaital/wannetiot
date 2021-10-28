package main

import (
	"github.com/klaital/max31855"
	log "github.com/sirupsen/logrus"
	//"github.com/warthog618/gpiod"
	//"github.com/warthog618/gpiod/spi/mcp3w0c"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"time"
)

func main() {
	var err error

	//gpioChip, err := gpiod.NewChip("gpiochip0", gpiod.WithConsumer("mbed"))
	//if err != nil {
	//	log.WithField("chips", gpiod.Chips()).WithError(err).Fatal("failed to open gpio chip")
	//}
	//adc, err := mcp3w0c.NewMCP3008(gpioChip, 5, 21, 20, 19)
	//if err != nil {
	//	log.WithError(err).Fatal("Failed to initialize ADC")
	//}
	//
	log.SetLevel(log.DebugLevel)

	_, err = host.Init()
	if err != nil {
		log.WithError(err).Fatal("Failed to init host")
	}

	if _, err = driverreg.Init(); err != nil {
		log.WithError(err).Fatal("Failed to init driverreg")
	}

	p, err := spireg.Open("")
	if err != nil {
		log.WithError(err).Fatal("Failed to find SPI bus")
	}
	defer p.Close()

	dev, err := max31855.New(p)
	if err != nil {
		log.WithError(err).Fatal("Failed to open SPI device!")
	}

	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C
		// Read the thermocouple voltage via the ADC, then print it
		t, err := dev.GetTemp()
		if err != nil {
			log.WithError(err).Fatal("Failed to read temperature")
		}
		log.WithFields(log.Fields{
			"Thermocouple_C": t.Thermocouple.Celsius(),
			"Thermocouple_F": t.Thermocouple.Fahrenheit(),
			"Thermocouple_S": t.Thermocouple.String(),
			"Internal_C": t.Internal.Celsius(),
			"Internal_F": t.Internal.Fahrenheit(),
			"Internal_S": t.Internal.String(),
		}).Debug("Read thermocouple")
	}
}
