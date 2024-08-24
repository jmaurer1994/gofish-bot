package infer

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmaurer1994/gofish-bot/internal/infer/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type InferenceClient struct {
	Host   string
	Port   string
	Task   string
	source string
	conn   *grpc.ClientConn
	cb     CallbackFunc
}
type CallbackFunc func(s *pb.TaskResultSet)

func NewInferenceClient(task, source, host, port string, cb CallbackFunc) *InferenceClient {
	return &InferenceClient{
		Host:   host,
		Port:   port,
		Task:   task,
		cb:     cb,
		source: source,
	}
}

var ConnectionClosedError = errors.New("Connection closed")

func (c *InferenceClient) RunTask(ctx context.Context) {
	// Connect to the server
	var err error
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", c.Host, c.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Error connecting:", err)
		return
	}

	client := pb.NewTaskServiceClient(c.conn)

	defer func() {
		c.conn.Close()
		log.Println("Connection Closed")
	}()

	// Create a request
	request := &pb.TaskRequest{
		TaskName: c.Task,
		Source:   c.source,
	}

	// Call StreamResults and receive results
	stream, err := client.StreamResults(ctx, request)
	if err != nil {
		log.Printf("could not stream results: %v\n", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done")
			c.cb(nil)
			return
		default:
			result, err := stream.Recv()
			if err != nil {
				log.Println("Error receiving message:", err)
				return
			}
			c.cb(result)
		}
	}
}
