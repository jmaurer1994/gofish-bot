package main

import (
	"fmt"
	"github.com/jmaurer1994/gofish/bot/scheduler"
	"github.com/jmaurer1994/gofish/bot/weather"
	"log"
	"time"
)

func UpdateChannelTitle(s *scheduler.Scheduler) {
	w, err := owm.GetCurrentCondiitons()

	if err != nil {
		log.Printf("Error retrieving current conditions: %v\n", err)
		return
	}

	currentTime := int(time.Now().UnixMilli() / 1000)

	var (
		nextSunEvent          string
		nextSunEventCountdown *SunEventCountdown
		lunarPhaseIcon        string
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
		nextSunEvent = "sunset"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Current.Sunset)

		if nextSunEventCountdown.hours < 1 && nextSunEventCountdown.minutes <= 10 {
			log.Println("Generating camera light on event")
			s.GenerateEvent("camera:light:check", "on")
		}

	} else {
		//get tomorrow's sunrise
		nextSunEvent = "sunrise"
		nextSunEventCountdown = getTimeUntil(currentTime, w.Daily[1].Sunrise)
	}

	celcius := weather.FToC(w.Current.Temp)
	conditionIcon := weather.GetConditionIcon(w.Current.Weather[0].Icon)

	lunarPhaseIcon, err = weather.LunarPhaseValueToEmoji(w.Daily[0].MoonPhase)
	if err != nil {
		log.Printf("Could not get get lunar phase icon: %v\n", err)
		return
	}

	conditionsString := fmt.Sprintf("[%s %s(%.0f°F/%.1f°C)]", conditionIcon, lunarPhaseIcon, w.Current.Temp, celcius)

	sunEventString := fmt.Sprintf("[%dh %dm until %s]", nextSunEventCountdown.hours, nextSunEventCountdown.minutes, nextSunEvent)
	newTitle := "go🐟fish " + conditionsString + sunEventString

	err = tac.SetChannelTitle(newTitle)

	if err != nil {
		log.Printf("Could not set channel title, %v\n", err)
		return
	}

	log.Printf("<Channel title updated>[%s]\n", newTitle)
}

type SunEventCountdown struct {
	hours   int
	minutes int
}

func getTimeUntil(start int, target int) *SunEventCountdown {
	minutes := ((target - start) % 3600) / 60
	approxMinutes := minutes - (minutes % 5)
	return &SunEventCountdown{
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
	gc.ToggleSourceVisibility("Main", "PondCamera")
	log.Printf("Toggled camera source\n")
}
