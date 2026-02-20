package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
	"github.com/dontpanicw/calendar/log_worker"
)

type MockUsecases struct {
	events map[int64]*domain.Event
	nextID int64
}

func NewMockUsecases() *MockUsecases {
	return &MockUsecases{
		events: make(map[int64]*domain.Event),
		nextID: 1,
	}
}

func (m *MockUsecases) CreateEvent(ctx context.Context, event *domain.Event) error {
	event.EventId = m.nextID
	m.nextID++
	m.events[event.EventId] = event
	return nil
}

func (m *MockUsecases) UpdateEvent(ctx context.Context, event domain.Event) error {
	m.events[event.EventId] = &event
	return nil
}

func (m *MockUsecases) DeleteEvent(ctx context.Context, eventId int64) error {
	delete(m.events, eventId)
	return nil
}

func (m *MockUsecases) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	var result []domain.Event
	for _, event := range m.events {
		if event.UserId == userID {
			result = append(result, *event)
		}
	}
	return result, nil
}

func (m *MockUsecases) GetEventsForWeek(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	return m.GetEventsForDay(ctx, userID, start)
}

func (m *MockUsecases) GetEventsForMonth(ctx context.Context, userID int64, start time.Time) ([]domain.Event, error) {
	return m.GetEventsForDay(ctx, userID, start)
}

func TestHandler_CreateEvent(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	body := map[string]interface{}{
		"user_id": 1,
		"date":    "2026-03-15",
		"event":   "Test event",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/create_event", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["result"] != "event created" {
		t.Errorf("Expected 'event created', got '%s'", response["result"])
	}
}

func TestHandler_CreateEvent_InvalidJSON(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	req := httptest.NewRequest("POST", "/create_event", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_UpdateEvent(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	// Создаем событие
	event := &domain.Event{UserId: 1, Date: time.Now(), Description: "Original"}
	usecases.CreateEvent(context.Background(), event)

	body := map[string]interface{}{
		"event_id": event.EventId,
		"user_id":  1,
		"date":     "2026-03-16",
		"event":    "Updated",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/update_event", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandler_DeleteEvent(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	// Создаем событие
	event := &domain.Event{UserId: 1, Date: time.Now(), Description: "To delete"}
	usecases.CreateEvent(context.Background(), event)

	body := map[string]interface{}{
		"event_id": event.EventId,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/delete_event", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DeleteEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if _, exists := usecases.events[event.EventId]; exists {
		t.Error("Event should be deleted")
	}
}

func TestHandler_EventsForDay(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	// Создаем событие
	event := &domain.Event{UserId: 1, Date: time.Now(), Description: "Test"}
	usecases.CreateEvent(context.Background(), event)

	req := httptest.NewRequest("GET", "/events_for_day?user_id=1&date=2026-03-15", nil)
	w := httptest.NewRecorder()

	handler.EventsForDay(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["result"] == nil {
		t.Error("Expected result field")
	}
}

func TestHandler_EventsForDay_InvalidUserID(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	req := httptest.NewRequest("GET", "/events_for_day?user_id=invalid&date=2026-03-15", nil)
	w := httptest.NewRecorder()

	handler.EventsForDay(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_EventsForDay_InvalidDate(t *testing.T) {
	usecases := NewMockUsecases()
	logger := log_worker.NewLogger()
	handler := NewHandler(usecases, logger)

	req := httptest.NewRequest("GET", "/events_for_day?user_id=1&date=invalid", nil)
	w := httptest.NewRecorder()

	handler.EventsForDay(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
