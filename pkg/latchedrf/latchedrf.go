package latchedrf

import (
	"context"
	"errors"
	"github.com/klaital/wannetiot/pkg/util"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	_ "periph.io/x/host/v3/bcm283x"
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
			log.WithField("pin", r.Channels[i].String()).Debug("Initialized RF signal pin")
		}
	}
	return r.LatchResetPin != nil
}

var ErrPinNotRegistered = errors.New("pin not registered")

func initInputPin(pinId string) (gpio.PinIn, error) {
	pin := gpioreg.ByName(pinId)
	if pin == nil {
		return nil, ErrPinNotRegistered
	}
	return pin, pin.In(gpio.PullDown, gpio.RisingEdge)
}
func New(resetPin, a, b, c, d string) (*LatchedRadioReceiver, error) {
	var err error
	gpioA, err := initInputPin(a)
	if err != nil {
		log.WithField("pin", a).WithError(err).Error("Failed to initialize RF Pin")
		return nil, err
	} else {
		log.WithField("pin", a).Debug("RF Pin ready")
	}
	gpioB, err := initInputPin(b)
	if err != nil {
		log.WithField("pin", b).WithError(err).Error("Failed to initialize RF Pin")
		return nil, err
	} else {
		log.WithField("pin", b).Debug("RF Pin ready")
	}
	gpioC, err := initInputPin(c)
	if err != nil {
		log.WithField("pin", c).WithError(err).Error("Failed to initialize RF Pin")
		return nil, err
	} else {
		log.WithField("pin", c).Debug("RF Pin ready")
	}
	gpioD, err := initInputPin(d)
	if err != nil {
		log.WithField("pin", d).WithError(err).Error("Failed to initialize RF Pin")
		return nil, err
	} else {
		log.WithField("pin", d).Debug("RF Pin ready")
	}

	gpioR := gpioreg.ByName(resetPin)
	if gpioR == nil {
		return nil, errors.New("failed to register pin R")
	}
	err = gpioR.Out(gpio.Low)
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
