package app

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
)

func (app *Config) registerSchedulerEvents() {
	app.Scheduler.RegisterEventHandler("camera:light:check", app.handleCameraLightCheck)
	app.Scheduler.RegisterEventHandler("SensorEvent:Insert", app.handleDatabaseEvent)
}

func (app *Config) handleCameraLightCheck(m scheduler.Message) {
	log.Printf("Received light check event: %s\n", m)
	switch m {
	case "on":
		if app.Camera.CurrentLightLevel() == 0 {
			app.Camera.SetLightLevel(1)
		}
	case "off":
		if app.Camera.CurrentLightLevel() > 0 {
			app.Camera.ZeroLight()
		}
	}
}

type SensorEventPayload struct {
	Event_ID  int   `json:"event_id"`
	Timestamp int   `json:"timestamp"`
	Samples   []int `json:"samples"`
}

func (app *Config) handleDatabaseEvent(m scheduler.Message) {
	log.Printf("Received sensor event\n")
	var payload SensorEventPayload
	if err := json.Unmarshal([]byte(m), &payload); err != nil {
		log.Printf("Error while unmarshalling payload: %v\n", err)
		return
	}
	var pm int
	for i, e := range payload.Samples {
		if i == 0 || e > pm {
			pm = e
		}
	}

	app.TwitchIrc.SendChannelMessage(fmt.Sprintf("Food dispensed! Force of string pull: %d", pm))
}
