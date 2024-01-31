package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
    "github.com/jmaurer1994/gofish/bot/scheduler"
)

type PGClient struct {
	pool *pgxpool.Pool
    scheduler *scheduler.Scheduler
}

func NewPGClient(databaseUrl string, s *scheduler.Scheduler) (*PGClient, error) {
	var pgc = &PGClient{
        scheduler: s,
    }

	var err error
	pgc.pool, err = pgxpool.New(context.Background(), databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %s", err)

	}

	go pgc.listen()

	return pgc, err
}
func (pgc *PGClient) listen() {
	conn, err := pgc.pool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		return
	}
	defer conn.Release()

    _, err = conn.Exec(context.Background(), "listen \"SensorEvent:Insert\"")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to event", err)
		return
	}

	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
            return
		}

		fmt.Println("PID:", notification.PID, "Channel:", notification.Channel, "Payload:", notification.Payload)
        pgc.scheduler.GenerateEvent("SensorEvent:Insert", scheduler.Message(notification.Payload))
	}
}
