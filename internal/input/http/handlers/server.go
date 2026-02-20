package handlers

import (
	"github.com/dontpanicw/calendar/log_worker"
	"log"
	"net/http"
	"time"

	"github.com/dontpanicw/calendar/internal/port"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(usecases port.EventUsecases, logger *log_worker.Logger) *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	h := NewHandler(usecases, logger)

	s.mux.HandleFunc("POST /create_event", h.CreateEvent)
	s.mux.HandleFunc("POST /update_event", h.UpdateEvent)
	s.mux.HandleFunc("POST /delete_event", h.DeleteEvent)
	s.mux.HandleFunc("GET /events_for_day", h.EventsForDay)
	s.mux.HandleFunc("GET /events_for_week", h.EventsForWeek)
	s.mux.HandleFunc("GET /events_for_month", h.EventsForMonth)

	return s
}

// loggingMiddleware логирует каждый запрос: метод, URL, время (на stdout)
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.String(), time.Since(start))
	})
}

// recoveryMiddleware восстанавливает панику и возвращает 500
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP реализует http.Handler с middleware: recovery, затем логирование
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := loggingMiddleware(recoveryMiddleware(s.mux))
	handler.ServeHTTP(w, r)
}
