// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geobus

import (
	"context"
	"testing"
	"time"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/log"

	"github.com/wneessen/localweather/internal/types"
)

const (
	subID = "test"
)

func TestGeolocationState_Update(t *testing.T) {
	state := GeoLocationState{}
	state.Update(types.Coordinate{Latitude: 50.0, Longitude: 8.0})
	if state.last.Latitude != 50.0 || state.last.Longitude != 8.0 {
		t.Error("expected last coordinate to be updated")
	}
	if !state.haveLast {
		t.Error("expected haveLast to be true")
	}
}

func TestCoordinate_PosHasSignificantChange(t *testing.T) {
	tests := []struct {
		name    string
		coord   types.Coordinate
		other   types.Coordinate
		changed bool
	}{
		{
			name: "same point, no change",
			coord: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
			},
			other: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
			},
			changed: false,
		},
		{
			name: "small move within threshold",
			coord: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
			},
			other: types.Coordinate{
				Latitude:  50.01,
				Longitude: 8.01,
			},
			changed: false,
		},
		{
			name: "small move within threshold with negative coordinates",
			coord: types.Coordinate{
				Latitude:  -50.0,
				Longitude: -8.0,
			},
			other: types.Coordinate{
				Latitude:  -50.01,
				Longitude: -8.01,
			},
			changed: false,
		},
		{
			name: "just above threshold",
			coord: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
			},
			other: types.Coordinate{
				Latitude:  50.0225,
				Longitude: 8.0,
			},
			changed: true,
		},
		{
			name: "far above threshold but negative coordinates",
			coord: types.Coordinate{
				Latitude:  -50.0,
				Longitude: -8.0,
			},
			other: types.Coordinate{
				Latitude:  -50.0,
				Longitude: -8.1,
			},
			changed: true,
		},
		{
			name: "large movement, Berlin to Paris",
			coord: types.Coordinate{
				Latitude:  52.52,
				Longitude: 13.405,
			},
			other: types.Coordinate{
				Latitude:  48.8566,
				Longitude: 2.3522,
			},
			changed: true,
		},
		{
			name: "same place but significantly better accuracy",
			coord: types.Coordinate{
				Latitude:  52.52,
				Longitude: 13.405,
				Accuracy:  types.AccuracyCity,
			},
			other: types.Coordinate{
				Latitude:  52.52,
				Longitude: 13.405,
				Accuracy:  types.AccuracyCountry,
			},
			changed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.coord.PositionHasSignificantChange(tc.other) != tc.changed {
				t.Error("expected change to be", tc.changed, "but it wasn't")
			}
		})
	}
}

func TestResult_BetterThan(t *testing.T) {
	tests := []struct {
		name   string
		new    Result
		prev   Result
		better bool
	}{
		{
			name:   "same result, no difference",
			new:    Result{Key: "test", Coordinates: types.Coordinate{Latitude: 1, Longitude: 1, Accuracy: 1}},
			prev:   Result{Key: "test", Coordinates: types.Coordinate{Latitude: 1, Longitude: 1, Accuracy: 1}},
			better: false,
		},
		{
			name:   "previous result had no key",
			new:    Result{Key: "test", Coordinates: types.Coordinate{Latitude: 1, Longitude: 1, Accuracy: 1}},
			prev:   Result{Coordinates: types.Coordinate{Latitude: 1, Longitude: 1, Accuracy: 1}},
			better: true,
		},
		{
			name:   "previous result is newer",
			new:    Result{Key: "test", At: time.Date(2024, time.January, 1, 16, 56, 0, 0, time.UTC)},
			prev:   Result{Key: "test", At: time.Date(2025, time.January, 1, 16, 56, 0, 0, time.UTC)},
			better: false,
		},
		{
			name:   "new result is more accurate",
			new:    Result{Key: "test", Coordinates: types.Coordinate{Accuracy: types.AccuracyZip}},
			prev:   Result{Key: "test", Coordinates: types.Coordinate{Accuracy: types.AccuracyCity}},
			better: true,
		},
		{
			name:   "previous result is more accurate",
			new:    Result{Key: "test", Coordinates: types.Coordinate{Accuracy: types.AccuracyCity}},
			prev:   Result{Key: "test", Coordinates: types.Coordinate{Accuracy: types.AccuracyZip}},
			better: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.new.BetterThan(tc.prev) != tc.better {
				t.Error("expected new to be", tc.better, "but it wasn't")
			}
		})
	}
}

func TestResult_IsExpired(t *testing.T) {
	result := Result{At: time.Now().Add(-time.Hour), TTL: time.Hour}
	if !result.IsExpired() {
		t.Error("expected result to be expired")
	}
	result = Result{At: time.Now().Add(-time.Hour), TTL: time.Hour * 2}
	if result.IsExpired() {
		t.Error("expected result to not be expired")
	}
}

func TestNew(t *testing.T) {
	logger := log.New(new(config.Config))
	bus, err := New(logger)
	if err != nil {
		t.Fatalf("failed to create bus: %s", err)
	}
	if bus == nil {
		t.Fatal("expected bus to be non-nil")
	}
	if bus.best == nil {
		t.Fatal("expected best provider to be non-nil")
	}
	if bus.subscribers == nil {
		t.Fatal("expected subscribers to be non-nil")
	}
}

func TestGeoBus_Publish(t *testing.T) {
	logger := log.New(new(config.Config))
	t.Run("a siggnificant change publishes a result", func(t *testing.T) {
		bus, err := New(logger)
		if err != nil {
			t.Fatalf("failed to create bus: %s", err)
		}
		ch, unsub := bus.Subscribe(subID, 1)
		defer unsub()

		bus.Publish(Result{
			Key: subID,
			Coordinates: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
				Accuracy:  20,
			},
			At:       time.Now(),
			Provider: "mock-provider",
		})
		<-ch
		bus.Publish(Result{
			Key: subID,
			Coordinates: types.Coordinate{
				Latitude:  55.0001,
				Longitude: 9.0001,
				Accuracy:  20,
			},
			At:       time.Now(),
			Provider: "mock-provider",
		})
		select {
		case <-ch:
			t.Fatalf("did not expect update for insignificant movement")
		case <-time.After(50 * time.Millisecond):
		}
	})
	t.Run("do not publish results without accuracy", func(t *testing.T) {
		bus, err := New(logger)
		if err != nil {
			t.Fatalf("failed to create bus: %s", err)
		}
		ch, unsub := bus.Subscribe(subID, 1)
		defer unsub()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		bus.Publish(Result{
			Key: subID,
			Coordinates: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
				Accuracy:  0,
			},
			At:       time.Now(),
			Provider: "mock-provider",
			TTL:      time.Millisecond * 500,
		})

		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				t.Fatalf("did not expect update for insignificant movement")
			}
		}
	})
	t.Run("no At time sets it to 'now'", func(t *testing.T) {
		bus, err := New(logger)
		if err != nil {
			t.Fatalf("failed to create bus: %s", err)
		}
		ch, unsub := bus.Subscribe(subID, 1)
		defer unsub()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		bus.Publish(Result{
			Key: subID,
			Coordinates: types.Coordinate{
				Latitude:  50.0,
				Longitude: 8.0,
				Accuracy:  1,
			},
			Provider: "mock-provider",
			TTL:      time.Millisecond * 500,
		})

		var result *Result
		for result == nil {
			select {
			case <-ctx.Done():
				return
			case r := <-ch:
				result = &r
			}
		}
		if result == nil {
			t.Fatal("expected result to be non-nil")
		}
		if result.At.IsZero() {
			t.Fatal("expected At time to be set")
		}
	})
}

func TestTrackProviders(t *testing.T) {
	logger := log.New(new(config.Config))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	bus, err := New(logger)
	if err != nil {
		t.Fatalf("failed to create bus: %s", err)
	}
	fp := &fakeProvider{name: "test", ch: make(chan Result, 1)}
	TrackProviders(ctx, bus, "k", fp)

	sub, unsub := bus.Subscribe(subID, 1)
	defer unsub()

	r := Result{
		Key: subID,
		Coordinates: types.Coordinate{
			Latitude:  1,
			Longitude: 2,
			Accuracy:  10,
		},
		At:  time.Now(),
		TTL: time.Millisecond * 500,
	}
	fp.ch <- r

	got := <-sub
	if got.Coordinates.Latitude != r.Coordinates.Latitude || got.Coordinates.Longitude != r.Coordinates.Longitude {
		t.Fatalf("unexpected result: %+v", got)
	}
}

type fakeProvider struct {
	name string
	ch   chan Result
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) LookupStream(context.Context, string) <-chan Result {
	return f.ch
}
