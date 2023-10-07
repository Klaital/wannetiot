// Package ctlpanel provides utilities for interacting with my custom-built
// Control Panel over GPIO.
// The hardware features a potentiometer dial for adjusting light brightness,
// a physical pager button, a capacitive-touch button for turning the lights
// on and off, and an LED and speaker for confirming user input.
package ctlpanel

import (
	"context"
	"github.com/klaital/wannetiot/pkg/lights"
	"github.com/klaital/wannetiot/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod/spi/mcp3w0c"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"time"
)

// ControlPanel models the hardware interface to the panel's devices. It also tracks the
// reads from the dimmer to only trigger an update when the value actually changes.
type ControlPanel struct {
	ResetPin        gpio.PinOut
	PagerPin        gpio.PinIn
	LightSwitchPin  gpio.PinIn
	LedPin          gpio.PinOut
	Speaker         gpio.PinOut
	DimmerChannel   int
	DimmerAdc       *mcp3w0c.MCP3w0c
	DimmerLastValue uint16

	logger *log.Entry
}

// New generates the ControlPanel struct with the specified pins.
// It will ensure that there is a valid logger attached.
func New(pager, light gpio.PinIn, reset, led, speaker gpio.PinOut, dimmerSelect int, dimmerAdc *mcp3w0c.MCP3w0c, logger *log.Entry) ControlPanel {
	if logger == nil {
		logger = log.NewEntry(log.New())
	}
	return ControlPanel{
		ResetPin:        reset,
		PagerPin:        pager,
		LightSwitchPin:  light,
		LedPin:          led,
		Speaker:         speaker,
		DimmerChannel:   dimmerSelect,
		DimmerAdc:       dimmerAdc,
		DimmerLastValue: 0,
		logger:          logger,
	}
}

// ReadDimmerPct reads the ADC supporting this control panel, and normalizes
// the response to an integer in the range [0, 100]
func (p *ControlPanel) ReadDimmerPct() uint16 {
	rawValue, err := p.DimmerAdc.Read(p.DimmerChannel)
	if err != nil {
		p.logger.WithFields(log.Fields{
			"operation": "ControlPanel#ReadDimmerPct",
			"channel":   p.DimmerChannel,
		}).WithError(err).Error("Failed to read value from ADC")
		return 0
	}

	raw32 := uint32(rawValue) * 100
	return uint16(raw32 / 1023)
}

// BlinkLED drives the panel's button's LED on for the specified duration.
func (p *ControlPanel) BlinkLED(d time.Duration) {
	logger := p.logger.WithField("pin", p.LedPin.String())
	if err := p.LedPin.Out(gpio.Low); err != nil {
		logger.WithError(err).Error("Failed to drive LED pin low")
	}
	if err := p.LedPin.Out(gpio.High); err != nil {
		logger.WithError(err).Error("Failed to drive LED pin high")
	}
	time.Sleep(d)
	if err := p.LedPin.Out(gpio.Low); err != nil {
		logger.WithError(err).Error("Failed to drive LED pin low")
	}
}

// Chirp plays the specified tone for the given duration on the Panel's onboard speaker.
func (p *ControlPanel) Chirp(tone physic.Frequency, d time.Duration) {
	logger := p.logger.WithFields(log.Fields{
		"pin":      p.Speaker.String(),
		"op":       "ControlPanel#Chirp",
		"tone":     tone.String(),
		"duration": d.String(),
	})
	duty, err := gpio.ParseDuty("50%")
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse duty 10%")
	}
	if err := p.Speaker.PWM(duty, tone); err != nil {
		logger.WithError(err).Error("Failed to drive speaker")
	}
	time.Sleep(d)
	if err := p.Speaker.Out(gpio.Low); err != nil {
		logger.WithError(err).Error("Failed to set speaker off")
	}
}

func (p *ControlPanel) ResetLatches() {
	util.Flash(p.ResetPin, 1*time.Millisecond, p.logger)
}

func (p *ControlPanel) HandlePager() {
	if p.PagerPin.Read() {
		log.Debug("Pager request detected")
		// Flash the LED and chirp the speaker, then reset the latch
		p.BlinkLED(500 * time.Millisecond)
		p.Chirp(6*physic.KiloHertz, 80*time.Millisecond)
		time.Sleep(50 * time.Millisecond)
		p.Chirp(6*physic.KiloHertz, 80*time.Millisecond)

		util.Flash(p.ResetPin, 1*time.Millisecond, p.logger)
		// TODO: send the notification to slack, then chirp and flash the LED again
		time.Sleep(500 * time.Millisecond)
		p.BlinkLED(500 * time.Millisecond)
		p.Chirp(5*physic.KiloHertz, 200*time.Millisecond)
	}
}

func (p *ControlPanel) HandleTouchSwitch(oldLightSettings *lights.LightConfig) (updated bool) {
	if p.LightSwitchPin.Read() {
		log.Debug("Light Switch touch detected")

		p.BlinkLED(500 * time.Millisecond)
		p.Chirp(5*physic.KiloHertz, 80*time.Millisecond)
		time.Sleep(50 * time.Millisecond)
		p.Chirp(7*physic.KiloHertz, 80*time.Millisecond)

		util.Flash(p.ResetPin, 1*time.Millisecond, nil)
		// If the lights are off, turn them to low. If low, go high. If high, turn off.
		// If in the middle of using the dimmer, move directly to "high".
		log.WithFields(log.Fields{
			"oldsettings": oldLightSettings,
		}).Debug("Changing light settings")
		var newSettings lights.LightConfig
		switch oldLightSettings.Name {
		case "LIGHTS_OFF":
			newSettings.Name = "LIGHTS_LOW"
			newSettings.R = lights.Full.R / 10
			newSettings.G = lights.Full.G / 10
			newSettings.W = lights.Full.W / 10
			newSettings.B = lights.Full.B / 10
		case "LIGHTS_FULL":
			newSettings.Name = "LIGHTS_OFF"
			// Use default zero values
		case "LIGHTS_LOW":
			newSettings = lights.Full
		default:
			newSettings = lights.Full
		}
		log.WithFields(log.Fields{
			"old": oldLightSettings,
			"new": newSettings,
		}).Debug("light settings updated")

		oldLightSettings.Name = newSettings.Name
		oldLightSettings.R = newSettings.R
		oldLightSettings.G = newSettings.G
		oldLightSettings.W = newSettings.W
		oldLightSettings.B = newSettings.B
		return true
	}
	return false
}

// StartPagerInterrupt will watch for edges on the Pager pin, and write into the given
// channel when one is detected. It runs forever until the context is cancelled.
func (p *ControlPanel) StartPagerInterrupt(ctx context.Context, ch chan uint8, panelCode uint8) {
	//go func() {
	//	for {
	//		p.PagerPin.WaitForEdge(1 * time.Second)
	//		if p.PagerPin.Read() {
	//			ch <- panelCode
	//		}
	//	}
	//}()

	<-ctx.Done()
}

// AcknowledgePager plays a hardcoded sequence of lights and beeps to confirm the pager request was received.
func (p *ControlPanel) AcknowledgePager() {
	p.logger.Info("Pager acknowledged")
	//go func() {
	//	util.Flash(p.LedPin, 250*time.Millisecond, nil)
	//	time.Sleep(80 * time.Millisecond)
	//	util.Flash(p.LedPin, 250*time.Millisecond, nil)
	//}()
	//go func() {
	//	p.Chirp(5*physic.KiloHertz, 80*time.Millisecond)
	//	time.Sleep(80 * time.Millisecond)
	//	p.Chirp(5*physic.KiloHertz, 80*time.Millisecond)
	//}()
}

// StartLightsInterrupt will watch for edges on the LightSwitch pin, and write into the given
// channel when one is detected. It runs forever until the context is cancelled
func (p *ControlPanel) StartLightsInterrupt(ctx context.Context, ch chan uint8, panelCode uint8) {
	//go func() {
	//	for {
	//		if p.LightSwitchPin.WaitForEdge(1 * time.Second) {
	//			ch <- panelCode
	//		}
	//	}
	//}()
	//
	<-ctx.Done()
}
