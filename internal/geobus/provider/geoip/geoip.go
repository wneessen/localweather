// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geoip

import (
	"context"
	"fmt"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/types"
)

const (
	apiEndpoint   = "https://reallyfreegeoip.org/json/"
	lookupTimeout = time.Second * 10
	name          = "geoip"
	ttlTime       = time.Hour * 2
	pollTime      = time.Minute * 15
)

type Provider struct {
	name     string
	http     *http.Client
	period   time.Duration
	ttl      time.Duration
	locateFn func(ctx context.Context) (lat, lon float64, acc types.Accuracy, err error)
}

type APIResult struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	Country     string  `json:"country_name"`
	RegionCode  string  `json:"region_code,omitempty"`
	Region      string  `json:"region_name,omitempty"`
	City        string  `json:"city,omitempty"`
	ZipCode     string  `json:"zip_code,omitempty"`
	TimeZone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

func NewGeoIPProvider(http *http.Client) (*Provider, error) {
	if http == nil {
		return nil, fmt.Errorf("http client is required")
	}
	provider := &Provider{
		name:   name,
		http:   http,
		period: pollTime,
		ttl:    ttlTime,
	}
	provider.locateFn = provider.locate
	return provider, nil
}

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

			lat, lon, acc, err := p.locateFn(ctx)
			if err != nil {
				continue
			}
			coord := types.Coordinate{
				Accuracy:  acc,
				Latitude:  lat,
				Longitude: lon,
			}
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
func (p *Provider) createResult(key string, coord types.Coordinate) geobus.Result {
	return geobus.Result{
		Key:       key,
		Altitude:  0, // GeoIP does not provide altitude information
		Accuracy:  coord.Accuracy,
		Latitude:  coord.Latitude,
		Longitude: coord.Longitude,
		Provider:  p.name,
		At:        time.Now(),
		TTL:       p.ttl,
	}
}

func (p *Provider) locate(ctx context.Context) (lat, lon float64, acc types.Accuracy, err error) {
	ctxHttp, cancelHttp := context.WithTimeout(ctx, lookupTimeout)
	defer cancelHttp()

	result := new(APIResult)
	if _, err = p.http.Get(ctxHttp, apiEndpoint, result, nil, nil); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get geolocation data from API: %w", err)
	}

	acc = types.AccuracyUnknown
	if result.CountryCode != "" {
		acc = types.AccuracyCountry
	}
	if result.RegionCode != "" {
		acc = types.AccuracyRegion
	}
	if result.City != "" {
		acc = types.AccuracyCity
	}
	if result.ZipCode != "" {
		acc = types.AccuracyZip
	}

	return geobus.Truncate(result.Latitude, geobus.TruncPrecision),
		geobus.Truncate(result.Longitude, geobus.TruncPrecision),
		types.Accuracy(geobus.Truncate(acc.Float64(), geobus.TruncPrecision)),
		nil
}
