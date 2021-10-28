package latchedrf

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"iot-bedroom-pi/pkg/util"
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
	r := LatchedRadioReceiver{
		LatchResetPin: gpioreg.ByName(resetPin),
		Channels: []gpio.PinIn{
			gpioreg.ByName(a),
			gpioreg.ByName(b),
			gpioreg.ByName(c),
			gpioreg.ByName(d),
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
	for i, c := range r.Channels {
		if err = c.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"channel": i,
				"pin":     c.String(),
			}).Error("Failed to initialize pin")
			return nil, err
		}
	}

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
		for {
			select {
			case <- ctx.Done():
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
