// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geoapi

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/types"
)

const (
	apiEndpoint  = "https://geoapi.info/api/geo"
	lookupTimeou = time.Second * 5
	name         = "geoapi"
	ttlTime      = time.Hour * 2
	pollTime     = time.Minute * 5
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
	IP       string `json:"ip"`
	Location struct {
		CountryCode string `json:"country,omitempty"`
		Country     string `json:"countryName,omitempty"`
		Region      string `json:"region,omitempty"`
		City        string `json:"city,omitempty"`
		ZipCode     string `json:"postalCode,omitempty"`
		TimeZone    string `json:"timezone"`
		Coordinates struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"coordinates"`
	} `json:"location"`
}

func NewGeoAPIProvider(http *http.Client, log *log.Logger) (*Provider, error) {
	if http == nil {
		return nil, fmt.Errorf("http client is required")
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
	ctxHttp, cancelHttp := context.WithTimeout(ctx, lookupTimeou)
	defer cancelHttp()

	result := new(APIResult)
	if _, err := p.http.Get(ctxHttp, apiEndpoint, result, nil, nil); err != nil {
		return coords, fmt.Errorf("failed to get geolocation data from API: %w", err)
	}

	coords.Accuracy = types.AccuracyUnknown
	if result.Location.CountryCode != "" {
		coords.Accuracy = types.AccuracyCountry
	}
	if result.Location.Region != "" {
		coords.Accuracy = types.AccuracyRegion
	}
	if result.Location.City != "" {
		coords.Accuracy = types.AccuracyCity
	}
	if result.Location.ZipCode != "" {
		coords.Accuracy = types.AccuracyZip
	}

	var err error
	coords.Latitude, err = strconv.ParseFloat(result.Location.Coordinates.Latitude, 64)
	if err != nil {
		return coords, fmt.Errorf("failed to parse latitude from API response: %w", err)
	}
	coords.Longitude, err = strconv.ParseFloat(result.Location.Coordinates.Longitude, 64)
	if err != nil {
		return coords, fmt.Errorf("failed to parse longitude from API response: %w", err)
	}

	return coords, nil
}
