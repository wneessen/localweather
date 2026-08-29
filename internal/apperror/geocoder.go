package apperror

type geoCoderRequiredError struct{}

func (*geoCoderRequiredError) Error() string {
	return "function requires a geocoder"
}

var ErrGeoCoderRequired = new(geoCoderRequiredError)
