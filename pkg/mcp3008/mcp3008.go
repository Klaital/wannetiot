package mcp3008

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type Mcp3008 struct {
	spi     spi.Conn
	adcMode AdcMode
}

type AdcMode byte

const SingleEndedMode AdcMode = 0x8
const DifferentialMode AdcMode = 0x0

func New(spiPort spi.Port, mode AdcMode) *Mcp3008 {
	c, err := spiPort.Connect(10 * physic.KiloHertz, spi.Mode3, 8)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to ADC port")
	}
	return &Mcp3008{spi: c, adcMode: mode}
}

func (adc *Mcp3008) ReadChannel(channel byte) (int, error) {
	tx := make([]byte, 3)
	rx := make([]byte, 3)

	// First send the channel ID
	tx[0] = 0x01
	tx[1] = (byte(adc.adcMode)+channel) << 4
	tx[2] = 0x00

	adc.spi.Tx(tx, rx)

	result := int(rx[1]&0x3)<<8 + int(rx[2])
	log.Debugf("Read ADC: %X %X %X => %X", rx[2], rx[1], rx[0], result)
	return result, nil
}

func (adc *Mcp3008) ReadChannelAsPct(channel byte) (float64, error) {
	raw, err := adc.ReadChannel(channel)
	if err != nil {return 0.0, err}
	return float64(raw * 100) / 1023.0, nil
}

func (adc *Mcp3008) ReadChannelAsVoltage(channel byte) (float64, error) {
	raw, err := adc.ReadChannel(channel)
	if err != nil {return 0.0, err}
	return float64(raw) * 3.3 / 1023.0, nil
}

func convertValueToVoltage(val int) float64 {
	return (float64(val) * 100.0) / 1023
}
