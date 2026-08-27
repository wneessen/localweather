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

// Provider reads geolocation data from a file and emits updates via a stream.
// It periodically reads a specified file, parses its data, and updates geolocation results based on changes.
// Each result includes details about the location, accuracy, confidence, and timestamp of the data.
// Results are subject to a time-to-live (TTL) duration, ensuring outdated data is discarded.
type Provider struct {
	name     string
	path     string
	period   time.Duration
	ttl      time.Duration
	locateFn func() (types.Coordinate, error)
}

// NewCoordinatesFileProvider initializes a Provider with a file path and default update
// interval and TTL settings.
func NewCoordinatesFileProvider(path string) *Provider {
	provider := &Provider{
		name:   name,
		path:   path,
		period: pollTime,
		ttl:    ttlTime,
	}
	provider.locateFn = provider.readFile
	return provider
}

// Name returns the name of the Provider instance.
func (p *Provider) Name() string {
	return p.name
}

// LookupStream continuously streams geolocation results from a file, emitting updates when data changes
// or context ends.
func (p *Provider) LookupStream(ctx context.Context, key string) <-chan geobus.Result {
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

			coords, err := p.locateFn()
			if err != nil {
				continue
			}
			state.Update(coords)
			r := p.createResult(key, coords)

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
func (p *Provider) createResult(key string, coord types.Coordinate) geobus.Result {
	return geobus.Result{
		Key:         key,
		Coordinates: coord,
		Provider:    p.name,
		At:          time.Now(),
		TTL:         p.ttl,
	}
}

// readFile reads geolocation data from the file at the configured path.
// Returns latitude, longitude, altitude, accuracy, or an error if the file cannot be
// read or parsed correctly.
func (p *Provider) readFile() (types.Coordinate, error) {
	coords := types.Coordinate{}
	data, err := os.ReadFile(p.path)
	if err != nil {
		return coords, fmt.Errorf("failed to read coordinates file %q: %w", p.path, err)
	}
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		coordSplit := strings.Split(line, ",")
		if len(coordSplit) < 2 || len(coordSplit) > 3 {
			continue
		}
		coords.Latitude, err = strconv.ParseFloat(strings.TrimSpace(coordSplit[0]), 64)
		if err != nil {
			continue
		}
		coords.Longitude, err = strconv.ParseFloat(strings.TrimSpace(coordSplit[1]), 64)
		if err != nil {
			continue
		}
		if len(coordSplit) == 3 {
			coords.Altitude, err = strconv.ParseFloat(strings.TrimSpace(coordSplit[2]), 64)
			if err != nil {
				continue
			}
		}

		return coords, nil
	}
	return coords, apperror.ErrNoValidCoordinates
}
