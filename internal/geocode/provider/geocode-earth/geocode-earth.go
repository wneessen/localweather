// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geocodeearth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"golang.org/x/text/language"

	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/types"
)

const (
	reverseAPIEndpoint = "https://api.geocode.earth/v1/reverse"
	searchAPIEndpoint  = "https://api.geocode.earth/v1/search"
	APITimeout         = time.Second * 10
	name               = "geocode-earth"
)

type GeocodeEarth struct {
	apikey string
	http   *http.Client
	lang   language.Tag
}

type ReverseResponse struct {
	Features []ReverseFeature `json:"features"`
	Type     string           `json:"type"`
}

type SearchResponse struct {
	Features []SearchFeature `json:"features"`
	Type     string          `json:"type"`
}

type ReverseFeature struct {
	Properties Properties `json:"properties"`
	Type       string     `json:"type"`
}

type SearchFeature struct {
	Geometry Geometry `json:"geometry"`
	Type     string   `json:"type"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Properties struct {
	DisplayName    string `json:"label"`
	City           string `json:"locality"`
	CityDistrict   string `json:"county"`
	Continent      string `json:"continent"`
	Country        string `json:"country"`
	CountryCode    string `json:"country_code"`
	HouseNumber    string `json:"housenumber"`
	PoliticalUnion string `json:"political_union"`
	Municipality   string `json:"neighbourhood"`
	Postcode       string `json:"postalcode"`
	Road           string `json:"street"`
	State          string `json:"region"`
	StateCode      string `json:"region_a"`
}

func New(client *http.Client, lang language.Tag, apikey string) *GeocodeEarth {
	return &GeocodeEarth{
		apikey: apikey,
		lang:   lang,
		http:   client,
	}
}

func (g *GeocodeEarth) Name() string {
	return name
}

func (g *GeocodeEarth) Reverse(ctx context.Context, coords types.Coordinate) (types.Address, error) {
	var response ReverseResponse
	address := types.Address{}

	query := url.Values{}
	query.Set("api_key", g.apikey)
	query.Set("point.lat", fmt.Sprintf("%f", coords.Latitude))
	query.Set("point.lon", fmt.Sprintf("%f", coords.Longitude))
	query.Set("lang", g.lang.String())

	code, err := g.http.GetWithTimeout(ctx, reverseAPIEndpoint, &response, query, nil, APITimeout)
	if err != nil {
		return address, fmt.Errorf("failed to retrieve address details from geocode.earth API: %w", err)
	}
	if code != 200 {
		return address, fmt.Errorf("received non-positive response code from geocode.earth API: %d", code)
	}
	if len(response.Features) < 1 {
		return address, fmt.Errorf("no address found for coordinates")
	}

	// Fill the geocode.Address struct
	result := response.Features[0].Properties
	address = types.Address{
		Found:        true,
		Latitude:     coords.Latitude,
		Longitude:    coords.Longitude,
		DisplayName:  result.DisplayName,
		Country:      result.Country,
		State:        result.State,
		Municipality: result.Municipality,
		CityDistrict: result.CityDistrict,
		Postcode:     result.Postcode,
		City:         result.City,
		Street:       result.Road,
		HouseNumber:  result.HouseNumber,
	}

	return address, nil
}

func (g *GeocodeEarth) Search(ctx context.Context, address string) (types.Coordinate, error) {
	var response SearchResponse
	coords := types.Coordinate{}

	query := url.Values{}
	query.Set("api_key", g.apikey)
	query.Set("text", address)
	query.Set("lang", g.lang.String())

	code, err := g.http.GetWithTimeout(ctx, searchAPIEndpoint, &response, query, nil, APITimeout)
	if err != nil {
		return coords, fmt.Errorf("failed to retrieve address details from geocode.earth API: %w", err)
	}
	if code != 200 {
		return coords, fmt.Errorf("received non-positive response code from geocode.earth API: %d", code)
	}
	if len(response.Features) < 1 {
		return coords, fmt.Errorf("no coordinates found for address %q", address)
	}

	// Fill the geocode.Address struct
	result := response.Features[0].Geometry
	if len(result.Coordinates) != 2 {
		return coords, fmt.Errorf("unexpected 2 coordinates in response, got: %d",
			len(result.Coordinates))
	}
	coords = types.Coordinate{
		Latitude:  result.Coordinates[1],
		Longitude: result.Coordinates[0],
		Found:     true,
	}

	return coords, nil
}
