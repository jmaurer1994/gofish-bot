package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/app"
	"github.com/jmaurer1994/gofish-bot/internal/app/views/components"
	"github.com/jmaurer1994/gofish-bot/internal/infer"
	"github.com/jmaurer1994/gofish-bot/internal/infer/pb"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
	"github.com/joho/godotenv"
)

var (
	a      *app.Config
	router *gin.Engine
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load environment file")
	}

	a = app.NewApp()

	a.Routes()

	go a.Router.Run()
	go a.Overlay.Listen()

	i := infer.NewInferenceClient("track", os.Getenv("INFERENCE_SOURCE"), os.Getenv("INFERENCE_HOST"), os.Getenv("INFERENCE_PORT"),
		func(s *pb.TaskResultSet) {
			a.Overlay.Render("inference", components.InferenceResult(s))
		})

	go i.RunTask(context.TODO())
	// Create a channel to receive os.Signal values.operator
	sigs := make(chan os.Signal, 1)
	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Starting overlay")
	//go timeUpdate()
	<-sigs
}

func timeUpdate() {
	for {

		time.Sleep(time.Second * 10)
		//		currentTime := fmt.Sprintf("The Current Time Is %v", now)

		// Send current time to clients message channel
		//        event.Render("countdown", components.CountdownWidget())
		a.Overlay.Render("weather", components.WeatherWidget(weather.OneCallResponse{}))
		a.Overlay.Render("feeder", components.FeederWidget(rand.Float64()*100))
	}
}
