package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/jmaurer1994/gofish-bot/internal/app"
	"github.com/jmaurer1994/gofish-bot/internal/camera"
	"github.com/jmaurer1994/gofish-bot/internal/commandprocessor"
	"github.com/jmaurer1994/gofish-bot/internal/database"
	"github.com/jmaurer1994/gofish-bot/internal/obs"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
	"github.com/jmaurer1994/gofish-bot/internal/twitch"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
)

var (
	tac     twitch.TwitchApiClient
	tic     twitch.TwitchIrcClient
	gc      *obs.GoobsClient
	owm     *weather.OwmClient
	db      *database.PGClient
	c       camera.IpCamera
	sch     *scheduler.Scheduler
	cmdproc commandprocessor.CommandProcessor
	event   *app.Event
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cameraSetup()
	owm, err = weather.Setup()
	obsSetup()
	ttvSetup()
	schedulerSetup()

	db, err = database.NewPGClient(os.Getenv("DB_CONNECTION_URL"), sch)
	db.StartListener()

	cmdproc = commandprocessor.NewCommandProcessor("!")
	registerChatCommands()
	if err != nil {
		log.Printf("Error creating db client %v\n", err)
	}

	log.Println("Starting task scheduler")
	sch.Start()
	// Create a channel to receive os.Signal values.operator
	sigs := make(chan os.Signal, 1)

	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}

func cameraSetup() {
	c = camera.IpCamera{
		Address:  os.Getenv("IPCAMERA_ADDRESS"),
		Username: os.Getenv("IPCAMERA_USERNAME"),
		Password: os.Getenv("IPCAMERA_PASSWORD"),
	}

	c.ZeroLight()
}

func obsSetup() {
	obsScreenshotQuality, qualityErr := strconv.ParseFloat(os.Getenv("OBS_SCREENSHOT_QUALITY"), 64)
	if qualityErr != nil {
		log.Fatalf("Could not parse image quality from env value")
		obsScreenshotQuality = 0.8
	}
	var err error
	gc, err = obs.NewGoobsClient(os.Getenv("OBS_HOST"), os.Getenv("OBS_PASSWORD"), os.Getenv("OBS_SCREENSHOT_DIRECTORY"), os.Getenv("OBS_SCREENSHOT_FORMAT"), obsScreenshotQuality)
	if err != nil {
		log.Printf("Error creating goobs client\n")
	}

	event = app.StartOverlay()
}

func ttvSetup() {
	tac = twitch.TwitchApiClient{
		ClientId:      os.Getenv("TTV_CLIENT_ID"),
		BroadcasterId: os.Getenv("TTV_BROADCASTER_ID"),
		TokenSource: twitch.NewTwitchTokenSource(
			"api-client",
			os.Getenv("TTV_CLIENT_ID"),
			os.Getenv("TTV_CLIENT_SECRET"),
			os.Getenv("TTV_REDIRECT_URI"),
			os.Getenv("TTV_AUTHSERVER_PORT"),
			[]string{"channel:manage:broadcast"}),
	}

	tic = twitch.TwitchIrcClient{
		Channel:  os.Getenv("TTV_CHANNEL_NAME"),
		Username: os.Getenv("TTV_BOT_USERNAME"),
		TokenSource: twitch.NewTwitchTokenSource("irc-client",
			os.Getenv("TTV_CLIENT_ID"),
			os.Getenv("TTV_CLIENT_SECRET"),
			os.Getenv("TTV_REDIRECT_URI"),
			os.Getenv("TTV_AUTHSERVER_PORT"),
			[]string{"user:read:email", "channel:moderate", "chat:edit", "chat:read", "whispers:read", "whispers:edit"}),
	}

	if err := tic.InitializeConnection(); err != nil {
		log.Printf("Twitch IRC error: %v", err)
	}

	registerIrcHandlers()
}

func schedulerSetup() {
	sch = scheduler.NewScheduler()
	registerSchedulerTasks(sch)
	registerSchedulerEvents(sch)

}
