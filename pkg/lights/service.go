package lights

import (
	"errors"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"time"
)

type PowerSetting string

const (
	POWER_FULL PowerSetting = "full"
	POWER_LOW  PowerSetting = "low"
	POWER_OFF  PowerSetting = "off"
)

// Service will enable the Central Control module to operate a RGWB LED light strip.
// Supported operations:
//   - Turn Lights On (full power)
//   - Turn Lights Dim (by default, ~20% power)
//   - Turn Lights Off
//   - Gradually turn lights on (0% -> 100% over 30 minutes)
//   - Should be cancelled if any other light commands are given
//   - Configure how strong the colored lights are when at full/low power.
//   - Flash the lights in a specific color
type Service struct {

	// The GPIO pins to use for driving the lights
	RedPin   gpio.PinOut
	GreenPin gpio.PinOut
	WhitePin gpio.PinOut
	BluePin  gpio.PinOut

	// BaseSetting specifies the color mix at 100% power. Other modes are calculated based on this.
	BaseSetting LightConfig
	// DimPowerSetting specifies the dimmer % for Low power operation, in the form of a multiplier in the range 0.0 - 1.0
	DimPowerSetting float64
	// CurrentPower specifies the current requested power level
	CurrentPower float64

	// haltWakeup is used to signal a request to stop the gradual wakeup sequence.
	// The value passed through the channel dictates the new state of the lights
	haltWakeup chan PowerSetting
	// wakeupInProgress is used to signal that the wakeup sequence is in progress.
	wakeupInProgress *time.Timer
	// wakeupTicker is used to increment the lights power gradually
	wakeupTicker *time.Ticker

	// wakeupStarted records the time when the wakeup sequence started, to
	// facilitate calculating the elapsed time. It will be reset back to 0 when
	// the wakeup is completed (or cancelled).
	wakeupStarted time.Time
}

var ErrInvalidMultiplier = errors.New("invalid power multiplier - valid range 0.0-1.0")

func (s *Service) ConfigurePowerDim(newMultiplier float64) error {
	if newMultiplier > 1.0 || newMultiplier < 0.0 {
		return ErrInvalidMultiplier
	}
	s.DimPowerSetting = newMultiplier
	return nil
}

func (s *Service) ConfigureColors(newColors LightConfig) {
	s.BaseSetting = newColors
	// Base Setting always has white LED channel at full power
	s.BaseSetting.W = gpio.DutyMax
}

func (s *Service) driveLights() {
	// TODO: hook up pins
	s.RedPin.PWM(gpio.Duty(float64(s.BaseSetting.R)*s.CurrentPower), 5*physic.KiloHertz)
	s.GreenPin.PWM(gpio.Duty(float64(s.BaseSetting.G)*s.CurrentPower), 5*physic.KiloHertz)
	s.WhitePin.PWM(gpio.Duty(float64(s.BaseSetting.W)*s.CurrentPower), 5*physic.KiloHertz)
	s.BluePin.PWM(gpio.Duty(float64(s.BaseSetting.B)*s.CurrentPower), 5*physic.KiloHertz)
}

// StartWakeup causes the lights to start coming on
func (s *Service) StartWakeup() {
	s.wakeupStarted = time.Now()
	s.wakeupInProgress = time.NewTimer(30 * time.Minute) // TODO: use a config setting for the wakeup duration
	s.wakeupTicker = time.NewTicker(30 * time.Second)
	s.CurrentPower = 0.0
	go s.tickWakeup()
}

// Halt causes the Wakeup program to halt (if it was running to begin with)
func (s *Service) Halt(desiredPower PowerSetting) {
	if s.wakeupInProgress != nil {
		s.haltWakeup <- desiredPower
	}
}

// tickWakeup is run as a background goroutine.
// It will drive the lights as appropriate, including the wakeup sequence.
func (s *Service) tickWakeup() {
	for {
		select {
		case pwr := <-s.haltWakeup:
			switch pwr {
			case POWER_FULL:
				s.CurrentPower = 1.0
			case POWER_OFF:
				s.CurrentPower = 0.0
			case POWER_LOW:
				s.CurrentPower = s.DimPowerSetting
			default:
				s.CurrentPower = 1.0
			}
			s.driveLights()
			// cleanup
			s.wakeupStarted = time.Time{}
			s.wakeupInProgress.Stop()
			s.wakeupInProgress = nil
			s.wakeupTicker = nil
			return
		case <-s.wakeupInProgress.C:
			// Wakeup sequence has completed
			s.CurrentPower = 1.0
			s.driveLights()
			// cleanup
			s.wakeupStarted = time.Time{}
			s.wakeupTicker.Stop()
			s.wakeupTicker = nil
			s.wakeupInProgress = nil
			return
		case <-s.wakeupTicker.C:
			// Wakeup sequence is in progress
			// Calculate what % of the way through the sequence we are
			secondsElapsed := time.Since(s.wakeupStarted).Seconds()
			secondsTotal := 30.0 * 60.0
			s.CurrentPower = secondsElapsed / secondsTotal
			s.driveLights()
		}
	}
}
