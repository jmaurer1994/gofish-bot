package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/camera"
	"github.com/jmaurer1994/gofish-bot/internal/chat"
	"github.com/jmaurer1994/gofish-bot/internal/database"
	"github.com/jmaurer1994/gofish-bot/internal/obs"
	"github.com/jmaurer1994/gofish-bot/internal/scheduler"
	"github.com/jmaurer1994/gofish-bot/internal/twitch"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
	"github.com/joho/godotenv"
)

const appTimeout = time.Second * 30

type Config struct {
	Router    *gin.Engine
	Overlay   *SSEvent
	TwitchApi *twitch.TwitchApiClient
	TwitchIrc *twitch.TwitchIrcClient
	CmdProc   *chat.CommandProcessor
	Obs       *obs.GoobsClient
	Db        *database.PGClient
	Camera    *camera.IpCamera
	Scheduler *scheduler.Scheduler
	OwmApi    *weather.OwmClient

	Data struct {
		Countdown    Countdown
		Weather      weather.OneCallResponse
		FeederWeight float64
	}
}

func (app *Config) Start() {
	log.Println("Starting task scheduler")
	app.Scheduler.Start()

	go app.Router.Run(":8080")
	go app.Overlay.Listen()
	go app.Db.Listen(context.Background())
	// Create a channel to receive os.Signal values.operator
	sigs := make(chan os.Signal, 1)

	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

}

func NewApp() *Config {
	router := gin.Default()
	sse := NewServer()

	app := &Config{Router: router, Overlay: sse}

	return app
}

func (app *Config) Init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load environment file")
	}
	app.Routes()

	app.cameraSetup()
	app.owmSetup()
	app.obsSetup()
	app.twitchSetup()

	app.schedulerSetup()

}

func (app *Config) dbSetup() {
	db, err := database.NewPGClient(os.Getenv("DB_CONNECTION_URL"))

	if err != nil {
		log.Printf("Error creating db client %v\n", err)
		return
	}

	app.Db = db
}

func (app *Config) owmSetup() {
	weatherLatitude, latErr := strconv.ParseFloat(os.Getenv("WEATHER_LATITUDE"), 64)
	weatherLongitude, longErr := strconv.ParseFloat(os.Getenv("WEATHER_LONGITUDE"), 64)

	if latErr != nil || longErr != nil {
		log.Printf("Could not parse latitude(%v) or longitude(%v)\n", latErr, longErr)
		return
	}

	app.OwmApi = &weather.OwmClient{
		Latitude:  weatherLatitude,
		Longitude: weatherLongitude,
		OwmApiKey: os.Getenv("OWM_API_KEY"),
	}
}
func (app *Config) twitchSetup() {
	app.TwitchApi = &twitch.TwitchApiClient{
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

	app.TwitchIrc = &twitch.TwitchIrcClient{
		Channel:  os.Getenv("TTV_CHANNEL_NAME"),
		Username: os.Getenv("TTV_BOT_USERNAME"),
		TokenSource: twitch.NewTwitchTokenSource("irc-client",
			os.Getenv("TTV_CLIENT_ID"),
			os.Getenv("TTV_CLIENT_SECRET"),
			os.Getenv("TTV_REDIRECT_URI"),
			os.Getenv("TTV_AUTHSERVER_PORT"),
			[]string{"user:read:email", "channel:moderate", "chat:edit", "chat:read", "whispers:read", "whispers:edit"}),
	}

	app.CmdProc = chat.NewCommandProcessor("!")

	if err := app.TwitchIrc.InitializeConnection(); err != nil {
		log.Printf("Twitch IRC error: %v", err)
	}

	app.registerIrcHandlers()
	app.registerChatCommands()
}

func (app *Config) cameraSetup() {
	app.Camera = &camera.IpCamera{
		Address:  os.Getenv("IPCAMERA_ADDRESS"),
		Username: os.Getenv("IPCAMERA_USERNAME"),
		Password: os.Getenv("IPCAMERA_PASSWORD"),
	}

}

func (app *Config) obsSetup() {
	obsScreenshotQuality, qualityErr := strconv.ParseFloat(os.Getenv("OBS_SCREENSHOT_QUALITY"), 64)
	if qualityErr != nil {
		log.Println("Could not parse image quality from env value")
		obsScreenshotQuality = 0.8
	}
	var err error
	app.Obs, err = obs.NewGoobsClient(os.Getenv("OBS_HOST"), os.Getenv("OBS_PASSWORD"), os.Getenv("OBS_SCREENSHOT_DIRECTORY"), os.Getenv("OBS_SCREENSHOT_FORMAT"), obsScreenshotQuality)
	if err != nil {
		log.Printf("Error creating goobs client\n")
	}
}

func (app *Config) schedulerSetup() {
	app.Scheduler = scheduler.NewScheduler()
	app.registerSchedulerTasks()
	app.registerSchedulerEvents()

}
