package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dontpanicw/calendar/cleaning_worker"
	"github.com/dontpanicw/calendar/config"
	"github.com/dontpanicw/calendar/internal/adapter/repository/postgres"
	"github.com/dontpanicw/calendar/internal/input/http/handlers"
	"github.com/dontpanicw/calendar/internal/usecases"
	"github.com/dontpanicw/calendar/log_worker"
	"github.com/dontpanicw/calendar/pkg/migrations"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(cfg *config.Config) error {

	// 1. Создаём контекст, который отменяется при нажатии Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := log_worker.NewLogger()
	go logger.Log(ctx)

	// Retry подключения к PostgreSQL
	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", cfg.PostgresConnStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		logger.Writef("Waiting for PostgreSQL... (attempt %d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			logger.Writef("Failed to close database: %v", err)
		}
	}()
	logger.Write("Connected to PostgreSQL")

	if err := migrations.Migrate(db); err != nil {
		logger.Writef("error %v", err)
		return err
	}
	logger.Write("Migrations applied successfully")

	//eventCache := cache.NewCacheMap()
	eventRepo, err := postgres.NewRepository(cfg, logger)
	if err != nil {
		return err
	}

	cleaningWorker := cleaning_worker.NewCleaningWorker(10, eventRepo)
	go cleaningWorker.Start(ctx)

	eventUsecase := usecases.NewUsecaseEvent(eventRepo, logger)
	srv := handlers.NewServer(eventUsecase, logger)

	httpServer := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Writef("Starting server on %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Writef("HTTP server error: %v", err)
		}
	}()

	// Ждём сигнала остановки
	<-ctx.Done()
	logger.Write("Shutting down gracefully...")

	// Даём серверу и воркеру время на завершение
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Writef("HTTP shutdown error: %v", err)
	}

	logger.Write("Application stopped")
	return nil
}
