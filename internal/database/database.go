package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
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
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		return
	}

	defer func() {
		conn.Release()
	}()

	_, err = conn.Exec(ctx, "listen sensoreventinsert")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to event", err)
		return
	}

	for {

		notification, err := conn.Conn().WaitForNotification(ctx)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
			return
		}

		if ctx.Err() != nil {
			return
		}

		pgc.NotificationChannel <- notification.Payload

	}
}
