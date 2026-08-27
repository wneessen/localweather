package server

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/text/language"

	"github.com/wneessen/localweather/internal/geocode"
	"github.com/wneessen/localweather/internal/geocode/provider/opencage"
	nominatim "github.com/wneessen/localweather/internal/geocode/provider/osm-nominatim"
	"github.com/wneessen/localweather/internal/http"
)

const (
	cacheHitTTL  = 1 * time.Hour
	cacheMissTTL = 10 * time.Minute
)

// initGeocoder initializes the geocoding provider and assigns it to the server, using the configuration and logger.
func (s *Server) initGeocoder() error {
	geocodeProvider, err := s.selectGeocodeProvider(language.English) // TODO: Integrate i18n
	if err != nil {
		return fmt.Errorf("failed to create geocode provider: %w", err)
	}
	s.geocoder = geocodeProvider
	s.log.Debug("geocoder initialized", slog.String("provider", s.geocoder.Name()))
	return nil
}

func (s *Server) selectGeocodeProvider(lang language.Tag) (geocode.Geocoder, error) {
	var geocoder geocode.Geocoder

	switch strings.ToLower(s.conf.Geocoder.Provider) {
	case "nominatim":
		geocoder = geocode.NewCachedGeocoder(nominatim.New(http.New(s.log), lang), cacheHitTTL, cacheMissTTL)
	case "opencage":
		if s.conf.Geocoder.APIKey == "" {
			return nil, fmt.Errorf("opencage geocoder requires an API key")
		}
		geocoder = geocode.NewCachedGeocoder(opencage.New(http.New(s.log), lang, s.conf.Geocoder.APIKey),
			cacheHitTTL, cacheMissTTL)
		/*
			case "geocode-earth":
				if conf.GeoCoder.APIKey == "" {
					return nil, fmt.Errorf("geocode-earth geocoder requires an API key")
				}
				geocoder = geocode.NewCachedGeocoder(geocodeearth.New(http.New(log), lang, conf.GeoCoder.APIKey),
					cacheHitTTL, cacheMissTTL)
		*/
	default:
		return nil, fmt.Errorf("unsupported geocoder type: %s", s.conf.Geocoder.Provider)
	}

	return geocoder, nil
}
