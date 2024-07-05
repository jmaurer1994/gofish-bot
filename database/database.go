package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmaurer1994/gofish/bot/scheduler"
	"log"
	"os"
)

type PGClient struct {
	pool           *pgxpool.Pool
	scheduler      *scheduler.Scheduler
	cancelListener context.CancelCauseFunc
	listening      bool
}

type DailyEventStats struct {
	Day_event_count int
	Day_max_force   float32
	Day_min_force   float32
	Day_avg_force   float32
}

type WeeklyEventStats struct {
	Week_total_events int
	Daily_avg_events  float32
	Week_max_force    float32
	Week_min_force    float32
	Week_avg_force    float32
}

type MonthlyEventStats struct {
	Month_total_events int
	Daily_avg_events   float32
	Month_max_force    float32
	Month_min_force    float32
	Month_avg_force    float32
}

const (
	listenerCancelError = "Cancelled by caller"
)

func NewPGClient(databaseUrl string, s *scheduler.Scheduler) (*PGClient, error) {
	var pgc = &PGClient{
		scheduler: s,
	}

	var err error
	pgc.pool, err = pgxpool.New(context.Background(), databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %s", err)

	}

	return pgc, err
}

func (pgc *PGClient) RetrieveDailyStats() (DailyEventStats, error) {
	conn, err := pgc.pool.Acquire(context.Background())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		return DailyEventStats{}, err
	}

	var s DailyEventStats
	if err := conn.QueryRow(context.Background(), `SELECT * FROM data."v_DailyEventStats`).Scan(&s.Day_event_count, &s.Day_max_force, &s.Day_min_force, &s.Day_avg_force); err != nil {
		return DailyEventStats{}, err
	}

	return s, nil
}

func (pgc *PGClient) RetrieveWeeklyStats() (WeeklyEventStats, error) {
	conn, err := pgc.pool.Acquire(context.Background())

	if err != nil {
		return WeeklyEventStats{}, err
	}

	var s WeeklyEventStats
	if err := conn.QueryRow(context.Background(), `SELECT * FROM data."v_WeeklyEventStats`).Scan(&s.Week_total_events, &s.Daily_avg_events, &s.Week_max_force, &s.Week_min_force, &s.Week_avg_force); err != nil {
		return WeeklyEventStats{}, err
	}

	return s, nil
}

func (pgc *PGClient) RetrieveMonthlyStats() (MonthlyEventStats, error) {
	conn, err := pgc.pool.Acquire(context.Background())

	if err != nil {
		return MonthlyEventStats{}, err
	}

	var s MonthlyEventStats
	if err := conn.QueryRow(context.Background(), `SELECT * FROM data."v_MonthlyEventStats`).Scan(&s.Month_total_events, &s.Daily_avg_events, &s.Month_max_force, &s.Month_min_force, &s.Month_avg_force); err != nil {
		return MonthlyEventStats{}, err
	}

	return s, nil
}

func (pgc *PGClient) IsListening() bool {
	return pgc.listening
}

func (pgc *PGClient) StartListener() {
	if pgc.listening {
		log.Println("Database listener is already running!")
	}

	go pgc.listen()

	return
}

func (pgc *PGClient) StopListener() {
	if pgc.cancelListener != nil {
		log.Println("Database listener cancel function not available")
		return
	}

	pgc.cancelListener(errors.New("Cancelled by caller"))
}

func (pgc *PGClient) listen() {
	pgc.listening = true
	conn, err := pgc.pool.Acquire(context.Background())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		return
	}

	defer func() {
		pgc.listening = false
		conn.Release()
	}()

	_, err = conn.Exec(context.Background(), "listen sensoreventinsert")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to event", err)
		return
	}

	for {
		var listenerCtx context.Context

		listenerCtx, pgc.cancelListener = context.WithCancelCause(context.Background())
		notification, err := conn.Conn().WaitForNotification(listenerCtx)

		log.Printf("ListenerCtx err: %v\n", listenerCtx.Err())

		if err = context.Cause(listenerCtx); err.Error() == listenerCancelError {
			log.Println("Database listener cancelled")
			return
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
			return
		}

		fmt.Println("PID:", notification.PID, "Channel:", notification.Channel, "Payload:", notification.Payload)
		pgc.scheduler.GenerateEvent("SensorEvent:Insert", scheduler.Message(notification.Payload))
	}
}
