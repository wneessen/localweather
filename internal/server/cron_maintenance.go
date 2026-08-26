package server

import (
	"context"
	"time"
)

func (s *Server) cronjobMaintenance(ctx context.Context) {
	now := time.Now()
	const action = "maintenance"
	_ = ctx

	s.logJobCompletion(action, now, false)
}
