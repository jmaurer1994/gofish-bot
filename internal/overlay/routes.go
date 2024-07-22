package overlay

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jmaurer1994/gofish-bot/internal/overlay/views"
	"io"
	"net/http"
)

func (app *Config) Routes() {
	app.Router.GET("/stream", HeadersMiddleware(), app.Event.serveHTTP(), func(c *gin.Context) {
		v, ok := c.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}
		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				c.SSEvent(msg.Channel, msg.Data)
				return true
			}
			return false
		})
	})

	app.Router.GET("/", app.indexPageHandler())

	app.Router.Static("/assets", "../../assets")
	app.Router.StaticFile("/style.css", "../../resources/style.css")
}

func (app *Config) indexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.Index())
	}
}
