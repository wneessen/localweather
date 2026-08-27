// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package gpsd

import (
	"context"
	"time"

	"github.com/wneessen/localweather/internal/gpsdpoll"
	"github.com/wneessen/localweather/internal/types"

	"github.com/wneessen/localweather/internal/geobus"
)

const (
	host     = "localhost"
	port     = "2947"
	name     = "gpsd"
	ttlTime  = time.Minute * 30
	pollTime = time.Second * 30
)

type Provider struct {
	name     string
	period   time.Duration
	ttl      time.Duration
	client   *gpsdpoll.Client
	locateFn func(ctx context.Context) (types.Coordinate, error)
}

func NewGPSdProvider() *Provider {
	provider := &Provider{
		name:   name,
		period: pollTime,
		ttl:    ttlTime,
		client: gpsdpoll.New(host, port),
	}
	provider.locateFn = provider.client.Poll

	return provider
}

func (p *Provider) Name() string {
	return p.name
}

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

			coord, err := p.locateFn(ctx)
			if err != nil {
				continue
			}
			if !coord.Has2DFix() {
				continue
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
		Key:         key,
		Coordinates: coord,
		Provider:    p.name,
		At:          time.Now(),
		TTL:         p.ttl,
	}
}
