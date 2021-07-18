package sensors

import (
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/spi/mcp3w0c"
)

func NewAdc(logger *log.Entry, chip *gpiod.Chip) *mcp3w0c.MCP3w0c {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}
	adc, err := mcp3w0c.NewMCP3008(chip, 11, 22, 10, 9)
	if err != nil {
		logger.WithError(err).Fatal("Failed to setup MCP3008 ADC")
	}
	return adc
}
