package cleaning_worker

import (
	"context"
	"log"
	"time"
)

//Чистка событий: отдельная горутина, каждые X минут должна переносить в архив старые события

type CleaningWorker struct {
	period uint32
	repo   RepoProvider
}

type RepoProvider interface {
	ArchiveOldEvents(ctx context.Context) error
}

func NewCleaningWorker(period uint32, repo RepoProvider) *CleaningWorker {
	return &CleaningWorker{
		period: period,
		repo:   repo,
	}
}

func (c *CleaningWorker) Start(ctx context.Context) {

	c.runCleanup(ctx)

	ticker := time.NewTicker(time.Duration(c.period) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.runCleanup(ctx)
		case <-ctx.Done():
			log.Println("Cleaning worker stopped")
			return
		}
	}
}

func (c *CleaningWorker) runCleanup(ctx context.Context) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err := c.repo.ArchiveOldEvents(cleanupCtx)
	cancel()

	if err != nil {
		log.Printf("Error archiving old events: %v", err)
	}
}
