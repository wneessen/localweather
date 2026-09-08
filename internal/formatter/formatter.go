// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/vorlif/spreak"
	"github.com/vorlif/spreak/humanize"
	"github.com/vorlif/spreak/humanize/locale/de"
	"github.com/vorlif/spreak/localize"
	"golang.org/x/text/message"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/vartype"
)

type Formatter struct {
	conf          *config.Config
	localizer     *spreak.Localizer
	humanizer     *humanize.Humanizer
	printer       *message.Printer
	forecastHours uint
}

// Supported languages for humanize
var supportedHumanizers = []*humanize.LocaleData{de.New()}

const (
	// iconURLFormat defines the format string for constructing the URL path to access SVG icons
	// stored in a public directory.
	iconURLFormat     = "/static/icons/%s/%s.svg"
	moonIconURLFormat = "/static/icons/%s/%s.svg"
)

// New initializes and returns a new Formatter instance with the provided configuration and localizer
func New(conf *config.Config, loc *spreak.Localizer) (*Formatter, error) {
	formatter := &Formatter{conf: conf, localizer: loc}

	// Create humanizer
	collection, err := humanize.New(humanize.WithLocale(supportedHumanizers...))
	if err != nil {
		return formatter, fmt.Errorf("failed to create humanizer: %w", err)
	}
	formatter.humanizer = collection.CreateHumanizer(loc.Language())

	// Create localized/humanized printer
	formatter.printer = message.NewPrinter(loc.Language())

	return formatter, nil
}

func (f *Formatter) Humanize[V float64](val V) string {
	return f.printer.Sprintf("%.1f", val)
}

func (f *Formatter) Localize(val string) localize.MsgID {
	want := strings.ToLower(val)
	if i18n, ok := i18nVars[want]; ok {
		return i18n
	}
	return val
}

func (f *Formatter) LocalizeTime(val time.Time) string {
	return f.humanizer.FormatTime(val, humanize.TimeFormat)
}

// WeatherCategory categorizes a WMO 4677 present-weather code (ww, 00-99) into
// general weather conditions such as clear, cloudy, rain, snow, etc.
func (f *Formatter) WeatherCategory(code int) string {
	switch code {
	case 0, 1:
		return "clear"
	case 2, 3, 14:
		return "cloudy"
	case 4, 5:
		return "haze"
	case 6, 7, 8, 9, 30, 31, 32, 33, 34, 35:
		return "dust"
	case 10, 11, 12, 28, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49:
		return "fog"
	case 18, 19:
		return "wind"
	case 15, 16,
		20, 21, 24, 25, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 80, 81, 82:
		return "rain"
	case 22, 23, 26, 36, 37, 38, 39, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 83, 84, 85, 86:
		return "snow"
	case 27, 87, 88, 89, 90:
		return "hail"
	case 13, 17, 29, 91, 92, 93, 94, 95, 96, 97, 98, 99:
		return "thunderstorm"

	default:
		return ""
	}
}

func (f *Formatter) WeatherCondition(code int) localize.MsgID {
	if condition, ok := WMOWeatherCodes[code]; ok {
		return condition
	}
	return ""
}

func (f *Formatter) WeatherSymbol(code int, isDay bool) string {
	if symbol, ok := WMOWeatherIcons[code][isDay]; ok {
		return symbol
	}
	return "󰨹 "
}

func (f *Formatter) WeatherSymbolURL(code int, isDay bool) string {
	nightDayTag := "day"
	if !isDay {
		nightDayTag = "night"
	}
	return fmt.Sprintf(iconURLFormat, f.conf.Weather.IconSet, fmt.Sprintf("wmo-%d-%s", code, nightDayTag))
}

func (f *Formatter) MoonphaseIconURL(phase string) string {
	filename, ok := MoonPhaseIconURL[phase]
	if !ok {
		return ""
	}
	return fmt.Sprintf(moonIconURLFormat, f.conf.Weather.IconSet, fmt.Sprintf("%s", filename))
}

func (f *Formatter) WindDirectionSymbol(deg vartype.VarFloat64) string {
	dir := f.DegToString(deg)
	if symbol, ok := windDirIcons[strings.ToUpper(dir)]; ok {
		return symbol
	}
	return ""
}

func (f *Formatter) DegToString(deg vartype.VarFloat64) string {
	switch {
	case deg.Value() < 22.5:
		return "N"
	case deg.Value() < 67.5:
		return "NE"
	case deg.Value() < 112.5:
		return "E"
	case deg.Value() < 157.5:
		return "SE"
	case deg.Value() < 202.5:
		return "S"
	case deg.Value() < 247.5:
		return "SW"
	case deg.Value() < 292.5:
		return "W"
	case deg.Value() < 337.5:
		return "NW"
	default:
		return "N"
	}
}
