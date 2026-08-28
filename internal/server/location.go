package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/types"
)

func (s *Server) updateCurrentLocation(ctx context.Context, coords types.Coordinate) error {
	if !coords.Valid() {
		return fmt.Errorf("invalid coordinates: %f, %f", coords.Latitude, coords.Longitude)
	}

	location, err := s.addressByCoords(ctx, coords)
	if err != nil {
		return fmt.Errorf("failed to get location by coordinates: %w", err)
	}
	s.log.Info("found location", slog.Any("location", location))

	return nil
}

func (s *Server) addressByCoords(ctx context.Context, coords types.Coordinate) (model.Address, error) {
	address, err := s.queries.AddressByCoords(ctx, model.AddressByCoordsParams{
		Latitude:  coords.Latitude,
		Longitude: coords.Longitude,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			lookup, aerr := s.geocoder.Reverse(ctx, coords)
			if aerr != nil {
				return address, fmt.Errorf("failed to reverse geocode coordinates: %w", aerr)
			}
			if !lookup.Found {
				return address, fmt.Errorf("no address found for coordinates: %f, %f", coords.Latitude,
					coords.Longitude)
			}
			params := model.NewAddressParams{
				Latitude:     coords.Latitude,
				Longitude:    coords.Longitude,
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
			}
			address, err = s.queries.NewAddress(ctx, params)
			if err != nil {
				return address, fmt.Errorf("failed to create new location: %w", err)
			}
			return address, nil
		default:
			return address, fmt.Errorf("failed to retrieve location from database: %w", err)
		}
	}

	return address, nil
}
