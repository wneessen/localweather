// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package ichnaea

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/ttlcache"
	"github.com/wneessen/localweather/internal/types"

	"github.com/mdlayher/wifi"
)

const (
	apiEndpoint     = "https://api.beacondb.net/v1/geolocate"
	lookupTimeout   = time.Second * 5
	wifiScanTime    = time.Second * 5
	wifiMaxPollTime = time.Minute * 10
	name            = "ichnaea"
	ttlTime         = time.Hour * 1
	pollTime        = time.Second * 30

	// wifiCacheTTL is the duration for which a WiFi-based (precise) geolocation result is
	// considered valid. It replaces the former periodic cache clearing.
	wifiCacheTTL = time.Hour * 12
	// fallbackCacheTTL is the duration for which an IP-based fallback result is considered
	// valid. Fallback results are stored as cache "misses" and therefore expire earlier.
	fallbackCacheTTL = time.Minute * 30
	// ipFallbackKey is the cache key that is used whenever no WiFi access points are known,
	// in which case the API can only perform an IP-based lookup.
	ipFallbackKey = "ip-fallback"
)

type Provider struct {
	name     string
	http     *http.Client
	wlan     *wifi.Client
	period   time.Duration
	ttl      time.Duration
	locateFn func(context.Context) (types.Coordinate, error)
	log      *log.Logger

	apLock sync.RWMutex
	aps    []WirelessNetwork
	apHash string

	cache *ttlcache.Cache[string, cachedLocation]
}

// cachedLocation holds a geolocation result together with the information whether it was
// derived from the WiFi access points or from the IP-based fallback of the API.
type cachedLocation struct {
	coords     types.Coordinate
	isFallback bool
}

type APIResult struct {
	Location struct {
		Altitude  float64 `json:"alt"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lng"`
	} `json:"location"`
	Accuracy   float64 `json:"accuracy"`
	IsFallback string  `json:"fallback"`
}

type WirelessNetwork struct {
	LastSeen       int64  `json:"age"`
	MACAddress     string `json:"macAddress"`
	SignalStrength int32  `json:"signalStrength"`
}

func NewICHNAEAProvider(http *http.Client, log *log.Logger) (*Provider, error) {
	if http == nil {
		return nil, fmt.Errorf("http client is required")
	}
	wlan, err := wifi.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wifi client: %w", err)
	}

	provider := &Provider{
		name:   name,
		http:   http,
		wlan:   wlan,
		period: pollTime,
		ttl:    ttlTime,
		apHash: ipFallbackKey,
		cache:  ttlcache.NewCache[string, cachedLocation](wifiCacheTTL, fallbackCacheTTL),
		log:    log,
	}
	provider.locateFn = provider.locate
	return provider, nil
}

func (p *Provider) Name() string {
	return p.name
}

// LookupStream continuously streams from the ICHNAEA API, emitting updates when data changes or context ends.
func (p *Provider) LookupStream(ctx context.Context, key string) <-chan geobus.Result {
	out := make(chan geobus.Result)
	go p.monitorWifiAccessPoints(ctx)
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

func (p *Provider) monitorWifiAccessPoints(ctx context.Context) {
	firstRun := true
	nextScanTime := time.Second * 2
	hasher := sha256.New()
	for {
		if !firstRun {
			select {
			case <-ctx.Done():
				return
			case <-time.After(nextScanTime):
			}
		}
		firstRun = false

		list, err := p.wifiAccessPoints(ctx)
		if err != nil {
			continue
		}

		// The hash is used as cache key, therefore it must be stable for an identical set
		// of access points. We hash the MAC addresses in lexicographical order, so that
		// fluctuating signal strengths do not change the key (and cause needless API calls).
		hash := ipFallbackKey
		if len(list) > 0 {
			macs := make([]string, 0, len(list))
			for _, ap := range list {
				macs = append(macs, ap.MACAddress)
			}
			slices.Sort(macs)
			for _, mac := range macs {
				hasher.Write([]byte(mac))
			}
			hash = fmt.Sprintf("%x", hasher.Sum(nil))
			hasher.Reset()
		}

		// The API prefers the strongest access points first.
		slices.SortFunc(list, func(a, b WirelessNetwork) int {
			return int(b.SignalStrength - a.SignalStrength)
		})

		p.apLock.Lock()
		p.apHash = hash
		p.aps = list
		p.apLock.Unlock()

		if len(list) == 0 {
			if nextScanTime < wifiMaxPollTime {
				nextScanTime = nextScanTime * 2
			}
			continue
		}
		nextScanTime = wifiMaxPollTime
	}
}

func (p *Provider) wifiAccessPoints(ctx context.Context) ([]WirelessNetwork, error) {
	var checkIfaces []*wifi.Interface
	var list []WirelessNetwork

	ifaces, err := p.wlan.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}
	for _, iface := range ifaces {
		if iface.Type != wifi.InterfaceTypeStation {
			continue
		}
		checkIfaces = append(checkIfaces, iface)
	}
	if len(checkIfaces) == 0 {
		return nil, nil
	}

	for _, iface := range checkIfaces {
		err = p.wlan.Scan(ctx, iface)
		if err != nil {
			continue
		}
		time.Sleep(wifiScanTime)

		aps, err := p.wlan.AccessPoints(iface)
		if err != nil {
			continue
		}
		for _, ap := range aps {
			if ap.SSID == "" || ap.SSID[0] == '\x00' || strings.HasSuffix(ap.SSID, "_nomap") {
				continue
			}
			list = append(list, WirelessNetwork{
				SignalStrength: ap.Signal / 100,
				MACAddress:     ap.BSSID.String(),
				LastSeen:       ap.LastSeen.Milliseconds(),
			})
		}
	}

	return list, nil
}

// locate returns the current coordinates, either from the TTL cache or, on a cache miss, from
// the ICHNAEA API. The cache key is the hash over the currently visible access points, so a
// changed environment automatically invalidates the previous result.
func (p *Provider) locate(ctx context.Context) (types.Coordinate, error) {
	p.apLock.RLock()
	wifiList := p.aps
	wifiHash := p.apHash
	p.apLock.RUnlock()

	cacheHit := false
	cached, err := p.cache.Fetch(ctx, wifiHash,
		func(ctx context.Context) (cachedLocation, error) {
			return p.lookup(ctx, wifiList)
		},
		func(location cachedLocation) bool {
			// IP-based fallback results are stored as "not found", so that they only live
			// for the shorter fallback TTL and a proper WiFi fix is retried earlier.
			return !location.isFallback
		},
		func(location *cachedLocation) {
			cacheHit = true
		},
	)
	if err != nil {
		return types.Coordinate{}, err
	}
	cached.coords.CacheHit = cacheHit
	return cached.coords, nil
}

// lookup performs the actual API request against the ICHNAEA endpoint.
func (p *Provider) lookup(ctx context.Context, wifiList []WirelessNetwork) (cachedLocation, error) {
	location := cachedLocation{}

	type request struct {
		ConsiderIP   bool              `json:"considerIp"`
		Accesspoints []WirelessNetwork `json:"wifiAccessPoints,omitempty"`
	}
	req := request{
		ConsiderIP:   true,
		Accesspoints: wifiList,
	}
	bodyBuffer := bytes.NewBuffer(nil)
	if err := json.MarshalWrite(bodyBuffer, req); err != nil {
		return location, fmt.Errorf("failed to encode wifi list to JSON: %w", err)
	}

	ctxHttp, cancelHttp := context.WithTimeout(ctx, lookupTimeout)
	defer cancelHttp()
	result := new(APIResult)
	if _, err := p.http.Post(ctxHttp, apiEndpoint, result, bodyBuffer,
		map[string]string{"Content-Type": "application/json"}); err != nil {
		return location, fmt.Errorf("failed to get geolocation data from API: %w", err)
	}

	location.coords.Accuracy = types.Accuracy(geobus.Truncate(result.Accuracy, geobus.TruncPrecision))
	location.coords.Altitude = geobus.Truncate(result.Location.Altitude, geobus.TruncPrecision)
	location.coords.Latitude = geobus.Truncate(result.Location.Latitude, geobus.TruncPrecision)
	location.coords.Longitude = geobus.Truncate(result.Location.Longitude, geobus.TruncPrecision)
	location.isFallback = result.IsFallback != ""

	return location, nil
}
