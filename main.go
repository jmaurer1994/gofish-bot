package main

import (
	"github.com/Adeithe/go-twitch/irc"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jmaurer1994/gofish/bot/camera"
	"github.com/jmaurer1994/gofish/bot/obs"
	"github.com/jmaurer1994/gofish/bot/scheduler"
	"github.com/jmaurer1994/gofish/bot/ttv"
	"github.com/jmaurer1994/gofish/bot/weather"
)

var (
	tac              ttv.TwitchApiClient
	tic              ttv.TwitchIrcClient
	gc               *obs.GoobsClient
	owm              weather.OwmClient
	c                camera.IpCamera
	ttvClientId      string
	ttvChannelName   string
	ttvBroadcasterId string
	ttvBotUsername   string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	weatherLatitude, latErr := strconv.ParseFloat(os.Getenv("WEATHER_LATITUDE"), 64)
	weatherLongitude, longErr := strconv.ParseFloat(os.Getenv("WEATHER_LONGITUDE"), 64)

	if latErr != nil || longErr != nil {
		log.Fatalf("Could not parse latitude(%v) or longitude(%v)", latErr, longErr)
	}

	ttvClientId = os.Getenv("TTV_CLIENT_ID")
	ttvClientSecret := os.Getenv("TTV_CLIENT_SECRET")
	ttvRedirectUri := os.Getenv("TTV_REDIRECT_URI")
	ttvChannelName = os.Getenv("TTV_CHANNEL_NAME")
	ttvBroadcasterId = os.Getenv("TTV_BROADCASTER_ID")
	ttvBotUsername = os.Getenv("TTV_BOT_USERNAME")

	ttvAuthServerPort := os.Getenv("TTV_AUTHSERVER_PORT")

	c = camera.IpCamera{
		Address:  os.Getenv("IPCAMERA_ADDRESS"),
		Username: os.Getenv("IPCAMERA_USERNAME"),
		Password: os.Getenv("IPCAMERA_PASSWORD"),
	}

	c.ZeroLight()

	obsHost := os.Getenv("OBS_HOST")
	obsPassword := os.Getenv("OBS_PASSWORD")
    obsScreenshotDirectory := os.Getenv("OBS_SCREENSHOT_DIRECTORY")
    obsScreenshotFormat := os.Getenv("OBS_SCREENSHOT_FORMAT")
    obsScreenshotQuality, qualityErr := strconv.ParseFloat(os.Getenv("OBS_SCREENSHOT_QUALITY"), 64)
    if(qualityErr != nil){
        log.Fatalf("Could not parse image quality from env value")
    }

	gc, err = obs.NewGoobsClient(obsHost, obsPassword, obsScreenshotDirectory, obsScreenshotFormat, obsScreenshotQuality)
	if err != nil {
		log.Fatal("Error creating goobs client")
	}

	owm = weather.OwmClient{
		Latitude:  weatherLatitude,
		Longitude: weatherLongitude,
		OwmApiKey: os.Getenv("OWM_API_KEY"),
	}

	tac = ttv.TwitchApiClient{
		ClientId:      ttvClientId,
		BroadcasterId: ttvBroadcasterId,
		TokenSource:   ttv.NewTwitchTokenSource("api-client", ttvClientId, ttvClientSecret, ttvRedirectUri, ttvAuthServerPort, []string{"channel:manage:broadcast"}),
	}

	tic = ttv.TwitchIrcClient{
		Channel:     ttvChannelName,
		Username:    ttvBotUsername,
		TokenSource: ttv.NewTwitchTokenSource("irc-client", ttvClientId, ttvClientSecret, ttvRedirectUri, ttvAuthServerPort, []string{"user:read:email", "channel:moderate", "chat:edit", "chat:read", "whispers:read", "whispers:edit"}),
	}

	err = tic.Connect()
	if err != nil {
		log.Fatalf("IRC Error: %v", err)
	}
	tic.RegisterHandlers(func(ircReader *irc.Client) {
		ircReader.OnShardReconnect(onShardReconnect)
		ircReader.OnShardServerNotice(onShardServerNotice)
		ircReader.OnShardLatencyUpdate(onShardLatencyUpdate)
		ircReader.OnShardMessage(onChannelMessage)
		ircReader.OnShardRawMessage(onRawMessage)
	})

	sch := scheduler.NewScheduler()

	sch.RegisterTask(scheduler.Task{
		T:          "channel:title:update",
		Enabled:    true,
		Interval:   time.Duration(5) * time.Minute,
		F:          UpdateChannelTitle,
		RunAtStart: true,
	})

	sch.RegisterTask(scheduler.Task{
		T:          "source:screenshot:save",
		Enabled:    true,
		Interval:   time.Duration(30) * time.Second,
		F:          SavePondCameraScreenshot,
		RunAtStart: true,
	})

	sch.RegisterEventHandler("camera:light:check", handleCameraLightCheck)
    sch.RegisterEventHandler("ForceSensor:Insert", handleDatabaseEvent)

	// Create a channel to receive os.Signal values.
	sigs := make(chan os.Signal, 1)

	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting task scheduler")
	sch.Start()

	<-sigs
}

func onShardReconnect(shardID int) {
	log.Printf("Shard #%d reconnected\n", shardID)
}

func onShardServerNotice(shardID int, sn irc.ServerNotice) {
	log.Printf("Shard #%d recv: %s\n", shardID, sn.Message)
}

func onShardChannelUserNotice(shardID int, n irc.UserNotice) {
	log.Printf("Shard #%d recv: %s\n", shardID, n.Message)
}

func onShardLatencyUpdate(shardID int, latency time.Duration) {
	log.Printf("Shard #%d has %dms ping\n", shardID, latency.Milliseconds())
}

// TODO: command processor
func onChannelMessage(shardID int, msg irc.ChatMessage) {
	log.Printf("#%s %s: %s\n", msg.Channel, msg.Sender.DisplayName, msg.Text)
	tokens := strings.Fields(msg.Text)
	if len(tokens) > 0 {
		switch tokens[0] {
		case "!help":
			tic.SendChannelMessage("Available commands: !help, !info, !clapclap*, !fixcamera*")
		case "!info":
			tic.SendChannelMessage("Welcome to the channel and thanks for stopping by! " +
				"This is an ongoing personal/hobby project - the goal of which is " +
				"to gather some data and monitor the feeding habits of the pond residents. " +
				"See below for more information.")
		case "!clapclap":
			if msg.Sender.IsModerator {
				if c.CurrentLightLevel() > 0 {
					c.ZeroLight()
				} else {
					c.IncreaseLight()
				}
			}
		case "!fixcamera":
			if msg.Sender.IsModerator {
				err := gc.ToggleSourceVisibility("Main", "PondCamera")
				if err != nil {
					log.Printf("%v\n", err)
				}
			}
		}
	}
}

func onRawMessage(shardID int, msg irc.Message) {
	log.Printf("#%s: %s\n", msg.Sender.Username, msg.Raw)
}
