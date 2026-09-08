// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"

	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/types"
)

const dbCoordinatePrecision = 2

func (s *Server) updateCurrentLocation(ctx context.Context, coords types.Coordinate, provider string) error {
	if !coords.Valid() {
		return fmt.Errorf("invalid coordinates: %f, %f", coords.Latitude, coords.Longitude)
	}

	location, err := s.addressByCoords(ctx, coords, provider)
	if err != nil {
		return fmt.Errorf("failed to get location by coordinates: %w", err)
	}
	if location.ID == 0 {
		return nil
	}

	params := model.UpdateCurrentAddressParams{
		AddressID: location.ID,
		CreatedAt: time.Now().UnixMicro(),
		UpdatedAt: time.Now().UnixMicro(),
	}
	if err = s.queries.UpdateCurrentAddress(ctx, params); err != nil {
		return fmt.Errorf("failed to set current address: %w", err)
	}

	return nil
}

func (s *Server) addressByCoords(ctx context.Context, coords types.Coordinate, provider string) (model.Address, error) {
	lat := types.TruncateFloat64(coords.Latitude, types.CoordinatePrecision)
	lon := types.TruncateFloat64(coords.Longitude, types.CoordinatePrecision)
	latTrunc := types.TruncateFloat64(lat, dbCoordinatePrecision)
	lonTrunc := types.TruncateFloat64(lon, dbCoordinatePrecision)
	address, err := s.queries.AddressByCoords(ctx, model.AddressByCoordsParams{
		LatTrunc: latTrunc,
		LonTrunc: lonTrunc,
		Locale:   s.t.Language().String(),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return address, fmt.Errorf("failed to retrieve location from database: %w", err)
	}

	if address.ID != 0 {
		return address, nil
	}

	lookup, aerr := s.geocoder.Reverse(ctx, coords)
	if aerr != nil {
		return address, fmt.Errorf("failed to reverse geocode coordinates: %w", aerr)
	}
	if !lookup.Found {
		return address, fmt.Errorf("no address found for coordinates: %f (%f), %f (%f)", coords.Latitude, lat,
			coords.Longitude, lon)
	}
	params := model.NewAddressParams{
		Latitude:     lat,
		Longitude:    lon,
		LatTrunc:     latTrunc,
		LonTrunc:     lonTrunc,
		Altitude:     sql.NullFloat64{Float64: coords.Altitude, Valid: true},
		Accuracy:     coords.Accuracy.Float64(),
		DisplayName:  lookup.DisplayName,
		Country:      sql.NullString{String: lookup.Country, Valid: true},
		State:        sql.NullString{String: lookup.State, Valid: true},
		Municipality: sql.NullString{String: lookup.Municipality, Valid: true},
		CityDistrict: sql.NullString{String: lookup.CityDistrict, Valid: true},
		Postcode:     sql.NullString{String: lookup.Postcode, Valid: true},
		City:         sql.NullString{String: lookup.City, Valid: true},
		Suburb:       sql.NullString{String: lookup.Suburb, Valid: true},
		Street:       sql.NullString{String: lookup.Street, Valid: true},
		HouseNumber:  sql.NullString{String: lookup.HouseNumber, Valid: true},
		Provider:     provider,
		Locale:       s.t.Language().String(),
		CreatedAt:    time.Now().UnixMicro(),
	}
	address, err = s.queries.NewAddress(ctx, params)
	if err != nil {
		if sqlErr, ok := errors.AsType[*sqlite.Error](err); ok && sqlErr.Code() == 2067 {
			return address, nil
		}
		return address, fmt.Errorf("failed to create new location: %w", err)
	}

	return address, nil
}
