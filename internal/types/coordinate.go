// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package types

import "math"

const (
	AccuracyThreshold   = 50.0
	DistanceThreshold   = 2500.0    // 2.5km
	EarthRadius         = 6371000.0 // meters
	CoordinatePrecision = 6
)

// Coordinate represents a geographic coordinate.
type Coordinate struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
	Accuracy  Accuracy

	GPSMode  int
	CacheHit bool
	Found    bool
}

// PositionHasSignificantChange checks if the geographic position differs significantly from
// another based on the distance threshold. We are using the Haversine formula to calculate
// great-circle distance between two points on a sphere (in our case: Earth).
func (c Coordinate) PositionHasSignificantChange(other Coordinate) bool {
	// Higher accuracy always trumps the distance threshold.
	if c.Accuracy < other.Accuracy && math.Abs(c.Accuracy.Float64()-other.Accuracy.Float64()) > AccuracyThreshold {
		return true
	}

	dLat := (c.Latitude - other.Latitude) * math.Pi / 180
	dLon := (c.Longitude - other.Longitude) * math.Pi / 180
	lat1 := c.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	distance := 2 * EarthRadius * math.Asin(math.Sqrt(h))

	return distance > DistanceThreshold
}

// Valid checks if the coordinate is valid according to the EPSG logic
func (c Coordinate) Valid() bool {
	return c.Latitude >= -90 && c.Latitude <= 90 && c.Longitude >= -180 && c.Longitude <= 180
}

// Has2DFix reports whether the fix has at least a 2D fix.
func (c Coordinate) Has2DFix() bool {
	return c.GPSMode >= 2
}

// TruncateFloat64 truncates a floating-point number to a specified number of decimal places without rounding.
func TruncateFloat64[V Accuracy | float64](val V, precision int) V {
	p := math.Pow(10, float64(precision))
	return V(math.Trunc(float64(val)*p) / p)
}
