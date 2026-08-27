// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geobus

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/types"
)

const (
	TruncPrecision  = 4
	accuracyEpsilon = 1e-6
)

// Provider defines an interface for geolocation service providers.
// It supports retrieving streamed results for a given key.
type Provider interface {
	Name() string
	LookupStream(ctx context.Context, key string) <-chan Result
}

// Service coordinates the publishing and subscribing of geolocation
// results between providers and consumers.
type Service struct {
	mu          sync.RWMutex
	best        map[string]Result
	subscribers map[string]map[chan Result]struct{}
	log         *log.Logger
}

// Result represents a geolocation result with associated metadata.
type Result struct {
	Key         string
	Coordinates types.Coordinate
	Provider    string
	At          time.Time
	TTL         time.Duration
}

// New initializes and returns a new instance of Service to handle
// geolocation result coordination.
func New(log *log.Logger) (*Service, error) {
	if log == nil {
		return nil, apperror.ErrLoggerRequired
	}
	return &Service{
		best:        make(map[string]Result),
		subscribers: make(map[string]map[chan Result]struct{}),
		log:         log,
	}, nil
}

// Subscribe adds a subscriber for updates associated with the given key and
// buffer size, returning a result channel and an unsubscribe function.
func (b *Service) Subscribe(key string, size int) (<-chan Result, func()) {
	ch := make(chan Result, size)

	b.mu.Lock()
	if _, ok := b.subscribers[key]; !ok {
		b.subscribers[key] = make(map[chan Result]struct{})
	}
	b.subscribers[key][ch] = struct{}{}

	// Immediately send the current best if we have it and it’s not expired.
	if best, ok := b.best[key]; ok && !best.IsExpired() {
		ch <- best
	}
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		if subs, ok := b.subscribers[key]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(b.subscribers, key)
			}
		}
		b.mu.Unlock()
		close(ch)
	}

	b.log.Debug("subscribed to geobus updates", slog.String("key", key))
	return ch, unsub
}

// Publish updates the best result for a key and notifies subscribers
func (b *Service) Publish(r Result) {
	// Ignore zero-accuracy results; they’re meaningless.
	if r.Coordinates.Accuracy <= 0 {
		return
	}

	// Ensure At is set.
	if r.At.IsZero() {
		r.At = time.Now()
	}

	newCoord := types.Coordinate{
		Accuracy:  r.Coordinates.Accuracy,
		Altitude:  r.Coordinates.Altitude,
		Latitude:  r.Coordinates.Latitude,
		Longitude: r.Coordinates.Longitude,
	}

	b.mu.Lock()
	superseded := false

	prev, have := b.best[r.Key]
	prevCoord := types.Coordinate{
		Altitude:  prev.Coordinates.Altitude,
		Accuracy:  prev.Coordinates.Accuracy,
		Latitude:  prev.Coordinates.Latitude,
		Longitude: prev.Coordinates.Longitude,
	}

	// If the result is not expired or better and the position has changed significantly, update it.
	if !have || prev.IsExpired() || r.BetterThan(prev) && newCoord.PositionHasSignificantChange(prevCoord) {
		superseded = true
	}

	b.log.Debug("a geobus provider published a new geolocation update",
		slog.Float64("accuracy", r.Coordinates.Accuracy.Float64()),
		slog.Float64("altitude", r.Coordinates.Altitude),
		slog.Float64("latitude", r.Coordinates.Latitude),
		slog.Float64("longitude", r.Coordinates.Longitude),
		slog.String("provider", r.Provider),
		slog.Bool("current_location_superseded", superseded),
	)
	if !superseded {
		b.mu.Unlock()
		return
	}

	b.best[r.Key] = r

	// Copy subscribers into a slice while we still hold the lock.
	var subs []chan Result
	if m, ok := b.subscribers[r.Key]; ok {
		subs = make([]chan Result, 0, len(m))
		for ch := range m {
			subs = append(subs, ch)
		}
	}
	b.mu.Unlock()

	// Non-blocking broadcast; slow subscribers just drop updates.
	for _, ch := range subs {
		select {
		case ch <- r:
		default:
		}
	}
}

// BetterThan compares two Result objects to determine if the current instance
// is better than the provided one.
func (r Result) BetterThan(prev Result) bool {
	if prev.Key == "" {
		return true
	}

	// Reject out-of-order results.
	if r.At.Before(prev.At) {
		return false
	}

	// More accurate?
	if r.Coordinates.Accuracy < prev.Coordinates.Accuracy-accuracyEpsilon {
		return true
	}
	if prev.Coordinates.Accuracy < r.Coordinates.Accuracy-accuracyEpsilon {
		return false
	}

	// Same-ish accuracy; we treat them as "not better".
	return false
}

// IsExpired checks if the Result has exceeded its time-to-live (TTL)
// based on the current time and the timestamp.
func (r Result) IsExpired() bool {
	return r.TTL > 0 && time.Since(r.At) > r.TTL
}

// Truncate truncates a float to a fixed decimal precision.
func Truncate(x float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Trunc(x*p) / p
}

// TrackProviders starts one goroutine per provider that streams results into the bus.
// It returns immediately; goroutines exit when ctx is canceled or the provider channel closes.
func TrackProviders(ctx context.Context, bus *Service, key string, providers ...Provider) {
	for _, p := range providers {
		go func() {
			ch := p.LookupStream(ctx, key)
			for {
				select {
				case <-ctx.Done():
					return
				case r, ok := <-ch:
					if !ok {
						return
					}
					bus.Publish(r)
				}
			}
		}()
	}
}
