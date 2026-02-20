package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/calendar_test?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
	}

	// Очистка таблицы перед тестами
	_, _ = db.Exec("TRUNCATE TABLE events RESTART IDENTITY CASCADE")

	return db
}

func TestRepository_CreateEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	event := &domain.Event{
		UserId:      1,
		Date:        time.Now().Add(24 * time.Hour),
		IsArchived:  false,
		Description: "Test event",
	}

	err := repo.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if event.EventId == 0 {
		t.Error("EventId should be set after creation")
	}
}

func TestRepository_UpdateEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	// Создаем событие
	event := &domain.Event{
		UserId:      1,
		Date:        time.Now().Add(24 * time.Hour),
		IsArchived:  false,
		Description: "Original description",
	}
	_ = repo.CreateEvent(ctx, event)

	// Обновляем
	event.Description = "Updated description"
	err := repo.UpdateEvent(ctx, *event)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	// Проверяем обновление
	var desc string
	err = db.QueryRow("SELECT description FROM events WHERE event_id = $1", event.EventId).Scan(&desc)
	if err != nil {
		t.Fatalf("Failed to query updated event: %v", err)
	}

	if desc != "Updated description" {
		t.Errorf("Expected 'Updated description', got '%s'", desc)
	}
}

func TestRepository_DeleteEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	// Создаем событие
	event := &domain.Event{
		UserId:      1,
		Date:        time.Now().Add(24 * time.Hour),
		IsArchived:  false,
		Description: "To be deleted",
	}
	_ = repo.CreateEvent(ctx, event)

	// Удаляем
	err := repo.DeleteEvent(ctx, event.EventId)
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	// Проверяем, что удалено
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM events WHERE event_id = $1", event.EventId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query deleted event: %v", err)
	}

	if count != 0 {
		t.Error("Event should be deleted")
	}
}

func TestRepository_GetEventsForDay(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	targetDate := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	// Создаем события на разные дни
	event1 := &domain.Event{UserId: 1, Date: targetDate, Description: "Event 1"}
	event2 := &domain.Event{UserId: 1, Date: targetDate.Add(2 * time.Hour), Description: "Event 2"}
	event3 := &domain.Event{UserId: 1, Date: targetDate.Add(24 * time.Hour), Description: "Event 3"}

	_ = repo.CreateEvent(ctx, event1)
	_ = repo.CreateEvent(ctx, event2)
	_ = repo.CreateEvent(ctx, event3)

	// Получаем события за день
	events, err := repo.GetEventsForDay(ctx, 1, targetDate)
	if err != nil {
		t.Fatalf("GetEventsForDay failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestRepository_GetEventsForWeek(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	startDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	// Создаем события в пределах недели и за её пределами
	event1 := &domain.Event{UserId: 1, Date: startDate, Description: "Day 1"}
	event2 := &domain.Event{UserId: 1, Date: startDate.Add(3 * 24 * time.Hour), Description: "Day 4"}
	event3 := &domain.Event{UserId: 1, Date: startDate.Add(8 * 24 * time.Hour), Description: "Day 9"}

	_ = repo.CreateEvent(ctx, event1)
	_ = repo.CreateEvent(ctx, event2)
	_ = repo.CreateEvent(ctx, event3)

	// Получаем события за неделю
	events, err := repo.GetEventsForWeek(ctx, 1, startDate)
	if err != nil {
		t.Fatalf("GetEventsForWeek failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events within week, got %d", len(events))
	}
}

func TestRepository_GetEventsForMonth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	startDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	// Создаем события в марте и апреле
	event1 := &domain.Event{UserId: 1, Date: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Description: "March"}
	event2 := &domain.Event{UserId: 1, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), Description: "March"}
	event3 := &domain.Event{UserId: 1, Date: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC), Description: "April"}

	_ = repo.CreateEvent(ctx, event1)
	_ = repo.CreateEvent(ctx, event2)
	_ = repo.CreateEvent(ctx, event3)

	// Получаем события за март
	events, err := repo.GetEventsForMonth(ctx, 1, startDate)
	if err != nil {
		t.Fatalf("GetEventsForMonth failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events in March, got %d", len(events))
	}
}

func TestRepository_ArchiveOldEvents(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := &Repository{DB: db}
	ctx := context.Background()

	// Создаем старое и будущее событие
	oldEvent := &domain.Event{
		UserId:      1,
		Date:        time.Now().Add(-48 * time.Hour),
		IsArchived:  false,
		Description: "Old event",
	}
	futureEvent := &domain.Event{
		UserId:      1,
		Date:        time.Now().Add(48 * time.Hour),
		IsArchived:  false,
		Description: "Future event",
	}

	_ = repo.CreateEvent(ctx, oldEvent)
	_ = repo.CreateEvent(ctx, futureEvent)

	// Архивируем старые события
	err := repo.ArchiveOldEvents(ctx)
	if err != nil {
		t.Fatalf("ArchiveOldEvents failed: %v", err)
	}

	// Проверяем, что старое событие заархивировано
	var isArchived bool
	err = db.QueryRow("SELECT is_archived FROM events WHERE event_id = $1", oldEvent.EventId).Scan(&isArchived)
	if err != nil {
		t.Fatalf("Failed to query old event: %v", err)
	}

	if !isArchived {
		t.Error("Old event should be archived")
	}

	// Проверяем, что будущее событие не заархивировано
	err = db.QueryRow("SELECT is_archived FROM events WHERE event_id = $1", futureEvent.EventId).Scan(&isArchived)
	if err != nil {
		t.Fatalf("Failed to query future event: %v", err)
	}

	if isArchived {
		t.Error("Future event should not be archived")
	}
}
