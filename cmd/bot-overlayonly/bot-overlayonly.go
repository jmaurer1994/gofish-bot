package main

import (
	"github.com/jmaurer1994/gofish-bot/internal/overlay"
	"github.com/jmaurer1994/gofish-bot/internal/overlay/views/components"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	event *overlay.Event
)

func main() {
	event = overlay.StartOverlay()
	// Create a channel to receive os.Signal values.operator
	sigs := make(chan os.Signal, 1)
	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Starting overlay")
	go timeUpdate()
	<-sigs
}

func timeUpdate() {
	for {

		time.Sleep(time.Second * 100)
		//		currentTime := fmt.Sprintf("The Current Time Is %v", now)

		// Send current time to clients message channel
		component := components.TimeWidget("10", "30", "moonrise")
		event.RenderSSE("weather", components.WeatherWidget([]string{"01d"}, "72", "23.5", "55", "Full Moon", "moon-full"))
		event.RenderSSE("time", component)
	}
}
