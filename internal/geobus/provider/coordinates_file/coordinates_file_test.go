// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package coordinates_file

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/testhelper"
	"github.com/wneessen/localweather/internal/types"
)

const (
	testFile = "../../../../testdata/coordinates"
	testLat  = 40.7185
	testLon  = -74.0025
)

func TestNewGeolocationFileProvider(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	t.Run("new geolocation file provider succeeds", func(t *testing.T) {
		provider := NewCoordinatesFileProvider(testFile, logger)
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
	})
}

func TestGeolocationFileProvider_Name(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	provider := NewCoordinatesFileProvider(testFile, logger)
	if provider == nil {
		t.Fatal("expected provider to be non-nil")
	}
	if !strings.EqualFold(provider.Name(), name) {
		t.Errorf("expected provider name to be %s, got %s", name, provider.Name())
	}
}

func TestNewGeolocationFileProvider_readFile(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	t.Run("read file succeeds", func(t *testing.T) {
		provider := NewCoordinatesFileProvider(testFile, logger)
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		coords, err := provider.readFile(t.Context())
		if err != nil {
			t.Fatalf("failed to read file: %s", err)
		}
		if coords.Latitude != testLat {
			t.Errorf("expected latitude to be %f, got %f", testLat, coords.Latitude)
		}
		if coords.Longitude != testLon {
			t.Errorf("expected longitude to be %f, got %f", testLon, coords.Longitude)
		}
	})
	t.Run("reading invalid file fails", func(t *testing.T) {
		provider := NewCoordinatesFileProvider(testFile+"_nocoord", logger)
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		_, err := provider.readFile(t.Context())
		if err == nil {
			t.Error("expected error, but didn't get one")
		}
		if !errors.Is(err, apperror.ErrNoValidCoordinates) {
			t.Errorf("expected error to be %s, got %s", apperror.ErrNoValidCoordinates, err)
		}
	})
	t.Run("parsing invalid coordinates fails", func(t *testing.T) {
		tests := []struct {
			name string
			file string
		}{
			{"latitude", testFile + "_brokenlat"},
			{"longitude", testFile + "_brokenlon"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				provider := NewCoordinatesFileProvider(tt.file, logger)
				if provider == nil {
					t.Fatal("expected provider to be non-nil")
				}
				_, err := provider.readFile(t.Context())
				if err == nil {
					t.Error("expected error, but didn't get one")
				}
				if !errors.Is(err, apperror.ErrNoValidCoordinates) {
					t.Errorf("expected error to be %s, got %s", apperror.ErrNoValidCoordinates, err)
				}
			})
		}
	})
}

func TestGeolocationFileProvider_LookupStream(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	t.Run("lookup stream succeeds", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			provider := NewCoordinatesFileProvider(testFile, logger)
			if provider == nil {
				t.Fatal("expected provider to be non-nil")
			}
			provider.ttl = time.Millisecond * 10
			provider.period = time.Millisecond * 10

			out := provider.LookupStream(ctx, "test")
			if out == nil {
				t.Fatal("expected stream to be non-nil")
			}

			var results []geobus.Result
			for len(results) < 1 {
				select {
				case r := <-out:
					results = append(results, r)
					cancel()
				default:
					synctest.Wait()
				}
			}

			synctest.Wait()
			if len(results) != 1 {
				t.Fatalf("expected at least one result, got %d", len(results))
			}
			result := results[0]
			if result.Key != "test" {
				t.Errorf("expected key to be %s, got %s", "test", result.Key)
			}
			if result.Coordinates.Latitude != testLat {
				t.Errorf("expected latitude to be %f, got %f", testLat, result.Coordinates.Latitude)
			}
			if result.Coordinates.Longitude != testLon {
				t.Errorf("expected longitude to be %f, got %f", testLon, result.Coordinates.Longitude)
			}
			if result.Coordinates.Accuracy.Float64() != types.AccuracyExact.Float64() {
				t.Errorf("expected accuracy to be %f, got %f", types.AccuracyExact.Float64(),
					result.Coordinates.Accuracy.Float64())
			}
			if result.Provider != provider.Name() {
				t.Errorf("expected source to be %s, got %s", provider.Name(), result.Provider)
			}
		})
	})
	t.Run("lookup stream fails during lookup", func(t *testing.T) {
		runCount := 0
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			provider := NewCoordinatesFileProvider(testFile, logger)
			if provider == nil {
				t.Fatal("expected provider to be non-nil")
			}
			provider.period = time.Millisecond * 10
			provider.locateFn = func(context.Context) (types.Coordinate, error) {
				if runCount == 0 {
					runCount++
					return types.Coordinate{}, errors.New("intentionally failing")
				}
				return types.Coordinate{Latitude: 1.0, Longitude: 2.0, Accuracy: types.AccuracyExact}, nil
			}

			out := provider.LookupStream(ctx, "test")
			if out == nil {
				t.Fatal("expected stream to be non-nil")
			}

			var result geobus.Result
			select {
			case r := <-out:
				result = r
				cancel()
			case <-ctx.Done():
				t.Fatalf("context done before result: %v", ctx.Err())
			}
			synctest.Wait()

			if result.Coordinates.Latitude != 1.0 {
				t.Errorf("expected latitude to be %f, got %f", 1.0, result.Coordinates.Latitude)
			}
			if result.Coordinates.Longitude != 2.0 {
				t.Errorf("expected longitude to be %f, got %f", 2.0, result.Coordinates.Longitude)
			}
			if result.Coordinates.Accuracy.Float64() != types.AccuracyExact.Float64() {
				t.Errorf("expected accuracy to be %f, got %f", types.AccuracyExact.Float64(),
					result.Coordinates.Accuracy)
			}
		})
	})
}
