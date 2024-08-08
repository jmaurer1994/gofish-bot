package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGClient struct {
	pool      *pgxpool.Pool
	listening bool
}

func NewPGClient(databaseUrl string) (*PGClient, error) {
	var pgc = &PGClient{}

	var err error
	pgc.pool, err = pgxpool.New(context.Background(), databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %s", err)

	}

	return pgc, err
}

type NotificationFunction func(n *pgconn.Notification)

func (pgc *PGClient) Listen(ctx context.Context, channel string, nf NotificationFunction) {
	conn, err := pgc.pool.Acquire(ctx)

	if err != nil {
		log.Printf("[Database] Error acquiring connection: %v\n", err)
		return
	}

	defer func() {
		conn.Release()
	}()

	_, err = conn.Exec(ctx, fmt.Sprintf("listen %s", channel))
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
			log.Printf("[Database] Listener context error: %v\n", ctx.Err())
			return
		}

		log.Printf("[Database] Received event on %s\n%s\n", notification.Channel, notification.Payload)
		nf(notification)
	}
}
