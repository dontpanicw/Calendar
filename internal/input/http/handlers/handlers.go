package handlers

import (
	"encoding/json"
	"errors"
	"github.com/dontpanicw/calendar/log_worker"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dontpanicw/calendar/internal/domain"
	"github.com/dontpanicw/calendar/internal/input/http/types"
	"github.com/dontpanicw/calendar/internal/port"
)

const dateLayout = "2006-01-02"

type Handler struct {
	usecases port.EventUsecases
	logger   *log_worker.Logger
}

func NewHandler(usecases port.EventUsecases, logger *log_worker.Logger) *Handler {
	return &Handler{
		usecases: usecases,
		logger:   logger,
	}
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	event, err := parseBodyCreate(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.usecases.CreateEvent(r.Context(), event); err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResultMessage(w, "event created")
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	event, err := parseBodyUpdate(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.usecases.UpdateEvent(r.Context(), event); err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResultMessage(w, "event updated")
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := parseBodyDelete(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.usecases.DeleteEvent(r.Context(), eventID); err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResultMessage(w, "event deleted")
}

func (h *Handler) EventsForDay(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryUserDate(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	events, err := h.usecases.GetEventsForDay(r.Context(), userID, date)
	if err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResult(w, events)
}

func (h *Handler) EventsForWeek(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryUserDate(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Неделя: от начала дня date до конца недели (7 дней)
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	events, err := h.usecases.GetEventsForWeek(r.Context(), userID, start)
	if err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResult(w, events)
}

func (h *Handler) EventsForMonth(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryUserDate(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Месяц: первый день месяца
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	events, err := h.usecases.GetEventsForMonth(r.Context(), userID, start)
	if err != nil {
		if isBusinessError(err) {
			writeError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResult(w, events)
}

// writeResult отправляет успешный ответ 200 OK с {"result": ...}
func writeResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.EventsResponse{Result: result})
}

// writeResultMessage отправляет 200 OK с {"result": "строка"}
func writeResultMessage(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.APIResponse{Result: msg})
}

// writeError отправляет JSON {"error": "..."} с заданным статусом
func writeError(w http.ResponseWriter, errMsg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.ErrorResponse{Error: errMsg})
}

// parseDate проверяет формат YYYY-MM-DD
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("date is required")
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, use YYYY-MM-DD")
	}
	return t, nil
}

// parseBodyCreate парсит JSON или form для создания события
func parseBodyCreate(r *http.Request) (*domain.Event, error) {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var req types.CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, errors.New("invalid JSON body")
		}
		date, err := parseDate(req.Date)
		if err != nil {
			return nil, err
		}
		if req.UserID <= 0 {
			return nil, errors.New("user_id is required and must be positive")
		}
		return &domain.Event{UserId: req.UserID, Date: date, Description: req.Event}, nil
	}
	if err := r.ParseForm(); err != nil {
		return nil, errors.New("invalid form body")
	}
	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		return nil, errors.New("user_id is required and must be positive integer")
	}
	date, err := parseDate(r.FormValue("date"))
	if err != nil {
		return nil, err
	}
	desc := r.FormValue("event")
	return &domain.Event{UserId: userID, Date: date, Description: desc}, nil
}

// parseBodyUpdate парсит JSON или form для обновления события
func parseBodyUpdate(r *http.Request) (domain.Event, error) {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var req types.UpdateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return domain.Event{}, errors.New("invalid JSON body")
		}
		date, err := parseDate(req.Date)
		if err != nil {
			return domain.Event{}, err
		}
		if req.EventID <= 0 || req.UserID <= 0 {
			return domain.Event{}, errors.New("event_id and user_id are required and must be positive")
		}
		return domain.Event{EventId: req.EventID, UserId: req.UserID, Date: date, Description: req.Event}, nil
	}
	if err := r.ParseForm(); err != nil {
		return domain.Event{}, errors.New("invalid form body")
	}
	eventID, err := strconv.ParseInt(r.FormValue("event_id"), 10, 64)
	if err != nil || eventID <= 0 {
		return domain.Event{}, errors.New("event_id is required and must be positive integer")
	}
	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		return domain.Event{}, errors.New("user_id is required and must be positive integer")
	}
	date, err := parseDate(r.FormValue("date"))
	if err != nil {
		return domain.Event{}, err
	}
	return domain.Event{EventId: eventID, UserId: userID, Date: date, Description: r.FormValue("event")}, nil
}

// parseBodyDelete парсит JSON или form для удаления события
func parseBodyDelete(r *http.Request) (int64, error) {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var req types.DeleteEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return 0, errors.New("invalid JSON body")
		}
		if req.EventID <= 0 {
			return 0, errors.New("event_id is required and must be positive")
		}
		return req.EventID, nil
	}
	if err := r.ParseForm(); err != nil {
		return 0, errors.New("invalid form body")
	}
	eventID, err := strconv.ParseInt(r.FormValue("event_id"), 10, 64)
	if err != nil || eventID <= 0 {
		return 0, errors.New("event_id is required and must be positive integer")
	}
	return eventID, nil
}

// parseQueryUserDate парсит query user_id и date для GET
func parseQueryUserDate(r *http.Request) (userID int64, date time.Time, err error) {
	userID, err = strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		return 0, time.Time{}, errors.New("user_id is required and must be positive integer")
	}
	dateStr := r.URL.Query().Get("date")
	date, err = parseDate(dateStr)
	if err != nil {
		return 0, time.Time{}, err
	}
	return userID, date, nil
}

// isBusinessError ошибки бизнес-логики (событие не найдено и т.д.) — 503
func isBusinessError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return s == "event not found" || s == "event already exists"
}
