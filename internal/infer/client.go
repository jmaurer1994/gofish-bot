package infer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jmaurer1994/gofish-bot/internal/infer/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InferenceClient struct {
	Host   string
	Port   string
	Task   string
	source string
	conn   *grpc.ClientConn
	cb     CallbackFunc
}
type CallbackFunc func(s *protos.TaskResultSet)

func NewInferenceClient(task, source, host, port string, cb CallbackFunc) *InferenceClient {
	return &InferenceClient{
		Host:   host,
		Port:   port,
		Task:   task,
		cb:     cb,
		source: source,
	}
}

var ConnectionClosed error = errors.New("Connection closed")

func (c *InferenceClient) RunTask(ctx context.Context) {
	// Connect to the server
	var err error
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", c.Host, c.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Error connecting:", err)
		return
	}

	client := protos.NewTaskServiceClient(c.conn)

	defer func() {
		c.conn.Close()
		log.Println("Connection Closed")
	}()

	// Create a request
	request := &protos.TaskRequest{
		TaskName: c.Task,
		Source:   c.source,
	}

	// Call StreamResults and receive results
	stream, err := client.StreamResults(ctx, request)

	if err != nil {
		log.Printf("could not stream results: %v\n", err)
		return
	}

	frameTimeLimit := time.Duration(int64(math.Round((1.0 / 30) * 1e9)))
	frames := 0
	for {
		frameStart := time.Now()
		frames++
		result, err := stream.Recv()
		if err != nil {
			log.Println("Error receiving message:", err)
			break
		}

		if t := time.Now().Sub(frameStart); t > frameTimeLimit {
			log.Printf("[%d] time exceeded: %dms\t%dms\n", frames, t.Milliseconds(), frameTimeLimit.Milliseconds())
		}

		c.cb(result)
	}

	log.Printf("Processed %d frames", frames)
}
