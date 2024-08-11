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
	switch m {
	case "on":
		if app.Camera.CurrentLightLevel() == 0 {
			log.Println("[Camera] Turning light on")
			app.Camera.SetLightLevel(1)
		}
	case "off":
		if app.Camera.CurrentLightLevel() > 0 {
			log.Println("[Camera] Turning light off")
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
	var payload SensorEventPayload
	if err := json.Unmarshal([]byte(m), &payload); err != nil {
		log.Printf("[Database] Error while unmarshalling payload: %v\n", err)
		return
	}
	var pm int
	for i, e := range payload.Samples {
		if i == 0 || e > pm {
			pm = e
		}
	}
	if app.Camera.CurrentLightLevel() == 0 {
		fn := fmt.Sprintf("%d-%d", payload.Timestamp, payload.Event_ID)
		err := app.Obs.ScreenhotToBucket("PondCamera", fn, "pond-cam", app.S3)
		if err != nil {
			log.Printf("[OBS] Error saving screenshot: %v\n", err)
		} else {
			log.Printf("[OBS] Saved screenshot to cloud %s.png\n", fn)
		}
	}
	//app.TwitchIrc.SendChannelMessage(fmt.Sprintf("Food dispensed! Force of string pull: %d", pm))
}
