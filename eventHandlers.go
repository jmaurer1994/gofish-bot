package main

import (
	"github.com/jmaurer1994/gofish/bot/scheduler"
	"log"
)

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

func handleDatabaseEvent(s *scheduler.Scheduler, m scheduler.Message) {
    log.Printf("Received insert event - did someone pull the string?\n%s\n", m)
}
