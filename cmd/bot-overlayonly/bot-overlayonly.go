package main

import (
	"fmt"
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

		time.Sleep(time.Second * 10)
		now := time.Now().Format("2006-01-02 15:04:05")
		currentTime := fmt.Sprintf("The Current Time Is %v", now)

		// Send current time to clients message channel
		component := components.TimeWidget(currentTime)

		event.RenderSSE("time", component)
	}
}
