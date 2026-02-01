package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
	"github.com/dontpanicw/calendar/internal/port"
)

var (
	_ port.EventUsecases = (*UsecaseEvent)(nil)
)

type UsecaseEvent struct {
	repo port.EventRepository
}

func NewUsecaseEvent(repo port.EventRepository) *UsecaseEvent {
	return &UsecaseEvent{repo: repo}
}

func (u *UsecaseEvent) CreateEvent(ctx context.Context, event *domain.Event) error {
	if event.UserId <= 0 {
		return errors.New("invalid user id")
	}
	if event.Description == "" {
		return errors.New("event description is required")
	}
	return u.repo.CreateEvent(ctx, event)
}

func (u *UsecaseEvent) UpdateEvent(ctx context.Context, event domain.Event) error {
	if event.EventId <= 0 || event.UserId <= 0 {
		return errors.New("invalid event or user id")
	}
	return u.repo.UpdateEvent(ctx, event)
}

func (u *UsecaseEvent) DeleteEvent(ctx context.Context, eventId int64) error {
	if eventId <= 0 {
		return errors.New("invalid event id")
	}
	return u.repo.DeleteEvent(ctx, eventId)
}

func (u *UsecaseEvent) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	return u.repo.GetEventsForDay(ctx, userID, date)
}

func (u *UsecaseEvent) GetEventsForWeek(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	return u.repo.GetEventsForWeek(ctx, userID, start)
}

func (u *UsecaseEvent) GetEventsForMonth(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	return u.repo.GetEventsForMonth(ctx, userID, start)
}
