package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/jmaurer1994/gofish-bot/internal/camera"
	"github.com/jmaurer1994/gofish-bot/internal/command_processor"
	"github.com/jmaurer1994/gofish-bot/internal/database"
	"github.com/jmaurer1994/gofish-bot/internal/obs"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
	"github.com/jmaurer1994/gofish-bot/internal/ttv"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
)

var (
	tac     ttv.TwitchApiClient
	tic     ttv.TwitchIrcClient
	gc      *obs.GoobsClient
	owm     weather.OwmClient
	db      *database.PGClient
	c       camera.IpCamera
	sch     *scheduler.Scheduler
	cmdproc command_processor.CommandProcessor
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cameraSetup()
	owmSetup()
	obsSetup()
	ttvSetup()
	schedulerSetup()

	db, err = database.NewPGClient(os.Getenv("DB_CONNECTION_URL"), sch)
	db.StartListener()

	cmdproc = command_processor.New("!")
	registerChatCommands()
	if err != nil {
		log.Printf("Error creating db client %v\n", err)
	}

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

func owmSetup() {
	weatherLatitude, latErr := strconv.ParseFloat(os.Getenv("WEATHER_LATITUDE"), 64)
	weatherLongitude, longErr := strconv.ParseFloat(os.Getenv("WEATHER_LONGITUDE"), 64)

	if latErr != nil || longErr != nil {
		log.Fatalf("Could not parse latitude(%v) or longitude(%v)", latErr, longErr)
	}

	owm = weather.OwmClient{
		Latitude:  weatherLatitude,
		Longitude: weatherLongitude,
		OwmApiKey: os.Getenv("OWM_API_KEY"),
	}
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

	router := gin.Default()

	app := obs.Config{Router: router}

	app.Routes()

	router.Run(":8080")
}

func ttvSetup() {
	tac = ttv.TwitchApiClient{
		ClientId:      os.Getenv("TTV_CLIENT_ID"),
		BroadcasterId: os.Getenv("TTV_BROADCASTER_ID"),
		TokenSource: ttv.NewTwitchTokenSource(
			"api-client",
			os.Getenv("TTV_CLIENT_ID"),
			os.Getenv("TTV_CLIENT_SECRET"),
			os.Getenv("TTV_REDIRECT_URI"),
			os.Getenv("TTV_AUTHSERVER_PORT"),
			[]string{"channel:manage:broadcast"}),
	}

	tic = ttv.TwitchIrcClient{
		Channel:  os.Getenv("TTV_CHANNEL_NAME"),
		Username: os.Getenv("TTV_BOT_USERNAME"),
		TokenSource: ttv.NewTwitchTokenSource("irc-client",
			os.Getenv("TTV_CLIENT_ID"),
			os.Getenv("TTV_CLIENT_SECRET"),
			os.Getenv("TTV_REDIRECT_URI"),
			os.Getenv("TTV_AUTHSERVER_PORT"),
			[]string{"user:read:email", "channel:moderate", "chat:edit", "chat:read", "whispers:read", "whispers:edit"}),
	}

	if err := tic.Connect(); err != nil {
		log.Printf("Twitch IRC error: %v", err)
	}

	registerIrcHandlers()
}

func schedulerSetup() {
	sch = scheduler.NewScheduler()
	registerSchedulerTasks(sch)
	registerSchedulerEvents(sch)

	log.Println("Starting task scheduler")
	sch.Start()
}
