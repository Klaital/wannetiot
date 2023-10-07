package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"

	"github.com/klaital/wannetiot/pkg/config"
	"github.com/klaital/wannetiot/pkg/lights"
)

type Server struct {
	Addr   string `env:"ADDR" envDefault:":8080"`
	Logger *logrus.Logger
	app    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	var srv Server
	srv.Logger = logrus.New()
	srv.Logger.SetLevel(logrus.DebugLevel)
	//srv.Logger.SetFormatter(&logrus.JSONFormatter{
	//	PrettyPrint: true,
	//})
	err := env.Parse(&srv)
	if err != nil {
		srv.Logger.WithError(err).Fatal("Failed to load env")
	}

	srv.app = cfg

	return &srv
}

//
//func (s *Server) LoggingHandler() http.Handler {
//
//}

type PagerNotice struct {
	Callback struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"callback"`
}
type LightsRequest struct {
	Dimmer     *float64 `json:"dimmer"` // value between 0.0 and 1.0, with 1.0 being full power.
	ColorPower *struct {
		Red   string
		Green string
		White string
		Blue  string
	} `json:"colors"`
}

func (srv *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	// Set up router
	pathTokens := strings.Split(req.RequestURI, "/")
	if len(pathTokens) <= 1 {
		srv.Logger.WithField("path", req.RequestURI).Error("invalid path")
		http.Error(resp, "invalid path", http.StatusInternalServerError)
		return
	}

	b, bodyReadErr := io.ReadAll(req.Body)

	// Request logging
	srv.Logger.WithFields(logrus.Fields{
		"method": req.Method,
		"path":   req.RequestURI,
		"host":   req.Host,
		"body":   string(b),
	}).Debug("Request received")

	switch pathTokens[1] {
	case "pager":
		if bodyReadErr != nil {
			resp.WriteHeader(400)
			return
		}
		var p PagerNotice
		err := json.Unmarshal(b, &p)
		if err != nil {
			resp.WriteHeader(500)
			return
		}

		// make an asynchronous call to Slack, then ping the control panel
		// to signal receipt
		go func() {
			// TODO: actually make the slack api call
			callbackUrl := fmt.Sprintf("http://%s:%d/pager/ack", p.Callback.Host, p.Callback.Port)
			resp, err := http.Get(callbackUrl)
			if err != nil {
				srv.Logger.WithField("callback", callbackUrl).WithError(err).Error("Failed to ack pager request")
			}
			srv.Logger.WithFields(logrus.Fields{
				"code":   resp.StatusCode,
				"status": resp.Status,
			}).Debug("Got pager ack callback response")
		}()

		resp.WriteHeader(200)
		return

	case "lights":
		//if len(pathTokens) < 3 {
		//	srv.Logger.Error("no light state given")
		//	http.Error(resp, "no light state given", http.StatusBadRequest)
		//	return
		//}
		var lightState lights.LightConfig
		switch pathTokens[2] {
		case "toggle":
			lights.HaltWakeup()
			srv.Logger.Debug("Wakeup halted, toggling lights")
			// TODO: lookup current light state.
			// TODO: advance lights one level in cycle: on -> dim -> off -> on
			resp.WriteHeader(200)
		case "off":
			lights.HaltWakeup()
			srv.Logger.Debug("Wakeup halted, turning lights off")
			lightState = lights.Off
			resp.WriteHeader(200)
		case "on":
			lights.HaltWakeup()
			srv.Logger.Debug("Wakeup halted, turning lights on")
			lightState = lights.LightSettingsFull()
			resp.WriteHeader(200)
		case "dim":
			lights.HaltWakeup()
			srv.Logger.Debug("Wakeup halted, dimming lights")
			lightState = lights.LightSettingLow()
			resp.WriteHeader(200)
		case "configure":
			var lightsReq LightsRequest
			if bodyReadErr != nil {
				srv.Logger.WithError(bodyReadErr).Error("Unable to read light configuration request")
				resp.WriteHeader(400)
				return
			}
			err := json.Unmarshal(b, &lightsReq)
			if err != nil {
				srv.Logger.WithError(err).Error("Unable to unmarshal light configuration request")
				resp.WriteHeader(400)
				return
			}

			if lightsReq.Dimmer != nil {
				lights.LightSettingLow()
			} else if lightsReq.ColorPower != nil {
				if lightState.R, err = gpio.ParseDuty(lightsReq.ColorPower.Red); err != nil {
					srv.Logger.WithField("raw", lightsReq.ColorPower.Red).WithError(err).Error("Failed to parse requested Red duty cycle")
				}
				if lightState.G, err = gpio.ParseDuty(lightsReq.ColorPower.Green); err != nil {
					srv.Logger.WithField("raw", lightsReq.ColorPower.Red).WithError(err).Error("Failed to parse requested Green duty cycle")
				}
				if lightState.W, err = gpio.ParseDuty(lightsReq.ColorPower.White); err != nil {
					srv.Logger.WithField("raw", lightsReq.ColorPower.Red).WithError(err).Error("Failed to parse requested White duty cycle")
				}
				if lightState.B, err = gpio.ParseDuty(lightsReq.ColorPower.Blue); err != nil {
					srv.Logger.WithField("raw", lightsReq.ColorPower.Red).WithError(err).Error("Failed to parse requested Blue duty cycle")
				}
				// TODO: write color settings to a config file to be loaded at startup
			}

			srv.Logger.WithField("cfg", lightState).Debug("Configured lights")
			resp.WriteHeader(204)
		case "wakeup":
			srv.Logger.Debug("Starting wakeup")
			lights.DoWakeup()
			resp.WriteHeader(200)
		default:
			srv.Logger.WithField("state", pathTokens[2]).Error("invalid light state")
			http.Error(resp, "invalid light state", http.StatusBadRequest)
			return
		}
		globalState.LightState = lightState
		lights.DriveLights(srv.app, &globalState.LightState)
	}
}
