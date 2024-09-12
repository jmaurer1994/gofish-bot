package app

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/app/views"
	"github.com/jmaurer1994/gofish-bot/internal/app/views/components"
)

func (app *Config) Routes() {
	app.Router.GET("/stream", HeadersMiddleware(), app.EventServer.ServeHTTP(), EventHandler())

	app.Router.GET("/", app.OverlayViewHandler())
	app.Router.GET("/overlay-weather", app.weatherComponentHandler())
	app.Router.GET("/overlay-countdown", app.countdownComponentHandler())
	app.Router.GET("/overlay-feeder", app.feederComponentHandler())

	app.Router.Static("/assets", "./assets")
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

func (app *Config) OverlayViewHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.Overlay())
	}
}

func render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}
