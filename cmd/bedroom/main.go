package main

import (
	"context"
	"github.com/klaital/wannetiot/pkg/config"
	"github.com/klaital/wannetiot/pkg/ctlpanel"
	"github.com/klaital/wannetiot/pkg/latchedrf"
	"github.com/klaital/wannetiot/pkg/lights"
	"github.com/klaital/wannetiot/pkg/util"
	log "github.com/sirupsen/logrus"
	"net/http"
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

	cfg.InitPins()
	defer cfg.HaltPins()

	// Initialize the attached sensors
	cfg.InitSensors()

	// Initialize the RF receiver
	logger.Debug("Initializing RF Receiver")
	globalState.RadioReceiver, err = latchedrf.New(cfg.RadioLatchResetPin, cfg.RadioChannelAPin, cfg.RadioChannelBPin, cfg.RadioChannelCPin, cfg.RadioChannelDPin)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"ResetPin": cfg.RadioLatchResetPin,
			"A":        cfg.RadioChannelAPin,
			"B":        cfg.RadioChannelBPin,
			"C":        cfg.RadioChannelCPin,
			"D":        cfg.RadioChannelDPin,
		}).Fatal("Failed to instantiate LatchedRadioReceiver")
	}
	globalState.RadioReceiver.WaitTimeout = cfg.RadioWaitTimeout
	globalState.RadioReceiver.RegisterChannelAHandler(func() {
		logger.WithField("channel", "A").Debug("RF signal received")
		globalState.LightState = lights.LightSettingsFull()
		if cfg.LedStripEnabled {
			logger.WithField("lights", globalState.LightState).Debug("Driving new light settings")
			lights.HaltWakeup()
			lights.DriveLights(cfg, &globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelBHandler(func() {
		logger.WithField("channel", "B").Debug("RF signal received")
		globalState.LightState = lights.LightSettingLow()
		if cfg.LedStripEnabled {
			logger.WithField("lights", globalState.LightState).Debug("Driving new light settings")
			lights.HaltWakeup()
			lights.DriveLights(cfg, &globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelCHandler(func() {
		logger.WithField("channel", "C").Debug("RF signal received")
		globalState.LightState = lights.LightsOff
		if cfg.LedStripEnabled {
			logger.WithField("lights", globalState.LightState).Debug("Driving new light settings")
			lights.HaltWakeup()
			lights.DriveLights(cfg, &globalState.LightState)
		}
	})
	globalState.RadioReceiver.RegisterChannelDHandler(func() {
		logger.WithField("channel", "D").Debug("RF Pager signal received")
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
			logger.WithError(err).Error("Failed to read from AM2302")
		} else {
			logger.WithFields(log.Fields{
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
			logger.WithError(err).Error("Error flushing influx buffer")
		} else {
			logger.Debug("Influx data buffer flushed")
			influxBuffer = make([]util.InfluxDataPoint, 0, 10)
		}
	}

	// Start polling the sensors
	//sensorTicker := time.NewTicker(cfg.PollInterval)
	//go pollSensors(ctx, sensorTicker, cfg)

	//pagerNotice := make(chan uint8, 1)
	//lightsNotice := make(chan uint8, 1)
	//
	//go func(cfg *config.Config) {
	//	var n uint8
	//	for {
	//		select {
	//		case n = <-pagerNotice:
	//			cfg.Logger.WithField("channel", n).Debug("Pager received")
	//		case n = <-lightsNotice:
	//			cfg.Logger.WithField("channel", n).Debug("Updating lights state")
	//			// Change the lights brightness, Low -> High -> Off -> Low
	//			switch globalState.LightState.Name {
	//			case "LIGHTS_OFF":
	//				globalState.LightState = lights.LightsOff
	//				lights.DriveLights(cfg, &globalState.LightState)
	//			case "LIGHTS_LOW":
	//				globalState.LightState = lights.LightSettingLow()
	//				lights.DriveLights(cfg, &globalState.LightState)
	//			case "LIGHTS_HIGH":
	//				globalState.LightState = lights.LightsOff
	//				lights.DriveLights(cfg, &globalState.LightState)
	//			}
	//		case <-ctx.Done():
	//			return
	//		}
	//	}
	//}(cfg)

	// Start a background thread to handle the gradual wakup light routine
	go lights.StartWakeupRunner(ctx, cfg)

	// Start a webserver to listen for remote control commands
	webServer := &http.Server{Addr: ":8080", Handler: NewServer(cfg)}
	go func() {
		if err := webServer.ListenAndServe(); err != nil {
			cfg.Logger.WithError(err).Fatal("Failed to initialize webserver")
		}
	}()

	// Shut down the background handlers and webserver listener when a shutdown signal is received
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.Info("Trapped Ctrl+C, shutting down")
	halt()

	logger.Debug("Halting wakeup")
	lights.HaltWakeup()

	shutdownContext, forceShutdown := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer forceShutdown()
	logger.Debug("Shutting down webserver")
	if err = webServer.Shutdown(shutdownContext); err != nil {
		if err != http.ErrServerClosed {
			cfg.Logger.WithError(err).Fatalf("Error shutting down webserver: %t, %v", err, err)
		}
	}
	os.Exit(0)

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
