<!--
SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>

SPDX-License-Identifier: MIT
//-->

# localweather
A local weather web service with automatic geolocation lookup

## What is this?
localweather is the successor of [waybar-weather](https://github.com/wneessen/waybar-weather). It provides the same functionality as waybar-weather, 
but provides the weather data in form of a web service instead. While waybar-weather was built exclusively for 
use with waybar, localweather is designed to be a standalone web service that can be used with any 
application or platform. The reason behind this decision is that I didn't use waybar anymore and switched to
Noctalia. With localweather it is possible to integrate the provided weather data as well as the geolocation
bus and geocoding data into any application or platform. A rough Noctalia plugin already exists and will be 
added to the repository, once I cleaned it up a bit.

## Work in progress
localweather is currently in development and not yet ready for use. Feel free to test it, but everything is currently
subject to change. PRs are not accepted at this time.
