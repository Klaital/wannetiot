package main

import (
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type Server struct {
	Addr   string `env:"ADDR" envDefault:":8080"`
	Logger *logrus.Logger
}

func NewServer() *Server {
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

	if pathTokens[1] == "dimmer" {
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
	}
}
