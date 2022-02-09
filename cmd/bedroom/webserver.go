package main

import (
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"github.com/klaital/wannetiot/pkg/config"
	"github.com/klaital/wannetiot/pkg/lights"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
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
	srv.Logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
	err := env.Parse(&srv)
	if err != nil {
		srv.Logger.WithError(err).Fatal("Failed to load env")
	}

	srv.app = cfg

	return &srv
}

type Telemetry struct {
	Dimmer int `json:"Dimmer"`
	WiFi   int `json:"WiFi"`
}
type PagerNotice struct {
	Callback struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"callback"`
}

func (srv *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// Set up router
	pathTokens := strings.Split(req.RequestURI, "/")
	if len(pathTokens) <= 1 {
		srv.Logger.WithField("path", req.RequestURI).Error("invalid path")
		http.Error(resp, "invalid path", http.StatusInternalServerError)
		return
	}

	switch pathTokens[1] {
	case "dimmer":
		bodyBytes, err := ioutil.ReadAll(req.Body)
		var telem Telemetry
		defer req.Body.Close()
		if err != nil {
			srv.Logger.WithError(err).Error("Unable to read body")
			http.Error(resp, "", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bodyBytes, &telem)
		if err != nil {
			srv.Logger.WithError(err).Error("Unable to deserialize body")
			http.Error(resp, "", http.StatusBadRequest)
			return
		}
		srv.Logger.WithField("telem", telem).Debug("Got telemetry from control panel")
	case "lights":
		if len(pathTokens) < 3 {
			srv.Logger.Error("no light state given")
			http.Error(resp, "no light state given", http.StatusBadRequest)
			return
		}
		var lightState lights.LightConfig
		switch pathTokens[2] {
		case "off":
			lights.HaltWakeup()
			lightState = lights.LightsOff
			resp.WriteHeader(200)
		case "on":
			lights.HaltWakeup()
			srv.Logger.Debug("Wakeup halted, turning lights on")
			lightState = lights.LightSettingsFull()
			resp.WriteHeader(200)
		case "dim":
			lights.HaltWakeup()
			lightState = lights.LightSettingLow()
			resp.WriteHeader(200)
		case "wakeup":
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
