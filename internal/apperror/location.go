package apperror

import "fmt"

type LocationNotSetError struct{}

func (*LocationNotSetError) Error() string {
	return "no current location set"
}

type GeoLocationAPIFetchError struct {
	Err error
}

func (e *GeoLocationAPIFetchError) Error() string {
	return fmt.Sprintf("failed to fetch geolocation data from API: %s", e.Err)
}

var ErrLocationNotSet = new(LocationNotSetError)
