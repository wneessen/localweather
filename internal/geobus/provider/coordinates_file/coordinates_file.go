// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package coordinates_file

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/types"
)

const (
	name     = "coordinates_file"
	ttlTime  = time.Hour * 12
	pollTime = time.Minute * 5
)

// CoordinatesFileProvider reads geolocation data from a file and emits updates via a stream.
// It periodically reads a specified file, parses its data, and updates geolocation results based on changes.
// Each result includes details about the location, accuracy, confidence, and timestamp of the data.
// Results are subject to a time-to-live (TTL) duration, ensuring outdated data is discarded.
type CoordinatesFileProvider struct {
	name     string
	path     string
	period   time.Duration
	ttl      time.Duration
	locateFn func() (lat, lon, alt float64, err error)
}

// NewLocationFileProvider initializes a CoordinatesFileProvider with a file path and default update
// interval and TTL settings.
func NewLocationFileProvider(path string) *CoordinatesFileProvider {
	provider := &CoordinatesFileProvider{
		name:   name,
		path:   path,
		period: pollTime,
		ttl:    ttlTime,
	}
	provider.locateFn = provider.readFile
	return provider
}

// Name returns the name of the CoordinatesFileProvider instance.
func (p *CoordinatesFileProvider) Name() string {
	return p.name
}

// LookupStream continuously streams geolocation results from a file, emitting updates when data changes
// or context ends.
func (p *CoordinatesFileProvider) LookupStream(ctx context.Context, key string) <-chan geobus.Result {
	out := make(chan geobus.Result)
	go func() {
		defer close(out)
		state := geobus.GeoLocationState{}
		firstRun := true

		for {
			if !firstRun {
				select {
				case <-ctx.Done():
					return
				case <-time.After(p.period):
				}
			}
			firstRun = false

			lat, lon, alt, err := p.locateFn()
			if err != nil {
				continue
			}
			coord := types.Coordinate{
				Altitude:  alt,
				Accuracy:  types.AccuracyExact,
				Latitude:  lat,
				Longitude: lon,
			}
			fmt.Printf("Coordinate: %+v\n", coord)
			state.Update(coord)
			r := p.createResult(key, coord)

			select {
			case <-ctx.Done():
				return
			case out <- r:
			}
		}
	}()
	return out
}

// createResult composes and returns a Result using provided geolocation data and metadata.
func (p *CoordinatesFileProvider) createResult(key string, coord types.Coordinate) geobus.Result {
	return geobus.Result{
		Key:       key,
		Altitude:  coord.Altitude,
		Accuracy:  coord.Accuracy,
		Latitude:  coord.Latitude,
		Longitude: coord.Longitude,
		Provider:  p.name,
		At:        time.Now(),
		TTL:       p.ttl,
	}
}

// readFile reads geolocation data from the file at the configured path.
// Returns latitude, longitude, altitude, accuracy, or an error if the file cannot be
// read or parsed correctly.
func (p *CoordinatesFileProvider) readFile() (lat, lon, alt float64, err error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read coordinates file %q: %w", p.path, err)
	}
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		coords := strings.Split(line, ",")
		if len(coords) < 2 || len(coords) > 3 {
			continue
		}
		lat, err = strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
		if err != nil {
			continue
		}
		lon, err = strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
		if err != nil {
			continue
		}
		if len(coords) == 3 {
			alt, err = strconv.ParseFloat(strings.TrimSpace(coords[2]), 64)
			if err != nil {
				continue
			}
		}

		return lat, lon, alt, nil
	}
	return 0, 0, 0, apperror.ErrNoValidCoordinates
}
