// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package cityname_file

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/testhelper"
	"github.com/wneessen/localweather/internal/types"
)

const (
	testFile = "../../../../testdata/cityname"
	testLat  = 40.7185
	testLon  = -74.0025
)

func TestNewCitynameFileProvider(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	t.Run("new cityname file provider succeeds", func(t *testing.T) {
		provider := testProvider(t, testFile)
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
	})
	t.Run("new cityname file provider without geocoder fails", func(t *testing.T) {
		provider, err := NewCitynameFileProvider(testFile, nil, logger)
		if err == nil {
			t.Fatal("expected provider to fail")
		}
		if provider != nil {
			t.Fatal("expected provider to be nil")
		}
	})
}

func TestCitynameFileProvider_Name(t *testing.T) {
	provider := testProvider(t, testFile)
	if provider == nil {
		t.Fatal("expected provider to be non-nil")
	}
	if !strings.EqualFold(provider.Name(), name) {
		t.Errorf("expected provider name to be %s, got %s", name, provider.Name())
	}
}

func TestNewCitynameFileProvider_readFile(t *testing.T) {
	t.Run("read file succeeds", func(t *testing.T) {
		provider := testProvider(t, testFile)
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
	t.Run("empty file fails", func(t *testing.T) {
		provider := testProvider(t, testFile+"_empty")
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		_, err := provider.readFile(t.Context())
		if err == nil {
			t.Error("expected error, but didn't get one")
		}
	})
	t.Run("geocoder lookup fails", func(t *testing.T) {
		provider := testProvider(t, testFile+"_fails")
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		_, err := provider.readFile(t.Context())
		if err == nil {
			t.Error("expected error, but didn't get one")
		}
	})
}

func TestCitynameFileProvider_LookupStream(t *testing.T) {
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	t.Run("lookup stream succeeds", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			coder := new(mockCoder)
			provider, err := NewCitynameFileProvider(testFile, coder, logger)
			if err != nil {
				t.Fatalf("failed to create cityname file provider: %s", err)
			}
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
			if result.Coordinates.Accuracy != types.AccuracyCity {
				t.Errorf("expected accuracy to be %f, got %f", types.AccuracyCity, result.Coordinates.Accuracy)
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

			logger = testhelper.NewLogger(t, config.LevelError, "discard")
			coder := new(mockCoder)
			provider, err := NewCitynameFileProvider(testFile, coder, logger)
			if err != nil {
				t.Fatalf("failed to create cityname file provider: %s", err)
			}
			if provider == nil {
				t.Fatal("expected provider to be non-nil")
			}
			provider.period = time.Millisecond * 10
			provider.locateFn = func(context.Context) (types.Coordinate, error) {
				if runCount == 0 {
					runCount++
					return types.Coordinate{}, errors.New("intentionally failing")
				}
				return types.Coordinate{Latitude: 1.0, Longitude: 2.0, Accuracy: types.AccuracyCity}, nil
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
			if result.Coordinates.Accuracy.Float64() != types.AccuracyCity.Float64() {
				t.Errorf("expected accuracy to be %f, got %f", types.AccuracyCity.Float64(),
					result.Coordinates.Accuracy.Float64())
			}
		})
	})
}

func testProvider(t *testing.T, file string) *Provider {
	t.Helper()
	logger := testhelper.NewLogger(t, config.LevelInfo, "")
	coder := new(mockCoder)
	provider, err := NewCitynameFileProvider(file, coder, logger)
	if err != nil {
		t.Fatalf("failed to create cityname file provider: %s", err)
	}
	return provider
}

type mockCoder struct{}

func (m *mockCoder) Name() string { return "mock" }
func (m *mockCoder) Reverse(_ context.Context, _ types.Coordinate) (types.Address, error) {
	return types.Address{}, errors.New("not implemented")
}

func (m *mockCoder) Search(_ context.Context, addr string) (types.Coordinate, error) {
	if addr == "Invalid, United Nations" {
		return types.Coordinate{}, errors.New("intentionally failing")
	}
	return types.Coordinate{Latitude: testLat, Longitude: testLon}, nil
}
