// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package apperror

type DBusSignalChannelClosedError struct{}

func (e *DBusSignalChannelClosedError) Error() string {
	return "dbus signal channel closed"
}

var ErrDBusSignalChannelClosed = new(DBusSignalChannelClosedError)
