package lookupstream

import (
	"context"
	"log/slog"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/types"
)

type Params struct {
	OutChannel chan geobus.Result
	Provider   string
	Key        string
	Period     time.Duration
	TTL        time.Duration
	LocateFn   func(ctx context.Context) (types.Coordinate, error)
	Log        *log.Logger
}

func NewLookupStream(ctx context.Context, params Params) func() {
	return func() {
		defer close(params.OutChannel)
		state := geobus.GeoLocationState{}
		firstRun := true

		for {
			if !firstRun {
				select {
				case <-ctx.Done():
					return
				case <-time.After(params.Period):
				}
			}
			firstRun = false

			coords, err := params.LocateFn(ctx)
			if err != nil {
				params.Log.Error("geobus lookup failed", log.ErrAttr(err), slog.String("provider", params.Provider))
				continue
			}
			state.Update(coords)
			r := createResult(params.Provider, params.Key, params.TTL, coords)

			select {
			case <-ctx.Done():
				return
			case params.OutChannel <- r:
			}
		}
	}
}

func createResult(name, key string, ttl time.Duration, coords types.Coordinate) geobus.Result {
	return geobus.Result{
		Key:         key,
		Coordinates: coords,
		Provider:    name,
		At:          time.Now(),
		TTL:         ttl,
	}
}
