package main

import (
	"context"
	_ "context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmaurer1994/gofish-bot/internal/app"
	"github.com/jmaurer1994/gofish-bot/internal/app/views/components"
	"github.com/jmaurer1994/gofish-bot/internal/weather"
	"github.com/joho/godotenv"
)

var (
	a *app.Config
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load environment file")
	}

	a = app.NewApp()

	a.Routes()

	go a.Router.Run()
	go a.EventServer.Listen(context.TODO())

	//i := infer.NewInferenceClient("track", os.Getenv("INFERENCE_SOURCE"), os.Getenv("INFERENCE_HOST"), os.Getenv("INFERENCE_PORT"),
	//	func(s *pb.TaskResultSet) {
	//		a.Overlay.Render("inference", components.InferenceResult(s))
	//	})

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	//go i.RunTask(ctx)
	//	go i.RunTask(context.Background())

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
		a.EventServer.SendEvent("weather", components.WeatherWidget(weather.OneCallResponse{}))
		a.EventServer.SendEvent("feeder", components.FeederWidget(rand.Float64()*100))
	}
}
