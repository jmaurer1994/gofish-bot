package database

import (
	"context"
)

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

func (pgc *PGClient) RetrieveDailyStats(ctx context.Context) (DailyEventStats, error) {
	conn, err := pgc.pool.Acquire(ctx)

	if err != nil {
		return DailyEventStats{}, err
	}

	var s DailyEventStats
	if err := conn.QueryRow(ctx, `SELECT * FROM data."v_DailyEventStats"`).Scan(&s.Day_event_count, &s.Day_max_force, &s.Day_min_force, &s.Day_avg_force); err != nil {
		return DailyEventStats{}, err
	}

	return s, nil
}

func (pgc *PGClient) RetrieveWeeklyStats(ctx context.Context) (WeeklyEventStats, error) {
	conn, err := pgc.pool.Acquire(ctx)

	if err != nil {
		return WeeklyEventStats{}, err
	}

	var s WeeklyEventStats
	if err := conn.QueryRow(ctx, `SELECT * FROM data."v_WeeklyEventStats"`).Scan(&s.Week_total_events, &s.Daily_avg_events, &s.Week_max_force, &s.Week_min_force, &s.Week_avg_force); err != nil {
		return WeeklyEventStats{}, err
	}

	return s, nil
}

func (pgc *PGClient) RetrieveMonthlyStats(ctx context.Context) (MonthlyEventStats, error) {
	conn, err := pgc.pool.Acquire(ctx)

	if err != nil {
		return MonthlyEventStats{}, err
	}

	var s MonthlyEventStats
	if err := conn.QueryRow(ctx, `SELECT * FROM data."v_MonthlyEventStats"`).Scan(&s.Month_total_events, &s.Daily_avg_events, &s.Month_max_force, &s.Month_min_force, &s.Month_avg_force); err != nil {
		return MonthlyEventStats{}, err
	}

	return s, nil
}
