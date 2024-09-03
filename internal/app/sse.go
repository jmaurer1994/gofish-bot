package app

import (
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

func (s *EventServer) Render(channel string, template templ.Component) {
	s.Event <- Message{Channel: channel, Data: template}
}

func render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}

func NewServer() (event *EventServer) {
	event = &EventServer{
		Event:         make(chan Message),
		NewClients:    make(chan chan Message),
		ClosedClients: make(chan chan Message),
		TotalClients:  make(map[chan Message]bool),
	}

	return
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (s *EventServer) Listen() {
	for {
		select {
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

func (s *EventServer) serveHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		s.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			s.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-s")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
