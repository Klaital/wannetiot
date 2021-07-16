package sensors

import (
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
)

func NewAdc(logger *log.Entry) AdcRead {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}
	return AdcRead{
		Cs:      &gpiod.Line{},
		Clock:   &gpiod.Line{},
		Miso:    &gpiod.Line{},
		NumBits: 0,
		Results: nil,
		logger:  logger,
	}
}
type AdcRead struct {
	Cs *gpiod.Line
	Clock *gpiod.Line
	Miso *gpiod.Line
	NumBits int
	Results chan uint32
	logger *log.Entry
}

// Read fetches the current value from the ADC register
func (r AdcRead) Read() {
	var result uint32
	r.Cs.SetValue(0)

	for i := 0; i < r.NumBits; i++ {
		r.Clock.SetValue(1)
		bit, err := r.Miso.Value()
		if err != nil {
			r.logger.WithError(err).WithField("bit", i).Error("Failed to read MISO value")
			// TODO: bubble the error up
			return
		}
		if bit > 0 {
			result |= 0x1
		}
		// Shift left
		if i != r.NumBits-1 {
			result <<= 1
		}
		r.Clock.SetValue(0)
	}

	r.Cs.SetValue(1)

	r.Results <- result
}
