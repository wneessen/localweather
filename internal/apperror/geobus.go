// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

type NoGeobusEnabledError struct{}

func (*NoGeobusEnabledError) Error() string {
	return "no geobus provider enabled"
}

type NoValidCoordinatesFoundError struct{}

func (*NoValidCoordinatesFoundError) Error() string {
	return "no valid coordinates found in coordinates file"
}

var (
	ErrNoGeobusProvider   = new(NoGeobusEnabledError)
	ErrNoValidCoordinates = new(NoValidCoordinatesFoundError)
)
