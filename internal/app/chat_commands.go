package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmaurer1994/gofish-bot/internal/chat"
)

func (app *Config) registerChatCommands() {
	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "help",
		F:            app.botHelp,
		IsModCommand: false,
		Cooldown:     10 * time.Second,
	})

	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "info",
		F:            app.channelInfo,
		IsModCommand: false,
		Cooldown:     1 * time.Minute,
	})

	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "stats",
		F:            app.botHelp,
		IsModCommand: false,
		Cooldown:     10 * time.Second,
	})

	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "clapclap",
		F:            app.toggleLight,
		IsModCommand: true,
		Cooldown:     10 * time.Second,
	})
	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "fixcamera",
		F:            app.fixCamera,
		IsModCommand: true,
		Cooldown:     10 * time.Second,
	})
	app.CmdProc.RegisterCommand(chat.Command{
		Key:          "track",
		F:            app.runTracker,
		IsModCommand: true,
		Cooldown:     45 * time.Second,
	})
}

func (app *Config) botHelp(args []string) {
	if len(args) == 0 {
		app.TwitchIrc.SendChannelMessage("Available commands: !help [command], !info, !stats")
	}
}

func (app *Config) channelInfo(args []string) {

	app.TwitchIrc.SendChannelMessage("Welcome to the channel and thanks for stopping by! " +
		"This is an ongoing project - the goal of which is " +
		"to gather some data and monitor the feeding habits of the pond residents. " +
		"See below for more information.")
}

func (app *Config) toggleLight(args []string) {
	app.Camera.ToggleLight()
}

func (app *Config) fixCamera(args []string) {

	err := app.Obs.ToggleSourceVisibility("Main", "PondCamera")
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func (app *Config) getStats(args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {

	case "daily":
		dailyStats, err := app.Db.RetrieveDailyStats(context.Background())
		if err != nil {
			log.Printf("Error retrieving daily stats: %v", err)
			app.TwitchIrc.SendChannelMessage("Error retrieving daily stats")
			return
		}
		app.TwitchIrc.SendChannelMessage(fmt.Sprintf("Daily Stats: Food was dispensed %d times today with a max/min/avg force of %.0f/%.0f/%.0f", dailyStats.Day_event_count, dailyStats.Day_max_force, dailyStats.Day_min_force, dailyStats.Day_avg_force))
	case "weekly":
		weeklyStats, err := app.Db.RetrieveWeeklyStats(context.Background())
		if err != nil {
			log.Printf("Error retrieving weekly stats: %v", err)
			app.TwitchIrc.SendChannelMessage("Error retrieving weekly stats")
			return
		}
		app.TwitchIrc.SendChannelMessage(fmt.Sprintf("Weekly Stats: Food was dispensed %d times this week with an average of %.0f per day, having a max/min/avg force of %.0f/%.0f/%.0f", weeklyStats.Week_total_events, weeklyStats.Daily_avg_events, weeklyStats.Week_max_force, weeklyStats.Week_min_force, weeklyStats.Week_avg_force))
	case "monthly":
		monthlyStats, err := app.Db.RetrieveMonthlyStats(context.Background())
		if err != nil {
			log.Printf("Error retrieving monthly stats: %v", err)
			app.TwitchIrc.SendChannelMessage("Error retrieving monthly stats")
			return
		}
		app.TwitchIrc.SendChannelMessage(fmt.Sprintf("Monthly Stats: Food was dispensed %d times this Month with an average of %.0f per day, having a max/min/avg force of %.0f/%.0f/%.0f", monthlyStats.Month_total_events, monthlyStats.Daily_avg_events, monthlyStats.Month_max_force, monthlyStats.Month_min_force, monthlyStats.Month_avg_force))
	}
}

func (app *Config) runTracker(args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	app.Tracker.RunTask(ctx)

}
