package app

import (
	"github.com/dontpanicw/calendar/config"
	"github.com/dontpanicw/calendar/internal/adapter/repository/cache"
	"github.com/dontpanicw/calendar/internal/input/http/handlers"
	"github.com/dontpanicw/calendar/internal/usecases"
	"log"
	"net/http"
)

func Start(cfg config.Config) error {

	eventRepo := cache.NewCacheMap()

	eventUsecase := usecases.NewUsecaseEvent(eventRepo)

	srv := handlers.NewServer(eventUsecase)

	log.Printf("Starting server on %s", cfg.HTTPPort)
	return http.ListenAndServe(cfg.HTTPPort, srv)
}
