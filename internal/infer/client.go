package infer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

type InferenceClient struct {
	Host   string
	Port   string
	Task   string
	source string
	conn   net.Conn
	cb     CallbackFunc
}
type CallbackFunc func(results []TaskResult)

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
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer func() {
		c.conn.Close()
		fmt.Println("Connection Closed")
	}()
	message := fmt.Sprintf("%s\t%s", c.Task, c.source)
	// Continuously receive messages from the server

	fmt.Printf("Sending message: %s\n", message)
	if err := c.sendMessage([]byte(message)); err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done")
			return
		default:
			message, err := c.receiveMessage()
			if err != nil {
				fmt.Println("Error receiving message:", err)
				return
			}
			c.handleTaskResult(message)
		}
	}
}

func (c *InferenceClient) handleTaskResult(message []byte) {
	var results []TaskResult
	err := json.Unmarshal([]byte(string(message)), &results)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	if len(results) == 0 {
		return
	}
	c.cb(results)
}

func (c *InferenceClient) sendMessage(message []byte) error {
	// Prefix the message with a 4-byte length
	if c.conn == nil {
		return errors.New("connection closed")
	}

	msglen := make([]byte, 4)
	binary.BigEndian.PutUint32(msglen, uint32(len(message)))

	// Send the length and the message
	if _, err := c.conn.Write(msglen); err != nil {
		return err
	}
	if _, err := c.conn.Write(message); err != nil {
		return err
	}
	return nil
}

func (c *InferenceClient) receiveMessage() ([]byte, error) {
	if c.conn == nil {
		return nil, errors.New("connection closed")
	}

	// Read the message length (4 bytes)
	msglen := make([]byte, 4)
	if _, err := c.conn.Read(msglen); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(msglen)

	// Read the actual message
	data := make([]byte, length)
	if _, err := c.conn.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

type TaskResult struct {
	Class      int       `json:"class"`
	Name       string    `json:"name"`
	Track_Id   int       `json:"track_id"`
	Confidence float64   `json:"confidence"`
	Box        *Box      `json:"box,omitempty"`      // Only for detection models
	Segments   *Segments `json:"segments,omitempty"` // Only for segmentation models
	Speed      Speed     `json:"speed"`
	Shape      []int     `json:"shape"`
}

type Box struct {
	X1 float64 `json:"x1"`
	X2 float64 `json:"x2"`
	Y1 float64 `json:"y1"`
	Y2 float64 `json:"y2"`
	// OBB models might have more points
	X3 float64 `json:"x3,omitempty"`
	X4 float64 `json:"x4,omitempty"`
	Y3 float64 `json:"y3,omitempty"`
	Y4 float64 `json:"y4,omitempty"`
}

type Segments struct {
	X []float64 `json:"x,omitempty"`
	Y []float64 `json:"y,omitempty"`
}

type Speed struct {
	Inference   float64 `json:"inference"`
	Postprocess float64 `json:"postprocess"`
	Preprocess  float64 `json:"preprocess"`
}
