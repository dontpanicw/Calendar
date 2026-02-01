package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
	"github.com/dontpanicw/calendar/internal/port"
)

var (
	_ port.EventRepository = (*CacheMap)(nil)
)

type CacheMap struct {
	mu     sync.RWMutex
	events map[int64]domain.Event
	nextID int64
}

func NewCacheMap() *CacheMap {
	return &CacheMap{
		events: make(map[int64]domain.Event, 128),
		nextID: 1,
	}
}

func (c *CacheMap) CreateEvent(ctx context.Context, event *domain.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	event.EventId = c.nextID
	c.nextID++
	c.events[event.EventId] = *event
	return nil
}

func (c *CacheMap) UpdateEvent(ctx context.Context, event domain.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.events[event.EventId]
	if !ok {
		return errors.New("event not found")
	}
	c.events[event.EventId] = event
	return nil
}

func (c *CacheMap) DeleteEvent(ctx context.Context, eventId int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.events[eventId]; !ok {
		return errors.New("event not found")
	}
	delete(c.events, eventId)
	return nil
}

func sameDay(a, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()
	return ya == yb && ma == mb && da == db
}

func inWeek(start time.Time, t time.Time) bool {
	weekEnd := start.AddDate(0, 0, 7)
	return !t.Before(start) && t.Before(weekEnd)
}

func inMonth(start time.Time, t time.Time) bool {
	monthEnd := start.AddDate(0, 1, 0)
	return !t.Before(start) && t.Before(monthEnd)
}

func (c *CacheMap) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []domain.Event
	for _, e := range c.events {
		if e.UserId == userID && sameDay(e.Date, date) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (c *CacheMap) GetEventsForWeek(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []domain.Event
	for _, e := range c.events {
		if e.UserId == userID && inWeek(start, e.Date) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (c *CacheMap) GetEventsForMonth(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []domain.Event
	for _, e := range c.events {
		if e.UserId == userID && inMonth(start, e.Date) {
			result = append(result, e)
		}
	}
	return result, nil
}
