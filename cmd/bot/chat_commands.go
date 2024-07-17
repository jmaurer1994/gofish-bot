package main

import (
	"fmt"
	"github.com/jmaurer1994/gofish-bot/internal/commandprocessor"
	"log"
	"time"
)

func registerChatCommands() {
	cmdproc.RegisterCommand(commandprocessor.Command{
		Key:          "help",
		F:            botHelp,
		IsModCommand: false,
		Cooldown:     10 * time.Second,
	})

	cmdproc.RegisterCommand(commandprocessor.Command{
		Key:          "info",
		F:            channelInfo,
		IsModCommand: false,
		Cooldown:     1 * time.Minute,
	})

	cmdproc.RegisterCommand(commandprocessor.Command{
		Key:          "stats",
		F:            botHelp,
		IsModCommand: false,
		Cooldown:     10 * time.Second,
	})

	cmdproc.RegisterCommand(commandprocessor.Command{
		Key:          "clapclap",
		F:            toggleLight,
		IsModCommand: true,
		Cooldown:     10 * time.Second,
	})
	cmdproc.RegisterCommand(commandprocessor.Command{
		Key:          "fixcamera",
		F:            fixCamera,
		IsModCommand: true,
		Cooldown:     10 * time.Second,
	})
}

func botHelp(args []string) {
	if len(args) == 0 {
		tic.SendChannelMessage("Available commands: !help [command], !info, !stats")
	}
}

func channelInfo(args []string) {

	tic.SendChannelMessage("Welcome to the channel and thanks for stopping by! " +
		"This is an ongoing project - the goal of which is " +
		"to gather some data and monitor the feeding habits of the pond residents. " +
		"See below for more information.")
}

func toggleLight(args []string) {
	if c.CurrentLightLevel() > 0 {
		c.ZeroLight()
	} else {
		c.IncreaseLight()
	}
}

func fixCamera(args []string) {

	err := gc.ToggleSourceVisibility("Main", "PondCamera")
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func getStats(args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {

	case "daily":
		dailyStats, err := db.RetrieveDailyStats()
		if err != nil {
			log.Printf("Error retrieving daily stats: %v", err)
			tic.SendChannelMessage("Error retrieving daily stats")
			return
		}
		tic.SendChannelMessage(fmt.Sprintf("Daily Stats: Food was dispensed %d times today with a max/min/avg force of %.0f/%.0f/%.0f", dailyStats.Day_event_count, dailyStats.Day_max_force, dailyStats.Day_min_force, dailyStats.Day_avg_force))
	case "weekly":
		weeklyStats, err := db.RetrieveWeeklyStats()
		if err != nil {
			log.Printf("Error retrieving weekly stats: %v", err)
			tic.SendChannelMessage("Error retrieving weekly stats")
			return
		}
		tic.SendChannelMessage(fmt.Sprintf("Weekly Stats: Food was dispensed %d times this week with an average of %.0f per day, having a max/min/avg force of %.0f/%.0f/%.0f", weeklyStats.Week_total_events, weeklyStats.Daily_avg_events, weeklyStats.Week_max_force, weeklyStats.Week_min_force, weeklyStats.Week_avg_force))
	case "monthly":
		monthlyStats, err := db.RetrieveMonthlyStats()
		if err != nil {
			log.Printf("Error retrieving monthly stats: %v", err)
			tic.SendChannelMessage("Error retrieving monthly stats")
			return
		}
		tic.SendChannelMessage(fmt.Sprintf("Monthly Stats: Food was dispensed %d times this Month with an average of %.0f per day, having a max/min/avg force of %.0f/%.0f/%.0f", monthlyStats.Month_total_events, monthlyStats.Daily_avg_events, monthlyStats.Month_max_force, monthlyStats.Month_min_force, monthlyStats.Month_avg_force))
	}
}
