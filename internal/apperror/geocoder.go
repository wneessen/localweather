// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

type GeoCoderRequiredError struct{}

func (*GeoCoderRequiredError) Error() string {
	return "function requires a geocoder"
}

var ErrGeoCoderRequired = new(GeoCoderRequiredError)
