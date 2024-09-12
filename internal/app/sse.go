package app

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

type Message struct {
	Channel string
	Data    templ.Component
}
type EventServer struct {
	// Events are pushed to this channel by the main events-gathering routine
	Event chan Message

	// New client connections
	NewClients chan chan Message

	// Closed client connections
	ClosedClients chan chan Message

	// Total client connections
	TotalClients map[chan Message]bool
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan Message

func (s *EventServer) SendEvent(channel string, template templ.Component) {
	s.Event <- Message{Channel: channel, Data: template}
}

func NewServer() *EventServer {
	return &EventServer{
		Event:         make(chan Message),
		NewClients:    make(chan chan Message),
		ClosedClients: make(chan chan Message),
		TotalClients:  make(map[chan Message]bool),
	}

}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (s *EventServer) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[SSE] Server stopped: %d", ctx.Err())
		// Add new available client
		case client := <-s.NewClients:
			s.TotalClients[client] = true
			log.Printf("[SSE] Client added. %d registered clients", len(s.TotalClients))

		// Remove closed client
		case client := <-s.ClosedClients:
			delete(s.TotalClients, client)
			close(client)
			log.Printf("[SSE] Removed client. %d registered clients", len(s.TotalClients))

		// Broadcast message to client
		case eventMsg := <-s.Event:
			for clientMessageChan := range s.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (s *EventServer) ServeHTTP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		s.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			s.ClosedClients <- clientChan
		}()

		ctx.Set("clientChan", clientChan)

		ctx.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")
		ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
		ctx.Next()
	}
}

func EventHandler() gin.HandlerFunc {
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
				buff := new(bytes.Buffer)

				if err := template.Render(ctx, buff); err != nil {
					log.Printf("[SSE] Render error: %v\n", err)
					return false
				}
				ctx.SSEvent(msg.Channel, buff.String())

				return true
			}
			return false
		})
	}
}
