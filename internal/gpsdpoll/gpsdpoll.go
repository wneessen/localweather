// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package gpsdpoll

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/wneessen/localweather/internal/types"
)

const (
	fallbackAccuracy3DFix = 10  // ~10 m typical consumer GPS in open sky
	fallbackAccuracy2DFix = 25  // worse than 3D, but still accurate enough
	fallbackAccuracyNoFix = 1e6 // effectively unusable
	watchTimeout          = time.Second * 2
)

// Client is a minimal GPSd client
type Client struct {
	Addr string
}

// gpsdPollResponse matches the subset of gpsd's POLL response we care about.
type gpsdPollResponse struct {
	Class     string   `json:"class"`
	Latitude  *float64 `json:"lat"`
	Longitude *float64 `json:"lon"`
	Accuracy  float64
	Altitude  *float64 `json:"altMSL"`
	Mode      int      `json:"mode"`
	EPX       *float64 `json:"epx"`
	EPY       *float64 `json:"epy"`
	EPH       *float64 `json:"eph"`
	EPV       *float64 `json:"epv"`
}

// New constructs a new Client for the given host and port.
func New(host, port string) *Client {
	return &Client{
		Addr: net.JoinHostPort(host, port),
	}
}

// Poll connects to gpsd, sends a POLL request, and returns the first TPV
// entry from the POLL response. The connection is closed before returning.
func (c *Client) Poll(ctx context.Context) (types.Coordinate, error) {
	var zero types.Coordinate

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return zero, fmt.Errorf("failed to connect to GPSd: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Respect context deadline if present, otherwise we add a safety net so we don't hang
	// forever if ctx has no deadline.
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(watchTimeout))
	}

	// Request a WATCH.
	if _, err = fmt.Fprint(conn, `?WATCH={"enable":true,"json":true}`+"\n"); err != nil {
		return zero, fmt.Errorf("gpspoll: write POLL: %w", err)
	}

	// Wait for a TPV response or timeout.
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var resp gpsdPollResponse

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if err = json.Unmarshal(line, &resp); err != nil {
			return zero, fmt.Errorf("failed to unmarshal JSON from GPSd: %w", err)
		}
		if resp.Class != "TPV" {
			continue
		}
		if resp.Longitude == nil || resp.Latitude == nil || resp.Altitude == nil {
			continue
		}

		return types.Coordinate{
			Accuracy:  types.Accuracy(horizontalAccuracyMeters(resp)),
			Altitude:  *resp.Altitude,
			Latitude:  *resp.Latitude,
			Longitude: *resp.Longitude,
			GPSMode:   resp.Mode,
		}, nil
	}

	if err = scanner.Err(); err != nil {
		return zero, fmt.Errorf("failed to scan GPSd response: %w", err)
	}

	return zero, fmt.Errorf("no TPV response received from GPSd")
}

func horizontalAccuracyMeters(resp gpsdPollResponse) float64 {
	if resp.EPH == nil || resp.EPX == nil || resp.EPY == nil {
		return horizontalAccuracyFallback(resp)
	}
	switch {
	case *resp.EPH > 0:
		return *resp.EPH
	case *resp.EPX > 0 && *resp.EPY > 0:
		return math.Hypot(*resp.EPX, *resp.EPY)
	default:
		return horizontalAccuracyFallback(resp)
	}
}

func horizontalAccuracyFallback(resp gpsdPollResponse) float64 {
	switch resp.Mode {
	case 3:
		return fallbackAccuracy3DFix
	case 2:
		return fallbackAccuracy2DFix
	default:
		return fallbackAccuracyNoFix
	}
}
