package scheduler

import (
	"log"
	"time"
)

type Scheduler struct {
	tasks     []*Task
	ec        chan Event
	listeners map[EventID][]EventHandler
}

type EventHandler func(m Message)

type Event struct {
	E EventID
	M Message
}

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
	s.tasks = append(s.tasks, &T)
}

func (s *Scheduler) GenerateEvent(e EventID, m Message) {
	if handlers, ok := s.listeners[e]; ok {
		for _, h := range handlers {
			go h(m)
		}
	}
}

func (s *Scheduler) Start() {
	log.Println("[Scheduler] Starting scheduler")

	for _, task := range s.tasks {
		if task.ticker != nil {
			task.ticker.Reset(task.Interval)
			continue
		}

		task.ticker = time.NewTicker(task.Interval)
		go func(task *Task) {
			if task.RunAtStart && task.Enabled {
				task.Run()
			}
			for {
				<-task.ticker.C
				if task.Enabled {
					task.Run()
				}
			}
		}(task)
	}
}

func (s *Scheduler) Stop() {
	for _, task := range s.tasks {
		task.ticker.Stop()
	}
}
