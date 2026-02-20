package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dontpanicw/calendar/config"
	"github.com/dontpanicw/calendar/internal/domain"
	"github.com/dontpanicw/calendar/internal/port"
	"github.com/dontpanicw/calendar/log_worker"
	"time"
)

const (
	updateArchiveEventsQuery = `UPDATE events 
						  SET is_archived = true 
						  WHERE date < NOW() AND is_archived = false;`
	updateEventsQuery = `UPDATE events 
			  SET user_id = $1, date = $2, is_archived = $3, description = $4, updated_at = NOW() 
			  WHERE event_id = $5`
	deleteEventQuery     = `DELETE FROM events WHERE event_id = $1`
	getEventsForDayQuery = `SELECT event_id, user_id, date, is_archived, description 
			  FROM events 
			  WHERE user_id = $1 AND DATE(date) = DATE($2)
			  ORDER BY date`
	getEventsForWeekQuery = `SELECT event_id, user_id, date, is_archived, description
				FROM events
				WHERE user_id = $1 AND date >= $2 AND date < $2 + INTERVAL '7 days'
				ORDER BY date`
	getEventsForMonthQuery = `SELECT event_id, user_id, date, is_archived, description 
			  FROM events 
			  WHERE user_id = $1 AND date >= $2 AND date < $2 + INTERVAL '1 month'
			  ORDER BY date`
)

type Repository struct {
	DB     *sql.DB
	logger *log_worker.Logger
}

var (
	_ port.EventRepository = (*Repository)(nil)
)

func NewRepository(cfg *config.Config, logger *log_worker.Logger) (*Repository, error) {
	db, err := sql.Open("postgres", cfg.PostgresConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(25)                 // Максимум одновременных соединений
	db.SetMaxIdleConns(25)                 // Максимум простаивающих соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения
	db.SetConnMaxIdleTime(2 * time.Minute) // Максимальное время простоя

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Repository{
		DB:     db,
		logger: logger,
	}, nil
}

func (r *Repository) CreateEvent(ctx context.Context, event *domain.Event) error {
	query := `INSERT INTO events (user_id, date, is_archived, description) 
			  VALUES ($1, $2, $3, $4) 
			  RETURNING event_id`

	err := r.DB.QueryRowContext(ctx, query, event.UserId, event.Date, event.IsArchived, event.Description).Scan(&event.EventId)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (r *Repository) UpdateEvent(ctx context.Context, event domain.Event) error {
	result, err := r.DB.ExecContext(ctx, updateEventsQuery, event.UserId, event.Date, event.IsArchived, event.Description, event.EventId)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event with id %d not found", event.EventId)
	}

	return nil
}

func (r *Repository) DeleteEvent(ctx context.Context, eventId int64) error {
	result, err := r.DB.ExecContext(ctx, deleteEventQuery, eventId)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event with id %d not found", eventId)
	}

	return nil
}

func (r *Repository) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	rows, err := r.DB.QueryContext(ctx, getEventsForDayQuery, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for day: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.EventId, &event.UserId, &event.Date, &event.IsArchived, &event.Description); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

func (r *Repository) GetEventsForWeek(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {

	rows, err := r.DB.QueryContext(ctx, getEventsForWeekQuery, userID, start)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for week: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.EventId, &event.UserId, &event.Date, &event.IsArchived, &event.Description); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

func (r *Repository) GetEventsForMonth(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {

	rows, err := r.DB.QueryContext(ctx, getEventsForMonthQuery, userID, start)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for month: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.EventId, &event.UserId, &event.Date, &event.IsArchived, &event.Description); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

func (r *Repository) ArchiveOldEvents(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, updateArchiveEventsQuery)
	if err != nil {
		return fmt.Errorf("failed to archive old events: %w", err)
	}

	return nil
}
