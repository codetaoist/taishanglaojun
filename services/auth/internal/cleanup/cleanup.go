package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/codetaoist/taishanglaojun/auth/internal/repository"
)

// CleanupWorker periodically cleans up expired sessions and blacklist entries
type CleanupWorker struct {
	sessionRepo  repository.SessionRepository
	blacklistRepo repository.BlacklistRepository
	interval     time.Duration
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(
	sessionRepo repository.SessionRepository,
	blacklistRepo repository.BlacklistRepository,
	interval time.Duration,
) *CleanupWorker {
	return &CleanupWorker{
		sessionRepo:  sessionRepo,
		blacklistRepo: blacklistRepo,
		interval:     interval,
	}
}

// Start starts the cleanup worker
func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Cleanup worker stopped")
			return
		case <-ticker.C:
			if err := w.cleanup(); err != nil {
				log.Printf("Cleanup failed: %v", err)
			}
		}
	}
}

// cleanup performs the cleanup of expired sessions and blacklist entries
func (w *CleanupWorker) cleanup() error {
	log.Println("Starting cleanup of expired sessions and blacklist entries")

	// Clean up expired sessions
	if err := w.sessionRepo.DeleteExpired(); err != nil {
		log.Printf("Failed to delete expired sessions: %v", err)
		return err
	}

	// Clean up expired blacklist entries
	if err := w.blacklistRepo.DeleteExpired(); err != nil {
		log.Printf("Failed to delete expired blacklist entries: %v", err)
		return err
	}

	log.Println("Cleanup completed successfully")
	return nil
}