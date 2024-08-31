package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Task struct {
	T          TaskID
	Enabled    bool
	Interval   time.Duration
	Timeout    time.Duration
	F          TaskFunction
	RunAtStart bool

	ticker *time.Ticker
}
type TaskID string

func (t *Task) Run() {
	t.Log("Running task\n")

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	if err := t.F(t, ctx); err != nil {
		t.Log(fmt.Sprintf("Error running task: %v\n", err))
	}
}

func (t *Task) Log(msg string) {
	log.Printf("[Task][%s] %s", t.T, msg)
}

func (t *Task) LogLn(msg string) {
	log.Printf("[Task][%s] %s\n", t.T, msg)
}

type TaskFunction func(t *Task, ctx context.Context) error
