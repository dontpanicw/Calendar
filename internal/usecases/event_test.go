package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/dontpanicw/calendar/internal/adapter/repository/cache"
	"github.com/dontpanicw/calendar/internal/domain"
)

func TestUsecaseEvent_CreateEvent(t *testing.T) {
	ctx := context.Background()
	repo := cache.NewCacheMap()
	uc := NewUsecaseEvent(repo)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	err := uc.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "Test"})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
}

func TestUsecaseEvent_CreateEvent_InvalidUserID(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecaseEvent(cache.NewCacheMap())
	err := uc.CreateEvent(ctx, &domain.Event{UserId: 0, Date: time.Now(), Description: "X"})
	if err == nil {
		t.Fatal("expected error for invalid user id")
	}
	if err.Error() != "invalid user id" {
		t.Errorf("expected 'invalid user id', got %q", err.Error())
	}
}

func TestUsecaseEvent_CreateEvent_EmptyDescription(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecaseEvent(cache.NewCacheMap())
	err := uc.CreateEvent(ctx, &domain.Event{UserId: 1, Date: time.Now(), Description: ""})
	if err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestUsecaseEvent_GetEventsForDay(t *testing.T) {
	ctx := context.Background()
	repo := cache.NewCacheMap()
	uc := NewUsecaseEvent(repo)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_ = uc.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "A"})
	_ = uc.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "B"})

	events, err := uc.GetEventsForDay(ctx, 1, date)
	if err != nil {
		t.Fatalf("GetEventsForDay: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestUsecaseEvent_GetEventsForDay_InvalidUserID(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecaseEvent(cache.NewCacheMap())
	_, err := uc.GetEventsForDay(ctx, 0, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid user id")
	}
}

func TestUsecaseEvent_DeleteEvent(t *testing.T) {
	ctx := context.Background()
	repo := cache.NewCacheMap()
	uc := NewUsecaseEvent(repo)
	event := &domain.Event{UserId: 1, Date: time.Now(), Description: "X"}
	_ = uc.CreateEvent(ctx, event)

	err := uc.DeleteEvent(ctx, event.EventId)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
}

func TestUsecaseEvent_DeleteEvent_NotFound(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecaseEvent(cache.NewCacheMap())
	err := uc.DeleteEvent(ctx, 999)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
	if err.Error() != "event not found" {
		t.Errorf("expected 'event not found', got %q", err.Error())
	}
}
