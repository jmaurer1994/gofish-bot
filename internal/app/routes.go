package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/app/views"
	"io"
	"net/http"
)

func (app *Config) Routes() {
	app.Router.GET("/stream", HeadersMiddleware(), app.Event.serveHTTP(), app.eventHandler())

	app.Router.GET("/", app.indexPageHandler())

	app.Router.Static("/assets", "./assets")
	app.Router.StaticFile("/style.css", "./resources/style.css")
	app.Router.StaticFile("/main.js", "./resources/main.js")

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
				ctx.SSEvent(msg.Channel, msg.Data)
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
