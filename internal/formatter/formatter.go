package formatter

import (
	"fmt"
	"strings"

	"github.com/vorlif/spreak"
	"github.com/vorlif/spreak/humanize"
	"github.com/vorlif/spreak/humanize/locale/de"
	"github.com/vorlif/spreak/localize"
	"golang.org/x/text/message"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/vartype"
)

type Formatter struct {
	localizer     *spreak.Localizer
	humanizer     *humanize.Humanizer
	printer       *message.Printer
	forecastHours uint
}

// Supported languages for humanize
var supportedHumanizers = []*humanize.LocaleData{de.New()}

// New initializes and returns a new Formatter instance with the provided configuration and localizer
func New(conf *config.Config, loc *spreak.Localizer) (*Formatter, error) {
	formatter := &Formatter{localizer: loc}

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

// WeatherCategory categorizes a weather code into general weather conditions such as clear, cloudy, rain, snow, etc.
func (f *Formatter) WeatherCategory(code int) string {
	switch code {
	case 0, 1:
		return "clear"
	case 2, 3:
		return "cloudy"
	case 45, 48:
		return "fog"
	case 51, 53, 55,
		56, 57,
		61, 63, 65,
		66, 67,
		80, 81, 82:
		return "rain"
	case 71, 73, 75, 77, 85, 86:
		return "snow"
	case 95, 96, 99:
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
