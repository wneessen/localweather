package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	"github.com/wneessen/localweather/internal/log"
)

type cronParams struct {
	name     string
	interval time.Duration
	fn       func(context.Context)
	idFn     func(uuid.UUID)
}

func (s *Server) cronjobs(ctx context.Context) error {
	jobs := []cronParams{
		{
			name:     "maintenance",
			interval: s.conf.Scheduler.MaintenanceInterval,
			fn:       s.cronjobMaintenance,
			idFn:     nil,
		},
		{
			name:     "weatherdata_update",
			interval: s.conf.Scheduler.WeatherUpdateInterval,
			fn:       s.cronjobWeatherdataUpdate,
			idFn:     func(id uuid.UUID) { s.weatherJobID = id },
		},
	}
	for _, job := range jobs {
		if err := s.newCronjob(ctx, job); err != nil {
			return fmt.Errorf("failed to set up cron job %q: %w", job.name, err)
		}
	}

	return nil
}

func (s *Server) newCronjob(ctx context.Context, params cronParams) error {
	job, err := s.cron.NewJob(gocron.DurationJob(params.interval),
		gocron.NewTask(params.fn),
		gocron.WithContext(ctx),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName(params.name),
	)
	if err != nil {
		return fmt.Errorf("failed to set up cron job %q: %w", params.name, err)
	}

	s.cron.Start()
	runat, err := job.NextRun()
	if err != nil {
		return fmt.Errorf("failed to get next runtime of cron job %q: %w", params.name, err)
	}
	if params.idFn != nil {
		params.idFn(job.ID())
	}
	s.log.Info("cron job successfully scheduled", slog.String("job_name", params.name),
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
