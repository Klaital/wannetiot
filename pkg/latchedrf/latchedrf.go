package latchedrf

import (
	"context"
	"errors"
	"github.com/klaital/wannetiot/pkg/util"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"time"
)

type LatchedRadioReceiver struct {
	LatchResetPin gpio.PinOut
	Channels      []gpio.PinIn
	handlers      []Handler
	WaitTimeout   time.Duration
	stop          chan bool
}

type Handler func()

func (r *LatchedRadioReceiver) IsValid() bool {
	for i := range r.Channels {

		if r.Channels[i] == nil {
			log.WithField("pin", i).Error("Failed to initialize pin")
			return false
		} else {
			log.WithField("pin", r.Channels[i].String()).Debug("Initialized pin")
		}
	}
	return r.LatchResetPin != nil
}

func New(resetPin, a, b, c, d string) (*LatchedRadioReceiver, error) {
	var err error
	gpioA := gpioreg.ByName(a)
	if gpioA == nil {
		return nil, errors.New("failed to register pin A")
	}
	err = gpioA.In(gpio.PullNoChange, gpio.RisingEdge)
	if err != nil {
		log.WithError(err).Error("Failed to initialize pin A")
		return nil, err
	}
	log.Debug("RF Channel A ready")

	gpioB := gpioreg.ByName(b)
	if gpioB == nil {
		return nil, errors.New("failed to register pin B")
	}
	err = gpioB.In(gpio.PullNoChange, gpio.RisingEdge)
	if err != nil {
		log.WithError(err).Error("Failed to initialize pin B")
		return nil, err
	}
	log.Debug("RF Channel B ready")

	gpioC := gpioreg.ByName(c)
	if gpioA == nil {
		return nil, errors.New("failed to register pin C")
	}
	err = gpioC.In(gpio.PullNoChange, gpio.RisingEdge)
	if err != nil {
		log.WithError(err).Error("Failed to initialize pin C")
		return nil, err
	}
	log.Debug("RF Channel C ready")

	gpioD := gpioreg.ByName(d)
	if gpioD == nil {
		return nil, errors.New("failed to register pin D")
	}
	err = gpioD.In(gpio.PullNoChange, gpio.RisingEdge)
	if err != nil {
		log.WithError(err).Error("Failed to initialize pin D")
		return nil, err
	}
	log.Debug("RF Channel D ready")

	gpioR := gpioreg.ByName(resetPin)
	if gpioR == nil {
		return nil, errors.New("failed to register pin R")
	}
	err = gpioA.Out(gpio.Low)
	if err != nil {
		log.WithError(err).Error("Failed to initialize pin R")
		return nil, err
	}

	r := LatchedRadioReceiver{
		LatchResetPin: gpioR,
		Channels: []gpio.PinIn{
			gpioA,
			gpioB,
			gpioC,
			gpioD,
		},
		handlers:    make([]Handler, 4),
		WaitTimeout: 1 * time.Second,
	}
	if !r.IsValid() {
		return nil, errors.New("failed to initialize all latch pins")
	}

	if err = r.LatchResetPin.Out(gpio.Low); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"pin":   "LatchResetPin",
			"pinid": r.LatchResetPin.String(),
		}).Error("Failed to initialize pin")
		return nil, err
	}
	//for i, c := range r.Channels {
	//	if err = c.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
	//		log.WithError(err).WithFields(log.Fields{
	//			"channel": i,
	//			"pin":     c.String(),
	//		}).Error("Failed to initialize pin")
	//		return nil, err
	//	}
	//}

	return &r, nil
}

func (r *LatchedRadioReceiver) ReadChannels() (a, b, c, d bool) {
	return bool(r.Channels[0].Read()), bool(r.Channels[1].Read()), bool(r.Channels[2].Read()), bool(r.Channels[3].Read())
}

func (r *LatchedRadioReceiver) Reset() {
	util.Flash(r.LatchResetPin, 1*time.Microsecond, nil)
}

func (r *LatchedRadioReceiver) RegisterChannelAHandler(handlerFunc Handler) {
	r.handlers[0] = handlerFunc
}
func (r *LatchedRadioReceiver) RegisterChannelBHandler(handlerFunc Handler) {
	r.handlers[1] = handlerFunc
}
func (r *LatchedRadioReceiver) RegisterChannelCHandler(handlerFunc Handler) {
	r.handlers[2] = handlerFunc
}
func (r *LatchedRadioReceiver) RegisterChannelDHandler(handlerFunc Handler) {
	r.handlers[3] = handlerFunc
}

// Run sets up watches on each of the input channels.
// When one is triggered, the registered handler is executed, then the latch is reset.
func (r *LatchedRadioReceiver) Run(ctx context.Context) {
	handlePin := func(c gpio.PinIn, h Handler) {
		log.WithField("pin", c.Name()).Debug("Listening for edges on RF pin")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if c.WaitForEdge(r.WaitTimeout) {
					if h != nil {
						h()
					}
					r.Reset()
				}
			}
		}
	}

	for i := range r.Channels {
		go handlePin(r.Channels[i], r.handlers[i])
	}
}
