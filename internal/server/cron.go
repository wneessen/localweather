package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/wneessen/localweather/internal/log"
)

func (s *Server) cronjobs(ctx context.Context) error {
	if err := s.newCronjob(ctx, "maintenance", s.conf.Scheduler.MaintenanceInterval, s.cronjobMaintenance); err != nil {
		return fmt.Errorf("failed to set up maintenance job: %w", err)
	}

	return nil
}

func (s *Server) newCronjob(ctx context.Context, name string, interval time.Duration, fn func(context.Context)) error {
	job, err := s.cron.NewJob(gocron.DurationJob(interval),
		gocron.NewTask(fn),
		gocron.WithContext(ctx),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to set up cron job %q: %w", name, err)
	}

	s.cron.Start()
	runat, err := job.NextRun()
	if err != nil {
		return fmt.Errorf("failed to get next runtime of cron job %q: %w", name, err)
	}
	s.log.Info("cron job successfully scheduled", slog.String("job_name", name),
		slog.String("job_id", job.ID().String()),
		slog.String("next_runtime", runat.Format(time.RFC3339)))

	return nil
}

func (s *Server) logJobCompletion(name string, startTime time.Time, failed bool) {
	s.log.Info("scheduled job completed",
		slog.Group("job_details",
			slog.String("job_name", name),
			slog.String("start_time", startTime.Format(time.RFC3339)),
			slog.String("runtime", time.Since(startTime).String()),
			slog.Bool("succeeded", !failed),
		),
	)
}

func (s *Server) logJobCompletionWithError(name string, startTime time.Time, err error) {
	s.log.Info("scheduled job completed",
		slog.Group("job_details",
			slog.String("job_name", name),
			slog.String("start_time", startTime.Format(time.RFC3339)),
			slog.String("runtime", time.Since(startTime).String()),
			slog.Bool("succeeded", false),
			log.ErrAttr(err),
		),
	)
}
