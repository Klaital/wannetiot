package lights

import (
	"context"
	"github.com/klaital/wannetiot/pkg/config"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"time"
)

// RunWakeup will turn the lights off, then gradually dim them to 100% brightness over the course of 30 minutes.
func RunWakeup(ctx context.Context, cfg *config.Config, targetSettings LightConfig) {
	// how much to change the lights every second
	settingsDelta := LightConfig{
		Name: "",
		R:    gpio.Duty(float64(targetSettings.R) / cfg.WakeupDuration.Minutes()),
		G:    gpio.Duty(float64(targetSettings.G) / cfg.WakeupDuration.Minutes()),
		W:    gpio.Duty(float64(targetSettings.W) / cfg.WakeupDuration.Minutes()),
		B:    gpio.Duty(float64(targetSettings.B) / cfg.WakeupDuration.Minutes()),
	}

	t := time.NewTicker(1 * time.Minute)
	wakeupTimer = time.NewTimer(cfg.WakeupDuration)
	current := LightConfig{
		Name: "WAKEUP",
		R:    0,
		G:    0,
		W:    0,
		B:    0,
	}
	cfg.Logger.WithFields(log.Fields{
		"delta":  settingsDelta,
		"target": targetSettings,
	}).Debug("Starting wakeup lights")
	DriveLights(cfg, &current)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-wakeupTimer.C:
			cfg.Logger.Info("Wakeup lights complete")
			DriveLights(cfg, &targetSettings)
			t.Stop()
			return
		case <-t.C:
			// New Settings
			AddSettings(&current, &settingsDelta)

			DriveLights(cfg, &current)
		case <-haltWakeup:
			t.Stop()
			wakeupTimer.Stop()
			return
		}
	}
}

var startWakeup chan bool
var haltWakeup chan bool
var wakeupTimer *time.Timer

func StartWakeupRunner(ctx context.Context, cfg *config.Config) {
	wakeupTimer = time.NewTimer(cfg.WakeupDuration)
	startWakeup = make(chan bool)
	haltWakeup = make(chan bool)
	cfg.Logger.Info("Starting background job: wakeup lights runner")
	for {
		select {
		case <-ctx.Done():
			cfg.Logger.Info("Halting wakeup lights process")
			return
		case <-startWakeup:
			cfg.Logger.Debug("Signalling wakeup lights to start")
			RunWakeup(ctx, cfg, LightSettingsFull())
		}
	}
}

func DoWakeup() {
	startWakeup <- true
}
func HaltWakeup() {
	if wakeupTimer != nil {
		wakeupTimer.Stop()
		//haltWakeup <- true
	}
}
