// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

import "fmt"

type CurrentLocationNotSetError struct{}

func (*CurrentLocationNotSetError) Error() string {
	return "no current location set"
}

type GeoLocationAPIFetchError struct {
	Err error
}

func (e *GeoLocationAPIFetchError) Error() string {
	return fmt.Sprintf("failed to fetch geolocation data from API: %s", e.Err)
}

var ErrCurrentLocationNotSet = new(CurrentLocationNotSetError)
