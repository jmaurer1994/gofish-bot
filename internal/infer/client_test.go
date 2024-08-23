package infer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func cb(results []TaskResult) {
	for _, result := range results {
		fmt.Print(result.Name)
	}

	fmt.Println("")
}

func TestClient(t *testing.T) {
	c := NewInferenceClient("track", os.Getenv("INFERENCE_SOURCE"), os.Getenv("INFERENCE_HOST"), os.Getenv("INFERENCE_PORT"), cb)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	go c.RunTask(ctx)

	// Create a channel to receive os.Signal values.operator
	sigs := make(chan os.Signal, 1)

	// Notify the channel if a SIGINT or SIGTERM is received.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}
