// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package geocode

import (
	"context"

	"github.com/wneessen/localweather/internal/types"
)

type Geocoder interface {
	Name() string
	Reverse(context.Context, types.Coordinate) (types.Address, error)
	Search(context.Context, string) (types.Coordinate, error)
}
