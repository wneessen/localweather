// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package ichnaea

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/lookupstream"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/types"

	"github.com/mdlayher/wifi"
)

const (
	apiEndpoint        = "https://api.beacondb.net/v1/geolocate"
	lookupTimeout      = time.Second * 5
	wifiScanTime       = time.Second * 5
	wifiMaxPollTime    = time.Minute * 10
	wifiClearCacheTime = time.Hour * 12
	name               = "ichnaea"
	ttlTime            = time.Hour * 1
	pollTime           = time.Second * 30
	fallbackCacheTime  = time.Minute * 30
)

type Provider struct {
	name     string
	http     *http.Client
	wlan     *wifi.Client
	period   time.Duration
	ttl      time.Duration
	locateFn func(context.Context) (types.Coordinate, error)

	apLock    sync.RWMutex
	aps       []WirelessNetwork
	apHash    string
	ipfLock   sync.RWMutex
	ipfcache  *ipFallbackCache
	wifiLock  sync.RWMutex
	wifiCache map[string]types.Coordinate
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

type ipFallbackCache struct {
	expires time.Time
	coords  types.Coordinate
}

func NewICHNAEAProvider(http *http.Client) (*Provider, error) {
	if http == nil {
		return nil, fmt.Errorf("http client is required")
	}
	wlan, err := wifi.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wifi client: %w", err)
	}

	provider := &Provider{
		name:      name,
		http:      http,
		wlan:      wlan,
		period:    pollTime,
		ttl:       ttlTime,
		ipfcache:  &ipFallbackCache{},
		wifiCache: make(map[string]types.Coordinate),
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
	go p.clearWifiCache(ctx)
	go lookupstream.NewLookupStream(ctx, p.name, key, out, p.ttl, p.period, p.locateFn)()
	return out
}

func (p *Provider) clearWifiCache(ctx context.Context) {
	ticker := time.NewTicker(wifiClearCacheTime)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.wifiLock.Lock()
			p.wifiCache = make(map[string]types.Coordinate)
			p.wifiLock.Unlock()
		}
	}
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
		slices.SortFunc(list, func(a, b WirelessNetwork) int {
			return int(b.SignalStrength - a.SignalStrength)
		})
		for _, ap := range list {
			hasher.Write([]byte(ap.MACAddress))
		}
		p.apLock.Lock()
		p.apHash = fmt.Sprintf("%x", hasher.Sum(nil))
		p.aps = list
		p.apLock.Unlock()
		hasher.Reset()

		if len(list) == 0 {
			if nextScanTime < wifiMaxPollTime {
				nextScanTime = nextScanTime * 2
			}
			continue
		}
		p.ipfLock.Lock()
		p.ipfcache.expires = time.Time{}
		p.ipfLock.Unlock()
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

func (p *Provider) locate(ctx context.Context) (types.Coordinate, error) {
	coords := types.Coordinate{}
	p.apLock.RLock()
	wifiList := p.aps
	wifiHash := p.apHash
	p.apLock.RUnlock()

	// If WiFi cache is valid, return cached coordinates
	p.wifiLock.RLock()
	if cache, ok := p.wifiCache[wifiHash]; ok {
		p.wifiLock.RUnlock()
		return cache, nil
	}
	p.wifiLock.RUnlock()

	// If IP fallback cache is still valid, return cached coordinates
	p.ipfLock.RLock()
	if p.ipfcache.expires.After(time.Now()) {
		p.ipfLock.RUnlock()
		return p.ipfcache.coords, nil
	}
	p.ipfLock.RUnlock()

	type request struct {
		ConsiderIP   bool              `json:"considerIp"`
		Accesspoints []WirelessNetwork `json:"wifiAccessPoints,omitempty"`
	}
	req := request{
		ConsiderIP:   true,
		Accesspoints: wifiList,
	}
	bodyBuffer := bytes.NewBuffer(nil)
	if err := json.NewEncoder(bodyBuffer).Encode(req); err != nil {
		return coords, fmt.Errorf("failed to encode wifi list to JSON: %w", err)
	}

	ctxHttp, cancelHttp := context.WithTimeout(ctx, lookupTimeout)
	defer cancelHttp()
	result := new(APIResult)
	if _, err := p.http.Post(ctxHttp, apiEndpoint, result, bodyBuffer,
		map[string]string{"Content-Type": "application/json"}); err != nil {
		return coords, fmt.Errorf("failed to get geolocation data from API: %w", err)
	}

	coords.Accuracy = types.Accuracy(geobus.Truncate(result.Accuracy, geobus.TruncPrecision))
	coords.Altitude = geobus.Truncate(result.Location.Altitude, geobus.TruncPrecision)
	coords.Latitude = geobus.Truncate(result.Location.Latitude, geobus.TruncPrecision)
	coords.Longitude = geobus.Truncate(result.Location.Longitude, geobus.TruncPrecision)

	if result.IsFallback != "" {
		p.ipfLock.Lock()
		p.ipfcache.expires = time.Now().Add(fallbackCacheTime)
		p.ipfcache.coords = coords
		p.ipfLock.Unlock()
		return coords, nil
	}

	p.wifiLock.Lock()
	p.wifiCache[wifiHash] = coords
	p.wifiLock.Unlock()
	return coords, nil
}
