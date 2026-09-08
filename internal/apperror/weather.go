// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

type CurrentWeatherdataNotFoundError struct{}

func (*CurrentWeatherdataNotFoundError) Error() string {
	return "no current weather data found"
}

type PingRequestFailedError struct{}

func (*PingRequestFailedError) Error() string {
	return "failed to ping weather provider"
}

var (
	ErrCurrentWeatherdataNotFound = new(CurrentWeatherdataNotFoundError)
	ErrPingRequestFailed          = new(PingRequestFailedError)
)
