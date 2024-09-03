package app

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/app/views"
	"github.com/jmaurer1994/gofish-bot/internal/app/views/components"
)

func (app *Config) Routes() {
	app.Router.GET("/stream", HeadersMiddleware(), app.Overlay.serveHTTP(), app.eventHandler())

	app.Router.GET("/", app.indexPageHandler())
	app.Router.GET("/weather", app.weatherComponentHandler())
	app.Router.GET("/countdown", app.countdownComponentHandler())
	app.Router.GET("/feeder", app.feederComponentHandler())

	app.Router.Static("/assets", "./assets")
	app.Router.StaticFile("/style.css", "./resources/style.css")
	app.Router.StaticFile("/main.js", "./resources/main.js")

}

func (app *Config) weatherComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, components.WeatherWidget(app.Data.Weather))
	}
}

func (app *Config) countdownComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, components.CountdownWidget(app.Data.Countdown.Hours(), app.Data.Countdown.Minutes(), app.Data.Countdown.Target))
	}
}

func (app *Config) feederComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, components.FeederWidget(app.Data.FeederWeight))
	}
}

func (app *Config) eventHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, ok := ctx.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}
		ctx.Stream(func(w io.Writer) bool {

			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				template := msg.Data

				if err := template.Render(ctx, w); err != nil {
					log.Printf("[SSE] Render error: %v\n", err)
					return false
				}
				ctx.SSEvent(msg.Channel, w)

				return true
			}
			return false
		})
	}
}
func (app *Config) indexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.Index())
	}
}
