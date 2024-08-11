package app

import (
	"fmt"
	"io"
	"log"
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
		F:          app.SavePondCameraScreenshot,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "data:weather:update",
		Enabled:    true,
		Interval:   time.Duration(5) * time.Minute,
		F:          app.OwmUpdate,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "data:weight:update",
		Enabled:    true,
		Interval:   time.Duration(15) * time.Minute,
		F:          app.UpdateFeederWeight,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "channel:reader:check",
		Enabled:    true,
		Interval:   time.Duration(1) * time.Hour,
		F:          app.CheckReaderStatus,
		RunAtStart: false,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "source:camera:cycle",
		Enabled:    true,
		Interval:   time.Duration(4) * time.Hour,
		F:          app.ResetCamera,
		RunAtStart: false,
	})
}

func (app *Config) OwmUpdate() {
	w, err := app.OwmApi.GetCurrentCondiitons()

	if err != nil {
		log.Printf("Error retrieving current conditions: %v\n", err)
		return
	}
	app.Data.Weather = w
	app.Data.Countdown = NewCountdown(w)

	switch {
	case app.Data.Countdown.Target == "sunrise":
		app.Scheduler.GenerateEvent("camera:light:check", "on")
	case app.Data.Countdown.Target == "moonrise":
		app.Scheduler.GenerateEvent("camera:light:check", "off")
	}

	app.Overlay.Render("weather", components.WeatherWidget(w))
	app.Overlay.Render("countdown", components.CountdownWidget(app.Data.Countdown.Hours(), app.Data.Countdown.Minutes(), app.Data.Countdown.Target))
}

func (app *Config) SavePondCameraScreenshot() {
	if app.Camera.CurrentLightLevel() > 0 {
		return //light on, don't take screenshot
	}

	fn := fmt.Sprintf("%d.png", time.Now().Unix())
	err := app.Obs.ScreenhotToBucket("PondCamera", fn, "pond-cam", app.S3)
	if err != nil {
		log.Printf("[Task] Error saving object to storage: %v\n", err)
	}
}

func (app *Config) CheckReaderStatus() {
	var _, connected = app.TwitchIrc.ReaderIsConnected()

	if !connected {
		log.Printf("!! Reader is not connected !!\n")
		if err := app.TwitchIrc.ConnectToChannel(); err != nil {
			log.Printf("Error while reconnecting to channel!:\n%v\n", err)
		}
	}
}

func (app *Config) ResetCamera() {
	if err := app.TwitchIrc.Sendf("Resetting camera... We'll be back in a moment!"); err != nil {
		log.Printf("Unable to send camera reset notification: %v\n", err)
	}
	app.Obs.ToggleSourceVisibility("Main", "PondCamera")
	log.Printf("Toggled camera source\n")
}

func (app *Config) UpdateFeederWeight() {
	resp, err := http.Get("https://sensor.gofish.cam/scale/read?samples=10")

	if err != nil {
		log.Printf("Error updating feeder capacity: %v\n", err)
		return
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)

	if readErr != nil {
		log.Printf("Error updating feeder capacity2: %v\n", err)
		return
	}

	str := string(body)

	f, convErr := strconv.ParseFloat(str, 64)

	if convErr != nil {
		log.Printf("Error updating feeder capacity due to conversion error: %v\n", convErr)
		return
	}

	app.Data.FeederWeight = f

	app.Overlay.Render("feeder", components.FeederWidget(f))
}
