// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/log"
)

const (
	dbusPath      = dbus.ObjectPath("/org/freedesktop/login1")
	dbusInterface = "org.freedesktop.login1.Manager"
	dbusMember    = "PrepareForSleep"
	dbusSignal    = dbusInterface + "." + dbusMember

	signalBufferSize = 8
	debounceWindow   = 2 * time.Second
	retryDelay       = 5 * time.Second

	// networkWakeupDelay gives DHCP/DNS a chance to settle before we hit the network.
	networkWakeupDelay = 15 * time.Second
)

// monitorSleepResume watches logind's PrepareForSleep signal and refreshes the weather data
// whenever the machine comes back from suspend or hibernate. It reconnects to the system bus
// until ctx is canceled.
func (s *Server) monitorSleepResume(ctx context.Context) {
	for ctx.Err() == nil {
		if err := s.watchSleepSignals(ctx); err != nil && ctx.Err() == nil {
			s.log.Error("sleep monitor failed, retrying", log.ErrAttr(err), slog.Duration("delay", retryDelay))
		}
		if !wait(ctx, retryDelay) {
			return
		}
	}
}

// watchSleepSignals runs a single connect-subscribe-consume cycle. It returns when the bus
// connection dies or ctx is canceled.
func (s *Server) watchSleepSignals(ctx context.Context) error {
	conn, err := dbus.ConnectSystemBus(dbus.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err = conn.AddMatchSignalContext(ctx,
		dbus.WithMatchObjectPath(dbusPath),
		dbus.WithMatchInterface(dbusInterface),
		dbus.WithMatchMember(dbusMember),
	); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", dbusSignal, err)
	}

	sigCh := make(chan *dbus.Signal, signalBufferSize)
	conn.Signal(sigCh)
	s.log.Debug("subscribed to dbus signal", slog.String("signal", dbusSignal))

	var lastResume time.Time
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sgn, ok := <-sigCh:
			if !ok {
				return apperror.ErrDBusSignalChannelClosed
			}
			if !isResumeSignal(sgn) || time.Since(lastResume) < debounceWindow {
				continue
			}
			lastResume = time.Now()
			s.handleResume(ctx)
		}
	}
}

// isResumeSignal reports whether the provided signal is a PrepareForSleep signal announcing
// a wake-up
func isResumeSignal(sgn *dbus.Signal) bool {
	if sgn.Name != dbusSignal || len(sgn.Body) != 1 {
		return false
	}
	sleeping, ok := sgn.Body[0].(bool)
	return ok && !sleeping
}

// handleResume refreshes the weather data after the machine woke up.
func (s *Server) handleResume(ctx context.Context) {
	if !wait(ctx, networkWakeupDelay) {
		return
	}
	s.log.Debug("resumed from sleep, fetching latest weather data")

	/*
		s.weatherLock.Lock()
		s.weatherIsSet = false
		s.weatherLock.Unlock()

		s.fetchWeather(ctx)
		s.printWeather(ctx)
	*/
}

// wait sleeps for a given duration and reports whether it completed without ctx being canceled.
func wait(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
