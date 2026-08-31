package types

import "fmt"

type Accuracy float64

const (
	AccuracyCountry Accuracy = 300000
	AccuracyRegion  Accuracy = 100000
	AccuracyCity    Accuracy = 15000
	AccuracyZip     Accuracy = 3000
	AccuracyStreet  Accuracy = 500
	AccuracyExact   Accuracy = 5
	AccuracyManual  Accuracy = 1
	AccuracyUnknown Accuracy = 1000000
)

func (a Accuracy) Float64() float64 {
	return float64(a)
}

func (a Accuracy) String() string {
	switch a {
	case AccuracyExact:
		return "Exact (<=5m)"
	case AccuracyZip:
		return "ZIP area (~3km)"
	case AccuracyCity:
		return "City (~15km)"
	case AccuracyRegion:
		return "Region (~100km)"
	case AccuracyCountry:
		return "Country (~300km)"
	case AccuracyUnknown:
		return "Unknown (>1000km)"
	default:
		return fmt.Sprintf("%.2fm", a)
	}
}
