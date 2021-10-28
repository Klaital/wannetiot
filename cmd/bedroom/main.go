package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"iot-bedroom-pi/pkg/config"
	"iot-bedroom-pi/pkg/ctlpanel"
	"iot-bedroom-pi/pkg/latchedrf"
	"iot-bedroom-pi/pkg/lights"
	"iot-bedroom-pi/pkg/util"
	"os"
	"os/signal"
	"time"
)

type ControlState struct {
	LightState    lights.LightConfig
	ControlPanel1 ctlpanel.ControlPanel
	ControlPanel2 ctlpanel.ControlPanel
	RadioReceiver *latchedrf.LatchedRadioReceiver
}

var globalState ControlState = ControlState{
	LightState:    lights.LightsOff,
	ControlPanel1: ctlpanel.ControlPanel{},
	ControlPanel2: ctlpanel.ControlPanel{},
	RadioReceiver: nil,
}

var influxBuffer []util.InfluxDataPoint

func main() {
	var err error
	ctx, halt := context.WithCancel(context.Background())

	//log.SetLevel(log.DebugLevel)
	//log.SetFormatter(&log.JSONFormatter{
	//	PrettyPrint: true,
	//})

	influxBuffer = make([]util.InfluxDataPoint, 0, 10)

	cfg := config.New()
	logger := cfg.Logger.WithFields(log.Fields{
		"op": "main",
	})
	//cfg.InitPins()
	//defer cfg.HaltPins()

	// Initialize the control panels
	globalState.ControlPanel1 = ctlpanel.New(cfg.Panel1PagerPin, cfg.Panel1LightSwitchPin, cfg.Panel1ResetPin, cfg.Panel1LEDPin, cfg.Panel1SpeakerPin, cfg.Panel1DimmerChannel, cfg.ControlPanelsAdc, nil)
	globalState.ControlPanel2 = ctlpanel.New(cfg.Panel2PagerPin, cfg.Panel2LightSwitchPin, cfg.Panel2ResetPin, cfg.Panel2LEDPin, cfg.Panel2SpeakerPin, cfg.Panel2DimmerChannel, cfg.ControlPanelsAdc, nil)

	// Initialize the attached sensors
	//cfg.InitSensors()

	// Initialize the RF receiver
	globalState.RadioReceiver, err = latchedrf.New(cfg.RadioLatchResetPin, cfg.RadioChannelAPin, cfg.RadioChannelBPin, cfg.RadioChannelCPin, cfg.RadioChannelDPin)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"ResetPin": cfg.RadioLatchResetPin,
			"A": cfg.RadioChannelAPin,
			"B": cfg.RadioChannelBPin,
			"C": cfg.RadioChannelCPin,
			"D": cfg.RadioChannelDPin,
		}).Fatal("Failed to instantiate LatchedRadioReceiver")
	}
	globalState.RadioReceiver.WaitTimeout = cfg.RadioWaitTimeout
	globalState.RadioReceiver.RegisterChannelAHandler(func() {
		log.WithField("channel", "A").Debug("RF signal received")
		// TODO: implement handler that turns the lights on full power
		globalState.LightState = lights.LightSettingsFull()
		if cfg.LedStripEnabled {
			lights.DriveLights(cfg, globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelBHandler(func() {
		log.WithField("channel", "B").Debug("RF signal received")
		// TODO: implement handler that turns the lights on low power
		globalState.LightState = lights.LightSettingLow()
		if cfg.LedStripEnabled {
			lights.DriveLights(cfg, globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelCHandler(func() {
		log.WithField("channel", "C").Debug("RF signal received")
		// TODO: implement handler that turns the lights off
		globalState.LightState = lights.LightsOff
		if cfg.LedStripEnabled {
			lights.DriveLights(cfg, globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelDHandler(func() {
		log.WithField("channel", "D").Debug("RF signal received")
		// Handler that sends out pager notifications
		if cfg.Panel1Enabled {
			globalState.ControlPanel1.AcknowledgePager()
		}
		if cfg.Panel2Enabled {
			globalState.ControlPanel2.AcknowledgePager()
		}

		// TODO: send out slack notifications
	})

	globalState.RadioReceiver.Run(ctx)

	// Take initial readings from the sensors and record them in Influx
	if cfg.AM2302Enabled {
		temperature, humidity, err := cfg.AM2302Sensor.Read()
		if err != nil {
			log.WithError(err).Error("Failed to read from AM2302")
		} else {
			log.WithFields(log.Fields{
				"t": temperature,
				"h": humidity,
			}).Debug("Got initial atmo readings")
		}
		p := util.AtmoData{
			T:    temperature,
			H:    humidity,
			PM25: 0,
			PM10: 0,
			Ts:   time.Now(),
		}
		influxBuffer = append(influxBuffer, p)
		err = util.FlushInfluxBuffer(influxBuffer, cfg.GetInfluxDB())
		if err != nil {
			log.WithError(err).Error("Error flushing influx buffer")
		} else {
			log.Debug("Influx data buffer flushed")
			influxBuffer = make([]util.InfluxDataPoint, 0, 10)
		}
	}

	// Start polling the sensors
	sensorTicker := time.NewTicker(cfg.PollInterval)
	go pollSensors(ctx, sensorTicker, cfg)

	pagerNotice := make(chan uint8, 1)
	lightsNotice := make(chan uint8, 1)
	// Register interrupts for the pager and lightswitch buttons
	go globalState.ControlPanel1.StartPagerInterrupt(ctx, pagerNotice, 0)
	go globalState.ControlPanel2.StartPagerInterrupt(ctx, pagerNotice, 1)
	go globalState.ControlPanel1.StartLightsInterrupt(ctx, lightsNotice, 0)
	go globalState.ControlPanel2.StartLightsInterrupt(ctx, lightsNotice, 1)

	go func(cfg *config.Config) {
		var n uint8
		for {
			select {
			case n = <-pagerNotice:
				if n == 0 {
					globalState.ControlPanel1.AcknowledgePager()
				} else if n == 1 {
					globalState.ControlPanel2.AcknowledgePager()
				} else {
					log.Error("Invalid panel code")
				}
			case n = <-lightsNotice:
				// Change the lights brightness, Low -> High -> Off -> Low
				switch globalState.LightState.Name {
				case "LIGHTS_OFF":
					globalState.LightState = lights.LightSettingLow()
					lights.DriveLights(cfg, globalState.LightState)
				case "LIGHTS_LOW":
					globalState.LightState = lights.LightsFull
					lights.DriveLights(cfg, globalState.LightState)
				case "LIGHTS_HIGH":
					globalState.LightState = lights.LightsOff
					lights.DriveLights(cfg, globalState.LightState)
				}
			case <-ctx.Done():
				return
			}
		}
	}(cfg)

	// TODO: cancel the interrupts when a shutdown signal is received
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		logger.Info("Trapped Ctrl+C, shutting down")
		halt()
		os.Exit(0)
	}()
	// TODO: start webserver to listen for remote control commands
}

func pollSensors(ctx context.Context, ticker *time.Ticker, cfg *config.Config) {
	influxBuffer := make([]util.InfluxDataPoint, 0, cfg.InfluxBufferSize)
	var temperature, humidity float64
	var err error
	var atmoPoint util.AtmoData
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			temperature, humidity, err = cfg.AM2302Sensor.Read()
			atmoPoint = util.AtmoData{Ts: time.Now()}

			if err != nil {
				cfg.Logger.WithError(err).Error("Failed to read from AM2302 sensor")
			} else {
				atmoPoint.T = temperature
				atmoPoint.H = humidity
			}

			// TODO: poll the SDS011 Air Quality sensor as well

			influxBuffer = append(influxBuffer, atmoPoint)

			// Flush the buffer when it's full
			if len(influxBuffer) == cap(influxBuffer) {
				err = util.FlushInfluxBuffer(influxBuffer, cfg.GetInfluxDB())
				if err != nil {
					cfg.Logger.WithError(err).Error("Failed to write points to influx")
				}
			}
		}
	}
}
