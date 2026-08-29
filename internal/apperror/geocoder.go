package apperror

type GeoCoderRequiredError struct{}

func (*GeoCoderRequiredError) Error() string {
	return "function requires a geocoder"
}

var ErrGeoCoderRequired = new(GeoCoderRequiredError)
