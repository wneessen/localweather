// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geocode

import (
	"context"
	"math"
	"time"

	"github.com/wneessen/localweather/internal/ttlcache"
	"github.com/wneessen/localweather/internal/types"
)

// coordPrecision is the precision used to quantize coordinates (0.01 degrees ≈ 1.1 km)
const coordPrecision = 1e-2

type reverseKey struct {
	Provider string
	LatQ     int32
	LonQ     int32
}

type CachedGeocoder struct {
	coder   Geocoder
	reverse *ttlcache.Cache[reverseKey, types.Address]
	search  *ttlcache.Cache[string, types.Coordinate]
}

func NewCachedGeocoder(coder Geocoder, ttlHit, ttlMiss time.Duration) *CachedGeocoder {
	return &CachedGeocoder{
		coder:   coder,
		reverse: ttlcache.NewCache[reverseKey, types.Address](ttlHit, ttlMiss),
		search:  ttlcache.NewCache[string, types.Coordinate](ttlHit, ttlMiss),
	}
}

func (c *CachedGeocoder) Name() string {
	return "geocoder cache using " + c.coder.Name()
}

func (c *CachedGeocoder) Reverse(ctx context.Context, coords types.Coordinate) (types.Address, error) {
	key := reverseKey{
		Provider: c.coder.Name(),
		LatQ:     quantizeCoord(coords.Latitude),
		LonQ:     quantizeCoord(coords.Longitude),
	}
	return c.reverse.Fetch(ctx,
		key,
		func(ctx context.Context) (types.Address, error) { return c.coder.Reverse(ctx, coords) },
		func(a types.Address) bool { return a.AddressFound },
		func(a *types.Address) { a.CacheHit = true },
	)
}

func (c *CachedGeocoder) Search(ctx context.Context, query string) (types.Coordinate, error) {
	return c.search.Fetch(ctx, query,
		func(ctx context.Context) (types.Coordinate, error) { return c.coder.Search(ctx, query) },
		func(co types.Coordinate) bool { return co.Found },
		func(co *types.Coordinate) { co.CacheHit = true },
	)
}

func quantizeCoord(val float64) int32 {
	return int32(math.Round(val / coordPrecision))
}
