// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geoip

import (
	"context"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/log"
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
	locateFn func(context.Context) (types.Coordinate, error)
	log      *log.Logger
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

func NewGeoIPProvider(http *http.Client, log *log.Logger) (*Provider, error) {
	if http == nil {
		return nil, apperror.ErrHTTPClientRequired
	}
	provider := &Provider{
		name:   name,
		http:   http,
		period: pollTime,
		ttl:    ttlTime,
		log:    log,
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

func (p *Provider) locate(ctx context.Context) (types.Coordinate, error) {
	coords := types.Coordinate{}
	ctxHttp, cancelHttp := context.WithTimeout(ctx, lookupTimeout)
	defer cancelHttp()

	result := new(APIResult)
	if _, err := p.http.Get(ctxHttp, apiEndpoint, result, nil, nil); err != nil {
		return coords, &apperror.GeoLocationAPIFetchError{Err: err}
	}

	coords.Accuracy = types.AccuracyUnknown
	if result.CountryCode != "" {
		coords.Accuracy = types.AccuracyCountry
	}
	if result.RegionCode != "" {
		coords.Accuracy = types.AccuracyRegion
	}
	if result.City != "" {
		coords.Accuracy = types.AccuracyCity
	}
	if result.ZipCode != "" {
		coords.Accuracy = types.AccuracyZip
	}

	coords.Latitude = types.TruncateFloat64(result.Latitude, types.CoordinatePrecision)
	coords.Longitude = types.TruncateFloat64(result.Longitude, types.CoordinatePrecision)

	return coords, nil
}
