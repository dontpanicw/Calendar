package port

import (
	"context"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
)

// EventRepository интерфейс для работы с репозиторием событий
type EventRepository interface {
	CreateEvent(ctx context.Context, event *domain.Event) error
	UpdateEvent(ctx context.Context, event domain.Event) error
	DeleteEvent(ctx context.Context, eventId int64) error
	GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error)
	GetEventsForWeek(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error)
	GetEventsForMonth(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error)
}
