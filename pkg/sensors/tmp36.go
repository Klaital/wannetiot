package sensors

import (
	"github.com/klaital/wannetiot/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod/spi/mcp3w0c"
)

type TMP36 struct {
	adc     *mcp3w0c.MCP3w0c
	channel int
	logger  *log.Entry
}

func NewTMP36(cfg *config.Config, adc *mcp3w0c.MCP3w0c, channel int) *TMP36 {
	return &TMP36{
		adc:     adc,
		channel: channel,
		logger:  cfg.LogContext,
	}
}

func (sensor *TMP36) Read() float64 {
	// rawValue will be a value [0-1023] indicating a ratio of the voltage 0-3.3V
	rawValue, err := sensor.adc.Read(sensor.channel)
	if err != nil {
		sensor.logger.WithError(err).WithFields(log.Fields{
			"operation": "TMP36#Read",
			"channel":   sensor.channel,
		}).Error("Failed to read value from ADC")
	}

	degC := float64(rawValue) / 3.3
	return degC
}

func (sensor *TMP36) ReadFahrenheit() float64 {
	degC := sensor.Read()
	return (degC * 1.8) + 32
}
