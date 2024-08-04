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
	currentTime := int(time.Now().UnixMilli() / 1000)
	riseTime := w.Current.Sunrise - currentTime
	setTime := w.Current.Sunset - currentTime
	target := ""
	var targetTime int

	switch {
	case riseTime > 0:
		target = "sunrise"

		targetTime = w.Current.Sunrise
	case setTime > 0:

		target = "moonrise"
		targetTime = w.Current.Sunset
	default:
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
