package scheduler

import (
	//"fmt"
	"time"
)

type Task struct {
	T          TaskID
	Enabled    bool
	Interval   time.Duration
	F          ExecuteFunction
	RunAtStart bool

	ticker *time.Ticker
}

type Scheduler struct {
	tasks     []Task
	ec        chan Event
	listeners map[EventID][]EventHandler
}

type ExecuteFunction func(s *Scheduler)

type EventHandler func(s *Scheduler, m Message)

type Event struct {
	E EventID
	M Message
}

type TaskID string
type EventID string
type Message string

func NewScheduler() *Scheduler {
	sch := &Scheduler{}
    sch.listeners = make(map[EventID][]EventHandler)
	return sch
}

func (s *Scheduler) RegisterEventHandler(e EventID, F EventHandler) {
	s.listeners[e] = append(s.listeners[e], F)
}

func (s *Scheduler) RegisterTask(T Task) {
	s.tasks = append(s.tasks, T)
}

func (s *Scheduler) GenerateEvent(e EventID, m Message) {
	if handlers, ok := s.listeners[e]; ok {

		for _, h := range handlers {
			go h(s, m)
		}
	}
}

func (s *Scheduler) Start() {
	for _, task := range s.tasks {

		if task.ticker != nil {
			task.ticker.Reset(task.Interval)
			continue
		}

		task.ticker = time.NewTicker(task.Interval)
		go func(task *Task) {
			if task.RunAtStart {
				task.F(s)
			}
			for {
				<-task.ticker.C

				task.F(s)
			}
		}(&task)
	}
}

func (s *Scheduler) Stop() {
	for _, task := range s.tasks {
		task.ticker.Stop()
	}
}
