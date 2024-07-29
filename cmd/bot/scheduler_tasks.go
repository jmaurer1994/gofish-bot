package main

import (
	"fmt"
	"github.com/jmaurer1994/gofish-bot/internal/overlay/views/components"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
	"log"
	"time"
)

func registerSchedulerTasks(sch *scheduler.Scheduler) {
	sch.RegisterTask(scheduler.Task{
		T:          "source:screenshot:save",
		Enabled:    false,
		Interval:   time.Duration(30) * time.Second,
		F:          SavePondCameraScreenshot,
		RunAtStart: true,
	})

	sch.RegisterTask(scheduler.Task{
		T:          "channel:title:update",
		Enabled:    true,
		Interval:   time.Duration(5) * time.Minute,
		F:          UpdateOverlay,
		RunAtStart: true,
	})

	sch.RegisterTask(scheduler.Task{
		T:          "channel:reader:check",
		Enabled:    true,
		Interval:   time.Duration(1) * time.Hour,
		F:          CheckReaderStatus,
		RunAtStart: false,
	})

	sch.RegisterTask(scheduler.Task{
		T:          "source:camera:cycle",
		Enabled:    true,
		Interval:   time.Duration(4) * time.Hour,
		F:          ResetCamera,
		RunAtStart: false,
	})
}

func UpdateOverlay(s *scheduler.Scheduler) {
	w, err := owm.GetCurrentCondiitons()

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
			s.GenerateEvent("camera:light:check", "off")
		}
	} else if setTime > 0 {
		nextSunEvent = "moonrise"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Current.Sunset)

		if nextSunEventCountdown.hours < 1 && nextSunEventCountdown.minutes == 0 {
			log.Println("Generating camera light on event")
			s.GenerateEvent("camera:light:check", "on")
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

	event.RenderSSE("weather", components.WeatherWidget(conditions, fmt.Sprintf("%.0f", w.Current.Temp), fmt.Sprintf("%.1f", weather.FToC(w.Current.Temp)), fmt.Sprintf("%d", w.Current.Humidity), phaseText, phaseIcon))
	event.RenderSSE("time", components.TimeWidget(fmt.Sprintf("%02d", nextSunEventCountdown.hours), fmt.Sprintf("%02d", nextSunEventCountdown.minutes), nextSunEvent))
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

func SavePondCameraScreenshot(s *scheduler.Scheduler) {
	if c.CurrentLightLevel() > 0 {
		return //light on, don't take screenshot
	}

	err := gc.ScreenshotSource("PondCamera")

	if err != nil {
		log.Printf("Could not save screenshot: %v\n", err)
	}
}

func CheckReaderStatus(s *scheduler.Scheduler) {
	var _, connected = tic.ReaderIsConnected()

	if !connected {
		log.Printf("!! Reader is not connected !!\n")
		if err := tic.ConnectToChannel(); err != nil {
			log.Printf("Error while reconnecting to channel!:\n%v\n", err)
		}
	}
}

func ResetCamera(s *scheduler.Scheduler) {
	if err := tic.Sendf("Resetting camera... We'll be back in a moment!"); err != nil {
		log.Printf("Unable to send camera reset notification: %v\n", err)
	}
	gc.ToggleSourceVisibility("Main", "PondCamera")
	log.Printf("Toggled camera source\n")
}
