// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package nominatim

import (
	"errors"
	stdhttp "net/http"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/language"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/geocode"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/testhelper"
	"github.com/wneessen/localweather/internal/types"
)

const (
	cityExpected             = "Quartier 205, 67, Friedrichstrasse, Friedrichstadt, Mitte, Berlin, 10117, Germany"
	cityFile                 = "../../../../testdata/nominatim_berlin.json"
	cityFileForward          = "../../../../testdata/nominatim_berlin_forward.json"
	emptyArray               = "../../../../testdata/empty_array.json"
	cityFileBrokenLat        = "../../../../testdata/nominatim_berlin_brokenlat.json"
	cityFileBrokenLon        = "../../../../testdata/nominatim_berlin_brokenlon.json"
	cityFileForwardBrokenLat = "../../../../testdata/nominatim_berlin_forward_brokenlat.json"
	cityFileForwardBrokenLon = "../../../../testdata/nominatim_berlin_forward_brokenlon.json"
	testHitTTL               = 1 * time.Second
	testMissTTL              = 1 * time.Second

	villageExpected = "Marshfield"
	villageFile     = "../../../../testdata/nominatim_marshfield.json"

	townExpected = "Otley"
	townFile     = "../../../../testdata/nominatim_otley.json"
)

var (
	cityCoords        = types.Coordinate{Latitude: 52.5129, Longitude: 13.3910}
	cityCoordsForward = types.Coordinate{Latitude: 52.5126051, Longitude: 13.3898616}
	villageCoords     = types.Coordinate{Latitude: 51.46292, Longitude: -2.31850}
	townCoords        = types.Coordinate{Latitude: 53.90712, Longitude: -1.69404}
)

func TestNew(t *testing.T) {
	t.Run("creating a new provider succeeds", func(t *testing.T) {
		coder := testCoder(t)
		if coder == nil {
			t.Fatal("expected a non-nil geocoder")
		}
	})
	t.Run("provider name is correct", func(t *testing.T) {
		coder := testCoder(t)
		if coder.Name() != name {
			t.Errorf("expected provider name to be %q, got %q", name, coder.Name())
		}
	})
}

func TestNominatim_Reverse(t *testing.T) {
	t.Run("reverse geocoding succeeds", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFile)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		addr, err := coder.Reverse(t.Context(), cityCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Found {
			t.Fatal("expected address to be found")
		}
		if !strings.EqualFold(addr.DisplayName, cityExpected) {
			t.Errorf("expected address to be %q, got %q", cityExpected, addr.DisplayName)
		}
	})
	t.Run("reverse cached geocoding succeeds", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFile)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := geocode.NewCachedGeocoder(testCoderWithRoundtripFunc(t, rtFn), testHitTTL, testMissTTL)
		addr, err := coder.Reverse(t.Context(), cityCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Found {
			t.Fatal("expected address to be found")
		}
		if !strings.EqualFold(addr.DisplayName, cityExpected) {
			t.Errorf("expected address to be %q, got %q", cityExpected, addr.DisplayName)
		}
		addr, err = coder.Reverse(t.Context(), cityCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.CacheHit {
			t.Error("expected cache hit")
		}
	})
	t.Run("reverse geocoding with town set should return the correct city", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(townFile)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		addr, err := coder.Reverse(t.Context(), townCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Found {
			t.Fatal("expected address to be found")
		}
		if !strings.EqualFold(addr.City, townExpected) {
			t.Errorf("expected city to be %q, got %q", townExpected, addr.DisplayName)
		}
	})
	t.Run("reverse geocoding with village set should return the correct city", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(villageFile)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		addr, err := coder.Reverse(t.Context(), villageCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Found {
			t.Fatal("expected address to be found")
		}
		if !strings.EqualFold(addr.City, villageExpected) {
			t.Errorf("expected city to be %q, got %q", villageExpected, addr.DisplayName)
		}
	})
	t.Run("reverse geocoding fails", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return nil, errors.New("intentionally failing")
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Reverse(t.Context(), cityCoords)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
	})
	t.Run("reverse geocoding fails on NaN latitude response", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFileBrokenLat)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Reverse(t.Context(), villageCoords)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
		if _, ok := errors.AsType[*apperror.CoordinateParsingError](err); !ok {
			t.Errorf("expected error to be %s, got: %s", &apperror.CoordinateParsingError{}, err)
		}
	})
	t.Run("reverse geocoding fails on NaN longitude response", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFileBrokenLon)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Reverse(t.Context(), villageCoords)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
		if _, ok := errors.AsType[*apperror.CoordinateParsingError](err); !ok {
			t.Errorf("expected error to be %s, got: %s", &apperror.CoordinateParsingError{}, err)
		}
	})
}

func TestNominatim_Search(t *testing.T) {
	t.Run("forward geocoding succeeds", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFileForward)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		coords, err := coder.Search(t.Context(), cityExpected)
		if err != nil {
			t.Fatal(err)
		}
		if !coords.Found {
			t.Fatal("expected address to be found")
		}
		if coords.Latitude != cityCoordsForward.Latitude {
			t.Errorf("expected latitude to be %f, got %f", cityCoordsForward.Latitude, coords.Latitude)
		}
		if coords.Longitude != cityCoordsForward.Longitude {
			t.Errorf("expected longitude to be %f, got %f", cityCoordsForward.Longitude, coords.Longitude)
		}
	})
	t.Run("forward geocoding fails", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return nil, errors.New("intentionally failing")
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Search(t.Context(), cityExpected)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
	})
	t.Run("forward geocoding fails on empty array response", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(emptyArray)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Search(t.Context(), cityExpected)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
	})
	t.Run("forward geocoding fails on NaN latitude response", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFileForwardBrokenLat)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Search(t.Context(), cityExpected)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
		if _, ok := errors.AsType[*apperror.CoordinateParsingError](err); !ok {
			t.Errorf("expected error to be %s, got: %s", &apperror.CoordinateParsingError{}, err)
		}
	})
	t.Run("forward geocoding fails on NaN longitue response", func(t *testing.T) {
		rtFn := func(req *stdhttp.Request) (*stdhttp.Response, error) {
			data, err := os.Open(cityFileForwardBrokenLon)
			if err != nil {
				t.Fatalf("failed to open JSON response file: %s", err)
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Body:       data,
				Header:     make(stdhttp.Header),
			}, nil
		}

		coder := testCoderWithRoundtripFunc(t, rtFn)
		_, err := coder.Search(t.Context(), cityExpected)
		if err == nil {
			t.Fatal("expected API request to fail")
		}
		if _, ok := errors.AsType[*apperror.CoordinateParsingError](err); !ok {
			t.Errorf("expected error to be %s, got: %s", &apperror.CoordinateParsingError{}, err)
		}
	})
}

func TestNominatim_integration(t *testing.T) {
	testhelper.PerformIntegrationTests(t)
	t.Run("reverse geocoding succeeds", func(t *testing.T) {
		coder := testCoder(t)
		addr, err := coder.Reverse(t.Context(), cityCoords)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Found {
			t.Fatal("expected address to be found")
		}
		want := "A.T. Kearney, 57, Charlottenstraße, Friedrichswerder, Friedrichstadt, Mitte, Berlin, 10117, Germany"
		if !strings.EqualFold(addr.DisplayName, want) {
			t.Errorf("expected address to be %q, got %q", want, addr.DisplayName)
		}
	})
	t.Run("geocoding forward-search succeeds", func(t *testing.T) {
		coder := testCoder(t)
		addr := "A.T. Kearney, 57, Charlottenstraße, Friedrichswerder, Friedrichstadt, Mitte, Berlin, 10117, Germany"
		coords, err := coder.Search(t.Context(), addr)
		if err != nil {
			t.Fatal(err)
		}
		want := types.Coordinate{Latitude: 52.5128745, Longitude: 13.3911058}
		if coords.Latitude != want.Latitude || coords.Longitude != want.Longitude {
			t.Errorf("expected coordinates to be %v, got %v", want, coords)
		}
	})
}

func testCoder(t *testing.T) geocode.Geocoder {
	logger := testhelper.NewLogger(t, config.LevelDebug, "")
	testHttpClient := http.New(logger)
	testLang := language.English
	return New(testHttpClient, testLang)
}

func testCoderWithRoundtripFunc(t *testing.T, fn func(req *stdhttp.Request) (*stdhttp.Response, error)) geocode.Geocoder {
	logger := testhelper.NewLogger(t, config.LevelDebug, "")
	testHttpClient := http.New(logger)
	testHttpClient.Transport = testhelper.MockRoundTripper{Fn: fn}
	testLang := language.English
	return New(testHttpClient, testLang)
}
