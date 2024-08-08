package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGClient struct {
	pool                *pgxpool.Pool
	listening           bool
	NotificationChannel chan (string)
}

func NewPGClient(databaseUrl string) (*PGClient, error) {
	var pgc = &PGClient{
		NotificationChannel: make(chan string, 10),
	}

	var err error
	pgc.pool, err = pgxpool.New(context.Background(), databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %s", err)

	}

	return pgc, err
}

func (pgc *PGClient) Listen(ctx context.Context) {
	conn, err := pgc.pool.Acquire(ctx)

	if err != nil {
		log.Printf("[Database] Error acquiring connection: %v\n", err)
		return
	}

	defer func() {
		conn.Release()
	}()

	_, err = conn.Exec(ctx, "listen sensoreventinsert")
	if err != nil {
		log.Printf("[Database] Error listening to event: %v\n", err)
		return
	}

	for {

		notification, err := conn.Conn().WaitForNotification(ctx)

		if err != nil {
			log.Printf("[Database] Error waiting for notification: %v\n", err)
			return
		}

		if ctx.Err() != nil {
			return
		}

		log.Printf("[Database] Received event on %s\n%s\n", notification.Channel, notification.Payload)
		pgc.NotificationChannel <- notification.Payload

	}
}
