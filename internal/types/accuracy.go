package types

type Accuracy float64

const (
	AccuracyCountry Accuracy = 300000
	AccuracyRegion  Accuracy = 100000
	AccuracyCity    Accuracy = 15000
	AccuracyZip     Accuracy = 3000
	AccuracyExact   Accuracy = 5
	AccuracyUnknown Accuracy = 1000000
)

func (a Accuracy) Float64() float64 {
	return float64(a)
}
