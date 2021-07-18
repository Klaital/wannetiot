package sensors

import (
	"github.com/klaital/wannetiot/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod/spi/mcp3w0c"
)

type TMP36 struct {
	adc *mcp3w0c.MCP3w0c
	channel int
	logger *log.Entry
}

func NewTMP36(cfg *config.Config, adc *mcp3w0c.MCP3w0c, channel int) *TMP36 {
	return &TMP36{
		adc: adc,
		channel: channel,
		logger: cfg.LogContext,
	}
}

func (sensor *TMP36) Read() uint16 {
	val, err := sensor.adc.Read(sensor.channel)
	if err != nil {
		sensor.logger.WithError(err).WithFields(log.Fields{
			"operation": "TMP36#Read",
			"channel": sensor.channel,
		}).Error("Failed to read value from ADC")
	}
	return val
}
