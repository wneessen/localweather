// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package gpsd

import (
	"context"
	"time"

	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/gpsdpoll"
	"github.com/wneessen/localweather/internal/log"
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
	log      *log.Logger
}

func NewGPSdProvider(log *log.Logger) *Provider {
	provider := &Provider{
		name:   name,
		period: pollTime,
		ttl:    ttlTime,
		client: gpsdpoll.New(host, port),
		log:    log,
	}
	provider.locateFn = provider.client.Poll

	return provider
}

func (p *Provider) Name() string {
	return p.name
}

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
