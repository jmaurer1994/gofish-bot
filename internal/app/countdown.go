package app

import (
	"time"

	"github.com/jmaurer1994/gofish-bot/internal/weather"
)

type Countdown struct {
	targetTime int
	Target     string
}

func NewCountdown(w weather.OneCallResponse) Countdown {
	currentSeconds := int(time.Now().UnixMilli() / 1000)
	minutesToSunRise := (w.Current.Sunrise - currentSeconds) / 60
	minutesToMoonRise := (w.Current.Sunset - currentSeconds) / 60

	target := ""
	var targetTime int

	switch {
	case minutesToSunRise >= 5: //[12am, sunrise)
		target = "sunrise"
		targetTime = w.Current.Sunrise
	case minutesToMoonRise >= 5: //[sunrise, sunset)
		target = "moonrise"
		targetTime = w.Current.Sunset
	default: //[sunset, 12am)
		target = "sunrise"
		targetTime = w.Daily[1].Sunrise
	}

	return Countdown{
		Target:     target,
		targetTime: targetTime,
	}
}

func (c Countdown) Hours() int {
	start := int(time.Now().UnixMilli() / 1000)
	return (c.targetTime - start) / 3600

}

func (c Countdown) Minutes() int {
	start := int(time.Now().UnixMilli() / 1000)

	minutes := ((c.targetTime - start) % 3600) / 60

	return minutes - (minutes % 5)
}
