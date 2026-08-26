package types

import "math"

const (
	EarthRadius       = 6371000.0 // meters
	DistanceThreshold = 2500.0    // 2.5km
	AccuracyThreshold = 50.0
)

// Coordinate represents a geographic coordinate.
type Coordinate struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
	Accuracy  Accuracy

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
