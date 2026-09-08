// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

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
