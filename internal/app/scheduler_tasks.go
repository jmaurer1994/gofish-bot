package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jmaurer1994/gofish-bot/internal/app/views/components"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
)

func (app *Config) registerSchedulerTasks() {
	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "source:screenshot:save",
		Enabled:    true,
		Interval:   time.Duration(30) * time.Minute,
		Timeout:    time.Duration(2) * time.Minute,
		F:          app.SavePondCameraScreenshot,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "data:weather:update",
		Enabled:    true,
		Interval:   time.Duration(5) * time.Minute,
		Timeout:    time.Duration(2) * time.Minute,
		F:          app.OwmUpdate,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "data:weight:update",
		Enabled:    true,
		Interval:   time.Duration(15) * time.Minute,
		Timeout:    time.Duration(2) * time.Minute,
		F:          app.UpdateFeederWeight,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "channel:reader:check",
		Enabled:    true,
		Interval:   time.Duration(1) * time.Hour,
		Timeout:    time.Duration(2) * time.Minute,
		F:          app.CheckReaderStatus,
		RunAtStart: false,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "source:camera:cycle",
		Enabled:    true,
		Interval:   time.Duration(4) * time.Hour,
		Timeout:    time.Duration(2) * time.Minute,
		F:          app.ResetCamera,
		RunAtStart: false,
	})
}

var ConditionsUpdate = errors.New("Could not retrieve current conditions")

func (app *Config) OwmUpdate(t *scheduler.Task, ctx context.Context) error {
	w, err := app.OwmApi.GetCurrentCondiitons()

	if err != nil {
		return errors.Join(ConditionsUpdate, err)
	}
	app.Data.Weather = w
	app.Data.Countdown = NewCountdown(w)

	switch {
	case app.Data.Countdown.Target == "sunrise":
		app.Scheduler.GenerateEvent("camera:light:check", "on")
	case app.Data.Countdown.Target == "moonrise":
		app.Scheduler.GenerateEvent("camera:light:check", "off")
	}

	app.EventServer.SendEvent("weather", components.WeatherWidget(w))
	app.EventServer.SendEvent("countdown", components.CountdownWidget(app.Data.Countdown.Hours(), app.Data.Countdown.Minutes(), app.Data.Countdown.Target))

	return nil
}

var ScreenshotStorage = errors.New("Could not save screenshot")

func (app *Config) SavePondCameraScreenshot(t *scheduler.Task, ctx context.Context) error {
	if app.Camera.CurrentLightLevel() > 0 {
		return nil //light on, don't take screenshot
	}

	fn := fmt.Sprintf("%d", time.Now().Unix())
	err := app.Obs.ScreenhotToBucket("PondCamera", fn, "pond-cam", app.S3)
	if err != nil {
		return errors.Join(ScreenshotStorage, err)
	}

	return nil
}

func (app *Config) CheckReaderStatus(t *scheduler.Task, ctx context.Context) error {
	var _, connected = app.TwitchIrc.ReaderIsConnected()

	if !connected {
		t.Log("!! Reader is not connected !!\n")
		if err := app.TwitchIrc.ConnectToChannel(); err != nil {
			t.Log(fmt.Sprintf("Error while reconnecting to channel!:\n%v\n", err))
		}
	}

	return nil
}

var ResetNotification = errors.New("Unable to send reset notification")

func (app *Config) ResetCamera(t *scheduler.Task, ctx context.Context) error {
	if err := app.TwitchIrc.Sendf("Resetting camera... We'll be back in a moment!"); err != nil {
		return errors.Join(ResetNotification, err)
	}
	app.Obs.ToggleSourceVisibility("Main", "PondCamera")
	t.Log("Toggled camera source\n")

	return nil
}

var FeederWeightRequest = errors.New("Could not make request")
var FeederWeightResponse = errors.New("Unable to read response body")
var FeederWeightConversion = errors.New("Unable to convert response value to float")

func (app *Config) UpdateFeederWeight(t *scheduler.Task, tx context.Context) error {
	resp, err := http.Get("https://sensor.gofish.cam/scale/read?samples=10")

	if err != nil {
		return errors.Join(FeederWeightRequest, err)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)

	if readErr != nil {
		return errors.Join(FeederWeightResponse, err)
	}

	f, convErr := strconv.ParseFloat(string(body), 64)

	if convErr != nil {
		return errors.Join(FeederWeightConversion, err)
	}

	app.Data.FeederWeight = f

	app.EventServer.SendEvent("feeder", components.FeederWidget(f))
	return nil
}
