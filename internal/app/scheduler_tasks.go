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
	"github.com/jmaurer1994/gofish-bot/internal/weather"
)

func (app *Config) registerSchedulerTasks() {
	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "source:screenshot:save",
		Enabled:    false,
		Interval:   time.Duration(30) * time.Second,
		F:          app.SavePondCameraScreenshot,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "channel:overlay:update",
		Enabled:    true,
		Interval:   time.Duration(5) * time.Minute,
		F:          app.UpdateOverlay,
		RunAtStart: true,
	})

	app.Scheduler.RegisterTask(scheduler.Task{
		T:          "channel:overlay:feeder",
		Enabled:    true,
		Interval:   time.Duration(15) * time.Minute,
		F:          app.UpdateFeederCapacity,
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

func (app *Config) UpdateOverlay() {
	w, err := app.OwmApi.GetCurrentCondiitons()

	if err != nil {
		log.Printf("Error retrieving current conditions: %v\n", err)
		return
	}

	currentTime := int(time.Now().UnixMilli() / 1000)

	var (
		nextSunEvent          string
		nextSunEventCountdown SunEventCountdown
		phaseText             string
		phaseIcon             string
	)

	riseTime := w.Current.Sunrise - currentTime
	setTime := w.Current.Sunset - currentTime
	if riseTime > 0 {
		//currently >12am and counting down until sunrise
		nextSunEvent = "sunrise"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Current.Sunrise)

		if nextSunEventCountdown.hours < 1 && nextSunEventCountdown.minutes <= 10 {
			log.Println("Generating camera light off event")
			app.Scheduler.GenerateEvent("camera:light:check", "off")
		}
	} else if setTime > 0 {
		nextSunEvent = "moonrise"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Current.Sunset)

		if nextSunEventCountdown.hours < 1 && nextSunEventCountdown.minutes == 0 {
			log.Println("Generating camera light on event")
			app.Scheduler.GenerateEvent("camera:light:check", "on")
		}

	} else {
		//get tomorrow's sunrise
		nextSunEvent = "sunrise"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Daily[1].Sunrise)
	}

	phaseText, phaseIcon, err = weather.LunarPhaseValueToString(w.Daily[0].MoonPhase)
	if err != nil {
		log.Printf("Could not get get lunar phase icon: %v\n", err)
		return
	}

	conditions := make([]string, 0)
	for _, condition := range w.Current.Weather {
		conditions = append(conditions, condition.Icon)
	}

	log.Println("Updating weather")

	app.Event.RenderSSE("weather", components.WeatherWidget(conditions, fmt.Sprintf("%.0f", w.Current.Temp), fmt.Sprintf("%.1f", weather.FToC(w.Current.Temp)), fmt.Sprintf("%d", w.Current.Humidity), phaseText, phaseIcon))

	log.Printf("Time: %02d %02d", nextSunEventCountdown.hours, nextSunEventCountdown.minutes)

	app.Event.RenderSSE("time", components.TimeWidget(fmt.Sprintf("%02d", nextSunEventCountdown.hours), fmt.Sprintf("%02d", nextSunEventCountdown.minutes), nextSunEvent))
}

type SunEventCountdown struct {
	hours   int
	minutes int
}

func getTimeUntil(start int, target int) SunEventCountdown {
	minutes := ((target - start) % 3600) / 60
	approxMinutes := minutes - (minutes % 5)
	return SunEventCountdown{
		hours:   (target - start) / 3600,
		minutes: approxMinutes,
	}
}

func (app *Config) SavePondCameraScreenshot() {
	if app.Camera.CurrentLightLevel() > 0 {
		return //light on, don't take screenshot
	}

	err := app.Obs.ScreenshotSource("PondCamera")

	if err != nil {
		log.Printf("Could not save screenshot: %v\n", err)
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

func (app *Config) UpdateFeederCapacity() {
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
	log.Printf("Scale value: %f\n", f)
	capacity := (f / 1500.00) * 100
	log.Printf("Feeder capacity: %f %.0f%%\n", capacity, capacity)
	app.Event.RenderSSE("feeder", components.FeederWidget(capacity))
}
