package cache

import (
	"context"
	"testing"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
)

func TestCacheMap_CreateEvent(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	event := &domain.Event{UserId: 1, Date: date, Description: "Meeting"}

	err := c.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if event.EventId == 0 {
		t.Error("expected EventId to be assigned")
	}
	events, _ := c.GetEventsForDay(ctx, 1, date)
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestCacheMap_GetEventsForDay(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "A"})
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "B"})
	otherDay := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: otherDay, Description: "C"})

	events, err := c.GetEventsForDay(ctx, 1, date)
	if err != nil {
		t.Fatalf("GetEventsForDay: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for day, got %d", len(events))
	}
	eventsOther, _ := c.GetEventsForDay(ctx, 1, otherDay)
	if len(eventsOther) != 1 {
		t.Errorf("expected 1 event for other day, got %d", len(eventsOther))
	}
}

func TestCacheMap_GetEventsForWeek(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // понедельник
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: start, Description: "A"})
	mid := time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: mid, Description: "B"})
	outOfWeek := time.Date(2024, 1, 23, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: outOfWeek, Description: "C"})

	events, err := c.GetEventsForWeek(ctx, 1, start)
	if err != nil {
		t.Fatalf("GetEventsForWeek: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for week, got %d", len(events))
	}
}

func TestCacheMap_GetEventsForMonth(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), Description: "A"})
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Description: "B"})
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), Description: "C"})

	events, err := c.GetEventsForMonth(ctx, 1, start)
	if err != nil {
		t.Fatalf("GetEventsForMonth: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for month, got %d", len(events))
	}
}

func TestCacheMap_UpdateEvent(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "Old"})
	events, _ := c.GetEventsForDay(ctx, 1, date)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	id := events[0].EventId
	updated := domain.Event{EventId: id, UserId: 1, Date: date, Description: "New"}
	err := c.UpdateEvent(ctx, updated)
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	eventsAfter, _ := c.GetEventsForDay(ctx, 1, date)
	if len(eventsAfter) != 1 || eventsAfter[0].Description != "New" {
		t.Errorf("expected one event with Description New, got %v", eventsAfter)
	}
}

func TestCacheMap_UpdateEvent_NotFound(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	event := domain.Event{EventId: 999, UserId: 1, Date: time.Now(), Description: "X"}
	err := c.UpdateEvent(ctx, event)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
	if err.Error() != "event not found" {
		t.Errorf("expected 'event not found', got %q", err.Error())
	}
}

func TestCacheMap_DeleteEvent(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_ = c.CreateEvent(ctx, &domain.Event{UserId: 1, Date: date, Description: "X"})
	events, _ := c.GetEventsForDay(ctx, 1, date)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	id := events[0].EventId
	err := c.DeleteEvent(ctx, id)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	eventsAfter, _ := c.GetEventsForDay(ctx, 1, date)
	if len(eventsAfter) != 0 {
		t.Errorf("expected 0 events after delete, got %d", len(eventsAfter))
	}
}

func TestCacheMap_DeleteEvent_NotFound(t *testing.T) {
	ctx := context.Background()
	c := NewCacheMap()
	err := c.DeleteEvent(ctx, 999)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
	if err.Error() != "event not found" {
		t.Errorf("expected 'event not found', got %q", err.Error())
	}
}
