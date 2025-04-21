package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"syncdocs/internal/config"
	"syncdocs/internal/syncer"
)

// Scheduler manages scheduled tasks.
type Scheduler struct {
	cron   *cron.Cron
	syncer *syncer.Syncer
	cfg    *config.Config
}

// NewScheduler creates a new Scheduler.
func NewScheduler(cfg *config.Config, syncer *syncer.Syncer) *Scheduler {
	// Create a new cron scheduler with seconds field support (optional)
	// c := cron.New(cron.WithSeconds())
	// Or standard cron without seconds:
	c := cron.New()

	return &Scheduler{
		cron:   c,
		syncer: syncer,
		cfg:    cfg,
	}
}

// Start initializes and starts the scheduled tasks.
func (s *Scheduler) Start() {
	log.Println("Starting task scheduler...")

	// Schedule the SyncAllRepositories task based on the configured interval.
	// cron format: "Minute Hour DayOfMonth Month DayOfWeek"
	// We need to convert the duration interval into a cron spec.
	// This is a bit tricky for arbitrary durations. A common approach is
	// to run frequently (e.g., every minute) and check internally if the
	// interval has passed, OR use a simpler cron spec like "@hourly", "@daily".
	// For simplicity, let's use the interval directly if it's a standard one,
	// otherwise default to hourly.

	// Convert duration to a cron-compatible spec string (simplistic approach)
	intervalSpec := fmt.Sprintf("@every %s", s.cfg.SyncInterval.String())
	log.Printf("Scheduling 'SyncAllRepositories' with spec: %s", intervalSpec)

	// Add the job to the scheduler
	_, err := s.cron.AddFunc(intervalSpec, func() {
		log.Println("Running scheduled task: SyncAllRepositories")
		// Use context.Background() for scheduled tasks as they are not tied to requests
		s.syncer.SyncAllRepositories(context.Background())
	})

	if err != nil {
		log.Fatalf("Error scheduling sync task: %v", err)
	}

	// Start the cron scheduler in a new goroutine
	go s.cron.Start()

	log.Println("Task scheduler started.")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	log.Println("Stopping task scheduler...")
	// Stop the scheduler. This waits for running jobs to complete.
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
		log.Println("Task scheduler stopped gracefully.")
	case <-time.After(10 * time.Second): // Add a timeout for shutdown
		log.Println("Task scheduler shutdown timed out.")
	}
}
