// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package coordinates_file

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/log"
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
	locateFn func(ctx context.Context) (types.Coordinate, error)
	log      *log.Logger
}

// NewCoordinatesFileProvider initializes a Provider with a file path and default update
// interval and TTL settings.
func NewCoordinatesFileProvider(path string, log *log.Logger) *Provider {
	provider := &Provider{
		name:   name,
		path:   path,
		period: pollTime,
		ttl:    ttlTime,
		log:    log,
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
	params := lookupstream.Params{
		Key:        key,
		LocateFn:   p.locateFn,
		Log:        p.log,
		OutChannel: out,
		Period:     p.period,
		Provider:   p.name,
		TTL:        p.ttl,
	}
	go lookupstream.NewLookupStream(ctx, params)()
	return out
}

// readFile reads geolocation data from the file at the configured path.
// Returns latitude, longitude, altitude, accuracy, or an error if the file cannot be
// read or parsed correctly.
func (p *Provider) readFile(_ context.Context) (types.Coordinate, error) {
	coords := types.Coordinate{}
	data, err := os.ReadFile(p.path)
	if err != nil {
		if _, ok := errors.AsType[*fs.PathError](err); ok {
			return coords, nil
		}
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
		coords.Accuracy = types.AccuracyManual

		return coords, nil
	}
	return coords, apperror.ErrNoValidCoordinates
}
