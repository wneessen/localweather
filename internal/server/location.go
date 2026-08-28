package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"modernc.org/sqlite"

	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/types"
)

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

	if err = s.queries.CurrentAddress(ctx, location.ID); err != nil {
		return fmt.Errorf("failed to set current address: %w", err)
	}
	s.log.Info("set current address", slog.Any("location", location))

	return nil
}

func (s *Server) addressByCoords(ctx context.Context, coords types.Coordinate, provider string) (model.Address, error) {
	lat := types.TruncateFloat64(coords.Latitude, types.CoordinatePrecision)
	lon := types.TruncateFloat64(coords.Longitude, types.CoordinatePrecision)
	address, err := s.queries.AddressByCoords(ctx, model.AddressByCoordsParams{
		Latitude:  lat,
		Longitude: lon,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
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
				Altitude:     coords.Altitude,
				Accuracy:     coords.Accuracy,
				DisplayName:  lookup.DisplayName,
				Country:      lookup.Country,
				State:        lookup.State,
				Municipality: lookup.Municipality,
				CityDistrict: lookup.CityDistrict,
				Postcode:     lookup.Postcode,
				City:         lookup.City,
				Suburb:       lookup.Suburb,
				Street:       lookup.Street,
				HouseNumber:  lookup.HouseNumber,
				Provider:     provider,
			}
			address, err = s.queries.NewAddress(ctx, params)
			if err != nil {
				if sqlErr, ok := errors.AsType[*sqlite.Error](err); ok && sqlErr.Code() == 2067 {
					return address, nil
				}
				return address, fmt.Errorf("failed to create new location: %w", err)
			}
			return address, nil
		default:
			return address, fmt.Errorf("failed to retrieve location from database: %w", err)
		}
	}

	return address, nil
}
