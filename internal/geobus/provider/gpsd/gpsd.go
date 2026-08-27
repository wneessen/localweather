// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package gpsd

import (
	"context"
	"time"

	"github.com/wneessen/localweather/internal/geobus/lookupstream"
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
	go lookupstream.NewLookupStream(ctx, p.name, key, out, p.ttl, p.period, p.locateFn)()
	return out
}
