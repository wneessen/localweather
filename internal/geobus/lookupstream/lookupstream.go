package lookupstream

import (
	"context"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/types"
)

func NewLookupStream(ctx context.Context, name, key string, out chan geobus.Result, ttl, period time.Duration,
	locateFn func(ctx context.Context) (types.Coordinate, error),
) func() {
	lookupstream := func() {
		defer close(out)
		state := geobus.GeoLocationState{}
		firstRun := true

		for {
			if !firstRun {
				select {
				case <-ctx.Done():
					return
				case <-time.After(period):
				}
			}
			firstRun = false

			coords, err := locateFn(ctx)
			if err != nil {
				continue
			}
			state.Update(coords)
			r := createResult(name, key, ttl, coords)

			select {
			case <-ctx.Done():
				return
			case out <- r:
			}
		}
	}
	return lookupstream
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
