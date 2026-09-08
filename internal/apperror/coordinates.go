// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

import "fmt"

type CoordinateParsingError struct {
	Val string
	Err error
}

func (e *CoordinateParsingError) Error() string {
	return fmt.Sprintf("failed to parse coordinate value (%s): %s", e.Val, e.Err)
}

type NoCoordinatesFoundForAddressError struct {
	Address string
}

func (e *NoCoordinatesFoundForAddressError) Error() string {
	return fmt.Sprintf("no coordinates found for address %q", e.Address)
}
