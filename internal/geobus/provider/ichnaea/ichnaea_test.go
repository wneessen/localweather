// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package ichnaea

import (
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/mdlayher/wifi"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/testhelper"
	"github.com/wneessen/localweather/internal/types"
)

const (
	testFile         = "../../../../testdata/beacondb.json"
	testLat          = 40.7185
	testLon          = -74.0025
	testAcc  float64 = 2000
)

func TestNewICHNAEAProvider(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	t.Run("new ICHNAEA provider succeeds", func(t *testing.T) {
		provider, err := NewICHNAEAProvider(http.New(logger), logger)
		if err != nil {
			t.Fatalf("failed to create ICHNAEA provider: %s", err)
		}
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
	})
	t.Run("ICHNAEA without http client fails ", func(t *testing.T) {
		provider, err := NewICHNAEAProvider(nil, logger)
		if err == nil {
			t.Fatal("expected provider to fail")
		}
		if provider != nil {
			t.Fatal("expected provider to be nil")
		}
	})
}

func TestICHNAEAProvider_Name(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	provider, err := NewICHNAEAProvider(http.New(logger), logger)
	if err != nil {
		t.Fatalf("failed to create ICHNAEA provider: %s", err)
	}
	if !strings.EqualFold(provider.Name(), name) {
		t.Errorf("expected provider name to be %s, got %s", name, provider.Name())
	}
}

// This test is very flacky, since it depends on the WiFi hardware
func TestNewICHNAEAProvider_wifiList(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	provider, err := NewICHNAEAProvider(http.New(logger), logger)
	if err != nil {
		t.Fatalf("failed to create ICHNAEA provider: %s", err)
	}
	list, err := provider.wifiAccessPoints(t.Context())
	if err != nil {
		t.Fatalf("failed to get WiFi list: %s", err)
	}
	if len(list) == 0 {
		t.Skip("no WiFi access points found, test results are meaningless")
	}
}

func TestICHNAEAProvider_locate(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	t.Run("locate succeeds with different accuracies", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(testFile)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}
		client := http.New(logger)
		client.Transport = testhelper.MockRoundTripper{Fn: rtFn}
		provider, err := NewICHNAEAProvider(client, logger)
		if err != nil {
			t.Fatalf("failed to create ICHNAEA provider: %s", err)
		}

		coords, err := provider.locate(t.Context())
		if err != nil {
			t.Fatalf("failed to locate coordinates via ICHNAEA: %s", err)
		}
		if coords.Latitude != testLat {
			t.Errorf("expected latitude to be %f, got %f", testLat, coords.Latitude)
		}
		if coords.Longitude != testLon {
			t.Errorf("expected longitude to be %f, got %f", testLon, coords.Longitude)
		}
		if types.TruncateFloat64(coords.Accuracy.Float64(), 1) != types.TruncateFloat64(testAcc, 1) {
			t.Errorf("expected accuracy to be %f, got %f", types.TruncateFloat64(testAcc, 1),
				types.TruncateFloat64(coords.Accuracy.Float64(), 1))
		}
	})
	t.Run("locate fails with broken JSON", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return &stdhttp.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("NOT_JSON")),
				Header:     make(stdhttp.Header),
			}, nil
		}
		client := http.New(logger)
		client.Transport = testhelper.MockRoundTripper{Fn: rtFn}
		provider, err := NewICHNAEAProvider(client, logger)
		if err != nil {
			t.Fatalf("failed to create ICHNAEA provider: %s", err)
		}

		_, err = provider.locate(t.Context())
		if err == nil {
			t.Fatal("expected locate to fail")
		}
	})
}

func TestICHNAEAProvider_LookupStream(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	t.Run("lookup stream succeeds", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
				data, err := os.Open(testFile)
				if err != nil {
					t.Fatalf("failed to open JSON response file: %s", err)
				}

				return &stdhttp.Response{
					StatusCode: 200,
					Body:       data,
					Header:     make(stdhttp.Header),
				}, nil
			}
			client := http.New(logger)
			client.Transport = testhelper.MockRoundTripper{Fn: rtFn}
			provider, err := NewICHNAEAProvider(client, logger)
			if err != nil {
				t.Fatalf("failed to create GeoIP provider: %s", err)
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
					// Block until all goroutines are durably blocked, then advance
					// fake time to the next wakeup (e.g. time.After/ Sleep).
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
			wantAcc := 2000.0
			if result.Coordinates.Accuracy.Float64() != wantAcc {
				t.Errorf("expected accuracy to be %f, got %f", wantAcc, result.Coordinates.Accuracy.Float64())
			}
			if result.Provider != provider.Name() {
				t.Errorf("expected provider to be %s, got %s", provider.Name(), result.Provider)
			}
		})
	})
	t.Run("lookup stream fails during lookup", func(t *testing.T) {
		runCount := 0
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			conf := new(config.Config)
			conf.Log.Output = "discard"
			discarder := log.New(conf)
			provider, err := NewICHNAEAProvider(http.New(discarder), discarder)
			if err != nil {
				t.Fatalf("failed to create GeoIP provider: %s", err)
			}
			provider.period = time.Millisecond * 10
			provider.locateFn = func(ctx context.Context) (types.Coordinate, error) {
				if runCount == 0 {
					runCount++
					return types.Coordinate{}, errors.New("intentionally failing")
				}
				return types.Coordinate{Latitude: 1.0, Longitude: 2.0, Accuracy: 3.0}, nil
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
			if result.Coordinates.Accuracy != 3.0 {
				t.Errorf("expected accuracy to be %f, got %f", 3.0, result.Coordinates.Accuracy)
			}
		})
	})
}

func TestNewICHNAEAProvider_monitorWifiAccessPoints(t *testing.T) {
	testRequiresWiFi(t)
	logger := log.New(&config.Config{})
	t.Run("monitor WiFi access points succeeds", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			isCancelled := false
			context.AfterFunc(ctx, func() {
				isCancelled = true
			})

			provider, err := NewICHNAEAProvider(http.New(logger), logger)
			if err != nil {
				t.Fatalf("failed to create ICHNAEA provider: %s", err)
			}
			go provider.monitorWifiAccessPoints(ctx)
			synctest.Wait()
			cancel()
			synctest.Wait()
			if !isCancelled {
				t.Fatal("expected monitor to be cancelled")
			}
		})
	})
}

func testRequiresWiFi(t *testing.T) {
	wlan, err := wifi.New()
	if err != nil {
		t.Skip("system has no WiFi support, skipping WiFi related tests")
	}

	checkIfaces := make([]*wifi.Interface, 0)
	ifaces, err := wlan.Interfaces()
	if err != nil {
		t.Skip("no WiFi interfaces found, skipping WiFi related tests")
	}
	for _, iface := range ifaces {
		if iface.Type != wifi.InterfaceTypeStation {
			continue
		}
		checkIfaces = append(checkIfaces, iface)
	}
	if len(checkIfaces) == 0 {
		t.Skip("no WiFi interfaces found, skipping WiFi related tests")
	}
}
