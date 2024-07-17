package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
	"log"
)

func registerSchedulerEvents(sch *scheduler.Scheduler) {
	sch.RegisterEventHandler("camera:light:check", handleCameraLightCheck)
	sch.RegisterEventHandler("SensorEvent:Insert", handleDatabaseEvent)
}

func handleCameraLightCheck(s *scheduler.Scheduler, m scheduler.Message) {
	log.Printf("Received light check event: %s\n", m)
	switch m {
	case "on":
		if c.CurrentLightLevel() == 0 {
			c.SetLightLevel(1)
		}
	case "off":
		if c.CurrentLightLevel() > 0 {
			c.ZeroLight()
		}
	}
}

type SensorEventPayload struct {
	Event_ID  int   `json:"event_id"`
	Timestamp int   `json:"timestamp"`
	Samples   []int `json:"samples"`
}

func handleDatabaseEvent(s *scheduler.Scheduler, m scheduler.Message) {
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

	tic.SendChannelMessage(fmt.Sprintf("Food dispensed! Force of string pull: %d", pm))
}
