package server

import (
	"context"
	"time"
)

func (s *Server) cronjobMaintenance(_ context.Context) {
	now := time.Now()
	const action = "maintenance"

	s.logJobCompletion(action, now, false)
}
